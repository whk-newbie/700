package handlers

import (
	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GetContactPoolSummary 获取底库统计汇总
// @Summary 获取底库统计汇总
// @Description 获取底库统计汇总（导入数量、平台工单数量、总数量）
// @Tags 底库管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} schemas.ContactPoolSummaryResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /contact-pool/summary [get]
func GetContactPoolSummary(c *gin.Context) {
	service := services.NewContactPoolService()
	summary, err := service.GetSummary(c)
	if err != nil {
		logger.Errorf("获取底库统计汇总失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取底库统计汇总失败", "internal_error")
		return
	}

	utils.Success(c, summary)
}

// GetContactPoolList 获取底库列表（按激活码+平台）
// @Summary 获取底库列表
// @Description 获取底库列表（按激活码+平台分组）
// @Tags 底库管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param platform_type query string false "平台类型" Enums(line, line_business)
// @Param search query string false "激活码搜索"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /contact-pool/list [get]
func GetContactPoolList(c *gin.Context) {
	var params schemas.ContactPoolListQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewContactPoolService()
	list, total, err := service.GetList(c, &params)
	if err != nil {
		logger.Errorf("获取底库列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取底库列表失败", "internal_error")
		return
	}

	utils.SuccessWithPagination(c, list, params.Page, params.PageSize, total)
}

// GetContactPoolDetail 获取底库详细列表
// @Summary 获取底库详细列表
// @Description 获取底库详细列表（支持筛选和分页）
// @Tags 底库管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param activation_code query string false "激活码"
// @Param platform_type query string false "平台类型" Enums(line, line_business)
// @Param start_time query string false "开始时间" format(date-time)
// @Param end_time query string false "结束时间" format(date-time)
// @Param search query string false "搜索（用户名或手机号）"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /contact-pool/detail [get]
func GetContactPoolDetail(c *gin.Context) {
	var params schemas.ContactPoolDetailQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewContactPoolService()
	list, total, err := service.GetDetailList(c, &params)
	if err != nil {
		logger.Errorf("获取底库详细列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取底库详细列表失败", "internal_error")
		return
	}

	utils.SuccessWithPagination(c, list, params.Page, params.PageSize, total)
}

// ImportContacts 导入联系人
// @Summary 导入联系人
// @Description 从Excel/CSV/TXT文件导入联系人到底库
// @Tags 底库管理
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件（Excel/CSV/TXT）"
// @Param platform_type formData string true "平台类型" Enums(line, line_business)
// @Param dedup_scope formData string true "去重范围" Enums(current, global)
// @Param group_id formData int true "分组ID"
// @Success 200 {object} schemas.ImportContactResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /contact-pool/import [post]
func ImportContacts(c *gin.Context) {
	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请选择要上传的文件", "file_required")
		return
	}

	// 验证文件大小（最大10MB）
	if file.Size > 10*1024*1024 {
		utils.ErrorWithErrorCode(c, 1001, "文件大小不能超过10MB", "file_too_large")
		return
	}

	// 绑定表单参数
	var req schemas.ImportContactRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewContactPoolService()
	result, err := service.ImportContacts(c, file, &req)
	if err != nil {
		logger.Errorf("导入联系人失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, err.Error(), "import_failed")
		return
	}

	utils.Success(c, result)
}

// GetImportBatchList 获取导入批次列表
// @Summary 获取导入批次列表
// @Description 获取导入批次列表（支持分页和筛选）
// @Tags 底库管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param platform_type query string false "平台类型" Enums(line, line_business)
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /contact-pool/import-batches [get]
func GetImportBatchList(c *gin.Context) {
	var params schemas.ImportBatchListQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewContactPoolService()
	list, total, err := service.GetImportBatchList(c, &params)
	if err != nil {
		logger.Errorf("获取导入批次列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取导入批次列表失败", "internal_error")
		return
	}

	utils.SuccessWithPagination(c, list, params.Page, params.PageSize, total)
}

