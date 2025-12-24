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
	return &MessageHandler{
		db:               database.GetDB(),
		groupService:     services.NewGroupService(),
		lineAccountService: services.NewLineAccountService(),
		incomingService:  services.NewIncomingService(nil), // 移除incoming_update回调
		manager:          manager,
	}
}

// SetManager 设置WebSocket管理器引用
func (h *MessageHandler) SetManager(manager *Manager) {
	h.manager = manager
}

// HandleMessage 处理消息
func (h *MessageHandler) HandleMessage(client *Client, message []byte) error {
	// 收到任何消息都更新心跳时间，表示连接活跃
	h.manager.UpdateHeartbeat(client.ID, client.Type)

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

	// 推送分组统计更新
	h.pushGroupStatsUpdate(group.ID)

	// 推送账号统计更新
	h.pushAccountStatsUpdate(group.ID, lineAccount.ID)

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

	// 确定客户类型
	customerType := customerMsg.Data.CustomerType
	if customerType == "" {
		// 检查是否有关联的进线记录，如果有则为"新增线索-实时"，否则为"新增线索-补录"
		var incomingCount int64
		if customerMsg.Data.LineAccountID != "" {
			h.db.Model(&models.IncomingLog{}).
				Where("group_id = ? AND incoming_line_id = ? AND line_account_id = (SELECT id FROM line_accounts WHERE group_id = ? AND line_id = ? AND deleted_at IS NULL)",
					group.ID, customerMsg.Data.CustomerID, group.ID, customerMsg.Data.LineAccountID).
				Count(&incomingCount)
		}
		if incomingCount > 0 {
			customerType = "新增线索-实时"
		} else {
			customerType = "新增线索-补录"
		}
	}

	// 转换数据格式
	customerSyncData := &schemas.CustomerSyncData{
		LineAccountID:  customerMsg.Data.LineAccountID,
		CustomerID:     customerMsg.Data.CustomerID,
		PlatformType:   platformType,
		CustomerType:   customerType,
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

	logger.Infof("处理账号状态变化: line_account_id=%s, status=%s", statusMsg.Data.LineAccountID, statusMsg.Data.OnlineStatus)

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
	// LineAccountID 优先按line_id查询，失败则按id查询
	if err := h.db.Where("group_id = ? AND line_id = ? AND deleted_at IS NULL", group.ID, statusMsg.Data.LineAccountID).First(&lineAccount).Error; err != nil {
		// 如果按line_id查询失败，尝试按id查询
		if err := h.db.Where("id = ? AND group_id = ? AND deleted_at IS NULL", statusMsg.Data.LineAccountID, group.ID).First(&lineAccount).Error; err != nil {
			return fmt.Errorf("Line账号不存在: %w", err)
		}
	}

	// 更新账号状态
	oldStatus := lineAccount.OnlineStatus
	lineAccount.OnlineStatus = statusMsg.Data.OnlineStatus
	now := time.Now()
	lineAccount.LastActiveAt = &now
	if statusMsg.Data.OnlineStatus == "online" {
		lineAccount.LastOnlineTime = &now
	}

	if err := h.db.Save(&lineAccount).Error; err != nil {
		return fmt.Errorf("更新账号状态失败: %w", err)
	}

	logger.Infof("账号状态已更新: %s -> %s (ID: %d)", oldStatus, lineAccount.OnlineStatus, lineAccount.ID)

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
	logger.Infof("推送账号状态更新到前端: group_id=%d, line_account_id=%s, status=%s", groupID, account.LineID, account.OnlineStatus)
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
	logger.Debugf("账号状态更新消息已广播: %s", string(messageBytes))
}

// PushAccountDelete 推送账号删除消息到前端
func (h *MessageHandler) PushAccountDelete(groupID uint, accountID uint, lineAccountID string) {
	logger.Infof("推送账号删除到前端: group_id=%d, account_id=%d, line_account_id=%s", groupID, accountID, lineAccountID)
	deleteMsg := Message{
		Type: "account_deleted",
		Data: map[string]interface{}{
			"group_id":       groupID,
			"account_id":     accountID,
			"line_account_id": lineAccountID,
			"timestamp":      time.Now().Unix(),
		},
	}
	messageBytes, _ := json.Marshal(deleteMsg)
	h.manager.BroadcastToGroup(groupID, messageBytes)
	logger.Debugf("账号删除消息已广播: %s", string(messageBytes))
}

// pushGroupStatsUpdate 推送分组统计更新到前端看板
func (h *MessageHandler) pushGroupStatsUpdate(groupID uint) {
	logger.Infof("推送分组统计更新到前端: group_id=%d", groupID)

	// 获取分组的activation_code
	var group models.Group
	if err := h.db.Select("activation_code").Where("id = ? AND deleted_at IS NULL", groupID).First(&group).Error; err != nil {
		logger.Errorf("获取分组activation_code失败: %v", err)
		return
	}

	// 实时计算统计数据
	stats := h.calculateGroupStats(groupID)

	logger.Infof("分组统计: group_id=%d, total_accounts=%d, online_accounts=%d", groupID, stats["total_accounts"], stats["online_accounts"])

	updateMsg := Message{
		Type: "group_stats_update",
		Data: map[string]interface{}{
			"activation_code":   group.ActivationCode,
			"total_accounts":    stats["total_accounts"],
			"online_accounts":   stats["online_accounts"],
			"total_incoming":    stats["total_incoming"],
			"today_incoming":    stats["today_incoming"],
			"duplicate_incoming": stats["duplicate_incoming"],
			"today_duplicate":   stats["today_duplicate"],
			"timestamp":         time.Now().Unix(),
		},
	}
	messageBytes, _ := json.Marshal(updateMsg)
	h.manager.BroadcastToGroup(groupID, messageBytes)
	logger.Debugf("分组统计更新消息已广播: %s", string(messageBytes))
}

// pushAccountStatsUpdate 推送账号统计更新到前端看板
func (h *MessageHandler) pushAccountStatsUpdate(groupID uint, lineAccountID uint) {
	logger.Infof("推送账号统计更新到前端: group_id=%d, line_account_id=%d", groupID, lineAccountID)

	// 获取账号的line_id
	var lineAccount models.LineAccount
	if err := h.db.Select("line_id").Where("id = ? AND deleted_at IS NULL", lineAccountID).First(&lineAccount).Error; err != nil {
		logger.Errorf("获取账号line_id失败: %v", err)
		return
	}

	// 实时计算统计数据
	stats := h.calculateAccountStats(lineAccountID)

	logger.Infof("账号统计: line_account_id=%d, total_incoming=%d, today_incoming=%d", lineAccountID, stats["total_incoming"], stats["today_incoming"])

	updateMsg := Message{
		Type: "account_stats_update",
		Data: map[string]interface{}{
			"line_id":           lineAccount.LineID,
			"total_incoming":    stats["total_incoming"],
			"today_incoming":    stats["today_incoming"],
			"duplicate_incoming": stats["duplicate_incoming"],
			"today_duplicate":   stats["today_duplicate"],
			"timestamp":         time.Now().Unix(),
		},
	}
	messageBytes, _ := json.Marshal(updateMsg)
	h.manager.BroadcastToGroup(groupID, messageBytes)
	logger.Debugf("账号统计更新消息已广播: %s", string(messageBytes))
}

// calculateGroupStats 实时计算分组统计数据
func (h *MessageHandler) calculateGroupStats(groupID uint) map[string]int64 {
	stats := make(map[string]int64)

	var count int64

	// 总账号数
	h.db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Count(&count)
	stats["total_accounts"] = count

	// 在线账号数
	h.db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL AND online_status = ?", groupID, "online").
		Count(&count)
	stats["online_accounts"] = count

	// 总进线数
	h.db.Model(&models.IncomingLog{}).
		Where("group_id = ?", groupID).
		Count(&count)
	stats["total_incoming"] = count

	// 计算今日时间范围（从重置时间开始）
	todayStartTime := h.getTodayStartTime(groupID)

	// 今日进线数（从重置时间开始）
	h.db.Model(&models.IncomingLog{}).
		Where("group_id = ? AND incoming_time >= ?", groupID, todayStartTime).
		Count(&count)
	stats["today_incoming"] = count

	// 总重复数
	h.db.Model(&models.IncomingLog{}).
		Where("group_id = ? AND is_duplicate = ?", groupID, true).
		Count(&count)
	stats["duplicate_incoming"] = count

	// 今日重复数
	h.db.Model(&models.IncomingLog{}).
		Where("group_id = ? AND incoming_time >= ? AND is_duplicate = ?", groupID, todayStartTime, true).
		Count(&count)
	stats["today_duplicate"] = count

	return stats
}

// calculateAccountStats 实时计算账号统计数据
func (h *MessageHandler) calculateAccountStats(lineAccountID uint) map[string]int64 {
	stats := make(map[string]int64)

	var count int64

	// 获取该账号所属的分组ID，用于计算今日时间范围
	var lineAccount models.LineAccount
	if err := h.db.Where("id = ? AND deleted_at IS NULL", lineAccountID).First(&lineAccount).Error; err != nil {
		logger.Errorf("获取Line账号信息失败: %v", err)
		return stats
	}

	// 总进线数
	h.db.Model(&models.IncomingLog{}).
		Where("line_account_id = ?", lineAccountID).
		Count(&count)
	stats["total_incoming"] = count

	// 计算今日时间范围（从重置时间开始）
	todayStartTime := h.getTodayStartTime(lineAccount.GroupID)

	// 今日进线数（从重置时间开始）
	h.db.Model(&models.IncomingLog{}).
		Where("line_account_id = ? AND incoming_time >= ?", lineAccountID, todayStartTime).
		Count(&count)
	stats["today_incoming"] = count

	// 总重复数
	h.db.Model(&models.IncomingLog{}).
		Where("line_account_id = ? AND is_duplicate = ?", lineAccountID, true).
		Count(&count)
	stats["duplicate_incoming"] = count

	// 今日重复数
	h.db.Model(&models.IncomingLog{}).
		Where("line_account_id = ? AND incoming_time >= ? AND is_duplicate = ?", lineAccountID, todayStartTime, true).
		Count(&count)
	stats["today_duplicate"] = count

	return stats
}

// getTodayStartTime 获取分组今日统计开始时间（重置时间）
func (h *MessageHandler) getTodayStartTime(groupID uint) time.Time {
	var group models.Group
	if err := h.db.Where("id = ? AND deleted_at IS NULL", groupID).First(&group).Error; err != nil {
		// 如果查询失败，使用默认时间（今天09:00:00）
		logger.Warnf("查询分组重置时间失败 (group_id=%d): %v，使用默认时间", groupID, err)
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	}

	// 解析重置时间
	resetTime, err := h.parseResetTime(group.ResetTime)
	if err != nil {
		logger.Warnf("解析分组重置时间失败 (group_id=%d, reset_time=%s): %v，使用默认时间", groupID, group.ResetTime, err)
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	}

	now := time.Now()
	todayResetTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		resetTime.Hour(), resetTime.Minute(), resetTime.Second(),
		0, now.Location(),
	)

	// 如果当前时间还没到今天的重置时间，则使用昨天的重置时间作为开始时间
	if now.Before(todayResetTime) {
		yesterdayResetTime := todayResetTime.AddDate(0, 0, -1)
		return yesterdayResetTime
	}

	return todayResetTime
}

