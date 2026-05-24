package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// TaskExecutionLogRepository 任务执行日志仓库接口
type TaskExecutionLogRepository interface {
	GetByID(id int64) (*entity.TaskExecutionLog, error)
	Create(log *entity.TaskExecutionLog) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.TaskExecutionLog, int64, error)
	GetByTaskID(taskID int64, page, pageSize int) ([]*entity.TaskExecutionLog, int64, error)
	GetByTaskName(taskName string, page, pageSize int) ([]*entity.TaskExecutionLog, int64, error)
	CountByTaskID(taskID int64) (int64, error)
	CountByTaskIDAndResult(taskID int64, result string) (int64, error)
	CountByResult(result string) (int64, error)
	GetLatestByTaskID(taskID int64, limit int) ([]*entity.TaskExecutionLog, error)
	ClearExpiredLogs(beforeDate string) (int64, error)
	GetSuccessRateByTaskID(taskID int64, days int) (float64, error)
	GetGroupStatistics() ([]*entity.TaskGroupStat, error)
	// Search 多条件搜索任务执行日志（支持任务ID、执行结果、日期范围筛选）
	Search(page, pageSize int, taskID *int64, result string, startDate, endDate string) ([]*entity.TaskExecutionLog, int64, error)
}
