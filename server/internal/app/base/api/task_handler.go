package baseapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
)

// TaskController 任务控制器
//
//	@Summary		任务管理 API
//	@Description	提供任务添加、移除、启动、停止等功能（管理员权限）
//	@Tags			任务管理
//	@Router			/api/admin/task [get]
type TaskController struct {
	manager *task.TaskManager
	log     logger.Logger
	service *service.TaskExecutionLogService
}

// NewTaskController 创建任务控制器
func NewTaskController(
	manager *task.TaskManager,
	log logger.Logger,
	service *service.TaskExecutionLogService,
) *TaskController {
	return &TaskController{
		manager: manager,
		log:     log,
		service: service,
	}
}

// MockTask 模拟任务实现，用于测试
type MockTask struct {
	TaskID      string
	Description string
} // Run 实现Task接口的Run方法
func (t *MockTask) Run(ctx context.Context) error {
	// 模拟任务执行
	fmt.Printf("Running task: %s, Description: %s\n", t.TaskID, t.Description)
	return nil
} // GetName 实现Task接口的GetName方法
func (t *MockTask) GetName() string {
	return t.TaskID
}

// AddTaskRequest 添加任务请求参数
// 已移动到 internal/dto/base/request/task.go

// TaskResponse 任务响应
// 已移动到 internal/dto/base/response/task.go

// AddTask 添加任务
//
//	@Summary		添加任务
//	@Description	添加一个定时任务或一次性任务（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			task	body		request.AddTaskRequest				true	"添加任务请求参数"
//	@Success		201		{object}	map[string]response.TaskResponse	"任务添加成功"
//	@Failure		400		{object}	map[string]string					"请求参数错误"
//	@Failure		401		{object}	map[string]string					"未授权"
//	@Failure		403		{object}	map[string]string					"无权限"
//	@Failure		500		{object}	map[string]string					"服务器内部错误"
//	@Router			/api/admin/task [post]
func (c *TaskController) AddTask(ctx *gin.Context) {
	var req request.AddTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户角色是否有权限添加任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 根据任务类型添加任务
	if req.Type == "cron" {
		// 创建定时任务
		cronTask := &MockTask{TaskID: req.TaskID, Description: req.Description}
		scheduledTask := &task.ScheduledTask{
			Task:     cronTask,
			CronExpr: req.Expression,
		}

		if err := c.manager.AddScheduledTask(scheduledTask); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if req.Type == "one_time" {
		// 创建一次性任务
		onetimeTask := &MockTask{TaskID: req.TaskID, Description: req.Description}

		// 解析执行时间
		executeTime, err := time.Parse(time.RFC3339, req.Expression)
		if err != nil {
			c.log.Error("无效的时间表达式", "error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time expression"})
			return
		}

		// 计算延迟时间
		delay := time.Until(executeTime)
		if delay < 0 {
			delay = 0
		}

		c.manager.AddOneTimeTask(onetimeTask, delay)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task type"})
		return
	}

	// 构建响应
	taskResponse := response.TaskResponse{
		TaskID:      req.TaskID,
		Type:        req.Type,
		Expression:  req.Expression,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   "",
		RetryCount:  req.RetryCount,
		MaxRetries:  req.RetryCount,
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": taskResponse})
}

// RemoveTask 移除任务
//
//	@Summary		移除任务
//	@Description	根据ID移除一个任务（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"任务ID"
//	@Success		200	{object}	map[string]string	"任务移除成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		500	{object}	map[string]string	"服务器内部错误"
//	@Router			/api/admin/task/:id [delete]
func (c *TaskController) RemoveTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限移除任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	if err := c.manager.RemoveScheduledTask(taskID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task removed successfully"})
}

// StartTask 启动任务
//
//	@Summary		启动任务
//	@Description	根据ID启动一个任务（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"任务ID"
//	@Success		200	{object}	map[string]string	"任务启动成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		500	{object}	map[string]string	"服务器内部错误"
//	@Router			/api/admin/task/:id/start [post]
func (c *TaskController) StartTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限启动任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 在TaskManager中，定时任务添加后会自动启动，这里我们只是立即执行一次
	if success := c.manager.RunTaskNow(taskID); !success {
		c.log.Error("立即运行任务失败", "taskID", taskID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run task"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task run successfully"})
}

