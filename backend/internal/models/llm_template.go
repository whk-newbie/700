package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONB 用于处理PostgreSQL的JSONB类型
type JSONB map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

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

