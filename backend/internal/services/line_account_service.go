package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/internal/utils"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LineAccountService Line账号服务
type LineAccountService struct {
	db *gorm.DB
}

// NewLineAccountService 创建Line账号服务实例
func NewLineAccountService() *LineAccountService {
	return &LineAccountService{
		db: database.GetDB(),
	}
}

// CreateLineAccount 创建Line账号
func (s *LineAccountService) CreateLineAccount(c *gin.Context, req *schemas.CreateLineAccountRequest) (*models.LineAccount, error) {
	// 检查分组是否存在
	var group models.Group
	if err := s.db.Where("id = ? AND deleted_at IS NULL", req.GroupID).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分组不存在")
		}
		return nil, err
	}

	// 检查分组是否激活
	if !group.IsActive {
		return nil, errors.New("分组已被禁用")
	}

	// 检查账号数量限制
	// 规则：nil 或 -1 表示无限制，0 表示显示为0但实际允许，>0 表示有限制
	if group.AccountLimit != nil && *group.AccountLimit > 0 {
		var count int64
		if err := s.db.Model(&models.LineAccount{}).
			Where("group_id = ? AND deleted_at IS NULL", req.GroupID).
			Count(&count).Error; err != nil {
			return nil, err
		}

		if int(count) >= *group.AccountLimit {
			return nil, fmt.Errorf("已达到分组账号数量限制: %d", *group.AccountLimit)
		}
	}

	// 检查同一分组下是否已存在相同的line_id（未删除的）
	var existingAccount models.LineAccount
	if err := s.db.Where("group_id = ? AND line_id = ? AND deleted_at IS NULL", req.GroupID, req.LineID).
		First(&existingAccount).Error; err == nil {
		return nil, errors.New("该Line ID在此分组中已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 创建账号
	account := &models.LineAccount{
		GroupID:       req.GroupID,
		ActivationCode: group.ActivationCode,
		PlatformType:  req.PlatformType,
		LineID:        req.LineID,
		DisplayName:   req.DisplayName,
		PhoneNumber:   req.PhoneNumber,
		ProfileURL:    req.ProfileURL,
		AvatarURL:     req.AvatarURL,
		Bio:           req.Bio,
		StatusMessage: req.StatusMessage,
		AddFriendLink: req.AddFriendLink,
		AccountRemark: req.AccountRemark,
		OnlineStatus:  "offline",
	}

	if err := s.db.Create(account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("创建账号失败")
		}
		return nil, err
	}

	// 初始化line_account_stats
	stats := &models.LineAccountStats{
		LineAccountID: account.ID,
	}
	if err := s.db.Create(stats).Error; err != nil {
		logger.Warnf("创建账号统计失败: %v", err)
		// 不影响账号创建，只记录日志
	}

	// 更新group_stats
	if err := s.updateGroupStatsForAccountChange(account.GroupID, account.PlatformType, account.OnlineStatus, true); err != nil {
		logger.Warnf("更新分组统计失败: %v", err)
		// 不影响账号创建，只记录日志
	}

	return account, nil
}

