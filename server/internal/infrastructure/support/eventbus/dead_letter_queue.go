package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// DeadLetterQueue 死信队列
type DeadLetterQueue struct {
	redisClient *redis.Client
	repo        EventRepository
	streamKey   string
}

// NewDeadLetterQueue 创建死信队列
func NewDeadLetterQueue(redisClient *redis.Client, repo EventRepository) *DeadLetterQueue {
	return &DeadLetterQueue{
		redisClient: redisClient,
		repo:        repo,
		streamKey:   "eventbus:dead-letter",
	}
}

// Add 添加事件到死信队列
func (dlq *DeadLetterQueue) Add(ctx context.Context, event *Event, reason string) error {
	// 更新数据库中的事件状态
	if err := dlq.repo.MoveToDeadLetter(ctx, event.ID, reason); err != nil {
		return fmt.Errorf("移动事件到死信队列失败: %w", err)
	}

	// 同时写入 Redis Stream（用于快速查询）
	eventData := map[string]interface{}{
		"event_id":    event.ID,
		"event_name":  event.Name,
		"payload":     event.Payload,
		"reason":      reason,
		"retry_count": event.RetryCount,
		"max_retries": event.MaxRetries,
		"created_at":  time.Now().Unix(),
	}

	data, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("序列化死信事件失败: %w", err)
	}

	if err := dlq.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: dlq.streamKey,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Err(); err != nil {
		return fmt.Errorf("写入 Redis 死信队列失败: %w", err)
	}

	return nil
}

// List 列出死信队列中的事件
func (dlq *DeadLetterQueue) List(ctx context.Context, start string, count int64) ([]DeadLetterEntry, error) {
	result, err := dlq.redisClient.XRevRangeN(ctx, dlq.streamKey, "+", start, count).Result()
	if err != nil {
		return nil, fmt.Errorf("读取死信队列失败: %w", err)
	}

	entries := make([]DeadLetterEntry, 0, len(result))
	for _, msg := range result {
		var entry DeadLetterEntry
		entry.ID = msg.ID
		if data, ok := msg.Values["data"].(string); ok {
			if err := json.Unmarshal([]byte(data), &entry.Data); err != nil {
				continue
			}
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Remove 从死信队列移除
func (dlq *DeadLetterQueue) Remove(ctx context.Context, id string) error {
	return dlq.redisClient.XDel(ctx, dlq.streamKey, id).Err()
}

// Count 获取死信队列大小
func (dlq *DeadLetterQueue) Count(ctx context.Context) (int64, error) {
	return dlq.redisClient.XLen(ctx, dlq.streamKey).Result()
}

// Retry 重试死信队列中的事件
func (dlq *DeadLetterQueue) Retry(ctx context.Context, eventID int64) error {
	// 从数据库获取事件
	event, err := dlq.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("获取事件失败: %w", err)
	}

	// 重置事件状态
	event.Status = EventStatusPending
	event.RetryCount = 0
	event.ErrorMessage = ""
	event.NextRetryAt = nil

	if err := dlq.repo.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("更新事件状态失败: %w", err)
	}

	return nil
}

// DeadLetterEntry 死信队列条目
type DeadLetterEntry struct {
	ID   string
	Data map[string]interface{}
}
