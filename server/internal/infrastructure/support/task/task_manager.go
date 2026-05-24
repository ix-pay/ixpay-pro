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

// ConcurrencyMode 并发模式
type ConcurrencyMode string

const (
	ConcurrencyAllow ConcurrencyMode = "allow" // 允许并发
	ConcurrencySkip  ConcurrencyMode = "skip"  // 跳过（如果正在运行）
	ConcurrencyWait  ConcurrencyMode = "wait"  // 排队等待
)

// ScheduledTask 定时任务
type ScheduledTask struct {
	Task        Task
	CronExpr    string
	Group       string
	Concurrency ConcurrencyMode // 并发模式，默认 skip
	Timeout     time.Duration   // 超时时间，默认 5 分钟
}

// TaskManager 任务管理器
type TaskManager struct {
	cron                *cron.Cron
	log                 logger.Logger
	failedTasks         map[string]*TaskInfo
	failedTasksMux      sync.RWMutex
	runningTasks        map[string]cron.EntryID
	runningTasksMux     sync.RWMutex
	runningTaskFlags    map[string]bool // 并发控制标志
	runningFlagMux      sync.RWMutex
	taskInfoMap         map[string]*TaskInfo
	taskInfoMux         sync.RWMutex
	taskGroupMap        map[string]string
	taskGroupMux        sync.RWMutex
	executionLogRepo    model.TaskExecutionLogRepository
	executionLogRepoMux sync.RWMutex
	taskRepo            model.TaskRepository
	taskRepoMux         sync.RWMutex
	factory             *TaskFactory
	nodeRegistry        *NodeRegistry
	defaultTimeout      time.Duration
}

// TaskInfo 任务信息
type TaskInfo struct {
	Task        Task
	ExecuteAt   time.Time
	Attempts    int
	MaxAttempts int
	CronExpr    string
	Group       string
}

// NewTaskManager 创建任务管理器
func SetupTaskManager(log logger.Logger) *TaskManager {
	c := cron.New(cron.WithSeconds())

	taskLogger := logger.GetGlobalLogger(logger.TaskLogger)
	if taskLogger == nil {
		taskLogger = log
	}

	return &TaskManager{
		cron:             c,
		log:              taskLogger,
		failedTasks:      make(map[string]*TaskInfo),
		runningTasks:     make(map[string]cron.EntryID),
		runningTaskFlags: make(map[string]bool),
		taskInfoMap:      make(map[string]*TaskInfo),
		taskGroupMap:     make(map[string]string),
		executionLogRepo: nil,
		defaultTimeout:   5 * time.Minute,
	}
}

// SetExecutionLogRepository 设置任务执行日志仓库
func (tm *TaskManager) SetExecutionLogRepository(repo model.TaskExecutionLogRepository) {
	tm.executionLogRepoMux.Lock()
	defer tm.executionLogRepoMux.Unlock()
	tm.executionLogRepo = repo
	tm.log.Info("Task execution log repository set")
}

// SetTaskRepository 设置任务配置仓库
func (tm *TaskManager) SetTaskRepository(repo model.TaskRepository) {
	tm.taskRepoMux.Lock()
	defer tm.taskRepoMux.Unlock()
	tm.taskRepo = repo
	tm.log.Info("Task repository set")
}

// SetTaskFactory 设置任务工厂
func (tm *TaskManager) SetTaskFactory(factory *TaskFactory) {
	tm.factory = factory
	tm.log.Info("Task factory set")
}

// SetNodeRegistry 设置节点注册表
func (tm *TaskManager) SetNodeRegistry(registry *NodeRegistry) {
	tm.nodeRegistry = registry
	tm.log.Info("Node registry set", "node_id", registry.GetNodeID())
}