// GetLineAccountList 获取Line账号列表（带分页和筛选）
func (s *LineAccountService) GetLineAccountList(c *gin.Context, params *schemas.LineAccountQueryParams) ([]schemas.LineAccountListResponse, int64, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.LineAccount{}), "line_accounts")

	// 添加查询条件
	query = query.Where("deleted_at IS NULL")

	if params.GroupID != nil {
		query = query.Where("group_id = ?", *params.GroupID)
	}

	if params.PlatformType != "" {
		query = query.Where("platform_type = ?", params.PlatformType)
	}

	if params.OnlineStatus != "" {
		query = query.Where("online_status = ?", params.OnlineStatus)
	}

	if params.ActivationCode != "" {
		query = query.Where("activation_code = ?", params.ActivationCode)
	}

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("line_id LIKE ? OR display_name LIKE ?", search, search)
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

	// 查询账号列表
	var accounts []models.LineAccount
	if err := query.
		Preload("Group").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&accounts).Error; err != nil {
		return nil, 0, err
	}

	// 获取账号ID列表
	accountIDs := make([]uint, 0, len(accounts))
	for _, a := range accounts {
		accountIDs = append(accountIDs, a.ID)
	}

	// 批量查询统计信息
	var statsList []models.LineAccountStats
	if len(accountIDs) > 0 {
		if err := s.db.Where("line_account_id IN ?", accountIDs).Find(&statsList).Error; err != nil {
			logger.Warnf("查询账号统计失败: %v", err)
		}
	}

	// 构建统计信息映射
	statsMap := make(map[uint]*models.LineAccountStats)
	for i := range statsList {
		statsMap[statsList[i].LineAccountID] = &statsList[i]
	}

	// 转换为响应格式
	result := make([]schemas.LineAccountListResponse, 0, len(accounts))
	for _, a := range accounts {
		var lastActiveAt, lastOnlineTime, firstLoginAt *string
		if a.LastActiveAt != nil {
			timeStr := a.LastActiveAt.Format(time.RFC3339)
			lastActiveAt = &timeStr
		}
		if a.LastOnlineTime != nil {
			timeStr := a.LastOnlineTime.Format(time.RFC3339)
			lastOnlineTime = &timeStr
		}
		if a.FirstLoginAt != nil {
			timeStr := a.FirstLoginAt.Format(time.RFC3339)
			firstLoginAt = &timeStr
		}

		// 获取统计信息
		stats := statsMap[a.ID]
		if stats == nil {
			stats = &models.LineAccountStats{} // 默认空统计
		}

		groupRemark := ""
		if a.Group != nil {
			groupRemark = a.Group.Remark
		}

		result = append(result, schemas.LineAccountListResponse{
			ID:              a.ID,
			GroupID:         a.GroupID,
			ActivationCode:  a.ActivationCode,
			PlatformType:    a.PlatformType,
			LineID:          a.LineID,
			DisplayName:     a.DisplayName,
			PhoneNumber:     a.PhoneNumber,
			ProfileURL:      a.ProfileURL,
			AvatarURL:       a.AvatarURL,
			Bio:             a.Bio,
			StatusMessage:   a.StatusMessage,
			AddFriendLink:   a.AddFriendLink,
			QRCodePath:      a.QRCodePath,
			OnlineStatus:    a.OnlineStatus,
			LastActiveAt:    lastActiveAt,
			LastOnlineTime:  lastOnlineTime,
			FirstLoginAt:    firstLoginAt,
			AccountRemark:   a.AccountRemark,
			CreatedAt:       a.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       a.UpdatedAt.Format(time.RFC3339),
			TodayIncoming:   stats.TodayIncoming,
			TotalIncoming:   stats.TotalIncoming,
			DuplicateIncoming: stats.DuplicateIncoming,
			TodayDuplicate:    stats.TodayDuplicate,
			GroupRemark:       groupRemark,
		})
	}

	return result, total, nil
}

// GetLineAccountByID 根据ID获取Line账号
func (s *LineAccountService) GetLineAccountByID(c *gin.Context, id uint) (*models.LineAccount, error) {
	query := utils.ApplyDataFilter(c, s.db.Model(&models.LineAccount{}), "line_accounts")

	var account models.LineAccount
	if err := query.Where("id = ? AND deleted_at IS NULL", id).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号不存在")
		}
		return nil, err
	}

	return &account, nil
}

