package handlers

import (
	"fmt"
	"strconv"
	"time"

	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"
	"line-management/pkg/redis"

	"github.com/gin-gonic/gin"
)

// CreateGroupShare 创建分组分享
// @Summary 创建分组分享
// @Description 为指定分组创建分享链接
// @Tags 分组分享
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /groups/:id/share [post]
func CreateGroupShare(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的分组ID", "invalid_id")
		return
	}

	shareService := services.NewGroupShareService()
	share, err := shareService.CreateGroupShare(c, uint(id), nil) // nil 表示永久有效
	if err != nil {
		logger.Warnf("创建分组分享失败: %v", err)
		if err.Error() == "分组不存在" {
			utils.ErrorWithErrorCode(c, 3002, err.Error(), "group_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "创建分享失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "创建成功", gin.H{
		"share_code": share.ShareCode,
		"password":   share.Password, // 返回密码给管理员（默认与分享码相同）
	})
}

// GetGroupShareInfo 获取分享信息
// @Summary 获取分享信息
// @Description 通过分享码获取分组和账号信息（公开接口，不需要认证）
// @Tags 分组分享
// @Accept json
// @Produce json
// @Param code query string true "分享码"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /share/info [get]
func GetGroupShareInfo(c *gin.Context) {
	shareCode := c.Query("code")
	if shareCode == "" {
		utils.ErrorWithErrorCode(c, 1001, "分享码不能为空", "invalid_params")
		return
	}

	shareService := services.NewGroupShareService()
	share, err := shareService.GetGroupShareByCode(c, shareCode)
	if err != nil {
		logger.Warnf("获取分享信息失败: %v", err)
		if err.Error() == "分享不存在或已失效" || err.Error() == "分享已过期" {
			utils.ErrorWithErrorCode(c, 3003, err.Error(), "share_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "获取分享信息失败", "internal_error")
		}
		return
	}

	// 返回分组基本信息（不验证密码，只是基本信息预览）
	utils.SuccessWithMessage(c, "获取成功", gin.H{
		"group_id":        share.GroupID,
		"activation_code": share.Group.ActivationCode,
		"remark":          share.Group.Remark,
		"description":     share.Group.Description,
		"view_count":      share.ViewCount,
		"require_password": true, // 标记需要密码
	})
}

// VerifySharePassword 验证分享密码
// @Summary 验证分享密码
// @Description 验证分享密码并返回分组详细信息（公开接口，不需要认证）
// @Tags 分组分享
// @Accept json
// @Produce json
// @Param request body map[string]string true "验证请求 {code: 分享码, password: 密码}"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /share/verify [post]
func VerifySharePassword(c *gin.Context) {
	var req struct {
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	shareService := services.NewGroupShareService()
	share, err := shareService.VerifySharePassword(c, req.Code, req.Password)
	if err != nil {
		logger.Warnf("验证分享密码失败: %v", err)
		if err.Error() == "密码错误" {
			utils.ErrorWithErrorCode(c, 2008, "密码错误", "invalid_password")
		} else if err.Error() == "分享不存在或已失效" || err.Error() == "分享已过期" {
			utils.ErrorWithErrorCode(c, 3003, err.Error(), "share_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "验证失败", "internal_error")
		}
		return
	}

	// 验证成功，生成临时访问 token
	shareToken := fmt.Sprintf("share_%s_%d", req.Code, time.Now().Unix())
	
	// 将分享信息存储到 Redis（有效期 24 小时，与 JWT 一致）
	rdb := redis.GetClient()
	shareKey := fmt.Sprintf("share_token:%s", shareToken)
	shareData := map[string]interface{}{
		"group_id":        share.GroupID,
		"share_code":      req.Code,
		"activation_code": share.Group.ActivationCode,
	}
	
	if err := rdb.HSet(c, shareKey, shareData).Err(); err != nil {
		logger.Errorf("存储分享token失败: %v", err)
	}
	if err := rdb.Expire(c, shareKey, 24*time.Hour).Err(); err != nil {
		logger.Errorf("设置分享token过期时间失败: %v", err)
	}

	utils.SuccessWithMessage(c, "验证成功", gin.H{
		"group_id":        share.GroupID,
		"activation_code": share.Group.ActivationCode,
		"remark":          share.Group.Remark,
		"description":     share.Group.Description,
		"view_count":      share.ViewCount,
		"verified":        true,
		"share_token":     shareToken, // 返回临时 token
	})
}

// GetGroupShareByGroupID 获取分组的分享信息
// @Summary 获取分组的分享信息
// @Description 获取指定分组的分享链接
// @Tags 分组分享
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /groups/:id/share [get]
func GetGroupShareByGroupID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的分组ID", "invalid_id")
		return
	}

	shareService := services.NewGroupShareService()
	share, err := shareService.GetGroupShareByGroupID(c, uint(id))
	if err != nil {
		logger.Warnf("获取分组分享信息失败: %v", err)
		if err.Error() == "分享不存在" {
			utils.ErrorWithErrorCode(c, 3003, "该分组还未创建分享", "share_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "获取分享信息失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "获取成功", gin.H{
		"share_code": share.ShareCode,
		"password":   share.Password, // 返回密码给管理员
		"view_count": share.ViewCount,
		"expires_at": share.ExpiresAt,
	})
}

// DeleteGroupShare 删除分组分享
// @Summary 删除分组分享
// @Description 删除指定分组的分享链接
// @Tags 分组分享
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /groups/:id/share [delete]
func DeleteGroupShare(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的分组ID", "invalid_id")
		return
	}

	shareService := services.NewGroupShareService()
	
	// 先获取分享信息
	share, err := shareService.GetGroupShareByGroupID(c, uint(id))
	if err != nil {
		logger.Warnf("获取分组分享信息失败: %v", err)
		if err.Error() == "分享不存在" {
			utils.ErrorWithErrorCode(c, 3003, "该分组还未创建分享", "share_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "获取分享信息失败", "internal_error")
		}
		return
	}

	// 删除分享
	if err := shareService.DeleteGroupShare(c, share.ID); err != nil {
		logger.Warnf("删除分组分享失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "删除分享失败", "internal_error")
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

