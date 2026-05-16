package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// ChainErrorStrategy 任务链错误处理策略
type ChainErrorStrategy string

const (
	ChainStopOnError  ChainErrorStrategy = "stop"  // 失败后停止整个链
	ChainSkipOnError  ChainErrorStrategy = "skip"  // 跳过失败任务继续执行
	ChainRetryOnError ChainErrorStrategy = "retry" // 重试失败任务
)

// ChainTask 任务链中的任务
type ChainTask struct {
	ID        string
	Task      Task
	DependsOn []string          // 依赖的任务 ID
	OnSuccess func() error      // 成功回调
	OnFailure func(error) error // 失败回调
	Status    ChainTaskStatus
	ExecuteAt time.Time
	Duration  int64
	Error     error
}

// ChainTaskStatus 任务链任务状态
type ChainTaskStatus string

const (
	ChainTaskPending ChainTaskStatus = "pending"
	ChainTaskRunning ChainTaskStatus = "running"
	ChainTaskSuccess ChainTaskStatus = "success"
	ChainTaskFailed  ChainTaskStatus = "failed"
	ChainTaskSkipped ChainTaskStatus = "skipped"
)

// TaskChain 任务链
type TaskChain struct {
	name        string
	tasks       map[string]*ChainTask
	taskOrder   []string
	onError     ChainErrorStrategy
	maxRetries  int
	timeout     time.Duration
	description string
	log         logger.Logger
	status      ChainStatus
	startAt     time.Time
	endAt       time.Time
	mux         sync.RWMutex
}

// ChainStatus 任务链状态
type ChainStatus string

const (
	ChainPending ChainStatus = "pending"
	ChainRunning ChainStatus = "running"
	ChainSuccess ChainStatus = "success"
	ChainFailed  ChainStatus = "failed"
	ChainStopped ChainStatus = "stopped"
)

// TaskChainBuilder 任务链构建器
type TaskChainBuilder struct {
	chain *TaskChain
}

// NewTaskChain 创建任务链构建器
func NewTaskChain(name string, log logger.Logger) *TaskChainBuilder {
	return &TaskChainBuilder{
		chain: &TaskChain{
			name:       name,
			tasks:      make(map[string]*ChainTask),
			taskOrder:  make([]string, 0),
			onError:    ChainStopOnError,
			maxRetries: 3,
			timeout:    30 * time.Minute,
			status:     ChainPending,
			log:        log,
		},
	}
}

// OnError 设置错误处理策略
func (b *TaskChainBuilder) OnError(strategy ChainErrorStrategy) *TaskChainBuilder {
	b.chain.onError = strategy
	return b
}

// Timeout 设置任务链超时时间
func (b *TaskChainBuilder) Timeout(timeout time.Duration) *TaskChainBuilder {
	b.chain.timeout = timeout
	return b
}

// MaxRetries 设置最大重试次数
func (b *TaskChainBuilder) MaxRetries(count int) *TaskChainBuilder {
	b.chain.maxRetries = count
	return b
}

// Description 设置任务链描述
func (b *TaskChainBuilder) Description(desc string) *TaskChainBuilder {
	b.chain.description = desc
	return b
}

// AddTask 添加任务到任务链
func (b *TaskChainBuilder) AddTask(id string, task Task) *TaskChainBuilder {
	b.chain.tasks[id] = &ChainTask{
		ID:     id,
		Task:   task,
		Status: ChainTaskPending,
	}
	b.chain.taskOrder = append(b.chain.taskOrder, id)
	return b
}

// AddTaskWithDeps 添加带依赖的任务到任务链
func (b *TaskChainBuilder) AddTaskWithDeps(id string, task Task, dependsOn []string) *TaskChainBuilder {
	b.chain.tasks[id] = &ChainTask{
		ID:        id,
		Task:      task,
		DependsOn: dependsOn,
		Status:    ChainTaskPending,
	}
	b.chain.taskOrder = append(b.chain.taskOrder, id)
	return b
}

