package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// MenuRelation 菜单关联关系类型（类型安全的枚举）
type MenuRelation string

const (
	MenuRelationChildren  MenuRelation = "Children"  // 子菜单
	MenuRelationParent    MenuRelation = "Parent"    // 父菜单
	MenuRelationAPIRoutes MenuRelation = "APIRoutes" // API 路由
	MenuRelationBtnPerms  MenuRelation = "BtnPerms"  // 按钮权限
)

// MenuRepository 菜单仓库接口
type MenuRepository interface {
	// GetByID 根据 ID 查询菜单并支持加载关联数据
	// relations 参数使用 MenuRelation 类型，提供编译期类型检查
	GetByID(id int64, relations ...MenuRelation) (*entity.Menu, error)
	GetByPath(path string) (*entity.Menu, error)
	GetByCode(code string) (*entity.Menu, error)
	GetAll() ([]*entity.Menu, error)
	GetMenusByRole(roleID int64) ([]*entity.Menu, error)
	GetMenusByUserID(userID int64) ([]*entity.Menu, error)
	GetMenusByType(menuType entity.MenuType) ([]*entity.Menu, error)
	GetDefaultRouter(roleID int64) (string, error)
	GetMenuTree(parentID int64) ([]*entity.Menu, error)
	GetAllMenuTree() ([]*entity.Menu, error)
	GetMenuPath(menuID int64) ([]*entity.Menu, error)
	Create(menu *entity.Menu) error
	Update(menu *entity.Menu) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Menu, int64, error)
	BatchDelete(ids []int64) error
	CheckMenuChildren(menuID int64) (bool, error)
}
