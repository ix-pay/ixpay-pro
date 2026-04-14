package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// ConfigRepository 配置仓库接口
type ConfigRepository interface {
	GetByID(id int64) (*entity.Config, error)
	GetByKey(configKey string) (*entity.Config, error)
	Create(config *entity.Config) error
	Update(config *entity.Config) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Config, int64, error)
	GetAllActive() ([]*entity.Config, error)
}
