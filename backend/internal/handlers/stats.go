package handlers

import (
	"strconv"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetGroupStats 获取分组统计
// @Summary 获取分组统计
// @Description 获取指定分组的统计数据
// @Tags 统计
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} utils.Response{data=object}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /stats/group/{id} [get]
// @Security BearerAuth
func GetGroupStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, 1001, "无效的分组ID")
		return
	}

	statsService := services.NewStatsService()
	stats, err := statsService.GetGroupStats(uint(id))
	if err != nil {
		utils.Error(c, 5001, "获取分组统计失败: "+err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetAccountStats 获取账号统计
// @Summary 获取账号统计
// @Description 获取指定Line账号的统计数据
// @Tags 统计
// @Accept json
// @Produce json
// @Param id path int true "账号ID"
// @Success 200 {object} utils.Response{data=object}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /stats/account/{id} [get]
// @Security BearerAuth
func GetAccountStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, 1001, "无效的账号ID")
		return
	}

	statsService := services.NewStatsService()
	stats, err := statsService.GetAccountStats(uint(id))
	if err != nil {
		utils.Error(c, 5001, "获取账号统计失败: "+err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetOverviewStats 获取总览统计
// @Summary 获取总览统计
// @Description 获取系统总览统计数据
// @Tags 统计
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=object}
// @Failure 500 {object} utils.Response
// @Router /stats/overview [get]
// @Security BearerAuth
func GetOverviewStats(c *gin.Context) {
	statsService := services.NewStatsService()
	stats, err := statsService.GetOverviewStats(c)
	if err != nil {
		utils.Error(c, 5001, "获取总览统计失败: "+err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetGroupIncomingTrend 获取分组进线趋势
// @Summary 获取分组进线趋势
// @Description 获取指定分组最近N天的进线趋势数据
// @Tags 统计
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Param days query int false "天数（默认7天，最多30天）" default(7)
// @Success 200 {object} utils.Response{data=[]object}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /stats/group/{id}/trend [get]
// @Security BearerAuth
func GetGroupIncomingTrend(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, 1001, "无效的分组ID")
		return
	}

	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	statsService := services.NewStatsService()
	trend, err := statsService.GetGroupIncomingTrend(uint(id), days)
	if err != nil {
		utils.Error(c, 5001, "获取分组进线趋势失败: "+err.Error())
		return
	}

	utils.Success(c, trend)
}

// GetAccountIncomingTrend 获取账号进线趋势
// @Summary 获取账号进线趋势
// @Description 获取指定账号最近N天的进线趋势数据
// @Tags 统计
// @Accept json
// @Produce json
// @Param id path int true "账号ID"
// @Param days query int false "天数（默认7天，最多30天）" default(7)
// @Success 200 {object} utils.Response{data=[]object}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /stats/account/{id}/trend [get]
// @Security BearerAuth
func GetAccountIncomingTrend(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, 1001, "无效的账号ID")
		return
	}

	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	statsService := services.NewStatsService()
	trend, err := statsService.GetAccountIncomingTrend(uint(id), days)
	if err != nil {
		utils.Error(c, 5001, "获取账号进线趋势失败: "+err.Error())
		return
	}

	utils.Success(c, trend)
}

// GetIncomingLogs 获取进线日志列表
// @Summary 获取进线日志列表
// @Description 获取进线日志列表（支持分页和筛选）
// @Tags 统计
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param group_id query int false "分组ID"
// @Param line_account_id query int false "账号ID"
// @Param is_duplicate query bool false "是否重复"
// @Param start_time query string false "开始时间（ISO 8601格式）"
// @Param end_time query string false "结束时间（ISO 8601格式）"
// @Param search query string false "搜索（进线Line ID或显示名称）"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /stats/incoming-logs [get]
// @Security BearerAuth
func GetIncomingLogs(c *gin.Context) {
	var params schemas.IncomingLogQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	incomingService := services.NewIncomingService(nil)
	list, total, err := incomingService.GetIncomingLogList(c, &params)
	if err != nil {
		utils.ErrorWithErrorCode(c, 5001, "获取进线日志列表失败", "internal_error")
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