// SetDefaultTimeout 设置默认超时时间
func (tm *TaskManager) SetDefaultTimeout(timeout time.Duration) {
	tm.defaultTimeout = timeout
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
	tm.runningTasksMux.RLock()
	_, exists := tm.runningTasks[task.Task.GetName()]
	tm.runningTasksMux.RUnlock()

	if exists {
		err := fmt.Errorf("任务 %s 已存在", task.Task.GetName())
		tm.log.Error("添加定时任务失败", "error", err)
		return err
	}

	group := task.Group
	if group == "" {
		if taskWithGroup, ok := task.Task.(TaskWithGroup); ok {
			group = taskWithGroup.GetGroup()
		}
		if group == "" {
			group = "default"
		}
	}

	if task.Concurrency == "" {
		task.Concurrency = ConcurrencySkip
	}

	if task.Timeout == 0 {
		task.Timeout = tm.defaultTimeout
	}

	tm.log.Info("Adding scheduled task",
		"task", task.Task.GetName(),
		"cron_expr", task.CronExpr,
		"group", group,
		"concurrency", task.Concurrency,
		"timeout", task.Timeout)

	taskFunc := func() {
		tm.executeTask(task, group)
	}

	entryID, err := tm.cron.AddFunc(task.CronExpr, taskFunc)
	if err != nil {
		tm.log.Error("Failed to add task to cron", "error", err)
		return err
	}

	tm.runningTasksMux.Lock()
	tm.runningTasks[task.Task.GetName()] = entryID
	tm.runningTasksMux.Unlock()

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

	tm.taskGroupMux.Lock()
	tm.taskGroupMap[task.Task.GetName()] = group
	tm.taskGroupMux.Unlock()

	tm.log.Info("Scheduled task added",
		"task", task.Task.GetName(),
		"cron_expr", task.CronExpr,
		"group", group)
	return nil
}

