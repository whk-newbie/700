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

// CustomerService 客户服务
type CustomerService struct {
	db *gorm.DB
}

// NewCustomerService 创建客户服务实例
func NewCustomerService() *CustomerService {
	return &CustomerService{
		db: database.GetDB(),
	}
}

// CreateCustomer 创建客户
func (s *CustomerService) CreateCustomer(c *gin.Context, req *schemas.CreateCustomerRequest) (*models.Customer, error) {
	// 检查分组是否存在
	var group models.Group
	if err := s.db.Where("id = ? AND deleted_at IS NULL", req.GroupID).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分组不存在")
		}
		return nil, err
	}

	// 检查Line账号是否存在（如果提供了line_account_id）
	if req.LineAccountID != nil {
		var lineAccount models.LineAccount
		if err := s.db.Where("id = ? AND group_id = ? AND deleted_at IS NULL", *req.LineAccountID, req.GroupID).
			First(&lineAccount).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("Line账号不存在或不属于该分组")
			}
			return nil, err
		}
	}

	// 检查同一分组下是否已存在相同的customer_id（未删除的）
	var existingCustomer models.Customer
	if err := s.db.Where("group_id = ? AND customer_id = ? AND platform_type = ? AND deleted_at IS NULL",
		req.GroupID, req.CustomerID, req.PlatformType).First(&existingCustomer).Error; err == nil {
		return nil, errors.New("该客户在此分组中已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 解析生日
	var birthday *time.Time
	if req.Birthday != "" {
		parsed, err := time.Parse("2006-01-02", req.Birthday)
		if err == nil {
			birthday = &parsed
		}
	}

	// 创建客户
	customer := &models.Customer{
		GroupID:        req.GroupID,
		ActivationCode: group.ActivationCode,
		LineAccountID:  req.LineAccountID,
		PlatformType:   req.PlatformType,
		CustomerID:     req.CustomerID,
		DisplayName:    req.DisplayName,
		AvatarURL:      req.AvatarURL,
		PhoneNumber:    req.PhoneNumber,
		CustomerType:   req.CustomerType,
		Country:        req.Country,
		Birthday:       birthday,
		Address:        req.Address,
		NicknameRemark: req.NicknameRemark,
		Remark:         req.Remark,
	}

	// 处理Gender字段：如果提供了有效值，则设置；如果为空字符串，则不设置（让数据库使用NULL）
	// 数据库约束要求gender必须是'male', 'female', 'unknown'之一，或者NULL，不能是空字符串
	if req.Gender != "" {
		customer.Gender = req.Gender
	}

	// 如果Gender为空字符串，使用Omit排除它，让数据库使用NULL
	db := s.db
	if req.Gender == "" {
		db = db.Omit("gender")
	}

	if err := db.Create(customer).Error; err != nil {
		return nil, fmt.Errorf("创建客户失败: %w", err)
	}

	return customer, nil
}

// GetCustomerList 获取客户列表（带分页和筛选）
func (s *CustomerService) GetCustomerList(c *gin.Context, params *schemas.CustomerQueryParams) ([]schemas.CustomerListResponse, int64, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Customer{}), "customers")

	// 添加查询条件
	query = query.Where("deleted_at IS NULL")

	if params.GroupID != nil {
		query = query.Where("group_id = ?", *params.GroupID)
	}

	if params.LineAccountID != nil {
		query = query.Where("line_account_id = ?", *params.LineAccountID)
	}

	if params.PlatformType != "" {
		query = query.Where("platform_type = ?", params.PlatformType)
	}

	if params.CustomerType != "" {
		query = query.Where("customer_type = ?", params.CustomerType)
	}

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("customer_id LIKE ? OR display_name LIKE ?", search, search)
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

	// 查询客户列表
	var customers []models.Customer
	if err := query.
		Preload("Group").
		Preload("LineAccount").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	result := make([]schemas.CustomerListResponse, 0, len(customers))
	for _, customer := range customers {
		var birthday *string
		if customer.Birthday != nil {
			birthdayStr := customer.Birthday.Format("2006-01-02")
			birthday = &birthdayStr
		}

		lineAccountDisplayName := ""
		lineAccountLineID := ""
		if customer.LineAccount != nil {
			lineAccountDisplayName = customer.LineAccount.DisplayName
			lineAccountLineID = customer.LineAccount.LineID
		}

		groupRemark := ""
		if customer.Group != nil {
			groupRemark = customer.Group.Remark
		}

		result = append(result, schemas.CustomerListResponse{
			ID:                     customer.ID,
			GroupID:                customer.GroupID,
			ActivationCode:         customer.ActivationCode,
			LineAccountID:          customer.LineAccountID,
			PlatformType:           customer.PlatformType,
			CustomerID:             customer.CustomerID,
			DisplayName:            customer.DisplayName,
			AvatarURL:              customer.AvatarURL,
			PhoneNumber:            customer.PhoneNumber,
			CustomerType:           customer.CustomerType,
			Gender:                 customer.Gender,
			Country:                customer.Country,
			Birthday:               birthday,
			Address:                customer.Address,
			NicknameRemark:         customer.NicknameRemark,
			Remark:                 customer.Remark,
			CreatedAt:              customer.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              customer.UpdatedAt.Format(time.RFC3339),
			LineAccountDisplayName: lineAccountDisplayName,
			LineAccountLineID:      lineAccountLineID,
			GroupRemark:            groupRemark,
		})
	}

	return result, total, nil
}

