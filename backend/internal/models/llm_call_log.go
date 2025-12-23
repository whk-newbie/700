package models

import (
	"time"
)

// LLMCallLog 大模型调用日志模型
type LLMCallLog struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigID         *uint          `gorm:"type:integer" json:"config_id"`
	TemplateID       *uint          `gorm:"type:integer" json:"template_id"`
	GroupID          *uint          `gorm:"type:integer" json:"group_id"`
	ActivationCode   string         `gorm:"type:varchar(32)" json:"activation_code"`
	RequestMessages  JSONB          `gorm:"type:jsonb;not null" json:"request_messages"`
	RequestParams    JSONB          `gorm:"type:jsonb" json:"request_params"`
	ResponseContent  string         `gorm:"type:text" json:"response_content"`
	ResponseData     JSONB          `gorm:"type:jsonb" json:"response_data"`
	Status           string         `gorm:"type:varchar(20);not null" json:"status"`
	ErrorMessage     string         `gorm:"type:text" json:"error_message"`
	TokensUsed       *int           `gorm:"type:integer" json:"tokens_used"`
	PromptTokens     *int           `gorm:"type:integer" json:"prompt_tokens"`
	CompletionTokens *int           `gorm:"type:integer" json:"completion_tokens"`
	CallTime         time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"call_time"`
	DurationMs       *int           `gorm:"type:integer" json:"duration_ms"`
}

// TableName 指定表名
func (LLMCallLog) TableName() string {
	return "llm_call_logs"
}

