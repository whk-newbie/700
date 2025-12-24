package database

import (
	"database/sql"
	"fmt"
	"time"

	"line-management/internal/config"
	"line-management/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	cfg := config.GlobalConfig.Database

	// 调试输出（生产环境应移除）
	logger.Infof("数据库配置: host=%s, user=%s, dbname=%s, port=%d", cfg.Host, cfg.User, cfg.DBName, cfg.Port)

	// 构建DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		cfg.SSLMode,
		cfg.TimeZone,
	)

	// 配置GORM日志
	gormLogLevel := gormLogger.Default
	if config.GlobalConfig.Log.Level == "debug" {
		gormLogLevel = gormLogger.Default.LogMode(gormLogger.Info)
	}

	// 连接数据库
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogLevel,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		// 启用准备语句缓存以提高性能
		PrepareStmt: true,
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层sql.DB以配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 翻倍当前连接池配置
	maxOpenConns := 200  // 原100翻倍
	maxIdleConns := 20   // 原10翻倍

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(maxIdleConns)                    // 最大空闲连接数
	sqlDB.SetMaxOpenConns(maxOpenConns)                   // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour * 2)              // 连接最大生存时间 (2小时)
	sqlDB.SetConnMaxIdleTime(time.Minute * 30)           // 连接最大空闲时间 (30分钟)

	logger.Infof("数据库连接池配置 - MaxOpenConns:%d, MaxIdleConns:%d",
		maxOpenConns, maxIdleConns)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 启动连接池监控
	go monitorConnectionPool(sqlDB)

	logger.Info("数据库连接成功")
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// monitorConnectionPool 监控连接池状态
func monitorConnectionPool(sqlDB *sql.DB) {
	ticker := time.NewTicker(time.Minute * 5) // 每5分钟检查一次
	defer ticker.Stop()

	for range ticker.C {
		stats := sqlDB.Stats()
		logger.Debugf("数据库连接池状态 - 打开连接:%d, 空闲连接:%d, 使用中:%d, 等待:%d",
			stats.OpenConnections,
			stats.Idle,
			stats.InUse,
			stats.WaitCount,
		)

		// 如果连接池使用率过高，记录警告
		if stats.InUse > stats.MaxOpenConnections*8/10 {
			logger.Warnf("数据库连接池使用率过高 - 使用中:%d/%d", stats.InUse, stats.MaxOpenConnections)
		}
	}
}

// GetConnectionStats 获取连接池统计信息
func GetConnectionStats() sql.DBStats {
	if DB != nil {
		if sqlDB, err := DB.DB(); err == nil {
			return sqlDB.Stats()
		}
	}
	return sql.DBStats{}
}

// HealthCheck 执行数据库健康检查
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 执行简单查询检查连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接检查失败: %w", err)
	}

	return nil
}

