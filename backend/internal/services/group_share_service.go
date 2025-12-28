package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"line-management/internal/models"
	"line-management/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GroupShareService 分组分享服务
type GroupShareService struct{}

// NewGroupShareService 创建分组分享服务实例
func NewGroupShareService() *GroupShareService {
	return &GroupShareService{}
}

// generateShareCode 生成随机分享码
func (s *GroupShareService) generateShareCode() (string, error) {
	bytes := make([]byte, 4) // 4字节 = 8个十六进制字符
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateGroupShare 创建分组分享
func (s *GroupShareService) CreateGroupShare(c *gin.Context, groupID uint, expiresAt *time.Time) (*models.GroupShare, error) {
	db := database.GetDB()

	// 检查分组是否存在
	var group models.Group
	if err := db.First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分组不存在")
		}
		return nil, err
	}

	// 检查是否已经存在有效的分享（一个分组只保留一个有效分享）
	var existingShare models.GroupShare
	err := db.Where("group_id = ? AND is_active = ? AND deleted_at IS NULL", groupID, true).First(&existingShare).Error
	if err == nil {
		// 如果已存在有效分享，直接返回
		return &existingShare, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 生成唯一的分享码
	var shareCode string
	for {
		shareCode, err = s.generateShareCode()
		if err != nil {
			return nil, err
		}

		// 检查是否重复
		var count int64
		db.Model(&models.GroupShare{}).Where("share_code = ? AND deleted_at IS NULL", shareCode).Count(&count)
		if count == 0 {
			break
		}
	}

	// 默认密码为分享码本身
	password := shareCode

	// 创建分享记录
	share := &models.GroupShare{
		GroupID:   groupID,
		ShareCode: shareCode,
		Password:  password,
		ExpiresAt: expiresAt,
		IsActive:  true,
		ViewCount: 0,
	}

	if err := db.Create(share).Error; err != nil {
		return nil, err
	}

	return share, nil
}

// GetGroupShareByCode 通过分享码获取分组分享信息
func (s *GroupShareService) GetGroupShareByCode(c *gin.Context, shareCode string) (*models.GroupShare, error) {
	db := database.GetDB()

	var share models.GroupShare
	err := db.Preload("Group").Where("share_code = ? AND is_active = ? AND deleted_at IS NULL", shareCode, true).First(&share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分享不存在或已失效")
		}
		return nil, err
	}

	// 检查是否过期
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("分享已过期")
	}

	// 增加访问次数
	db.Model(&share).Update("view_count", gorm.Expr("view_count + 1"))

	return &share, nil
}

// VerifySharePassword 验证分享密码
func (s *GroupShareService) VerifySharePassword(c *gin.Context, shareCode, password string) (*models.GroupShare, error) {
	db := database.GetDB()

	var share models.GroupShare
	err := db.Preload("Group").Where("share_code = ? AND is_active = ? AND deleted_at IS NULL", shareCode, true).First(&share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分享不存在或已失效")
		}
		return nil, err
	}

	// 检查是否过期
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("分享已过期")
	}

	// 验证密码（简单比较，默认密码就是分享码）
	if share.Password != password {
		return nil, errors.New("密码错误")
	}

	// 增加访问次数
	db.Model(&share).Update("view_count", gorm.Expr("view_count + 1"))

	return &share, nil
}

// GetGroupShareByGroupID 获取分组的分享信息
func (s *GroupShareService) GetGroupShareByGroupID(c *gin.Context, groupID uint) (*models.GroupShare, error) {
	db := database.GetDB()

	var share models.GroupShare
	err := db.Where("group_id = ? AND is_active = ? AND deleted_at IS NULL", groupID, true).First(&share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分享不存在")
		}
		return nil, err
	}

	return &share, nil
}

// DeleteGroupShare 删除分组分享
func (s *GroupShareService) DeleteGroupShare(c *gin.Context, shareID uint) error {
	db := database.GetDB()

	var share models.GroupShare
	if err := db.First(&share, shareID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分享不存在")
		}
		return err
	}

	// 软删除
	return db.Delete(&share).Error
}

// DisableGroupShare 禁用分组分享
func (s *GroupShareService) DisableGroupShare(c *gin.Context, shareID uint) error {
	db := database.GetDB()

	var share models.GroupShare
	if err := db.First(&share, shareID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分享不存在")
		}
		return err
	}

	return db.Model(&share).Update("is_active", false).Error
}

