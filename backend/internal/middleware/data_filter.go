package middleware

import (
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
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

// 注意：ApplyDataFilter 和 GetFilteredDB 函数已移至 utils 包
// 请使用 utils.ApplyDataFilter 替代

// LogDataAccess 记录数据访问日志（可选）
func LogDataAccess(c *gin.Context, resource string, action string) {
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")
	groupID, _ := c.Get("group_id")

	logger.Infof("数据访问: role=%v, user_id=%v, group_id=%v, resource=%s, action=%s",
		role, userID, groupID, resource, action)
}

