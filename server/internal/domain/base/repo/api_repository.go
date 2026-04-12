package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// APIRepository API 路由仓库接口
type APIRepository interface {
	BatchSave(routes []*entity.API) error
	GetByID(id string) (*entity.API, error)
	GetAllRoutes() ([]*entity.API, error)
	GetByPathAndMethod(path, method string) (*entity.API, error)
	Create(route *entity.API) error
	Update(route *entity.API) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.API, int64, error)
	GetAPIsByRole(roleID string) ([]*entity.API, error)
}
