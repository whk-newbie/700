package services

import (
	"errors"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/internal/utils"
	"line-management/pkg/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{
		db: database.GetDB(),
	}
}

// Login 用户登录（管理员/普通用户）
func (s *AuthService) Login(req *schemas.LoginRequest) (*schemas.LoginResponse, error) {
	var user models.User

	// 查找用户
	if err := s.db.Where("username = ? AND deleted_at IS NULL", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成Token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &schemas.LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: 24 * 3600, // 24小时
		User: &schemas.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	return response, nil
}

// LoginSubAccount 子账号登录
func (s *AuthService) LoginSubAccount(req *schemas.SubAccountLoginRequest) (*schemas.LoginResponse, error) {
	var group models.Group

	// 查找分组（激活码）
	if err := s.db.Where("activation_code = ? AND deleted_at IS NULL", req.ActivationCode).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("激活码或密码错误")
		}
		return nil, err
	}

	// 检查分组是否激活
	if !group.IsActive {
		return nil, errors.New("分组已被禁用")
	}

	// 验证密码
	if group.LoginPassword == "" {
		return nil, errors.New("该分组未设置登录密码")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(group.LoginPassword), []byte(req.Password)); err != nil {
		return nil, errors.New("激活码或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	group.LastLoginAt = &now
	s.db.Save(&group)

	// 生成Token
	token, err := utils.GenerateSubAccountToken(group.ID, group.ActivationCode)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &schemas.LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: 24 * 3600, // 24小时
		Group: &schemas.GroupInfo{
			ID:             group.ID,
			ActivationCode: group.ActivationCode,
			Category:       group.Category,
		},
	}

	return response, nil
}

// GetUserByID 根据ID获取用户信息
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetGroupByID 根据ID获取分组信息
func (s *AuthService) GetGroupByID(groupID uint) (*models.Group, error) {
	var group models.Group
	if err := s.db.Where("id = ? AND deleted_at IS NULL", groupID).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}


