package task

import (
	"context"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// TaskHelper 任务帮助方法集
type TaskHelper struct {
	manager      *TaskManager
	factory      *TaskFactory
	delayedQueue *DelayedQueue
	log          logger.Logger
}

// NewTaskHelper 创建任务帮助方法集
func NewTaskHelper(manager *TaskManager, factory *TaskFactory, queue *DelayedQueue, log logger.Logger) *TaskHelper {
	return &TaskHelper{
		manager:      manager,
		factory:      factory,
		delayedQueue: queue,
		log:          log,
	}
}

// ==================== 定时任务快捷方法 ====================

// CreateCronTask 快速创建定时任务
func (h *TaskHelper) CreateCronTask(
	taskID string,
	taskType TaskType,
	cronExpr string,
	params map[string]interface{},
	opts ...TaskOption,
) error {
	options := defaultTaskOptions()
	for _, opt := range opts {
		opt(options)
	}

	scheduledTask, err := NewTaskBuilder(taskID).
		Type(taskType).
		CronExpr(cronExpr).
		Group(options.group).
		Timeout(options.timeout).
		Concurrency(options.concurrency).
		MaxRetries(options.maxRetries).
		Params(params).
		Build()

	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	if err := h.manager.AddScheduledTask(scheduledTask); err != nil {
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	h.log.Info("定时任务创建成功", "task_id", taskID, "cron_expr", cronExpr)
	return nil
}

// CreateDailyTask 创建每天执行的任务
func (h *TaskHelper) CreateDailyTask(
	taskID string,
	taskType TaskType,
	hour, minute int,
	params map[string]interface{},
	opts ...TaskOption,
) error {
	cronExpr := fmt.Sprintf("0 %d %d * * *", minute, hour)
	return h.CreateCronTask(taskID, taskType, cronExpr, params, opts...)
}

// CreateHourlyTask 创建每小时执行的任务
func (h *TaskHelper) CreateHourlyTask(
	taskID string,
	taskType TaskType,
	params map[string]interface{},
	opts ...TaskOption,
) error {
	return h.CreateCronTask(taskID, taskType, "0 0 * * * *", params, opts...)
}

// CreateEveryMinuteTask 创建每分钟执行的任务
func (h *TaskHelper) CreateEveryMinuteTask(
	taskID string,
	taskType TaskType,
	params map[string]interface{},
	opts ...TaskOption,
) error {
	return h.CreateCronTask(taskID, taskType, "0 * * * * *", params, opts...)
}

// CreateWeeklyTask 创建每周执行的任务
func (h *TaskHelper) CreateWeeklyTask(
	taskID string,
	taskType TaskType,
	dayOfWeek, hour, minute int,
	params map[string]interface{},
	opts ...TaskOption,
) error {
	cronExpr := fmt.Sprintf("0 %d %d * * %d", minute, hour, dayOfWeek)
	return h.CreateCronTask(taskID, taskType, cronExpr, params, opts...)
}

// CreateMonthlyTask 创建每月执行的任务
func (h *TaskHelper) CreateMonthlyTask(
	taskID string,
	taskType TaskType,
	day, hour, minute int,
	params map[string]interface{},
	opts ...TaskOption,
) error {
	cronExpr := fmt.Sprintf("0 %d %d %d * *", minute, hour, day)
	return h.CreateCronTask(taskID, taskType, cronExpr, params, opts...)
}

// ==================== 延迟任务快捷方法 ====================

// AddDelayedTask 快速添加延迟任务
func (h *TaskHelper) AddDelayedTask(
	taskType string,
	payload map[string]interface{},
	delay time.Duration,
	maxRetries int,
) error {
	if h.delayedQueue == nil {
		return fmt.Errorf("延迟任务队列未初始化")
	}

	delayedTask := &DelayedTask{
		Type:       taskType,
		Payload:    payload,
		ExecuteAt:  time.Now().Add(delay),
		MaxRetries: maxRetries,
	}

	ctx := context.Background()
	if err := h.delayedQueue.Add(ctx, delayedTask); err != nil {
		return fmt.Errorf("添加延迟任务失败: %w", err)
	}

	h.log.Info("延迟任务添加成功", "type", taskType, "delay", delay)
	return nil
}

// AddPaymentTimeoutTask 添加支付超时任务
func (h *TaskHelper) AddPaymentTimeoutTask(orderID string, timeout time.Duration) error {
	return h.AddDelayedTask("payment_timeout", map[string]interface{}{
		"order_id": orderID,
	}, timeout, 3)
}

// AddOrderCancelTask 添加订单取消任务
func (h *TaskHelper) AddOrderCancelTask(orderID string, cancelAt time.Time) error {
	delay := time.Until(cancelAt)
	if delay < 0 {
		delay = 0
	}
	return h.AddDelayedTask("order_cancel", map[string]interface{}{
		"order_id": orderID,
	}, delay, 3)
}

// AddDelayedTaskAt 在指定时间执行延迟任务
func (h *TaskHelper) AddDelayedTaskAt(
	taskType string,
	payload map[string]interface{},
	executeAt time.Time,
	maxRetries int,
) error {
	if h.delayedQueue == nil {
		return fmt.Errorf("延迟任务队列未初始化")
	}

	delayedTask := &DelayedTask{
		Type:       taskType,
		Payload:    payload,
		ExecuteAt:  executeAt,
		MaxRetries: maxRetries,
	}

	ctx := context.Background()
	if err := h.delayedQueue.Add(ctx, delayedTask); err != nil {
		return fmt.Errorf("添加延迟任务失败: %w", err)
	}

	h.log.Info("延迟任务添加成功", "type", taskType, "execute_at", executeAt)
	return nil
}

// ==================== 任务状态查询方法 ====================

// TaskStatus 任务状态
type TaskStatus struct {
	TaskID        string
	IsRunning     bool
	Group         string
	CronExpr      string
	Attempts      int
	MaxAttempts   int
	LastExecuteAt time.Time
}

// GetTaskStatus 获取任务状态
func (h *TaskHelper) GetTaskStatus(taskID string) (*TaskStatus, error) {
	taskInfo, exists := h.manager.GetTask(taskID)
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}

	return &TaskStatus{
		TaskID:      taskID,
		IsRunning:   h.manager.IsTaskRunning(taskID),
		Group:       taskInfo.Group,
		CronExpr:    taskInfo.CronExpr,
		Attempts:    taskInfo.Attempts,
		MaxAttempts: taskInfo.MaxAttempts,
	}, nil
}

// IsTaskRunning 检查任务是否正在运行
func (h *TaskHelper) IsTaskRunning(taskID string) bool {
	return h.manager.IsTaskRunning(taskID)
}

// GetTaskNextExecuteTime 获取任务下次执行时间
func (h *TaskHelper) GetTaskNextExecuteTime(taskID string) (time.Time, error) {
	taskInfo, exists := h.manager.GetTask(taskID)
	if !exists {
		return time.Time{}, fmt.Errorf("任务不存在: %s", taskID)
	}
	return taskInfo.ExecuteAt, nil
}

// GetTaskLastExecuteTime 获取任务上次执行时间
func (h *TaskHelper) GetTaskLastExecuteTime(taskID string) (time.Time, error) {
	taskInfo, exists := h.manager.GetTask(taskID)
	if !exists {
		return time.Time{}, fmt.Errorf("任务不存在: %s", taskID)
	}
	return taskInfo.ExecuteAt, nil
}

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	TaskID        string
	TotalExecutes int64
	SuccessCount  int64
	FailedCount   int64
	SuccessRate   float64
	AvgDuration   float64
}

