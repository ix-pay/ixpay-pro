package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Config EventBus 配置
type Config struct {
	WorkerCount     int           // 工作协程数量
	BatchSize       int           // 批量处理大小
	FlushInterval   time.Duration // 刷新间隔
	MaxRetries      int           // 最大重试次数
	RetryInterval   time.Duration // 重试间隔
	DeadLetterAfter int           // 进入死信队列前的最大重试次数
	BufferEnabled   bool          // 是否启用缓冲
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		WorkerCount:     4,
		BatchSize:       100,
		FlushInterval:   time.Second * 5,
		MaxRetries:      3,
		RetryInterval:   time.Second * 30,
		DeadLetterAfter: 5,
		BufferEnabled:   true,
	}
}

// EventBus 事件总线
type EventBus struct {
	config    *Config
	db        *gorm.DB
	redis     *redis.Client
	repo      EventRepository
	buffer    *EventBuffer
	dlq       *DeadLetterQueue
	tracker   *EventTracker
	subMgr    *SubscriberManager
	publisher *EventPublisher
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	eventCh   chan *Event
	isRunning bool
	mu        sync.RWMutex
}

// NewEventBus 创建事件总线
func NewEventBus(db *gorm.DB, redisClient *redis.Client, config *Config) *EventBus {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	repo := NewEventRepository(db)
	dlq := NewDeadLetterQueue(redisClient, repo)
	buffer := NewEventBuffer(redisClient, config.BatchSize, config.FlushInterval)
	tracker := NewEventTracker(redisClient, repo, time.Hour*24*7)
	subMgr := NewSubscriberManager(repo)

	eb := &EventBus{
		config:  config,
		db:      db,
		redis:   redisClient,
		repo:    repo,
		buffer:  buffer,
		dlq:     dlq,
		tracker: tracker,
		subMgr:  subMgr,
		ctx:     ctx,
		cancel:  cancel,
		eventCh: make(chan *Event, config.BatchSize*10),
	}

	eb.publisher = NewEventPublisher(eb)

	// 设置缓冲区的刷新函数
	buffer.SetFlushFunc(eb.flushEventsToDB)

	return eb
}

// Start 启动事件总线
func (eb *EventBus) Start() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.isRunning {
		return fmt.Errorf("事件总线已在运行中")
	}

	// 从数据库加载订阅者
	if err := eb.subMgr.LoadFromDB(eb.ctx); err != nil {
		return fmt.Errorf("加载订阅者失败: %w", err)
	}

	// 启动工作协程
	for i := 0; i < eb.config.WorkerCount; i++ {
		eb.wg.Add(1)
		go eb.worker(i)
	}

	// 启动重试协程
	eb.wg.Add(1)
	go eb.retryWorker()

	// 启动缓冲区
	if eb.config.BufferEnabled {
		eb.buffer.Start(eb.ctx)
	}

	eb.isRunning = true
	return nil
}

// Stop 停止事件总线
func (eb *EventBus) Stop() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if !eb.isRunning {
		return nil
	}

	eb.cancel()
	eb.isRunning = false

	// 等待所有工作协程完成
	eb.wg.Wait()

	// 停止缓冲区
	if eb.config.BufferEnabled {
		if err := eb.buffer.Stop(eb.ctx); err != nil {
			return err
		}
	}

	return nil
}

// Publish 同步发布事件
func (eb *EventBus) Publish(ctx context.Context, event *Event) error {
	// 保存到数据库
	if err := eb.repo.CreateEvent(ctx, event); err != nil {
		return fmt.Errorf("保存事件失败: %w", err)
	}

	// 获取订阅者
	subscribers := eb.subMgr.GetSubscribers(event.Name)
	event.SubscriberCount = len(subscribers)

	// 创建投递记录
	for _, sub := range subscribers {
		delivery := &EventDelivery{
			EventID:      event.ID,
			SubscriberID: sub.ID,
			Status:       EventStatusPending,
			MaxAttempts:  sub.MaxRetries,
		}
		if err := eb.repo.CreateDelivery(ctx, delivery); err != nil {
			return fmt.Errorf("创建投递记录失败: %w", err)
		}
	}

	// 投递给订阅者
	for _, sub := range subscribers {
		if err := eb.deliverToSubscriber(ctx, event, sub); err != nil {
			// 记录错误，但继续投递给其他订阅者
			fmt.Printf("投递事件失败，订阅者: %s, 错误: %v\n", sub.Name, err)
		}
	}

	// 追踪事件
	if err := eb.tracker.TraceEvent(ctx, event, "发布"); err != nil {
		fmt.Printf("追踪事件失败: %v\n", err)
	}

	return nil
}

