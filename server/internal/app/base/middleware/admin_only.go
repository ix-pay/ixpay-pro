package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminOnly 管理端接口仅管理员角色放行
// 用于管理平台专属功能（租户管理/微信渠道配置等），与租户管理 isAdmin 语义一致：
// base 模块 PermissionMiddleware 中 admin 角色本就直通，此处用于语义化拦截并给出明确提示
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可执行管理操作"})
			c.Abort()
			return
		}
		c.Next()
	}
}
