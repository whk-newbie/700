package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// HandleDashboardConnection 处理前端看板WebSocket连接
func HandleDashboardConnection(c *gin.Context, manager *Manager) error {
	// 需要认证
	claims, exists := c.Get("claims")
	if !exists {
		return fmt.Errorf("未认证")
	}

	// 从claims获取用户信息
	userClaims, ok := claims.(*utils.JWTClaims)
	if !ok {
		return fmt.Errorf("无效的认证信息")
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return fmt.Errorf("升级WebSocket连接失败: %w", err)
	}

	// 生成客户端ID（使用用户ID+时间戳）
	clientID := fmt.Sprintf("dashboard_%d_%d", userClaims.UserID, time.Now().Unix())

	// 获取分组ID（如果有）
	groupID := uint(0)
	if userClaims.GroupID > 0 {
		groupID = userClaims.GroupID
	}

	// 创建客户端
	client := &Client{
		ID:            clientID,
		Type:          ClientTypeDashboard,
		UserID:        userClaims.UserID,
		GroupID:       groupID,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		LastHeartbeat: time.Now(),
	}

	// 注册客户端
	manager.RegisterClient(client)

	// 发送连接成功消息
	connectSuccessMsg := Message{
		Type: "connected",
		Data: map[string]interface{}{
			"user_id":  userClaims.UserID,
			"group_id": groupID,
			"message":  "WebSocket连接成功",
		},
	}
	connectSuccessBytes, _ := json.Marshal(connectSuccessMsg)
	conn.WriteMessage(websocket.TextMessage, connectSuccessBytes)

	// 启动读写协程
	go client.writePump(manager)
	go client.readPumpDashboard(manager)

	return nil
}

// readPumpDashboard 前端看板读取消息协程
func (c *Client) readPumpDashboard(manager *Manager) {
	defer func() {
		manager.UnregisterClient(c)
		c.Conn.Close()
	}()

	// 设置读超时
	SetReadDeadline(c.Conn)
	c.Conn.SetReadDeadline(time.Now().Add(PongWait * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		SetReadDeadline(c.Conn)
		c.Conn.SetReadDeadline(time.Now().Add(PongWait * time.Second))
		manager.UpdateHeartbeat(c.ID, c.Type)
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("WebSocket读取错误: %v", err)
			}
			break
		}

		// 前端看板主要接收消息，也可以发送心跳
		var msg Message
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg.Type == "heartbeat" {
				// 处理心跳并回复确认
				manager.UpdateHeartbeat(c.ID, c.Type)
				response := Message{
					Type:      "heartbeat_ack",
					Timestamp: time.Now().Unix(),
					Data: map[string]interface{}{
						"status":  "ok",
						"message": "心跳正常",
					},
				}
				responseBytes, _ := json.Marshal(response)
				c.Conn.WriteMessage(websocket.TextMessage, responseBytes)
			}
		}
	}
}

