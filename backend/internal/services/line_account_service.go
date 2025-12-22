package services

import (
	"errors"
	"fmt"
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
	if group.AccountLimit != nil {
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

	// 更新字段
	if req.DisplayName != "" {
		account.DisplayName = req.DisplayName
	}

	if req.PhoneNumber != "" {
		account.PhoneNumber = req.PhoneNumber
	}

	if req.ProfileURL != "" {
		account.ProfileURL = req.ProfileURL
	}

	if req.AvatarURL != "" {
		account.AvatarURL = req.AvatarURL
	}

	if req.Bio != "" {
		account.Bio = req.Bio
	}

	if req.StatusMessage != "" {
		account.StatusMessage = req.StatusMessage
	}

	if req.AccountRemark != "" {
		account.AccountRemark = req.AccountRemark
	}

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

