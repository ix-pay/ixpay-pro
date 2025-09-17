package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/robfig/cron/v3"
)

// Task 定义任务接口
type Task interface {
	Run(ctx context.Context) error
	GetName() string
}

// ScheduledTask 定时任务
type ScheduledTask struct {
	Task     Task
	CronExpr string
}

// TaskManager 任务管理器
type TaskManager struct {
	cron            *cron.Cron
	log             logger.Logger
	failedTasks     map[string]*TaskInfo
	failedTasksMux  sync.RWMutex
	runningTasks    map[string]cron.EntryID
	runningTasksMux sync.RWMutex
}

// TaskInfo 任务信息
type TaskInfo struct {
	Task        Task
	ExecuteAt   time.Time
	Attempts    int
	MaxAttempts int
}

// NewTaskManager 创建任务管理器
func NewTaskManager(log logger.Logger) *TaskManager {
	// 创建cron实例
	c := cron.New(cron.WithSeconds())

	return &TaskManager{
		cron:         c,
		log:          log,
		failedTasks:  make(map[string]*TaskInfo),
		runningTasks: make(map[string]cron.EntryID),
	}
}

// Start 启动任务管理器
func (tm *TaskManager) Start() {
	tm.cron.Start()
	tm.log.Info("Task manager started")
}

// Stop 停止任务管理器
func (tm *TaskManager) Stop() {
	_ = tm.cron.Stop()
	tm.log.Info("Task manager stopped")
}

// AddScheduledTask 添加定时任务
func (tm *TaskManager) AddScheduledTask(task *ScheduledTask) error {
	// 检查任务是否已存在
	tm.runningTasksMux.RLock()
	_, exists := tm.runningTasks[task.Task.GetName()]
	tm.runningTasksMux.RUnlock()

	if exists {
		err := fmt.Errorf("task %s already exists", task.Task.GetName())
		tm.log.Error("Failed to add scheduled task", "error", err)
		return err
	}

	// 创建任务函数
	taskFunc := func() {
		ctx := context.Background()
		tm.log.Info("Running scheduled task", "task", task.Task.GetName())
		if err := task.Task.Run(ctx); err != nil {
			tm.log.Error("Scheduled task failed", "task", task.Task.GetName(), "error", err)
			tm.addFailedTask(task.Task, 3) // 默认最多重试3次
		} else {
			tm.log.Info("Scheduled task completed successfully", "task", task.Task.GetName())
		}
	}

	// 添加任务到cron
	entryID, err := tm.cron.AddFunc(task.CronExpr, taskFunc)
	if err != nil {
		tm.log.Error("Failed to add task to cron", "error", err)
		return err
	}

	// 记录运行中的任务
	tm.runningTasksMux.Lock()
	tm.runningTasks[task.Task.GetName()] = entryID
	tm.runningTasksMux.Unlock()

	tm.log.Info("Scheduled task added", "task", task.Task.GetName(), "cron_expr", task.CronExpr)
	return nil
}

// RemoveScheduledTask 移除定时任务
func (tm *TaskManager) RemoveScheduledTask(taskName string) error {
	tm.runningTasksMux.RLock()
	entryID, exists := tm.runningTasks[taskName]
	tm.runningTasksMux.RUnlock()

	if !exists {
		err := fmt.Errorf("task %s not found", taskName)
		tm.log.Error("Failed to remove scheduled task", "error", err)
		return err
	}

	// 移除任务
	tm.cron.Remove(entryID)

	// 从运行中任务列表移除
	tm.runningTasksMux.Lock()
	delete(tm.runningTasks, taskName)
	tm.runningTasksMux.Unlock()

	tm.log.Info("Scheduled task removed", "task", taskName)
	return nil
}

// AddOneTimeTask 添加一次性任务
func (tm *TaskManager) AddOneTimeTask(task Task, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		ctx := context.Background()
		tm.log.Info("Running one-time task", "task", task.GetName())
		if err := task.Run(ctx); err != nil {
			tm.log.Error("One-time task failed", "task", task.GetName(), "error", err)
			tm.addFailedTask(task, 3) // 默认最多重试3次
		} else {
			tm.log.Info("One-time task completed successfully", "task", task.GetName())
		}
	}()
}

// addFailedTask 添加失败的任务
func (tm *TaskManager) addFailedTask(task Task, maxAttempts int) {
	tm.failedTasksMux.Lock()
	defer tm.failedTasksMux.Unlock()

	taskInfo, exists := tm.failedTasks[task.GetName()]
	if exists {
		taskInfo.Attempts++
	} else {
		taskInfo = &TaskInfo{
			Task:        task,
			ExecuteAt:   time.Now(),
			Attempts:    1,
			MaxAttempts: maxAttempts,
		}
	}

	tm.failedTasks[task.GetName()] = taskInfo
}

// GetFailedTasks 获取所有失败的任务
func (tm *TaskManager) GetFailedTasks() []*TaskInfo {
	tm.failedTasksMux.RLock()
	defer tm.failedTasksMux.RUnlock()

	tasks := make([]*TaskInfo, 0, len(tm.failedTasks))
	for _, task := range tm.failedTasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// RetryFailedTask 重试失败的任务
func (tm *TaskManager) RetryFailedTask(taskName string) bool {
	tm.failedTasksMux.RLock()
	taskInfo, exists := tm.failedTasks[taskName]
	tm.failedTasksMux.RUnlock()

	if !exists {
		tm.log.Error("Failed task not found", "task", taskName)
		return false
	}

	// 检查是否超过最大重试次数
	if taskInfo.Attempts >= taskInfo.MaxAttempts {
		tm.log.Error("Task has reached maximum retry attempts", "task", taskName)
		return false
	}

	// 执行任务
	go func() {
		ctx := context.Background()
		tm.log.Info("Retrying failed task", "task", taskName, "attempt", taskInfo.Attempts+1)
		if err := taskInfo.Task.Run(ctx); err != nil {
			tm.log.Error("Retry task failed", "task", taskName, "error", err)
			tm.addFailedTask(taskInfo.Task, taskInfo.MaxAttempts)
		} else {
			tm.log.Info("Retry task completed successfully", "task", taskName)
			// 从失败任务列表移除
			tm.failedTasksMux.Lock()
			delete(tm.failedTasks, taskName)
			tm.failedTasksMux.Unlock()
		}
	}()

	return true
}

// RunTaskNow 立即运行指定的任务
func (tm *TaskManager) RunTaskNow(taskName string) bool {
	tm.runningTasksMux.RLock()
	entryID, exists := tm.runningTasks[taskName]
	tm.runningTasksMux.RUnlock()

	if !exists {
		tm.log.Error("Task not found", "task", taskName)
		return false
	}

	// 立即执行任务
	tm.cron.Entry(entryID).Job.Run()
	return true
}
