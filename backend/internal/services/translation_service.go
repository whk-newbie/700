package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/pkg/database"
	"line-management/pkg/logger"
	redisClient "line-management/pkg/redis"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TranslationService 翻译服务
type TranslationService struct {
	db          *gorm.DB
	maxMessages int           // 最大消息数，超过后重置对话
	ctx         context.Context
}

// ConversationHistory 对话历史
type ConversationHistory struct {
	Messages []map[string]interface{} `json:"messages"`   // 对话消息列表
	LastUsed time.Time                 `json:"last_used"` // 最后使用时间
}

// 系统提示词（System Prompt）
const translationSystemPrompt = `你是一个专业的中日翻译助手。请遵循以下规则：
1. 如果输入是中文，请翻译成日文
2. 如果输入是日文，请翻译成中文
3. 只返回翻译结果，不要添加任何解释或额外内容
4. 保持原文的语气和风格
5. 对于专业术语，请使用标准翻译`

var translationServiceInstance *TranslationService
var translationServiceOnce sync.Once

// GetTranslationService 获取翻译服务单例
func GetTranslationService() *TranslationService {
	translationServiceOnce.Do(func() {
		translationServiceInstance = &TranslationService{
			db:          database.GetDB(),
			maxMessages: 100, // 默认最多保留100条消息（50轮对话）
			ctx:         context.Background(),
		}
	})
	return translationServiceInstance
}

// getRedisKey 获取Redis中存储对话历史的key
func (s *TranslationService) getRedisKey(userKey string) string {
	return fmt.Sprintf("translation:conversation:%s", userKey)
}

// DetectLanguage 检测语言类型
func (s *TranslationService) DetectLanguage(text string) string {
	// 检测是否包含中文字符
	chineseRegex := regexp.MustCompile(`[\p{Han}]`)
	// 检测是否包含日文假名
	japaneseRegex := regexp.MustCompile(`[\p{Hiragana}\p{Katakana}]`)
	
	hasChinese := chineseRegex.MatchString(text)
	hasJapanese := japaneseRegex.MatchString(text)
	
	// 如果同时包含中日文字符，优先判断为中文
	if hasChinese {
		return "zh"
	}
	if hasJapanese {
		return "ja"
	}
	
	// 默认返回中文
	return "zh"
}

// getUserKey 获取用户的唯一标识
func (s *TranslationService) getUserKey(c *gin.Context) string {
	// 优先使用用户ID
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uint); ok {
			return fmt.Sprintf("user_%d", uid)
		}
	}
	
	// 子账号使用激活码
	if activationCode, exists := c.Get("activation_code"); exists {
		if code, ok := activationCode.(string); ok && code != "" {
			return fmt.Sprintf("group_%s", code)
		}
	}
	
	// 默认返回空字符串（不使用对话历史）
	return ""
}

// getConversationHistory 获取对话历史（从Redis）
func (s *TranslationService) getConversationHistory(userKey string) []map[string]interface{} {
	systemMessage := []map[string]interface{}{
		{
			"role":    "system",
			"content": translationSystemPrompt,
		},
	}
	
	if userKey == "" {
		// 如果没有用户标识，返回新的对话
		return systemMessage
	}
	
	// 从Redis获取对话历史
	redisKey := s.getRedisKey(userKey)
	data, err := redisClient.GetClient().Get(s.ctx, redisKey).Result()
	if err != nil {
		// 如果没有找到或出错，返回新的对话
		return systemMessage
	}
	
	// 反序列化对话历史
	var history ConversationHistory
	if err := json.Unmarshal([]byte(data), &history); err != nil {
		logger.Errorf("反序列化对话历史失败: %v", err)
		return systemMessage
	}
	
	// 检查是否超过最大消息数
	if len(history.Messages) >= s.maxMessages {
		// 重置对话，只保留系统提示词
		logger.Infof("用户 %s 的对话历史达到上限 %d，重置对话", userKey, s.maxMessages)
		return systemMessage
	}
	
	// 返回现有历史
	return history.Messages
}

