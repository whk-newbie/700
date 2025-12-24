package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"line-management/internal/models"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源（生产环境应该限制）
		return true
	},
}

// HandleClientConnection 处理Windows客户端WebSocket连接
func HandleClientConnection(c *gin.Context, manager *Manager) error {
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return fmt.Errorf("升级WebSocket连接失败: %w", err)
	}

	// 获取激活码和token
	activationCode := c.Query("activation_code")
	token := c.Query("token")

	if activationCode == "" {
		conn.Close()
		return errors.New("缺少激活码参数")
	}

	// 验证激活码和token
	group, err := validateActivationCode(activationCode, token)
	if err != nil {
		conn.Close()
		return err
	}

	// 生成客户端ID
	clientID := generateClientID()

	// 创建客户端
	client := &Client{
		ID:             clientID,
		Type:           ClientTypeWindows,
		ActivationCode: activationCode,
		GroupID:        group.ID,
		Conn:           conn,
		Send:           make(chan []byte, 1024),
		LastHeartbeat:  time.Now(),
	}

	// 注册客户端
	manager.RegisterClient(client)

	// 发送认证成功消息
	authSuccessMsg := Message{
		Type: "auth_success",
		Data: map[string]interface{}{
			"group_id":        group.ID,
			"activation_code": activationCode,
			"message":         "认证成功，请同步Line账号列表",
		},
	}
	authSuccessBytes, _ := json.Marshal(authSuccessMsg)
	conn.WriteMessage(websocket.TextMessage, authSuccessBytes)

	// 启动读写协程
	go client.writePump(manager)
	go client.readPump(manager)

	return nil
}

// validateActivationCode 验证激活码
func validateActivationCode(activationCode, token string) (*models.Group, error) {
	db := database.GetDB()

	var group models.Group
	if err := db.Where("activation_code = ? AND deleted_at IS NULL", activationCode).First(&group).Error; err != nil {
		return nil, fmt.Errorf("激活码不存在或已被删除")
	}

	// 检查分组是否激活
	if !group.IsActive {
		return nil, errors.New("分组已被禁用")
	}

	// TODO: 验证token（如果需要）
	// 目前先简单验证激活码

	return &group, nil
}

// generateClientID 生成客户端ID
func generateClientID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// readPump 读取消息协程
func (c *Client) readPump(manager *Manager) {
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

	// 创建消息处理器
	handler := NewMessageHandler(manager)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("WebSocket读取错误: %v", err)
			}
			break
		}

		// 处理消息
		if err := handler.HandleMessage(c, message); err != nil {
			logger.Errorf("处理消息失败: %v", err)
			// 通过Send通道发送错误消息，避免并发写入
			errorMsg := Message{
				Type:  "error",
				Error: err.Error(),
			}
			errorBytes, _ := json.Marshal(errorMsg)
			select {
			case c.Send <- errorBytes:
			default:
				// 如果Send通道已满，记录错误但不阻塞
				logger.Warnf("发送错误消息失败，通道已满: %v", err)
			}
		}
	}
}

// writePump 写入消息协程
func (c *Client) writePump(manager *Manager) {
	ticker := time.NewTicker(PingPeriod * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			SetWriteDeadline(c.Conn)
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			SetWriteDeadline(c.Conn)
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

