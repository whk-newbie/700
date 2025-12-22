package models

import (
	"time"

	"gorm.io/gorm"
)

// ContactPool 底库模型
type ContactPool struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceType    string         `gorm:"type:varchar(20);not null;check:source_type IN ('import', 'platform')" json:"source_type"`
	ImportBatchID *uint          `gorm:"type:integer" json:"import_batch_id"`
	GroupID       uint           `gorm:"type:integer;not null;index" json:"group_id"`
	ActivationCode string        `gorm:"type:varchar(32);not null;index" json:"activation_code"`
	LineAccountID *uint          `gorm:"type:integer" json:"line_account_id"`
	PlatformType  string         `gorm:"type:varchar(20);not null;check:platform_type IN ('line', 'line_business')" json:"platform_type"`
	LineID        string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_contact_pool_global_unique" json:"line_id"`
	DisplayName   string         `gorm:"type:varchar(100)" json:"display_name"`
	PhoneNumber   string         `gorm:"type:varchar(20)" json:"phone_number"`
	AvatarURL     string         `gorm:"type:varchar(500)" json:"avatar_url"`
	DedupScope    string         `gorm:"type:varchar(20)" json:"dedup_scope"`
	FirstSeenAt   *time.Time     `gorm:"type:timestamp" json:"first_seen_at"`
	Remark        string         `gorm:"type:text" json:"remark"`
	Metadata      JSONB          `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 关联关系
	Group       *Group       `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	LineAccount *LineAccount `gorm:"foreignKey:LineAccountID" json:"line_account,omitempty"`
}

// TableName 指定表名
func (ContactPool) TableName() string {
	return "contact_pool"
}

