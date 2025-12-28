package middleware

import (
	"fmt"
	"strings"

	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"
	"line-management/pkg/redis"

	"github.com/gin-gonic/gin"
)

// AuthRequired 认证中间件（需要登录）
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorWithErrorCode(c, 2001, "未提供认证Token", "missing_token")
			c.Abort()
			return
		}

		// 检查Token格式（Bearer <token>）
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorWithErrorCode(c, 2002, "Token格式错误", "invalid_token_format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 检查是否是分享 token (格式: share_xxx_timestamp)
		if strings.HasPrefix(tokenString, "share_") {
			// 验证分享 token
			rdb := redis.GetClient()
			shareKey := fmt.Sprintf("share_token:%s", tokenString)
			
			// 从 Redis 获取分享信息
			shareData, err := rdb.HGetAll(c, shareKey).Result()
			if err != nil || len(shareData) == 0 {
				logger.Warnf("分享Token无效或已过期: %s", tokenString)
				utils.ErrorWithErrorCode(c, 2003, "分享链接已过期，请重新验证", "share_token_expired")
				c.Abort()
				return
			}

			// 将分享信息存储到上下文
			c.Set("is_share", true)
			c.Set("share_token", tokenString)
			c.Set("share_code", shareData["share_code"])
			c.Set("group_id", shareData["group_id"])
			c.Set("activation_code", shareData["activation_code"])

			c.Next()
			return
		}

		// 普通 JWT token 验证逻辑
		// 解析Token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			logger.Warnf("Token解析失败: %v", err)
			utils.ErrorWithErrorCode(c, 2003, "Token无效或已过期", "invalid_token")
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
			utils.ErrorWithErrorCode(c, 2003, "Session已过期，请重新登录", "session_expired")
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
			utils.ErrorWithErrorCode(c, 2007, "需要管理员权限", "admin_required")
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
			utils.ErrorWithErrorCode(c, 2007, "需要用户权限", "user_required")
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
			utils.ErrorWithErrorCode(c, 2007, "需要子账号权限", "subaccount_required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// WebSocketAuthRequired WebSocket认证中间件（支持URL参数中的token）
func WebSocketAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 优先从URL参数获取token（WebSocket连接使用）
		tokenString = c.Query("token")
		if tokenString == "" {
			// 如果URL参数中没有，从Header获取Token（兼容普通HTTP请求）
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				utils.ErrorWithErrorCode(c, 2001, "未提供认证Token", "missing_token")
				c.Abort()
				return
			}

			// 检查Token格式（Bearer <token>）
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.ErrorWithErrorCode(c, 2002, "Token格式错误", "invalid_token_format")
				c.Abort()
				return
			}

			tokenString = parts[1]
		}

		// 解析Token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			logger.Warnf("Token解析失败: %v", err)
			utils.ErrorWithErrorCode(c, 2003, "Token无效或已过期", "invalid_token")
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
			utils.ErrorWithErrorCode(c, 2003, "Session已过期，请重新登录", "session_expired")
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

