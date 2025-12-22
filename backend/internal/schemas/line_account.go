package schemas

// CreateLineAccountRequest 创建Line账号请求
type CreateLineAccountRequest struct {
	GroupID       uint   `json:"group_id" binding:"required" example:"1"`
	PlatformType  string `json:"platform_type" binding:"required,oneof=line line_business" example:"line"`
	LineID        string `json:"line_id" binding:"required" example:"U1234567890abcdef"`
	DisplayName   string `json:"display_name" example:"测试账号"`
	PhoneNumber   string `json:"phone_number" example:"13800138000"`
	ProfileURL    string `json:"profile_url" example:"https://profile.line-scdn.net/..."`
	AvatarURL     string `json:"avatar_url" example:"https://profile.line-scdn.net/..."`
	Bio           string `json:"bio" example:"这是个人简介"`
	StatusMessage string `json:"status_message" example:"这是状态消息"`
	AccountRemark string `json:"account_remark" example:"这是备注"`
}

// UpdateLineAccountRequest 更新Line账号请求
type UpdateLineAccountRequest struct {
	DisplayName   string `json:"display_name" example:"测试账号"`
	PhoneNumber   string `json:"phone_number" example:"13800138000"`
	ProfileURL    string `json:"profile_url" example:"https://profile.line-scdn.net/..."`
	AvatarURL     string `json:"avatar_url" example:"https://profile.line-scdn.net/..."`
	Bio           string `json:"bio" example:"这是个人简介"`
	StatusMessage string `json:"status_message" example:"这是状态消息"`
	AccountRemark string `json:"account_remark" example:"这是备注"`
	OnlineStatus  string `json:"online_status" binding:"omitempty,oneof=online offline user_logout abnormal_offline" example:"online"`
}

// LineAccountListResponse Line账号列表响应
type LineAccountListResponse struct {
	ID             uint    `json:"id" example:"1"`
	GroupID        uint    `json:"group_id" example:"1"`
	ActivationCode string  `json:"activation_code" example:"ABC12345"`
	PlatformType   string  `json:"platform_type" example:"line"`
	LineID         string  `json:"line_id" example:"U1234567890abcdef"`
	DisplayName    string  `json:"display_name" example:"测试账号"`
	PhoneNumber    string  `json:"phone_number" example:"13800138000"`
	ProfileURL     string  `json:"profile_url" example:"https://profile.line-scdn.net/..."`
	AvatarURL      string  `json:"avatar_url" example:"https://profile.line-scdn.net/..."`
	Bio            string  `json:"bio" example:"这是个人简介"`
	StatusMessage  string  `json:"status_message" example:"这是状态消息"`
	QRCodePath     string  `json:"qr_code_path" example:"/static/qrcodes/1.png"`
	OnlineStatus   string  `json:"online_status" example:"online"`
	LastActiveAt   *string `json:"last_active_at,omitempty" example:"2024-01-01T00:00:00Z"`
	LastOnlineTime *string `json:"last_online_time,omitempty" example:"2024-01-01T00:00:00Z"`
	FirstLoginAt   *string `json:"first_login_at,omitempty" example:"2024-01-01T00:00:00Z"`
	AccountRemark  string  `json:"account_remark" example:"这是备注"`
	CreatedAt      string  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      string  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	// 统计信息
	TodayIncoming     int `json:"today_incoming" example:"10"`
	TotalIncoming     int `json:"total_incoming" example:"100"`
	DuplicateIncoming int `json:"duplicate_incoming" example:"5"`
	TodayDuplicate    int `json:"today_duplicate" example:"2"`
	// 分组信息
	GroupRemark string `json:"group_remark,omitempty" example:"测试分组"`
}

// LineAccountQueryParams Line账号查询参数
type LineAccountQueryParams struct {
	Page          int     `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize      int     `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	GroupID       *uint   `form:"group_id" example:"1"`
	PlatformType  string  `form:"platform_type" example:"line"`
	OnlineStatus  string  `form:"online_status" example:"online"`
	ActivationCode string  `form:"activation_code" example:"ABC12345"`
	Search        string  `form:"search" example:"测试账号"` // 搜索line_id或display_name
}

// GenerateQRCodeResponse 生成二维码响应
type GenerateQRCodeResponse struct {
	QRCodePath string `json:"qr_code_path" example:"/static/qrcodes/1.png"`
	QRCodeURL  string `json:"qr_code_url" example:"http://localhost:8080/static/qrcodes/1.png"`
}

// BatchDeleteLineAccountsRequest 批量删除Line账号请求
type BatchDeleteLineAccountsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,dive,min=1"`
}

// BatchUpdateLineAccountsRequest 批量更新Line账号请求
type BatchUpdateLineAccountsRequest struct {
	IDs          []uint `json:"ids" binding:"required,min=1,dive,min=1"`
	OnlineStatus string `json:"online_status" binding:"omitempty,oneof=online offline user_logout abnormal_offline" example:"offline"`
}