// StopTask 停止任务
//
//	@Summary		停止任务
//	@Description	根据ID停止一个任务（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"任务ID"
//	@Success		200	{object}	map[string]string	"任务停止成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		500	{object}	map[string]string	"服务器内部错误"
//	@Router			/api/admin/task/:id/stop [post]
func (c *TaskController) StopTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限停止任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 在 TaskManager 中，停止任务相当于移除任务
	if err := c.manager.RemoveScheduledTask(taskID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task stopped successfully"})
}

// RetryTask 重试任务
//
//	@Summary		重试任务
//	@Description	根据ID重试一个失败的任务（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"任务ID"
//	@Success		200	{object}	map[string]string	"任务重试触发成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		500	{object}	map[string]string	"服务器内部错误"
//	@Router			/api/admin/task/:id/retry [post]
func (c *TaskController) RetryTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限重试任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	if success := c.manager.RetryFailedTask(taskID); !success {
		c.log.Error("重试任务失败", "taskID", taskID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry task"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task retry triggered successfully"})
}

// GetTasks 获取所有任务
//
//	@Summary		获取所有任务
//	@Description	获取所有任务列表（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string][]response.TaskResponse	"任务列表"
//	@Failure		401	{object}	map[string]string			"未授权"
//	@Failure		403	{object}	map[string]string			"无权限"
//	@Router			/api/admin/task [get]
func (c *TaskController) GetTasks(ctx *gin.Context) {
	// 检查用户角色是否有权限获取任务列表
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 获取所有任务
	tasks := c.manager.GetAllTasks()

	// 转换为响应格式
	taskResponses := make([]response.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		// 获取任务状态
		status := "running"
		if !c.manager.IsTaskRunning(t.Task.GetName()) {
			status = "stopped"
		}

		taskResponses = append(taskResponses, response.TaskResponse{
			TaskID:      t.Task.GetName(),
			Type:        "cron", // 目前只支持定时任务
			Expression:  t.CronExpr,
			Description: "", // Task 接口没有提供描述方法
			Status:      status,
			CreatedAt:   "",
			RetryCount:  t.Attempts,
			MaxRetries:  t.MaxAttempts,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"data": taskResponses})
}

// ExecutionLogResponse 执行日志响应
// 已移动到 internal/dto/base/response/task.go

// ExecutionLogsResponse 执行日志列表响应
// 已移动到 internal/dto/base/response/task.go

// GetExecutionLogs 查询任务执行历史
//
//	@Summary		查询任务执行历史
//	@Description	根据任务 ID 查询任务执行历史记录（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string														true	"任务 ID"
//	@Param			page	query		int															false	"页码"
//	@Param			pageSize	query		int														false	"每页数量"
//	@Success		200		{object}	map[string]response.TaskExecutionLogsResponse				"执行日志列表"
//	@Failure		401		{object}	map[string]string		"未授权"
//	@Failure		403		{object}	map[string]string		"无权限"
//	@Router			/api/admin/task/:id/execution-logs [get]
func (c *TaskController) GetExecutionLogs(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限获取任务日志
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 获取分页参数
	var req request.GetTaskExecutionLogsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.TaskID = taskID

	// 将 TaskID 从 string 转换为 int64
	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID 格式"})
		return
	}

	// 查询执行历史
	logs, total, err := c.service.GetExecutionHistory(taskIDInt, req.Page, req.PageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务执行历史失败"})
		return
	}

	// 转换为响应格式
	logResponses := make([]response.TaskExecutionLogResponse, 0, len(logs))
	for _, log := range logs {
		logResponses = append(logResponses, response.TaskExecutionLogResponse{
			ID:          strconv.FormatInt(log.ID, 10),
			TaskID:      strconv.FormatInt(log.TaskID, 10),
			TaskName:    log.TaskName,
			Group:       log.Group,
			ExecuteAt:   log.ExecuteAt,
			Duration:    log.Duration,
			Result:      log.Result,
			ErrorInfo:   log.ErrorInfo,
			RetryCount:  log.RetryCount,
			CronExpr:    log.CronExpr,
			TriggerType: log.TriggerType,
			OperatorID:  strconv.FormatInt(log.OperatorID, 10),
		})
	}

	logsResponse := response.TaskExecutionLogsResponse{
		Logs:  logResponses,
		Total: total,
	}

	ctx.JSON(http.StatusOK, gin.H{"data": logsResponse})
}

