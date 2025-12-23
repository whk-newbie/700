package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"gorm.io/gorm"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	db              *gorm.DB
	groupService    *services.GroupService
	lineAccountService *services.LineAccountService
	incomingService *services.IncomingService
	manager         *Manager
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(manager *Manager) *MessageHandler {
	// 创建进线更新回调函数
	updateCallback := func(groupID uint, lineAccountID uint, incomingLineID string, isDuplicate bool) {
		hub := GetHub()
		if hub != nil {
			updateData := map[string]interface{}{
				"group_id":        groupID,
				"line_account_id": lineAccountID,
				"incoming_line_id": incomingLineID,
				"is_duplicate":    isDuplicate,
				"timestamp":       time.Now().Unix(),
			}
			hub.BroadcastToGroup(groupID, "incoming_update", updateData)
		}
	}
	
	return &MessageHandler{
		db:               database.GetDB(),
		groupService:     services.NewGroupService(),
		lineAccountService: services.NewLineAccountService(),
		incomingService:  services.NewIncomingService(updateCallback),
		manager:          manager,
	}
}

// HandleMessage 处理消息
func (h *MessageHandler) HandleMessage(client *Client, message []byte) error {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	logger.Debugf("收到消息: Type=%s, ActivationCode=%s", msg.Type, msg.ActivationCode)

	switch msg.Type {
	case "heartbeat":
		return h.handleHeartbeat(client, &msg)
	case "sync_line_accounts":
		return h.handleSyncLineAccounts(client, message)
	case "incoming":
		return h.handleIncoming(client, message)
	case "customer_sync":
		return h.handleCustomerSync(client, message)
	case "follow_up_sync":
		return h.handleFollowUpSync(client, message)
	case "account_status_change":
		return h.handleAccountStatusChange(client, message)
	default:
		return fmt.Errorf("未知的消息类型: %s", msg.Type)
	}
}

// handleHeartbeat 处理心跳消息
func (h *MessageHandler) handleHeartbeat(client *Client, msg *Message) error {
	// 更新心跳时间
	h.manager.UpdateHeartbeat(client.ID, client.Type)

	// 发送心跳响应（告知客户端服务器正常）
	response := Message{
		Type:          "heartbeat_ack",
		ActivationCode: client.ActivationCode,
		Timestamp:     time.Now().Unix(),
		Data: map[string]interface{}{
			"status": "ok",
			"message": "心跳正常",
		},
	}
	return h.sendMessage(client, response)
}

