package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// DictRelation 字典关联关系类型（类型安全的枚举）
type DictRelation string

const (
	DictRelationItems DictRelation = "DictItems" // 字典项
)

// DictRepository 字典仓库接口
type DictRepository interface {
	// GetByID 根据 ID 查询字典并支持加载关联数据
	// relations 参数使用 DictRelation 类型，提供编译期类型检查
	GetByID(id int64, relations ...DictRelation) (*entity.Dict, error)
	// GetByCode 根据编码查询字典并支持加载关联数据
	GetByCode(dictCode string, relations ...DictRelation) (*entity.Dict, error)
	Create(dict *entity.Dict) error
	Update(dict *entity.Dict) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Dict, int64, error)
	GetAllActive() ([]*entity.Dict, error)
}

// DictItemRepository 字典项仓库接口
type DictItemRepository interface {
	GetByID(id int64) (*entity.DictItem, error)
	GetByDictID(dictID int64) ([]*entity.DictItem, error)
	Create(dictItem *entity.DictItem) error
	Update(dictItem *entity.DictItem) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.DictItem, int64, error)
	GetActiveByDictID(dictID int64) ([]*entity.DictItem, error)
}
