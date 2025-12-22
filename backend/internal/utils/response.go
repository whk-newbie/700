package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`                // 业务状态码
	Message   string      `json:"message"`             // 响应消息
	Data      interface{} `json:"data,omitempty"`      // 响应数据
	Timestamp string      `json:"timestamp"`           // 时间戳
	Error     string      `json:"error,omitempty"`     // 错误代码（仅错误时返回）
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      1000,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      1000,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// ErrorWithErrorCode 带错误代码的错误响应
func ErrorWithErrorCode(c *gin.Context, code int, message, errorCode string) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		Data:      nil,
		Error:     errorCode,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// getHTTPStatus 根据业务状态码获取HTTP状态码
func getHTTPStatus(code int) int {
	// 1xxx - 通用错误 -> 400
	if code >= 1001 && code < 2000 {
		return http.StatusBadRequest
	}
	// 2xxx - 认证与授权错误
	if code >= 2001 && code < 3000 {
		if code >= 2001 && code <= 2006 {
			return http.StatusUnauthorized // 2001-2006: 401
		}
		return http.StatusForbidden // 2007-2010: 403
	}
	// 3xxx - 资源不存在错误 -> 404
	if code >= 3001 && code < 4000 {
		return http.StatusNotFound
	}
	// 4xxx - 业务逻辑错误 -> 400
	if code >= 4001 && code < 5000 {
		return http.StatusBadRequest
	}
	// 5xxx - 数据操作错误 -> 500
	if code >= 5001 && code < 6000 {
		return http.StatusInternalServerError
	}
	// 6xxx - WebSocket错误 -> 400
	if code >= 6001 && code < 7000 {
		return http.StatusBadRequest
	}
	// 7xxx - 外部服务错误 -> 502
	if code >= 7001 && code < 8000 {
		return http.StatusBadGateway
	}
	// 8xxx - 系统错误 -> 500
	if code >= 8001 && code < 9000 {
		if code == 8003 {
			return http.StatusServiceUnavailable // 8003: 503
		}
		return http.StatusInternalServerError
	}
	// 9xxx - 业务状态（非错误） -> 200
	if code >= 9001 && code < 10000 {
		return http.StatusOK
	}
	// 默认返回200
	return http.StatusOK
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	List       interface{} `json:"list"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页数量
	Total      int64 `json:"total"`       // 总记录数
	TotalPages int   `json:"total_pages"` // 总页数
}

// SuccessWithPagination 分页成功响应
func SuccessWithPagination(c *gin.Context, list interface{}, page, pageSize int, total int64) {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	SuccessWithMessage(c, "查询成功", PaginationResponse{
		List: list,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

