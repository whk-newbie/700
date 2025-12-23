package services

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/internal/utils"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GroupService 分组服务
type GroupService struct {
	db *gorm.DB
}

// NewGroupService 创建分组服务实例
func NewGroupService() *GroupService {
	return &GroupService{
		db: database.GetDB(),
	}
}

// GenerateActivationCode 生成激活码（8位大写字母+数字）
func (s *GroupService) GenerateActivationCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8
	
	rand.Seed(time.Now().UnixNano())
	
	for i := 0; i < 100; i++ { // 最多尝试100次
		var code strings.Builder
		for j := 0; j < length; j++ {
			code.WriteByte(charset[rand.Intn(len(charset))])
		}
		
		activationCode := code.String()
		
		// 检查是否已存在
		var count int64
		if err := s.db.Model(&models.Group{}).
			Where("activation_code = ? AND deleted_at IS NULL", activationCode).
			Count(&count).Error; err != nil {
			return "", err
		}
		
		if count == 0 {
			return activationCode, nil
		}
	}
	
	return "", errors.New("无法生成唯一的激活码")
}

// CreateGroup 创建分组
func (s *GroupService) CreateGroup(c *gin.Context, req *schemas.CreateGroupRequest) (*models.Group, error) {
	// 检查用户是否存在
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", req.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}

	// 检查普通用户的分组数量限制
	if user.Role == "user" && user.MaxGroups != nil {
		var count int64
		if err := s.db.Model(&models.Group{}).
			Where("user_id = ? AND deleted_at IS NULL", req.UserID).
			Count(&count).Error; err != nil {
			return nil, err
		}
		
		if int(count) >= *user.MaxGroups {
			return nil, fmt.Errorf("已达到最大分组数量限制: %d", *user.MaxGroups)
		}
	}

	// 生成激活码
	activationCode, err := s.GenerateActivationCode()
	if err != nil {
		return nil, err
	}

	// 处理登录密码
	var loginPasswordHash string
	if req.LoginPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.LoginPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		loginPasswordHash = string(hash)
	}

	// 设置默认值
	category := req.Category
	if category == "" {
		category = "default"
	}
	
	dedupScope := req.DedupScope
	if dedupScope == "" {
		dedupScope = "current"
	}
	
	resetTime := req.ResetTime
	if resetTime == "" {
		resetTime = "09:00:00"
	}

	// 创建分组
	group := &models.Group{
		UserID:        req.UserID,
		ActivationCode: activationCode,
		AccountLimit:  req.AccountLimit,
		IsActive:      req.IsActive,
		Remark:        req.Remark,
		Description:   req.Description,
		Category:      category,
		DedupScope:    dedupScope,
		ResetTime:     resetTime,
		LoginPassword: loginPasswordHash,
	}

	if err := s.db.Create(group).Error; err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("激活码已存在，请重试")
		}
		return nil, err
	}

	// 初始化group_stats
	groupStats := &models.GroupStats{
		GroupID: group.ID,
	}
	if err := s.db.Create(groupStats).Error; err != nil {
		logger.Warnf("创建分组统计失败: %v", err)
		// 不影响分组创建，只记录日志
	}

	return group, nil
}

