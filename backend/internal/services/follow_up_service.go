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

// FollowUpService 跟进记录服务
type FollowUpService struct {
	db *gorm.DB
}

// NewFollowUpService 创建跟进记录服务实例
func NewFollowUpService() *FollowUpService {
	return &FollowUpService{
		db: database.GetDB(),
	}
}

// CreateFollowUp 创建跟进记录
func (s *FollowUpService) CreateFollowUp(c *gin.Context, req *schemas.CreateFollowUpRequest) (*models.FollowUpRecord, error) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		userID = nil
	}

	// 检查分组是否存在
	var group models.Group
	if err := s.db.Where("id = ? AND deleted_at IS NULL", req.GroupID).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分组不存在")
		}
		return nil, err
	}

	// 检查Line账号是否存在（如果提供了line_account_id）
	var lineAccount *models.LineAccount
	if req.LineAccountID != nil {
		var la models.LineAccount
		if err := s.db.Where("id = ? AND group_id = ? AND deleted_at IS NULL", *req.LineAccountID, req.GroupID).
			First(&la).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("Line账号不存在或不属于该分组")
			}
			return nil, err
		}
		lineAccount = &la
	}

	// 检查客户是否存在（如果提供了customer_id）
	var customer *models.Customer
	if req.CustomerID != nil {
		var cus models.Customer
		if err := s.db.Where("id = ? AND group_id = ? AND deleted_at IS NULL", *req.CustomerID, req.GroupID).
			First(&cus).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("客户不存在或不属于该分组")
			}
			return nil, err
		}
		customer = &cus
	}

	// 创建跟进记录
	record := &models.FollowUpRecord{
		GroupID:        req.GroupID,
		ActivationCode: group.ActivationCode,
		LineAccountID:  req.LineAccountID,
		CustomerID:     req.CustomerID,
		PlatformType:   req.PlatformType,
		Content:        req.Content,
	}

	// 设置创建者
	if userID != nil {
		uid := userID.(uint)
		record.CreatedBy = &uid
	}

	// 填充账号和客户信息（用于显示）
	if lineAccount != nil {
		record.LineAccountDisplayName = lineAccount.DisplayName
		record.LineAccountLineID = lineAccount.LineID
		record.LineAccountAvatarURL = lineAccount.AvatarURL
	}

	if customer != nil {
		record.CustomerDisplayName = customer.DisplayName
		record.CustomerLineID = customer.CustomerID
		record.CustomerAvatarURL = customer.AvatarURL
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("创建跟进记录失败: %w", err)
	}

	return record, nil
}

// GetFollowUpList 获取跟进记录列表（带分页和筛选）
func (s *FollowUpService) GetFollowUpList(c *gin.Context, params *schemas.FollowUpQueryParams) ([]schemas.FollowUpListResponse, int64, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.FollowUpRecord{}), "follow_up_records")

	// 添加查询条件
	query = query.Where("deleted_at IS NULL")

	if params.GroupID != nil {
		query = query.Where("group_id = ?", *params.GroupID)
	}

	if params.LineAccountID != nil {
		query = query.Where("line_account_id = ?", *params.LineAccountID)
	}

	if params.CustomerID != nil {
		query = query.Where("customer_id = ?", *params.CustomerID)
	}

	if params.PlatformType != "" {
		query = query.Where("platform_type = ?", params.PlatformType)
	}

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("content LIKE ?", search)
	}

	// 时间范围筛选
	if params.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, params.StartTime)
		if err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if params.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, params.EndTime)
		if err == nil {
			query = query.Where("created_at <= ?", endTime)
		}
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

	// 查询跟进记录列表
	var records []models.FollowUpRecord
	if err := query.
		Preload("Group").
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	result := make([]schemas.FollowUpListResponse, 0, len(records))
	for _, record := range records {
		groupRemark := ""
		if record.Group != nil {
			groupRemark = record.Group.Remark
		}

		createdByUsername := ""
		if record.User != nil {
			createdByUsername = record.User.Username
		}

		result = append(result, schemas.FollowUpListResponse{
			ID:                     record.ID,
			GroupID:                record.GroupID,
			ActivationCode:         record.ActivationCode,
			LineAccountID:          record.LineAccountID,
			CustomerID:             record.CustomerID,
			PlatformType:           record.PlatformType,
			LineAccountDisplayName: record.LineAccountDisplayName,
			LineAccountLineID:      record.LineAccountLineID,
			LineAccountAvatarURL:   record.LineAccountAvatarURL,
			CustomerDisplayName:    record.CustomerDisplayName,
			CustomerLineID:         record.CustomerLineID,
			CustomerAvatarURL:      record.CustomerAvatarURL,
			Content:                record.Content,
			CreatedBy:              record.CreatedBy,
			CreatedAt:              record.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              record.UpdatedAt.Format(time.RFC3339),
			GroupRemark:            groupRemark,
			CreatedByUsername:      createdByUsername,
		})
	}

	return result, total, nil
}

// UpdateFollowUp 更新跟进记录内容
func (s *FollowUpService) UpdateFollowUp(c *gin.Context, id uint64, req *schemas.UpdateFollowUpRequest) (*models.FollowUpRecord, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.FollowUpRecord{}), "follow_up_records")

	var record models.FollowUpRecord
	if err := query.Where("id = ? AND deleted_at IS NULL", id).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("跟进记录不存在")
		}
		return nil, err
	}

	// 更新内容
	record.Content = req.Content

	if err := s.db.Save(&record).Error; err != nil {
		return nil, fmt.Errorf("更新跟进记录失败: %w", err)
	}

	return &record, nil
}

