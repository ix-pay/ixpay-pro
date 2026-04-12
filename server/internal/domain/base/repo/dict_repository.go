package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// DictRepository 字典仓库接口
type DictRepository interface {
	GetByID(id string) (*entity.Dict, error)
	GetByCode(dictCode string) (*entity.Dict, error)
	Create(dict *entity.Dict) error
	Update(dict *entity.Dict) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Dict, int64, error)
	GetAllActive() ([]*entity.Dict, error)
}

// DictItemRepository 字典项仓库接口
type DictItemRepository interface {
	GetByID(id string) (*entity.DictItem, error)
	GetByDictID(dictID string) ([]*entity.DictItem, error)
	Create(dictItem *entity.DictItem) error
	Update(dictItem *entity.DictItem) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.DictItem, int64, error)
	GetActiveByDictID(dictID string) ([]*entity.DictItem, error)
}
