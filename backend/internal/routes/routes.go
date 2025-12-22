package routes

import (
	"line-management/docs"
	"line-management/internal/handlers"
	"line-management/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.RouterGroup) {
	// 认证相关路由（不需要JWT）
	auth := r.Group("/auth")
	{
		auth.POST("/login", handlers.Login)
		auth.POST("/login-subaccount", handlers.LoginSubAccount)
		auth.POST("/logout", middleware.AuthRequired(), handlers.Logout)
		auth.GET("/me", middleware.AuthRequired(), handlers.GetMe)
		auth.POST("/refresh", handlers.RefreshToken)
		auth.GET("/sessions", middleware.AuthRequired(), handlers.GetActiveSessions)
	}

	// 需要认证的路由
	api := r.Group("")
	api.Use(middleware.AuthRequired())
	api.Use(middleware.DataFilter()) // 应用数据过滤中间件
	{
		// 分组管理路由
		groups := api.Group("/groups")
		{
			groups.GET("", handlers.GetGroups)
			groups.POST("", handlers.CreateGroup)
			groups.PUT("/:id", handlers.UpdateGroup)
			groups.DELETE("/:id", handlers.DeleteGroup)
			groups.POST("/:id/regenerate-code", handlers.RegenerateActivationCode)
			groups.POST("/:id/generate-subaccount-token", handlers.GenerateSubAccountToken)
			groups.GET("/categories", handlers.GetGroupCategories)
			// 批量操作
			groups.POST("/batch/delete", handlers.BatchDeleteGroups)
			groups.POST("/batch/update", handlers.BatchUpdateGroups)
		}

		// Line账号管理路由
		lineAccounts := api.Group("/line-accounts")
		{
			lineAccounts.GET("", handlers.GetLineAccounts)
			lineAccounts.POST("", handlers.CreateLineAccount)
			lineAccounts.PUT("/:id", handlers.UpdateLineAccount)
			lineAccounts.DELETE("/:id", handlers.DeleteLineAccount)
			lineAccounts.POST("/:id/generate-qr", handlers.GenerateQRCode)
			// 批量操作
			lineAccounts.POST("/batch/delete", handlers.BatchDeleteLineAccounts)
			lineAccounts.POST("/batch/update", handlers.BatchUpdateLineAccounts)
		}

		// 统计路由
		stats := api.Group("/stats")
		{
			stats.GET("/overview", handlers.GetOverviewStats)
			stats.GET("/group/:id", handlers.GetGroupStats)
			stats.GET("/group/:id/trend", handlers.GetGroupIncomingTrend)
			stats.GET("/account/:id", handlers.GetAccountStats)
			stats.GET("/account/:id/trend", handlers.GetAccountIncomingTrend)
			stats.GET("/incoming-logs", handlers.GetIncomingLogs)
		}
	}

	// 健康检查（不需要认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}

// SetupWebSocketRoutes 设置WebSocket路由
func SetupWebSocketRoutes(r *gin.Engine) {
	// Windows客户端WebSocket连接（不需要JWT认证，使用激活码+token认证）
	r.GET("/api/ws/client", handlers.HandleClientWebSocket)
	
	// 前端看板WebSocket连接（需要JWT认证）
	r.GET("/api/ws/dashboard", middleware.AuthRequired(), handlers.HandleDashboardWebSocket)
}

// SetupSwagger 设置Swagger文档
func SetupSwagger(r *gin.Engine) {
	// 导入docs包以确保SwaggerInfo被初始化
	_ = docs.SwaggerInfo
	
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// WebSocket文档页面（也可以通过 /static/websocket-docs.html 访问）
	r.GET("/docs/websocket", func(c *gin.Context) {
		c.File("./static/websocket-docs.html")
	})
}

