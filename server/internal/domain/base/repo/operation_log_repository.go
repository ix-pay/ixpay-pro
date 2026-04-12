package repo

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
)

// OperationLogRepository 操作日志仓库接口
type OperationLogRepository interface {
	Create(log *entity.OperationLog) error
	BatchCreate(logs []*entity.OperationLog) error
	GetByID(id string) (*entity.OperationLog, error)
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.OperationLog, int64, error)
	Delete(id string) error
	BatchDelete(ids []string) error
	DeleteByTimeRange(startTime, endTime time.Time) error
	CountByDate(date time.Time) (int64, error)
	CountByModule(module string) (int64, error)
	CountByUser(userID string) (int64, error)
	CountByOperationType(operationType entity.OperationType) (int64, error)
	CountByResult(isSuccess bool) (int64, error)
}