// GetTaskStatistics 获取任务统计信息
func (h *TaskHelper) GetTaskStatistics(taskID string) (*TaskStatistics, error) {
	// 从 TaskManager 获取基本统计
	taskInfo, exists := h.manager.GetTask(taskID)
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}

	return &TaskStatistics{
		TaskID:        taskID,
		TotalExecutes: int64(taskInfo.Attempts),
		SuccessCount:  int64(taskInfo.Attempts),
		FailedCount:   0,
		SuccessRate:   100.0,
		AvgDuration:   0,
	}, nil
}

// ==================== 批量操作方法 ====================

// EnableAllTasks 启用所有任务
func (h *TaskHelper) EnableAllTasks() error {
	tasks := h.manager.GetAllTasks()
	for _, taskInfo := range tasks {
		if err := h.manager.EnableTask(taskInfo.Task.GetName()); err != nil {
			h.log.Warn("启用任务失败", "task_id", taskInfo.Task.GetName(), "error", err)
		}
	}
	h.log.Info("所有任务启用完成", "count", len(tasks))
	return nil
}

// DisableAllTasks 禁用所有任务
func (h *TaskHelper) DisableAllTasks() error {
	tasks := h.manager.GetAllTasks()
	for _, taskInfo := range tasks {
		if err := h.manager.DisableTask(taskInfo.Task.GetName()); err != nil {
			h.log.Warn("禁用任务失败", "task_id", taskInfo.Task.GetName(), "error", err)
		}
	}
	h.log.Info("所有任务禁用完成", "count", len(tasks))
	return nil
}

