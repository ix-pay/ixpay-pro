package repo

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
)

// LoginLogRepository 登录日志仓库接口
type LoginLogRepository interface {
	Create(log *entity.LoginLog) error
	GetByID(id int64) (*entity.LoginLog, error)
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.LoginLog, int64, error)
	GetByUserID(userID int64, page, pageSize int) ([]*entity.LoginLog, int64, error)
	GetByIP(ip string, page, pageSize int) ([]*entity.LoginLog, int, error)
	GetFailedByIP(ip string, hours int) ([]*entity.LoginLog, error)
	GetByTimeRange(startTime, endTime time.Time, page, pageSize int) ([]*entity.LoginLog, int, error)
	CountByDate(date time.Time) (int64, error)
	CountByUser(userID int64) (int64, error)
	CountByResult(result entity.LoginResult) (int64, error)
	GetStatistics(startTime, endTime time.Time) (*entity.LoginStatistics, error)
	// 批量操作
	BatchDelete(ids []int64) error
	ClearByTimeRange(startTime, endTime time.Time) error
}
