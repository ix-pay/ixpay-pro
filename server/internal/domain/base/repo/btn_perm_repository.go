package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// BtnPermRelation 按钮权限关联关系类型（类型安全的枚举）
type BtnPermRelation string

const (
	BtnPermRelationMenu      BtnPermRelation = "Menu"      // 所属菜单
	BtnPermRelationAPIRoutes BtnPermRelation = "APIRoutes" // API 路由
)

// BtnPermRepository 按钮权限仓库接口
type BtnPermRepository interface {
	// GetByID 根据 ID 查询按钮权限并支持加载关联数据
	// relations 参数使用 BtnPermRelation 类型，提供编译期类型检查
	GetByID(id string, relations ...BtnPermRelation) (*entity.BtnPerm, error)
	GetByCode(code string) (*entity.BtnPerm, error)
	GetBtnPermsByMenu(menuID string) ([]*entity.BtnPerm, error)
	Create(button *entity.BtnPerm) error
	Update(button *entity.BtnPerm) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.BtnPerm, int64, error)

	// 按钮权限关联 API 路由操作
	AddAPIToBtnPerm(buttonID, routeID string) error
	RemoveAPIFromBtnPerm(buttonID, routeID string) error
	GetAPIsByBtnPerm(buttonID string) ([]*entity.API, error)
	GetBtnPermsByAPI(routeID string) ([]*entity.BtnPerm, error)
}