// PublishAsync 异步发布事件
func (eb *EventBus) PublishAsync(ctx context.Context, event *Event) error {
	// 检查是否有延迟
	if event.DelaySeconds > 0 || event.ScheduledAt != nil {
		return eb.publishDelayed(ctx, event)
	}

	// 如果启用缓冲，加入缓冲区
	if eb.config.BufferEnabled {
		return eb.buffer.Add(ctx, event)
	}

	// 否则直接发送到事件通道
	select {
	case eb.eventCh <- event:
		return nil
	default:
		return fmt.Errorf("事件通道已满")
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(ctx context.Context, sub *Subscriber) error {
	return eb.subMgr.AddSubscriber(ctx, sub)
}

// Unsubscribe 取消订阅
func (eb *EventBus) Unsubscribe(ctx context.Context, subscriberID int64) error {
	return eb.subMgr.RemoveSubscriber(ctx, subscriberID)
}

// GetEvent 获取事件
func (eb *EventBus) GetEvent(ctx context.Context, id int64) (*Event, error) {
	return eb.repo.GetEventByID(ctx, id)
}

// ListEvents 列出事件
func (eb *EventBus) ListEvents(ctx context.Context, filter EventFilter) ([]*Event, int64, error) {
	return eb.repo.ListEvents(ctx, filter)
}

// ListDeadLetters 列出死信
func (eb *EventBus) ListDeadLetters(ctx context.Context, filter EventFilter) ([]*Event, int64, error) {
	return eb.repo.ListDeadLetters(ctx, filter)
}

// RetryDeadLetter 重试死信
func (eb *EventBus) RetryDeadLetter(ctx context.Context, eventID int64) error {
	return eb.dlq.Retry(ctx, eventID)
}

// Publisher 获取发布器
func (eb *EventBus) Publisher() *EventPublisher {
	return eb.publisher
}

// Tracker 获取追踪器
func (eb *EventBus) Tracker() *EventTracker {
	return eb.tracker
}

// DLQ 获取死信队列
func (eb *EventBus) DLQ() *DeadLetterQueue {
	return eb.dlq
}

// worker 工作协程
func (eb *EventBus) worker(id int) {
	defer eb.wg.Done()

	for {
		select {
		case <-eb.ctx.Done():
			return
		case event := <-eb.eventCh:
			if err := eb.processEvent(eb.ctx, event); err != nil {
				fmt.Printf("工作协程 %d 处理事件失败: %v\n", id, err)
			}
		}
	}
}

// retryWorker 重试工作协程
func (eb *EventBus) retryWorker() {
	defer eb.wg.Done()

	ticker := time.NewTicker(eb.config.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-eb.ctx.Done():
			return
		case <-ticker.C:
			eb.processRetries()
		}
	}
}

// processEvent 处理事件
func (eb *EventBus) processEvent(ctx context.Context, event *Event) error {
	// 保存到数据库
	if err := eb.repo.CreateEvent(ctx, event); err != nil {
		return fmt.Errorf("保存事件失败: %w", err)
	}

	// 获取订阅者
	subscribers := eb.subMgr.GetSubscribers(event.Name)
	event.SubscriberCount = len(subscribers)

	// 创建投递记录并投递
	for _, sub := range subscribers {
		delivery := &EventDelivery{
			EventID:      event.ID,
			SubscriberID: sub.ID,
			Status:       EventStatusPending,
			MaxAttempts:  sub.MaxRetries,
		}
		if err := eb.repo.CreateDelivery(ctx, delivery); err != nil {
			return fmt.Errorf("创建投递记录失败: %w", err)
		}

		if err := eb.deliverToSubscriber(ctx, event, sub); err != nil {
			fmt.Printf("投递事件失败，订阅者: %s, 错误: %v\n", sub.Name, err)
		}
	}

	// 追踪事件
	if err := eb.tracker.TraceEvent(ctx, event, "异步发布"); err != nil {
		fmt.Printf("追踪事件失败: %v\n", err)
	}

	return nil
}

// deliverToSubscriber 投递给订阅者
func (eb *EventBus) deliverToSubscriber(ctx context.Context, event *Event, sub *Subscriber) error {
	if sub.Handler == nil {
		return fmt.Errorf("订阅者 %s 没有处理器", sub.Name)
	}

	// 设置超时
	timeoutCtx := ctx
	if sub.Timeout > 0 {
		var cancel context.CancelFunc
		timeoutCtx, cancel = context.WithTimeout(ctx, time.Duration(sub.Timeout)*time.Second)
		defer cancel()
	}

	// 执行处理器
	err := sub.Handler(event)

	// 更新投递记录
	if err != nil {
		// 失败处理
		return eb.handleDeliveryFailure(timeoutCtx, event, sub, err)
	}

	// 成功处理
	return eb.handleDeliverySuccess(timeoutCtx, event, sub)
}

// handleDeliverySuccess 处理投递成功
func (eb *EventBus) handleDeliverySuccess(ctx context.Context, event *Event, sub *Subscriber) error {
	// 更新投递记录
	deliveries, err := eb.repo.ListDeliveriesByEventID(ctx, event.ID)
	if err != nil {
		return err
	}

	for _, delivery := range deliveries {
		if delivery.SubscriberID == sub.ID {
			delivery.Status = EventStatusSuccess
			now := time.Now()
			delivery.DeliveredAt = &now
			return eb.repo.UpdateDelivery(ctx, delivery)
		}
	}

	return nil
}

// handleDeliveryFailure 处理投递失败
func (eb *EventBus) handleDeliveryFailure(ctx context.Context, event *Event, sub *Subscriber, err error) error {
	// 更新投递记录
	deliveries, err := eb.repo.ListDeliveriesByEventID(ctx, event.ID)
	if err != nil {
		return err
	}

	for _, delivery := range deliveries {
		if delivery.SubscriberID == sub.ID {
			delivery.Attempts++
			delivery.ErrorMessage = err.Error()

			if delivery.Attempts >= delivery.MaxAttempts {
				// 超过最大重试次数，标记为死信
				delivery.Status = EventStatusDeadLetter
				if dlqErr := eb.dlq.Add(ctx, event, fmt.Sprintf("订阅者 %s 投递失败: %v", sub.Name, err)); dlqErr != nil {
					return dlqErr
				}
			} else {
				// 设置下次重试时间
				delivery.Status = EventStatusFailed
				nextRetry := time.Now().Add(time.Duration(delivery.Attempts) * eb.config.RetryInterval)
				delivery.NextRetryAt = &nextRetry
			}

			return eb.repo.UpdateDelivery(ctx, delivery)
		}
	}

	return nil
}

// processRetries 处理重试
func (eb *EventBus) processRetries() {
	ctx := context.Background()

	// 获取需要重试的投递记录
	deliveries, err := eb.repo.ListFailedDeliveries(ctx, 100)
	if err != nil {
		fmt.Printf("获取失败投递记录失败: %v\n", err)
		return
	}

	for _, delivery := range deliveries {
		// 检查是否可以重试
		if delivery.NextRetryAt != nil && time.Now().Before(*delivery.NextRetryAt) {
			continue
		}

		// 获取事件
		event, err := eb.repo.GetEventByID(ctx, delivery.EventID)
		if err != nil {
			continue
		}

		// 获取订阅者
		sub, err := eb.repo.GetSubscriberByID(ctx, delivery.SubscriberID)
		if err != nil {
			continue
		}

		// 重新投递
		if err := eb.deliverToSubscriber(ctx, event, sub); err != nil {
			fmt.Printf("重试投递失败，事件ID: %d, 订阅者ID: %d, 错误: %v\n", event.ID, sub.ID, err)
		}
	}
}

// publishDelayed 发布延迟事件
func (eb *EventBus) publishDelayed(ctx context.Context, event *Event) error {
	// 保存到数据库
	if err := eb.repo.CreateEvent(ctx, event); err != nil {
		return fmt.Errorf("保存延迟事件失败: %w", err)
	}

	// 添加到 Redis 延迟队列
	scheduledAt := event.ScheduledAt
	if event.DelaySeconds > 0 && event.ScheduledAt == nil {
		t := time.Now().Add(time.Duration(event.DelaySeconds) * time.Second)
		scheduledAt = &t
	}

	if scheduledAt != nil {
		score := float64(scheduledAt.Unix())
		member, _ := json.Marshal(map[string]interface{}{
			"event_id": event.ID,
			"name":     event.Name,
		})

		if err := eb.redis.ZAdd(ctx, "eventbus:delayed", redis.Z{
			Score:  score,
			Member: string(member),
		}).Err(); err != nil {
			return fmt.Errorf("添加延迟事件到 Redis 失败: %w", err)
		}
	}

	return nil
}

// flushEventsToDB 批量刷新事件到数据库
func (eb *EventBus) flushEventsToDB(events []*Event) error {
	ctx := context.Background()

	for _, event := range events {
		if err := eb.processEvent(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

// GetSubscriberManager 获取订阅者管理器
func (eb *EventBus) GetSubscriberManager() *SubscriberManager {
	return eb.subMgr
}

// GetEventRepository 获取事件仓库
func (eb *EventBus) GetEventRepository() EventRepository {
	return eb.repo
}
