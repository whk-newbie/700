package handlers

import (
	"strconv"

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
// @Success 200 {object} utils.Response{data=models.GroupStats}
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
// @Success 200 {object} utils.Response{data=models.LineAccountStats}
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
	stats, err := statsService.GetOverviewStats()
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

