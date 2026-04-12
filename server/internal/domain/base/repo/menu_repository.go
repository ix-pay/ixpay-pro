package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// MenuRepository 菜单仓库接口
type MenuRepository interface {
	GetByID(id string) (*entity.Menu, error)
	GetByPath(path string) (*entity.Menu, error)
	GetByCode(code string) (*entity.Menu, error)
	GetAll() ([]*entity.Menu, error)
	GetMenusByRole(role string) ([]*entity.Menu, error)
	GetMenusByUserID(userID string) ([]*entity.Menu, error)
	GetMenusByType(menuType entity.MenuType) ([]*entity.Menu, error)
	GetDefaultRouter(role string) (string, error)
	GetMenuTree(parentID string) ([]*entity.Menu, error)
	GetAllMenuTree() ([]*entity.Menu, error)
	GetMenuPath(menuID string) ([]*entity.Menu, error)
	Create(menu *entity.Menu) error
	Update(menu *entity.Menu) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Menu, int64, error)
	BatchDelete(ids []string) error
	CheckMenuChildren(menuID string) (bool, error)
}
