package websocket

import (
	"sync"
	"time"

	"line-management/internal/config"
	"line-management/pkg/logger"

	"github.com/gorilla/websocket"
)

const (
	// 心跳超时时间（秒）
	HeartbeatTimeout = 65
	// 心跳检测间隔（秒）
	HeartbeatCheckInterval = 10
	// 写超时时间（秒）
	WriteTimeout = 10
	// 读超时时间（秒）
	ReadTimeout = 60
	// Pong等待时间（秒）
	PongWait = 60
	// Ping周期（秒）
	PingPeriod = 54
)

// ClientDisconnectCallback 客户端断开连接回调函数
type ClientDisconnectCallback func(groupID uint, activationCode string)

// Manager WebSocket连接管理器
type Manager struct {
	// Windows客户端连接池 key: activation_code + conn_id
	clientClients map[string]*Client
	// 前端看板连接池 key: user_id + conn_id
	dashboardClients map[string]*Client
	// 注册通道
	register chan *Client
	// 注销通道
	unregister chan *Client
	// 广播通道（发送给所有前端看板）
	broadcast chan []byte
	// 互斥锁
	mu sync.RWMutex
	// 关闭通道
	close chan struct{}
	// Windows客户端断开连接回调
	onClientDisconnect ClientDisconnectCallback
}

// NewManager 创建连接管理器
func NewManager(onClientDisconnect ClientDisconnectCallback) *Manager {
	return &Manager{
		clientClients:      make(map[string]*Client),
		dashboardClients:   make(map[string]*Client),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		broadcast:          make(chan []byte, 256),
		close:              make(chan struct{}),
		onClientDisconnect: onClientDisconnect,
	}
}

// Run 启动管理器
func (m *Manager) Run() {
	// 启动心跳检测
	go m.startHeartbeatChecker()

	for {
		select {
		case client := <-m.register:
			m.registerClient(client)

		case client := <-m.unregister:
			m.unregisterClient(client)

		case message := <-m.broadcast:
			m.broadcastToDashboards(message)

		case <-m.close:
			return
		}
	}
}

// RegisterClient 注册客户端
func (m *Manager) RegisterClient(client *Client) {
	m.register <- client
}

// UnregisterClient 注销客户端
func (m *Manager) UnregisterClient(client *Client) {
	m.unregister <- client
}

// BroadcastToDashboards 广播消息到所有前端看板
func (m *Manager) BroadcastToDashboards(message []byte) {
	select {
	case m.broadcast <- message:
	default:
		logger.Warn("广播通道已满，丢弃消息")
	}
}

// BroadcastToGroup 广播消息到指定分组的所有前端看板
func (m *Manager) BroadcastToGroup(groupID uint, message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.dashboardClients {
		if client.GroupID == groupID || client.GroupID == 0 {
			select {
			case client.Send <- message:
			default:
				// 发送失败，关闭连接
				close(client.Send)
				delete(m.dashboardClients, client.ID)
			}
		}
	}
}

// GetClientCount 获取客户端数量
func (m *Manager) GetClientCount() (clientCount, dashboardCount int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clientClients), len(m.dashboardClients)
}

// GetClientsByActivationCode 根据激活码获取客户端列表
func (m *Manager) GetClientsByActivationCode(activationCode string) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var clients []*Client
	for _, client := range m.clientClients {
		if client.ActivationCode == activationCode {
			clients = append(clients, client)
		}
	}
	return clients
}

// registerClient 注册客户端（内部方法）
func (m *Manager) registerClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client.RegisteredAt = time.Now()
	client.LastHeartbeat = time.Now()

	if client.Type == ClientTypeWindows {
		m.clientClients[client.ID] = client
		logger.Infof("Windows客户端已注册: ID=%s, ActivationCode=%s, GroupID=%d", client.ID, client.ActivationCode, client.GroupID)
	} else if client.Type == ClientTypeDashboard {
		m.dashboardClients[client.ID] = client
		logger.Infof("前端看板已注册: ID=%s, UserID=%d, GroupID=%d", client.ID, client.UserID, client.GroupID)
	}
}

