package middleware

import (
	"sync"
	"time"

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

	// 清理相关
	cleanupInterval time.Duration
	lastAccess      map[string]time.Time
	stopCh          chan struct{}
}

// NewRateLimiter 创建一个新的速率限制器
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		limiters:        make(map[string]*rate.Limiter),
		rate:            r,
		burst:           b,
		cleanupInterval: 10 * time.Minute,
		lastAccess:      make(map[string]time.Time),
		stopCh:          make(chan struct{}),
	}

	// 启动后台清理协程
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop 定期清理过期条目
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, last := range rl.lastAccess {
				// 清理超过 30 分钟未访问的条目
				if now.Sub(last) > 30*time.Minute {
					delete(rl.limiters, key)
					delete(rl.lastAccess, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCh:
			return
		}
	}
}

// Stop 停止后台清理协程
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
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
	rl.lastAccess[key] = time.Now()

	return limiter
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	limiter := NewRateLimiter(r, b)

	return func(c *gin.Context) {
		// 使用客户端IP作为限流键
		clientIP := c.ClientIP()

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