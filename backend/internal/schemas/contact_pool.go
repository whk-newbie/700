package schemas

import "time"

// ContactPoolSummaryResponse 底库统计汇总响应
type ContactPoolSummaryResponse struct {
	ImportCount      int64 `json:"import_count"`       // 导入原始联系人数量
	PlatformCount    int64 `json:"platform_count"`     // 平台工单原始联系人数量
	TotalCount       int64 `json:"total_count"`          // 原始联系人数量汇总
}

// ContactPoolListQueryParams 底库列表查询参数（按激活码+平台）
type ContactPoolListQueryParams struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	PlatformType string `form:"platform_type"` // line, line_business
	Search       string `form:"search"`        // 激活码搜索
}

// ContactPoolListResponse 底库列表响应（按激活码+平台）
type ContactPoolListResponse struct {
	ActivationCode string `json:"activation_code"`
	Remark         string `json:"remark"`
	PlatformType   string `json:"platform_type"`
	ContactCount   int64  `json:"contact_count"`
}

// ContactPoolDetailQueryParams 底库详细列表查询参数
type ContactPoolDetailQueryParams struct {
	Page         int       `form:"page" binding:"omitempty,min=1"`
	PageSize     int       `form:"page_size" binding:"omitempty,min=1,max=100"`
	ActivationCode string  `form:"activation_code"`
	PlatformType string    `form:"platform_type"`
	StartTime    *time.Time `form:"start_time" time_format:"2006-01-02 15:04:05"`
	EndTime      *time.Time `form:"end_time" time_format:"2006-01-02 15:04:05"`
	Search       string    `form:"search"` // 用户名或手机号搜索
}

// ContactPoolDetailResponse 底库详细列表响应
type ContactPoolDetailResponse struct {
	ID            uint64     `json:"id"`
	LineID        string     `json:"line_id"`
	DisplayName   string     `json:"display_name"`
	PhoneNumber   string     `json:"phone_number"`
	Source        string     `json:"source"` // 系统上报, 手动导入
	CreatedAt     time.Time  `json:"created_at"`
}

// ImportContactRequest 导入联系人请求
type ImportContactRequest struct {
	PlatformType string `form:"platform_type" binding:"required,oneof=line line_business"`
	DedupScope   string `form:"dedup_scope" binding:"required,oneof=current global"`
	GroupID      uint   `form:"group_id" binding:"required"`
}

// ImportContactResponse 导入联系人响应
type ImportContactResponse struct {
	BatchID      uint   `json:"batch_id"`
	TotalCount   int    `json:"total_count"`
	SuccessCount int    `json:"success_count"`
	DuplicateCount int  `json:"duplicate_count"`
	ErrorCount   int    `json:"error_count"`
}

// ImportBatchListQueryParams 导入批次列表查询参数
type ImportBatchListQueryParams struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	PlatformType string `form:"platform_type"`
}

// ImportBatchListResponse 导入批次列表响应
type ImportBatchListResponse struct {
	ID            uint       `json:"id"`
	BatchName     string     `json:"batch_name"`
	PlatformType  string     `json:"platform_type"`
	TotalCount    int        `json:"total_count"`
	SuccessCount  int        `json:"success_count"`
	DuplicateCount int       `json:"duplicate_count"`
	ErrorCount    int        `json:"error_count"`
	DedupScope    string     `json:"dedup_scope"`
	FileName      string     `json:"file_name"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

