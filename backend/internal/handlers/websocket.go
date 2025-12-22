package handlers

import (
	"line-management/internal/utils"
	"line-management/internal/websocket"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

var wsManager *websocket.Manager

// InitWebSocketManager 初始化WebSocket管理器
func InitWebSocketManager() {
	wsManager = websocket.NewManager()
	go wsManager.Run()
	websocket.InitHub(wsManager)
	logger.Info("WebSocket管理器已启动")
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

