package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// DepartmentRepository 部门仓库接口
type DepartmentRepository interface {
	GetByID(id string) (*entity.Department, error)
	Create(department *entity.Department) error
	Update(department *entity.Department) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Department, int64, error)
	GetAll() ([]*entity.Department, error)
	GetChildrenByParentID(parentID string) ([]*entity.Department, error)
	GetDepartmentTree() ([]*entity.Department, error)
	GetDepartmentPath(id string) ([]*entity.Department, error)
}
