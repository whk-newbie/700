package handlers

import (
	"strconv"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GetFollowUps 获取跟进记录列表
// @Summary 获取跟进记录列表
// @Description 获取跟进记录列表（支持分页和筛选）
// @Tags 跟进记录
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param group_id query int false "分组ID"
// @Param line_account_id query int false "Line账号ID"
// @Param customer_id query int false "客户ID"
// @Param platform_type query string false "平台类型" Enums(line, line_business)
// @Param search query string false "搜索（跟进内容）"
// @Param start_time query string false "开始时间" format(date-time)
// @Param end_time query string false "结束时间" format(date-time)
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /follow-ups [get]
func GetFollowUps(c *gin.Context) {
	var params schemas.FollowUpQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewFollowUpService()
	list, total, err := service.GetFollowUpList(c, &params)
	if err != nil {
		logger.Errorf("获取跟进记录列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取跟进记录列表失败", "internal_error")
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

// CreateFollowUp 创建跟进记录
// @Summary 创建跟进记录
// @Description 创建新的跟进记录
// @Tags 跟进记录
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.CreateFollowUpRequest true "创建跟进记录请求"
// @Success 200 {object} schemas.FollowUpListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /follow-ups [post]
func CreateFollowUp(c *gin.Context) {
	var req schemas.CreateFollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewFollowUpService()
	record, err := service.CreateFollowUp(c, &req)
	if err != nil {
		logger.Errorf("创建跟进记录失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, err.Error(), "internal_error")
		return
	}

	// 转换为响应格式
	response := schemas.FollowUpListResponse{
		ID:                     record.ID,
		GroupID:                record.GroupID,
		ActivationCode:         record.ActivationCode,
		LineAccountID:          record.LineAccountID,
		CustomerID:             record.CustomerID,
		PlatformType:           record.PlatformType,
		LineAccountDisplayName: record.LineAccountDisplayName,
		LineAccountLineID:      record.LineAccountLineID,
		LineAccountAvatarURL:   record.LineAccountAvatarURL,
		CustomerDisplayName:    record.CustomerDisplayName,
		CustomerLineID:         record.CustomerLineID,
		CustomerAvatarURL:      record.CustomerAvatarURL,
		Content:                record.Content,
		CreatedBy:              record.CreatedBy,
		CreatedAt:              record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:              record.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// UpdateFollowUp 更新跟进记录
// @Summary 更新跟进记录
// @Description 更新跟进记录内容
// @Tags 跟进记录
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "跟进记录ID"
// @Param request body schemas.UpdateFollowUpRequest true "更新跟进记录请求"
// @Success 200 {object} schemas.FollowUpListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /follow-ups/{id} [put]
func UpdateFollowUp(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的跟进记录ID", "invalid_id")
		return
	}

	var req schemas.UpdateFollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewFollowUpService()
	record, err := service.UpdateFollowUp(c, id, &req)
	if err != nil {
		logger.Errorf("更新跟进记录失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, err.Error(), "internal_error")
		return
	}

	// 转换为响应格式
	response := schemas.FollowUpListResponse{
		ID:                     record.ID,
		GroupID:                record.GroupID,
		ActivationCode:         record.ActivationCode,
		LineAccountID:          record.LineAccountID,
		CustomerID:             record.CustomerID,
		PlatformType:           record.PlatformType,
		LineAccountDisplayName: record.LineAccountDisplayName,
		LineAccountLineID:      record.LineAccountLineID,
		LineAccountAvatarURL:   record.LineAccountAvatarURL,
		CustomerDisplayName:    record.CustomerDisplayName,
		CustomerLineID:         record.CustomerLineID,
		CustomerAvatarURL:      record.CustomerAvatarURL,
		Content:                record.Content,
		CreatedBy:              record.CreatedBy,
		CreatedAt:              record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:              record.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// DeleteFollowUp 删除跟进记录
// @Summary 删除跟进记录
// @Description 删除跟进记录（软删除）
// @Tags 跟进记录
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "跟进记录ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /follow-ups/{id} [delete]
func DeleteFollowUp(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的跟进记录ID", "invalid_id")
		return
	}

	service := services.NewFollowUpService()
	if err := service.DeleteFollowUp(c, id); err != nil {
		logger.Errorf("删除跟进记录失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, err.Error(), "internal_error")
		return
	}

	utils.Success(c, nil)
}

// BatchCreateFollowUp 批量创建跟进记录
// @Summary 批量创建跟进记录
// @Description 批量创建跟进记录
// @Tags 跟进记录
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.BatchCreateFollowUpRequest true "批量创建跟进记录请求"
// @Success 200 {array} schemas.FollowUpListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /follow-ups/batch [post]
func BatchCreateFollowUp(c *gin.Context) {
	var req schemas.BatchCreateFollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	service := services.NewFollowUpService()
	records, err := service.BatchCreateFollowUp(c, &req)
	if err != nil {
		logger.Errorf("批量创建跟进记录失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, err.Error(), "internal_error")
		return
	}

	// 转换为响应格式
	response := make([]schemas.FollowUpListResponse, 0, len(records))
	for _, record := range records {
		response = append(response, schemas.FollowUpListResponse{
			ID:                     record.ID,
			GroupID:                record.GroupID,
			ActivationCode:         record.ActivationCode,
			LineAccountID:          record.LineAccountID,
			CustomerID:             record.CustomerID,
			PlatformType:           record.PlatformType,
			LineAccountDisplayName: record.LineAccountDisplayName,
			LineAccountLineID:      record.LineAccountLineID,
			LineAccountAvatarURL:   record.LineAccountAvatarURL,
			CustomerDisplayName:    record.CustomerDisplayName,
			CustomerLineID:         record.CustomerLineID,
			CustomerAvatarURL:      record.CustomerAvatarURL,
			Content:                record.Content,
			CreatedBy:              record.CreatedBy,
			CreatedAt:              record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:              record.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	utils.Success(c, response)
}

