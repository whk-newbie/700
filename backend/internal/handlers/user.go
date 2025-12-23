package handlers

import (
	"strconv"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表（管理员专用，支持分页和筛选）
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param role query string false "角色" Enums(admin, user)
// @Param is_active query bool false "是否激活"
// @Param search query string false "搜索（用户名或邮箱）"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/users [get]
func GetUsers(c *gin.Context) {
	var params schemas.UserQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	userService := services.NewUserService()
	list, total, err := userService.GetUserList(c, &params)
	if err != nil {
		logger.Errorf("获取用户列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取用户列表失败", "internal_error")
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

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建普通用户（管理员专用）
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.CreateUserRequest true "创建用户请求"
// @Success 200 {object} schemas.UserListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/users [post]
func CreateUser(c *gin.Context) {
	var req schemas.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	userService := services.NewUserService()
	user, err := userService.CreateUser(c, &req)
	if err != nil {
		logger.Warnf("创建用户失败: %v", err)
		if err.Error() == "用户名已存在" {
			utils.ErrorWithErrorCode(c, 4001, err.Error(), "username_exists")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "创建用户失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	response := schemas.UserListResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		MaxGroups: user.MaxGroups,
		IsActive:  user.IsActive,
		CreatedBy: user.CreatedBy,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息（管理员专用）
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body schemas.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} schemas.UserListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /admin/users/{id} [put]
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的用户ID", "invalid_id")
		return
	}

	var req schemas.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	userService := services.NewUserService()
	user, err := userService.UpdateUser(c, uint(id), &req)
	if err != nil {
		logger.Warnf("更新用户失败: %v", err)
		if err.Error() == "用户不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "user_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "更新用户失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	response := schemas.UserListResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		MaxGroups: user.MaxGroups,
		IsActive:  user.IsActive,
		CreatedBy: user.CreatedBy,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户（软删除，管理员专用）
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /admin/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的用户ID", "invalid_id")
		return
	}

	userService := services.NewUserService()
	if err := userService.DeleteUser(c, uint(id)); err != nil {
		logger.Warnf("删除用户失败: %v", err)
		if err.Error() == "用户不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "user_not_found")
		} else if err.Error() == "该用户下还有分组，无法删除" {
			utils.ErrorWithErrorCode(c, 4001, err.Error(), "user_has_groups")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "删除用户失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

