package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// EventBuffer 事件缓冲区，使用 Redis 进行临时存储
type EventBuffer struct {
	redisClient   *redis.Client
	bufferKey     string
	batchSize     int
	flushInterval time.Duration
	buffer        []*Event
	mu            sync.Mutex
	stopCh        chan struct{}
	flushFunc     func([]*Event) error
}

// NewEventBuffer 创建事件缓冲区
func NewEventBuffer(redisClient *redis.Client, batchSize int, flushInterval time.Duration) *EventBuffer {
	return &EventBuffer{
		redisClient:   redisClient,
		bufferKey:     "eventbus:buffer",
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]*Event, 0, batchSize),
		stopCh:        make(chan struct{}),
	}
}

// SetFlushFunc 设置批量写入函数
func (b *EventBuffer) SetFlushFunc(fn func([]*Event) error) {
	b.flushFunc = fn
}

// Add 添加事件到缓冲区
func (b *EventBuffer) Add(ctx context.Context, event *Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, event)

	// 将事件推入 Redis 列表（持久化备份）
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化事件失败: %w", err)
	}

	pipe := b.redisClient.Pipeline()
	pipe.RPush(ctx, b.bufferKey, eventJSON)
	pipe.Exec(ctx)

	// 如果达到批次大小，立即刷新
	if len(b.buffer) >= b.batchSize {
		return b.flushLocked(ctx)
	}

	return nil
}

// Start 启动异步刷新协程
func (b *EventBuffer) Start(ctx context.Context) {
	go b.flushLoop(ctx)
}

// Stop 停止缓冲区
func (b *EventBuffer) Stop(ctx context.Context) error {
	close(b.stopCh)
	// 最后一次刷新
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.buffer) > 0 {
		return b.flushLocked(ctx)
	}
	return nil
}

// flushLoop 定时刷新循环
func (b *EventBuffer) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			if len(b.buffer) > 0 {
				if err := b.flushLocked(ctx); err != nil {
					// 记录错误日志
					fmt.Printf("刷新事件缓冲区失败: %v\n", err)
				}
			}
			b.mu.Unlock()
		case <-ctx.Done():
			return
		case <-b.stopCh:
			return
		}
	}
}

// flushLocked 刷新缓冲区（必须在持有锁的情况下调用）
func (b *EventBuffer) flushLocked(ctx context.Context) error {
	if len(b.buffer) == 0 {
		return nil
	}

	if b.flushFunc != nil {
		if err := b.flushFunc(b.buffer); err != nil {
			return err
		}
	}

	// 刷新成功后清理 Redis 中的对应数据
	b.buffer = b.buffer[:0]

	return nil
}

// RecoverFromRedis 从 Redis 恢复未处理的事件
func (b *EventBuffer) RecoverFromRedis(ctx context.Context) ([]*Event, error) {
	// 获取所有缓冲的事件
	result, err := b.redisClient.LRange(ctx, b.bufferKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("从 Redis 恢复事件失败: %w", err)
	}

	events := make([]*Event, 0, len(result))
	for _, item := range result {
		var event Event
		if err := json.Unmarshal([]byte(item), &event); err != nil {
			continue // 跳过无法解析的事件
		}
		events = append(events, &event)
	}

	return events, nil
}

// Clear 清空缓冲区
func (b *EventBuffer) Clear(ctx context.Context) error {
	return b.redisClient.Del(ctx, b.bufferKey).Err()
}
