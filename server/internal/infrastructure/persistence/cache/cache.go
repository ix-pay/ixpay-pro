package cache

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/redis"
)

// Cache 定义缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(key string) (string, error)
	// Set 设置缓存值
	Set(key string, value interface{}, expiration time.Duration) error
	// Delete 删除缓存值
	Delete(key string) error
	// Exists 检查缓存键是否存在
	Exists(key string) (bool, error)
	// Close 关闭缓存连接
	Close() error
}

// SetupCache 设置缓存服务
func SetupCache(cfg *config.Config, redisClient *redis.RedisClient) (Cache, error) {
	return NewRedisCache(redisClient, redisClient.GetContext(), cfg.Redis.PreKey, 5*time.Minute), nil
}