// handleSyncLineAccounts 处理同步Line账号消息
func (h *MessageHandler) handleSyncLineAccounts(client *Client, message []byte) error {
	var syncMsg SyncLineAccountsMessage
	if err := json.Unmarshal(message, &syncMsg); err != nil {
		return fmt.Errorf("解析同步账号消息失败: %w", err)
	}

	// 验证激活码
	if syncMsg.ActivationCode != client.ActivationCode {
		return errors.New("激活码不匹配")
	}

	// 获取分组信息
	var group models.Group
	if err := h.db.Where("activation_code = ? AND deleted_at IS NULL", syncMsg.ActivationCode).First(&group).Error; err != nil {
		return fmt.Errorf("分组不存在: %w", err)
	}

	createdCount := 0
	updatedCount := 0
	var accountResults []map[string]interface{}

	// 处理每个账号
	for _, accountData := range syncMsg.Data {
		// 查找或创建账号
		var account models.LineAccount
		err := h.db.Where("group_id = ? AND line_id = ? AND deleted_at IS NULL", group.ID, accountData.LineID).First(&account).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新账号
				account = models.LineAccount{
					GroupID:        group.ID,
					ActivationCode: syncMsg.ActivationCode,
					LineID:         accountData.LineID,
					PlatformType:   accountData.PlatformType,
					DisplayName:    accountData.DisplayName,
					PhoneNumber:    accountData.PhoneNumber,
					ProfileURL:     accountData.ProfileURL,
					AvatarURL:      accountData.AvatarURL,
					Bio:            accountData.Bio,
					StatusMessage:  accountData.StatusMessage,
					OnlineStatus:   accountData.OnlineStatus,
				}
				if account.OnlineStatus == "" {
					account.OnlineStatus = "online"
				}
				now := time.Now()
				account.FirstLoginAt = &now
				account.LastActiveAt = &now
				if account.OnlineStatus == "online" {
					account.LastOnlineTime = &now
				}

				if err := h.db.Create(&account).Error; err != nil {
					logger.Errorf("创建Line账号失败: %v", err)
					continue
				}

				// 初始化统计
				stats := models.LineAccountStats{
					LineAccountID: account.ID,
				}
				h.db.Create(&stats)

				// 生成二维码
				if account.ProfileURL != "" {
					qrService := services.NewQRService()
					qrPath, err := qrService.GenerateQRCode(account.ID, account.ProfileURL)
					if err == nil {
						account.QRCodePath = qrPath
						h.db.Save(&account)
					}
				}

				createdCount++
				accountResults = append(accountResults, map[string]interface{}{
					"line_id":    accountData.LineID,
					"account_id": account.ID,
					"status":     "created",
				})
			} else {
				logger.Errorf("查询Line账号失败: %v", err)
				continue
			}
		} else {
			// 更新现有账号
			account.DisplayName = accountData.DisplayName
			account.PhoneNumber = accountData.PhoneNumber
			account.ProfileURL = accountData.ProfileURL
			account.AvatarURL = accountData.AvatarURL
			account.Bio = accountData.Bio
			account.StatusMessage = accountData.StatusMessage
			if accountData.OnlineStatus != "" {
				account.OnlineStatus = accountData.OnlineStatus
			}
			now := time.Now()
			account.LastActiveAt = &now
			if account.OnlineStatus == "online" {
				account.LastOnlineTime = &now
			}

			if err := h.db.Save(&account).Error; err != nil {
				logger.Errorf("更新Line账号失败: %v", err)
				continue
			}

			updatedCount++
			accountResults = append(accountResults, map[string]interface{}{
				"line_id":    accountData.LineID,
				"account_id": account.ID,
				"status":     "updated",
			})
		}
	}

	// 发送同步结果
	response := Message{
		Type: "sync_result",
		Data: map[string]interface{}{
			"success":       true,
			"created_count": createdCount,
			"updated_count": updatedCount,
			"accounts":      accountResults,
		},
	}
	return h.sendMessage(client, response)
}

// handleIncoming 处理进线消息
func (h *MessageHandler) handleIncoming(client *Client, message []byte) error {
	var incomingMsg IncomingMessage
	if err := json.Unmarshal(message, &incomingMsg); err != nil {
		return fmt.Errorf("解析进线消息失败: %w", err)
	}

	// 验证激活码
	if incomingMsg.ActivationCode != client.ActivationCode {
		return errors.New("激活码不匹配")
	}

	// 获取分组信息
	var group models.Group
	if err := h.db.Where("activation_code = ? AND deleted_at IS NULL", incomingMsg.ActivationCode).First(&group).Error; err != nil {
		return fmt.Errorf("分组不存在: %w", err)
	}

	// 查找Line账号
	var lineAccount models.LineAccount
	if err := h.db.Where("group_id = ? AND line_id = ? AND deleted_at IS NULL", group.ID, incomingMsg.Data.LineAccountID).First(&lineAccount).Error; err != nil {
		return fmt.Errorf("Line账号不存在: %w", err)
	}

	// 转换数据格式（从websocket.IncomingData转换为services.IncomingData）
	incomingData := services.IncomingData{
		LineAccountID:  incomingMsg.Data.LineAccountID,
		IncomingLineID: incomingMsg.Data.IncomingLineID,
		Timestamp:      incomingMsg.Data.Timestamp,
		DisplayName:    incomingMsg.Data.DisplayName,
		AvatarURL:      incomingMsg.Data.AvatarURL,
		PhoneNumber:    incomingMsg.Data.PhoneNumber,
	}
	
	// 调用进线处理服务
	if err := h.incomingService.ProcessIncoming(&incomingData, lineAccount.ID, group.ID, group.DedupScope); err != nil {
		logger.Errorf("处理进线数据失败: %v", err)
		return fmt.Errorf("处理进线数据失败: %w", err)
	}

	// 发送确认消息
	response := Message{
		Type: "incoming_received",
		Data: map[string]interface{}{
			"line_account_id": incomingMsg.Data.LineAccountID,
			"incoming_line_id": incomingMsg.Data.IncomingLineID,
			"status": "processed",
		},
	}
	return h.sendMessage(client, response)
}

