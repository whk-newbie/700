package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// IncomingLog 进线日志模型（分区表）
type IncomingLog struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	LineAccountID uint           `gorm:"type:integer;not null;index:idx_incoming_logs_line_account" json:"line_account_id"`
	GroupID       uint           `gorm:"type:integer;not null;index:idx_incoming_logs_group_id" json:"group_id"`
	IncomingLineID string        `gorm:"type:varchar(100);not null;index:idx_incoming_logs_incoming_line_id" json:"incoming_line_id"`
	IncomingTime  time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;primaryKey;index:idx_incoming_logs_time" json:"incoming_time"`
	DisplayName   string         `gorm:"type:varchar(100)" json:"display_name"`
	AvatarURL     string         `gorm:"type:varchar(500)" json:"avatar_url"`
	PhoneNumber   string         `gorm:"type:varchar(20)" json:"phone_number"`
	IsDuplicate   bool           `gorm:"type:boolean;default:false;index:idx_incoming_logs_duplicate" json:"is_duplicate"`
	DuplicateScope string        `gorm:"type:varchar(20)" json:"duplicate_scope"` // 'current' or 'global'
	CustomerType  string         `gorm:"type:varchar(50)" json:"customer_type"`
	RawData       JSONB          `gorm:"type:jsonb" json:"raw_data,omitempty"`

	// 关联关系
	LineAccount *LineAccount `gorm:"foreignKey:LineAccountID" json:"line_account,omitempty"`
	Group       *Group       `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

// TableName 指定表名
func (IncomingLog) TableName() string {
	return "incoming_logs"
}

// JSONB 自定义JSONB类型
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

