package middleware

import (
	"net/http"
	"strings"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AuthRequired 认证中间件（需要登录）
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "未提供认证Token",
				Error:   "missing_token",
			})
			c.Abort()
			return
		}

		// 检查Token格式（Bearer <token>）
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Token格式错误",
				Error:   "invalid_token_format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析Token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			logger.Warnf("Token解析失败: %v", err)
			c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Token无效或已过期",
				Error:   "invalid_token",
			})
			c.Abort()
			return
		}

		// 检查Session是否存在（如果启用了Session管理）
		sessionService := services.NewSessionService()
		var userID uint
		if claims.Role == "subaccount" {
			userID = claims.GroupID
		} else {
			userID = claims.UserID
		}

		// 验证Session是否存在
		if !sessionService.CheckSession(userID, tokenString) {
			logger.Warnf("Session不存在或已过期: user_id=%d", userID)
			c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Session已过期，请重新登录",
				Error:   "session_expired",
			})
			c.Abort()
			return
		}

		// 将claims存储到上下文
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		if claims.GroupID > 0 {
			c.Set("group_id", claims.GroupID)
		}
		if claims.ActivationCode != "" {
			c.Set("activation_code", claims.ActivationCode)
		}

		c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证中间件
		AuthRequired()(c)

		// 如果认证失败，直接返回
		if c.IsAborted() {
			return
		}

		// 检查角色
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, schemas.ErrorResponse{
				Code:    http.StatusForbidden,
				Message: "需要管理员权限",
				Error:   "admin_required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRequired 用户权限中间件（管理员或普通用户）
func UserRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证中间件
		AuthRequired()(c)

		// 如果认证失败，直接返回
		if c.IsAborted() {
			return
		}

		// 检查角色（管理员或普通用户都可以）
		role, exists := c.Get("role")
		if !exists || (role != "admin" && role != "user") {
			c.JSON(http.StatusForbidden, schemas.ErrorResponse{
				Code:    http.StatusForbidden,
				Message: "需要用户权限",
				Error:   "user_required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SubAccountRequired 子账号权限中间件
func SubAccountRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证中间件
		AuthRequired()(c)

		// 如果认证失败，直接返回
		if c.IsAborted() {
			return
		}

		// 检查角色
		role, exists := c.Get("role")
		if !exists || role != "subaccount" {
			c.JSON(http.StatusForbidden, schemas.ErrorResponse{
				Code:    http.StatusForbidden,
				Message: "需要子账号权限",
				Error:   "subaccount_required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

