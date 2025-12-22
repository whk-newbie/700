package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Email        string         `gorm:"type:varchar(100)" json:"email"`
	Role         string         `gorm:"type:varchar(20);not null;default:'user';check:role IN ('admin', 'user')" json:"role"`
	MaxGroups    *int           `gorm:"type:integer" json:"max_groups"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedBy    *uint          `gorm:"type:integer" json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsAdmin 判断是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsUser 判断是否为普通用户
func (u *User) IsUser() bool {
	return u.Role == "user"
}