// AddTaskWithCallbacks 添加带回调的任务到任务链
func (b *TaskChainBuilder) AddTaskWithCallbacks(id string, task Task, onSuccess func() error, onFailure func(error) error) *TaskChainBuilder {
	b.chain.tasks[id] = &ChainTask{
		ID:        id,
		Task:      task,
		Status:    ChainTaskPending,
		OnSuccess: onSuccess,
		OnFailure: onFailure,
	}
	b.chain.taskOrder = append(b.chain.taskOrder, id)
	return b
}

// Build 构建任务链
func (b *TaskChainBuilder) Build() *TaskChain {
	if len(b.chain.taskOrder) == 0 {
		b.chain.log.Error("任务链不能为空")
		return nil
	}

	// 验证依赖关系
	if err := b.chain.validateDependencies(); err != nil {
		b.chain.log.Error("任务链依赖验证失败", "error", err)
		return nil
	}

	return b.chain
}

// validateDependencies 验证依赖关系
func (tc *TaskChain) validateDependencies() error {
	for id, task := range tc.tasks {
		for _, dep := range task.DependsOn {
			if _, exists := tc.tasks[dep]; !exists {
				return fmt.Errorf("任务 %s 依赖的任务 %s 不存在", id, dep)
			}
		}
	}

	// 检查循环依赖
	visited := make(map[string]bool)
	inProgress := make(map[string]bool)

	var hasCycle func(id string) bool
	hasCycle = func(id string) bool {
		if inProgress[id] {
			return true
		}
		if visited[id] {
			return false
		}

		visited[id] = true
		inProgress[id] = true

		task, exists := tc.tasks[id]
		if !exists {
			return false
		}

		for _, dep := range task.DependsOn {
			if hasCycle(dep) {
				return true
			}
		}

		inProgress[id] = false
		return false
	}

	for id := range tc.tasks {
		if hasCycle(id) {
			return fmt.Errorf("任务链存在循环依赖")
		}
	}

	return nil
}

// Execute 执行任务链
func (tc *TaskChain) Execute(ctx context.Context) error {
	tc.mux.Lock()
	tc.status = ChainRunning
	tc.startAt = time.Now()
	tc.mux.Unlock()

	tc.log.Info("开始执行任务链", "name", tc.name, "tasks", len(tc.taskOrder))

	ctx, cancel := context.WithTimeout(ctx, tc.timeout)
	defer cancel()

	for _, taskID := range tc.taskOrder {
		select {
		case <-ctx.Done():
			tc.mux.Lock()
			tc.status = ChainFailed
			tc.endAt = time.Now()
			tc.mux.Unlock()
			return fmt.Errorf("任务链执行超时")
		default:
		}

		if err := tc.executeTask(ctx, taskID); err != nil {
			switch tc.onError {
			case ChainStopOnError:
				tc.mux.Lock()
				tc.status = ChainFailed
				tc.endAt = time.Now()
				tc.mux.Unlock()
				tc.log.Error("任务链执行失败，停止执行", "name", tc.name, "failed_task", taskID)
				return err
			case ChainSkipOnError:
				tc.log.Warn("任务执行失败，跳过继续执行", "task_id", taskID, "error", err)
				continue
			case ChainRetryOnError:
				if retryErr := tc.retryTask(ctx, taskID); retryErr != nil {
					tc.mux.Lock()
					tc.status = ChainFailed
					tc.endAt = time.Now()
					tc.mux.Unlock()
					return retryErr
				}
			}
		}
	}

	tc.mux.Lock()
	tc.status = ChainSuccess
	tc.endAt = time.Now()
	tc.mux.Unlock()

	tc.log.Info("任务链执行完成", "name", tc.name, "duration", tc.endAt.Sub(tc.startAt))
	return nil
}