// GetGroupList 获取分组列表（带分页和筛选）
func (s *GroupService) GetGroupList(c *gin.Context, params *schemas.GroupQueryParams) ([]schemas.GroupListResponse, int64, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Group{}), "groups")
	
	// 添加查询条件
	query = query.Where("deleted_at IS NULL")
	
	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}
	
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}
	
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}
	
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("activation_code LIKE ? OR remark LIKE ?", search, search)
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

	// 查询分组列表
	var groups []models.Group
	if err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	// 获取分组ID列表
	groupIDs := make([]uint, 0, len(groups))
	for _, g := range groups {
		groupIDs = append(groupIDs, g.ID)
	}

	// 实时计算统计信息
	statsMap := make(map[uint]*models.GroupStats)
	for _, groupID := range groupIDs {
		stats, err := s.calculateGroupStats(groupID)
		if err != nil {
			logger.Warnf("计算分组统计失败 group_id=%d: %v", groupID, err)
			// 如果计算失败，使用空统计
			statsMap[groupID] = &models.GroupStats{GroupID: groupID}
		} else {
			statsMap[groupID] = stats
		}
	}

	// 转换为响应格式
	result := make([]schemas.GroupListResponse, 0, len(groups))
	for _, g := range groups {
		var lastLoginAt *string
		if g.LastLoginAt != nil {
			timeStr := g.LastLoginAt.Format(time.RFC3339)
			lastLoginAt = &timeStr
		}

		// 获取统计信息
		stats := statsMap[g.ID]
		if stats == nil {
			stats = &models.GroupStats{} // 默认空统计
		}

		username := ""
		if g.User != nil {
			username = g.User.Username
		}
		
		result = append(result, schemas.GroupListResponse{
			ID:                 g.ID,
			UserID:             g.UserID,
			ActivationCode:     g.ActivationCode,
			AccountLimit:       g.AccountLimit,
			IsActive:           g.IsActive,
			Remark:             g.Remark,
			Description:        g.Description,
			Category:           g.Category,
			DedupScope:         g.DedupScope,
			ResetTime:          g.ResetTime,
			CreatedAt:          g.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          g.UpdatedAt.Format(time.RFC3339),
			LastLoginAt:        lastLoginAt,
			TotalAccounts:      stats.TotalAccounts,
			OnlineAccounts:     stats.OnlineAccounts,
			LineAccounts:       stats.LineAccounts,
			LineBusinessAccounts: stats.LineBusinessAccounts,
			TodayIncoming:      stats.TodayIncoming,
			TotalIncoming:      stats.TotalIncoming,
			DuplicateIncoming: stats.DuplicateIncoming,
			TodayDuplicate:     stats.TodayDuplicate,
			Username:           username,
		})
	}

	return result, total, nil
}

// GetGroupByID 根据ID获取分组
func (s *GroupService) GetGroupByID(c *gin.Context, id uint) (*models.Group, error) {
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Group{}), "groups")
	
	var group models.Group
	if err := query.Where("id = ? AND deleted_at IS NULL", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分组不存在")
		}
		return nil, err
	}
	
	return &group, nil
}

// UpdateGroup 更新分组
func (s *GroupService) UpdateGroup(c *gin.Context, id uint, req *schemas.UpdateGroupRequest) (*models.Group, error) {
	// 先获取分组
	group, err := s.GetGroupByID(c, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.AccountLimit != nil {
		group.AccountLimit = req.AccountLimit
	}
	
	if req.IsActive != nil {
		group.IsActive = *req.IsActive
	}
	
	if req.Remark != "" {
		group.Remark = req.Remark
	}
	
	if req.Description != "" {
		group.Description = req.Description
	}
	
	if req.Category != "" {
		group.Category = req.Category
	}
	
	if req.DedupScope != "" {
		group.DedupScope = req.DedupScope
	}
	
	if req.ResetTime != "" {
		group.ResetTime = req.ResetTime
	}
	
	// 更新登录密码
	if req.LoginPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.LoginPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		group.LoginPassword = string(hash)
	}

	if err := s.db.Save(group).Error; err != nil {
		return nil, err
	}

	return group, nil
}

// DeleteGroup 删除分组（软删除）
func (s *GroupService) DeleteGroup(c *gin.Context, id uint) error {
	// 先获取分组
	group, err := s.GetGroupByID(c, id)
	if err != nil {
		return err
	}

	// 软删除
	if err := s.db.Delete(group).Error; err != nil {
		return err
	}

	return nil
}

// RegenerateActivationCode 重新生成激活码
func (s *GroupService) RegenerateActivationCode(c *gin.Context, id uint) (string, error) {
	// 先获取分组
	group, err := s.GetGroupByID(c, id)
	if err != nil {
		return "", err
	}

	// 生成新的激活码
	newCode, err := s.GenerateActivationCode()
	if err != nil {
		return "", err
	}

	// 更新激活码
	group.ActivationCode = newCode
	if err := s.db.Save(group).Error; err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return "", errors.New("激活码已存在，请重试")
		}
		return "", err
	}

	return newCode, nil
}

// GetCategories 获取所有分组分类
func (s *GroupService) GetCategories(c *gin.Context) ([]string, error) {
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Group{}), "groups")
	
	var categories []string
	if err := query.
		Where("deleted_at IS NULL AND category != ''").
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		return nil, err
	}
	
	return categories, nil
}

