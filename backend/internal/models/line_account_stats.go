package models

import (
	"time"
)

// LineAccountStats Line账号统计模型
type LineAccountStats struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	LineAccountID     uint       `gorm:"type:integer;not null;uniqueIndex" json:"line_account_id"`
	TodayIncoming     int        `gorm:"type:integer;default:0" json:"today_incoming"`
	TotalIncoming     int        `gorm:"type:integer;default:0" json:"total_incoming"`
	DuplicateIncoming int        `gorm:"type:integer;default:0" json:"duplicate_incoming"`
	TodayDuplicate    int        `gorm:"type:integer;default:0" json:"today_duplicate"`
	LastResetDate     *time.Time `gorm:"type:date" json:"last_reset_date"`
	LastResetTime     *time.Time `gorm:"type:timestamp" json:"last_reset_time"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// 关联关系
	LineAccount *LineAccount `gorm:"foreignKey:LineAccountID" json:"line_account,omitempty"`
}

// TableName 指定表名
func (LineAccountStats) TableName() string {
	return "line_account_stats"
}

