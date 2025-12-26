package schemas

// UpdateOpenAIAPIKeyRequest 更新OpenAI API Key请求（API Key已使用RSA加密）
type UpdateOpenAIAPIKeyRequest struct {
	EncryptedAPIKey string `json:"encrypted_api_key" binding:"required"` // RSA加密后的API Key（Base64编码）
}

// RSAPublicKeyResponse RSA公钥响应
type RSAPublicKeyResponse struct {
	PublicKey string `json:"public_key"` // PEM格式的公钥
}

// OpenAIAPIKeyResponse OpenAI API Key响应
type OpenAIAPIKeyResponse struct {
	HasKey   bool   `json:"has_key"`   // 是否已配置API Key
	UpdatedAt string `json:"updated_at,omitempty"` // 更新时间
}

// PromptTemplateQueryParams Prompt模板查询参数
type PromptTemplateQueryParams struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	ConfigID *uint  `form:"config_id"`
	IsActive *bool  `form:"is_active"`
	Search   string `form:"search"`
}

// CreatePromptTemplateRequest 创建Prompt模板请求
type CreatePromptTemplateRequest struct {
	ConfigID       uint   `json:"config_id" binding:"required"`
	TemplateName   string `json:"template_name" binding:"required,min=1,max=100"`
	TemplateContent string `json:"template_content" binding:"required"`
	Variables      map[string]interface{} `json:"variables"`
	Description    string `json:"description"`
	IsActive       bool   `json:"is_active"`
}

// UpdatePromptTemplateRequest 更新Prompt模板请求
type UpdatePromptTemplateRequest struct {
	TemplateName   string `json:"template_name" binding:"omitempty,min=1,max=100"`
	TemplateContent string `json:"template_content"`
	Variables      map[string]interface{} `json:"variables"`
	Description    string `json:"description"`
	IsActive       *bool  `json:"is_active"`
}

// PromptTemplateResponse Prompt模板响应
type PromptTemplateResponse struct {
	ID             uint                   `json:"id"`
	ConfigID       uint                   `json:"config_id"`
	TemplateName   string                 `json:"template_name"`
	TemplateContent string                `json:"template_content"`
	Variables      map[string]interface{} `json:"variables"`
	Description    string                 `json:"description"`
	IsActive       bool                   `json:"is_active"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
}

// LLMCallRequest 大模型调用请求
type LLMCallRequest struct {
	ConfigID      *uint                  `json:"config_id" binding:"required"`
	Messages      []map[string]interface{} `json:"messages" binding:"required,min=1"`
	Temperature   *float64                `json:"temperature"`
	MaxTokens     *int                    `json:"max_tokens"`
	GroupID       *uint                   `json:"group_id"`
	ActivationCode string                 `json:"activation_code"`
}

// LLMCallTemplateRequest 使用模板调用请求
type LLMCallTemplateRequest struct {
	ConfigID      uint                   `json:"config_id" binding:"required"`
	TemplateID    uint                   `json:"template_id" binding:"required"`
	Variables     map[string]interface{} `json:"variables"`
	Temperature   *float64               `json:"temperature"`
	MaxTokens     *int                    `json:"max_tokens"`
	GroupID       *uint                   `json:"group_id"`
	ActivationCode string                `json:"activation_code"`
}

// LLMCallResponse 大模型调用响应
type LLMCallResponse struct {
	Content       string                 `json:"content"`
	ResponseData  map[string]interface{} `json:"response_data,omitempty"`
	TokensUsed    *int                   `json:"tokens_used,omitempty"`
	PromptTokens  *int                   `json:"prompt_tokens,omitempty"`
	CompletionTokens *int                `json:"completion_tokens,omitempty"`
}

// LLMCallLogQueryParams 调用日志查询参数
type LLMCallLogQueryParams struct {
	Page          int    `form:"page" binding:"omitempty,min=1"`
	PageSize      int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	ConfigID      *uint  `form:"config_id"`
	TemplateID    *uint  `form:"template_id"`
	GroupID       *uint  `form:"group_id"`
	ActivationCode string `form:"activation_code"`
	Status        string `form:"status" binding:"omitempty,oneof=success error"`
	StartTime     string `form:"start_time"`
	EndTime       string `form:"end_time"`
}

// LLMCallLogResponse 调用日志响应
type LLMCallLogResponse struct {
	ID               uint64                 `json:"id"`
	ConfigID         *uint                  `json:"config_id"`
	TemplateID       *uint                  `json:"template_id"`
	GroupID          *uint                  `json:"group_id"`
	ActivationCode   string                 `json:"activation_code"`
	RequestMessages  []map[string]interface{} `json:"request_messages"`
	RequestParams    map[string]interface{} `json:"request_params"`
	ResponseContent  string                 `json:"response_content"`
	ResponseData     map[string]interface{} `json:"response_data"`
	Status           string                 `json:"status"`
	ErrorMessage     string                 `json:"error_message"`
	TokensUsed       *int                   `json:"tokens_used"`
	PromptTokens     *int                   `json:"prompt_tokens"`
	CompletionTokens *int                   `json:"completion_tokens"`
	CallTime         string                 `json:"call_time"`
	DurationMs       *int                   `json:"duration_ms"`
}

// TestLLMConfigRequest 测试LLM配置请求
type TestLLMConfigRequest struct {
	ConfigID    uint   `json:"config_id" binding:"required"`
	TestMessage string `json:"test_message" binding:"required"`
}

// OpenAIProxyRequest OpenAI API转发请求
// 前端传参格式与OpenAI文档一致，所有字段都会原样转发给OpenAI API
// 后端会自动从配置中获取API Key并添加到请求头
type OpenAIProxyRequest struct {
	// 以下字段与OpenAI API文档一致，支持所有OpenAI API参数
	Model       string                   `json:"model" binding:"required"`
	Messages    []map[string]interface{} `json:"messages" binding:"required,min=1"`
	Temperature *float64                 `json:"temperature,omitempty"`
	TopP        *float64                 `json:"top_p,omitempty"`
	MaxTokens   *int                     `json:"max_tokens,omitempty"`
	Stream      *bool                    `json:"stream,omitempty"`
	N           *int                     `json:"n,omitempty"`
	Stop        interface{}              `json:"stop,omitempty"` // string or []string
	PresencePenalty *float64             `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64            `json:"frequency_penalty,omitempty"`
	LogitBias   map[string]interface{}   `json:"logit_bias,omitempty"`
	User        string                   `json:"user,omitempty"`
}

