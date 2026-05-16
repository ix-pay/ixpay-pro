package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
)

// CacheMiddleware 缓存中间件
func CacheMiddleware(cache cache.Cache, expiration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只缓存GET请求
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := generateCacheKey(c)

		// 尝试从缓存获取数据
		cachedData, err := cache.Get(cacheKey)
		if err == nil && cachedData != "" {
			// 缓存命中，直接返回缓存数据
			c.Data(http.StatusOK, "application/json", []byte(cachedData))
			c.Abort()
			return
		}

		// 创建响应写入器包装器，用于捕获响应体
		w := &responseWriterWrapper{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 检查响应状态码，只缓存200 OK的响应
		if c.Writer.Status() == http.StatusOK {
			// 将响应体写入缓存
			_ = cache.Set(cacheKey, w.body.String(), expiration)
		}
	}
}

// responseWriterWrapper 包装响应写入器，用于捕获响应体
type responseWriterWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法，将响应体同时写入缓冲区
func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// generateCacheKey 生成缓存键
func generateCacheKey(c *gin.Context) string {
	// 结合请求方法、路径和查询参数生成唯一键
	key := c.Request.Method + ":" + c.Request.URL.String()

	// 使用SHA256哈希生成固定长度的键
	hash := sha256.New()
	hash.Write([]byte(key))
	return hex.EncodeToString(hash.Sum(nil))
}

// CacheKeyContext 缓存键上下文，用于在请求处理过程中生成和使用缓存键
type CacheKeyContext struct {
	Key      string
	Expire   time.Duration
	Cache    cache.Cache
	skip     bool
	forceSet bool
}

// NewCacheKeyContext 创建新的缓存键上下文
func NewCacheKeyContext(cache cache.Cache, key string, expire time.Duration) *CacheKeyContext {
	return &CacheKeyContext{
		Key:    key,
		Expire: expire,
		Cache:  cache,
	}
}

// Skip 跳过缓存
func (c *CacheKeyContext) Skip() {
	c.skip = true
}

// ForceSet 强制设置缓存，即使已经存在
func (c *CacheKeyContext) ForceSet() {
	c.forceSet = true
}

// CacheControlMiddleware 缓存控制中间件，用于在请求处理过程中控制缓存行为
func CacheControlMiddleware(cache cache.Cache, defaultExpiration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只处理GET请求
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := generateCacheKey(c)

		// 创建缓存上下文
		cacheCtx := NewCacheKeyContext(cache, cacheKey, defaultExpiration)
		c.Set("cacheCtx", cacheCtx)

		// 尝试从缓存获取数据
		cachedData, err := cache.Get(cacheKey)
		if err == nil && cachedData != "" && !cacheCtx.skip {
			// 缓存命中，直接返回缓存数据
			c.Data(http.StatusOK, "application/json", []byte(cachedData))
			c.Abort()
			return
		}

		// 创建响应写入器包装器，用于捕获响应体
		w := &responseWriterWrapper{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 检查响应状态码，只缓存200 OK的响应
		if c.Writer.Status() == http.StatusOK && (!cacheCtx.skip || cacheCtx.forceSet) {
			// 将响应体写入缓存
			_ = cache.Set(cacheKey, w.body.String(), cacheCtx.Expire)
		}
	}
}

// GetCacheKeyContext 从Gin上下文获取缓存上下文
func GetCacheKeyContext(c *gin.Context) (*CacheKeyContext, bool) {
	cacheCtx, exists := c.Get("cacheCtx")
	if !exists {
		return nil, false
	}
	ctx, ok := cacheCtx.(*CacheKeyContext)
	return ctx, ok
}
