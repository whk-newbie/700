package services

import (
	"errors"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"gorm.io/gorm"
)

// LLMConfigService LLM配置服务（简化版，只处理OpenAI API Key）
type LLMConfigService struct {
	db *gorm.DB
}

// NewLLMConfigService 创建LLM配置服务实例
func NewLLMConfigService() *LLMConfigService {
	return &LLMConfigService{
		db: database.GetDB(),
	}
}

// GetOpenAIAPIKey 获取OpenAI API Key配置（只返回一条记录，如果没有则创建）
func (s *LLMConfigService) GetOpenAIAPIKey() (*models.LLMConfig, error) {
	var config models.LLMConfig
	
	// 尝试获取第一条记录（应该只有一条）
	result := s.db.First(&config)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果没有记录，创建一条空记录
			config = models.LLMConfig{
				APIKey:    "",
				UpdatedAt: time.Now(),
			}
			if err := s.db.Create(&config).Error; err != nil {
				return nil, err
			}
			return &config, nil
		}
		return nil, result.Error
	}
	
	return &config, nil
}

// UpdateOpenAIAPIKey 更新OpenAI API Key（已废弃，使用UpdateOpenAIAPIKeyWithPlainText）
func (s *LLMConfigService) UpdateOpenAIAPIKey(req *schemas.UpdateOpenAIAPIKeyRequest) (*models.LLMConfig, error) {
	return s.UpdateOpenAIAPIKeyWithPlainText("")
}

// UpdateOpenAIAPIKeyWithPlainText 使用明文API Key更新（内部使用，已通过RSA解密）
func (s *LLMConfigService) UpdateOpenAIAPIKeyWithPlainText(plainTextAPIKey string) (*models.LLMConfig, error) {
	// 获取或创建配置
	config, err := s.GetOpenAIAPIKey()
	if err != nil {
		return nil, err
	}

	// 使用AES加密API Key存储到数据库
	encryptionService := GetEncryptionService()
	encryptedAPIKey, err := encryptionService.Encrypt(plainTextAPIKey)
	if err != nil {
		logger.Errorf("加密API Key失败: %v", err)
		return nil, errors.New("加密API Key失败")
	}

	// 更新API Key
	config.APIKey = encryptedAPIKey
	config.UpdatedAt = time.Now()
	
	if err := s.db.Save(config).Error; err != nil {
		logger.Errorf("更新OpenAI API Key失败: %v", err)
		return nil, errors.New("更新OpenAI API Key失败")
	}

	return config, nil
}
