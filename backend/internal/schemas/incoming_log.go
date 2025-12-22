package schemas

// IncomingLogQueryParams 进线日志查询参数
type IncomingLogQueryParams struct {
	Page          int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize      int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	GroupID       *uint  `form:"group_id" example:"1"`
	LineAccountID *uint  `form:"line_account_id" example:"1"`
	IsDuplicate   *bool  `form:"is_duplicate" example:"false"`
	StartTime     string `form:"start_time" example:"2024-01-01T00:00:00Z"` // ISO 8601格式
	EndTime       string `form:"end_time" example:"2024-01-31T23:59:59Z"`   // ISO 8601格式
	Search        string `form:"search" example:"U1234567890"`              // 搜索进线Line ID或显示名称
}

// IncomingLogListResponse 进线日志列表响应
type IncomingLogListResponse struct {
	ID             uint64 `json:"id" example:"1"`
	LineAccountID  uint   `json:"line_account_id" example:"1"`
	GroupID        uint   `json:"group_id" example:"1"`
	IncomingLineID string `json:"incoming_line_id" example:"U1234567890"`
	IncomingTime   string `json:"incoming_time" example:"2024-01-01T00:00:00Z"`
	DisplayName    string `json:"display_name" example:"张三"`
	AvatarURL      string `json:"avatar_url" example:"https://example.com/avatar.jpg"`
	PhoneNumber    string `json:"phone_number" example:"13800138000"`
	IsDuplicate    bool   `json:"is_duplicate" example:"false"`
	DuplicateScope string `json:"duplicate_scope" example:"current"`
	CustomerType   string `json:"customer_type" example:"lead"`
	// 关联信息
	LineAccount *LineAccountInfo `json:"line_account,omitempty"`
	Group       *GroupInfo        `json:"group,omitempty"`
}

// LineAccountInfo Line账号信息
type LineAccountInfo struct {
	ID          uint   `json:"id" example:"1"`
	LineID      string `json:"line_id" example:"U9876543210"`
	DisplayName string `json:"display_name" example:"客服账号"`
	PlatformType string `json:"platform_type" example:"line"`
}

// GroupInfo 分组信息
type GroupInfo struct {
	ID            uint   `json:"id" example:"1"`
	ActivationCode string `json:"activation_code" example:"ABC123"`
	Remark        string `json:"remark" example:"测试分组"`
}

