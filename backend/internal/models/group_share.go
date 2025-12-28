package models

import (
	"time"

	"gorm.io/gorm"
)

// GroupShare 分组分享模型
type GroupShare struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GroupID   uint           `gorm:"type:integer;not null;index" json:"group_id"`
	ShareCode string         `gorm:"type:varchar(16);uniqueIndex;not null" json:"share_code"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"` // 访问密码，不返回给前端
	ExpiresAt *time.Time     `json:"expires_at"` // 过期时间，为空表示永久有效
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	ViewCount int            `gorm:"default:0" json:"view_count"` // 访问次数统计
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 关联关系
	Group *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

// TableName 指定表名
func (GroupShare) TableName() string {
	return "group_shares"
}

