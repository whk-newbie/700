package models

import (
	"time"

	"gorm.io/gorm"
)

// Customer 客户模型
type Customer struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID       uint           `gorm:"type:integer;not null;index" json:"group_id"`
	ActivationCode string        `gorm:"type:varchar(32);not null;index" json:"activation_code"`
	LineAccountID *uint          `gorm:"type:integer;index" json:"line_account_id"`
	PlatformType  string         `gorm:"type:varchar(20);not null;check:platform_type IN ('line', 'line_business')" json:"platform_type"`
	CustomerID    string         `gorm:"type:varchar(100);not null;index" json:"customer_id"`
	DisplayName   string         `gorm:"type:varchar(100)" json:"display_name"`
	AvatarURL     string         `gorm:"type:varchar(500)" json:"avatar_url"`
	PhoneNumber   string         `gorm:"type:varchar(20)" json:"phone_number"`
	CustomerType  string         `gorm:"type:varchar(50)" json:"customer_type"`
	Gender        string         `gorm:"type:varchar(10);check:gender IN ('male', 'female', 'unknown')" json:"gender"`
	Country       string         `gorm:"type:varchar(50)" json:"country"`
	Birthday      *time.Time     `gorm:"type:date" json:"birthday,omitempty"`
	Address       string         `gorm:"type:text" json:"address"`
	NicknameRemark string        `gorm:"type:varchar(20)" json:"nickname_remark"`
	Remark        string         `gorm:"type:text" json:"remark"`
	Tags          JSONB          `gorm:"type:jsonb" json:"tags,omitempty"`
	ProfileData   JSONB          `gorm:"type:jsonb" json:"profile_data,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 关联关系
	Group       *Group       `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	LineAccount *LineAccount `gorm:"foreignKey:LineAccountID" json:"line_account,omitempty"`
}

// TableName 指定表名
func (Customer) TableName() string {
	return "customers"
}

