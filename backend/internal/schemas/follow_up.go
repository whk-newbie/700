package schemas

// CreateFollowUpRequest 创建跟进记录请求
type CreateFollowUpRequest struct {
	GroupID        uint   `json:"group_id" binding:"required" example:"1"`
	LineAccountID  *uint  `json:"line_account_id" example:"1"`
	CustomerID     *uint64 `json:"customer_id" example:"1"`
	PlatformType   string `json:"platform_type" binding:"required,oneof=line line_business" example:"line"`
	Content        string `json:"content" binding:"required" example:"跟进内容"`
}

// BatchCreateFollowUpRequest 批量创建跟进记录请求
type BatchCreateFollowUpRequest struct {
	Records []CreateFollowUpRequest `json:"records" binding:"required,min=1,dive"`
}

// UpdateFollowUpRequest 更新跟进记录请求
type UpdateFollowUpRequest struct {
	Content string `json:"content" binding:"required" example:"更新后的跟进内容"`
}

// FollowUpListResponse 跟进记录列表响应
type FollowUpListResponse struct {
	ID                     uint64  `json:"id" example:"1"`
	GroupID                uint    `json:"group_id" example:"1"`
	ActivationCode         string  `json:"activation_code" example:"ABC12345"`
	LineAccountID         *uint   `json:"line_account_id,omitempty" example:"1"`
	CustomerID             *uint64 `json:"customer_id,omitempty" example:"1"`
	PlatformType           string  `json:"platform_type" example:"line"`
	LineAccountDisplayName string  `json:"line_account_display_name" example:"账号名称"`
	LineAccountLineID      string  `json:"line_account_line_id" example:"U1234567890abcdef"`
	LineAccountAvatarURL   string  `json:"line_account_avatar_url" example:"https://profile.line-scdn.net/..."`
	CustomerDisplayName    string  `json:"customer_display_name" example:"客户名称"`
	CustomerLineID         string  `json:"customer_line_id" example:"U1234567890abcdef"`
	CustomerAvatarURL      string  `json:"customer_avatar_url" example:"https://profile.line-scdn.net/..."`
	Content                string  `json:"content" example:"跟进内容"`
	CreatedBy              *uint   `json:"created_by,omitempty" example:"1"`
	CreatedAt              string  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt              string  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	// 关联信息
	GroupRemark            string `json:"group_remark,omitempty" example:"分组备注"`
	CreatedByUsername      string `json:"created_by_username,omitempty" example:"admin"`
}

// FollowUpQueryParams 跟进记录查询参数
type FollowUpQueryParams struct {
	Page          int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize      int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	GroupID       *uint  `form:"group_id" example:"1"`
	LineAccountID *uint  `form:"line_account_id" example:"1"`
	CustomerID    *uint64 `form:"customer_id" example:"1"`
	PlatformType  string `form:"platform_type" example:"line"`
	Search        string `form:"search" example:"搜索内容"` // 搜索跟进内容
	StartTime     string `form:"start_time" example:"2024-01-01T00:00:00Z"`
	EndTime       string `form:"end_time" example:"2024-01-31T23:59:59Z"`
}

// FollowUpSyncData 跟进记录同步数据（用于WebSocket）
type FollowUpSyncData struct {
	LineAccountID string `json:"line_account_id"` // Line账号的line_id
	CustomerID    string `json:"customer_id"`     // 客户的customer_id
	PlatformType  string `json:"platform_type"`
	Content        string `json:"content"`
	Timestamp      string `json:"timestamp,omitempty"`
}

