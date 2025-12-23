package schemas

// LLMConfigQueryParams LLM配置查询参数
type LLMConfigQueryParams struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Provider string `form:"provider" binding:"omitempty"`
	IsActive *bool  `form:"is_active"`
	Search   string `form:"search"`
}

// CreateLLMConfigRequest 创建LLM配置请求
type CreateLLMConfigRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=100"`
	Provider        string  `json:"provider" binding:"required,oneof=openai anthropic aliyun xunfei baidu zhipu custom"`
	APIURL          string  `json:"api_url" binding:"required,url"`
	APIKey          string  `json:"api_key" binding:"required"`
	Model           string  `json:"model" binding:"required,min=1,max=100"`
	MaxTokens       int     `json:"max_tokens" binding:"omitempty,min=1,max=100000"`
	Temperature     float64 `json:"temperature" binding:"omitempty,min=0,max=2"`
	TopP            float64 `json:"top_p" binding:"omitempty,min=0,max=1"`
	FrequencyPenalty float64 `json:"frequency_penalty" binding:"omitempty,min=-2,max=2"`
	PresencePenalty  float64 `json:"presence_penalty" binding:"omitempty,min=-2,max=2"`
	SystemPrompt    string  `json:"system_prompt"`
	TimeoutSeconds  int     `json:"timeout_seconds" binding:"omitempty,min=1,max=300"`
	MaxRetries      int     `json:"max_retries" binding:"omitempty,min=0,max=10"`
	IsActive        bool    `json:"is_active"`
}

// UpdateLLMConfigRequest 更新LLM配置请求
type UpdateLLMConfigRequest struct {
	Name            string  `json:"name" binding:"omitempty,min=1,max=100"`
	Provider        string  `json:"provider" binding:"omitempty,oneof=openai anthropic aliyun xunfei baidu zhipu custom"`
	APIURL          string  `json:"api_url" binding:"omitempty,url"`
	APIKey          string  `json:"api_key"`
	Model           string  `json:"model" binding:"omitempty,min=1,max=100"`
	MaxTokens       *int    `json:"max_tokens" binding:"omitempty,min=1,max=100000"`
	Temperature     *float64 `json:"temperature" binding:"omitempty,min=0,max=2"`
	TopP            *float64 `json:"top_p" binding:"omitempty,min=0,max=1"`
	FrequencyPenalty *float64 `json:"frequency_penalty" binding:"omitempty,min=-2,max=2"`
	PresencePenalty  *float64 `json:"presence_penalty" binding:"omitempty,min=-2,max=2"`
	SystemPrompt    string  `json:"system_prompt"`
	TimeoutSeconds  *int    `json:"timeout_seconds" binding:"omitempty,min=1,max=300"`
	MaxRetries      *int    `json:"max_retries" binding:"omitempty,min=0,max=10"`
	IsActive        *bool   `json:"is_active"`
}

// LLMConfigResponse LLM配置响应
type LLMConfigResponse struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Provider        string  `json:"provider"`
	APIURL          string  `json:"api_url"`
	Model           string  `json:"model"`
	MaxTokens       int     `json:"max_tokens"`
	Temperature     float64 `json:"temperature"`
	TopP            float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	SystemPrompt    string  `json:"system_prompt"`
	TimeoutSeconds  int     `json:"timeout_seconds"`
	MaxRetries      int     `json:"max_retries"`
	IsActive        bool    `json:"is_active"`
	CreatedBy       *uint   `json:"created_by"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// LLMConfigPublicResponse LLM配置公开响应（不包含API Key）
type LLMConfigPublicResponse struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Provider        string  `json:"provider"`
	APIURL          string  `json:"api_url"`
	Model           string  `json:"model"`
	MaxTokens       int     `json:"max_tokens"`
	Temperature     float64 `json:"temperature"`
	TopP            float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	SystemPrompt    string  `json:"system_prompt"`
	TimeoutSeconds  int     `json:"timeout_seconds"`
	MaxRetries      int     `json:"max_retries"`
	IsActive        bool    `json:"is_active"`
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
	ID               uint                   `json:"id"`
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

