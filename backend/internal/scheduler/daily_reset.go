package scheduler

import (
	"time"

	"line-management/internal/models"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"gorm.io/gorm"
)

// DailyResetTask 每日重置任务
// 每分钟检查一次，根据每个分组的reset_time判断是否需要重置
// 每个分组可以在自己的重置时间点重置，账号统计跟随所属分组的重置时间
func DailyResetTask() {
	db := database.GetDB()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	logger.Info("开始执行每日重置任务检查")

	// 获取所有分组（包括已删除的，因为统计表可能还存在）
	var groups []models.Group
	if err := db.Find(&groups).Error; err != nil {
		logger.Errorf("查询分组失败: %v", err)
		return
	}

	resetCount := 0
	accountResetCount := 0

	for _, group := range groups {
		// 解析分组的重置时间
		resetTime, err := parseResetTime(group.ResetTime)
		if err != nil {
			logger.Warnf("解析分组重置时间失败 (GroupID=%d, ResetTime=%s): %v", group.ID, group.ResetTime, err)
			continue
		}

		// 计算今天该分组的重置时间点
		todayResetTime := time.Date(
			now.Year(), now.Month(), now.Day(),
			resetTime.Hour(), resetTime.Minute(), resetTime.Second(),
			0, now.Location(),
		)

		// 检查是否需要重置：
		// 1. 当前时间已经过了今天的重置时间点
		// 2. 上次重置日期不是今天，或者上次重置时间在今天重置时间点之前
		needReset := false

		// 获取分组统计
		var groupStats models.GroupStats
		if err := db.Where("group_id = ?", group.ID).First(&groupStats).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 统计记录不存在，跳过
				continue
			}
			logger.Errorf("查询分组统计失败 (GroupID=%d): %v", group.ID, err)
			continue
		}

		if groupStats.LastResetDate == nil {
			// 从未重置过，如果当前时间已过重置时间点，则重置
			if now.After(todayResetTime) || now.Equal(todayResetTime) {
				needReset = true
			}
		} else {
			lastResetDate := time.Date(
				groupStats.LastResetDate.Year(),
				groupStats.LastResetDate.Month(),
				groupStats.LastResetDate.Day(),
				0, 0, 0, 0,
				groupStats.LastResetDate.Location(),
			)

			// 如果上次重置日期不是今天，且当前时间已过今天的重置时间点，则需要重置
			if !lastResetDate.Equal(today) {
				if now.After(todayResetTime) || now.Equal(todayResetTime) {
					needReset = true
				}
			} else {
				// 上次重置日期是今天，检查上次重置时间是否在今天重置时间点之前
				// 如果是，且当前时间已过重置时间点，则需要再次重置（防止重置时间点被修改的情况）
				if groupStats.LastResetTime != nil {
					lastResetTime := *groupStats.LastResetTime
					if lastResetTime.Before(todayResetTime) && (now.After(todayResetTime) || now.Equal(todayResetTime)) {
						needReset = true
					}
				}
			}
		}

		if needReset {
			// 重置分组统计
			updates := map[string]interface{}{
				"today_incoming":  0,
				"today_duplicate": 0,
				"last_reset_date": today,
				"last_reset_time": now,
			}

			if err := db.Model(&models.GroupStats{}).
				Where("group_id = ?", group.ID).
				Updates(updates).Error; err != nil {
				logger.Errorf("重置分组统计失败 (GroupID=%d): %v", group.ID, err)
				continue
			}

			resetCount++
			logger.Infof("已重置分组统计 (GroupID=%d, ResetTime=%s)", group.ID, group.ResetTime)

			// 重置该分组下所有账号的统计（仅重置没有独立重置时间的账号）
			var accounts []models.LineAccount
			if err := db.Where("group_id = ? AND deleted_at IS NULL AND (reset_time IS NULL OR reset_time = '')", group.ID).Find(&accounts).Error; err != nil {
				logger.Errorf("查询账号列表失败 (GroupID=%d): %v", group.ID, err)
				continue
			}

			for _, account := range accounts {
				// 跳过有独立重置时间的账号（这些账号会在后面单独处理）
				if account.ResetTime != nil && *account.ResetTime != "" {
					continue
				}

				// 获取账号统计
				var accountStats models.LineAccountStats
				if err := db.Where("line_account_id = ?", account.ID).First(&accountStats).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						// 统计记录不存在，跳过
						continue
					}
					logger.Errorf("查询账号统计失败 (LineAccountID=%d): %v", account.ID, err)
					continue
				}

				// 重置账号统计（跟随分组重置时间）
				accountUpdates := map[string]interface{}{
					"today_incoming":  0,
					"today_duplicate": 0,
					"last_reset_date": today,
					"last_reset_time": now,
				}

				if err := db.Model(&models.LineAccountStats{}).
					Where("line_account_id = ?", account.ID).
					Updates(accountUpdates).Error; err != nil {
					logger.Errorf("重置账号统计失败 (LineAccountID=%d): %v", account.ID, err)
					continue
				}

				accountResetCount++
				logger.Infof("已重置账号统计 (LineAccountID=%d, GroupID=%d, 使用分组重置时间)", account.ID, group.ID)
			}
		}
	}

	// 单独处理有独立重置时间的账号
	// 获取所有账号，检查是否有独立的重置时间
	var allAccounts []models.LineAccount
	if err := db.Where("deleted_at IS NULL AND reset_time IS NOT NULL").Find(&allAccounts).Error; err != nil {
		logger.Errorf("查询有独立重置时间的账号失败: %v", err)
	} else {
		for _, account := range allAccounts {
			// 解析账号的重置时间
			accountResetTime, err := parseResetTime(*account.ResetTime)
			if err != nil {
				logger.Warnf("解析账号重置时间失败 (LineAccountID=%d, ResetTime=%s): %v", account.ID, *account.ResetTime, err)
				continue
			}

			// 计算今天该账号的重置时间点
			todayAccountResetTime := time.Date(
				now.Year(), now.Month(), now.Day(),
				accountResetTime.Hour(), accountResetTime.Minute(), accountResetTime.Second(),
				0, now.Location(),
			)

			// 获取账号统计
			var accountStats models.LineAccountStats
			if err := db.Where("line_account_id = ?", account.ID).First(&accountStats).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					continue
				}
				logger.Errorf("查询账号统计失败 (LineAccountID=%d): %v", account.ID, err)
				continue
			}

			// 检查是否需要重置
			needAccountReset := false

			if accountStats.LastResetDate == nil {
				// 从未重置过，如果当前时间已过重置时间点，则重置
				if now.After(todayAccountResetTime) || now.Equal(todayAccountResetTime) {
					needAccountReset = true
				}
			} else {
				lastResetDate := time.Date(
					accountStats.LastResetDate.Year(),
					accountStats.LastResetDate.Month(),
					accountStats.LastResetDate.Day(),
					0, 0, 0, 0,
					accountStats.LastResetDate.Location(),
				)

				// 如果上次重置日期不是今天，且当前时间已过今天的重置时间点，则需要重置
				if !lastResetDate.Equal(today) {
					if now.After(todayAccountResetTime) || now.Equal(todayAccountResetTime) {
						needAccountReset = true
					}
				} else {
					// 上次重置日期是今天，检查上次重置时间是否在今天重置时间点之前
					if accountStats.LastResetTime != nil {
						lastResetTime := *accountStats.LastResetTime
						if lastResetTime.Before(todayAccountResetTime) && (now.After(todayAccountResetTime) || now.Equal(todayAccountResetTime)) {
							needAccountReset = true
						}
					}
				}
			}

			if needAccountReset {
				// 重置账号统计（使用账号自己的重置时间）
				accountUpdates := map[string]interface{}{
					"today_incoming":  0,
					"today_duplicate": 0,
					"last_reset_date": today,
					"last_reset_time": now,
				}

				if err := db.Model(&models.LineAccountStats{}).
					Where("line_account_id = ?", account.ID).
					Updates(accountUpdates).Error; err != nil {
					logger.Errorf("重置账号统计失败 (LineAccountID=%d): %v", account.ID, err)
					continue
				}

				accountResetCount++
				logger.Infof("已重置账号统计 (LineAccountID=%d, 使用账号独立重置时间=%s)", account.ID, *account.ResetTime)
			}
		}
	}

	if resetCount > 0 || accountResetCount > 0 {
		logger.Infof("每日重置任务完成: 重置了 %d 个分组统计, %d 个账号统计", resetCount, accountResetCount)
	}
}

// parseResetTime 解析重置时间字符串（格式：HH:MM:SS）
func parseResetTime(resetTimeStr string) (time.Time, error) {
	// 默认重置时间为 09:00:00
	if resetTimeStr == "" {
		resetTimeStr = "09:00:00"
	}

	// 解析时间字符串
	layout := "15:04:05"
	parsedTime, err := time.Parse(layout, resetTimeStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

