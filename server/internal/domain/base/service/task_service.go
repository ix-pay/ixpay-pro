package service

import (
	"errors"
	"strconv"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// TaskService 任务服务
type TaskService struct {
	repo repo.TaskRepository
	log  logger.Logger
}

// NewTaskService 创建任务服务实例
func NewTaskService(repo repo.TaskRepository, log logger.Logger) *TaskService {
	return &TaskService{
		repo: repo,
		log:  log,
	}
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(task *entity.Task) error {
	if err := s.repo.Create(task); err != nil {
		s.log.Error("创建任务失败", "task_id", task.TaskID, "error", err)
		return errors.New("创建任务失败")
	}
	s.log.Info("创建任务成功", "task_id", task.TaskID)
	return nil
}

// GetTaskByID 根据数据库ID获取任务
func (s *TaskService) GetTaskByID(id int64) (*entity.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("获取任务失败", "id", id, "error", err)
		return nil, errors.New("获取任务失败")
	}
	return task, nil
}

// GetTaskByTaskID 根据任务ID获取任务
func (s *TaskService) GetTaskByTaskID(taskID string) (*entity.Task, error) {
	task, err := s.repo.GetByTaskID(taskID)
	if err != nil {
		s.log.Error("获取任务失败", "task_id", taskID, "error", err)
		return nil, errors.New("获取任务失败")
	}
	return task, nil
}

// ListTasks 分页获取任务列表
func (s *TaskService) ListTasks(filters map[string]interface{}, page, pageSize int) ([]*entity.Task, int64, error) {
	tasks, total, err := s.repo.List(filters, page, pageSize)
	if err != nil {
		s.log.Error("获取任务列表失败", "error", err)
		return nil, 0, errors.New("获取任务列表失败")
	}
	return tasks, total, nil
}

// ListTasksByType 根据类型获取任务列表
func (s *TaskService) ListTasksByType(taskType string, status *int) ([]*entity.Task, error) {
	tasks, err := s.repo.ListByType(taskType, status)
	if err != nil {
		s.log.Error("获取任务列表失败", "task_type", taskType, "error", err)
		return nil, errors.New("获取任务列表失败")
	}
	return tasks, nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(task *entity.Task) error {
	if err := s.repo.Update(task); err != nil {
		s.log.Error("更新任务失败", "task_id", task.TaskID, "error", err)
		return errors.New("更新任务失败")
	}
	s.log.Info("更新任务成功", "task_id", task.TaskID)
	return nil
}

// UpdateTaskStatus 更新任务状态
func (s *TaskService) UpdateTaskStatus(taskID string, status int) error {
	if err := s.repo.UpdateStatus(taskID, status); err != nil {
		s.log.Error("更新任务状态失败", "task_id", taskID, "error", err)
		return errors.New("更新任务状态失败")
	}
	s.log.Info("更新任务状态成功", "task_id", taskID, "status", status)
	return nil
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(taskID string) error {
	if err := s.repo.Delete(taskID); err != nil {
		s.log.Error("删除任务失败", "task_id", taskID, "error", err)
		return errors.New("删除任务失败")
	}
	s.log.Info("删除任务成功", "task_id", taskID)
	return nil
}

// GetTaskByTaskIDOrID 根据任务 ID 或数据库 ID 获取任务
func (s *TaskService) GetTaskByTaskIDOrID(id string) (*entity.Task, error) {
	// 先尝试按逻辑任务 ID 查找
	task, err := s.repo.GetByTaskID(id)
	if err == nil {
		return task, nil
	}

	// 尝试将 id 解析为 int64（数据库 ID）
	parsedID, parseErr := strconv.ParseInt(id, 10, 64)
	if parseErr != nil {
		return nil, errors.New("获取任务失败")
	}

	// 使用数据库 ID 查询
	task, err = s.repo.GetByID(parsedID)
	if err != nil {
		s.log.Error("获取任务失败", "id", id, "error", err)
		return nil, errors.New("获取任务失败")
	}

	return task, nil
}