// handleCustomerSync 处理客户同步消息
func (h *MessageHandler) handleCustomerSync(client *Client, message []byte) error {
	var customerMsg CustomerSyncMessage
	if err := json.Unmarshal(message, &customerMsg); err != nil {
		return fmt.Errorf("解析客户同步消息失败: %w", err)
	}

	// 验证激活码
	if customerMsg.ActivationCode != client.ActivationCode {
		return errors.New("激活码不匹配")
	}

	// 获取分组信息
	var group models.Group
	if err := h.db.Where("activation_code = ? AND deleted_at IS NULL", customerMsg.ActivationCode).First(&group).Error; err != nil {
		return fmt.Errorf("分组不存在: %w", err)
	}

	// 查找Line账号（如果提供了line_account_id）
	var lineAccountID *uint
	if customerMsg.Data.LineAccountID != "" {
		var lineAccount models.LineAccount
		if err := h.db.Where("group_id = ? AND line_id = ? AND deleted_at IS NULL", group.ID, customerMsg.Data.LineAccountID).
			First(&lineAccount).Error; err == nil {
			lineAccountID = &lineAccount.ID
		} else {
			logger.Warnf("Line账号不存在: line_id=%s, group_id=%d", customerMsg.Data.LineAccountID, group.ID)
		}
	}

	// 确定平台类型（如果没有提供，默认为line）
	platformType := customerMsg.Data.PlatformType
	if platformType == "" {
		platformType = "line"
	}

	// 转换数据格式
	customerSyncData := &schemas.CustomerSyncData{
		LineAccountID: customerMsg.Data.LineAccountID,
		CustomerID:     customerMsg.Data.CustomerID,
		PlatformType:   platformType,
		DisplayName:    customerMsg.Data.DisplayName,
		AvatarURL:      customerMsg.Data.AvatarURL,
		PhoneNumber:    customerMsg.Data.PhoneNumber,
		Gender:         customerMsg.Data.Gender,
		Country:        customerMsg.Data.Country,
		Birthday:       customerMsg.Data.Birthday,
		Address:        customerMsg.Data.Address,
		Remark:         customerMsg.Data.Remark,
	}

	// 调用客户同步服务
	customerService := services.NewCustomerService()
	customer, err := customerService.SyncCustomer(group.ID, group.ActivationCode, customerSyncData)
	if err != nil {
		logger.Errorf("同步客户失败: %v", err)
		return fmt.Errorf("同步客户失败: %w", err)
	}

	// 如果提供了line_account_id，更新客户的line_account_id关联
	if lineAccountID != nil && customer.LineAccountID == nil {
		if err := h.db.Model(customer).Update("line_account_id", lineAccountID).Error; err != nil {
			logger.Warnf("更新客户Line账号关联失败: %v", err)
		}
	}

	// 发送确认消息
	response := Message{
		Type: "customer_sync_received",
		Data: map[string]interface{}{
			"customer_id": customerMsg.Data.CustomerID,
			"customer_db_id": customer.ID,
			"status":      "processed",
		},
	}
	return h.sendMessage(client, response)
}

// handleFollowUpSync 处理跟进记录同步消息
func (h *MessageHandler) handleFollowUpSync(client *Client, message []byte) error {
	var followUpMsg FollowUpSyncMessage
	if err := json.Unmarshal(message, &followUpMsg); err != nil {
		return fmt.Errorf("解析跟进记录同步消息失败: %w", err)
	}

	// 验证激活码
	if followUpMsg.ActivationCode != client.ActivationCode {
		return errors.New("激活码不匹配")
	}

	// 获取分组信息
	var group models.Group
	if err := h.db.Where("activation_code = ? AND deleted_at IS NULL", followUpMsg.ActivationCode).First(&group).Error; err != nil {
		return fmt.Errorf("分组不存在: %w", err)
	}

	// 确定平台类型（如果没有提供，默认为line）
	platformType := followUpMsg.Data.PlatformType
	if platformType == "" {
		platformType = "line"
	}

	// 转换数据格式
	followUpSyncData := &schemas.FollowUpSyncData{
		LineAccountID: followUpMsg.Data.LineAccountID,
		CustomerID:    followUpMsg.Data.CustomerID,
		PlatformType:  platformType,
		Content:       followUpMsg.Data.Content,
		Timestamp:     followUpMsg.Data.Timestamp,
	}

	// 调用跟进记录同步服务
	followUpService := services.NewFollowUpService()
	record, err := followUpService.SyncFollowUp(group.ID, group.ActivationCode, followUpSyncData)
	if err != nil {
		logger.Errorf("同步跟进记录失败: %v", err)
		return fmt.Errorf("同步跟进记录失败: %w", err)
	}

	// 发送确认消息
	response := Message{
		Type: "follow_up_sync_received",
		Data: map[string]interface{}{
			"customer_id":   followUpMsg.Data.CustomerID,
			"follow_up_id":  record.ID,
			"status":        "processed",
		},
	}
	return h.sendMessage(client, response)
}