// UpdateLineAccount 更新Line账号
func (s *LineAccountService) UpdateLineAccount(c *gin.Context, id uint, req *schemas.UpdateLineAccountRequest) (*models.LineAccount, error) {
	// 先获取账号
	account, err := s.GetLineAccountByID(c, id)
	if err != nil {
		return nil, err
	}

	oldStatus := account.OnlineStatus
	oldPlatformType := account.PlatformType
	oldGroupID := account.GroupID

	// 如果更新分组，需要验证新分组
	if req.GroupID != nil && *req.GroupID != account.GroupID {
		var newGroup models.Group
		if err := s.db.Where("id = ? AND deleted_at IS NULL", *req.GroupID).First(&newGroup).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("分组不存在")
			}
			return nil, err
		}

		if !newGroup.IsActive {
			return nil, errors.New("分组已被禁用")
		}

		// 检查新分组下是否已存在相同的line_id（排除当前账号）
		var existingAccount models.LineAccount
		if err := s.db.Where("group_id = ? AND line_id = ? AND id != ? AND deleted_at IS NULL", *req.GroupID, account.LineID, id).
			First(&existingAccount).Error; err == nil {
			return nil, errors.New("该Line ID在新分组中已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		account.GroupID = *req.GroupID
		account.ActivationCode = newGroup.ActivationCode

		// 更新统计：从旧分组减少，向新分组增加
		if err := s.updateGroupStatsForAccountChange(oldGroupID, oldPlatformType, oldStatus, false); err != nil {
			logger.Warnf("更新分组统计失败: %v", err)
		}
		if err := s.updateGroupStatsForAccountChange(account.GroupID, oldPlatformType, oldStatus, true); err != nil {
			logger.Warnf("更新分组统计失败: %v", err)
		}
	}

	// 如果更新平台类型
	if req.PlatformType != "" && req.PlatformType != account.PlatformType {
		account.PlatformType = req.PlatformType
	}

	// 如果更新Line ID，需要检查是否重复
	if req.LineID != "" && req.LineID != account.LineID {
		// 检查同一分组下是否已存在相同的line_id（排除当前账号）
		var existingAccount models.LineAccount
		if err := s.db.Where("group_id = ? AND line_id = ? AND id != ? AND deleted_at IS NULL", account.GroupID, req.LineID, id).
			First(&existingAccount).Error; err == nil {
			return nil, errors.New("该Line ID在此分组中已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		account.LineID = req.LineID
	}

	// 更新其他字段（支持空值，使用指针判断是否提供）
	// 使用字符串指针的方式：如果字段在请求中（即使为空字符串），也更新
	// 这里我们通过检查字段是否在JSON中提供来判断，但由于Go的限制，我们使用特殊值判断
	// 实际上，前端会发送空字符串，我们直接更新即可

	// 对于字符串字段，如果请求中提供了（即使是空字符串），也更新
	// 这里我们直接赋值，因为前端会发送所有字段
	account.DisplayName = req.DisplayName
	account.PhoneNumber = req.PhoneNumber
	account.ProfileURL = req.ProfileURL
	account.AvatarURL = req.AvatarURL
	account.Bio = req.Bio
	account.StatusMessage = req.StatusMessage
	account.AddFriendLink = req.AddFriendLink
	account.AccountRemark = req.AccountRemark

	if req.OnlineStatus != "" {
		account.OnlineStatus = req.OnlineStatus
	}

	if err := s.db.Save(account).Error; err != nil {
		return nil, err
	}

	// 如果状态或平台类型发生变化，更新group_stats
	if oldStatus != account.OnlineStatus || oldPlatformType != account.PlatformType {
		// 先减少旧状态的统计
		if err := s.updateGroupStatsForAccountChange(account.GroupID, oldPlatformType, oldStatus, false); err != nil {
			logger.Warnf("更新分组统计失败: %v", err)
		}
		// 再增加新状态的统计
		if err := s.updateGroupStatsForAccountChange(account.GroupID, account.PlatformType, account.OnlineStatus, true); err != nil {
			logger.Warnf("更新分组统计失败: %v", err)
		}
	}

	return account, nil
}

// DeleteLineAccount 删除Line账号（软删除）
func (s *LineAccountService) DeleteLineAccount(c *gin.Context, id uint, deletedBy *uint) error {
	// 先获取账号
	account, err := s.GetLineAccountByID(c, id)
	if err != nil {
		return err
	}

	// 软删除
	account.DeletedBy = deletedBy
	if err := s.db.Delete(account).Error; err != nil {
		return err
	}

	// 更新group_stats（减少统计）
	if err := s.updateGroupStatsForAccountChange(account.GroupID, account.PlatformType, account.OnlineStatus, false); err != nil {
		logger.Warnf("更新分组统计失败: %v", err)
		// 不影响删除操作
	}

	return nil
}

// updateGroupStatsForAccountChange 更新分组统计（账号创建/删除/状态变化时调用）
// isAdd: true表示增加，false表示减少
func (s *LineAccountService) updateGroupStatsForAccountChange(groupID uint, platformType, onlineStatus string, isAdd bool) error {
	// 获取或创建group_stats
	var stats models.GroupStats
	if err := s.db.Where("group_id = ?", groupID).FirstOrCreate(&stats, models.GroupStats{GroupID: groupID}).Error; err != nil {
		return err
	}

	// 计算增量（增加为1，减少为-1）
	delta := 1
	if !isAdd {
		delta = -1
	}

	updates := make(map[string]interface{})

	// 更新总账号数
	updates["total_accounts"] = gorm.Expr("total_accounts + ?", delta)

	// 更新平台类型账号数
	if platformType == "line" {
		updates["line_accounts"] = gorm.Expr("line_accounts + ?", delta)
	} else if platformType == "line_business" {
		updates["line_business_accounts"] = gorm.Expr("line_business_accounts + ?", delta)
	}

	// 更新在线账号数
	if onlineStatus == "online" {
		updates["online_accounts"] = gorm.Expr("online_accounts + ?", delta)
	}

	// 更新updated_at
	updates["updated_at"] = time.Now()

	if err := s.db.Model(&stats).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// BatchDeleteLineAccounts 批量删除Line账号
func (s *LineAccountService) BatchDeleteLineAccounts(c *gin.Context, ids []uint, deletedBy *uint) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, errors.New("账号ID列表不能为空")
	}

	// 应用数据过滤，获取可操作的账号
	query := utils.ApplyDataFilter(c, s.db.Model(&models.LineAccount{}), "line_accounts")
	query = query.Where("id IN ? AND deleted_at IS NULL", ids)

	var accounts []models.LineAccount
	if err := query.Find(&accounts).Error; err != nil {
		return 0, nil, err
	}

	// 找出实际存在的账号ID
	existingIDs := make(map[uint]bool)
	for _, acc := range accounts {
		existingIDs[acc.ID] = true
	}

	// 找出不存在的ID
	var failedIDs []uint
	for _, id := range ids {
		if !existingIDs[id] {
			failedIDs = append(failedIDs, id)
		}
	}

	// 批量软删除并更新统计
	if len(accounts) > 0 {
		// 按分组和平台类型分组，用于批量更新统计
		groupStatsMap := make(map[uint]map[string]int) // groupID -> platformType -> count
		for _, acc := range accounts {
			if groupStatsMap[acc.GroupID] == nil {
				groupStatsMap[acc.GroupID] = make(map[string]int)
			}
			key := acc.PlatformType + "_" + acc.OnlineStatus
			groupStatsMap[acc.GroupID][key]++
		}

		// 执行批量删除
		accountIDs := make([]uint, 0, len(accounts))
		for _, acc := range accounts {
			accountIDs = append(accountIDs, acc.ID)
		}

		// 批量更新deleted_by
		if deletedBy != nil {
			if err := s.db.Model(&models.LineAccount{}).
				Where("id IN ?", accountIDs).
				Update("deleted_by", *deletedBy).Error; err != nil {
				return 0, ids, err
			}
		}

		// 批量软删除
		if err := s.db.Where("id IN ?", accountIDs).Delete(&models.LineAccount{}).Error; err != nil {
			return 0, ids, err
		}

		// 批量更新分组统计（减少统计）
		for groupID, stats := range groupStatsMap {
			for key, count := range stats {
				parts := strings.Split(key, "_")
				if len(parts) == 2 {
					platformType := parts[0]
					onlineStatus := parts[1]
					for i := 0; i < count; i++ {
						if err := s.updateGroupStatsForAccountChange(groupID, platformType, onlineStatus, false); err != nil {
							logger.Warnf("更新分组统计失败: %v", err)
							// 不影响删除操作
						}
					}
				}
			}
		}
	}

	return len(accounts), failedIDs, nil
}

// BatchUpdateLineAccounts 批量更新Line账号
func (s *LineAccountService) BatchUpdateLineAccounts(c *gin.Context, ids []uint, req *schemas.BatchUpdateLineAccountsRequest) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, errors.New("账号ID列表不能为空")
	}

	// 应用数据过滤，获取可操作的账号
	query := utils.ApplyDataFilter(c, s.db.Model(&models.LineAccount{}), "line_accounts")
	query = query.Where("id IN ? AND deleted_at IS NULL", ids)

	var accounts []models.LineAccount
	if err := query.Find(&accounts).Error; err != nil {
		return 0, nil, err
	}

	// 找出实际存在的账号ID
	existingIDs := make(map[uint]bool)
	for _, acc := range accounts {
		existingIDs[acc.ID] = true
	}

	// 找出不存在的ID
	var failedIDs []uint
	for _, id := range ids {
		if !existingIDs[id] {
			failedIDs = append(failedIDs, id)
		}
	}

	// 批量更新
	if len(accounts) > 0 {
		updates := make(map[string]interface{})
		
		if req.OnlineStatus != "" {
			updates["online_status"] = req.OnlineStatus
		}

		if len(updates) > 0 {
			accountIDs := make([]uint, 0, len(accounts))
			for _, acc := range accounts {
				accountIDs = append(accountIDs, acc.ID)
			}

			// 如果更新了在线状态，需要更新分组统计
			if req.OnlineStatus != "" {
				// 先减少旧状态的统计，再增加新状态的统计
				for _, acc := range accounts {
					// 减少旧状态
					if err := s.updateGroupStatsForAccountChange(acc.GroupID, acc.PlatformType, acc.OnlineStatus, false); err != nil {
						logger.Warnf("更新分组统计失败: %v", err)
					}
					// 增加新状态
					if err := s.updateGroupStatsForAccountChange(acc.GroupID, acc.PlatformType, req.OnlineStatus, true); err != nil {
						logger.Warnf("更新分组统计失败: %v", err)
					}
				}
			}

			// 执行批量更新
			if err := s.db.Model(&models.LineAccount{}).
				Where("id IN ?", accountIDs).
				Updates(updates).Error; err != nil {
				return 0, ids, err
			}
		}
	}

	return len(accounts), failedIDs, nil
}

