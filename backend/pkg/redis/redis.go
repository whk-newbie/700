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

	// 创建Redis客户端
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		// 连接池配置
		PoolSize:     10,
		MinIdleConns: 5,
		// 超时配置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		// 连接保活
		PoolTimeout: 4 * time.Second,
		IdleTimeout: 5 * time.Minute,
	})

	// 测试连接
	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}

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

