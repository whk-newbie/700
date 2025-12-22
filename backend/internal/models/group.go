package models

import (
	"time"

	"gorm.io/gorm"
)

// Group 分组模型
type Group struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"type:integer;not null;index" json:"user_id"`
	ActivationCode string        `gorm:"type:varchar(32);uniqueIndex;not null" json:"activation_code"`
	AccountLimit  *int           `gorm:"type:integer" json:"account_limit"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	Remark        string         `gorm:"type:varchar(255)" json:"remark"`
	Description   string         `gorm:"type:text" json:"description"`
	Category      string         `gorm:"type:varchar(50);default:'default';index" json:"category"`
	DedupScope    string         `gorm:"type:varchar(20);default:'current';check:dedup_scope IN ('current', 'global')" json:"dedup_scope"`
	ResetTime     string         `gorm:"type:time;default:'09:00:00'" json:"reset_time"`
	LoginPassword string         `gorm:"type:varchar(255)" json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	LastLoginAt   *time.Time     `json:"last_login_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (Group) TableName() string {
	return "groups"
}