// TaskStatisticsResponse 任务统计响应
// 已移动到 internal/dto/base/response/task.go

// GetStatistics 任务统计
//
//	@Summary		获取任务统计
//	@Description	获取所有任务的执行统计信息（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string][]response.TaskStatisticsResponse	"任务统计列表"
//	@Failure		401	{object}	map[string]string		"未授权"
//	@Failure		403	{object}	map[string]string		"无权限"
//	@Router			/api/admin/task/statistics [get]
func (c *TaskController) GetStatistics(ctx *gin.Context) {
	// 检查用户角色是否有权限获取统计
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 获取所有任务统计
	stats, err := c.service.GetAllTaskStatistics()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务统计失败"})
		return
	}

	// 转换为响应格式
	statResponses := make([]response.TaskStatisticsResponse, 0, len(stats))
	for _, stat := range stats {
		statResponses = append(statResponses, response.TaskStatisticsResponse{
			TaskID:        strconv.FormatInt(stat.TaskID, 10),
			TaskName:      stat.TaskName,
			Group:         stat.Group,
			TotalExecutes: stat.TotalExecutes,
			SuccessCount:  stat.SuccessCount,
			FailedCount:   stat.FailedCount,
			SuccessRate:   stat.SuccessRate,
			AvgDuration:   stat.AvgDuration,
			LastExecuteAt: stat.LastExecuteAt,
			NextExecuteAt: stat.NextExecuteAt,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"data": statResponses})
}

// SetTaskGroupRequest 设置任务分组请求
// 已移动到 internal/dto/base/request/task.go

// SetTaskGroup 设置任务分组
//
//	@Summary		设置任务分组
//	@Description	设置指定任务的分组（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string								true	"任务 ID"
//	@Param			task	body		request.SetTaskGroupRequest			true	"分组信息"
//	@Success		200		{object}	map[string]string					"设置成功"
//	@Failure		400		{object}	map[string]string					"请求参数错误"
//	@Failure		401		{object}	map[string]string					"未授权"
//	@Failure		403		{object}	map[string]string					"无权限"
//	@Failure		404		{object}	map[string]string					"任务不存在"
//	@Router			/api/admin/task/:id/group [post]
func (c *TaskController) SetTaskGroup(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限设置分组
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	var req request.SetTaskGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置任务分组
	if err := c.manager.SetTaskGroup(taskID, req.Group); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "任务分组设置成功"})
}

// GetManager 获取任务管理器（用于设置执行日志仓库）
func (c *TaskController) GetManager() *task.TaskManager {
	return c.manager
}

// GetTask 获取单个任务
//
//	@Summary		获取单个任务
//	@Description	根据 ID 获取单个任务详情（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"任务 ID"
//	@Success		200	{object}	map[string]response.TaskResponse	"任务详情"
//	@Failure		401	{object}	map[string]string		"未授权"
//	@Failure		403	{object}	map[string]string		"无权限"
//	@Failure		404	{object}	map[string]string		"任务不存在"
//	@Router			/api/admin/task/:id [get]
func (c *TaskController) GetTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限获取任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 获取单个任务
	task, exists := c.manager.GetTask(taskID)
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 转换为响应格式
	status := "running"
	if !c.manager.IsTaskRunning(task.Task.GetName()) {
		status = "stopped"
	}

	response := response.TaskResponse{
		TaskID:      task.Task.GetName(),
		Type:        "cron", // 目前只支持定时任务
		Expression:  task.CronExpr,
		Description: "", // Task 接口没有提供描述方法
		Status:      status,
		CreatedAt:   "",
		RetryCount:  task.Attempts,
		MaxRetries:  task.MaxAttempts,
	}

	ctx.JSON(http.StatusOK, gin.H{"data": response})
}
