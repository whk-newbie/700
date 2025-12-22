package handlers

import (
	"strconv"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GetCustomers 获取客户列表
// @Summary 获取客户列表
// @Description 获取客户列表（支持分页和筛选）
// @Tags 客户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param group_id query int false "分组ID"
// @Param line_account_id query int false "Line账号ID"
// @Param platform_type query string false "平台类型" Enums(line, line_business)
// @Param customer_type query string false "客户类型"
// @Param search query string false "搜索（客户ID或显示名称）"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /customers [get]
func GetCustomers(c *gin.Context) {
	var params schemas.CustomerQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewCustomerService()
	list, total, err := service.GetCustomerList(c, &params)
	if err != nil {
		logger.Errorf("获取客户列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取客户列表失败", "internal_error")
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

// GetCustomerDetail 获取客户详情
// @Summary 获取客户详情
// @Description 获取客户详细信息
// @Tags 客户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Success 200 {object} schemas.CustomerDetailResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /customers/{id} [get]
func GetCustomerDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的客户ID", "invalid_id")
		return
	}

	service := services.NewCustomerService()
	customer, err := service.GetCustomerDetail(c, id)
	if err != nil {
		logger.Errorf("获取客户详情失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, err.Error(), "internal_error")
		return
	}

	utils.Success(c, customer)
}

// UpdateCustomer 更新客户信息
// @Summary 更新客户信息
// @Description 更新客户信息
// @Tags 客户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param request body schemas.UpdateCustomerRequest true "更新客户请求"
// @Success 200 {object} schemas.CustomerListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /customers/{id} [put]
func UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的客户ID", "invalid_id")
		return
	}

	var req schemas.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewCustomerService()
	customer, err := service.UpdateCustomer(c, id, &req)
	if err != nil {
		logger.Errorf("更新客户失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, err.Error(), "internal_error")
		return
	}

	// 转换为响应格式
	var birthday *string
	if customer.Birthday != nil {
		birthdayStr := customer.Birthday.Format("2006-01-02")
		birthday = &birthdayStr
	}

	response := schemas.CustomerListResponse{
		ID:             customer.ID,
		GroupID:        customer.GroupID,
		ActivationCode: customer.ActivationCode,
		LineAccountID:  customer.LineAccountID,
		PlatformType:   customer.PlatformType,
		CustomerID:     customer.CustomerID,
		DisplayName:    customer.DisplayName,
		AvatarURL:      customer.AvatarURL,
		PhoneNumber:    customer.PhoneNumber,
		CustomerType:   customer.CustomerType,
		Gender:         customer.Gender,
		Country:        customer.Country,
		Birthday:       birthday,
		Address:        customer.Address,
		NicknameRemark: customer.NicknameRemark,
		Remark:         customer.Remark,
		CreatedAt:      customer.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      customer.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// DeleteCustomer 删除客户
// @Summary 删除客户
// @Description 删除客户（软删除）
// @Tags 客户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /customers/{id} [delete]
func DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的客户ID", "invalid_id")
		return
	}

	service := services.NewCustomerService()
	if err := service.DeleteCustomer(c, id); err != nil {
		logger.Errorf("删除客户失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, err.Error(), "internal_error")
		return
	}

	utils.Success(c, nil)
}

