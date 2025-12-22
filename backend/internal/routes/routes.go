package routes

import (
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
	{
		// 后续添加的业务路由将在这里
		// 例如：
		// groups := api.Group("/groups")
		// {
		// 	groups.GET("", handlers.GetGroups)
		// 	groups.POST("", handlers.CreateGroup)
		// 	groups.PUT("/:id", handlers.UpdateGroup)
		// 	groups.DELETE("/:id", handlers.DeleteGroup)
		// }
	}

	// 健康检查（不需要认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}

// SetupSwagger 设置Swagger文档
func SetupSwagger(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

