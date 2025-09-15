package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// RedisClient 提供Redis客户端功能
type RedisClient struct {
	Client *redis.Client
	ctx    context.Context
}

// NewRedisClient 创建新的Redis客户端
func NewRedisClient(cfg *config.Config, log logger.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Address,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  time.Duration(cfg.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Redis.WriteTimeout) * time.Second,
		PoolSize:     cfg.Redis.PoolSize,
	})

	ctx := context.Background()

	// 测试连接
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Redis connected successfully")

	return &RedisClient{
			Client: client,
			ctx:    ctx,
		},
		nil
}

// Set 设置键值对
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(r.ctx, key, value, expiration).Err()
}

// Get 获取键对应的值
func (r *RedisClient) Get(key string) (string, error) {
	return r.Client.Get(r.ctx, key).Result()
}

// Del 删除键
func (r *RedisClient) Del(key string) error {
	return r.Client.Del(r.ctx, key).Err()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(key string) (bool, error) {
	count, err := r.Client.Exists(r.ctx, key).Result()
	return count > 0, err
}

// Expire 设置键的过期时间
func (r *RedisClient) Expire(key string, expiration time.Duration) error {
	return r.Client.Expire(r.ctx, key, expiration).Err()
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
