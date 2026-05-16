package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var globalRedisClient *redis.Client

func SetupGlobalRedisClient(client *redis.Client) {
	globalRedisClient = client
}

type StreamMaintenanceTask struct {
	TaskID    string `json:"-"`
	StreamKey string `json:"stream_key"`
	MaxLength int64  `json:"max_length"`
	TrimType  string `json:"trim_type"`
	Timeout   int    `json:"timeout"`
}

func CreateStreamMaintenanceTask(taskID string, params json.RawMessage) (Task, error) {
	var t StreamMaintenanceTask
	if err := json.Unmarshal(params, &t); err != nil {
		return nil, fmt.Errorf("解析 Stream 维护任务参数失败: %w", err)
	}

	t.TaskID = taskID
	if t.StreamKey == "" {
		return nil, fmt.Errorf("stream_key 不能为空")
	}
	if t.MaxLength <= 0 {
		t.MaxLength = 10000
	}
	if t.TrimType == "" {
		t.TrimType = "approx"
	}
	if t.Timeout == 0 {
		t.Timeout = 30
	}

	return &t, nil
}

func (t *StreamMaintenanceTask) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(t.Timeout)*time.Second)
	defer cancel()

	if globalRedisClient == nil {
		return fmt.Errorf("Redis 客户端未初始化")
	}

	switch t.TrimType {
	case "approx":
		return t.xTrimApprox(ctx)
	case "exact":
		return t.xTrimExact(ctx)
	default:
		return fmt.Errorf("不支持的裁剪类型: %s", t.TrimType)
	}
}

func (t *StreamMaintenanceTask) xTrimApprox(ctx context.Context) error {
	deleted, err := globalRedisClient.XTrimMaxLenApprox(ctx, t.StreamKey, t.MaxLength, 0).Result()
	if err != nil {
		return fmt.Errorf("XTRIM 裁剪失败 (approx): stream=%s, max_len=%d, error=%w", t.StreamKey, t.MaxLength, err)
	}

	fmt.Printf("Stream 维护完成: stream=%s, 裁剪类型=approx, 最大长度=%d, 删除消息数=%d\n",
		t.StreamKey, t.MaxLength, deleted)
	return nil
}

func (t *StreamMaintenanceTask) xTrimExact(ctx context.Context) error {
	deleted, err := globalRedisClient.XTrimMaxLen(ctx, t.StreamKey, t.MaxLength).Result()
	if err != nil {
		return fmt.Errorf("XTRIM 裁剪失败 (exact): stream=%s, max_len=%d, error=%w", t.StreamKey, t.MaxLength, err)
	}

	fmt.Printf("Stream 维护完成: stream=%s, 裁剪类型=exact, 最大长度=%d, 删除消息数=%d\n",
		t.StreamKey, t.MaxLength, deleted)
	return nil
}

func (t *StreamMaintenanceTask) GetName() string {
	return t.TaskID
}
