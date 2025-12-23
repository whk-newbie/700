package schemas

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

// SubAccountLoginRequest 子账号登录请求
type SubAccountLoginRequest struct {
	ActivationCode string `json:"activation_code" binding:"required" example:"ABC123"`
	Password       string `json:"password" binding:"omitempty" example:"password123"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType string      `json:"token_type" example:"Bearer"`
	ExpiresIn int         `json:"expires_in" example:"86400"`
	User      *UserInfo   `json:"user,omitempty"`
	Group     *GroupInfo  `json:"group,omitempty"` // 子账号登录时返回
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint   `json:"id" example:"1"`
	Username string `json:"username" example:"admin"`
	Email    string `json:"email" example:"admin@example.com"`
	Role     string `json:"role" example:"admin"`
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType string `json:"token_type" example:"Bearer"`
	ExpiresIn int    `json:"expires_in" example:"86400"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"用户名或密码错误"`
	Error   string `json:"error,omitempty" example:"invalid_credentials"`
}

