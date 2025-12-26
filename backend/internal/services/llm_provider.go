package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProxyToOpenAI 转发请求到OpenAI API
// apiURL: OpenAI API的基础URL（例如：https://api.openai.com/v1）
// apiKey: OpenAI API密钥
// requestBody: 请求体（与OpenAI API文档格式一致）
// timeoutSeconds: 超时时间（秒）
func ProxyToOpenAI(apiURL, apiKey string, requestBody map[string]interface{}, timeoutSeconds int) (map[string]interface{}, error) {
	// 构建完整的API URL
	apiEndpoint := apiURL
	if apiEndpoint == "" {
		apiEndpoint = "https://api.openai.com/v1"
	}
	
	// 规范化URL：确保以/结尾，然后添加chat/completions
	if apiEndpoint[len(apiEndpoint)-1] != '/' {
		apiEndpoint += "/"
	}
	apiEndpoint += "chat/completions"

	// 序列化请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应为JSON
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		// 如果解析失败，返回原始响应
		return map[string]interface{}{
			"error": string(body),
		}, fmt.Errorf("解析响应失败: %v", err)
	}

	// 如果状态码不是200，返回错误信息
	if resp.StatusCode != http.StatusOK {
		errorMsg := "API调用失败"
		if errObj, ok := response["error"].(map[string]interface{}); ok {
			if message, ok := errObj["message"].(string); ok {
				errorMsg = message
			}
		}
		return response, fmt.Errorf("%s (状态码: %d)", errorMsg, resp.StatusCode)
	}

	return response, nil
}

