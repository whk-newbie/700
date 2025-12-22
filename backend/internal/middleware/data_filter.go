package middleware

import (
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DataFilter 数据过滤中间件（根据角色过滤数据）
// 这个中间件会在查询时自动添加过滤条件
// - 普通用户：只能看到自己创建的数据
// - 管理员：可以看到所有数据
// - 子账号：只能看到自己分组的数据
func DataFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色和ID
		role, roleExists := c.Get("role")
		userID, userIDExists := c.Get("user_id")
		groupID, groupIDExists := c.Get("group_id")

		if !roleExists {
			c.Next()
			return
		}

		// 根据角色设置数据过滤条件
		switch role {
		case "admin":
			// 管理员可以看到所有数据，不需要过滤
			c.Set("data_filter", nil)
		case "user":
			// 普通用户只能看到自己创建的数据
			if userIDExists {
				c.Set("data_filter", map[string]interface{}{
					"user_id": userID,
				})
			}
		case "subaccount":
			// 子账号只能看到自己分组的数据
			if groupIDExists {
				c.Set("data_filter", map[string]interface{}{
					"group_id": groupID,
				})
			}
		}

		c.Next()
	}
}

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

// GetFilteredDB 获取带过滤条件的数据库查询
// 这是一个辅助函数，用于在Service层快速应用过滤
func GetFilteredDB(c *gin.Context, tableName string) *gorm.DB {
	db := database.GetDB()
	return ApplyDataFilter(c, db.Table(tableName), tableName)
}

// LogDataAccess 记录数据访问日志（可选）
func LogDataAccess(c *gin.Context, resource string, action string) {
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")
	groupID, _ := c.Get("group_id")

	logger.Infof("数据访问: role=%v, user_id=%v, group_id=%v, resource=%s, action=%s",
		role, userID, groupID, resource, action)
}

