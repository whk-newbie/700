package handlers

import (
	"strconv"
	"strings"
	"time"

	"line-management/internal/schemas"
	"line-management/internal/services"
	"line-management/internal/utils"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GetLineAccounts 获取Line账号列表
// @Summary 获取Line账号列表
// @Description 获取Line账号列表（支持分页和筛选）
// @Tags Line账号管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param group_id query int false "分组ID"
// @Param platform_type query string false "平台类型" Enums(line, line_business)
// @Param online_status query string false "在线状态" Enums(online, offline, user_logout, abnormal_offline)
// @Param activation_code query string false "激活码"
// @Param search query string false "搜索（Line ID或显示名称）"
// @Success 200 {object} utils.PaginationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /line-accounts [get]
func GetLineAccounts(c *gin.Context) {
	var params schemas.LineAccountQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	lineAccountService := services.NewLineAccountService()
	list, total, err := lineAccountService.GetLineAccountList(c, &params)
	if err != nil {
		logger.Errorf("获取Line账号列表失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "获取Line账号列表失败", "internal_error")
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

// CreateLineAccount 创建Line账号
// @Summary 创建Line账号
// @Description 创建新的Line账号
// @Tags Line账号管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.CreateLineAccountRequest true "创建Line账号请求"
// @Success 200 {object} schemas.LineAccountListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /line-accounts [post]
func CreateLineAccount(c *gin.Context) {
	var req schemas.CreateLineAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	lineAccountService := services.NewLineAccountService()
	account, err := lineAccountService.CreateLineAccount(c, &req)
	if err != nil {
		logger.Warnf("创建Line账号失败: %v", err)
		if err.Error() == "分组不存在" {
			utils.ErrorWithErrorCode(c, 3002, err.Error(), "group_not_found")
		} else if err.Error() == "分组已被禁用" {
			utils.ErrorWithErrorCode(c, 4001, err.Error(), "group_disabled")
		} else if strings.Contains(err.Error(), "账号数量限制") {
			utils.ErrorWithErrorCode(c, 4002, err.Error(), "account_limit_exceeded")
		} else if err.Error() == "该Line ID在此分组中已存在" {
			utils.ErrorWithErrorCode(c, 4003, err.Error(), "line_id_exists")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "创建Line账号失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	var lastActiveAt, lastOnlineTime, firstLoginAt *string
	if account.LastActiveAt != nil {
		timeStr := account.LastActiveAt.Format(time.RFC3339)
		lastActiveAt = &timeStr
	}
	if account.LastOnlineTime != nil {
		timeStr := account.LastOnlineTime.Format(time.RFC3339)
		lastOnlineTime = &timeStr
	}
	if account.FirstLoginAt != nil {
		timeStr := account.FirstLoginAt.Format(time.RFC3339)
		firstLoginAt = &timeStr
	}

	response := schemas.LineAccountListResponse{
		ID:             account.ID,
		GroupID:        account.GroupID,
		ActivationCode: account.ActivationCode,
		PlatformType:   account.PlatformType,
		LineID:         account.LineID,
		DisplayName:    account.DisplayName,
		PhoneNumber:    account.PhoneNumber,
		ProfileURL:     account.ProfileURL,
		AvatarURL:      account.AvatarURL,
		Bio:            account.Bio,
		StatusMessage:  account.StatusMessage,
		AddFriendLink:  account.AddFriendLink,
		QRCodePath:     account.QRCodePath,
		OnlineStatus:   account.OnlineStatus,
		LastActiveAt:   lastActiveAt,
		LastOnlineTime: lastOnlineTime,
		FirstLoginAt:   firstLoginAt,
		AccountRemark:  account.AccountRemark,
		CreatedAt:      account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      account.UpdatedAt.Format(time.RFC3339),
	}

	utils.SuccessWithMessage(c, "创建成功", response)
}

// UpdateLineAccount 更新Line账号
// @Summary 更新Line账号
// @Description 更新Line账号信息
// @Tags Line账号管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "账号ID"
// @Param request body schemas.UpdateLineAccountRequest true "更新Line账号请求"
// @Success 200 {object} schemas.LineAccountListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /line-accounts/:id [put]
func UpdateLineAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的账号ID", "invalid_id")
		return
	}

	var req schemas.UpdateLineAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	lineAccountService := services.NewLineAccountService()
	account, err := lineAccountService.UpdateLineAccount(c, uint(id), &req)
	if err != nil {
		logger.Warnf("更新Line账号失败: %v", err)
		if err.Error() == "账号不存在" {
			utils.ErrorWithErrorCode(c, 3003, err.Error(), "account_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "更新Line账号失败", "internal_error")
		}
		return
	}

	// 转换为响应格式
	var lastActiveAt, lastOnlineTime, firstLoginAt *string
	if account.LastActiveAt != nil {
		timeStr := account.LastActiveAt.Format(time.RFC3339)
		lastActiveAt = &timeStr
	}
	if account.LastOnlineTime != nil {
		timeStr := account.LastOnlineTime.Format(time.RFC3339)
		lastOnlineTime = &timeStr
	}
	if account.FirstLoginAt != nil {
		timeStr := account.FirstLoginAt.Format(time.RFC3339)
		firstLoginAt = &timeStr
	}

	response := schemas.LineAccountListResponse{
		ID:             account.ID,
		GroupID:        account.GroupID,
		ActivationCode: account.ActivationCode,
		PlatformType:   account.PlatformType,
		LineID:         account.LineID,
		DisplayName:    account.DisplayName,
		PhoneNumber:    account.PhoneNumber,
		ProfileURL:     account.ProfileURL,
		AvatarURL:      account.AvatarURL,
		Bio:            account.Bio,
		StatusMessage:  account.StatusMessage,
		AddFriendLink:  account.AddFriendLink,
		QRCodePath:     account.QRCodePath,
		OnlineStatus:   account.OnlineStatus,
		LastActiveAt:   lastActiveAt,
		LastOnlineTime: lastOnlineTime,
		FirstLoginAt:   firstLoginAt,
		AccountRemark:  account.AccountRemark,
		CreatedAt:      account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      account.UpdatedAt.Format(time.RFC3339),
	}

	utils.SuccessWithMessage(c, "更新成功", response)
}

// DeleteLineAccount 删除Line账号
// @Summary 删除Line账号
// @Description 软删除Line账号
// @Tags Line账号管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "账号ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /line-accounts/:id [delete]
func DeleteLineAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的账号ID", "invalid_id")
		return
	}

	// 获取当前用户ID（用于记录删除者）
	userID, exists := c.Get("user_id")
	var deletedBy *uint
	if exists {
		if uid, ok := userID.(uint); ok {
			deletedBy = &uid
		}
	}

	lineAccountService := services.NewLineAccountService()
	if err := lineAccountService.DeleteLineAccount(c, uint(id), deletedBy); err != nil {
		logger.Warnf("删除Line账号失败: %v", err)
		if err.Error() == "账号不存在" {
			utils.ErrorWithErrorCode(c, 3003, err.Error(), "account_not_found")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "删除Line账号失败", "internal_error")
		}
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// GenerateQRCode 生成二维码
// @Summary 生成二维码
// @Description 为Line账号生成二维码（二维码内容为Line添加好友链接）
// @Tags Line账号管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "账号ID"
// @Param content query string false "二维码内容（默认为Line添加好友链接：https://line.me/ti/p/~{line_id}）"
// @Success 200 {object} schemas.GenerateQRCodeResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 404 {object} schemas.ErrorResponse
// @Router /line-accounts/:id/generate-qr [post]
func GenerateQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorWithErrorCode(c, 1001, "无效的账号ID", "invalid_id")
		return
	}

	// 获取二维码内容（可选，如果为空则使用账号的添加好友链接）
	content := c.Query("content")

	qrService := services.NewQRService()
	qrCodePath, err := qrService.GenerateQRCode(uint(id), content)
	if err != nil {
		logger.Warnf("生成二维码失败: %v", err)
		if err.Error() == "账号不存在" {
			utils.ErrorWithErrorCode(c, 3003, err.Error(), "account_not_found")
		} else if err.Error() == "账号没有添加好友链接，无法生成二维码" {
			utils.ErrorWithErrorCode(c, 4004, err.Error(), "no_add_friend_link")
		} else {
			utils.ErrorWithErrorCode(c, 5001, "生成二维码失败", "internal_error")
		}
		return
	}

	// 构建完整的访问URL（这里使用相对路径，前端可以根据实际情况拼接完整URL）
	qrCodeURL := qrCodePath // 或者可以拼接完整URL: fmt.Sprintf("http://%s%s", c.Request.Host, qrCodePath)

	response := schemas.GenerateQRCodeResponse{
		QRCodePath: qrCodePath,
		QRCodeURL:  qrCodeURL,
	}

	utils.SuccessWithMessage(c, "生成成功", response)
}

// BatchDeleteLineAccounts 批量删除Line账号
// @Summary 批量删除Line账号
// @Description 批量软删除Line账号
// @Tags Line账号管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.BatchDeleteLineAccountsRequest true "批量删除请求"
// @Success 200 {object} schemas.BatchOperationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /line-accounts/batch/delete [post]
func BatchDeleteLineAccounts(c *gin.Context) {
	var req schemas.BatchDeleteLineAccountsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	// 获取当前用户ID（用于记录删除者）
	userID, exists := c.Get("user_id")
	var deletedBy *uint
	if exists {
		if uid, ok := userID.(uint); ok {
			deletedBy = &uid
		}
	}

	lineAccountService := services.NewLineAccountService()
	successCount, failedIDs, err := lineAccountService.BatchDeleteLineAccounts(c, req.IDs, deletedBy)
	if err != nil {
		logger.Errorf("批量删除Line账号失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "批量删除Line账号失败", "internal_error")
		return
	}

	failCount := len(failedIDs)
	response := schemas.BatchOperationResponse{
		SuccessCount: successCount,
		FailCount:    failCount,
	}
	if failCount > 0 {
		response.FailedIDs = failedIDs
	}

	utils.SuccessWithMessage(c, "批量删除完成", response)
}

// BatchUpdateLineAccounts 批量更新Line账号
// @Summary 批量更新Line账号
// @Description 批量更新Line账号的在线状态等
// @Tags Line账号管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.BatchUpdateLineAccountsRequest true "批量更新请求"
// @Success 200 {object} schemas.BatchOperationResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Router /line-accounts/batch/update [post]
func BatchUpdateLineAccounts(c *gin.Context) {
	var req schemas.BatchUpdateLineAccountsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithErrorCode(c, 1001, "请求参数错误", "invalid_params")
		return
	}

	// 验证至少有一个更新字段
	if req.OnlineStatus == "" {
		utils.ErrorWithErrorCode(c, 1001, "至少需要提供一个更新字段", "invalid_params")
		return
	}

	lineAccountService := services.NewLineAccountService()
	successCount, failedIDs, err := lineAccountService.BatchUpdateLineAccounts(c, req.IDs, &req)
	if err != nil {
		logger.Errorf("批量更新Line账号失败: %v", err)
		utils.ErrorWithErrorCode(c, 5001, "批量更新Line账号失败", "internal_error")
		return
	}

	failCount := len(failedIDs)
	response := schemas.BatchOperationResponse{
		SuccessCount: successCount,
		FailCount:    failCount,
	}
	if failCount > 0 {
		response.FailedIDs = failedIDs
	}

	utils.SuccessWithMessage(c, "批量更新完成", response)
}

