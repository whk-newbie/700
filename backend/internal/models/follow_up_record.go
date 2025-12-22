package models

import (
	"time"

	"gorm.io/gorm"
)

// FollowUpRecord 跟进记录模型
type FollowUpRecord struct {
	ID                    uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID               uint           `gorm:"type:integer;not null;index" json:"group_id"`
	ActivationCode        string         `gorm:"type:varchar(32);not null" json:"activation_code"`
	LineAccountID         *uint          `gorm:"type:integer;index" json:"line_account_id"`
	CustomerID            *uint64        `gorm:"type:bigint;index" json:"customer_id"`
	PlatformType          string         `gorm:"type:varchar(20);not null;check:platform_type IN ('line', 'line_business')" json:"platform_type"`
	LineAccountDisplayName string        `gorm:"type:varchar(100)" json:"line_account_display_name"`
	LineAccountLineID     string         `gorm:"type:varchar(100)" json:"line_account_line_id"`
	LineAccountAvatarURL   string         `gorm:"type:varchar(500)" json:"line_account_avatar_url"`
	CustomerDisplayName    string         `gorm:"type:varchar(100)" json:"customer_display_name"`
	CustomerLineID         string         `gorm:"type:varchar(100)" json:"customer_line_id"`
	CustomerAvatarURL      string         `gorm:"type:varchar(500)" json:"customer_avatar_url"`
	Content                string         `gorm:"type:text;not null" json:"content"`
	CreatedBy              *uint          `gorm:"type:integer" json:"created_by"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 关联关系
	Group       *Group       `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	LineAccount *LineAccount `gorm:"foreignKey:LineAccountID" json:"line_account,omitempty"`
	Customer    *Customer    `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	User        *User         `gorm:"foreignKey:CreatedBy" json:"user,omitempty"`
}

// TableName 指定表名
func (FollowUpRecord) TableName() string {
	return "follow_up_records"
}