// EnableTasksByGroup 按分组启用任务
func (h *TaskHelper) EnableTasksByGroup(group string) error {
	tasks := h.manager.GetTasksByGroup(group)
	for _, taskInfo := range tasks {
		if err := h.manager.EnableTask(taskInfo.Task.GetName()); err != nil {
			h.log.Warn("启用任务失败", "task_id", taskInfo.Task.GetName(), "error", err)
		}
	}
	h.log.Info("分组任务启用完成", "group", group, "count", len(tasks))
	return nil
}

// DisableTasksByGroup 按分组禁用任务
func (h *TaskHelper) DisableTasksByGroup(group string) error {
	tasks := h.manager.GetTasksByGroup(group)
	for _, taskInfo := range tasks {
		if err := h.manager.DisableTask(taskInfo.Task.GetName()); err != nil {
			h.log.Warn("禁用任务失败", "task_id", taskInfo.Task.GetName(), "error", err)
		}
	}
	h.log.Info("分组任务禁用完成", "group", group, "count", len(tasks))
	return nil
}

// GetTasksByGroup 按分组获取任务
func (h *TaskHelper) GetTasksByGroup(group string) []*TaskInfo {
	return h.manager.GetTasksByGroup(group)
}

// GetAllTasks 获取所有任务
func (h *TaskHelper) GetAllTasks() []*TaskInfo {
	return h.manager.GetAllTasks()
}

// ==================== 任务选项 ====================

// TaskOption 任务选项
type TaskOption func(*taskOptions)

type taskOptions struct {
	group       string
	timeout     time.Duration
	concurrency ConcurrencyMode
	maxRetries  int
	description string
}

func defaultTaskOptions() *taskOptions {
	return &taskOptions{
		group:       "default",
		timeout:     5 * time.Minute,
		concurrency: ConcurrencySkip,
		maxRetries:  3,
	}
}

// WithGroup 设置任务分组
func WithGroup(group string) TaskOption {
	return func(o *taskOptions) {
		o.group = group
	}
}

// WithTimeout 设置任务超时
func WithTimeout(timeout time.Duration) TaskOption {
	return func(o *taskOptions) {
		o.timeout = timeout
	}
}

// WithConcurrency 设置并发模式
func WithConcurrency(mode ConcurrencyMode) TaskOption {
	return func(o *taskOptions) {
		o.concurrency = mode
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(count int) TaskOption {
	return func(o *taskOptions) {
		o.maxRetries = count
	}
}

// WithDescription 设置任务描述
func WithDescription(desc string) TaskOption {
	return func(o *taskOptions) {
		o.description = desc
	}
}