// executeTask 执行单个任务
func (tc *TaskChain) executeTask(ctx context.Context, taskID string) error {
	task := tc.tasks[taskID]

	// 检查依赖
	for _, dep := range task.DependsOn {
		depTask := tc.tasks[dep]
		if depTask.Status != ChainTaskSuccess {
			tc.log.Warn("任务依赖未成功，跳过执行", "task_id", taskID, "dependency", dep, "dep_status", depTask.Status)
			task.Status = ChainTaskSkipped
			return nil
		}
	}

	tc.mux.Lock()
	task.Status = ChainTaskRunning
	task.ExecuteAt = time.Now()
	tc.mux.Unlock()

	tc.log.Info("执行任务链任务", "task_id", taskID, "chain", tc.name)

	startTime := time.Now()
	err := task.Task.Run(ctx)
	duration := time.Since(startTime).Milliseconds()

	tc.mux.Lock()
	task.Duration = duration
	task.Error = err
	tc.mux.Unlock()

	if err != nil {
		tc.mux.Lock()
		task.Status = ChainTaskFailed
		tc.mux.Unlock()

		tc.log.Error("任务链任务执行失败", "task_id", taskID, "error", err, "duration", duration)

		if task.OnFailure != nil {
			if cbErr := task.OnFailure(err); cbErr != nil {
				tc.log.Error("任务失败回调执行失败", "task_id", taskID, "error", cbErr)
			}
		}
		return err
	}

	tc.mux.Lock()
	task.Status = ChainTaskSuccess
	tc.mux.Unlock()

	tc.log.Info("任务链任务执行成功", "task_id", taskID, "duration", duration)

	if task.OnSuccess != nil {
		if err := task.OnSuccess(); err != nil {
			tc.log.Error("任务成功回调执行失败", "task_id", taskID, "error", err)
		}
	}

	return nil
}

// retryTask 重试任务
func (tc *TaskChain) retryTask(ctx context.Context, taskID string) error {
	for attempt := 1; attempt <= tc.maxRetries; attempt++ {
		tc.log.Info("重试任务链任务", "task_id", taskID, "attempt", attempt, "max_retries", tc.maxRetries)

		if err := tc.executeTask(ctx, taskID); err == nil {
			return nil
		}

		// 等待后重试（指数退避）
		delay := time.Duration(attempt) * 5 * time.Second
		select {
		case <-ctx.Done():
			return fmt.Errorf("任务重试超时")
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("任务 %s 达到最大重试次数", taskID)
}

// GetStatus 获取任务链状态
func (tc *TaskChain) GetStatus() ChainStatus {
	tc.mux.RLock()
	defer tc.mux.RUnlock()
	return tc.status
}

// GetTaskStatus 获取任务状态
func (tc *TaskChain) GetTaskStatus(taskID string) (ChainTaskStatus, error) {
	tc.mux.RLock()
	defer tc.mux.RUnlock()

	task, exists := tc.tasks[taskID]
	if !exists {
		return "", fmt.Errorf("任务不存在: %s", taskID)
	}
	return task.Status, nil
}

// GetTasks 获取所有任务
func (tc *TaskChain) GetTasks() map[string]*ChainTask {
	tc.mux.RLock()
	defer tc.mux.RUnlock()

	result := make(map[string]*ChainTask)
	for k, v := range tc.tasks {
		result[k] = v
	}
	return result
}

// GetName 获取任务链名称
func (tc *TaskChain) GetName() string {
	return tc.name
}

// AddTaskChainToManager 将任务链添加到任务管理器
func (tm *TaskManager) AddTaskChain(chain *TaskChain) error {
	if chain == nil {
		return fmt.Errorf("任务链不能为空")
	}

	tm.log.Info("添加任务链", "name", chain.name, "tasks", len(chain.taskOrder))

	// 创建任务链执行任务
	chainTask := &chainExecutionTask{
		chain: chain,
		log:   tm.log,
	}

	scheduledTask := &ScheduledTask{
		Task:        chainTask,
		CronExpr:    "0 0 2 * * *", // 默认每天凌晨2点执行
		Group:       "chain",
		Concurrency: ConcurrencySkip,
		Timeout:     chain.timeout,
	}

	return tm.AddScheduledTask(scheduledTask)
}

// chainExecutionTask 任务链执行任务
type chainExecutionTask struct {
	chain *TaskChain
	log   logger.Logger
}

func (t *chainExecutionTask) Run(ctx context.Context) error {
	return t.chain.Execute(ctx)
}

func (t *chainExecutionTask) GetName() string {
	return fmt.Sprintf("chain:%s", t.chain.name)
}
