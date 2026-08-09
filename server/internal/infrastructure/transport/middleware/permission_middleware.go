// Package middleware 提供通用权限中间件
// 注意：此中间件为通用中间件，当前未被路由直接使用。
// 实际使用的权限中间件在 app/base/middleware/permission_middleware.go
// 如需启用，需在路由中注册并配置合适的依赖注入
package middleware

import (
	"net/http"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	httpresponse "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/http"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限管理中间件
func PermissionMiddleware(permissionManager *auth.PermissionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查用户是否有权限访问
		if !permissionManager.CheckPermission(c.Request.Context(), method, path) {
			// 检查是否需要按钮权限
			if !permissionManager.CheckAPIPermissionWithButton(c.Request.Context(), method, path) {
				// 无权限，返回403错误
				c.AbortWithStatusJSON(http.StatusForbidden, httpresponse.Response{
					Code:    http.StatusForbidden,
					Message: "Permission denied",
					Data:    nil,
					Error:   "You don't have permission to access this resource",
				})
				return
			}
		}

		c.Next()
	}
}
