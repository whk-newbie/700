package services

import (
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/internal/utils"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// IncomingData 进线数据（从websocket包复制，避免循环依赖）
type IncomingData struct {
	LineAccountID  string `json:"line_account_id"`  // Line账号的line_id
	IncomingLineID string `json:"incoming_line_id"` // 进线客户的Line User ID
	PlatformType   string `json:"platform_type,omitempty"` // 平台类型（line / line_business）
	Timestamp      string `json:"timestamp"`
	DisplayName    string `json:"display_name,omitempty"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`
}

// IncomingUpdateCallback 进线更新回调函数类型
type IncomingUpdateCallback func(groupID uint, lineAccountID uint, incomingLineID string, isDuplicate bool)

// IncomingService 进线处理服务
type IncomingService struct {
	db          *gorm.DB
	dedupService *DedupService
	updateCallback IncomingUpdateCallback
}

// NewIncomingService 创建进线处理服务实例
func NewIncomingService(updateCallback IncomingUpdateCallback) *IncomingService {
	return &IncomingService{
		db:          database.GetDB(),
		dedupService: NewDedupService(),
		updateCallback: updateCallback,
	}
}

// ProcessIncoming 处理进线数据
// 1. 去重判断
// 2. 记录incoming_logs
// 3. 增量更新统计表
// 4. 添加到底库（如果不重复）
// 返回值: isDuplicate, error
func (s *IncomingService) ProcessIncoming(data *IncomingData, lineAccountID uint, groupID uint, dedupScope string) (bool, error) {
	var isDuplicateResult bool
	
	// 使用事务处理
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 去重判断
		isDuplicate, duplicateScope, err := s.dedupService.CheckDuplicate(groupID, data.IncomingLineID, dedupScope)
		if err != nil {
			logger.Errorf("去重检查失败: %v", err)
			return err
		}
		isDuplicateResult = isDuplicate

		// 2. 记录进线日志
		incomingLog := models.IncomingLog{
			LineAccountID:  lineAccountID,
			GroupID:        groupID,
			IncomingLineID: data.IncomingLineID,
			IncomingTime:    time.Now(),
			DisplayName:    data.DisplayName,
			AvatarURL:      data.AvatarURL,
			PhoneNumber:    data.PhoneNumber,
			IsDuplicate:    isDuplicate,
			DuplicateScope: duplicateScope,
		}

		// 保存原始数据
		if data.Timestamp != "" || data.DisplayName != "" || data.AvatarURL != "" || data.PhoneNumber != "" {
			rawData := make(map[string]interface{})
			if data.Timestamp != "" {
				rawData["timestamp"] = data.Timestamp
			}
			if data.DisplayName != "" {
				rawData["display_name"] = data.DisplayName
			}
			if data.AvatarURL != "" {
				rawData["avatar_url"] = data.AvatarURL
			}
			if data.PhoneNumber != "" {
				rawData["phone_number"] = data.PhoneNumber
			}
			incomingLog.RawData = models.JSONB(rawData)
		}

		if err := tx.Create(&incomingLog).Error; err != nil {
			logger.Errorf("记录进线日志失败: %v", err)
			return err
		}

		// 3. 增量更新账号统计（如果不存在则创建）
		updates := map[string]interface{}{
			"total_incoming": gorm.Expr("total_incoming + ?", 1),
			"today_incoming": gorm.Expr("today_incoming + ?", 1),
		}
		if isDuplicate {
			updates["duplicate_incoming"] = gorm.Expr("duplicate_incoming + ?", 1)
			updates["today_duplicate"] = gorm.Expr("today_duplicate + ?", 1)
		}

		// 检查账号统计是否存在
		var accountStatsCount int64
		if err := tx.Model(&models.LineAccountStats{}).
			Where("line_account_id = ?", lineAccountID).
			Count(&accountStatsCount).Error; err != nil {
			logger.Errorf("检查账号统计失败: %v", err)
			return err
		}

		if accountStatsCount == 0 {
			// 创建账号统计记录
			accountStats := models.LineAccountStats{
				LineAccountID: lineAccountID,
			}
			if err := tx.Create(&accountStats).Error; err != nil {
				logger.Errorf("创建账号统计失败: %v", err)
				return err
			}
		}

		if err := tx.Model(&models.LineAccountStats{}).
			Where("line_account_id = ?", lineAccountID).
			Updates(updates).Error; err != nil {
			logger.Errorf("更新账号统计失败: %v", err)
			return err
		}

		// 4. 增量更新分组统计（如果不存在则创建）
		// 检查分组统计是否存在
		var groupStatsCount int64
		if err := tx.Model(&models.GroupStats{}).
			Where("group_id = ?", groupID).
			Count(&groupStatsCount).Error; err != nil {
			logger.Errorf("检查分组统计失败: %v", err)
			return err
		}

		if groupStatsCount == 0 {
			// 创建分组统计记录
			groupStats := models.GroupStats{
				GroupID: groupID,
			}
			if err := tx.Create(&groupStats).Error; err != nil {
				logger.Errorf("创建分组统计失败: %v", err)
				return err
			}
		}

		if err := tx.Model(&models.GroupStats{}).
			Where("group_id = ?", groupID).
			Updates(updates).Error; err != nil {
			logger.Errorf("更新分组统计失败: %v", err)
			return err
		}

		// 5. 添加到底库（如果不重复）
		if !isDuplicate {
			// 获取Line账号信息以确定platform_type
			var lineAccount models.LineAccount
			if err := tx.Where("id = ?", lineAccountID).First(&lineAccount).Error; err != nil {
				logger.Errorf("获取Line账号信息失败: %v", err)
				// 不返回错误，继续处理
			} else {
				// 检查底库中是否已存在
				exists, err := s.dedupService.CheckContactPoolDuplicate(data.IncomingLineID, lineAccount.PlatformType)
				if err != nil {
					logger.Errorf("检查底库重复失败: %v", err)
					// 不返回错误，继续处理
				} else if !exists {
					// 获取分组信息
					var group models.Group
					if err := tx.Where("id = ?", groupID).First(&group).Error; err != nil {
						logger.Errorf("获取分组信息失败: %v", err)
					} else {
						contact := models.ContactPool{
							SourceType:     "platform",
							GroupID:        groupID,
							ActivationCode: group.ActivationCode,
							LineAccountID:  &lineAccountID,
							PlatformType:   lineAccount.PlatformType,
							LineID:         data.IncomingLineID,
							DisplayName:    data.DisplayName,
							PhoneNumber:    data.PhoneNumber,
							AvatarURL:      data.AvatarURL,
							DedupScope:     dedupScope,
							FirstSeenAt:    &incomingLog.IncomingTime,
						}

						if err := tx.Create(&contact).Error; err != nil {
							logger.Errorf("添加到底库失败: %v", err)
							// 不返回错误，继续处理
						}
					}
				}
			}
		}

		// 6. 推送实时更新到前端看板
		if s.updateCallback != nil {
			s.updateCallback(groupID, lineAccountID, data.IncomingLineID, isDuplicate)
		}

		logger.Infof("进线数据处理完成: GroupID=%d, LineAccountID=%d, IncomingLineID=%s, IsDuplicate=%v",
			groupID, lineAccountID, data.IncomingLineID, isDuplicate)

		return nil
	})
	
	return isDuplicateResult, err
}