func (tm *TaskManager) executeTask(task *ScheduledTask, group string) {
	taskName := task.Task.GetName()

	// 并发控制
	switch task.Concurrency {
	case ConcurrencySkip:
		tm.runningFlagMux.RLock()
		isRunning := tm.runningTaskFlags[taskName]
		tm.runningFlagMux.RUnlock()
		if isRunning {
			tm.log.Warn("Task is running, skip this execution", "task", taskName)
			return
		}
	case ConcurrencyWait:
		tm.runningFlagMux.Lock()
		for tm.runningTaskFlags[taskName] {
			tm.runningFlagMux.Unlock()
			time.Sleep(100 * time.Millisecond)
			tm.runningFlagMux.Lock()
		}
		tm.runningTaskFlags[taskName] = true
		tm.runningFlagMux.Unlock()
	default:
		tm.runningFlagMux.Lock()
		tm.runningTaskFlags[taskName] = true
		tm.runningFlagMux.Unlock()
	}

	defer func() {
		tm.runningFlagMux.Lock()
		delete(tm.runningTaskFlags, taskName)
		tm.runningFlagMux.Unlock()
	}()

	ctx := logger.WithTraceID(context.Background(), logger.NewTraceID())
	taskCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	startTime := time.Now()
	taskLogger := logger.GetGlobalLogger(logger.TaskLogger).With("trace_id", logger.GetTraceID(ctx))
	taskLogger.Info("开始执行定时任务", "task", taskName)

	done := make(chan error, 1)
	go func() {
		done <- task.Task.Run(taskCtx)
	}()

	var err error
	select {
	case err = <-done:
		// 任务完成
	case <-taskCtx.Done():
		err = fmt.Errorf("任务执行超时（%v）", task.Timeout)
	}

	duration := time.Since(startTime).Milliseconds()
	traceID := logger.GetTraceID(ctx)

	if err != nil {
		taskLogger.Error("定时任务执行失败", "error", err)
		tm.addFailedTask(task.Task, 3)
		tm.recordExecutionLog(taskName, taskName, group, task.CronExpr, "cron", duration, "failed", err.Error(), 0, "", traceID)
	} else {
		taskLogger.Info("定时任务执行成功")
		tm.recordExecutionLog(taskName, taskName, group, task.CronExpr, "cron", duration, "success", "", 0, "", traceID)
	}
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

	tm.cron.Remove(entryID)

	tm.runningTasksMux.Lock()
	delete(tm.runningTasks, taskName)
	tm.runningTasksMux.Unlock()

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
		ctx := logger.WithTraceID(context.Background(), logger.NewTraceID())
		taskCtx, cancel := context.WithTimeout(ctx, tm.defaultTimeout)
		defer cancel()

		startTime := time.Now()
		taskLogger := logger.GetGlobalLogger(logger.TaskLogger).With("trace_id", logger.GetTraceID(ctx))
		taskLogger.Info("开始执行一次性任务", "task", task.GetName())

		done := make(chan error, 1)
		go func() {
			done <- task.Run(taskCtx)
		}()

		var err error
		select {
		case err = <-done:
		case <-taskCtx.Done():
			err = fmt.Errorf("任务执行超时")
		}

		duration := time.Since(startTime).Milliseconds()
		traceID := logger.GetTraceID(ctx)

		if err != nil {
			taskLogger.Error("一次性任务执行失败", "error", err)
			tm.addFailedTask(task, 3)
			tm.recordExecutionLog(task.GetName(), task.GetName(), "default", "", "one_time", duration, "failed", err.Error(), 0, "", traceID)
		} else {
			taskLogger.Info("一次性任务执行成功")
			tm.recordExecutionLog(task.GetName(), task.GetName(), "default", "", "one_time", duration, "success", "", 0, "", traceID)
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

	if taskInfo.Attempts >= taskInfo.MaxAttempts {
		tm.log.Error("Task has reached maximum retry attempts", "task", taskName)
		return false
	}

	go func() {
		ctx := logger.WithTraceID(context.Background(), logger.NewTraceID())
		taskCtx, cancel := context.WithTimeout(ctx, tm.defaultTimeout)
		defer cancel()

		startTime := time.Now()
		taskLogger := logger.GetGlobalLogger(logger.TaskLogger).With("trace_id", logger.GetTraceID(ctx))
		taskLogger.Info("开始重试失败任务", "task", taskName, "attempt", taskInfo.Attempts+1)

		done := make(chan error, 1)
		go func() {
			done <- taskInfo.Task.Run(taskCtx)
		}()

		var err error
		select {
		case err = <-done:
		case <-taskCtx.Done():
			err = fmt.Errorf("任务执行超时")
		}

		duration := time.Since(startTime).Milliseconds()
		traceID := logger.GetTraceID(ctx)

		if err != nil {
			taskLogger.Error("重试任务失败", "error", err)
			tm.addFailedTask(taskInfo.Task, taskInfo.MaxAttempts)
			tm.recordExecutionLog(taskName, taskName, taskInfo.Group, taskInfo.CronExpr, "retry", duration, "failed", err.Error(), taskInfo.Attempts, "", traceID)
		} else {
			taskLogger.Info("重试任务执行成功")
			tm.recordExecutionLog(taskName, taskName, taskInfo.Group, taskInfo.CronExpr, "retry", duration, "success", "", taskInfo.Attempts, "", traceID)
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

	go func() {
		ctx := logger.WithTraceID(context.Background(), logger.NewTraceID())
		taskLogger := logger.GetGlobalLogger(logger.TaskLogger).With("trace_id", logger.GetTraceID(ctx))

		startTime := time.Now()
		taskLogger.Info("手动执行任务", "task", taskName)

		tm.taskInfoMux.RLock()
		taskInfo, exists := tm.taskInfoMap[taskName]
		group := "default"
		cronExpr := ""
		if exists {
			group = taskInfo.Group
			cronExpr = taskInfo.CronExpr
		}
		tm.taskInfoMux.RUnlock()

		tm.cron.Entry(entryID).Job.Run()
		duration := time.Since(startTime).Milliseconds()
		traceID := logger.GetTraceID(ctx)

		taskLogger.Info("手动任务执行完成")
		tm.recordExecutionLog(taskName, taskName, group, cronExpr, "manual", duration, "success", "", 0, "", traceID)
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
	tm.runningFlagMux.RLock()
	isRunning := tm.runningTaskFlags[taskName]
	tm.runningFlagMux.RUnlock()

	return isRunning
}

// recordExecutionLog 记录任务执行日志
func (tm *TaskManager) recordExecutionLog(
	taskID, taskName, group, cronExpr, triggerType string,
	duration int64, result string, errorInfo string,
	retryCount int, operatorID string, traceID string,
) {
	tm.executionLogRepoMux.RLock()
	repo := tm.executionLogRepo
	tm.executionLogRepoMux.RUnlock()

	if repo == nil {
		tm.log.Info("Execution log repository not set, skipping database logging", "task_id", taskID)
		return
	}

	go func() {
		// 尝试解析 taskID 为数据库 ID
		logTaskID := common.TryParseInt64(taskID)

		// 如果解析失败（taskID 是逻辑名称），通过 taskRepo 查询获取真实 ID
		if logTaskID == 0 {
			tm.taskRepoMux.RLock()
			taskRepo := tm.taskRepo
			tm.taskRepoMux.RUnlock()

			if taskRepo != nil {
				if taskEntity, err := taskRepo.GetByTaskID(taskName); err == nil {
					logTaskID = taskEntity.ID
				}
			}
		}

		log := &entity.TaskExecutionLog{
			TaskID:      logTaskID,
			TraceID:     traceID,
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

		for attempt := 1; attempt <= 3; attempt++ {
			err := repo.Create(log)
			if err == nil {
				tm.log.Debug("Task execution log recorded", "task_id", taskID)
				return
			}
			tm.log.Warn("Failed to record task execution log, retrying",
				"task_id", taskID, "attempt", attempt, "error", err)
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		tm.log.Error("Failed to record task execution log after 3 attempts", "task_id", taskID)
	}()
}

// SetTaskGroup 设置任务分组
func (tm *TaskManager) SetTaskGroup(taskName, group string) error {
	tm.taskGroupMux.Lock()
	defer tm.taskGroupMux.Unlock()

	tm.taskInfoMux.RLock()
	_, exists := tm.taskInfoMap[taskName]
	tm.taskInfoMux.RUnlock()

	if !exists {
		err := fmt.Errorf("任务 %s 不存在", taskName)
		tm.log.Error("设置任务分组失败", "error", err)
		return err
	}

	tm.taskGroupMap[taskName] = group

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

// LoadTasksFromDB 从数据库加载已启用的任务
func (tm *TaskManager) LoadTasksFromDB() error {
	tm.taskRepoMux.RLock()
	repo := tm.taskRepo
	tm.taskRepoMux.RUnlock()

	if repo == nil {
		tm.log.Warn("Task repository not set, skipping load tasks from DB")
		return nil
	}

	if tm.factory == nil {
		tm.log.Warn("Task factory not set, skipping load tasks from DB")
		return nil
	}

	enabledTasks, err := repo.GetEnabledTasks()
	if err != nil {
		tm.log.Error("Failed to get enabled tasks from DB", "error", err)
		return err
	}

	tm.log.Info("Loading tasks from DB", "count", len(enabledTasks))
	for _, taskEntity := range enabledTasks {
		taskInst, err := tm.factory.CreateTask(taskEntity.TaskID, TaskType(taskEntity.TaskType), []byte(taskEntity.Params))
		if err != nil {
			tm.log.Error("Failed to create task instance from DB", "task_id", taskEntity.TaskID, "error", err)
			continue
		}

		scheduledTask := &ScheduledTask{
			Task:     taskInst,
			CronExpr: taskEntity.Expression,
			Group:    taskEntity.Group,
		}

		if err := tm.AddScheduledTask(scheduledTask); err != nil {
			tm.log.Error("Failed to add scheduled task from DB", "task_id", taskEntity.TaskID, "error", err)
			continue
		}
	}

	tm.log.Info("Tasks loaded from DB successfully", "count", len(enabledTasks))
	return nil
}

// EnableTask 启用任务（接收逻辑 taskID）
func (tm *TaskManager) EnableTask(taskID string) error {
	tm.taskRepoMux.RLock()
	repo := tm.taskRepo
	tm.taskRepoMux.RUnlock()

	if repo == nil {
		return fmt.Errorf("任务仓库未设置")
	}

	taskEntity, err := repo.GetByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败：%w", err)
	}

	if taskEntity.Status == 1 {
		return fmt.Errorf("任务已启用")
	}

	err = repo.UpdateStatus(taskEntity.TaskID, 1)
	if err != nil {
		return fmt.Errorf("更新任务状态失败：%w", err)
	}

	if tm.factory == nil {
		return fmt.Errorf("任务工厂未设置")
	}

	taskInst, err := tm.factory.CreateTask(taskEntity.TaskID, TaskType(taskEntity.TaskType), []byte(taskEntity.Params))
	if err != nil {
		return fmt.Errorf("创建任务实例失败: %w", err)
	}

	scheduledTask := &ScheduledTask{
		Task:     taskInst,
		CronExpr: taskEntity.Expression,
		Group:    taskEntity.Group,
	}

	if err := tm.AddScheduledTask(scheduledTask); err != nil {
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	tm.log.Info("Task enabled", "task_id", taskEntity.TaskID)
	return nil
}

// DisableTask 禁用任务（接收逻辑 taskID）
func (tm *TaskManager) DisableTask(taskID string) error {
	tm.taskRepoMux.RLock()
	repo := tm.taskRepo
	tm.taskRepoMux.RUnlock()

	if repo == nil {
		return fmt.Errorf("任务仓库未设置")
	}

	taskEntity, err := repo.GetByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if taskEntity.Status == 0 {
		return fmt.Errorf("任务已禁用")
	}

	err = repo.UpdateStatus(taskEntity.TaskID, 0)
	if err != nil {
		return fmt.Errorf("更新任务状态失败：%w", err)
	}

	if err := tm.RemoveScheduledTask(taskEntity.TaskID); err != nil {
		tm.log.Warn("移除定时任务失败", "task_id", taskEntity.TaskID, "error", err)
	}

	tm.log.Info("Task disabled", "task_id", taskEntity.TaskID)
	return nil
}