// parseResetTime 解析重置时间字符串（格式：HH:MM:SS）
func (h *MessageHandler) parseResetTime(resetTimeStr string) (time.Time, error) {
	// 默认重置时间为 09:00:00
	if resetTimeStr == "" {
		resetTimeStr = "09:00:00"
	}

	// 解析时间字符串
	layout := "15:04:05"
	parsedTime, err := time.Parse(layout, resetTimeStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

// HandleGroupClientDisconnect 处理分组Windows客户端断开连接
func (h *MessageHandler) HandleGroupClientDisconnect(groupID uint, activationCode string) {
	logger.Infof("处理分组客户端断开连接: group_id=%d, activation_code=%s", groupID, activationCode)

	// 将分组所有账号状态设置为下线
	now := time.Now()
	result := h.db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Updates(map[string]interface{}{
			"online_status":   "offline",
			"last_active_at":  &now,
		})

	if result.Error != nil {
		logger.Errorf("批量更新账号下线状态失败: %v", result.Error)
		return
	}

	affectedCount := result.RowsAffected
	logger.Infof("分组账号下线更新完成: group_id=%d, affected_accounts=%d", groupID, affectedCount)

	// 广播账号状态变化到前端看板
	if affectedCount > 0 {
		h.broadcastGroupAccountsOffline(groupID)
	}

	// 推送分组统计更新
	h.pushGroupStatsUpdate(groupID)
}

// broadcastGroupAccountsOffline 广播分组所有账号下线状态
func (h *MessageHandler) broadcastGroupAccountsOffline(groupID uint) {
	var accounts []models.LineAccount
	if err := h.db.Where("group_id = ? AND deleted_at IS NULL", groupID).Find(&accounts).Error; err != nil {
		logger.Errorf("查询分组账号失败: %v", err)
		return
	}

	// 广播每个账号的下线状态
	for _, account := range accounts {
		h.pushAccountStatusUpdate(groupID, account)
	}
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

