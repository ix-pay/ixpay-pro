package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// DepartmentRelation 部门关联关系类型（类型安全的枚举）
type DepartmentRelation string

const (
	DepartmentRelationChildren DepartmentRelation = "Children" // 子部门
	DepartmentRelationParent   DepartmentRelation = "Parent"   // 父部门
	DepartmentRelationLeader   DepartmentRelation = "Leader"   // 负责人
)

// DepartmentRepository 部门仓库接口
type DepartmentRepository interface {
	// GetByID 根据 ID 查询部门并支持加载关联数据
	// relations 参数使用 DepartmentRelation 类型，提供编译期类型检查
	GetByID(id int64, relations ...DepartmentRelation) (*entity.Department, error)
	Create(department *entity.Department) error
	Update(department *entity.Department) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Department, int64, error)
	GetAll() ([]*entity.Department, error)
	GetChildrenByParentID(parentID int64) ([]*entity.Department, error)
	GetDepartmentTree() ([]*entity.Department, error)
	GetDepartmentPath(id int64) ([]*entity.Department, error)
}
