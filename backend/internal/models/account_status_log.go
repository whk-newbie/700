package models

import (
	"time"
)

// AccountStatusLog 账号状态日志模型（分区表）
type AccountStatusLog struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	LineAccountID uint      `gorm:"type:integer;not null;index:idx_account_status_logs_account" json:"line_account_id"`
	FromStatus    string    `gorm:"type:varchar(20);not null;check:from_status IN ('online', 'offline', 'user_logout', 'abnormal_offline')" json:"from_status"`
	ToStatus      string    `gorm:"type:varchar(20);not null;check:to_status IN ('online', 'offline', 'user_logout', 'abnormal_offline')" json:"to_status"`
	Reason        string    `gorm:"type:varchar(50);not null;check:reason IN ('user_login', 'user_logout', 'abnormal_offline', 'force_offline')" json:"reason"`
	IPAddress     string    `gorm:"type:varchar(50)" json:"ip_address"`
	OccurredAt    time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;primaryKey;index:idx_account_status_logs_time" json:"occurred_at"`

	// 关联关系
	LineAccount *LineAccount `gorm:"foreignKey:LineAccountID" json:"line_account,omitempty"`
}

// TableName 指定表名
func (AccountStatusLog) TableName() string {
	return "account_status_logs"
}

