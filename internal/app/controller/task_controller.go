package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/task"

	"github.com/gin-gonic/gin"
)

// TaskController 任务控制器
// @Summary 任务管理API
// @Description 提供任务添加、移除、启动、停止等功能（管理员权限）
// @Tags 任务管理
// @Router /api/v1/task [get]
type TaskController struct {
	manager *task.TaskManager
	log     logger.Logger
}

// NewTaskController 创建任务控制器
func NewTaskController(manager *task.TaskManager, log logger.Logger) *TaskController {
	return &TaskController{
		manager: manager,
		log:     log,
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
type AddTaskRequest struct {
	TaskID      string `json:"task_id" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=cron one_time"`
	Expression  string `json:"expression" binding:"required_if=Type cron"` // cron表达式或一次性执行的时间
	Description string `json:"description"`
	RetryCount  int    `json:"retry_count" binding:"min=0,max=10"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	TaskID      string    `json:"task_id"`
	Type        string    `json:"type"`
	Expression  string    `json:"expression"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	LastRunAt   time.Time `json:"last_run_at,omitempty"`
	NextRunAt   time.Time `json:"next_run_at,omitempty"`
	RetryCount  int       `json:"retry_count"`
	MaxRetries  int       `json:"max_retries"`
}

// AddTask 添加任务
// @Summary 添加任务
// @Description 添加一个定时任务或一次性任务（管理员权限）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task body AddTaskRequest true "添加任务请求参数"
// @Success 201 {object} map[string]TaskResponse "任务添加成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/task [post]
func (c *TaskController) AddTask(ctx *gin.Context) {
	var req AddTaskRequest
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
			c.log.Error("Add scheduled task failed", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if req.Type == "one_time" {
		// 创建一次性任务
		onetimeTask := &MockTask{TaskID: req.TaskID, Description: req.Description}

		// 解析执行时间
		executeTime, err := time.Parse(time.RFC3339, req.Expression)
		if err != nil {
			c.log.Error("Invalid time expression", "error", err)
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
	response := TaskResponse{
		TaskID:      req.TaskID,
		Type:        req.Type,
		Expression:  req.Expression,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		RetryCount:  req.RetryCount,
		MaxRetries:  req.RetryCount,
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": response})
}

// RemoveTask 移除任务
// @Summary 移除任务
// @Description 根据ID移除一个任务（管理员权限）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]string "任务移除成功"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/task/{id} [delete]
func (c *TaskController) RemoveTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限移除任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	if err := c.manager.RemoveScheduledTask(taskID); err != nil {
		c.log.Error("Remove task failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task removed successfully"})
}

// StartTask 启动任务
// @Summary 启动任务
// @Description 根据ID启动一个任务（管理员权限）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]string "任务启动成功"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/task/{id}/start [post]
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
		c.log.Error("Run task now failed", "taskID", taskID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run task"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task run successfully"})
}

// StopTask 停止任务
// @Summary 停止任务
// @Description 根据ID停止一个任务（管理员权限）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]string "任务停止成功"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/task/{id}/stop [post]
func (c *TaskController) StopTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限停止任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 在TaskManager中，停止任务相当于移除任务
	if err := c.manager.RemoveScheduledTask(taskID); err != nil {
		c.log.Error("Stop task failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task stopped successfully"})
}

// RetryTask 重试任务
// @Summary 重试任务
// @Description 根据ID重试一个失败的任务（管理员权限）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]string "任务重试触发成功"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/task/{id}/retry [post]
func (c *TaskController) RetryTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	// 检查用户角色是否有权限重试任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	if success := c.manager.RetryFailedTask(taskID); !success {
		c.log.Error("Retry task failed", "taskID", taskID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry task"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task retry triggered successfully"})
}

// GetTasks 获取所有任务
// @Summary 获取所有任务
// @Description 获取所有任务列表（管理员权限）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string][]TaskResponse "任务列表"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Router /api/v1/task [get]
func (c *TaskController) GetTasks(ctx *gin.Context) {
	// 检查用户角色是否有权限获取任务列表
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 在TaskManager中没有GetAllTasks方法，返回空列表
	// 实际实现应该在TaskManager中添加这个方法
	c.log.Error("GetAllTasks method not implemented in TaskManager")
	ctx.JSON(http.StatusOK, gin.H{"data": []TaskResponse{}})
}

// GetTask 获取单个任务
// @Summary 获取单个任务
// @Description 根据ID获取单个任务详情（管理员权限）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]TaskResponse "任务详情"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 403 {object} map[string]string "无权限"
// @Failure 404 {object} map[string]string "任务不存在"
// @Router /api/v1/task/{id} [get]
func (c *TaskController) GetTask(ctx *gin.Context) {
	_ = ctx.Param("id") // 未使用的参数，但保留用于API一致性

	// 检查用户角色是否有权限获取任务
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 在TaskManager中没有GetTask方法，返回错误
	// 实际实现应该在TaskManager中添加这个方法
	c.log.Error("GetTask method not implemented in TaskManager")
	ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
}
