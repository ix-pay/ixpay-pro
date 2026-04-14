package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	model "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"

	"github.com/robfig/cron/v3"
)

// Task 定义任务接口
type Task interface {
	Run(ctx context.Context) error
	GetName() string
}

// TaskWithGroup 带分组的任务接口
type TaskWithGroup interface {
	Task
	GetGroup() string
}

// ScheduledTask 定时任务
type ScheduledTask struct {
	Task     Task
	CronExpr string
	Group    string // 任务分组
}

// TaskManager 任务管理器
type TaskManager struct {
	cron                *cron.Cron
	log                 logger.Logger
	failedTasks         map[string]*TaskInfo
	failedTasksMux      sync.RWMutex
	runningTasks        map[string]cron.EntryID
	runningTasksMux     sync.RWMutex
	taskInfoMap         map[string]*TaskInfo
	taskInfoMux         sync.RWMutex
	taskGroupMap        map[string]string // 任务 ID 到分组的映射
	taskGroupMux        sync.RWMutex
	executionLogRepo    model.TaskExecutionLogRepository // 任务执行日志仓库
	executionLogRepoMux sync.RWMutex
}

// TaskInfo 任务信息
type TaskInfo struct {
	Task        Task
	ExecuteAt   time.Time
	Attempts    int
	MaxAttempts int
	CronExpr    string // 用于存储定时任务的表达式
	Group       string // 任务分组
}

// NewTaskManager 创建任务管理器
func SetupTaskManager(log logger.Logger) *TaskManager {
	// 创建 cron 实例
	c := cron.New(cron.WithSeconds())

	// 尝试使用任务日志记录器
	taskLogger := logger.GetGlobalLogger(logger.TaskLogger)
	if taskLogger == nil {
		// 如果全局日志管理器未初始化，使用传入的默认日志
		taskLogger = log
	}

	return &TaskManager{
		cron:             c,
		log:              taskLogger,
		failedTasks:      make(map[string]*TaskInfo),
		runningTasks:     make(map[string]cron.EntryID),
		taskInfoMap:      make(map[string]*TaskInfo),
		taskGroupMap:     make(map[string]string),
		executionLogRepo: nil,
	}
}

// SetExecutionLogRepository 设置任务执行日志仓库
func (tm *TaskManager) SetExecutionLogRepository(repo model.TaskExecutionLogRepository) {
	tm.executionLogRepoMux.Lock()
	defer tm.executionLogRepoMux.Unlock()
	tm.executionLogRepo = repo
	tm.log.Info("Task execution log repository set")
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
		err := fmt.Errorf("任务 %s 已存在", task.Task.GetName())
		tm.log.Error("添加定时任务失败", "error", err)
		return err
	}

	// 获取任务分组
	group := task.Group
	if group == "" {
		// 尝试从 TaskWithGroup 接口获取分组
		if taskWithGroup, ok := task.Task.(TaskWithGroup); ok {
			group = taskWithGroup.GetGroup()
		}
		if group == "" {
			group = "default" // 默认分组
		}
	}

	// 创建任务函数
	taskFunc := func() {
		ctx := context.Background()
		startTime := time.Now()
		tm.log.Info("Running scheduled task", "task", task.Task.GetName())

		// 执行任务
		err := task.Task.Run(ctx)
		duration := time.Since(startTime).Milliseconds()

		// 记录执行结果
		if err != nil {
			tm.log.Error("Scheduled task failed", "task", task.Task.GetName(), "error", err)
			tm.addFailedTask(task.Task, 3) // 默认最多重试 3 次
			tm.recordExecutionLog(task.Task.GetName(), task.Task.GetName(), group, task.CronExpr, "cron", duration, "failed", err.Error(), 0, "")
		} else {
			tm.log.Info("Scheduled task completed successfully", "task", task.Task.GetName())
			tm.recordExecutionLog(task.Task.GetName(), task.Task.GetName(), group, task.CronExpr, "cron", duration, "success", "", 0, "")
		}
	}

	// 添加任务到 cron
	entryID, err := tm.cron.AddFunc(task.CronExpr, taskFunc)
	if err != nil {
		tm.log.Error("Failed to add task to cron", "error", err)
		return err
	}

	// 记录运行中的任务
	tm.runningTasksMux.Lock()
	tm.runningTasks[task.Task.GetName()] = entryID
	tm.runningTasksMux.Unlock()

	// 记录任务信息
	tm.taskInfoMux.Lock()
	tm.taskInfoMap[task.Task.GetName()] = &TaskInfo{
		Task:        task.Task,
		ExecuteAt:   time.Now(),
		Attempts:    0,
		MaxAttempts: 3,
		CronExpr:    task.CronExpr,
		Group:       group,
	}
	tm.taskInfoMux.Unlock()

	// 记录任务分组
	tm.taskGroupMux.Lock()
	tm.taskGroupMap[task.Task.GetName()] = group
	tm.taskGroupMux.Unlock()

	tm.log.Info("Scheduled task added", "task", task.Task.GetName(), "cron_expr", task.CronExpr, "group", group)
	return nil
}

