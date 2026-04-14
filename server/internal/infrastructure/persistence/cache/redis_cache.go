package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/redis"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	redisClient *redis.RedisClient
	ctx         context.Context
	prefix      string
	expiration  time.Duration
}

// NewRedisCache 创建新的Redis缓存实例
func NewRedisCache(redisClient *redis.RedisClient, ctx context.Context, prefix string, expiration time.Duration) *RedisCache {
	return &RedisCache{
		redisClient: redisClient,
		ctx:         ctx,
		prefix:      prefix,
		expiration:  expiration,
	}
}

// Get 获取缓存值
func (rc *RedisCache) Get(key string) (string, error) {
	return rc.redisClient.GetTry(rc.prefix + key)
}

// Set 设置缓存值
func (rc *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	// 如果未指定过期时间，使用默认过期时间
	if expiration == 0 {
		expiration = rc.expiration
	}
	// 将interface{}转换为string
	var valueStr string
	switch v := value.(type) {
	case string:
		valueStr = v
	case []byte:
		valueStr = string(v)
	default:
		valueStr = fmt.Sprintf("%v", v)
	}
	return rc.redisClient.SetTry(rc.prefix+key, valueStr, expiration)
}

// Delete 删除缓存值
func (rc *RedisCache) Delete(key string) error {
	return rc.redisClient.Del(rc.prefix + key)
}

// Exists 检查缓存是否存在
func (rc *RedisCache) Exists(key string) (bool, error) {
	return rc.redisClient.Exists(rc.prefix + key)
}

// Close 关闭缓存连接
func (rc *RedisCache) Close() error {
	return rc.redisClient.Close()
}

// CaptchaStoreAdapter 适配 RedisCache 以兼容 base64Captcha.Store 接口
type CaptchaStoreAdapter struct {
	cache *RedisCache
}

// NewCaptchaStoreAdapter 创建验证码存储适配器
func NewCaptchaStoreAdapter(cache *RedisCache) *CaptchaStoreAdapter {
	return &CaptchaStoreAdapter{cache: cache}
}

// Set 实现 base64Captcha.Store 接口的 Set 方法
func (a *CaptchaStoreAdapter) Set(id string, value string) error {
	return a.cache.redisClient.Client.Set(a.cache.ctx, a.cache.prefix+id, value, a.cache.expiration).Err()
}

// Get 实现 base64Captcha.Store 接口的 Get 方法
func (a *CaptchaStoreAdapter) Get(id string, clear bool) string {
	fullKey := a.cache.prefix + id
	val, err := a.cache.redisClient.Client.Get(a.cache.ctx, fullKey).Result()
	if err != nil {
		return ""
	}
	if clear {
		a.cache.redisClient.Client.Del(a.cache.ctx, fullKey)
	}
	return val
}

// Verify 实现 base64Captcha.Store 接口的 Verify 方法
func (a *CaptchaStoreAdapter) Verify(id, answer string, clear bool) bool {
	v := a.Get(id, clear)
	return v == answer
}
