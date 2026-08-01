package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// TaskRepository 任务仓库接口
type TaskRepository interface {
	Create(task *entity.Task) error
	Update(task *entity.Task) error
	GetByID(id int64) (*entity.Task, error)
	GetByTaskID(taskID string) (*entity.Task, error)
	List(filters map[string]interface{}, page, pageSize int) ([]*entity.Task, int64, error)
	ListByType(taskType string, status *int) ([]*entity.Task, error)
	ListAll() ([]*entity.Task, error)
	UpdateStatus(taskID string, status int) error
	Delete(taskID string) error
	GetEnabledTasks() ([]*entity.Task, error)
}
