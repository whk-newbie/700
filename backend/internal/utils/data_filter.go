package utils

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ApplyDataFilter 应用数据过滤到GORM查询
// 在Service层使用此函数来应用过滤条件
func ApplyDataFilter(c *gin.Context, query *gorm.DB, tableName string) *gorm.DB {
	filter, exists := c.Get("data_filter")
	if !exists || filter == nil {
		return query
	}

	filterMap, ok := filter.(map[string]interface{})
	if !ok {
		return query
	}

	// 根据表名应用不同的过滤条件
	switch tableName {
	case "groups":
		if userID, ok := filterMap["user_id"].(uint); ok {
			query = query.Where("user_id = ?", userID)
		}
		if gID, ok := filterMap["group_id"].(uint); ok {
			query = query.Where("id = ?", gID)
		}
	case "line_accounts", "customers", "follow_up_records", "contact_pool":
		if gID, ok := filterMap["group_id"].(uint); ok {
			query = query.Where("group_id = ?", gID)
		} else if userID, ok := filterMap["user_id"].(uint); ok {
			// 对于普通用户，需要通过groups表关联过滤
			query = query.Joins("JOIN groups ON groups.id = " + tableName + ".group_id").
				Where("groups.user_id = ?", userID)
		}
	case "incoming_logs":
		if gID, ok := filterMap["group_id"].(uint); ok {
			query = query.Where("group_id = ?", gID)
		} else if userID, ok := filterMap["user_id"].(uint); ok {
			query = query.Joins("JOIN groups ON groups.id = incoming_logs.group_id").
				Where("groups.user_id = ?", userID)
		}
	}

	return query
}

