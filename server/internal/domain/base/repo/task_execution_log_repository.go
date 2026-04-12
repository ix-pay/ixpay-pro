package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// TaskExecutionLogRepository 任务执行日志仓库接口
type TaskExecutionLogRepository interface {
	GetByID(id string) (*entity.TaskExecutionLog, error)
	Create(log *entity.TaskExecutionLog) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.TaskExecutionLog, int64, error)
	GetByTaskID(taskID string, page, pageSize int) ([]*entity.TaskExecutionLog, int64, error)
	CountByTaskID(taskID string) (int64, error)
	CountByResult(result string) (int64, error)
	GetLatestByTaskID(taskID string, limit int) ([]*entity.TaskExecutionLog, error)
	ClearExpiredLogs(beforeDate string) (int64, error)
	GetSuccessRateByTaskID(taskID string, days int) (float64, error)
	GetGroupStatistics() ([]*entity.TaskGroupStat, error)
}
