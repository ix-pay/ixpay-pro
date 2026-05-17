package baseapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/dictconst"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// TaskController 任务控制器
//
//	@Summary		任务管理 API
//	@Description	提供任务添加、移除、启动、停止等功能（管理员权限）
//	@Tags			任务管理
//	@Router			/api/admin/task [get]
type TaskController struct {
	manager     *task.TaskManager
	log         logger.Logger
	logService  *service.TaskExecutionLogService
	taskService *service.TaskService
	factory     *task.TaskFactory
}

// NewTaskController 创建任务控制器
func NewTaskController(
	manager *task.TaskManager,
	log logger.Logger,
	logService *service.TaskExecutionLogService,
	taskService *service.TaskService,
	factory *task.TaskFactory,
) *TaskController {
	return &TaskController{
		manager:     manager,
		log:         log,
		logService:  logService,
		taskService: taskService,
		factory:     factory,
	}
}

// AddTaskRequest 添加任务请求参数
// 已移动到 internal/dto/base/request/task.go

// TaskResponse 任务响应
// 已移动到 internal/dto/base/response/task.go

// CreateTask 创建任务
//
//	@Summary		创建任务
//	@Description	创建一个新的任务（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			task	body		request.AddTaskRequest				true	"创建任务请求参数"
//	@Success		201		{object}	map[string]response.TaskResponse	"任务创建成功"
//	@Failure		400		{object}	map[string]string					"请求参数错误"
//	@Failure		401		{object}	map[string]string					"未授权"
//	@Failure		403		{object}	map[string]string					"无权限"
//	@Failure		500		{object}	map[string]string					"服务器内部错误"
//	@Router			/api/admin/task [post]
func (c *TaskController) CreateTask(ctx *gin.Context) {
	var req request.AddTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误：" + err.Error()})
		return
	}

	// 检查用户角色是否有权限添加任务
	role, exists := ctx.Get("role")
	if !exists || role != dictconst.UserTypeAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 使用任务工厂创建任务实例
	taskInst, err := c.factory.CreateTask(req.TaskID, task.TaskType(req.TaskType), req.Params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "创建任务失败：" + err.Error()})
		return
	}

	// 根据调度类型添加任务
	if req.Type == "cron" {
		scheduledTask := &task.ScheduledTask{
			Task:     taskInst,
			CronExpr: req.Expression,
			Group:    req.Group,
		}

		if err := c.manager.AddScheduledTask(scheduledTask); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误：" + err.Error()})
			return
		}
	} else if req.Type == "one_time" {
		executeTime, err := time.Parse(time.RFC3339, req.Expression)
		if err != nil {
			c.log.Error("无效的时间表达式", "error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的时间表达式"})
			return
		}

		delay := time.Until(executeTime)
		if delay < 0 {
			delay = 0
		}

		c.manager.AddOneTimeTask(taskInst, delay)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务类型"})
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

