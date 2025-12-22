package models

import (
	"time"
)

// ImportBatch 导入批次模型
type ImportBatch struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchName     string         `gorm:"type:varchar(100)" json:"batch_name"`
	PlatformType  string         `gorm:"type:varchar(20);not null;check:platform_type IN ('line', 'line_business')" json:"platform_type"`
	TotalCount    int            `gorm:"type:integer;default:0" json:"total_count"`
	SuccessCount  int            `gorm:"type:integer;default:0" json:"success_count"`
	DuplicateCount int           `gorm:"type:integer;default:0" json:"duplicate_count"`
	ErrorCount    int            `gorm:"type:integer;default:0" json:"error_count"`
	DedupScope    string         `gorm:"type:varchar(20);check:dedup_scope IN ('current', 'global')" json:"dedup_scope"`
	FileName      string         `gorm:"type:varchar(255)" json:"file_name"`
	FilePath      string         `gorm:"type:varchar(500)" json:"file_path"`
	FileSize      int64          `gorm:"type:bigint" json:"file_size"`
	ImportedBy    *uint          `gorm:"type:integer" json:"imported_by"`
	CreatedAt     time.Time      `json:"created_at"`
	CompletedAt   *time.Time     `gorm:"type:timestamp" json:"completed_at,omitempty"`

	// 关联关系
	Importer *User `gorm:"foreignKey:ImportedBy" json:"importer,omitempty"`
}

// TableName 指定表名
func (ImportBatch) TableName() string {
	return "import_batches"
}

