package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限中间件
func PermissionMiddleware(permissionManager *auth.PermissionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前请求的路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查路径是否需要权限控制
		// 排除Swagger路径和公开API
		if strings.HasPrefix(path, "/swagger") ||
			strings.HasPrefix(path, "/api/v1/auth") ||
			strings.HasPrefix(path, "/api/v1/health") ||
			strings.HasPrefix(path, "/api/v1/pay/notify") {
			c.Next()
			return
		}

		// 从gin.Context中获取context.Context
		ctx := context.Background()

		// 将用户信息从gin.Context复制到context.Context
		if userID, exists := c.Get("userID"); exists {
			ctx = context.WithValue(ctx, "userID", userID)
		}
		if username, exists := c.Get("username"); exists {
			ctx = context.WithValue(ctx, "username", username)
		}
		if role, exists := c.Get("role"); exists {
			ctx = context.WithValue(ctx, "role", role)
		}
		if loginType, exists := c.Get("loginType"); exists {
			ctx = context.WithValue(ctx, "loginType", loginType)
		}
		if claims, exists := c.Get("claims"); exists {
			ctx = context.WithValue(ctx, "claims", claims)
		}

		// 检查权限
		if !permissionManager.CheckPermission(ctx, method, path) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}
