package models

import (
	"time"
)

// LLMPromptTemplate Prompt模板模型
type LLMPromptTemplate struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ConfigID       uint           `gorm:"type:integer;not null" json:"config_id"`
	TemplateName   string         `gorm:"type:varchar(100);not null" json:"template_name"`
	TemplateContent string        `gorm:"type:text;not null" json:"template_content"`
	Variables      JSONB          `gorm:"type:jsonb" json:"variables"`
	Description    string         `gorm:"type:text" json:"description"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (LLMPromptTemplate) TableName() string {
	return "llm_prompt_templates"
}

