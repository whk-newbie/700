package scheduler

import (
	"time"

	"line-management/pkg/database"
	"line-management/pkg/logger"
)

// ArchiveTask 数据归档任务
// 每天凌晨4点执行，归档12个月前的数据
func ArchiveTask() {
	db := database.GetDB()
	logger.Info("开始执行数据归档任务")

	// 计算12个月前的日期
	archiveDate := time.Now().AddDate(0, -12, 0)
	archiveDateStr := archiveDate.Format("2006-01-02")

	logger.Infof("准备归档 %s 之前的数据", archiveDateStr)

	// 归档进线日志（删除12个月前的数据）
	// 注意：由于是分区表，实际上只需要删除旧分区即可
	// 但为了安全，我们只删除数据，不删除分区表结构
	result := db.Exec(`
		DELETE FROM incoming_logs 
		WHERE incoming_time < ?
	`, archiveDate)
	
	if result.Error != nil {
		logger.Errorf("归档进线日志失败: %v", result.Error)
	} else {
		logger.Infof("已归档进线日志: 删除了 %d 条记录", result.RowsAffected)
	}

	// 归档账号状态日志（删除12个月前的数据）
	result = db.Exec(`
		DELETE FROM account_status_logs 
		WHERE occurred_at < ?
	`, archiveDate)
	
	if result.Error != nil {
		logger.Errorf("归档账号状态日志失败: %v", result.Error)
	} else {
		logger.Infof("已归档账号状态日志: 删除了 %d 条记录", result.RowsAffected)
	}

	logger.Info("数据归档任务完成")
}

