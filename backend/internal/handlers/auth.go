package handlers

import (
	"net/http"
	"strings"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Login 用户登录（管理员/普通用户）
// @Summary 用户登录
// @Description 管理员或普通用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body schemas.LoginRequest true "登录请求"
// @Success 200 {object} schemas.LoginResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req schemas.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取客户端IP和User-Agent
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	authService := services.NewAuthService()
	response, err := authService.Login(&req, ipAddress, userAgent)
	if err != nil {
		logger.Warnf("登录失败: %v", err)
		c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
			Error:   "invalid_credentials",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// LoginSubAccount 子账号登录
// @Summary 子账号登录
// @Description 使用激活码和密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body schemas.SubAccountLoginRequest true "子账号登录请求"
// @Success 200 {object} schemas.LoginResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /auth/login-subaccount [post]
func LoginSubAccount(c *gin.Context) {
	var req schemas.SubAccountLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取客户端IP和User-Agent
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	authService := services.NewAuthService()
	response, err := authService.LoginSubAccount(&req, ipAddress, userAgent)
	if err != nil {
		logger.Warnf("子账号登录失败: %v", err)
		c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
			Error:   "invalid_credentials",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout 登出
// @Summary 登出
// @Description 用户登出，删除Session
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	// 从上下文获取claims
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "未授权",
			Error:   "unauthorized",
		})
		return
	}

	jwtClaims := claims.(*utils.JWTClaims)

	// 获取Token
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Token格式错误",
			Error:   "invalid_token_format",
		})
		return
	}
	token := parts[1]

	// 删除Session
	sessionService := services.NewSessionService()
	var userID uint
	if jwtClaims.Role == "subaccount" {
		userID = jwtClaims.GroupID
	} else {
		userID = jwtClaims.UserID
	}

	if err := sessionService.DeleteSession(userID, token); err != nil {
		logger.Warnf("删除Session失败: %v", err)
		// 即使删除失败也返回成功，因为Token可能已过期
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// GetMe 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的信息
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} schemas.UserInfo
// @Failure 401 {object} schemas.ErrorResponse
// @Router /auth/me [get]
func GetMe(c *gin.Context) {
	// 从上下文获取用户信息（由中间件设置）
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "未授权",
			Error:   "unauthorized",
		})
		return
	}

	jwtClaims := claims.(*utils.JWTClaims)

	authService := services.NewAuthService()

	// 根据角色返回不同的信息
	if jwtClaims.Role == "subaccount" {
		// 子账号返回分组信息
		group, err := authService.GetGroupByID(jwtClaims.GroupID)
		if err != nil {
			c.JSON(http.StatusNotFound, schemas.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "分组不存在",
				Error:   "group_not_found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":              group.ID,
			"activation_code": group.ActivationCode,
			"category":        group.Category,
			"role":            "subaccount",
		})
		return
	}

	// 管理员/普通用户返回用户信息
	user, err := authService.GetUserByID(jwtClaims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, schemas.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "用户不存在",
			Error:   "user_not_found",
		})
		return
	}

	c.JSON(http.StatusOK, schemas.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	})
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Description 刷新过期的Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body schemas.RefreshTokenRequest true "刷新Token请求"
// @Success 200 {object} schemas.RefreshTokenResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var req schemas.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	newToken, err := utils.RefreshToken(req.Token)
	if err != nil {
		logger.Warnf("刷新Token失败: %v", err)
		c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
			Error:   "invalid_token",
		})
		return
	}

	c.JSON(http.StatusOK, schemas.RefreshTokenResponse{
		Token:     newToken,
		TokenType: "Bearer",
		ExpiresIn: 24 * 3600, // 24小时
	})
}

// GetActiveSessions 获取当前用户的活跃会话
// @Summary 获取活跃会话
// @Description 获取当前用户的所有活跃Session
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} schemas.ErrorResponse
// @Router /auth/sessions [get]
func GetActiveSessions(c *gin.Context) {
	// 从上下文获取claims
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "未授权",
			Error:   "unauthorized",
		})
		return
	}

	jwtClaims := claims.(*utils.JWTClaims)

	sessionService := services.NewSessionService()
	var userID uint
	if jwtClaims.Role == "subaccount" {
		userID = jwtClaims.GroupID
	} else {
		userID = jwtClaims.UserID
	}

	sessions, err := sessionService.GetUserSessions(userID)
	if err != nil {
		logger.Warnf("获取活跃会话失败: %v", err)
		c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取活跃会话失败",
			Error:   "internal_error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