// handleAccountStatusChange 处理账号状态变化消息
func (h *MessageHandler) handleAccountStatusChange(client *Client, message []byte) error {
	var statusMsg AccountStatusChangeMessage
	if err := json.Unmarshal(message, &statusMsg); err != nil {
		return fmt.Errorf("解析账号状态变化消息失败: %w", err)
	}

	// 验证激活码
	if statusMsg.ActivationCode != client.ActivationCode {
		return errors.New("激活码不匹配")
	}

	// 获取分组信息
	var group models.Group
	if err := h.db.Where("activation_code = ? AND deleted_at IS NULL", statusMsg.ActivationCode).First(&group).Error; err != nil {
		return fmt.Errorf("分组不存在: %w", err)
	}

	// 查找Line账号
	var lineAccount models.LineAccount
	if err := h.db.Where("group_id = ? AND line_id = ? AND deleted_at IS NULL", group.ID, statusMsg.Data.LineAccountID).First(&lineAccount).Error; err != nil {
		return fmt.Errorf("Line账号不存在: %w", err)
	}

	// 更新账号状态
	lineAccount.OnlineStatus = statusMsg.Data.OnlineStatus
	now := time.Now()
	lineAccount.LastActiveAt = &now
	if statusMsg.Data.OnlineStatus == "online" {
		lineAccount.LastOnlineTime = &now
	}

	if err := h.db.Save(&lineAccount).Error; err != nil {
		return fmt.Errorf("更新账号状态失败: %w", err)
	}

	// 推送状态更新到前端看板
	h.pushAccountStatusUpdate(group.ID, lineAccount)

	// 获取并推送分组统计更新
	h.pushGroupStatsUpdate(group.ID)

	response := Message{
		Type: "account_status_updated",
		Data: map[string]interface{}{
			"line_account_id": statusMsg.Data.LineAccountID,
			"online_status":   statusMsg.Data.OnlineStatus,
			"status":          "updated",
		},
	}
	return h.sendMessage(client, response)
}

// pushAccountStatusUpdate 推送账号状态更新到前端看板
func (h *MessageHandler) pushAccountStatusUpdate(groupID uint, account models.LineAccount) {
	updateMsg := Message{
		Type: "account_status_change",
		Data: map[string]interface{}{
			"line_account_id": account.LineID,
			"online_status":    account.OnlineStatus,
			"group_id":         groupID,
			"timestamp":        time.Now().Unix(),
		},
	}
	messageBytes, _ := json.Marshal(updateMsg)
	h.manager.BroadcastToGroup(groupID, messageBytes)
}

// pushGroupStatsUpdate 推送分组统计更新到前端看板
func (h *MessageHandler) pushGroupStatsUpdate(groupID uint) {
	statsService := services.NewStatsService()
	groupStats, err := statsService.GetGroupStats(groupID)
	if err != nil {
		logger.Errorf("获取分组统计失败: %v", err)
		return
	}

	// 获取在线账号数
	var onlineCount int64
	h.db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL AND online_status = ?", groupID, "online").
		Count(&onlineCount)

	updateMsg := Message{
		Type: "group_stats_update",
		Data: map[string]interface{}{
			"group_id":        groupID,
			"total_accounts":  groupStats.TotalAccounts,
			"online_accounts": onlineCount,
			"total_incoming":  groupStats.TotalIncoming,
			"today_incoming":  groupStats.TodayIncoming,
			"duplicate_incoming": groupStats.DuplicateIncoming,
			"today_duplicate": groupStats.TodayDuplicate,
			"timestamp":       time.Now().Unix(),
		},
	}
	messageBytes, _ := json.Marshal(updateMsg)
	h.manager.BroadcastToGroup(groupID, messageBytes)
}

// sendMessage 发送消息到客户端
func (h *MessageHandler) sendMessage(client *Client, message Message) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	select {
	case client.Send <- messageBytes:
		return nil
	default:
		return errors.New("发送消息失败：通道已满")
	}
}

