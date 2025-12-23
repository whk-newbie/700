package handlers

import (
	"strconv"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GetLLMConfigs 获取LLM配置列表
// @Summary 获取LLM配置列表
// @Description 获取LLM配置列表（管理员专用，支持分页和筛选）
// @Tags 大模型配置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param provider query string false "提供商"
// @Param is_active query bool false "是否激活"
// @Param search query string false "搜索（名称）"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/llm/configs [get]
func GetLLMConfigs(c *gin.Context) {
	var params schemas.LLMConfigQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	configService := services.NewLLMConfigService()
	list, total, err := configService.GetLLMConfigList(c, &params)
	if err != nil {
		logger.Errorf("获取LLM配置列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取LLM配置列表失败", "internal_error")
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

// CreateLLMConfig 创建LLM配置
// @Summary 创建LLM配置
// @Description 创建LLM配置（管理员专用）
// @Tags 大模型配置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.CreateLLMConfigRequest true "创建LLM配置请求"
// @Success 200 {object} schemas.LLMConfigResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/llm/configs [post]
func CreateLLMConfig(c *gin.Context) {
	var req schemas.CreateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	configService := services.NewLLMConfigService()
	config, err := configService.CreateLLMConfig(c, &req)
	if err != nil {
		logger.Warnf("创建LLM配置失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "创建LLM配置失败", "internal_error")
		return
	}

	// 转换为响应格式
	response := schemas.LLMConfigResponse{
		ID:              config.ID,
		Name:            config.Name,
		Provider:        config.Provider,
		APIURL:          config.APIURL,
		Model:           config.Model,
		MaxTokens:       config.MaxTokens,
		Temperature:     config.Temperature,
		TopP:            config.TopP,
		FrequencyPenalty: config.FrequencyPenalty,
		PresencePenalty:  config.PresencePenalty,
		SystemPrompt:    config.SystemPrompt,
		TimeoutSeconds:  config.TimeoutSeconds,
		MaxRetries:      config.MaxRetries,
		IsActive:        config.IsActive,
		CreatedBy:       config.CreatedBy,
		CreatedAt:       config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// UpdateLLMConfig 更新LLM配置
// @Summary 更新LLM配置
// @Description 更新LLM配置（管理员专用）
// @Tags 大模型配置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Param request body schemas.UpdateLLMConfigRequest true "更新LLM配置请求"
// @Success 200 {object} schemas.LLMConfigResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /admin/llm/configs/{id} [put]
func UpdateLLMConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的配置ID", "invalid_id")
		return
	}

	var req schemas.UpdateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	configService := services.NewLLMConfigService()
	config, err := configService.UpdateLLMConfig(c, uint(id), &req)
	if err != nil {
		logger.Warnf("更新LLM配置失败: %v", err)
		if err.Error() == "配置不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "config_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "更新LLM配置失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	response := schemas.LLMConfigResponse{
		ID:              config.ID,
		Name:            config.Name,
		Provider:        config.Provider,
		APIURL:          config.APIURL,
		Model:           config.Model,
		MaxTokens:       config.MaxTokens,
		Temperature:     config.Temperature,
		TopP:            config.TopP,
		FrequencyPenalty: config.FrequencyPenalty,
		PresencePenalty:  config.PresencePenalty,
		SystemPrompt:    config.SystemPrompt,
		TimeoutSeconds:  config.TimeoutSeconds,
		MaxRetries:      config.MaxRetries,
		IsActive:        config.IsActive,
		CreatedBy:       config.CreatedBy,
		CreatedAt:       config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// DeleteLLMConfig 删除LLM配置
// @Summary 删除LLM配置
// @Description 删除LLM配置（管理员专用）
// @Tags 大模型配置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /admin/llm/configs/{id} [delete]
func DeleteLLMConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的配置ID", "invalid_id")
		return
	}

	configService := services.NewLLMConfigService()
	if err := configService.DeleteLLMConfig(c, uint(id)); err != nil {
		logger.Warnf("删除LLM配置失败: %v", err)
		if err.Error() == "配置不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "config_not_found")
		} else if err.Error() == "该配置下还有模板，无法删除" {
			utils.ErrorWithErrorCode(c, 4001, err.Error(), "config_has_templates")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "删除LLM配置失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// TestLLMConfig 测试LLM配置连接
// @Summary 测试LLM配置连接
// @Description 测试LLM配置连接（管理员专用）
// @Tags 大模型配置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /admin/llm/configs/{id}/test [post]
func TestLLMConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的配置ID", "invalid_id")
		return
	}

	configService := services.NewLLMConfigService()
	if err := configService.TestLLMConfig(uint(id)); err != nil {
		logger.Warnf("测试LLM配置失败: %v", err)
		if err.Error() == "配置不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "config_not_found")
		} else if err.Error() == "配置未激活" {
			utils.ErrorWithErrorCode(c, 4001, err.Error(), "config_inactive")
		} else {
			utils.ErrorWithErrorCode(c, 7001, "测试连接失败: "+err.Error(), "test_connection_failed")
		}
		return
	}

	utils.SuccessWithMessage(c, "连接测试成功", nil)
}

// GetPromptTemplates 获取Prompt模板列表
// @Summary 获取Prompt模板列表
// @Description 获取Prompt模板列表（管理员专用，支持分页和筛选）
// @Tags Prompt模板
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param config_id query int false "配置ID"
// @Param is_active query bool false "是否激活"
// @Param search query string false "搜索（模板名称）"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/llm/templates [get]
func GetPromptTemplates(c *gin.Context) {
	var params schemas.PromptTemplateQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	templateService := services.NewLLMTemplateService()
	list, total, err := templateService.GetTemplateList(c, &params)
	if err != nil {
		logger.Errorf("获取Prompt模板列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取Prompt模板列表失败", "internal_error")
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

// CreatePromptTemplate 创建Prompt模板
// @Summary 创建Prompt模板
// @Description 创建Prompt模板（管理员专用）
// @Tags Prompt模板
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.CreatePromptTemplateRequest true "创建Prompt模板请求"
// @Success 200 {object} schemas.PromptTemplateResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Router /admin/llm/templates [post]
func CreatePromptTemplate(c *gin.Context) {
	var req schemas.CreatePromptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	templateService := services.NewLLMTemplateService()
	template, err := templateService.CreateTemplate(c, &req)
	if err != nil {
		logger.Warnf("创建Prompt模板失败: %v", err)
		if err.Error() == "配置不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "config_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "创建Prompt模板失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	variables := make(map[string]interface{})
	if template.Variables != nil {
		variables = template.Variables
	}

	response := schemas.PromptTemplateResponse{
		ID:             template.ID,
		ConfigID:       template.ConfigID,
		TemplateName:   template.TemplateName,
		TemplateContent: template.TemplateContent,
		Variables:      variables,
		Description:    template.Description,
		IsActive:       template.IsActive,
		CreatedAt:      template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// UpdatePromptTemplate 更新Prompt模板
// @Summary 更新Prompt模板
// @Description 更新Prompt模板（管理员专用）
// @Tags Prompt模板
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param request body schemas.UpdatePromptTemplateRequest true "更新Prompt模板请求"
// @Success 200 {object} schemas.PromptTemplateResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /admin/llm/templates/{id} [put]
func UpdatePromptTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的模板ID", "invalid_id")
		return
	}

	var req schemas.UpdatePromptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	templateService := services.NewLLMTemplateService()
	template, err := templateService.UpdateTemplate(c, uint(id), &req)
	if err != nil {
		logger.Warnf("更新Prompt模板失败: %v", err)
		if err.Error() == "模板不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "template_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "更新Prompt模板失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	variables := make(map[string]interface{})
	if template.Variables != nil {
		variables = template.Variables
	}

	response := schemas.PromptTemplateResponse{
		ID:             template.ID,
		ConfigID:       template.ConfigID,
		TemplateName:   template.TemplateName,
		TemplateContent: template.TemplateContent,
		Variables:      variables,
		Description:    template.Description,
		IsActive:       template.IsActive,
		CreatedAt:      template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.Success(c, response)
}

// DeletePromptTemplate 删除Prompt模板
// @Summary 删除Prompt模板
// @Description 删除Prompt模板（管理员专用）
// @Tags Prompt模板
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 403 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /admin/llm/templates/{id} [delete]
func DeletePromptTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的模板ID", "invalid_id")
		return
	}

	templateService := services.NewLLMTemplateService()
	if err := templateService.DeleteTemplate(c, uint(id)); err != nil {
		logger.Warnf("删除Prompt模板失败: %v", err)
		if err.Error() == "模板不存在" {
			utils.ErrorWithErrorCode(c, 3001, err.Error(), "template_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "删除Prompt模板失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// GetLLMConfigsPublic 获取可用配置（不返回API Key，Windows客户端使用）
// @Summary 获取可用配置
// @Description 获取可用配置（不返回API Key）
// @Tags 大模型调用
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {array} schemas.LLMConfigPublicResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /llm/configs [get]
func GetLLMConfigsPublic(c *gin.Context) {
	configService := services.NewLLMConfigService()
	params := &schemas.LLMConfigQueryParams{
		Page:     1,
		PageSize: 1000,
		IsActive: boolPtr(true),
	}
	list, _, err := configService.GetLLMConfigList(c, params)
	if err != nil {
		logger.Errorf("获取LLM配置列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取LLM配置列表失败", "internal_error")
		return
	}

	// 转换为公开响应格式（不包含API Key）
	var publicList []schemas.LLMConfigPublicResponse
	for _, config := range list {
		publicList = append(publicList, schemas.LLMConfigPublicResponse{
			ID:              config.ID,
			Name:            config.Name,
			Provider:        config.Provider,
			APIURL:          config.APIURL,
			Model:           config.Model,
			MaxTokens:       config.MaxTokens,
			Temperature:     config.Temperature,
			TopP:            config.TopP,
			FrequencyPenalty: config.FrequencyPenalty,
			PresencePenalty:  config.PresencePenalty,
			SystemPrompt:    config.SystemPrompt,
			TimeoutSeconds:  config.TimeoutSeconds,
			MaxRetries:      config.MaxRetries,
			IsActive:        config.IsActive,
		})
	}

	utils.Success(c, publicList)
}

// CallLLM 调用大模型
// @Summary 调用大模型
// @Description 调用大模型
// @Tags 大模型调用
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.LLMCallRequest true "调用请求"
// @Success 200 {object} schemas.LLMCallResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /llm/call [post]
func CallLLM(c *gin.Context) {
	var req schemas.LLMCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	llmService := services.NewLLMService()
	response, err := llmService.CallLLM(c, &req)
	if err != nil {
		logger.Warnf("调用大模型失败: %v", err)
		utils.ErrorWithErrorCode(c, 7001, "调用大模型失败: "+err.Error(), "llm_call_failed")
		return
	}

	utils.Success(c, response)
}

// CallLLMWithTemplate 使用模板调用大模型
// @Summary 使用模板调用大模型
// @Description 使用模板调用大模型
// @Tags 大模型调用
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.LLMCallTemplateRequest true "调用请求"
// @Success 200 {object} schemas.LLMCallResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /llm/call-template [post]
func CallLLMWithTemplate(c *gin.Context) {
	var req schemas.LLMCallTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	llmService := services.NewLLMService()
	response, err := llmService.CallLLMWithTemplate(c, &req)
	if err != nil {
		logger.Warnf("使用模板调用大模型失败: %v", err)
		utils.ErrorWithErrorCode(c, 7001, "调用大模型失败: "+err.Error(), "llm_call_failed")
		return
	}

	utils.Success(c, response)
}

// GetTemplatesPublic 获取模板列表（Windows客户端使用）
// @Summary 获取模板列表
// @Description 获取模板列表
// @Tags 大模型调用
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param config_id query int false "配置ID"
// @Success 200 {array} schemas.PromptTemplateResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /llm/templates [get]
func GetTemplatesPublic(c *gin.Context) {
	var params schemas.PromptTemplateQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	params.Page = 1
	params.PageSize = 1000
	params.IsActive = boolPtr(true)

	templateService := services.NewLLMTemplateService()
	list, _, err := templateService.GetTemplateList(c, &params)
	if err != nil {
		logger.Errorf("获取模板列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取模板列表失败", "internal_error")
		return
	}

	utils.Success(c, list)
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

