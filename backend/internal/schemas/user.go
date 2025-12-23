package schemas

// UserQueryParams 用户查询参数
type UserQueryParams struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=admin user"`
	IsActive *bool  `form:"is_active"`
	Search   string `form:"search"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=6,max=100"`
	Email     string `json:"email" binding:"omitempty,email"`
	Role      string `json:"role" binding:"required,oneof=admin user"`
	MaxGroups *int   `json:"max_groups" binding:"omitempty,min=0"`
	IsActive  bool   `json:"is_active"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email     string `json:"email" binding:"omitempty,email"`
	Role      string `json:"role" binding:"omitempty,oneof=admin user"`
	MaxGroups *int   `json:"max_groups" binding:"omitempty,min=0"`
	IsActive  *bool  `json:"is_active"`
	Password  string `json:"password" binding:"omitempty,min=6,max=100"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	MaxGroups *int   `json:"max_groups"`
	IsActive  bool   `json:"is_active"`
	CreatedBy *uint  `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

