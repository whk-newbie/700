package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LLMService LLM调用服务
type LLMService struct {
	db *gorm.DB
}

// NewLLMService 创建LLM调用服务实例
func NewLLMService() *LLMService {
	return &LLMService{
		db: database.GetDB(),
	}
}

// CallLLM 调用大模型
func (s *LLMService) CallLLM(c *gin.Context, req *schemas.LLMCallRequest) (*schemas.LLMCallResponse, error) {
	// 获取配置
	configService := NewLLMConfigService()
	config, err := configService.GetLLMConfigByID(*req.ConfigID)
	if err != nil {
		return nil, err
	}

	if !config.IsActive {
		return nil, errors.New("配置未激活")
	}

	// 构建消息列表
	messages := make([]map[string]interface{}, 0)
	
	// 如果有系统提示词，添加到消息列表开头
	if config.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": config.SystemPrompt,
		})
	}

	// 添加用户消息
	messages = append(messages, req.Messages...)

	// 构建参数
	params := make(map[string]interface{})
	if req.Temperature != nil {
		params["temperature"] = *req.Temperature
	} else {
		params["temperature"] = config.Temperature
	}
	if req.MaxTokens != nil {
		params["max_tokens"] = *req.MaxTokens
	} else {
		params["max_tokens"] = config.MaxTokens
	}

	// 获取提供商
	provider := GetLLMProvider(config.Provider)
	if provider == nil {
		return nil, errors.New("不支持的提供商")
	}

	// 记录开始时间
	startTime := time.Now()

	// 调用API（带重试）
	var response *LLMResponse
	var lastErr error
	for i := 0; i <= config.MaxRetries; i++ {
		response, lastErr = provider.Call(config, messages, params)
		if lastErr == nil {
			break
		}
		if i < config.MaxRetries {
			logger.Warnf("LLM调用失败，重试 %d/%d: %v", i+1, config.MaxRetries, lastErr)
			time.Sleep(time.Duration(i+1) * time.Second) // 递增延迟
		}
	}

	if lastErr != nil {
		// 记录失败日志
		s.recordCallLog(c, config, nil, req, nil, "error", lastErr.Error(), nil, nil, nil, time.Since(startTime))
		return nil, fmt.Errorf("LLM调用失败: %v", lastErr)
	}

	// 计算耗时
	duration := time.Since(startTime)

	// 记录成功日志
	s.recordCallLog(c, config, nil, req, response, "success", "", response.TokensUsed, response.PromptTokens, response.CompletionTokens, duration)

	// 构建响应
	result := &schemas.LLMCallResponse{
		Content:          response.Content,
		ResponseData:     response.ResponseData,
		TokensUsed:       response.TokensUsed,
		PromptTokens:     response.PromptTokens,
		CompletionTokens: response.CompletionTokens,
	}

	return result, nil
}

// CallLLMWithTemplate 使用模板调用大模型
func (s *LLMService) CallLLMWithTemplate(c *gin.Context, req *schemas.LLMCallTemplateRequest) (*schemas.LLMCallResponse, error) {
	// 获取模板
	templateService := NewLLMTemplateService()
	template, err := templateService.GetTemplateByID(req.TemplateID)
	if err != nil {
		return nil, err
	}

	if !template.IsActive {
		return nil, errors.New("模板未激活")
	}

	// 检查模板是否属于指定配置
	if template.ConfigID != req.ConfigID {
		return nil, errors.New("模板不属于指定配置")
	}

	// 获取配置
	configService := NewLLMConfigService()
	config, err := configService.GetLLMConfigByID(req.ConfigID)
	if err != nil {
		return nil, err
	}

	if !config.IsActive {
		return nil, errors.New("配置未激活")
	}

	// 替换模板变量
	content := template.TemplateContent
	if req.Variables != nil {
		for key, value := range req.Variables {
			placeholder := fmt.Sprintf("{{%s}}", key)
			content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
		}
	}

	// 构建消息列表
	messages := make([]map[string]interface{}, 0)
	
	// 如果有系统提示词，添加到消息列表开头
	if config.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": config.SystemPrompt,
		})
	}

	// 添加用户消息（使用模板内容）
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": content,
	})

	// 构建参数
	params := make(map[string]interface{})
	if req.Temperature != nil {
		params["temperature"] = *req.Temperature
	} else {
		params["temperature"] = config.Temperature
	}
	if req.MaxTokens != nil {
		params["max_tokens"] = *req.MaxTokens
	} else {
		params["max_tokens"] = config.MaxTokens
	}

	// 获取提供商
	provider := GetLLMProvider(config.Provider)
	if provider == nil {
		return nil, errors.New("不支持的提供商")
	}

	// 记录开始时间
	startTime := time.Now()

	// 调用API（带重试）
	var response *LLMResponse
	var lastErr error
	for i := 0; i <= config.MaxRetries; i++ {
		response, lastErr = provider.Call(config, messages, params)
		if lastErr == nil {
			break
		}
		if i < config.MaxRetries {
			logger.Warnf("LLM调用失败，重试 %d/%d: %v", i+1, config.MaxRetries, lastErr)
			time.Sleep(time.Duration(i+1) * time.Second) // 递增延迟
		}
	}

	if lastErr != nil {
		// 记录失败日志
		s.recordCallLog(c, config, &template.ID, nil, nil, "error", lastErr.Error(), nil, nil, nil, time.Since(startTime))
		return nil, fmt.Errorf("LLM调用失败: %v", lastErr)
	}

	// 计算耗时
	duration := time.Since(startTime)

	// 记录成功日志
	s.recordCallLog(c, config, &template.ID, nil, response, "success", "", response.TokensUsed, response.PromptTokens, response.CompletionTokens, duration)

	// 构建响应
	result := &schemas.LLMCallResponse{
		Content:          response.Content,
		ResponseData:     response.ResponseData,
		TokensUsed:       response.TokensUsed,
		PromptTokens:     response.PromptTokens,
		CompletionTokens: response.CompletionTokens,
	}

	return result, nil
}

