package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"line-management/internal/models"
	"line-management/pkg/logger"
)

// LLMProvider 大模型提供商接口
type LLMProvider interface {
	Call(config *models.LLMConfig, messages []map[string]interface{}, params map[string]interface{}) (*LLMResponse, error)
	TestConnection(config *models.LLMConfig) error
}

// LLMResponse 大模型响应
type LLMResponse struct {
	Content          string
	TokensUsed       *int
	PromptTokens     *int
	CompletionTokens *int
	ResponseData     map[string]interface{}
}

// OpenAIProvider OpenAI提供商
type OpenAIProvider struct{}

// Call 调用OpenAI API
func (p *OpenAIProvider) Call(config *models.LLMConfig, messages []map[string]interface{}, params map[string]interface{}) (*LLMResponse, error) {
	// 解密API Key
	encryptionService := GetEncryptionService()
	apiKey, err := encryptionService.Decrypt(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("解密API Key失败: %v", err)
	}

	// 构建请求
	requestBody := map[string]interface{}{
		"model":       config.Model,
		"messages":    messages,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
		"top_p":       config.TopP,
	}

	// 合并自定义参数
	if params != nil {
		if temp, ok := params["temperature"].(float64); ok {
			requestBody["temperature"] = temp
		}
		if maxTokens, ok := params["max_tokens"].(int); ok {
			requestBody["max_tokens"] = maxTokens
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	req, err := http.NewRequest("POST", config.APIURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API调用失败: %s", string(body))
	}

	// 解析响应
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens      int `json:"total_tokens"`
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("API返回空响应")
	}

	result := &LLMResponse{
		Content:          response.Choices[0].Message.Content,
		TokensUsed:       &response.Usage.TotalTokens,
		PromptTokens:     &response.Usage.PromptTokens,
		CompletionTokens: &response.Usage.CompletionTokens,
		ResponseData:     make(map[string]interface{}),
	}

	// 保存完整响应数据
	if err := json.Unmarshal(body, &result.ResponseData); err == nil {
		// 解析成功
	}

	return result, nil
}

// TestConnection 测试连接
func (p *OpenAIProvider) TestConnection(config *models.LLMConfig) error {
	testMessages := []map[string]interface{}{
		{
			"role":    "user",
			"content": "Hello",
		},
	}
	_, err := p.Call(config, testMessages, nil)
	return err
}

// AnthropicProvider Claude提供商
type AnthropicProvider struct{}

// Call 调用Claude API
func (p *AnthropicProvider) Call(config *models.LLMConfig, messages []map[string]interface{}, params map[string]interface{}) (*LLMResponse, error) {
	// 解密API Key
	encryptionService := GetEncryptionService()
	apiKey, err := encryptionService.Decrypt(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("解密API Key失败: %v", err)
	}

	// 转换消息格式（Claude使用不同的格式）
	claudeMessages := make([]map[string]interface{}, 0)
	for _, msg := range messages {
		claudeMsg := map[string]interface{}{
			"role":    msg["role"],
			"content": msg["content"],
		}
		claudeMessages = append(claudeMessages, claudeMsg)
	}

	// 构建请求
	requestBody := map[string]interface{}{
		"model":       config.Model,
		"max_tokens":  config.MaxTokens,
		"messages":    claudeMessages,
		"temperature": config.Temperature,
		"top_p":       config.TopP,
	}

	// 合并自定义参数
	if params != nil {
		if temp, ok := params["temperature"].(float64); ok {
			requestBody["temperature"] = temp
		}
		if maxTokens, ok := params["max_tokens"].(int); ok {
			requestBody["max_tokens"] = maxTokens
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	req, err := http.NewRequest("POST", config.APIURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API调用失败: %s", string(body))
	}

	// 解析响应
	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("API返回空响应")
	}

	totalTokens := response.Usage.InputTokens + response.Usage.OutputTokens
	result := &LLMResponse{
		Content:          response.Content[0].Text,
		TokensUsed:       &totalTokens,
		PromptTokens:     &response.Usage.InputTokens,
		CompletionTokens: &response.Usage.OutputTokens,
		ResponseData:     make(map[string]interface{}),
	}

	// 保存完整响应数据
	if err := json.Unmarshal(body, &result.ResponseData); err == nil {
		// 解析成功
	}

	return result, nil
}

// TestConnection 测试连接
func (p *AnthropicProvider) TestConnection(config *models.LLMConfig) error {
	testMessages := []map[string]interface{}{
		{
			"role":    "user",
			"content": "Hello",
		},
	}
	_, err := p.Call(config, testMessages, nil)
	return err
}

// AliyunProvider 通义千问提供商
type AliyunProvider struct{}

// Call 调用通义千问API
func (p *AliyunProvider) Call(config *models.LLMConfig, messages []map[string]interface{}, params map[string]interface{}) (*LLMResponse, error) {
	// 解密API Key
	encryptionService := GetEncryptionService()
	apiKey, err := encryptionService.Decrypt(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("解密API Key失败: %v", err)
	}

	// 转换消息格式
	dashscopeMessages := make([]map[string]interface{}, 0)
	for _, msg := range messages {
		dashscopeMsg := map[string]interface{}{
			"role":    msg["role"],
			"content": msg["content"],
		}
		dashscopeMessages = append(dashscopeMessages, dashscopeMsg)
	}

	// 构建请求
	requestBody := map[string]interface{}{
		"model":       config.Model,
		"input": map[string]interface{}{
			"messages": dashscopeMessages,
		},
		"parameters": map[string]interface{}{
			"max_tokens":  config.MaxTokens,
			"temperature": config.Temperature,
			"top_p":       config.TopP,
		},
	}

	// 合并自定义参数
	if params != nil {
		paramsMap := requestBody["parameters"].(map[string]interface{})
		if temp, ok := params["temperature"].(float64); ok {
			paramsMap["temperature"] = temp
		}
		if maxTokens, ok := params["max_tokens"].(int); ok {
			paramsMap["max_tokens"] = maxTokens
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	req, err := http.NewRequest("POST", config.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API调用失败: %s", string(body))
	}

	// 解析响应
	var response struct {
		Output struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"output"`
		Usage struct {
			TotalTokens      int `json:"total_tokens"`
			InputTokens      int `json:"input_tokens"`
			OutputTokens     int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if len(response.Output.Choices) == 0 {
		return nil, fmt.Errorf("API返回空响应")
	}

	result := &LLMResponse{
		Content:          response.Output.Choices[0].Message.Content,
		TokensUsed:       &response.Usage.TotalTokens,
		PromptTokens:     &response.Usage.InputTokens,
		CompletionTokens: &response.Usage.OutputTokens,
		ResponseData:     make(map[string]interface{}),
	}

	// 保存完整响应数据
	if err := json.Unmarshal(body, &result.ResponseData); err == nil {
		// 解析成功
	}

	return result, nil
}

// TestConnection 测试连接
func (p *AliyunProvider) TestConnection(config *models.LLMConfig) error {
	testMessages := []map[string]interface{}{
		{
			"role":    "user",
			"content": "Hello",
		},
	}
	_, err := p.Call(config, testMessages, nil)
	return err
}

// CustomProvider 自定义提供商（通用HTTP接口）
type CustomProvider struct{}

// Call 调用自定义API
func (p *CustomProvider) Call(config *models.LLMConfig, messages []map[string]interface{}, params map[string]interface{}) (*LLMResponse, error) {
	// 解密API Key
	encryptionService := GetEncryptionService()
	apiKey, err := encryptionService.Decrypt(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("解密API Key失败: %v", err)
	}

	// 构建请求（使用通用格式）
	requestBody := map[string]interface{}{
		"model":       config.Model,
		"messages":    messages,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
	}

	// 合并自定义参数
	if params != nil {
		if temp, ok := params["temperature"].(float64); ok {
			requestBody["temperature"] = temp
		}
		if maxTokens, ok := params["max_tokens"].(int); ok {
			requestBody["max_tokens"] = maxTokens
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	req, err := http.NewRequest("POST", config.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API调用失败: %s", string(body))
	}

	// 尝试解析通用响应格式
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	// 尝试提取内容（根据常见格式）
	content := ""
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if c, ok := message["content"].(string); ok {
					content = c
				}
			}
		}
	}

	if content == "" {
		// 如果无法解析，返回原始响应
		content = string(body)
	}

	result := &LLMResponse{
		Content:      content,
		ResponseData: response,
	}

	return result, nil
}

// TestConnection 测试连接
func (p *CustomProvider) TestConnection(config *models.LLMConfig) error {
	testMessages := []map[string]interface{}{
		{
			"role":    "user",
			"content": "Hello",
		},
	}
	_, err := p.Call(config, testMessages, nil)
	return err
}

// GetLLMProvider 获取LLM提供商实例
func GetLLMProvider(provider string) LLMProvider {
	switch provider {
	case "openai":
		return &OpenAIProvider{}
	case "anthropic":
		return &AnthropicProvider{}
	case "aliyun":
		return &AliyunProvider{}
	case "custom":
		return &CustomProvider{}
	default:
		logger.Warnf("未知的提供商: %s，使用自定义提供商", provider)
		return &CustomProvider{}
	}
}

