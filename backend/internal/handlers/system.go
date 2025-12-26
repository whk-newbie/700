package handlers

import (
	"line-management/internal/utils"
	"line-management/pkg/database"
	"line-management/pkg/redis"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查系统各组件的健康状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=HealthStatus}
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0",
		Uptime:    0, // TODO: 可以添加启动时间统计
	}

	// 检查数据库连接
	dbStats := database.GetConnectionStats()
	status.Database = DatabaseHealth{
		Status: "ok",
		Stats: ConnectionPoolStats{
			OpenConnections:     dbStats.OpenConnections,
			InUse:              dbStats.InUse,
			Idle:               dbStats.Idle,
			WaitCount:          dbStats.WaitCount,
			WaitDuration:       dbStats.WaitDuration.String(),
			MaxIdleClosed:      dbStats.MaxIdleClosed,
			MaxIdleTimeClosed:  dbStats.MaxIdleTimeClosed,
			MaxLifetimeClosed:  dbStats.MaxLifetimeClosed,
		},
	}

	// 执行数据库健康检查
	if err := database.HealthCheck(); err != nil {
		status.Database.Status = "error"
		status.Database.Error = err.Error()
		status.Status = "degraded"
	}

	// 检查Redis连接
	redisStats := redis.GetPoolStats()
	status.Redis = RedisHealth{
		Status: "ok",
		Stats: RedisPoolStats{
			TotalConns: int64(redisStats.TotalConns),
			IdleConns:  int64(redisStats.IdleConns),
			StaleConns: int64(redisStats.StaleConns),
			Hits:       int64(redisStats.Hits),
			Misses:     int64(redisStats.Misses),
			Timeouts:   int64(redisStats.Timeouts),
		},
		Config: redis.GetPoolConfig(),
	}

	// 执行Redis健康检查
	if err := redis.HealthCheck(); err != nil {
		status.Redis.Status = "error"
		status.Redis.Error = err.Error()
		status.Status = "degraded"
	}

	// 如果两个服务都异常，标记为错误
	if status.Database.Status == "error" && status.Redis.Status == "error" {
		status.Status = "error"
	}

	utils.Success(c, status)
}

// 健康状态结构体
type HealthStatus struct {
	Status    string         `json:"status"`    // ok, degraded, error
	Timestamp int64          `json:"timestamp"`
	Version   string         `json:"version"`
	Uptime    int64          `json:"uptime"`
	Database  DatabaseHealth `json:"database"`
	Redis     RedisHealth    `json:"redis"`
}

type DatabaseHealth struct {
	Status string             `json:"status"`
	Error  string             `json:"error,omitempty"`
	Stats  ConnectionPoolStats `json:"stats"`
}

type RedisHealth struct {
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
	Stats  RedisPoolStats `json:"stats"`
	Config map[string]interface{} `json:"config"`
}

// 数据库连接池统计
type ConnectionPoolStats struct {
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       string        `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed  int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
}

// Redis连接池统计
type RedisPoolStats struct {
	TotalConns int64 `json:"total_conns"`
	IdleConns  int64 `json:"idle_conns"`
	StaleConns int64 `json:"stale_conns"`
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Timeouts   int64 `json:"timeouts"`
}
