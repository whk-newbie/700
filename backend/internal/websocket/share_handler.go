package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"line-management/internal/services"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// HandleShareConnection 处理分享页面WebSocket连接
func HandleShareConnection(c *gin.Context, manager *Manager) error {
	// 从查询参数获取分享码
	shareCode := c.Query("code")
	if shareCode == "" {
		return fmt.Errorf("分享码不能为空")
	}

	// 验证分享码
	shareService := services.NewGroupShareService()
	share, err := shareService.GetGroupShareByCode(c, shareCode)
	if err != nil {
		return fmt.Errorf("分享码验证失败: %w", err)
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return fmt.Errorf("升级WebSocket连接失败: %w", err)
	}

	// 生成客户端ID
	clientID := fmt.Sprintf("share_%s_%d", shareCode, time.Now().Unix())

	// 创建客户端
	client := &Client{
		ID:            clientID,
		Type:          ClientTypeShare,
		ShareCode:     shareCode,
		GroupID:       share.GroupID,
		UserID:        0, // 分享页面没有用户ID
		Conn:          conn,
		Send:          make(chan []byte, 1024),
		LastHeartbeat: time.Now(),
	}

	// 注册客户端
	manager.RegisterClient(client)

	// 发送连接成功消息
	connectSuccessMsg := Message{
		Type: "connected",
		Data: map[string]interface{}{
			"group_id":   share.GroupID,
			"share_code": shareCode,
			"message":    "WebSocket连接成功",
		},
	}
	connectSuccessBytes, _ := json.Marshal(connectSuccessMsg)
	conn.WriteMessage(websocket.TextMessage, connectSuccessBytes)

	// 启动读写协程
	go client.writePump(manager)
	go client.readPumpShare(manager)

	return nil
}

// readPumpShare 分享页面读取消息协程
func (c *Client) readPumpShare(manager *Manager) {
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

		// 分享页面主要接收消息，也可以发送心跳
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

