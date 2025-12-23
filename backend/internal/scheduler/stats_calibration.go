package scheduler

import (
	"time"

	"line-management/internal/models"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"gorm.io/gorm"
)

// StatsCalibrationTask 全量校准任务
// 每天凌晨3点执行，重算所有统计数据，防止数据漂移
func StatsCalibrationTask() {
	db := database.GetDB()
	logger.Info("开始执行全量校准任务")

	// 1. 校准分组统计
	if err := calibrateGroupStats(db); err != nil {
		logger.Errorf("校准分组统计失败: %v", err)
	} else {
		logger.Info("分组统计校准完成")
	}

	// 2. 校准账号统计
	if err := calibrateAccountStats(db); err != nil {
		logger.Errorf("校准账号统计失败: %v", err)
	} else {
		logger.Info("账号统计校准完成")
	}

	logger.Info("全量校准任务完成")
}

// calibrateGroupStats 校准分组统计
func calibrateGroupStats(db *gorm.DB) error {
	// 获取所有分组
	var groups []models.Group
	if err := db.Where("deleted_at IS NULL").Find(&groups).Error; err != nil {
		return err
	}

	for _, group := range groups {
		// 计算实际统计数据
		var stats struct {
			TotalIncoming     int64
			TodayIncoming     int64
			DuplicateIncoming int64
			TodayDuplicate    int64
			TotalAccounts     int64
			OnlineAccounts    int64
			LineAccounts      int64
			LineBusinessAccounts int64
		}

		// 计算总进线数
		db.Model(&models.IncomingLog{}).
			Where("group_id = ?", group.ID).
			Count(&stats.TotalIncoming)

		// 计算今日进线数
		today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
		db.Model(&models.IncomingLog{}).
			Where("group_id = ? AND incoming_time >= ?", group.ID, today).
			Count(&stats.TodayIncoming)

		// 计算重复进线数
		db.Model(&models.IncomingLog{}).
			Where("group_id = ? AND is_duplicate = true", group.ID).
			Count(&stats.DuplicateIncoming)

		// 计算今日重复数
		db.Model(&models.IncomingLog{}).
			Where("group_id = ? AND is_duplicate = true AND incoming_time >= ?", group.ID, today).
			Count(&stats.TodayDuplicate)

		// 计算账号统计
		db.Model(&models.LineAccount{}).
			Where("group_id = ? AND deleted_at IS NULL", group.ID).
			Count(&stats.TotalAccounts)

		db.Model(&models.LineAccount{}).
			Where("group_id = ? AND deleted_at IS NULL AND online_status = ?", group.ID, "online").
			Count(&stats.OnlineAccounts)

		db.Model(&models.LineAccount{}).
			Where("group_id = ? AND deleted_at IS NULL AND platform_type = ?", group.ID, "line").
			Count(&stats.LineAccounts)

		db.Model(&models.LineAccount{}).
			Where("group_id = ? AND deleted_at IS NULL AND platform_type = ?", group.ID, "line_business").
			Count(&stats.LineBusinessAccounts)

		// 更新分组统计
		now := time.Now()
		todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		groupStats := models.GroupStats{
			GroupID:             group.ID,
			TotalIncoming:       int(stats.TotalIncoming),
			TodayIncoming:       int(stats.TodayIncoming),
			DuplicateIncoming:   int(stats.DuplicateIncoming),
			TodayDuplicate:      int(stats.TodayDuplicate),
			TotalAccounts:       int(stats.TotalAccounts),
			OnlineAccounts:      int(stats.OnlineAccounts),
			LineAccounts:        int(stats.LineAccounts),
			LineBusinessAccounts: int(stats.LineBusinessAccounts),
			LastResetDate:       &todayDate,
			LastResetTime:       &now,
		}

		// 使用Save方法，如果不存在则创建，存在则更新
		var existingStats models.GroupStats
		if err := db.Where("group_id = ?", group.ID).First(&existingStats).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建新记录
				if err := db.Create(&groupStats).Error; err != nil {
					logger.Errorf("创建分组统计失败 (GroupID=%d): %v", group.ID, err)
					continue
				}
			} else {
				logger.Errorf("查询分组统计失败 (GroupID=%d): %v", group.ID, err)
				continue
			}
		} else {
			// 更新现有记录
			groupStats.ID = existingStats.ID
			if err := db.Save(&groupStats).Error; err != nil {
				logger.Errorf("更新分组统计失败 (GroupID=%d): %v", group.ID, err)
				continue
			}
		}

		logger.Infof("已校准分组统计 (GroupID=%d)", group.ID)
	}

	return nil
}

// calibrateAccountStats 校准账号统计
func calibrateAccountStats(db *gorm.DB) error {
	// 获取所有账号
	var accounts []models.LineAccount
	if err := db.Where("deleted_at IS NULL").Find(&accounts).Error; err != nil {
		return err
	}

	for _, account := range accounts {
		// 计算实际统计数据
		var stats struct {
			TotalIncoming     int64
			TodayIncoming     int64
			DuplicateIncoming int64
			TodayDuplicate    int64
		}

		// 计算总进线数
		db.Model(&models.IncomingLog{}).
			Where("line_account_id = ?", account.ID).
			Count(&stats.TotalIncoming)

		// 计算今日进线数
		today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
		db.Model(&models.IncomingLog{}).
			Where("line_account_id = ? AND incoming_time >= ?", account.ID, today).
			Count(&stats.TodayIncoming)

		// 计算重复进线数
		db.Model(&models.IncomingLog{}).
			Where("line_account_id = ? AND is_duplicate = true", account.ID).
			Count(&stats.DuplicateIncoming)

		// 计算今日重复数
		db.Model(&models.IncomingLog{}).
			Where("line_account_id = ? AND is_duplicate = true AND incoming_time >= ?", account.ID, today).
			Count(&stats.TodayDuplicate)

		// 更新账号统计
		now := time.Now()
		todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		accountStats := models.LineAccountStats{
			LineAccountID:     account.ID,
			TotalIncoming:     int(stats.TotalIncoming),
			TodayIncoming:     int(stats.TodayIncoming),
			DuplicateIncoming: int(stats.DuplicateIncoming),
			TodayDuplicate:    int(stats.TodayDuplicate),
			LastResetDate:     &todayDate,
			LastResetTime:     &now,
		}

		// 使用Save方法，如果不存在则创建，存在则更新
		var existingStats models.LineAccountStats
		if err := db.Where("line_account_id = ?", account.ID).First(&existingStats).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建新记录
				if err := db.Create(&accountStats).Error; err != nil {
					logger.Errorf("创建账号统计失败 (LineAccountID=%d): %v", account.ID, err)
					continue
				}
			} else {
				logger.Errorf("查询账号统计失败 (LineAccountID=%d): %v", account.ID, err)
				continue
			}
		} else {
			// 更新现有记录
			accountStats.ID = existingStats.ID
			if err := db.Save(&accountStats).Error; err != nil {
				logger.Errorf("更新账号统计失败 (LineAccountID=%d): %v", account.ID, err)
				continue
			}
		}

		logger.Infof("已校准账号统计 (LineAccountID=%d)", account.ID)
	}

	return nil
}

