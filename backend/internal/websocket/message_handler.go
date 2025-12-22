package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"line-management/internal/models"
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
	manager         *Manager
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(manager *Manager) *MessageHandler {
	return &MessageHandler{
		db:               database.GetDB(),
		groupService:     services.NewGroupService(),
		lineAccountService: services.NewLineAccountService(),
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

	// 发送心跳响应
	response := Message{
		Type:      "heartbeat",
		Timestamp: time.Now().Unix(),
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

	// TODO: 这里应该调用进线处理服务（第6周实现）
	// 目前先记录日志
	logger.Infof("收到进线数据: GroupID=%d, LineAccountID=%d, IncomingLineID=%s", 
		group.ID, lineAccount.ID, incomingMsg.Data.IncomingLineID)

	// 发送确认消息
	response := Message{
		Type: "incoming_received",
		Data: map[string]interface{}{
			"line_account_id": incomingMsg.Data.LineAccountID,
			"incoming_line_id": incomingMsg.Data.IncomingLineID,
			"status": "received",
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

	// TODO: 第8周实现客户管理模块时处理
	logger.Infof("收到客户同步数据: ActivationCode=%s, CustomerID=%s", 
		customerMsg.ActivationCode, customerMsg.Data.CustomerID)

	response := Message{
		Type: "customer_sync_received",
		Data: map[string]interface{}{
			"customer_id": customerMsg.Data.CustomerID,
			"status":      "received",
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

	// TODO: 第9周实现跟进记录模块时处理
	logger.Infof("收到跟进记录同步数据: ActivationCode=%s, CustomerID=%s", 
		followUpMsg.ActivationCode, followUpMsg.Data.CustomerID)

	response := Message{
		Type: "follow_up_sync_received",
		Data: map[string]interface{}{
			"customer_id": followUpMsg.Data.CustomerID,
			"status":      "received",
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