// updateConversationHistory 更新对话历史（存储到Redis）
func (s *TranslationService) updateConversationHistory(userKey string, userMessage, assistantMessage map[string]interface{}) {
	if userKey == "" {
		return
	}
	
	// 从Redis获取现有历史
	redisKey := s.getRedisKey(userKey)
	var history ConversationHistory
	
	data, err := redisClient.GetClient().Get(s.ctx, redisKey).Result()
	if err == nil {
		// 如果存在历史，反序列化
		if err := json.Unmarshal([]byte(data), &history); err != nil {
			logger.Errorf("反序列化对话历史失败: %v", err)
			// 创建新的历史
			history = ConversationHistory{
				Messages: []map[string]interface{}{
					{
						"role":    "system",
						"content": translationSystemPrompt,
					},
				},
			}
		}
	} else {
		// 如果不存在，创建新的历史
		history = ConversationHistory{
			Messages: []map[string]interface{}{
				{
					"role":    "system",
					"content": translationSystemPrompt,
				},
			},
		}
	}
	
	// 检查是否需要重置对话
	if len(history.Messages) >= s.maxMessages {
		// 重置对话，只保留系统提示词
		history.Messages = []map[string]interface{}{
			{
				"role":    "system",
				"content": translationSystemPrompt,
			},
		}
	}
	
	// 添加新的对话消息
	history.Messages = append(history.Messages, userMessage, assistantMessage)
	history.LastUsed = time.Now()
	
	// 序列化并存储到Redis，设置1小时过期
	jsonData, err := json.Marshal(history)
	if err != nil {
		logger.Errorf("序列化对话历史失败: %v", err)
		return
	}
	
	// 存储到Redis，1小时后自动过期
	if err := redisClient.GetClient().Set(s.ctx, redisKey, jsonData, 1*time.Hour).Err(); err != nil {
		logger.Errorf("保存对话历史到Redis失败: %v", err)
	}
}

// Translate 执行翻译
func (s *TranslationService) Translate(c *gin.Context, req *schemas.TranslateRequest) (*schemas.TranslateResponse, error) {
	// 检测源语言
	sourceLang := s.DetectLanguage(req.Text)
	targetLang := "ja"
	if sourceLang == "ja" {
		targetLang = "zh"
	}
	
	// 获取用户标识
	userKey := s.getUserKey(c)
	
	// 获取对话历史
	messages := s.getConversationHistory(userKey)
	
	// 添加用户消息
	userMessage := map[string]interface{}{
		"role":    "user",
		"content": req.Text,
	}
	messages = append(messages, userMessage)
	
	// 获取OpenAI API Key配置
	configService := NewLLMConfigService()
	config, err := configService.GetOpenAIAPIKey()
	if err != nil {
		logger.Errorf("获取OpenAI API Key失败: %v", err)
		return nil, fmt.Errorf("获取OpenAI API Key失败: %v", err)
	}
	
	if config.APIKey == "" {
		return nil, fmt.Errorf("未配置OpenAI API Key，请先配置")
	}
	
	// 解密API Key
	encryptionService := GetEncryptionService()
	apiKey, err := encryptionService.Decrypt(config.APIKey)
	if err != nil {
		logger.Errorf("解密API Key失败: %v", err)
		return nil, fmt.Errorf("解密API Key失败: %v", err)
	}
	
	// 构建OpenAI API请求体
	requestBody := map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"messages":    messages,
		"temperature": 0.3, // 较低的温度以获得更准确的翻译
		"max_tokens":  2000,
	}
	
	// 记录开始时间
	startTime := time.Now()
	
	// 调用OpenAI API
	apiURL := "https://api.openai.com/v1"
	timeoutSeconds := 30
	response, err := ProxyToOpenAI(apiURL, apiKey, requestBody, timeoutSeconds)
	
	// 计算耗时
	duration := time.Since(startTime)
	
	// 构建请求消息用于日志记录
	requestMessages := models.JSONB{
		"messages": messages,
	}
	
	// 构建请求参数用于日志记录
	requestParams := models.JSONB{
		"temperature": 0.3,
		"max_tokens":  2000,
	}
	
	// 记录调用日志
	s.recordTranslationLog(c, config, requestMessages, requestParams, response, err, duration, sourceLang, targetLang)
	
	if err != nil {
		logger.Errorf("调用OpenAI API失败: %v", err)
		return nil, fmt.Errorf("调用OpenAI API失败: %v", err)
	}
	
	// 提取翻译结果
	var translatedText string
	var tokensUsed, promptTokens, completionTokens *int
	
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					translatedText = content
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
	
	// 更新对话历史
	assistantMessage := map[string]interface{}{
		"role":    "assistant",
		"content": translatedText,
	}
	s.updateConversationHistory(userKey, userMessage, assistantMessage)
	
	return &schemas.TranslateResponse{
		OriginalText:     req.Text,
		TranslatedText:   translatedText,
		SourceLanguage:   sourceLang,
		TargetLanguage:   targetLang,
		TokensUsed:       tokensUsed,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
	}, nil
}

// recordTranslationLog 记录翻译调用日志
func (s *TranslationService) recordTranslationLog(c *gin.Context, config *models.LLMConfig, requestMessages, requestParams models.JSONB, response map[string]interface{}, err error, duration time.Duration, sourceLang, targetLang string) {
	// 获取用户和分组信息
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
	
	// 在请求参数中添加翻译相关信息
	requestParams["source_language"] = sourceLang
	requestParams["target_language"] = targetLang
	requestParams["api_type"] = "translation"
	
	// 创建日志
	log := &models.LLMCallLog{
		ConfigID:         &config.ID,
		TemplateID:       nil, // 翻译调用不使用模板
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
		logger.Errorf("保存翻译调用日志失败: %v", err)
	}
}