// DeleteTask 移除任务
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
func (c *TaskController) DeleteTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限移除任务
	role, exists := ctx.Get("role")
	if !exists || role != dictconst.UserTypeAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	if err := c.manager.RemoveScheduledTask(taskID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误：" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "任务移除成功"})
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
	if !exists || role != dictconst.UserTypeAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 在 TaskManager 中，定时任务添加后会自动启动，这里我们只是立即执行一次
	if success := c.manager.RunTaskNow(taskID); !success {
		c.log.Error("立即运行任务失败", "taskID", taskID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "任务运行失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "任务运行成功"})
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
	if !exists || role != dictconst.UserTypeAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 在 TaskManager 中，停止任务相当于移除任务
	if err := c.manager.RemoveScheduledTask(taskID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误：" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "任务停止成功"})
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
	if !exists || role != dictconst.UserTypeAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	if success := c.manager.RetryFailedTask(taskID); !success {
		c.log.Error("重试任务失败", "taskID", taskID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "任务重试失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "任务重试触发成功"})
}

// GetTasks 获取所有任务
//
//	@Summary		获取所有任务
//	@Description	获取所有任务列表（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int															false	"页码"
//	@Param			pageSize	query		int															false	"每页数量"
//	@Param			taskId		query		string														false	"任务ID"
//	@Param			taskType	query		string														false	"任务类型"
//	@Success		200	{object}	response.TaskListResponse	"任务列表"
//	@Failure		401	{object}	map[string]string			"未授权"
//	@Failure		403	{object}	map[string]string			"无权限"
//	@Router			/api/admin/task [get]
func (c *TaskController) GetTasks(ctx *gin.Context) {
	// 检查用户角色是否有权限获取任务列表
	role, exists := ctx.Get("role")
	if !exists || role != dictconst.UserTypeAdmin {
		baseRes.FailWithMessage("权限不足", ctx)
		return
	}

	// 获取分页参数
	var page, pageSize int
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}
	if ps := ctx.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil {
			pageSize = parsed
		}
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	taskId := ctx.Query("taskId")
	taskType := ctx.Query("taskType")

	// 从数据库查询任务
	var tasks []*entity.Task
	var total int64
	var err error

	if taskId != "" {
		task, err := c.taskService.GetTaskByTaskID(taskId)
		if err != nil {
			pageResult := baseRes.PageResult{
				List:     []response.TaskResponse{},
				Total:    0,
				Page:     page,
				PageSize: pageSize,
			}
			baseRes.OkWithDetailed(pageResult, "获取任务列表成功", ctx)
			return
		}
		tasks = []*entity.Task{task}
		total = 1
	} else if taskType != "" {
		tasks, err = c.taskService.ListTasksByType(taskType, nil)
		if err != nil {
			baseRes.FailWithMessage("查询任务失败", ctx)
			return
		}
		total = int64(len(tasks))
	} else {
		tasks, total, err = c.taskService.ListTasks(nil, page, pageSize)
		if err != nil {
			baseRes.FailWithMessage("查询任务失败", ctx)
			return
		}
	}

	// 转换为响应格式
	taskResponses := make([]response.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		dbStatus := "disabled"
		if t.Status == 1 {
			dbStatus = "enabled"
		}

		runningStatus := "stopped"
		if c.manager.IsTaskRunning(t.TaskID) {
			runningStatus = "running"
		}

		lastRunAt := ""
		if t.LastRunAt != nil {
			lastRunAt = t.LastRunAt.Format("2006-01-02 15:04:05")
		}

		nextRunAt := ""
		if t.NextRunAt != nil {
			nextRunAt = t.NextRunAt.Format("2006-01-02 15:04:05")
		}

		concurrency := t.Concurrency
		if concurrency == "" {
			concurrency = "allow"
		}

		taskResponses = append(taskResponses, response.TaskResponse{
			ID:          t.ID,
			TaskID:      t.TaskID,
			TaskType:    t.TaskType,
			Type:        t.Type,
			Expression:  t.Expression,
			Description: t.Description,
			Group:       t.Group,
			Status:      dbStatus,
			StatusLabel: runningStatus,
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04:05"),
			LastRunAt:   lastRunAt,
			NextRunAt:   nextRunAt,
			RetryCount:  t.RetryCount,
			MaxRetries:  3,
			Concurrency: concurrency,
			Timeout:     t.Timeout,
		})
	}

	pageResult := baseRes.PageResult{
		List:     taskResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	baseRes.OkWithDetailed(pageResult, "获取任务列表成功", ctx)
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
	if !exists || role != dictconst.UserTypeAdmin {
		baseRes.FailWithMessage("权限不足", ctx)
		return
	}

	// 获取分页参数
	var req request.GetTaskExecutionLogsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误："+err.Error(), ctx)
		return
	}
	req.TaskID = taskID

	// 将 TaskID 从 string 转换为 int64
	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	// 查询执行历史
	logs, total, err := c.logService.GetExecutionHistory(taskIDInt, req.Page, req.PageSize)
	if err != nil {
		baseRes.FailWithMessage("查询任务执行历史失败", ctx)
		return
	}

	// 转换为响应格式
	logResponses := make([]response.TaskExecutionLogResponse, 0, len(logs))
	for _, log := range logs {
		logResponses = append(logResponses, response.TaskExecutionLogResponse{
			ID:          log.ID,
			TaskID:      log.TaskID,
			TaskName:    log.TaskName,
			Group:       log.Group,
			ExecuteAt:   log.ExecuteAt,
			Duration:    log.Duration,
			Result:      log.Result,
			ErrorInfo:   log.ErrorInfo,
			RetryCount:  log.RetryCount,
			CronExpr:    log.CronExpr,
			TriggerType: log.TriggerType,
			OperatorID:  log.OperatorID,
		})
	}

	logsResponse := response.TaskExecutionLogsResponse{
		Logs:  logResponses,
		Total: total,
	}

	baseRes.OkWithDetailed(logsResponse, "获取任务执行日志成功", ctx)
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
	if !exists || role != dictconst.UserTypeAdmin {
		baseRes.FailWithMessage("权限不足", ctx)
		return
	}

	// 获取所有任务统计
	stats, err := c.logService.GetAllTaskStatistics()
	if err != nil {
		baseRes.FailWithMessage("获取任务统计失败", ctx)
		return
	}

	// 转换为响应格式
	statResponses := make([]response.TaskStatisticsResponse, 0, len(stats))
	for _, stat := range stats {
		statResponses = append(statResponses, response.TaskStatisticsResponse{
			TaskID:        stat.TaskID,
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

	baseRes.OkWithDetailed(gin.H{"list": statResponses}, "获取任务统计成功", ctx)
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
	if !exists || role != dictconst.UserTypeAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req request.SetTaskGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误：" + err.Error()})
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

// SearchExecutionLogs 统一搜索任务执行日志
//
//	@Summary		搜索任务执行日志
//	@Description	支持多条件搜索任务执行日志，可按任务ID、执行结果、日期范围筛选（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			taskId		query		string														false	"任务 ID"
//	@Param			result		query		string														false	"执行结果（success/failed）"
//	@Param			startDate	query		string														false	"开始日期（RFC3339格式）"
//	@Param			endDate		query		string														false	"结束日期（RFC3339格式）"
//	@Param			page		query		int															false	"页码"
//	@Param			pageSize	query		int															false	"每页数量"
//	@Success		200		{object}	map[string]response.TaskExecutionLogsResponse				"执行日志列表"
//	@Failure		400		{object}	map[string]string		"请求参数错误"
//	@Failure		401		{object}	map[string]string		"未授权"
//	@Failure		403		{object}	map[string]string		"无权限"
//	@Router			/api/admin/task/execution-logs [get]
func (c *TaskController) SearchExecutionLogs(ctx *gin.Context) {
	// 检查用户角色是否有权限获取任务日志
	role, exists := ctx.Get("role")
	if !exists || role != dictconst.UserTypeAdmin {
		baseRes.FailWithMessage("权限不足", ctx)
		return
	}

	var req request.SearchTaskExecutionLogsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误："+err.Error(), ctx)
		return
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 将 TaskID 从 string 转换为 *int64
	var taskID *int64
	if req.TaskID != "" {
		id, err := strconv.ParseInt(req.TaskID, 10, 64)
		if err != nil {
			baseRes.FailWithMessage("无效的 ID 格式", ctx)
			return
		}
		taskID = &id
	}

	// 查询执行历史
	logs, total, err := c.logService.SearchExecutionLogs(req.Page, req.PageSize, taskID, req.Result, req.StartDate, req.EndDate)
	if err != nil {
		baseRes.FailWithMessage("查询任务执行日志失败", ctx)
		return
	}

	// 转换为响应格式
	logResponses := make([]response.TaskExecutionLogResponse, 0, len(logs))
	for _, log := range logs {
		logResponses = append(logResponses, response.TaskExecutionLogResponse{
			ID:          log.ID,
			TaskID:      log.TaskID,
			TaskName:    log.TaskName,
			Group:       log.Group,
			ExecuteAt:   log.ExecuteAt,
			Duration:    log.Duration,
			Result:      log.Result,
			ErrorInfo:   log.ErrorInfo,
			RetryCount:  log.RetryCount,
			CronExpr:    log.CronExpr,
			TriggerType: log.TriggerType,
			OperatorID:  log.OperatorID,
		})
	}

	logsResponse := response.TaskExecutionLogsResponse{
		Logs:  logResponses,
		Total: total,
	}

	baseRes.OkWithDetailed(logsResponse, "获取任务执行日志成功", ctx)
}

// GetFactory 获取任务工厂
func (c *TaskController) GetFactory() *task.TaskFactory {
	return c.factory
}

// EnableTask 启用任务
//
//	@Summary		启用任务
//	@Description	启用一个已禁用的任务（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"任务ID"
//	@Success		200	{object}	map[string]string	"任务启用成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		404	{object}	map[string]string	"任务不存在"
//	@Router			/api/admin/task/:id/enable [post]
func (c *TaskController) EnableTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	role, exists := ctx.Get("role")
	if !exists || role != dictconst.UserTypeAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	if err := c.manager.EnableTask(taskID); err != nil {
		c.log.Error("启用任务失败", "task_id", taskID, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "任务启用成功"})
}

// GetDashboard 获取任务统计面板数据
//
//	@Summary		获取任务统计面板
//	@Description	获取任务统计面板数据，包括任务总数、启用数、禁用数、今日执行数（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response	"统计面板数据"
//	@Failure		401	{object}	map[string]string		"未授权"
//	@Failure		403	{object}	map[string]string		"无权限"
//	@Router			/api/admin/task/dashboard [get]
func (c *TaskController) GetDashboard(ctx *gin.Context) {
	role, exists := ctx.Get("role")
	if !exists || role != dictconst.UserTypeAdmin {
		baseRes.FailWithMessage("权限不足", ctx)
		return
	}

	// 获取所有任务
	allTasks, _, err := c.taskService.ListTasks(nil, 1, 10000)
	if err != nil {
		allTasks = []*entity.Task{}
	}

	// 统计总数和状态
	totalTasks := len(allTasks)
	enabledCount := 0
	disabledCount := 0
	for _, t := range allTasks {
		if t.Status == 1 {
			enabledCount++
		} else {
			disabledCount++
		}
	}

	// 获取今日执行次数（从日志服务统计）
	todayExecutions := int64(0)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Format(time.RFC3339)
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Format(time.RFC3339)

	logs, _, err := c.logService.SearchExecutionLogs(1, 10000, nil, "", todayStart, todayEnd)
	if err == nil {
		todayExecutions = int64(len(logs))
	}

	baseRes.OkWithDetailed(gin.H{
		"totalTasks":      totalTasks,
		"enabledTasks":    enabledCount,
		"disabledTasks":   disabledCount,
		"todayExecutions": todayExecutions,
	}, "获取任务统计面板成功", ctx)
}

// DisableTask 禁用任务
//
//	@Summary		禁用任务
//	@Description	禁用一个已启用的任务（管理员权限）
//	@Tags			任务管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"任务ID"
//	@Success		200	{object}	map[string]string	"任务禁用成功"
//	@Failure		401	{object}	map[string]string	"未授权"
//	@Failure		403	{object}	map[string]string	"无权限"
//	@Failure		404	{object}	map[string]string	"任务不存在"
//	@Router			/api/admin/task/:id/disable [post]
func (c *TaskController) DisableTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	role, exists := ctx.Get("role")
	if !exists || role != dictconst.UserTypeAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	if err := c.manager.DisableTask(taskID); err != nil {
		c.log.Error("禁用任务失败", "task_id", taskID, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "任务禁用成功"})
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
	idStr := ctx.Param("id")

	// 检查用户角色是否有权限获取任务
	role, exists := ctx.Get("role")
	if !exists || role != dictconst.UserTypeAdmin {
		baseRes.FailWithMessage("权限不足", ctx)
		return
	}

	// 将 ID 从 string 转换为 int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	// 从数据库获取任务
	task, err := c.taskService.GetTaskByID(id)
	if err != nil {
		baseRes.FailWithMessage("任务不存在", ctx)
		return
	}

	// 转换为响应格式
	dbStatus := "disabled"
	if task.Status == 1 {
		dbStatus = "enabled"
	}

	runningStatus := "stopped"
	if c.manager.IsTaskRunning(task.TaskID) {
		runningStatus = "running"
	}

	lastRunAt := ""
	if task.LastRunAt != nil {
		lastRunAt = task.LastRunAt.Format("2006-01-02 15:04:05")
	}

	nextRunAt := ""
	if task.NextRunAt != nil {
		nextRunAt = task.NextRunAt.Format("2006-01-02 15:04:05")
	}

	concurrency := task.Concurrency
	if concurrency == "" {
		concurrency = "allow"
	}

	taskResponse := response.TaskResponse{
		ID:          task.ID,
		TaskID:      task.TaskID,
		TaskType:    task.TaskType,
		Type:        task.Type,
		Expression:  task.Expression,
		Description: task.Description,
		Group:       task.Group,
		Status:      dbStatus,
		StatusLabel: runningStatus,
		CreatedAt:   task.CreatedAt.Format("2006-01-02 15:04:05"),
		LastRunAt:   lastRunAt,
		NextRunAt:   nextRunAt,
		RetryCount:  task.RetryCount,
		MaxRetries:  3,
		Concurrency: concurrency,
		Timeout:     task.Timeout,
	}

	baseRes.OkWithDetailed(taskResponse, "获取任务详情成功", ctx)
}
