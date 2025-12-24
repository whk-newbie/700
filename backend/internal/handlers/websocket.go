package handlers

import (
	"line-management/internal/utils"
	"line-management/internal/websocket"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

var wsManager *websocket.Manager
var messageHandler *websocket.MessageHandler

// InitWebSocketManager 初始化WebSocket管理器
func InitWebSocketManager() {
	// 创建消息处理器
	messageHandler = websocket.NewMessageHandler(nil) // 暂时传nil，后面会设置manager

	// 创建客户端断开连接回调函数
	onClientDisconnect := func(groupID uint, activationCode string) {
		messageHandler.HandleGroupClientDisconnect(groupID, activationCode)
	}

	// 创建WebSocket管理器
	wsManager = websocket.NewManager(onClientDisconnect)
	messageHandler.SetManager(wsManager) // 设置manager引用

	go wsManager.Run()
	websocket.InitHub(wsManager)
	logger.Info("WebSocket管理器已启动")
}

// GetMessageHandler 获取消息处理器
func GetMessageHandler() *websocket.MessageHandler {
	return messageHandler
}

// GetWebSocketManager 获取WebSocket管理器
func GetWebSocketManager() *websocket.Manager {
	return wsManager
}

// HandleClientWebSocket Windows客户端WebSocket连接
func HandleClientWebSocket(c *gin.Context) {
	if err := websocket.HandleClientConnection(c, wsManager); err != nil {
		logger.Errorf("处理Windows客户端WebSocket连接失败: %v", err)
		utils.Error(c, 500, "WebSocket连接失败: "+err.Error())
		return
	}
}

// HandleDashboardWebSocket 前端看板WebSocket连接
func HandleDashboardWebSocket(c *gin.Context) {
	if err := websocket.HandleDashboardConnection(c, wsManager); err != nil {
		logger.Errorf("处理前端看板WebSocket连接失败: %v", err)
		utils.Error(c, 500, "WebSocket连接失败: "+err.Error())
		return
	}
}

