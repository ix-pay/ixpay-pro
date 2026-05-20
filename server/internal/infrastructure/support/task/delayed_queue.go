package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DelayedTask 延迟任务
type DelayedTask struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	ExecuteAt  time.Time              `json:"execute_at"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	CreatedAt  time.Time              `json:"created_at"`
}

// TaskHandler 任务处理函数
type TaskHandler func(ctx context.Context, task *DelayedTask) error

// Logger 日志接口
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// DelayedQueue 延迟任务队列（基于 Redis Streams + Sorted Set）
type DelayedQueue struct {
	redis        *redis.Client
	streamKey    string // Redis Stream 键
	groupName    string // 消费者组名称
	consumerID   string // 消费者 ID
	sortedSetKey string // Sorted Set 键（时间范围筛选）
	log          Logger
	handler      TaskHandler
	stopChan     chan struct{}
	mu           sync.Mutex
}

// NewDelayedQueue 创建延迟任务队列（Redis Streams 消费者组模式）
func NewDelayedQueue(redis *redis.Client, streamKey string, groupName string, consumerID string, log Logger) *DelayedQueue {
	return &DelayedQueue{
		redis:        redis,
		streamKey:    streamKey,
		groupName:    groupName,
		consumerID:   consumerID,
		sortedSetKey: streamKey + ":timeline",
		log:          log,
		stopChan:     make(chan struct{}),
	}
}

// SetHandler 设置任务处理器
func (dq *DelayedQueue) SetHandler(handler TaskHandler) {
	dq.handler = handler
}

// InitStream 初始化 Stream 和消费者组
func (dq *DelayedQueue) InitStream(ctx context.Context) error {
	// 尝试创建消费者组（如果不存在）
	err := dq.redis.XGroupCreateMkStream(ctx, dq.streamKey, dq.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("创建消费者组失败: %w", err)
	}
	dq.log.Info("延迟任务队列 Stream 已初始化",
		"stream_key", dq.streamKey,
		"group_name", dq.groupName)
	return nil
}

// Add 添加延迟任务
func (dq *DelayedQueue) Add(ctx context.Context, task *DelayedTask) error {
	dq.mu.Lock()
	defer dq.mu.Unlock()

	task.ID = generateTaskID()
	task.CreatedAt = time.Now()
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化任务失败: %w", err)
	}

	// 写入 Redis Stream
	if err := dq.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: dq.streamKey,
		Values: map[string]interface{}{
			"data":       string(data),
			"execute_at": task.ExecuteAt.UnixMilli(),
			"task_id":    task.ID,
			"task_type":  task.Type,
		},
	}).Err(); err != nil {
		return fmt.Errorf("添加延迟任务到 Stream 失败: %w", err)
	}

	// 同时写入 Sorted Set（用于时间范围筛选和快速到期查询）
	score := float64(task.ExecuteAt.UnixMilli())
	if err := dq.redis.ZAdd(ctx, dq.sortedSetKey, redis.Z{
		Score:  score,
		Member: task.ID,
	}).Err(); err != nil {
		return fmt.Errorf("添加任务到时间线失败: %w", err)
	}

	dq.log.Info("延迟任务已添加",
		"task_id", task.ID,
		"type", task.Type,
		"execute_at", task.ExecuteAt.Format(time.RFC3339))

	return nil
}

// Start 启动消费循环
func (dq *DelayedQueue) Start(ctx context.Context) {
	// 初始化 Stream
	if err := dq.InitStream(ctx); err != nil {
		dq.log.Error("初始化 Stream 失败", "error", err)
		return
	}

	// 恢复未处理的消息（XPENDING）
	go dq.recoverPending(ctx)

	// 启动消费循环
	go dq.consumeLoop(ctx)
	dq.log.Info("延迟任务队列已启动", "stream_key", dq.streamKey)
}

// Stop 停止消费
func (dq *DelayedQueue) Stop() {
	close(dq.stopChan)
	dq.log.Info("延迟任务队列已停止")
}

// consumeLoop 消费循环（使用 XREADGROUP）
func (dq *DelayedQueue) consumeLoop(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dq.processDueTasks(ctx)
		case <-ctx.Done():
			return
		case <-dq.stopChan:
			return
		}
	}
}

// processDueTasks 处理到期任务
func (dq *DelayedQueue) processDueTasks(ctx context.Context) {
	now := time.Now().UnixMilli()

	// 从 Sorted Set 获取到期任务的 ID
	taskIDs, err := dq.redis.ZRangeByScore(ctx, dq.sortedSetKey, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%d", now),
		Count: 50, // 每次最多处理 50 个
	}).Result()

	if err != nil {
		dq.log.Error("获取到期任务失败", "error", err)
		return
	}

	if len(taskIDs) == 0 {
		return
	}

	// 从 Stream 中读取这些任务
	for _, taskID := range taskIDs {
		// 使用 XREADGROUP 读取消息
		streams, err := dq.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    dq.groupName,
			Consumer: dq.consumerID,
			Streams:  []string{dq.streamKey, ">"},
			Count:    1,
			Block:    0,
		}).Result()

		if err != nil {
			if err != redis.Nil {
				dq.log.Error("读取 Stream 消息失败", "error", err)
			}
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				if msg.Values["task_id"] == taskID {
					dq.processMessage(ctx, msg)
				}
			}
		}
	}
}

// processMessage 处理单条消息
func (dq *DelayedQueue) processMessage(ctx context.Context, msg redis.XMessage) {
	dataStr, _ := msg.Values["data"].(string)
	if dataStr == "" {
		dq.log.Error("消息数据为空", "message_id", msg.ID)
		dq.redis.XAck(ctx, dq.streamKey, dq.groupName, msg.ID)
		return
	}

	var task DelayedTask
	if err := json.Unmarshal([]byte(dataStr), &task); err != nil {
		dq.log.Error("反序列化任务失败", "error", err, "message_id", msg.ID)
		dq.redis.XAck(ctx, dq.streamKey, dq.groupName, msg.ID)
		dq.redis.ZRem(ctx, dq.sortedSetKey, task.ID)
		return
	}

	// 执行任务
	dq.executeTask(ctx, &task)

	// 确认消息
	dq.redis.XAck(ctx, dq.streamKey, dq.groupName, msg.ID)
	// 从时间线中移除
	dq.redis.ZRem(ctx, dq.sortedSetKey, task.ID)
}

// recoverPending 恢复未处理的消息（XPENDING）
func (dq *DelayedQueue) recoverPending(ctx context.Context) {
	dq.log.Info("开始恢复未处理的任务消息")

	// 获取待处理的消息
	pending, err := dq.redis.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: dq.streamKey,
		Group:  dq.groupName,
		Start:  "0",
		End:    "+",
		Count:  100,
	}).Result()

	if err != nil {
		dq.log.Error("获取待处理消息失败", "error", err)
		return
	}

	for _, p := range pending {
		// 检查 idle 时间超过 5 分钟的消息
		if p.Idle > 5*time.Minute {
			// 将消息重新分配给当前消费者
			_, err := dq.redis.XClaim(ctx, &redis.XClaimArgs{
				Stream:   dq.streamKey,
				Group:    dq.groupName,
				Consumer: dq.consumerID,
				MinIdle:  5 * time.Minute,
				Messages: []string{p.ID},
			}).Result()

			if err != nil {
				dq.log.Error("恢复消息失败", "message_id", p.ID, "error", err)
				continue
			}

			dq.log.Info("已恢复未处理的消息", "message_id", p.ID)
		}
	}
}

// executeTask 执行任务
func (dq *DelayedQueue) executeTask(ctx context.Context, task *DelayedTask) {
	if dq.handler == nil {
		dq.log.Error("任务处理器未设置", "task_id", task.ID)
		return
	}

	dq.log.Info("执行延迟任务",
		"task_id", task.ID,
		"type", task.Type,
		"retry", task.RetryCount)

	err := dq.handler(ctx, task)

	if err != nil {
		dq.log.Error("延迟任务执行失败",
			"task_id", task.ID,
			"error", err,
			"retry", task.RetryCount,
			"max_retries", task.MaxRetries)

		if task.RetryCount < task.MaxRetries {
			task.RetryCount++
			delay := time.Duration(task.RetryCount) * 30 * time.Second
			task.ExecuteAt = time.Now().Add(delay)
			dq.Add(ctx, task)
		} else {
			dq.log.Error("延迟任务达到最大重试次数", "task_id", task.ID)
		}
	} else {
		dq.log.Info("延迟任务执行成功", "task_id", task.ID)
	}
}

// GetPendingCount 获取待执行任务数
func (dq *DelayedQueue) GetPendingCount(ctx context.Context) (int64, error) {
	return dq.redis.ZCard(ctx, dq.sortedSetKey).Result()
}

// GetTasksByTimeRange 按时间范围获取任务
func (dq *DelayedQueue) GetTasksByTimeRange(ctx context.Context, start, end time.Time) ([]string, error) {
	minScore := start.UnixMilli()
	maxScore := end.UnixMilli()

	return dq.redis.ZRangeByScore(ctx, dq.sortedSetKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", minScore),
		Max: fmt.Sprintf("%d", maxScore),
	}).Result()
}

// GetStreamLength 获取 Stream 长度
func (dq *DelayedQueue) GetStreamLength(ctx context.Context) (int64, error) {
	return dq.redis.XLen(ctx, dq.streamKey).Result()
}

func generateTaskID() string {
	return fmt.Sprintf("dt_%d", time.Now().UnixNano())
}