// GetIncomingLogList 获取进线日志列表（带分页和筛选）
func (s *IncomingService) GetIncomingLogList(c *gin.Context, params *schemas.IncomingLogQueryParams) ([]schemas.IncomingLogListResponse, int64, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.IncomingLog{}), "incoming_logs")

	// 添加查询条件
	if params.GroupID != nil {
		query = query.Where("group_id = ?", *params.GroupID)
	}

	if params.LineAccountID != nil {
		query = query.Where("line_account_id = ?", *params.LineAccountID)
	}

	if params.IsDuplicate != nil {
		query = query.Where("is_duplicate = ?", *params.IsDuplicate)
	}

	if params.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, params.StartTime)
		if err == nil {
			query = query.Where("incoming_time >= ?", startTime)
		}
	}

	if params.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, params.EndTime)
		if err == nil {
			query = query.Where("incoming_time <= ?", endTime)
		}
	}

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("incoming_line_id LIKE ? OR display_name LIKE ?", search, search)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	// 查询进线日志列表
	var logs []models.IncomingLog
	if err := query.
		Preload("LineAccount").
		Preload("Group").
		Order("incoming_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	result := make([]schemas.IncomingLogListResponse, 0, len(logs))
	for _, log := range logs {
		item := schemas.IncomingLogListResponse{
			ID:             log.ID,
			LineAccountID:  log.LineAccountID,
			GroupID:        log.GroupID,
			IncomingLineID: log.IncomingLineID,
			IncomingTime:   log.IncomingTime.Format(time.RFC3339),
			DisplayName:    log.DisplayName,
			AvatarURL:      log.AvatarURL,
			PhoneNumber:    log.PhoneNumber,
			IsDuplicate:    log.IsDuplicate,
			DuplicateScope: log.DuplicateScope,
			CustomerType:   log.CustomerType,
		}

		// 关联Line账号信息
		if log.LineAccount != nil {
			item.LineAccount = &schemas.LineAccountInfo{
				ID:          log.LineAccount.ID,
				LineID:      log.LineAccount.LineID,
				DisplayName: log.LineAccount.DisplayName,
				PlatformType: log.LineAccount.PlatformType,
			}
		}

		// 关联分组信息
		if log.Group != nil {
			item.Group = &schemas.GroupInfo{
				ID:            log.Group.ID,
				ActivationCode: log.Group.ActivationCode,
				Remark:        log.Group.Remark,
			}
		}

		result = append(result, item)
	}

	return result, total, nil
}


