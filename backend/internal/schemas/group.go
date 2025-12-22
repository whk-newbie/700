package schemas

// CreateGroupRequest 创建分组请求
type CreateGroupRequest struct {
	UserID        uint   `json:"user_id" binding:"required" example:"1"`
	AccountLimit  *int   `json:"account_limit" example:"10"`
	IsActive      bool   `json:"is_active" example:"true"`
	Remark        string `json:"remark" example:"测试分组"`
	Description   string `json:"description" example:"这是一个测试分组"`
	Category      string `json:"category" binding:"omitempty" example:"default"`
	DedupScope    string `json:"dedup_scope" binding:"omitempty,oneof=current global" example:"current"`
	ResetTime     string `json:"reset_time" binding:"omitempty" example:"09:00:00"`
	LoginPassword string `json:"login_password" binding:"omitempty,min=6" example:"password123"`
}

// UpdateGroupRequest 更新分组请求
type UpdateGroupRequest struct {
	AccountLimit  *int   `json:"account_limit" example:"10"`
	IsActive      *bool  `json:"is_active" example:"true"`
	Remark        string `json:"remark" example:"测试分组"`
	Description   string `json:"description" example:"这是一个测试分组"`
	Category      string `json:"category" example:"default"`
	DedupScope    string `json:"dedup_scope" binding:"omitempty,oneof=current global" example:"current"`
	ResetTime     string `json:"reset_time" example:"09:00:00"`
	LoginPassword string `json:"login_password" binding:"omitempty,min=6" example:"password123"`
}

// GroupListResponse 分组列表响应
type GroupListResponse struct {
	ID            uint   `json:"id" example:"1"`
	UserID        uint   `json:"user_id" example:"1"`
	ActivationCode string `json:"activation_code" example:"ABC123"`
	AccountLimit  *int   `json:"account_limit" example:"10"`
	IsActive      bool   `json:"is_active" example:"true"`
	Remark        string `json:"remark" example:"测试分组"`
	Description   string `json:"description" example:"这是一个测试分组"`
	Category      string `json:"category" example:"default"`
	DedupScope    string `json:"dedup_scope" example:"current"`
	ResetTime     string `json:"reset_time" example:"09:00:00"`
	CreatedAt     string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt     string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	LastLoginAt   *string `json:"last_login_at,omitempty" example:"2024-01-01T00:00:00Z"`
	// 统计信息
	TotalAccounts       int `json:"total_accounts" example:"5"`
	OnlineAccounts      int `json:"online_accounts" example:"3"`
	LineAccounts        int `json:"line_accounts" example:"4"`
	LineBusinessAccounts int `json:"line_business_accounts" example:"1"`
	TodayIncoming       int `json:"today_incoming" example:"10"`
	TotalIncoming       int `json:"total_incoming" example:"100"`
	DuplicateIncoming   int `json:"duplicate_incoming" example:"5"`
	TodayDuplicate      int `json:"today_duplicate" example:"2"`
	// 用户信息
	Username string `json:"username,omitempty" example:"user001"`
}

// GroupQueryParams 分组查询参数
type GroupQueryParams struct {
	Page       int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	UserID     *uint  `form:"user_id" example:"1"`
	Category   string `form:"category" example:"default"`
	IsActive   *bool  `form:"is_active" example:"true"`
	Search     string `form:"search" example:"ABC123"` // 搜索激活码或备注
}

// BatchDeleteGroupsRequest 批量删除分组请求
type BatchDeleteGroupsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,dive,min=1"`
}

// BatchUpdateGroupsRequest 批量更新分组请求
type BatchUpdateGroupsRequest struct {
	IDs        []uint `json:"ids" binding:"required,min=1,dive,min=1"`
	IsActive   *bool  `json:"is_active" example:"true"`
	Category   string `json:"category" example:"default"`
	DedupScope string `json:"dedup_scope" binding:"omitempty,oneof=current global" example:"current"`
}

// BatchOperationResponse 批量操作响应
type BatchOperationResponse struct {
	SuccessCount int    `json:"success_count" example:"3"`
	FailCount    int    `json:"fail_count" example:"0"`
	FailedIDs    []uint `json:"failed_ids,omitempty"`
}

