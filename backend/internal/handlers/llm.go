package handlers

import (
	"encoding/json"
	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetOpenAIAPIKey 获取OpenAI API Key配置
// @Summary 获取OpenAI API Key配置
// @Description 获取OpenAI API Key配置（管理员专用）
// @Tags 大模型配置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} schemas.OpenAIAPIKeyResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/llm/openai-key [get]
func GetOpenAIAPIKey(c *gin.Context) {
	configService := services.NewLLMConfigService()
	config, err := configService.GetOpenAIAPIKey()
	if err != nil {
		logger.Errorf("获取OpenAI API Key失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取OpenAI API Key失败", "internal_error")
		return
	}

	response := schemas.OpenAIAPIKeyResponse{
		HasKey:    config.APIKey != "",
		UpdatedAt: config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// GetRSAPublicKey 获取RSA公钥
// @Summary 获取RSA公钥
// @Description 获取RSA公钥，用于前端加密API Key
// @Tags 大模型配置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} schemas.RSAPublicKeyResponse
// @Failure 500 {object} schemas.ErrorResponse
// @Router /admin/llm/rsa-public-key [get]
func GetRSAPublicKey(c *gin.Context) {
	rsaService := services.GetRSAService()
	publicKeyPEM, err := rsaService.GetPublicKeyPEM()
	if err != nil {
		logger.Errorf("获取RSA公钥失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取RSA公钥失败", "internal_error")
		return
	}

	response := schemas.RSAPublicKeyResponse{
		PublicKey: publicKeyPEM,
	}

	utils.Success(c, response)
}

// UpdateOpenAIAPIKey 更新OpenAI API Key
// @Summary 更新OpenAI API Key
// @Description 更新OpenAI API Key（管理员专用），API Key需要使用RSA公钥加密后传输
// @Tags 大模型配置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.UpdateOpenAIAPIKeyRequest true "更新OpenAI API Key请求（encrypted_api_key为RSA加密后的Base64字符串）"
// @Success 200 {object} schemas.OpenAIAPIKeyResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/llm/openai-key [put]
func UpdateOpenAIAPIKey(c *gin.Context) {
	var req schemas.UpdateOpenAIAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	// 使用RSA私钥解密API Key
	rsaService := services.GetRSAService()
	apiKey, err := rsaService.Decrypt(req.EncryptedAPIKey)
	if err != nil {
		logger.Errorf("RSA解密API Key失败: %v", err)
		utils.ErrorWithErrorCode(c, 1002, "解密API Key失败，请确保使用正确的RSA公钥加密", "decrypt_failed")
		return
	}

	// 使用解密后的明文API Key更新配置
	configService := services.NewLLMConfigService()
	config, err := configService.UpdateOpenAIAPIKeyWithPlainText(apiKey)
	if err != nil {
		logger.Warnf("更新OpenAI API Key失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "更新OpenAI API Key失败", "internal_error")
		return
	}

	response := schemas.OpenAIAPIKeyResponse{
		HasKey:    config.APIKey != "",
		UpdatedAt: config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// GetLLMCallLogs 获取调用日志列表
// @Summary 获取调用日志列表
// @Description 获取调用日志列表（管理员专用，支持分页和筛选）
// @Tags 大模型调用记录
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param config_id query int false "配置ID"
// @Param template_id query int false "模板ID"
// @Param group_id query int false "分组ID"
// @Param activation_code query string false "激活码"
// @Param status query string false "状态" Enums(success, error)
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/llm/call-logs [get]
func GetLLMCallLogs(c *gin.Context) {
	var params schemas.LLMCallLogQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	llmService := services.NewLLMService()
	list, total, err := llmService.GetCallLogList(c, &params)
	if err != nil {
		logger.Errorf("获取调用日志列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取调用日志列表失败", "internal_error")
		return
	}

	// 分页参数
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	utils.SuccessWithPagination(c, list, page, pageSize, total)
}

// boolPtr 返回bool指针
func boolPtr(b bool) *bool {
	return &b
}

// TranslateText 文本翻译接口（中日互译）
// @Summary 文本翻译（中日互译）
// @Description 自动检测语言并翻译（中文翻译成日文，日文翻译成中文），支持对话历史复用
// @Tags 大模型调用
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.TranslateRequest true "翻译请求"
// @Success 200 {object} schemas.TranslateResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 500 {object} schemas.ErrorResponse
// @Router /llm/translate [post]
func TranslateText(c *gin.Context) {
	var req schemas.TranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误: "+err.Error(), "invalid_params")
		return
	}

	// 调用翻译服务
	translationService := services.GetTranslationService()
	result, err := translationService.Translate(c, &req)
	if err != nil {
		logger.Errorf("翻译失败: %v", err)
		
		// 根据错误类型返回不同的错误码
		if err.Error() == "未配置OpenAI API Key，请先配置" {
			utils.ErrorWithErrorCode(c, 4001, err.Error(), "key_not_configured")
		} else {
			utils.ErrorWithErrorCode(c, 7001, "翻译失败: "+err.Error(), "translation_failed")
		}
		return
	}

	utils.Success(c, result)
}

// ProxyOpenAIAPI OpenAI API转发接口
// @Summary OpenAI API转发
// @Description 转发OpenAI API请求，前端传参格式与OpenAI文档一致，后端自动添加授权码
// @Tags 大模型调用
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.OpenAIProxyRequest true "OpenAI API请求（不包含授权码）"
// @Success 200 {object} map[string]interface{} "OpenAI API响应"
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 500 {object} schemas.ErrorResponse
// @Router /llm/proxy/openai [post]
func ProxyOpenAIAPI(c *gin.Context) {
	var req schemas.OpenAIProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误: "+err.Error(), "invalid_params")
		return
	}

	// 获取OpenAI API Key配置
	configService := services.NewLLMConfigService()
	config, err := configService.GetOpenAIAPIKey()
	if err != nil {
		logger.Errorf("获取OpenAI API Key失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取OpenAI API Key失败", "get_key_failed")
		return
	}

	if config.APIKey == "" {
		utils.ErrorWithErrorCode(c, 4001, "未配置OpenAI API Key，请先配置", "key_not_configured")
		return
	}

	// 解密API Key
	encryptionService := services.GetEncryptionService()
	apiKey, err := encryptionService.Decrypt(config.APIKey)
	if err != nil {
		logger.Errorf("解密API Key失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "解密API Key失败", "decrypt_failed")
		return
	}

	// 构建OpenAI API请求体（使用前端传来的所有参数）
	var requestBody map[string]interface{}
	requestJSON, err := json.Marshal(req)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "序列化请求参数失败: "+err.Error(), "serialize_failed")
		return
	}
	if err := json.Unmarshal(requestJSON, &requestBody); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "解析请求参数失败: "+err.Error(), "parse_failed")
		return
	}

	// 记录开始时间
	startTime := time.Now()

	// 转发请求到OpenAI API（使用默认的OpenAI API URL和超时时间）
	apiURL := "https://api.openai.com/v1"
	timeoutSeconds := 30
	response, err := services.ProxyToOpenAI(apiURL, apiKey, requestBody, timeoutSeconds)
	
	// 计算耗时
	duration := time.Since(startTime)

	// 记录调用日志
	llmService := services.NewLLMService()
	llmService.RecordProxyCallLog(c, config, req, response, err, duration)

	if err != nil {
		logger.Errorf("转发OpenAI API请求失败: %v", err)
		utils.ErrorWithErrorCode(c, 7001, "转发请求失败: "+err.Error(), "proxy_failed")
		return
	}

	// 直接返回OpenAI的响应
	c.JSON(http.StatusOK, response)
}
