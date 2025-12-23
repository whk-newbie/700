package scheduler

import (
	"time"

	"line-management/internal/handlers"
	"line-management/internal/models"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"gorm.io/gorm"
)

// OfflineDetectionTask 离线检测任务
// 每5分钟检查一次WebSocket连接，标记超时离线的账号
func OfflineDetectionTask() {
	db := database.GetDB()
	manager := handlers.GetWebSocketManager()
	if manager == nil {
		logger.Warn("WebSocket管理器未初始化，跳过离线检测")
		return
	}

	logger.Info("开始执行离线检测任务")

	// 获取所有活跃的Windows客户端连接
	clientCount, _ := manager.GetClientCount()
	if clientCount == 0 {
		logger.Info("没有活跃的WebSocket连接，跳过离线检测")
		return
	}

	// 获取所有在线状态的账号
	var onlineAccounts []models.LineAccount
	if err := db.Where("deleted_at IS NULL AND online_status = ?", "online").Find(&onlineAccounts).Error; err != nil {
		logger.Errorf("查询在线账号失败: %v", err)
		return
	}

	offlineCount := 0
	now := time.Now()
	// 超时时间：如果账号超过5分钟没有WebSocket连接，标记为离线
	timeout := 5 * time.Minute

	for _, account := range onlineAccounts {
		// 查找该账号所属的分组
		var group models.Group
		if err := db.Where("id = ? AND deleted_at IS NULL", account.GroupID).First(&group).Error; err != nil {
			logger.Warnf("分组不存在 (GroupID=%d): %v", account.GroupID, err)
			continue
		}

		// 检查是否有该激活码的WebSocket连接
		clients := manager.GetClientsByActivationCode(group.ActivationCode)
		hasActiveConnection := false

		for _, client := range clients {
			// 检查连接是否活跃（最后心跳时间在超时范围内）
			if now.Sub(client.LastHeartbeat) < timeout {
				hasActiveConnection = true
				break
			}
		}

		// 如果没有活跃连接，且账号状态为online，标记为异常离线
		if !hasActiveConnection {
			// 检查账号的最后活跃时间
			if account.LastActiveAt != nil {
				// 如果最后活跃时间超过超时时间，标记为离线
				if now.Sub(*account.LastActiveAt) > timeout {
					account.OnlineStatus = "abnormal_offline"
					nowTime := time.Now()
					account.LastActiveAt = &nowTime

					if err := db.Save(&account).Error; err != nil {
						logger.Errorf("更新账号状态失败 (LineAccountID=%d): %v", account.ID, err)
						continue
					}

					offlineCount++
					logger.Infof("账号已标记为异常离线 (LineAccountID=%d, LineID=%s)", account.ID, account.LineID)

					// 更新分组统计中的在线账号数
					updateGroupOnlineCount(db, account.GroupID)
				}
			} else {
				// 如果没有最后活跃时间，且没有活跃连接，标记为离线
				account.OnlineStatus = "abnormal_offline"
				nowTime := time.Now()
				account.LastActiveAt = &nowTime

				if err := db.Save(&account).Error; err != nil {
					logger.Errorf("更新账号状态失败 (LineAccountID=%d): %v", account.ID, err)
					continue
				}

				offlineCount++
				logger.Infof("账号已标记为异常离线 (LineAccountID=%d, LineID=%s)", account.ID, account.LineID)

				// 更新分组统计中的在线账号数
				updateGroupOnlineCount(db, account.GroupID)
			}
		}
	}

	if offlineCount > 0 {
		logger.Infof("离线检测任务完成: 标记了 %d 个账号为异常离线", offlineCount)
	} else {
		logger.Info("离线检测任务完成: 所有账号连接正常")
	}
}

// updateGroupOnlineCount 更新分组统计中的在线账号数
func updateGroupOnlineCount(db *gorm.DB, groupID uint) {
	var onlineCount int64
	db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL AND online_status = ?", groupID, "online").
		Count(&onlineCount)

	db.Model(&models.GroupStats{}).
		Where("group_id = ?", groupID).
		Update("online_accounts", onlineCount)
}

