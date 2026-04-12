package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	http "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/http"
)

// RateLimiter 速率限制器结构体
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter 创建一个新的速率限制器
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter 根据客户端IP获取或创建一个速率限制器
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	limiter := NewRateLimiter(r, b)

	return func(c *gin.Context) {
		// 使用客户端IP作为限流键
		clientIP := c.ClientIP()
		// 也可以根据用户ID进行限流
		// userID, exists := c.Get("userID")

		// 获取速率限制器
		rateLimiter := limiter.getLimiter(clientIP)

		// 检查是否允许请求
		if !rateLimiter.Allow() {
			// 返回429 Too Many Requests
			http.TooManyRequestsResponse(c, "请求频率过高，请稍后再试")
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}