// DeleteFollowUp 删除跟进记录（软删除）
func (s *FollowUpService) DeleteFollowUp(c *gin.Context, id uint64) error {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.FollowUpRecord{}), "follow_up_records")

	var record models.FollowUpRecord
	if err := query.Where("id = ? AND deleted_at IS NULL", id).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("跟进记录不存在")
		}
		return err
	}

	if err := s.db.Delete(&record).Error; err != nil {
		return fmt.Errorf("删除跟进记录失败: %w", err)
	}

	return nil
}

// BatchCreateFollowUp 批量创建跟进记录
func (s *FollowUpService) BatchCreateFollowUp(c *gin.Context, req *schemas.BatchCreateFollowUpRequest) ([]*models.FollowUpRecord, error) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		userID = nil
	}

	records := make([]*models.FollowUpRecord, 0, len(req.Records))
	for _, recordReq := range req.Records {
		// 检查分组是否存在
		var group models.Group
		if err := s.db.Where("id = ? AND deleted_at IS NULL", recordReq.GroupID).First(&group).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("分组不存在: group_id=%d", recordReq.GroupID)
			}
			return nil, err
		}

		record := &models.FollowUpRecord{
			GroupID:        recordReq.GroupID,
			ActivationCode: group.ActivationCode,
			LineAccountID:  recordReq.LineAccountID,
			CustomerID:     recordReq.CustomerID,
			PlatformType:   recordReq.PlatformType,
			Content:        recordReq.Content,
		}

		// 设置创建者
		if userID != nil {
			uid := userID.(uint)
			record.CreatedBy = &uid
		}

		// 填充账号和客户信息（如果提供了ID）
		if recordReq.LineAccountID != nil {
			var lineAccount models.LineAccount
			if err := s.db.Where("id = ? AND deleted_at IS NULL", *recordReq.LineAccountID).
				First(&lineAccount).Error; err == nil {
				record.LineAccountDisplayName = lineAccount.DisplayName
				record.LineAccountLineID = lineAccount.LineID
				record.LineAccountAvatarURL = lineAccount.AvatarURL
			}
		}

		if recordReq.CustomerID != nil {
			var customer models.Customer
			if err := s.db.Where("id = ? AND deleted_at IS NULL", *recordReq.CustomerID).
				First(&customer).Error; err == nil {
				record.CustomerDisplayName = customer.DisplayName
				record.CustomerLineID = customer.CustomerID
				record.CustomerAvatarURL = customer.AvatarURL
			}
		}

		records = append(records, record)
	}

	// 批量插入
	if err := s.db.Create(records).Error; err != nil {
		return nil, fmt.Errorf("批量创建跟进记录失败: %w", err)
	}

	return records, nil
}

// SyncFollowUp 同步跟进记录（从Windows客户端）
func (s *FollowUpService) SyncFollowUp(groupID uint, activationCode string, data *schemas.FollowUpSyncData) (*models.FollowUpRecord, error) {
	// 查找Line账号（如果提供了line_account_id）
	var lineAccountID *uint
	var lineAccount *models.LineAccount
	if data.LineAccountID != "" {
		var la models.LineAccount
		if err := s.db.Where("group_id = ? AND line_id = ? AND deleted_at IS NULL", groupID, data.LineAccountID).
			First(&la).Error; err == nil {
			lineAccountID = &la.ID
			lineAccount = &la
		} else {
			logger.Warnf("Line账号不存在: line_id=%s, group_id=%d", data.LineAccountID, groupID)
		}
	}

	// 查找客户（如果提供了customer_id）
	var customerID *uint64
	var customer *models.Customer
	if data.CustomerID != "" {
		var cus models.Customer
		if err := s.db.Where("group_id = ? AND customer_id = ? AND platform_type = ? AND deleted_at IS NULL",
			groupID, data.CustomerID, data.PlatformType).First(&cus).Error; err == nil {
			customerID = &cus.ID
			customer = &cus
		} else {
			logger.Warnf("客户不存在: customer_id=%s, group_id=%d", data.CustomerID, groupID)
		}
	}

	// 确定平台类型（如果没有提供，默认为line）
	platformType := data.PlatformType
	if platformType == "" {
		platformType = "line"
	}

	// 创建跟进记录
	record := &models.FollowUpRecord{
		GroupID:        groupID,
		ActivationCode: activationCode,
		LineAccountID:  lineAccountID,
		CustomerID:     customerID,
		PlatformType:   platformType,
		Content:        data.Content,
	}

	// 填充账号和客户信息
	if lineAccount != nil {
		record.LineAccountDisplayName = lineAccount.DisplayName
		record.LineAccountLineID = lineAccount.LineID
		record.LineAccountAvatarURL = lineAccount.AvatarURL
	}

	if customer != nil {
		record.CustomerDisplayName = customer.DisplayName
		record.CustomerLineID = customer.CustomerID
		record.CustomerAvatarURL = customer.AvatarURL
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("创建跟进记录失败: %w", err)
	}

	logger.Infof("创建跟进记录: id=%d, group_id=%d, customer_id=%s", record.ID, groupID, data.CustomerID)
	return record, nil
}

