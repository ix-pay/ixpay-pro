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
	Client     *redis.Client
	ctx        context.Context
	PreKey     string
	Expiration time.Duration
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
			Client:     client,
			ctx:        ctx,
			PreKey:     cfg.Redis.PreKey,
			Expiration: time.Second * 180,
		},
		nil
}

// Set 设置键值对
func (r *RedisClient) SetTry(key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(r.ctx, r.PreKey+key, value, expiration).Err()
}

// Get 获取键对应的值
func (r *RedisClient) GetTry(key string) (string, error) {
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

func (r *RedisClient) Set(id string, value string) error {
	err := r.Client.Set(r.ctx, r.PreKey+id, value, r.Expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) Get(key string, clear bool) string {
	val, err := r.Client.Get(r.ctx, r.PreKey+key).Result()
	if err != nil {
		return ""
	}
	if clear {
		err := r.Client.Del(r.ctx, r.PreKey+key).Err()
		if err != nil {
			return ""
		}
	}
	return val
}

func (r *RedisClient) Verify(id, answer string, clear bool) bool {
	key := r.PreKey + id
	v := r.Get(key, clear)
	return v == answer
}

// Expire 设置键的过期时间
func (r *RedisClient) Expire(key string, expiration time.Duration) error {
	return r.Client.Expire(r.ctx, key, expiration).Err()
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
