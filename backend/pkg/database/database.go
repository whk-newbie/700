package database

import (
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
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层sql.DB以配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)                // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)       // 连接最大生存时间
	sqlDB.SetConnMaxIdleTime(time.Minute * 10) // 连接最大空闲时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

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