// RemoveScheduledTask 移除定时任务
func (tm *TaskManager) RemoveScheduledTask(taskName string) error {
	tm.runningTasksMux.RLock()
	entryID, exists := tm.runningTasks[taskName]
	tm.runningTasksMux.RUnlock()

	if !exists {
		err := fmt.Errorf("任务 %s 不存在", taskName)
		tm.log.Error("移除定时任务失败", "error", err)
		return err
	}

	// 移除任务
	tm.cron.Remove(entryID)

	// 从运行中任务列表移除
	tm.runningTasksMux.Lock()
	delete(tm.runningTasks, taskName)
	tm.runningTasksMux.Unlock()

	// 从任务信息列表移除
	tm.taskInfoMux.Lock()
	delete(tm.taskInfoMap, taskName)
	tm.taskInfoMux.Unlock()

	tm.log.Info("Scheduled task removed", "task", taskName)
	return nil
}

// AddOneTimeTask 添加一次性任务
func (tm *TaskManager) AddOneTimeTask(task Task, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		ctx := context.Background()
		startTime := time.Now()
		tm.log.Info("Running one-time task", "task", task.GetName())
		err := task.Run(ctx)
		duration := time.Since(startTime).Milliseconds()

		if err != nil {
			tm.log.Error("One-time task failed", "task", task.GetName(), "error", err)
			tm.addFailedTask(task, 3) // 默认最多重试 3 次
			tm.recordExecutionLog(task.GetName(), task.GetName(), "default", "", "one_time", duration, "failed", err.Error(), 0, "")
		} else {
			tm.log.Info("One-time task completed successfully", "task", task.GetName())
			tm.recordExecutionLog(task.GetName(), task.GetName(), "default", "", "one_time", duration, "success", "", 0, "")
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
		startTime := time.Now()
		tm.log.Info("Retrying failed task", "task", taskName, "attempt", taskInfo.Attempts+1)
		err := taskInfo.Task.Run(ctx)
		duration := time.Since(startTime).Milliseconds()

		if err != nil {
			tm.log.Error("Retry task failed", "task", taskName, "error", err)
			tm.addFailedTask(taskInfo.Task, taskInfo.MaxAttempts)
			tm.recordExecutionLog(taskName, taskName, taskInfo.Group, taskInfo.CronExpr, "retry", duration, "failed", err.Error(), taskInfo.Attempts, "")
		} else {
			tm.log.Info("Retry task completed successfully", "task", taskName)
			tm.recordExecutionLog(taskName, taskName, taskInfo.Group, taskInfo.CronExpr, "retry", duration, "success", "", taskInfo.Attempts, "")
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

	// 立即执行任务（带日志记录）
	go func() {
		startTime := time.Now()
		tm.log.Info("Manually running task", "task", taskName)

		// 获取任务信息
		tm.taskInfoMux.RLock()
		taskInfo, exists := tm.taskInfoMap[taskName]
		group := "default"
		cronExpr := ""
		if exists {
			group = taskInfo.Group
			cronExpr = taskInfo.CronExpr
		}
		tm.taskInfoMux.RUnlock()

		// 运行任务（Job.Run() 不返回错误）
		tm.cron.Entry(entryID).Job.Run()
		duration := time.Since(startTime).Milliseconds()

		// 假设任务成功完成（无法捕获 panic 外的错误）
		tm.log.Info("Manual task execution completed", "task", taskName)
		tm.recordExecutionLog(taskName, taskName, group, cronExpr, "manual", duration, "success", "", 0, "")
	}()

	return true
}

// GetAllTasks 获取所有任务信息
func (tm *TaskManager) GetAllTasks() []*TaskInfo {
	tm.taskInfoMux.RLock()
	defer tm.taskInfoMux.RUnlock()

	tasks := make([]*TaskInfo, 0, len(tm.taskInfoMap))
	for _, task := range tm.taskInfoMap {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetTask 获取指定任务信息
func (tm *TaskManager) GetTask(taskName string) (*TaskInfo, bool) {
	tm.taskInfoMux.RLock()
	taskInfo, exists := tm.taskInfoMap[taskName]
	tm.taskInfoMux.RUnlock()

	return taskInfo, exists
}

// IsTaskRunning 检查任务是否正在运行
func (tm *TaskManager) IsTaskRunning(taskName string) bool {
	tm.runningTasksMux.RLock()
	_, exists := tm.runningTasks[taskName]
	tm.runningTasksMux.RUnlock()

	return exists
}

// recordExecutionLog 记录任务执行日志（辅助方法）
func (tm *TaskManager) recordExecutionLog(
	taskID, taskName, group, cronExpr, triggerType string,
	duration int64, result string, errorInfo string,
	retryCount int, operatorID string,
) {
	tm.executionLogRepoMux.RLock()
	repo := tm.executionLogRepo
	tm.executionLogRepoMux.RUnlock()

	// 如果未设置日志仓库，则只记录到日志系统
	if repo == nil {
		tm.log.Info("Execution log repository not set, skipping database logging", "task_id", taskID)
		return
	}

	// 异步记录日志到数据库
	go func() {
		log := &entity.TaskExecutionLog{
			TaskID:      common.TryParseInt64(taskID),
			TaskName:    taskName,
			Group:       group,
			ExecuteAt:   time.Now().Format(time.RFC3339),
			Duration:    duration,
			Result:      result,
			ErrorInfo:   errorInfo,
			RetryCount:  retryCount,
			CronExpr:    cronExpr,
			TriggerType: triggerType,
			OperatorID:  common.TryParseInt64(operatorID),
		}

		if err := repo.Create(log); err != nil {
			tm.log.Error("Failed to record task execution log", "task_id", taskID, "error", err)
		} else {
			tm.log.Debug("Task execution log recorded", "task_id", taskID, "result", result)
		}
	}()
}

// SetTaskGroup 设置任务分组
func (tm *TaskManager) SetTaskGroup(taskName, group string) error {
	tm.taskGroupMux.Lock()
	defer tm.taskGroupMux.Unlock()

	// 检查任务是否存在
	tm.taskInfoMux.RLock()
	_, exists := tm.taskInfoMap[taskName]
	tm.taskInfoMux.RUnlock()

	if !exists {
		err := fmt.Errorf("任务 %s 不存在", taskName)
		tm.log.Error("设置任务分组失败", "error", err)
		return err
	}

	// 更新分组
	tm.taskGroupMap[taskName] = group

	// 更新任务信息中的分组
	tm.taskInfoMux.Lock()
	if taskInfo, exists := tm.taskInfoMap[taskName]; exists {
		taskInfo.Group = group
	}
	tm.taskInfoMux.Unlock()

	tm.log.Info("Task group updated", "task", taskName, "group", group)
	return nil
}

// GetTaskGroup 获取任务分组
func (tm *TaskManager) GetTaskGroup(taskName string) (string, error) {
	tm.taskGroupMux.RLock()
	group, exists := tm.taskGroupMap[taskName]
	tm.taskGroupMux.RUnlock()

	if !exists {
		err := fmt.Errorf("任务 %s 不存在", taskName)
		tm.log.Error("获取任务分组失败", "error", err)
		return "", err
	}

	return group, nil
}

// GetAllTaskGroups 获取所有任务分组
func (tm *TaskManager) GetAllTaskGroups() map[string]string {
	tm.taskGroupMux.RLock()
	defer tm.taskGroupMux.RUnlock()

	// 返回分组的副本
	groups := make(map[string]string)
	for k, v := range tm.taskGroupMap {
		groups[k] = v
	}
	return groups
}

// GetTasksByGroup 按分组获取任务
func (tm *TaskManager) GetTasksByGroup(group string) []*TaskInfo {
	tm.taskInfoMux.RLock()
	defer tm.taskInfoMux.RUnlock()

	var tasks []*TaskInfo
	for _, taskInfo := range tm.taskInfoMap {
		if taskInfo.Group == group {
			tasks = append(tasks, taskInfo)
		}
	}
	return tasks
}