// GetCustomerDetail 获取客户详情
func (s *CustomerService) GetCustomerDetail(c *gin.Context, id uint64) (*schemas.CustomerDetailResponse, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Customer{}), "customers")

	var customer models.Customer
	if err := query.
		Where("id = ? AND deleted_at IS NULL", id).
		Preload("Group").
		Preload("LineAccount").
		First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, err
	}

	var birthday *string
	if customer.Birthday != nil {
		birthdayStr := customer.Birthday.Format("2006-01-02")
		birthday = &birthdayStr
	}

	lineAccountDisplayName := ""
	lineAccountLineID := ""
	if customer.LineAccount != nil {
		lineAccountDisplayName = customer.LineAccount.DisplayName
		lineAccountLineID = customer.LineAccount.LineID
	}

	groupRemark := ""
	if customer.Group != nil {
		groupRemark = customer.Group.Remark
	}

	// 转换Tags和ProfileData
	tags := make(map[string]interface{})
	if customer.Tags != nil {
		tags = customer.Tags
	}

	profileData := make(map[string]interface{})
	if customer.ProfileData != nil {
		profileData = customer.ProfileData
	}

	return &schemas.CustomerDetailResponse{
		CustomerListResponse: schemas.CustomerListResponse{
			ID:                     customer.ID,
			GroupID:                customer.GroupID,
			ActivationCode:         customer.ActivationCode,
			LineAccountID:          customer.LineAccountID,
			PlatformType:           customer.PlatformType,
			CustomerID:             customer.CustomerID,
			DisplayName:            customer.DisplayName,
			AvatarURL:              customer.AvatarURL,
			PhoneNumber:            customer.PhoneNumber,
			CustomerType:           customer.CustomerType,
			Gender:                 customer.Gender,
			Country:                customer.Country,
			Birthday:               birthday,
			Address:                customer.Address,
			NicknameRemark:         customer.NicknameRemark,
			Remark:                 customer.Remark,
			CreatedAt:              customer.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              customer.UpdatedAt.Format(time.RFC3339),
			LineAccountDisplayName: lineAccountDisplayName,
			LineAccountLineID:      lineAccountLineID,
			GroupRemark:            groupRemark,
		},
		Tags:        tags,
		ProfileData: profileData,
	}, nil
}

