package services

import (
	"time"

	"line-management/internal/models"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"gorm.io/gorm"
)

// StatsService 统计服务
type StatsService struct {
	db *gorm.DB
}

// NewStatsService 创建统计服务实例
func NewStatsService() *StatsService {
	return &StatsService{
		db: database.GetDB(),
	}
}

// GetGroupStats 获取分组统计
func (s *StatsService) GetGroupStats(groupID uint) (*models.GroupStats, error) {
	var stats models.GroupStats
	
	err := s.db.Where("group_id = ?", groupID).First(&stats).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果统计记录不存在，创建一个空的统计记录
			stats = models.GroupStats{
				GroupID: groupID,
			}
			if err := s.db.Create(&stats).Error; err != nil {
				logger.Errorf("创建分组统计记录失败: %v", err)
				return nil, err
			}
		} else {
			logger.Errorf("获取分组统计失败: %v", err)
			return nil, err
		}
	}
	
	// 加载关联的分组信息
	s.db.Preload("Group").First(&stats, stats.ID)
	
	return &stats, nil
}

// GetAccountStats 获取账号统计
func (s *StatsService) GetAccountStats(accountID uint) (*models.LineAccountStats, error) {
	var stats models.LineAccountStats
	
	err := s.db.Where("line_account_id = ?", accountID).First(&stats).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果统计记录不存在，创建一个空的统计记录
			stats = models.LineAccountStats{
				LineAccountID: accountID,
			}
			if err := s.db.Create(&stats).Error; err != nil {
				logger.Errorf("创建账号统计记录失败: %v", err)
				return nil, err
			}
		} else {
			logger.Errorf("获取账号统计失败: %v", err)
			return nil, err
		}
	}
	
	// 加载关联的账号信息
	s.db.Preload("LineAccount").First(&stats, stats.ID)
	
	return &stats, nil
}

// GetOverviewStats 获取总览统计
func (s *StatsService) GetOverviewStats() (map[string]interface{}, error) {
	var result map[string]interface{} = make(map[string]interface{})
	
	// 总分组数
	var totalGroups int64
	s.db.Model(&models.Group{}).Where("deleted_at IS NULL").Count(&totalGroups)
	result["total_groups"] = totalGroups
	
	// 总账号数
	var totalAccounts int64
	s.db.Model(&models.LineAccount{}).Where("deleted_at IS NULL").Count(&totalAccounts)
	result["total_accounts"] = totalAccounts
	
	// 在线账号数
	var onlineAccounts int64
	s.db.Model(&models.LineAccount{}).Where("deleted_at IS NULL AND online_status = ?", "online").Count(&onlineAccounts)
	result["online_accounts"] = onlineAccounts
	
	// 总进线数（从所有分组统计汇总）
	var totalIncoming int64
	s.db.Model(&models.GroupStats{}).Select("COALESCE(SUM(total_incoming), 0)").Scan(&totalIncoming)
	result["total_incoming"] = totalIncoming
	
	// 今日进线数（从所有分组统计汇总）
	var todayIncoming int64
	s.db.Model(&models.GroupStats{}).Select("COALESCE(SUM(today_incoming), 0)").Scan(&todayIncoming)
	result["today_incoming"] = todayIncoming
	
	// 总重复数（从所有分组统计汇总）
	var totalDuplicate int64
	s.db.Model(&models.GroupStats{}).Select("COALESCE(SUM(duplicate_incoming), 0)").Scan(&totalDuplicate)
	result["total_duplicate"] = totalDuplicate
	
	// 今日重复数（从所有分组统计汇总）
	var todayDuplicate int64
	s.db.Model(&models.GroupStats{}).Select("COALESCE(SUM(today_duplicate), 0)").Scan(&todayDuplicate)
	result["today_duplicate"] = todayDuplicate
	
	// 底库总数
	var totalContacts int64
	s.db.Model(&models.ContactPool{}).Where("deleted_at IS NULL").Count(&totalContacts)
	result["total_contacts"] = totalContacts
	
	return result, nil
}

// GetGroupIncomingTrend 获取分组进线趋势（最近N天）
func (s *StatsService) GetGroupIncomingTrend(groupID uint, days int) ([]map[string]interface{}, error) {
	if days <= 0 {
		days = 7 // 默认7天
	}
	if days > 30 {
		days = 30 // 最多30天
	}
	
	var results []map[string]interface{}
	
	// 查询最近N天的进线数据
	startDate := time.Now().AddDate(0, 0, -days)
	
	rows, err := s.db.Model(&models.IncomingLog{}).
		Select("DATE(incoming_time) as date, COUNT(*) as count, COUNT(CASE WHEN is_duplicate = true THEN 1 END) as duplicate_count").
		Where("group_id = ? AND incoming_time >= ?", groupID, startDate).
		Group("DATE(incoming_time)").
		Order("date ASC").
		Rows()
	
	if err != nil {
		logger.Errorf("查询分组进线趋势失败: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var date time.Time
		var count, duplicateCount int64
		
		if err := rows.Scan(&date, &count, &duplicateCount); err != nil {
			logger.Errorf("扫描趋势数据失败: %v", err)
			continue
		}
		
		results = append(results, map[string]interface{}{
			"date":           date.Format("2006-01-02"),
			"count":          count,
			"duplicate_count": duplicateCount,
			"unique_count":   count - duplicateCount,
		})
	}
	
	return results, nil
}

// GetAccountIncomingTrend 获取账号进线趋势（最近N天）
func (s *StatsService) GetAccountIncomingTrend(accountID uint, days int) ([]map[string]interface{}, error) {
	if days <= 0 {
		days = 7 // 默认7天
	}
	if days > 30 {
		days = 30 // 最多30天
	}
	
	var results []map[string]interface{}
	
	// 查询最近N天的进线数据
	startDate := time.Now().AddDate(0, 0, -days)
	
	rows, err := s.db.Model(&models.IncomingLog{}).
		Select("DATE(incoming_time) as date, COUNT(*) as count, COUNT(CASE WHEN is_duplicate = true THEN 1 END) as duplicate_count").
		Where("line_account_id = ? AND incoming_time >= ?", accountID, startDate).
		Group("DATE(incoming_time)").
		Order("date ASC").
		Rows()
	
	if err != nil {
		logger.Errorf("查询账号进线趋势失败: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var date time.Time
		var count, duplicateCount int64
		
		if err := rows.Scan(&date, &count, &duplicateCount); err != nil {
			logger.Errorf("扫描趋势数据失败: %v", err)
			continue
		}
		
		results = append(results, map[string]interface{}{
			"date":           date.Format("2006-01-02"),
			"count":          count,
			"duplicate_count": duplicateCount,
			"unique_count":   count - duplicateCount,
		})
	}
	
	return results, nil
}

