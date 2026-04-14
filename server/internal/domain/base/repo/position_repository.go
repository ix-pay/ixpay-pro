package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// PositionRepository 岗位仓库接口
type PositionRepository interface {
	GetByID(id int64) (*entity.Position, error)
	Create(position *entity.Position) error
	Update(position *entity.Position) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Position, int64, error)
	GetAll() ([]*entity.Position, error)
	GetByName(name string) (*entity.Position, error)
}