// BatchDeleteGroups 批量删除分组
func (s *GroupService) BatchDeleteGroups(c *gin.Context, ids []uint) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, errors.New("分组ID列表不能为空")
	}

	// 应用数据过滤，获取可操作的分组
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Group{}), "groups")
	query = query.Where("id IN ? AND deleted_at IS NULL", ids)

	var groups []models.Group
	if err := query.Find(&groups).Error; err != nil {
		return 0, nil, err
	}

	// 找出实际存在的分组ID
	existingIDs := make(map[uint]bool)
	for _, g := range groups {
		existingIDs[g.ID] = true
	}

	// 找出不存在的ID
	var failedIDs []uint
	for _, id := range ids {
		if !existingIDs[id] {
			failedIDs = append(failedIDs, id)
		}
	}

	// 批量软删除
	if len(groups) > 0 {
		groupIDs := make([]uint, 0, len(groups))
		for _, g := range groups {
			groupIDs = append(groupIDs, g.ID)
		}
		
		if err := s.db.Where("id IN ?", groupIDs).Delete(&models.Group{}).Error; err != nil {
			return 0, ids, err
		}
	}

	return len(groups), failedIDs, nil
}

// BatchUpdateGroups 批量更新分组
func (s *GroupService) BatchUpdateGroups(c *gin.Context, ids []uint, req *schemas.BatchUpdateGroupsRequest) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, errors.New("分组ID列表不能为空")
	}

	// 应用数据过滤，获取可操作的分组
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Group{}), "groups")
	query = query.Where("id IN ? AND deleted_at IS NULL", ids)

	var groups []models.Group
	if err := query.Find(&groups).Error; err != nil {
		return 0, nil, err
	}

	// 找出实际存在的分组ID
	existingIDs := make(map[uint]bool)
	for _, g := range groups {
		existingIDs[g.ID] = true
	}

	// 找出不存在的ID
	var failedIDs []uint
	for _, id := range ids {
		if !existingIDs[id] {
			failedIDs = append(failedIDs, id)
		}
	}

	// 批量更新
	if len(groups) > 0 {
		updates := make(map[string]interface{})
		
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}
		
		if req.Category != "" {
			updates["category"] = req.Category
		}
		
		if req.DedupScope != "" {
			updates["dedup_scope"] = req.DedupScope
		}

		if len(updates) > 0 {
			groupIDs := make([]uint, 0, len(groups))
			for _, g := range groups {
				groupIDs = append(groupIDs, g.ID)
			}
			
			if err := s.db.Model(&models.Group{}).
				Where("id IN ?", groupIDs).
				Updates(updates).Error; err != nil {
				return 0, ids, err
			}
		}
	}

	return len(groups), failedIDs, nil
}

// GenerateSubAccountTokenForUser 为用户生成子账户Token（管理员或普通用户）
// 管理员可以为任何分组生成Token，普通用户只能为自己管理的分组生成Token
func (s *GroupService) GenerateSubAccountTokenForUser(c *gin.Context, id uint) (string, error) {
	// 获取用户角色和ID
	role, roleExists := c.Get("role")
	userID, userIDExists := c.Get("user_id")

	if !roleExists {
		return "", errors.New("无法获取用户角色")
	}

	// 获取分组（应用数据过滤，普通用户只能看到自己的分组）
	group, err := s.GetGroupByID(c, id)
	if err != nil {
		return "", err
	}

	// 如果是普通用户，额外检查分组是否属于该用户
	if role == "user" {
		if !userIDExists {
			return "", errors.New("无法获取用户ID")
		}
		if group.UserID != userID.(uint) {
			return "", errors.New("无权访问该分组")
		}
	}

	// 检查分组是否激活
	if !group.IsActive {
		return "", errors.New("分组已被禁用")
	}

	// 更新最后登录时间
	now := time.Now()
	group.LastLoginAt = &now
	s.db.Save(group)

	// 生成Token
	token, err := utils.GenerateSubAccountToken(group.ID, group.ActivationCode)
	if err != nil {
		return "", err
	}

	// 创建Session（子账号使用GroupID作为标识）
	sessionService := NewSessionService()
	sessionInfo := &SessionInfo{
		GroupID:        group.ID,
		ActivationCode: group.ActivationCode,
		Role:           "subaccount",
		LoginTime:      time.Now(),
		IPAddress:      c.ClientIP(),
		UserAgent:      c.GetHeader("User-Agent"),
	}
	// 子账号使用GroupID作为Session的用户ID标识
	if err := sessionService.CreateSession(uint(group.ID), token, sessionInfo, 24*time.Hour); err != nil {
		// Session创建失败不影响Token生成
		logger.Warnf("创建Session失败: %v", err)
	}

	return token, nil
}

