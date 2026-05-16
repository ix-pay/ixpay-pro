package middleware

import (
	"net/http"
	"strings"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtAuth *auth.JWTAuth, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization 头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization 头是必需的"})
			c.Abort()
			return
		}

		// 检查令牌格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization 头格式必须为 Bearer {token}"})
			c.Abort()
			return
		}

		// 解析和验证令牌
		claims, err := jwtAuth.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效或已过期"})
			c.Abort()
			return
		}

		// 将用户信息添加到上下文
		c.Set("userID", claims.UserID)
		c.Set("userName", claims.Username)
		c.Set("role", claims.Role)
		c.Set("loginType", claims.LoginType)
		c.Set("claims", claims)

		c.Next()
	}
}
