package models

import (
	"time"

	"gorm.io/gorm"
)

// LineAccount Line账号模型
type LineAccount struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	GroupID         uint           `gorm:"type:integer;not null;index" json:"group_id"`
	ActivationCode  string         `gorm:"type:varchar(32);not null;index" json:"activation_code"`
	PlatformType    string         `gorm:"type:varchar(20);not null;default:'line';check:platform_type IN ('line', 'line_business')" json:"platform_type"`
	LineID          string         `gorm:"type:varchar(100);not null;index" json:"line_id"`
	DisplayName     string         `gorm:"type:varchar(100)" json:"display_name"`
	PhoneNumber     string         `gorm:"type:varchar(20)" json:"phone_number"`
	ProfileURL      string         `gorm:"type:varchar(500)" json:"profile_url"`
	AvatarURL       string         `gorm:"type:varchar(500)" json:"avatar_url"`
	Bio             string         `gorm:"type:text" json:"bio"`
	StatusMessage   string         `gorm:"type:varchar(255)" json:"status_message"`
	AddFriendLink   string         `gorm:"type:varchar(500)" json:"add_friend_link"`
	QRCodePath      string         `gorm:"type:varchar(255)" json:"qr_code_path"`
	OnlineStatus    string         `gorm:"type:varchar(20);default:'offline';check:online_status IN ('online', 'offline', 'user_logout', 'abnormal_offline')" json:"online_status"`
	ResetTime       *string        `gorm:"type:time" json:"reset_time"` // 账号重置时间，为空时使用分组的重置时间
	LastActiveAt    *time.Time     `gorm:"type:timestamp" json:"last_active_at"`
	LastOnlineTime  *time.Time     `gorm:"type:timestamp" json:"last_online_time"`
	FirstLoginAt    *time.Time     `gorm:"type:timestamp" json:"first_login_at"`
	AccountRemark   string         `gorm:"type:text" json:"account_remark"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DeletedBy       *uint          `gorm:"type:integer" json:"deleted_by"`

	// 关联关系
	Group *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

// TableName 指定表名
func (LineAccount) TableName() string {
	return "line_accounts"
}

