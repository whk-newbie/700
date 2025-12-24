package redis

import (
	"context"
	"fmt"
	"time"

	"line-management/internal/config"
	"line-management/pkg/logger"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client
var ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis() error {
	cfg := config.GlobalConfig.Redis

	// 翻倍当前连接池配置
	poolSize := 20     // 原10翻倍
	minIdleConns := 10 // 原5翻倍

	// 创建Redis客户端
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		// 优化连接池配置
		PoolSize:     poolSize,           // 连接池大小
		MinIdleConns: minIdleConns,       // 最小空闲连接数
		// 超时配置
		DialTimeout:  10 * time.Second,   // 连接超时 (增加到10秒)
		ReadTimeout:  5 * time.Second,    // 读取超时 (增加到5秒)
		WriteTimeout: 5 * time.Second,    // 写入超时 (增加到5秒)
		// 连接保活
		PoolTimeout: 8 * time.Second,     // 池超时 (增加到8秒)
		IdleTimeout: 10 * time.Minute,    // 空闲超时 (增加到10分钟)
		// 连接检查
		IdleCheckFrequency: 60 * time.Second, // 空闲连接检查频率
		// 最大重试次数
		MaxRetries: 3,
		// 重试间隔
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 2 * time.Second,
	})

	logger.Infof("Redis连接池配置 - PoolSize:%d, MinIdleConns:%d",
		poolSize, minIdleConns)

	// 测试连接
	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}

	// 启动连接池监控
	go monitorRedisPool()

	logger.Info("Redis连接成功")
	return nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// GetClient 获取Redis客户端
func GetClient() *redis.Client {
	return Client
}

// GetContext 获取上下文
func GetContext() context.Context {
	return ctx
}

// monitorRedisPool 监控Redis连接池状态
func monitorRedisPool() {
	ticker := time.NewTicker(time.Minute * 5) // 每5分钟检查一次
	defer ticker.Stop()

	for range ticker.C {
		if Client == nil {
			continue
		}

		poolStats := Client.PoolStats()
		logger.Debugf("Redis连接池状态 - 总连接:%d, 空闲连接:%d, 等待中:%d, 命中:%d, 超时:%d, 总数:%d",
			poolStats.TotalConns,
			poolStats.IdleConns,
			poolStats.StaleConns,
			poolStats.Hits,
			poolStats.Timeouts,
			poolStats.TotalConns,
		)

		// 如果连接池使用率过高，记录警告
		if int(poolStats.TotalConns) > int(Client.Options().PoolSize)*8/10 {
			logger.Warnf("Redis连接池使用率过高 - 总连接:%d/%d", poolStats.TotalConns, Client.Options().PoolSize)
		}
	}
}

// GetPoolStats 获取Redis连接池统计信息
func GetPoolStats() *redis.PoolStats {
	if Client != nil {
		return Client.PoolStats()
	}
	return &redis.PoolStats{}
}

// HealthCheck 执行Redis健康检查
func HealthCheck() error {
	if Client == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	// 执行简单PING检查连接
	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接检查失败: %w", err)
	}

	return nil
}

// GetPoolConfig 获取连接池配置信息
func GetPoolConfig() map[string]interface{} {
	if Client == nil {
		return nil
	}

	opts := Client.Options()
	return map[string]interface{}{
		"addr":              opts.Addr,
		"db":                opts.DB,
		"pool_size":         opts.PoolSize,
		"min_idle_conns":    opts.MinIdleConns,
		"dial_timeout":      opts.DialTimeout,
		"read_timeout":      opts.ReadTimeout,
		"write_timeout":     opts.WriteTimeout,
		"pool_timeout":      opts.PoolTimeout,
		"idle_timeout":      opts.IdleTimeout,
		"idle_check_freq":   opts.IdleCheckFrequency,
		"max_retries":       opts.MaxRetries,
		"min_retry_backoff": opts.MinRetryBackoff,
		"max_retry_backoff": opts.MaxRetryBackoff,
	}
}

