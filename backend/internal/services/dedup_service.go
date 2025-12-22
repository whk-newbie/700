package services

import (
	"line-management/internal/models"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"gorm.io/gorm"
)

// DedupService 去重服务
type DedupService struct {
	db *gorm.DB
}

// NewDedupService 创建去重服务实例
func NewDedupService() *DedupService {
	return &DedupService{
		db: database.GetDB(),
	}
}

// CheckDuplicateCurrent 检查当前分组内是否重复
// 在当前分组的所有账号中，检查该incoming_line_id是否已经存在
func (s *DedupService) CheckDuplicateCurrent(groupID uint, incomingLineID string) (bool, error) {
	var count int64
	
	// 检查在当前分组的所有进线记录中是否已经存在该incoming_line_id
	err := s.db.Model(&models.IncomingLog{}).
		Where("group_id = ? AND incoming_line_id = ?", groupID, incomingLineID).
		Count(&count).Error
	
	if err != nil {
		logger.Errorf("检查当前分组重复失败: %v", err)
		return false, err
	}
	
	return count > 0, nil
}

// CheckDuplicateGlobal 检查全局是否重复
// 在所有分组的所有账号中，检查该incoming_line_id是否已经存在
func (s *DedupService) CheckDuplicateGlobal(incomingLineID string) (bool, error) {
	var count int64
	
	// 检查在所有进线记录中是否已经存在该incoming_line_id
	err := s.db.Model(&models.IncomingLog{}).
		Where("incoming_line_id = ?", incomingLineID).
		Count(&count).Error
	
	if err != nil {
		logger.Errorf("检查全局重复失败: %v", err)
		return false, err
	}
	
	return count > 0, nil
}

// CheckDuplicate 根据分组配置检查是否重复
// dedup_scope: 'current' 检查当前分组, 'global' 检查全局
func (s *DedupService) CheckDuplicate(groupID uint, incomingLineID string, dedupScope string) (bool, string, error) {
	if dedupScope == "global" {
		isDuplicate, err := s.CheckDuplicateGlobal(incomingLineID)
		if err != nil {
			return false, "", err
		}
		return isDuplicate, "global", nil
	} else {
		// 默认使用current
		isDuplicate, err := s.CheckDuplicateCurrent(groupID, incomingLineID)
		if err != nil {
			return false, "", err
		}
		return isDuplicate, "current", nil
	}
}

// CheckContactPoolDuplicate 检查底库中是否已存在
func (s *DedupService) CheckContactPoolDuplicate(lineID string, platformType string) (bool, error) {
	var count int64
	
	err := s.db.Model(&models.ContactPool{}).
		Where("line_id = ? AND platform_type = ? AND deleted_at IS NULL", lineID, platformType).
		Count(&count).Error
	
	if err != nil {
		logger.Errorf("检查底库重复失败: %v", err)
		return false, err
	}
	
	return count > 0, nil
}

