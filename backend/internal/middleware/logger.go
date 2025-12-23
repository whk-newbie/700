package middleware

import (
	"time"

	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 检查是否为WebSocket升级请求
		upgrade := c.GetHeader("Upgrade")
		if upgrade == "websocket" {
			// WebSocket升级请求，不记录HTTP日志
			c.Next()
			return
		}

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)

		// 记录日志
		logger.Info(
			"HTTP请求",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", raw,
			"ip", c.ClientIP(),
			"user-agent", c.Request.UserAgent(),
			"latency", latency,
			"error", c.Errors.String(),
		)
	}
}