// recordCallLog 记录调用日志
func (s *LLMService) recordCallLog(c *gin.Context, config *models.LLMConfig, templateID *uint, callReq *schemas.LLMCallRequest, response *LLMResponse, status, errorMsg string, tokensUsed, promptTokens, completionTokens *int, duration time.Duration) {
	// 获取分组ID和激活码
	var groupID *uint
	var activationCode string

	if callReq != nil {
		groupID = callReq.GroupID
		activationCode = callReq.ActivationCode
	} else {
		// 从上下文获取（如果是子账号）
		if gid, exists := c.Get("group_id"); exists {
			if gidUint, ok := gid.(uint); ok {
				groupID = &gidUint
			}
		}
		if ac, exists := c.Get("activation_code"); exists {
			if acStr, ok := ac.(string); ok {
				activationCode = acStr
			}
		}
	}

	// 构建请求消息
	var requestMessages models.JSONB
	if callReq != nil {
		requestMessages = models.JSONB{
			"messages": callReq.Messages,
		}
	} else {
		requestMessages = models.JSONB{}
	}

	// 构建请求参数
	requestParams := models.JSONB{}
	if callReq != nil {
		if callReq.Temperature != nil {
			requestParams["temperature"] = *callReq.Temperature
		}
		if callReq.MaxTokens != nil {
			requestParams["max_tokens"] = *callReq.MaxTokens
		}
	}

	// 构建响应数据
	var responseData models.JSONB
	if response != nil && response.ResponseData != nil {
		responseData = models.JSONB(response.ResponseData)
	}

	// 创建日志
	log := &models.LLMCallLog{
		ConfigID:         &config.ID,
		TemplateID:       templateID,
		GroupID:          groupID,
		ActivationCode:   activationCode,
		RequestMessages:  requestMessages,
		RequestParams:    requestParams,
		ResponseContent:  "",
		ResponseData:     responseData,
		Status:           status,
		ErrorMessage:     errorMsg,
		TokensUsed:       tokensUsed,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		CallTime:         time.Now(),
		DurationMs:       intPtr(int(duration.Milliseconds())),
	}

	if response != nil {
		log.ResponseContent = response.Content
	}

	// 保存日志
	if err := s.db.Create(log).Error; err != nil {
		logger.Errorf("保存LLM调用日志失败: %v", err)
		// 不返回错误，因为日志记录失败不应该影响主流程
	}
}

// intPtr 返回int指针
func intPtr(i int) *int {
	return &i
}

// GetCallLogList 获取调用日志列表
func (s *LLMService) GetCallLogList(c *gin.Context, params *schemas.LLMCallLogQueryParams) ([]schemas.LLMCallLogResponse, int64, error) {
	var logs []models.LLMCallLog
	var total int64

	query := s.db.Model(&models.LLMCallLog{})

	// 配置ID筛选
	if params.ConfigID != nil {
		query = query.Where("config_id = ?", *params.ConfigID)
	}

	// 模板ID筛选
	if params.TemplateID != nil {
		query = query.Where("template_id = ?", *params.TemplateID)
	}

	// 分组ID筛选
	if params.GroupID != nil {
		query = query.Where("group_id = ?", *params.GroupID)
	}

	// 激活码筛选
	if params.ActivationCode != "" {
		query = query.Where("activation_code = ?", params.ActivationCode)
	}

	// 状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 时间范围筛选
	if params.StartTime != "" {
		query = query.Where("call_time >= ?", params.StartTime)
	}
	if params.EndTime != "" {
		query = query.Where("call_time <= ?", params.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 查询列表
	if err := query.Order("call_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	var list []schemas.LLMCallLogResponse
	for _, log := range logs {
		requestMessages := make([]map[string]interface{}, 0)
		if log.RequestMessages != nil {
			if msgs, ok := log.RequestMessages["messages"].([]interface{}); ok {
				for _, msg := range msgs {
					if msgMap, ok := msg.(map[string]interface{}); ok {
						requestMessages = append(requestMessages, msgMap)
					}
				}
			}
		}

		requestParams := make(map[string]interface{})
		if log.RequestParams != nil {
			requestParams = log.RequestParams
		}

		responseData := make(map[string]interface{})
		if log.ResponseData != nil {
			responseData = log.ResponseData
		}

		list = append(list, schemas.LLMCallLogResponse{
			ID:               log.ID,
			ConfigID:         log.ConfigID,
			TemplateID:       log.TemplateID,
			GroupID:          log.GroupID,
			ActivationCode:   log.ActivationCode,
			RequestMessages:  requestMessages,
			RequestParams:    requestParams,
			ResponseContent:  log.ResponseContent,
			ResponseData:     responseData,
			Status:           log.Status,
			ErrorMessage:     log.ErrorMessage,
			TokensUsed:       log.TokensUsed,
			PromptTokens:     log.PromptTokens,
			CompletionTokens: log.CompletionTokens,
			CallTime:         log.CallTime.Format(time.RFC3339),
			DurationMs:       log.DurationMs,
		})
	}

	return list, total, nil
}

