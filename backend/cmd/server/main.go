package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"line-management/internal/config"
	"line-management/internal/handlers"
	"line-management/internal/middleware"
	"line-management/internal/routes"
	"line-management/internal/scheduler"
	"line-management/pkg/database"
	"line-management/pkg/logger"
	"line-management/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// @title Line账号管理系统API
// @version 1.0
// @description Line账号分组管理与进线统计系统API
//
// ## WebSocket 接口文档
// 由于 Swagger 主要支持 REST API，WebSocket 接口无法在此界面直接测试。
// 详细的 WebSocket 接口文档请访问：http://localhost:8080/docs/websocket
//
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Token，格式：Bearer {token}
func main() {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.InitLogger(); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化Redis
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}

	// 设置运行模式
	gin.SetMode(viper.GetString("gin.mode"))

	// 创建Gin引擎
	r := gin.New()

	// 使用自定义中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"timestamp": time.Now(),
		})
	})

	// 静态文件服务（用于提供二维码图片等）
	r.Static("/static", "./static")

	// 初始化WebSocket管理器
	handlers.InitWebSocketManager()

	// API路由
	apiV1 := r.Group("/api/v1")
	routes.SetupRoutes(apiV1)

	// WebSocket路由
	routes.SetupWebSocketRoutes(r)

	// Swagger文档
	if viper.GetBool("swagger.enable") {
		routes.SetupSwagger(r)
	}

	// 启动定时任务调度器
	taskScheduler := scheduler.NewScheduler()
	taskScheduler.Start()
	defer taskScheduler.Stop()

	// 启动服务器
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		fmt.Printf("服务器启动在端口 %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("正在关闭服务器...")

	// 关闭服务器
	if err := srv.Close(); err != nil {
		log.Printf("关闭服务器失败: %v", err)
	}

	fmt.Println("服务器已关闭")
}
