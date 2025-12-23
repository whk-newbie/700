package services

import (
	"errors"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		db: database.GetDB(),
	}
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList(c *gin.Context, params *schemas.UserQueryParams) ([]schemas.UserListResponse, int64, error) {
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{}).Where("deleted_at IS NULL")

	// 角色筛选
	if params.Role != "" {
		query = query.Where("role = ?", params.Role)
	}

	// 激活状态筛选
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	// 搜索（用户名或邮箱）
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("username LIKE ? OR email LIKE ?", searchPattern, searchPattern)
	}

	// 获取总数
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
	offset := (page - 1) * pageSize

	// 查询列表
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	var list []schemas.UserListResponse
	for _, user := range users {
		list = append(list, schemas.UserListResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			MaxGroups: user.MaxGroups,
			IsActive:  user.IsActive,
			CreatedBy: user.CreatedBy,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		})
	}

	return list, total, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(c *gin.Context, req *schemas.CreateUserRequest) (*models.User, error) {
	// 获取当前用户ID（创建者）
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("无法获取当前用户信息")
	}
	createdBy := uint(userID.(uint))

	// 检查用户名是否已存在
	var count int64
	if err := s.db.Model(&models.User{}).
		Where("username = ? AND deleted_at IS NULL", req.Username).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("密码加密失败: %v", err)
		return nil, errors.New("密码加密失败")
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Email:        req.Email,
		Role:         req.Role,
		MaxGroups:    req.MaxGroups,
		IsActive:     req.IsActive,
		CreatedBy:    &createdBy,
	}

	if err := s.db.Create(user).Error; err != nil {
		logger.Errorf("创建用户失败: %v", err)
		return nil, errors.New("创建用户失败")
	}

	return user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(c *gin.Context, userID uint, req *schemas.UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.MaxGroups != nil {
		updates["max_groups"] = req.MaxGroups
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Password != "" {
		// 加密新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Errorf("密码加密失败: %v", err)
			return nil, errors.New("密码加密失败")
		}
		updates["password_hash"] = string(hashedPassword)
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			logger.Errorf("更新用户失败: %v", err)
			return nil, errors.New("更新用户失败")
		}
	}

	// 重新查询用户
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(c *gin.Context, userID uint) error {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 检查是否有分组关联
	var groupCount int64
	if err := s.db.Model(&models.Group{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&groupCount).Error; err != nil {
		return err
	}
	if groupCount > 0 {
		return errors.New("该用户下还有分组，无法删除")
	}

	// 软删除
	if err := s.db.Delete(&user).Error; err != nil {
		logger.Errorf("删除用户失败: %v", err)
		return errors.New("删除用户失败")
	}

	return nil
}

