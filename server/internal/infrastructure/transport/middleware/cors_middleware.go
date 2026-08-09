package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/config"
)

// CORSMiddleware CORS中间件
func CORSMiddleware(cfg *config.CORSConfig) gin.HandlerFunc {
	// 设置默认值
	origins := cfg.AllowedOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	methods := cfg.AllowedMethods
	if len(methods) == 0 {
		methods = []string{"POST", "GET", "PUT", "DELETE", "OPTIONS", "PATCH"}
	}
	headers := cfg.AllowedHeaders
	if len(headers) == 0 {
		headers = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Requested-With"}
	}
	allowCredentials := cfg.AllowCredentials
	maxAge := cfg.MaxAge
	if maxAge <= 0 {
		maxAge = 86400
	}

	// 通配符模式时不能设置 Allow-Credentials=true
	if len(origins) == 1 && origins[0] == "*" {
		allowCredentials = false
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 设置允许的源
		if len(origins) == 1 && origins[0] == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			allowed := false
			for _, o := range origins {
				if o == origin {
					allowed = true
					break
				}
			}
			if allowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		// 设置允许的请求方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))

		// 设置允许的请求头
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))

		// 设置是否允许携带凭证
		if allowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求的缓存时间
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// 处理OPTIONS请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