// UpdateCustomer 更新客户信息
func (s *CustomerService) UpdateCustomer(c *gin.Context, id uint64, req *schemas.UpdateCustomerRequest) (*models.Customer, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Customer{}), "customers")

	var customer models.Customer
	if err := query.Where("id = ? AND deleted_at IS NULL", id).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, err
	}

	// 检查Line账号是否存在（如果提供了line_account_id）
	if req.LineAccountID != nil {
		var lineAccount models.LineAccount
		if err := s.db.Where("id = ? AND group_id = ? AND deleted_at IS NULL", *req.LineAccountID, customer.GroupID).
			First(&lineAccount).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("Line账号不存在或不属于该分组")
			}
			return nil, err
		}
		customer.LineAccountID = req.LineAccountID
	}

	// 更新字段
	if req.DisplayName != "" {
		customer.DisplayName = req.DisplayName
	}
	if req.AvatarURL != "" {
		customer.AvatarURL = req.AvatarURL
	}
	if req.PhoneNumber != "" {
		customer.PhoneNumber = req.PhoneNumber
	}
	if req.CustomerType != "" {
		customer.CustomerType = req.CustomerType
	}
	if req.Gender != "" {
		customer.Gender = req.Gender
	}
	if req.Country != "" {
		customer.Country = req.Country
	}
	if req.Birthday != "" {
		parsed, err := time.Parse("2006-01-02", req.Birthday)
		if err == nil {
			customer.Birthday = &parsed
		}
	}
	if req.Address != "" {
		customer.Address = req.Address
	}
	if req.NicknameRemark != "" {
		customer.NicknameRemark = req.NicknameRemark
	}
	if req.Remark != "" {
		customer.Remark = req.Remark
	}

	if err := s.db.Save(&customer).Error; err != nil {
		return nil, fmt.Errorf("更新客户失败: %w", err)
	}

	return &customer, nil
}

// DeleteCustomer 删除客户（软删除）
func (s *CustomerService) DeleteCustomer(c *gin.Context, id uint64) error {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.Customer{}), "customers")

	var customer models.Customer
	if err := query.Where("id = ? AND deleted_at IS NULL", id).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("客户不存在")
		}
		return err
	}

	if err := s.db.Delete(&customer).Error; err != nil {
		return fmt.Errorf("删除客户失败: %w", err)
	}

	return nil
}

// SyncCustomer 同步客户信息（从Windows客户端）
func (s *CustomerService) SyncCustomer(groupID uint, activationCode string, data *schemas.CustomerSyncData) (*models.Customer, error) {
	// 查找或创建客户
	var customer models.Customer
	err := s.db.Where("group_id = ? AND customer_id = ? AND platform_type = ? AND deleted_at IS NULL",
		groupID, data.CustomerID, data.PlatformType).First(&customer).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新客户
			customer = models.Customer{
				GroupID:        groupID,
				ActivationCode: activationCode,
				PlatformType:   data.PlatformType,
				CustomerID:     data.CustomerID,
				DisplayName:    data.DisplayName,
				AvatarURL:      data.AvatarURL,
				PhoneNumber:    data.PhoneNumber,
				Country:        data.Country,
				Address:        data.Address,
				Remark:         data.Remark,
			}

			// 处理Gender字段：如果提供了有效值，则设置；如果为空字符串，则不设置（让数据库使用NULL）
			if data.Gender != "" {
				customer.Gender = data.Gender
			}

			// 解析生日
			if data.Birthday != "" {
				parsed, err := time.Parse("2006-01-02", data.Birthday)
				if err == nil {
					customer.Birthday = &parsed
				}
			}

			// 如果Gender为空字符串，使用Omit排除它，让数据库使用NULL
			db := s.db
			if data.Gender == "" {
				db = db.Omit("gender")
			}

			if err := db.Create(&customer).Error; err != nil {
				return nil, fmt.Errorf("创建客户失败: %w", err)
			}
			logger.Infof("创建新客户: customer_id=%s, group_id=%d", data.CustomerID, groupID)
		} else {
			return nil, err
		}
	} else {
		// 更新现有客户
		updateFields := make(map[string]interface{})
		if data.DisplayName != "" {
			updateFields["display_name"] = data.DisplayName
		}
		if data.AvatarURL != "" {
			updateFields["avatar_url"] = data.AvatarURL
		}
		if data.PhoneNumber != "" {
			updateFields["phone_number"] = data.PhoneNumber
		}
		if data.Gender != "" {
			updateFields["gender"] = data.Gender
		}
		if data.Country != "" {
			updateFields["country"] = data.Country
		}
		if data.Address != "" {
			updateFields["address"] = data.Address
		}
		if data.Remark != "" {
			updateFields["remark"] = data.Remark
		}
		if data.Birthday != "" {
			parsed, err := time.Parse("2006-01-02", data.Birthday)
			if err == nil {
				updateFields["birthday"] = parsed
			}
		}

		if len(updateFields) > 0 {
			if err := s.db.Model(&customer).Updates(updateFields).Error; err != nil {
				return nil, fmt.Errorf("更新客户失败: %w", err)
			}
			logger.Debugf("更新客户: customer_id=%s, group_id=%d", data.CustomerID, groupID)
		}
	}

	return &customer, nil
}