// unregisterClient 注销客户端（内部方法）
func (m *Manager) unregisterClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client.Type == ClientTypeWindows {
		if _, ok := m.clientClients[client.ID]; ok {
			delete(m.clientClients, client.ID)
			logger.Infof("Windows客户端已注销: ID=%s, ActivationCode=%s, GroupID=%d", client.ID, client.ActivationCode, client.GroupID)

			// 处理分组客户端断开连接，将所有账号下线
			if m.onClientDisconnect != nil {
				go m.onClientDisconnect(client.GroupID, client.ActivationCode)
			}
		}
	} else if client.Type == ClientTypeDashboard {
		if _, ok := m.dashboardClients[client.ID]; ok {
			delete(m.dashboardClients, client.ID)
			logger.Infof("前端看板已注销: ID=%s, UserID=%d", client.ID, client.UserID)
		}
	}

	close(client.Send)
}

// broadcastToDashboards 广播消息到所有前端看板（内部方法）
func (m *Manager) broadcastToDashboards(message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.dashboardClients {
		select {
		case client.Send <- message:
		default:
			// 发送失败，关闭连接
			close(client.Send)
			delete(m.dashboardClients, client.ID)
		}
	}
}

// startHeartbeatChecker 启动心跳检测器
func (m *Manager) startHeartbeatChecker() {
	ticker := time.NewTicker(HeartbeatCheckInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkHeartbeats()
		case <-m.close:
			return
		}
	}
}

// checkHeartbeats 检查心跳
func (m *Manager) checkHeartbeats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	timeout := HeartbeatTimeout * time.Second

	// 检查Windows客户端
	for id, client := range m.clientClients {
		if now.Sub(client.LastHeartbeat) > timeout {
			logger.Warnf("Windows客户端心跳超时，断开连接: ID=%s, ActivationCode=%s", client.ID, client.ActivationCode)
			delete(m.clientClients, id)
			close(client.Send)
			client.Conn.Close()
		}
	}

	// 检查前端看板
	for id, client := range m.dashboardClients {
		if now.Sub(client.LastHeartbeat) > timeout {
			logger.Warnf("前端看板心跳超时，断开连接: ID=%s, UserID=%d", client.ID, client.UserID)
			delete(m.dashboardClients, id)
			close(client.Send)
			client.Conn.Close()
		}
	}
}

// UpdateHeartbeat 更新客户端心跳时间
func (m *Manager) UpdateHeartbeat(clientID string, clientType ClientType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clientType == ClientTypeWindows {
		if client, ok := m.clientClients[clientID]; ok {
			client.LastHeartbeat = time.Now()
		}
	} else if clientType == ClientTypeDashboard {
		if client, ok := m.dashboardClients[clientID]; ok {
			client.LastHeartbeat = time.Now()
		}
	}
}

// Close 关闭管理器
func (m *Manager) Close() {
	close(m.close)
	
	m.mu.Lock()
	defer m.mu.Unlock()

	// 关闭所有客户端连接
	for _, client := range m.clientClients {
		close(client.Send)
		client.Conn.Close()
	}
	for _, client := range m.dashboardClients {
		close(client.Send)
		client.Conn.Close()
	}
}

// SetReadDeadline 设置读超时
func SetReadDeadline(conn *websocket.Conn) error {
	cfg := config.GlobalConfig.WebSocket
	timeout := time.Duration(cfg.ReadTimeout) * time.Second
	if timeout == 0 {
		timeout = ReadTimeout * time.Second
	}
	return conn.SetReadDeadline(time.Now().Add(timeout))
}

// SetWriteDeadline 设置写超时
func SetWriteDeadline(conn *websocket.Conn) error {
	cfg := config.GlobalConfig.WebSocket
	timeout := time.Duration(cfg.WriteTimeout) * time.Second
	if timeout == 0 {
		timeout = WriteTimeout * time.Second
	}
	return conn.SetWriteDeadline(time.Now().Add(timeout))
}

