package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// BtnPermRepository 按钮权限仓库接口
type BtnPermRepository interface {
	GetByID(id string) (*entity.BtnPerm, error)
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
