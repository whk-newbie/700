package schemas

// CreateCustomerRequest 创建客户请求
type CreateCustomerRequest struct {
	GroupID        uint   `json:"group_id" binding:"required" example:"1"`
	LineAccountID  *uint  `json:"line_account_id" example:"1"`
	PlatformType   string `json:"platform_type" binding:"required,oneof=line line_business" example:"line"`
	CustomerID     string `json:"customer_id" binding:"required" example:"U1234567890abcdef"`
	DisplayName    string `json:"display_name" example:"客户名称"`
	AvatarURL      string `json:"avatar_url" example:"https://profile.line-scdn.net/..."`
	PhoneNumber    string `json:"phone_number" example:"13800138000"`
	CustomerType   string `json:"customer_type" example:"friend"`
	Gender         string `json:"gender" binding:"omitempty,oneof=male female unknown" example:"male"`
	Country        string `json:"country" example:"TW"`
	Birthday       string `json:"birthday" example:"1990-01-01"`
	Address        string `json:"address" example:"地址信息"`
	NicknameRemark string `json:"nickname_remark" example:"昵称备注"`
	Remark         string `json:"remark" example:"备注信息"`
}

// UpdateCustomerRequest 更新客户请求
type UpdateCustomerRequest struct {
	LineAccountID  *uint  `json:"line_account_id" example:"1"`
	DisplayName    string `json:"display_name" example:"客户名称"`
	AvatarURL      string `json:"avatar_url" example:"https://profile.line-scdn.net/..."`
	PhoneNumber    string `json:"phone_number" example:"13800138000"`
	CustomerType   string `json:"customer_type" example:"friend"`
	Gender         string `json:"gender" binding:"omitempty,oneof=male female unknown" example:"male"`
	Country        string `json:"country" example:"TW"`
	Birthday       string `json:"birthday" example:"1990-01-01"`
	Address        string `json:"address" example:"地址信息"`
	NicknameRemark string `json:"nickname_remark" example:"昵称备注"`
	Remark         string `json:"remark" example:"备注信息"`
}

// CustomerListResponse 客户列表响应
type CustomerListResponse struct {
	ID             uint64  `json:"id" example:"1"`
	GroupID        uint    `json:"group_id" example:"1"`
	ActivationCode string  `json:"activation_code" example:"ABC12345"`
	LineAccountID  *uint   `json:"line_account_id,omitempty" example:"1"`
	PlatformType   string  `json:"platform_type" example:"line"`
	CustomerID     string  `json:"customer_id" example:"U1234567890abcdef"`
	DisplayName    string  `json:"display_name" example:"客户名称"`
	AvatarURL      string  `json:"avatar_url" example:"https://profile.line-scdn.net/..."`
	PhoneNumber    string  `json:"phone_number" example:"13800138000"`
	CustomerType   string  `json:"customer_type" example:"friend"`
	Gender         string  `json:"gender" example:"male"`
	Country        string  `json:"country" example:"TW"`
	Birthday       *string `json:"birthday,omitempty" example:"1990-01-01"`
	Address        string  `json:"address" example:"地址信息"`
	NicknameRemark string  `json:"nickname_remark" example:"昵称备注"`
	Remark         string  `json:"remark" example:"备注信息"`
	CreatedAt      string  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      string  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	// 关联信息
	LineAccountDisplayName string `json:"line_account_display_name,omitempty" example:"账号名称"`
	LineAccountLineID      string `json:"line_account_line_id,omitempty" example:"U1234567890abcdef"`
	GroupRemark            string `json:"group_remark,omitempty" example:"分组备注"`
}

// CustomerDetailResponse 客户详情响应
type CustomerDetailResponse struct {
	CustomerListResponse
	Tags        map[string]interface{} `json:"tags,omitempty"`
	ProfileData map[string]interface{} `json:"profile_data,omitempty"`
}

// CustomerQueryParams 客户查询参数
type CustomerQueryParams struct {
	Page          int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize      int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	GroupID       *uint  `form:"group_id" example:"1"`
	LineAccountID *uint  `form:"line_account_id" example:"1"`
	PlatformType  string `form:"platform_type" example:"line"`
	CustomerType  string `form:"customer_type" example:"friend"`
	Search        string `form:"search" example:"客户名称"` // 搜索customer_id或display_name
}

// CustomerSyncData 客户同步数据（用于WebSocket）
type CustomerSyncData struct {
	LineAccountID string `json:"line_account_id"` // Line账号的line_id
	CustomerID    string `json:"customer_id"`
	PlatformType  string `json:"platform_type"`
	DisplayName   string `json:"display_name,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	Gender        string `json:"gender,omitempty"`
	Country       string `json:"country,omitempty"`
	Birthday      string `json:"birthday,omitempty"`
	Address       string `json:"address,omitempty"`
	Remark        string `json:"remark,omitempty"`
}

