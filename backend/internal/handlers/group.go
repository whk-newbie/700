package handlers

import (
	"strconv"
	"strings"
	"time"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GetGroups 获取分组列表
// @Summary 获取分组列表
// @Description 获取分组列表（支持分页和筛选）
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param user_id query int false "用户ID"
// @Param category query string false "分类"
// @Param is_active query bool false "是否激活"
// @Param search query string false "搜索（激活码或备注）"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /groups [get]
func GetGroups(c *gin.Context) {
	var params schemas.GroupQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	groupService := services.NewGroupService()
	list, total, err := groupService.GetGroupList(c, &params)
	if err != nil {
		logger.Errorf("获取分组列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取分组列表失败", "internal_error")
		return
	}

	// 分页参数
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	utils.SuccessWithPagination(c, list, page, pageSize, total)
}

// CreateGroup 创建分组
// @Summary 创建分组
// @Description 创建新分组（自动生成激活码）
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.CreateGroupRequest true "创建分组请求"
// @Success 200 {object} schemas.GroupListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /groups [post]
func CreateGroup(c *gin.Context) {
	var req schemas.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	groupService := services.NewGroupService()
	group, err := groupService.CreateGroup(c, &req)
	if err != nil {
		logger.Warnf("创建分组失败: %v", err)
		if err.Error() == "用户不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "user_not_found")
		} else if err.Error() == "用户已被禁用" {
			utils.ErrorWithErrorCode(c, 4001, err.Error(), "user_disabled")
		} else if strings.Contains(err.Error(), "最大分组数量限制") {
			utils.ErrorWithErrorCode(c, 4002, err.Error(), "max_groups_exceeded")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "创建分组失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	response := schemas.GroupListResponse{
		ID:            group.ID,
		UserID:        group.UserID,
		ActivationCode: group.ActivationCode,
		AccountLimit:  group.AccountLimit,
		IsActive:      group.IsActive,
		Remark:        group.Remark,
		Description:   group.Description,
		Category:      group.Category,
		DedupScope:    group.DedupScope,
		ResetTime:     group.ResetTime,
		CreatedAt:     group.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     group.UpdatedAt.Format(time.RFC3339),
	}

	utils.SuccessWithMessage(c, "创建成功", response)
}

// UpdateGroup 更新分组
// @Summary 更新分组
// @Description 更新分组信息
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Param request body schemas.UpdateGroupRequest true "更新分组请求"
// @Success 200 {object} schemas.GroupListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /groups/:id [put]
func UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的分组ID", "invalid_id")
		return
	}

	var req schemas.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	groupService := services.NewGroupService()
	group, err := groupService.UpdateGroup(c, uint(id), &req)
	if err != nil {
		logger.Warnf("更新分组失败: %v", err)
		if err.Error() == "分组不存在" {
			utils.ErrorWithErrorCode(c, 3002, err.Error(), "group_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "更新分组失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	var lastLoginAt *string
	if group.LastLoginAt != nil {
		timeStr := group.LastLoginAt.Format(time.RFC3339)
		lastLoginAt = &timeStr
	}

	response := schemas.GroupListResponse{
		ID:            group.ID,
		UserID:        group.UserID,
		ActivationCode: group.ActivationCode,
		AccountLimit:  group.AccountLimit,
		IsActive:      group.IsActive,
		Remark:        group.Remark,
		Description:   group.Description,
		Category:      group.Category,
		DedupScope:    group.DedupScope,
		ResetTime:     group.ResetTime,
		CreatedAt:     group.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     group.UpdatedAt.Format(time.RFC3339),
		LastLoginAt:   lastLoginAt,
	}

	utils.SuccessWithMessage(c, "更新成功", response)
}

// DeleteGroup 删除分组
// @Summary 删除分组
// @Description 软删除分组
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /groups/:id [delete]
func DeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的分组ID", "invalid_id")
		return
	}

	groupService := services.NewGroupService()
	if err := groupService.DeleteGroup(c, uint(id)); err != nil {
		logger.Warnf("删除分组失败: %v", err)
		if err.Error() == "分组不存在" {
			utils.ErrorWithErrorCode(c, 3002, err.Error(), "group_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "删除分组失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// RegenerateActivationCode 重新生成激活码
// @Summary 重新生成激活码
// @Description 为指定分组重新生成激活码
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /groups/:id/regenerate-code [post]
func RegenerateActivationCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的分组ID", "invalid_id")
		return
	}

	groupService := services.NewGroupService()
	newCode, err := groupService.RegenerateActivationCode(c, uint(id))
	if err != nil {
		logger.Warnf("重新生成激活码失败: %v", err)
		if err.Error() == "分组不存在" {
			utils.ErrorWithErrorCode(c, 3002, err.Error(), "group_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "重新生成激活码失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "重新生成成功", gin.H{
		"activation_code": newCode,
	})
}

// GetGroupCategories 获取分组分类列表
// @Summary 获取分组分类列表
// @Description 获取所有分组分类
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} schemas.ErrorResponse
// @Router /groups/categories [get]
func GetGroupCategories(c *gin.Context) {
	groupService := services.NewGroupService()
	categories, err := groupService.GetCategories(c)
	if err != nil {
		logger.Errorf("获取分组分类失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取分组分类失败", "internal_error")
		return
	}

	utils.SuccessWithMessage(c, "获取成功", gin.H{
		"categories": categories,
	})
}

// BatchDeleteGroups 批量删除分组
// @Summary 批量删除分组
// @Description 批量软删除分组
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.BatchDeleteGroupsRequest true "批量删除请求"
// @Success 200 {object} schemas.BatchOperationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /groups/batch/delete [post]
func BatchDeleteGroups(c *gin.Context) {
	var req schemas.BatchDeleteGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	groupService := services.NewGroupService()
	successCount, failedIDs, err := groupService.BatchDeleteGroups(c, req.IDs)
	if err != nil {
		logger.Errorf("批量删除分组失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "批量删除分组失败", "internal_error")
		return
	}

	failCount := len(failedIDs)
	response := schemas.BatchOperationResponse{
		SuccessCount: successCount,
		FailCount:    failCount,
	}
	if failCount > 0 {
		response.FailedIDs = failedIDs
	}

	utils.SuccessWithMessage(c, "批量删除完成", response)
}

// GenerateSubAccountToken 为分组生成子账户Token
// @Summary 生成子账户Token
// @Description 管理员可以为任何分组生成子账户登录Token，普通用户只能为自己管理的分组生成Token，用于在新标签页自动登录
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /groups/:id/generate-subaccount-token [post]
func GenerateSubAccountToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的分组ID", "invalid_id")
		return
	}

	groupService := services.NewGroupService()
	token, err := groupService.GenerateSubAccountTokenForUser(c, uint(id))
	if err != nil {
		logger.Warnf("生成子账户Token失败: %v", err)
		if err.Error() == "分组不存在" {
			utils.ErrorWithErrorCode(c, 3002, err.Error(), "group_not_found")
		} else if err.Error() == "分组已被禁用" {
			utils.ErrorWithErrorCode(c, 4003, err.Error(), "group_disabled")
		} else if err.Error() == "无权访问该分组" {
			utils.ErrorWithErrorCode(c, 2007, err.Error(), "permission_denied")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "生成Token失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "生成成功", gin.H{
		"token": token,
	})
}

// BatchUpdateGroups 批量更新分组
// @Summary 批量更新分组
// @Description 批量更新分组的状态、分类、去重范围等
// @Tags 分组管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.BatchUpdateGroupsRequest true "批量更新请求"
// @Success 200 {object} schemas.BatchOperationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /groups/batch/update [post]
func BatchUpdateGroups(c *gin.Context) {
	var req schemas.BatchUpdateGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	// 验证至少有一个更新字段
	if req.IsActive == nil && req.Category == "" && req.DedupScope == "" {
		utils.ErrorWithErrorCode(c, 1001, "至少需要提供一个更新字段", "invalid_params")
		return
	}

	groupService := services.NewGroupService()
	successCount, failedIDs, err := groupService.BatchUpdateGroups(c, req.IDs, &req)
	if err != nil {
		logger.Errorf("批量更新分组失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "批量更新分组失败", "internal_error")
		return
	}

	failCount := len(failedIDs)
	response := schemas.BatchOperationResponse{
		SuccessCount: successCount,
		FailCount:    failCount,
	}
	if failCount > 0 {
		response.FailedIDs = failedIDs
	}

	utils.SuccessWithMessage(c, "批量更新完成", response)
}

