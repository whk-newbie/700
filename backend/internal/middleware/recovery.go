package middleware

import (
	"net/http"
	"runtime/debug"

	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Recovery 错误恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				logger.Error(
					"Panic恢复",
					"error", err,
					"stack", string(debug.Stack()),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				// 返回错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误",
					"data":    nil,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

