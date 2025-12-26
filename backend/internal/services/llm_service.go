package services

import (
	"fmt"
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

// RecordProxyCallLog 记录代理调用的日志
func (s *LLMService) RecordProxyCallLog(c *gin.Context, config *models.LLMConfig, req schemas.OpenAIProxyRequest, response map[string]interface{}, err error, duration time.Duration) {
	// 获取用户和分组信息（从上下文）
	var groupID *uint
	var activationCode string
	var userID *uint
	var username string

	// 获取分组信息（子账号才有）
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

	// 获取用户信息
	if uid, exists := c.Get("user_id"); exists {
		if uidUint, ok := uid.(uint); ok {
			userID = &uidUint
		}
	}
	if uname, exists := c.Get("username"); exists {
		if unameStr, ok := uname.(string); ok {
			username = unameStr
		}
	}

	// 构建请求消息
	requestMessages := models.JSONB{
		"messages": req.Messages,
	}

	// 构建请求参数
	requestParams := models.JSONB{}
	if req.Temperature != nil {
		requestParams["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		requestParams["max_tokens"] = *req.MaxTokens
	}
	if req.TopP != nil {
		requestParams["top_p"] = *req.TopP
	}
	if req.Stream != nil {
		requestParams["stream"] = *req.Stream
	}

	// 解析响应数据
	var responseData models.JSONB
	var responseContent string
	var tokensUsed, promptTokens, completionTokens *int
	status := "success"
	errorMsg := ""

	if err != nil {
		status = "error"
		errorMsg = err.Error()
		responseData = models.JSONB{
			"error": errorMsg,
		}
	} else if response != nil {
		responseData = models.JSONB(response)

		// 提取响应内容
		if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						responseContent = content
					}
				}
			}
		}

		// 提取 tokens 信息
		if usage, ok := response["usage"].(map[string]interface{}); ok {
			if total, ok := usage["total_tokens"].(float64); ok {
				totalInt := int(total)
				tokensUsed = &totalInt
			}
			if prompt, ok := usage["prompt_tokens"].(float64); ok {
				promptInt := int(prompt)
				promptTokens = &promptInt
			}
			if completion, ok := usage["completion_tokens"].(float64); ok {
				completionInt := int(completion)
				completionTokens = &completionInt
			}
		}
	}

	// 创建日志
	log := &models.LLMCallLog{
		ConfigID:         &config.ID,
		TemplateID:       nil, // 代理调用不使用模板
		GroupID:          groupID,
		ActivationCode:   activationCode,
		RequestMessages:  requestMessages,
		RequestParams:    requestParams,
		ResponseContent:  responseContent,
		ResponseData:     responseData,
		Status:           status,
		ErrorMessage:     errorMsg,
		TokensUsed:       tokensUsed,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		CallTime:         time.Now(),
		DurationMs:       intPtr(int(duration.Milliseconds())),
	}

	// 如果没有分组信息（可能是管理员调用），在ActivationCode字段记录用户信息
	if groupID == nil && userID != nil && activationCode == "" {
		log.ActivationCode = fmt.Sprintf("ADMIN:%s(ID:%d)", username, *userID)
	}

	// 保存日志
	if err := s.db.Create(log).Error; err != nil {
		logger.Errorf("保存LLM代理调用日志失败: %v", err)
		// 不返回错误，因为日志记录失败不应该影响主流程
	}
}