// calculateGroupStats 实时计算分组统计
func (s *GroupService) calculateGroupStats(groupID uint) (*models.GroupStats, error) {
	stats := &models.GroupStats{
		GroupID: groupID,
	}

	// 计算账号统计
	var totalAccounts, onlineAccounts, lineAccounts, lineBusinessAccounts int64

	// 总账号数
	if err := s.db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Count(&totalAccounts).Error; err != nil {
		return nil, err
	}

	// 在线账号数
	if err := s.db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL AND online_status = ?", groupID, "online").
		Count(&onlineAccounts).Error; err != nil {
		return nil, err
	}

	// Line账号数
	if err := s.db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL AND platform_type = ?", groupID, "line").
		Count(&lineAccounts).Error; err != nil {
		return nil, err
	}

	// Line Business账号数
	if err := s.db.Model(&models.LineAccount{}).
		Where("group_id = ? AND deleted_at IS NULL AND platform_type = ?", groupID, "line_business").
		Count(&lineBusinessAccounts).Error; err != nil {
		return nil, err
	}

	// 获取分组信息以计算今日时间范围
	var group models.Group
	if err := s.db.Where("id = ? AND deleted_at IS NULL", groupID).First(&group).Error; err != nil {
		return nil, err
	}

	// 计算今日时间范围
	now := time.Now()
	todayStartTime := s.getTodayStartTime(group.ResetTime, now)

	var todayIncoming, totalIncoming, duplicateIncoming, todayDuplicate int64

	// 今日进线数（从今天开始时间到当前时间）
	if err := s.db.Model(&models.IncomingLog{}).
		Where("group_id = ? AND incoming_time >= ?", groupID, todayStartTime).
		Count(&todayIncoming).Error; err != nil {
		return nil, err
	}

	// 总进线数
	if err := s.db.Model(&models.IncomingLog{}).
		Where("group_id = ?", groupID).
		Count(&totalIncoming).Error; err != nil {
		return nil, err
	}

	// 重复进线数
	if err := s.db.Model(&models.IncomingLog{}).
		Where("group_id = ? AND is_duplicate = ?", groupID, true).
		Count(&duplicateIncoming).Error; err != nil {
		return nil, err
	}

	// 今日重复进线数
	if err := s.db.Model(&models.IncomingLog{}).
		Where("group_id = ? AND incoming_time >= ? AND is_duplicate = ?", groupID, todayStartTime, true).
		Count(&todayDuplicate).Error; err != nil {
		return nil, err
	}

	// 赋值给stats
	stats.TotalAccounts = int(totalAccounts)
	stats.OnlineAccounts = int(onlineAccounts)
	stats.LineAccounts = int(lineAccounts)
	stats.LineBusinessAccounts = int(lineBusinessAccounts)
	stats.TodayIncoming = int(todayIncoming)
	stats.TotalIncoming = int(totalIncoming)
	stats.DuplicateIncoming = int(duplicateIncoming)
	stats.TodayDuplicate = int(todayDuplicate)

	return stats, nil
}

// getTodayStartTime 根据重置时间计算今日的开始时间
func (s *GroupService) getTodayStartTime(resetTimeStr string, now time.Time) time.Time {
	// 解析重置时间
	resetTime, err := s.parseResetTime(resetTimeStr)
	if err != nil {
		// 解析失败，使用默认时间 09:00:00
		resetTime = time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)
	}

	// 计算今天的重置时间点
	todayResetTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		resetTime.Hour(), resetTime.Minute(), resetTime.Second(),
		0, now.Location(),
	)

	// 如果当前时间还没到今天的重置时间，则使用昨天的重置时间作为开始时间
	if now.Before(todayResetTime) {
		yesterdayResetTime := todayResetTime.AddDate(0, 0, -1)
		return yesterdayResetTime
	}

	return todayResetTime
}

// parseResetTime 解析重置时间字符串（格式：HH:MM:SS）
func (s *GroupService) parseResetTime(resetTimeStr string) (time.Time, error) {
	// 默认重置时间为 09:00:00
	if resetTimeStr == "" {
		resetTimeStr = "09:00:00"
	}

	// 解析时间字符串
	parsedTime, err := time.Parse("15:04:05", resetTimeStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

