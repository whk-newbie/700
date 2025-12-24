package websocket

import (
	"encoding/json"
	"line-management/pkg/logger"
)

// Hub 消息广播中心（单例）
var globalHub *Hub

// Hub 消息广播中心
type Hub struct {
	manager *Manager
}

// NewHub 创建消息广播中心
func NewHub(manager *Manager) *Hub {
	return &Hub{
		manager: manager,
	}
}

// InitHub 初始化全局Hub
func InitHub(manager *Manager) {
	globalHub = NewHub(manager)
}

// GetHub 获取全局Hub
func GetHub() *Hub {
	return globalHub
}

// BroadcastAccountStatusChange 广播账号状态变化
func (h *Hub) BroadcastAccountStatusChange(groupID uint, lineAccountID string, onlineStatus string) {
	message := Message{
		Type: "account_status_change",
		Data: map[string]interface{}{
			"group_id":        groupID,
			"line_account_id": lineAccountID,
			"online_status":   onlineStatus,
		},
	}
	h.broadcast(message)
}


// BroadcastStatsUpdate 广播统计更新
func (h *Hub) BroadcastStatsUpdate(groupID uint, stats map[string]interface{}) {
	message := Message{
		Type: "stats_update",
		Data: map[string]interface{}{
			"group_id": groupID,
			"stats":    stats,
		},
	}
	h.broadcast(message)
}

// BroadcastToGroup 广播消息到指定分组
func (h *Hub) BroadcastToGroup(groupID uint, messageType string, data interface{}) {
	message := Message{
		Type: messageType,
		Data: data,
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Errorf("序列化消息失败: %v", err)
		return
	}
	h.manager.BroadcastToGroup(groupID, messageBytes)
}

// BroadcastToAll 广播消息到所有前端看板
func (h *Hub) BroadcastToAll(messageType string, data interface{}) {
	message := Message{
		Type: messageType,
		Data: data,
	}
	h.broadcast(message)
}

// broadcast 广播消息（内部方法）
func (h *Hub) broadcast(message Message) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Errorf("序列化消息失败: %v", err)
		return
	}
	h.manager.BroadcastToDashboards(messageBytes)
}

