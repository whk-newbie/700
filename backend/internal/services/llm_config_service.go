package services

import (
	"errors"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LLMConfigService LLM配置服务
type LLMConfigService struct {
	db *gorm.DB
}

// NewLLMConfigService 创建LLM配置服务实例
func NewLLMConfigService() *LLMConfigService {
	return &LLMConfigService{
		db: database.GetDB(),
	}
}

// GetLLMConfigList 获取LLM配置列表
func (s *LLMConfigService) GetLLMConfigList(c *gin.Context, params *schemas.LLMConfigQueryParams) ([]schemas.LLMConfigResponse, int64, error) {
	var configs []models.LLMConfig
	var total int64

	query := s.db.Model(&models.LLMConfig{})

	// 提供商筛选
	if params.Provider != "" {
		query = query.Where("provider = ?", params.Provider)
	}

	// 激活状态筛选
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	// 搜索（名称）
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("name LIKE ?", searchPattern)
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
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	var list []schemas.LLMConfigResponse
	for _, config := range configs {
		list = append(list, schemas.LLMConfigResponse{
			ID:              config.ID,
			Name:            config.Name,
			Provider:        config.Provider,
			APIURL:          config.APIURL,
			Model:           config.Model,
			MaxTokens:       config.MaxTokens,
			Temperature:     config.Temperature,
			TopP:            config.TopP,
			FrequencyPenalty: config.FrequencyPenalty,
			PresencePenalty:  config.PresencePenalty,
			SystemPrompt:    config.SystemPrompt,
			TimeoutSeconds:  config.TimeoutSeconds,
			MaxRetries:      config.MaxRetries,
			IsActive:        config.IsActive,
			CreatedBy:       config.CreatedBy,
			CreatedAt:       config.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       config.UpdatedAt.Format(time.RFC3339),
		})
	}

	return list, total, nil
}

// GetLLMConfigByID 根据ID获取配置
func (s *LLMConfigService) GetLLMConfigByID(id uint) (*models.LLMConfig, error) {
	var config models.LLMConfig
	if err := s.db.Where("id = ?", id).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("配置不存在")
		}
		return nil, err
	}
	return &config, nil
}

// CreateLLMConfig 创建LLM配置
func (s *LLMConfigService) CreateLLMConfig(c *gin.Context, req *schemas.CreateLLMConfigRequest) (*models.LLMConfig, error) {
	// 获取当前用户ID（创建者）
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("无法获取当前用户信息")
	}
	createdBy := uint(userID.(uint))

	// 加密API Key
	encryptionService := GetEncryptionService()
	encryptedAPIKey, err := encryptionService.Encrypt(req.APIKey)
	if err != nil {
		logger.Errorf("加密API Key失败: %v", err)
		return nil, errors.New("加密API Key失败")
	}

	// 设置默认值
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}
	topP := req.TopP
	if topP == 0 {
		topP = 1.0
	}
	timeoutSeconds := req.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 30
	}
	maxRetries := req.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	// 创建配置
	config := &models.LLMConfig{
		Name:            req.Name,
		Provider:        req.Provider,
		APIURL:          req.APIURL,
		APIKey:          encryptedAPIKey,
		Model:           req.Model,
		MaxTokens:       maxTokens,
		Temperature:     temperature,
		TopP:            topP,
		FrequencyPenalty: req.FrequencyPenalty,
		PresencePenalty:  req.PresencePenalty,
		SystemPrompt:    req.SystemPrompt,
		TimeoutSeconds:  timeoutSeconds,
		MaxRetries:      maxRetries,
		IsActive:        req.IsActive,
		CreatedBy:       &createdBy,
	}

	if err := s.db.Create(config).Error; err != nil {
		logger.Errorf("创建LLM配置失败: %v", err)
		return nil, errors.New("创建LLM配置失败")
	}

	return config, nil
}

// UpdateLLMConfig 更新LLM配置
func (s *LLMConfigService) UpdateLLMConfig(c *gin.Context, id uint, req *schemas.UpdateLLMConfigRequest) (*models.LLMConfig, error) {
	var config models.LLMConfig
	if err := s.db.Where("id = ?", id).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("配置不存在")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Provider != "" {
		updates["provider"] = req.Provider
	}
	if req.APIURL != "" {
		updates["api_url"] = req.APIURL
	}
	if req.APIKey != "" {
		// 加密新API Key
		encryptionService := GetEncryptionService()
		encryptedAPIKey, err := encryptionService.Encrypt(req.APIKey)
		if err != nil {
			logger.Errorf("加密API Key失败: %v", err)
			return nil, errors.New("加密API Key失败")
		}
		updates["api_key"] = encryptedAPIKey
	}
	if req.Model != "" {
		updates["model"] = req.Model
	}
	if req.MaxTokens != nil {
		updates["max_tokens"] = *req.MaxTokens
	}
	if req.Temperature != nil {
		updates["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		updates["top_p"] = *req.TopP
	}
	if req.FrequencyPenalty != nil {
		updates["frequency_penalty"] = *req.FrequencyPenalty
	}
	if req.PresencePenalty != nil {
		updates["presence_penalty"] = *req.PresencePenalty
	}
	if req.SystemPrompt != "" {
		updates["system_prompt"] = req.SystemPrompt
	}
	if req.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *req.TimeoutSeconds
	}
	if req.MaxRetries != nil {
		updates["max_retries"] = *req.MaxRetries
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&config).Updates(updates).Error; err != nil {
			logger.Errorf("更新LLM配置失败: %v", err)
			return nil, errors.New("更新LLM配置失败")
		}
	}

	// 重新查询配置
	if err := s.db.Where("id = ?", id).First(&config).Error; err != nil {
		return nil, err
	}

	return &config, nil
}

// DeleteLLMConfig 删除LLM配置
func (s *LLMConfigService) DeleteLLMConfig(c *gin.Context, id uint) error {
	var config models.LLMConfig
	if err := s.db.Where("id = ?", id).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("配置不存在")
		}
		return err
	}

	// 检查是否有模板关联
	var templateCount int64
	if err := s.db.Model(&models.LLMPromptTemplate{}).
		Where("config_id = ?", id).
		Count(&templateCount).Error; err != nil {
		return err
	}
	if templateCount > 0 {
		return errors.New("该配置下还有模板，无法删除")
	}

	// 删除配置
	if err := s.db.Delete(&config).Error; err != nil {
		logger.Errorf("删除LLM配置失败: %v", err)
		return errors.New("删除LLM配置失败")
	}

	return nil
}

// TestLLMConfig 测试LLM配置连接
func (s *LLMConfigService) TestLLMConfig(id uint) error {
	config, err := s.GetLLMConfigByID(id)
	if err != nil {
		return err
	}

	if !config.IsActive {
		return errors.New("配置未激活")
	}

	// 获取提供商
	provider := GetLLMProvider(config.Provider)
	if provider == nil {
		return errors.New("不支持的提供商")
	}

	// 测试连接
	return provider.TestConnection(config)
}

