package models

import (
	"time"
)

// LLMConfig 大模型配置模型
type LLMConfig struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(100);not null" json:"name"`
	Provider        string         `gorm:"type:varchar(50);not null" json:"provider"`
	APIURL          string         `gorm:"type:varchar(500);not null" json:"api_url"`
	APIKey          string         `gorm:"type:text;not null" json:"-"` // 不返回给前端
	Model           string         `gorm:"type:varchar(100);not null" json:"model"`
	MaxTokens       int            `gorm:"type:integer;default:2000" json:"max_tokens"`
	Temperature     float64        `gorm:"type:decimal(3,2);default:0.7" json:"temperature"`
	TopP            float64        `gorm:"type:decimal(3,2);default:1.0" json:"top_p"`
	FrequencyPenalty float64       `gorm:"type:decimal(3,2);default:0.0" json:"frequency_penalty"`
	PresencePenalty  float64       `gorm:"type:decimal(3,2);default:0.0" json:"presence_penalty"`
	SystemPrompt    string         `gorm:"type:text" json:"system_prompt"`
	TimeoutSeconds  int            `gorm:"type:integer;default:30" json:"timeout_seconds"`
	MaxRetries      int            `gorm:"type:integer;default:3" json:"max_retries"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedBy       *uint          `gorm:"type:integer" json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (LLMConfig) TableName() string {
	return "llm_configs"
}

