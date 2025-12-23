package scheduler

import (
	"line-management/pkg/logger"

	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron *cron.Cron
}

// NewScheduler 创建调度器实例
func NewScheduler() *Scheduler {
	// 使用秒级精度（支持秒级任务）
	c := cron.New(cron.WithSeconds())
	return &Scheduler{
		cron: c,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	// 注册所有定时任务
	s.registerTasks()
	
	// 启动调度器
	s.cron.Start()
	logger.Info("定时任务调度器已启动")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("定时任务调度器已停止")
}

// registerTasks 注册所有定时任务
func (s *Scheduler) registerTasks() {
	// 1. 每日重置任务 - 每分钟检查一次
	_, err := s.cron.AddFunc("0 * * * * *", DailyResetTask)
	if err != nil {
		logger.Errorf("注册每日重置任务失败: %v", err)
	} else {
		logger.Info("每日重置任务已注册（每分钟检查）")
	}

	// 2. 全量校准任务 - 每天凌晨3点执行
	_, err = s.cron.AddFunc("0 0 3 * * *", StatsCalibrationTask)
	if err != nil {
		logger.Errorf("注册全量校准任务失败: %v", err)
	} else {
		logger.Info("全量校准任务已注册（每天凌晨3点）")
	}

	// 3. 离线检测任务 - 每5分钟检查一次
	_, err = s.cron.AddFunc("0 */5 * * * *", OfflineDetectionTask)
	if err != nil {
		logger.Errorf("注册离线检测任务失败: %v", err)
	} else {
		logger.Info("离线检测任务已注册（每5分钟检查）")
	}

	// 4. 分区自动创建任务 - 每月1号凌晨2点执行
	_, err = s.cron.AddFunc("0 0 2 1 * *", PartitionManagerTask)
	if err != nil {
		logger.Errorf("注册分区创建任务失败: %v", err)
	} else {
		logger.Info("分区创建任务已注册（每月1号凌晨2点）")
	}

	// 5. 数据归档任务 - 每天凌晨4点执行
	_, err = s.cron.AddFunc("0 0 4 * * *", ArchiveTask)
	if err != nil {
		logger.Errorf("注册数据归档任务失败: %v", err)
	} else {
		logger.Info("数据归档任务已注册（每天凌晨4点）")
	}
}

