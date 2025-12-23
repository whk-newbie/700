package services

import (
	"time"

	"line-management/internal/models"
	"line-management/internal/utils"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
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
func (s *StatsService) GetOverviewStats(c *gin.Context) (map[string]interface{}, error) {
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
	
	// 总进线数（实时计算）
	var totalIncoming int64
	incomingQuery := utils.ApplyDataFilter(c, s.db.Model(&models.IncomingLog{}), "incoming_logs")
	incomingQuery.Count(&totalIncoming)
	result["total_incoming"] = totalIncoming

	// 总重复数（实时计算）
	var totalDuplicate int64
	duplicateQuery := utils.ApplyDataFilter(c, s.db.Model(&models.IncomingLog{}), "incoming_logs")
	duplicateQuery.Where("is_duplicate = ?", true).Count(&totalDuplicate)
	result["duplicate_incoming"] = totalDuplicate

	// 今日进线数（实时计算，从今天00:00:00开始）
	now := time.Now()
	today := now.Format("2006-01-02")
	var todayIncoming int64
	todayIncomingQuery := utils.ApplyDataFilter(c, s.db.Model(&models.IncomingLog{}), "incoming_logs")
	todayIncomingQuery.Where("DATE(incoming_time) = ?", today).Count(&todayIncoming)
	result["today_incoming"] = todayIncoming

	// 今日重复数（实时计算，从今天00:00:00开始）
	var todayDuplicate int64
	todayDuplicateQuery := utils.ApplyDataFilter(c, s.db.Model(&models.IncomingLog{}), "incoming_logs")
	todayDuplicateQuery.Where("DATE(incoming_time) = ? AND is_duplicate = ?", today, true).Count(&todayDuplicate)
	result["today_duplicate"] = todayDuplicate
	
	// 底库总数（GORM会自动处理软删除）
	var totalContacts int64
	s.db.Model(&models.ContactPool{}).Count(&totalContacts)
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
	
	// 查询最近N天的进线数据
	startDate := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)
	
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
	
	// 先创建一个日期到数据的映射
	dataMap := make(map[string]map[string]interface{})
	for rows.Next() {
		var dateValue interface{}
		var count, duplicateCount int64
		
		if err := rows.Scan(&dateValue, &count, &duplicateCount); err != nil {
			logger.Errorf("扫描趋势数据失败: %v", err)
			continue
		}
		
		// 处理日期值，可能是time.Time或string
		var dateStr string
		switch v := dateValue.(type) {
		case time.Time:
			dateStr = v.Format("2006-01-02")
		case string:
			// 如果是字符串，尝试解析
			if t, err := time.Parse("2006-01-02", v); err == nil {
				dateStr = t.Format("2006-01-02")
			} else {
				dateStr = v[:10] // 取前10个字符（YYYY-MM-DD）
			}
		default:
			logger.Errorf("未知的日期类型: %T", dateValue)
			continue
		}
		
		dataMap[dateStr] = map[string]interface{}{
			"date":            dateStr,
			"incoming_count":  count,
			"duplicate_count": duplicateCount,
			"unique_count":    count - duplicateCount,
		}
	}
	
	// 填充完整的日期范围，即使某些日期没有数据也要填充0
	var results []map[string]interface{}
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		
		if data, exists := dataMap[dateStr]; exists {
			results = append(results, data)
		} else {
			results = append(results, map[string]interface{}{
				"date":            dateStr,
				"incoming_count":  int64(0),
				"duplicate_count": int64(0),
				"unique_count":    int64(0),
			})
		}
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
	
	// 查询最近N天的进线数据
	startDate := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)
	
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
	
	// 先创建一个日期到数据的映射
	dataMap := make(map[string]map[string]interface{})
	for rows.Next() {
		var date time.Time
		var count, duplicateCount int64
		
		if err := rows.Scan(&date, &count, &duplicateCount); err != nil {
			logger.Errorf("扫描趋势数据失败: %v", err)
			continue
		}
		
		dateStr := date.Format("2006-01-02")
		dataMap[dateStr] = map[string]interface{}{
			"date":            dateStr,
			"incoming_count":  count,
			"duplicate_count": duplicateCount,
			"unique_count":    count - duplicateCount,
		}
	}
	
	// 填充完整的日期范围，即使某些日期没有数据也要填充0
	var results []map[string]interface{}
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		
		if data, exists := dataMap[dateStr]; exists {
			results = append(results, data)
		} else {
			results = append(results, map[string]interface{}{
				"date":            dateStr,
				"incoming_count":  int64(0),
				"duplicate_count": int64(0),
				"unique_count":    int64(0),
			})
		}
	}
	
	return results, nil
}

