package models

import (
	"time"
)

// LLMConfig 大模型配置模型（简化版，只保留OpenAI API Key）
type LLMConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	APIKey    string    `gorm:"type:text;not null" json:"-"` // 不返回给前端
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (LLMConfig) TableName() string {
	return "llm_configs"
}

