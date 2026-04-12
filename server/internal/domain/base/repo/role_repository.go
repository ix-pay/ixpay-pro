package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// RoleRelation 角色关联关系类型（类型安全的枚举）
type RoleRelation string

const (
	RoleRelationUsers            RoleRelation = "Users"            // 用户
	RoleRelationMenus            RoleRelation = "Menus"            // 菜单
	RoleRelationAPIRoutes        RoleRelation = "APIRoutes"        // API 路由
	RoleRelationBtnPerms         RoleRelation = "BtnPerms"         // 按钮权限
	RoleRelationPermissionGroups RoleRelation = "PermissionGroups" // 权限组
	RoleRelationChildren         RoleRelation = "Children"         // 子角色
	RoleRelationParent           RoleRelation = "Parent"           // 父角色
)

// RoleRepository 角色仓库接口
type RoleRepository interface {
	// GetByID 根据 ID 查询角色并支持加载关联数据
	// relations 参数使用 RoleRelation 类型，提供编译期类型检查
	GetByID(id string, relations ...RoleRelation) (*entity.Role, error)
	GetByName(name string) (*entity.Role, error)
	GetByCode(code string) (*entity.Role, error)
	Create(role *entity.Role) error
	Update(role *entity.Role) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Role, int64, error)
	GetAllRoles() ([]*entity.Role, error)

	// 角色关联用户操作
	AddUserToRole(roleID, userID string) error
	RemoveUserFromRole(roleID, userID string) error
	ExistsUserInRole(roleID, userID string) (bool, error)
	GetUsersByRole(roleID string) ([]*entity.User, error)
	GetRolesByUser(userID string) ([]*entity.Role, error)

	// 角色关联菜单操作
	AddMenuToRole(roleID, menuID string) error
	RemoveMenuFromRole(roleID, menuID string) error
	GetMenusByRole(roleID string) ([]*entity.Menu, error)
	GetRolesByMenu(menuID string) ([]*entity.Role, error)

	// 角色关联接口路由操作
	AddToRole(roleID, routeID string) error
	RemoveFromRole(roleID, routeID string) error
	GetsByRole(roleID string) ([]*entity.API, error)
	GetRolesBy(routeID string) ([]*entity.Role, error)
	GetAPIByPathAndMethod(path, method string) (*entity.API, error)

	// 角色关联按钮权限操作
	AddBtnPermToRole(roleID, btnPermID string) error
	RemoveBtnPermFromRole(roleID, btnPermID string) error
	GetBtnPermsByRole(roleID string) ([]*entity.BtnPerm, error)
	GetRolesByBtnPerm(btnPermID string) ([]*entity.Role, error)
}
