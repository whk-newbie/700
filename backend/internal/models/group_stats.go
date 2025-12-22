package models

import (
	"time"
)

// GroupStats 分组统计模型
type GroupStats struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	GroupID             uint           `gorm:"type:integer;not null;uniqueIndex" json:"group_id"`
	TotalAccounts       int            `gorm:"type:integer;default:0" json:"total_accounts"`
	OnlineAccounts      int            `gorm:"type:integer;default:0" json:"online_accounts"`
	LineAccounts        int            `gorm:"type:integer;default:0" json:"line_accounts"`
	LineBusinessAccounts int           `gorm:"type:integer;default:0" json:"line_business_accounts"`
	TodayIncoming       int            `gorm:"type:integer;default:0" json:"today_incoming"`
	TotalIncoming       int            `gorm:"type:integer;default:0" json:"total_incoming"`
	DuplicateIncoming   int            `gorm:"type:integer;default:0" json:"duplicate_incoming"`
	TodayDuplicate      int            `gorm:"type:integer;default:0" json:"today_duplicate"`
	LastResetDate       *time.Time     `gorm:"type:date" json:"last_reset_date"`
	LastResetTime       *time.Time     `gorm:"type:timestamp" json:"last_reset_time"`
	UpdatedAt           time.Time      `json:"updated_at"`
	
	// 关联关系
	Group *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

// TableName 指定表名
func (GroupStats) TableName() string {
	return "group_stats"
}

