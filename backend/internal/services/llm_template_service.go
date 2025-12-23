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

// LLMTemplateService Prompt模板服务
type LLMTemplateService struct {
	db *gorm.DB
}

// NewLLMTemplateService 创建Prompt模板服务实例
func NewLLMTemplateService() *LLMTemplateService {
	return &LLMTemplateService{
		db: database.GetDB(),
	}
}

// GetTemplateList 获取模板列表
func (s *LLMTemplateService) GetTemplateList(c *gin.Context, params *schemas.PromptTemplateQueryParams) ([]schemas.PromptTemplateResponse, int64, error) {
	var templates []models.LLMPromptTemplate
	var total int64

	query := s.db.Model(&models.LLMPromptTemplate{})

	// 配置ID筛选
	if params.ConfigID != nil {
		query = query.Where("config_id = ?", *params.ConfigID)
	}

	// 激活状态筛选
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	// 搜索（模板名称）
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("template_name LIKE ?", searchPattern)
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
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	var list []schemas.PromptTemplateResponse
	for _, template := range templates {
		variables := make(map[string]interface{})
		if template.Variables != nil {
			variables = template.Variables
		}

		list = append(list, schemas.PromptTemplateResponse{
			ID:             template.ID,
			ConfigID:       template.ConfigID,
			TemplateName:   template.TemplateName,
			TemplateContent: template.TemplateContent,
			Variables:      variables,
			Description:    template.Description,
			IsActive:       template.IsActive,
			CreatedAt:      template.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      template.UpdatedAt.Format(time.RFC3339),
		})
	}

	return list, total, nil
}

// GetTemplateByID 根据ID获取模板
func (s *LLMTemplateService) GetTemplateByID(id uint) (*models.LLMPromptTemplate, error) {
	var template models.LLMPromptTemplate
	if err := s.db.Where("id = ?", id).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模板不存在")
		}
		return nil, err
	}
	return &template, nil
}

// CreateTemplate 创建模板
func (s *LLMTemplateService) CreateTemplate(c *gin.Context, req *schemas.CreatePromptTemplateRequest) (*models.LLMPromptTemplate, error) {
	// 检查配置是否存在
	var config models.LLMConfig
	if err := s.db.Where("id = ?", req.ConfigID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("配置不存在")
		}
		return nil, err
	}

	// 创建模板
	template := &models.LLMPromptTemplate{
		ConfigID:       req.ConfigID,
		TemplateName:   req.TemplateName,
		TemplateContent: req.TemplateContent,
		Variables:      models.JSONB(req.Variables),
		Description:    req.Description,
		IsActive:       req.IsActive,
	}

	if err := s.db.Create(template).Error; err != nil {
		logger.Errorf("创建模板失败: %v", err)
		return nil, errors.New("创建模板失败")
	}

	return template, nil
}

// UpdateTemplate 更新模板
func (s *LLMTemplateService) UpdateTemplate(c *gin.Context, id uint, req *schemas.UpdatePromptTemplateRequest) (*models.LLMPromptTemplate, error) {
	var template models.LLMPromptTemplate
	if err := s.db.Where("id = ?", id).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模板不存在")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.TemplateName != "" {
		updates["template_name"] = req.TemplateName
	}
	if req.TemplateContent != "" {
		updates["template_content"] = req.TemplateContent
	}
	if req.Variables != nil {
		updates["variables"] = models.JSONB(req.Variables)
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&template).Updates(updates).Error; err != nil {
			logger.Errorf("更新模板失败: %v", err)
			return nil, errors.New("更新模板失败")
		}
	}

	// 重新查询模板
	if err := s.db.Where("id = ?", id).First(&template).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

// DeleteTemplate 删除模板
func (s *LLMTemplateService) DeleteTemplate(c *gin.Context, id uint) error {
	var template models.LLMPromptTemplate
	if err := s.db.Where("id = ?", id).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("模板不存在")
		}
		return err
	}

	// 删除模板
	if err := s.db.Delete(&template).Error; err != nil {
		logger.Errorf("删除模板失败: %v", err)
		return errors.New("删除模板失败")
	}

	return nil
}

