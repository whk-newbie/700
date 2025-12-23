package scheduler

import (
	"line-management/pkg/database"
	"line-management/pkg/logger"
)

// PartitionManagerTask 分区自动创建任务
// 每月1号凌晨2点执行，创建下月分区
func PartitionManagerTask() {
	db := database.GetDB()
	logger.Info("开始执行分区创建任务")

	// 调用数据库函数创建下月分区
	// 使用PostgreSQL的create_next_month_partitions函数
	result := db.Exec("SELECT create_next_month_partitions()")
	if result.Error != nil {
		logger.Errorf("创建分区失败: %v", result.Error)
		return
	}

	logger.Info("分区创建任务完成")
}

