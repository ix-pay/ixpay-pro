package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// EventTracker 事件追踪器
type EventTracker struct {
	redisClient *redis.Client
	repo        EventRepository
	traceKey    string
	ttl         time.Duration
}

// NewEventTracker 创建事件追踪器
func NewEventTracker(redisClient *redis.Client, repo EventRepository, ttl time.Duration) *EventTracker {
	return &EventTracker{
		redisClient: redisClient,
		repo:        repo,
		traceKey:    "eventbus:trace",
		ttl:         ttl,
	}
}

// TraceEvent 追踪事件状态变更
func (t *EventTracker) TraceEvent(ctx context.Context, event *Event, action string) error {
	trace := EventTrace{
		EventID:   event.ID,
		EventName: event.Name,
		Action:    action,
		Status:    event.Status.String(),
		Timestamp: time.Now().Unix(),
	}

	traceJSON, err := json.Marshal(trace)
	if err != nil {
		return fmt.Errorf("序列化事件追踪数据失败: %w", err)
	}

	// 写入 Redis Hash
	key := fmt.Sprintf("%s:%d", t.traceKey, event.ID)
	if err := t.redisClient.HSet(ctx, key, fmt.Sprintf("trace_%d", trace.Timestamp), string(traceJSON)).Err(); err != nil {
		return fmt.Errorf("写入事件追踪数据失败: %w", err)
	}

	// 设置过期时间
	t.redisClient.Expire(ctx, key, t.ttl)

	return nil
}

// GetEventTrace 获取事件追踪历史
func (t *EventTracker) GetEventTrace(ctx context.Context, eventID int64) ([]EventTrace, error) {
	key := fmt.Sprintf("%s:%d", t.traceKey, eventID)

	// 获取所有追踪记录
	all, err := t.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("获取事件追踪数据失败: %w", err)
	}

	traces := make([]EventTrace, 0, len(all))
	for _, value := range all {
		var trace EventTrace
		if err := json.Unmarshal([]byte(value), &trace); err != nil {
			continue
		}
		traces = append(traces, trace)
	}

	return traces, nil
}

// GetEventStatus 获取事件当前状态
func (t *EventTracker) GetEventStatus(ctx context.Context, eventID int64) (*Event, error) {
	return t.repo.GetEventByID(ctx, eventID)
}

// EventTrace 事件追踪记录
type EventTrace struct {
	EventID   int64  `json:"eventId"`
	EventName string `json:"eventName"`
	Action    string `json:"action"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}
