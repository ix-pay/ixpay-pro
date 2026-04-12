package response

import (
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
	"time"
)

// BtnPermResponse 按钮权限响应模型
// 用于返回按钮权限的基本信息
type BtnPermResponse struct {
	ID          string    `json:"id"`
	MenuID      string    `json:"menuId"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// BtnPermDetailResponse 按钮权限详情响应模型
// 用于返回按钮权限的详细信息，包括关联的菜单和 API 路由
type BtnPermDetailResponse struct {
	BtnPermResponse
	Menu      MenuInfo    `json:"menu"`
	APIRoutes []RouteInfo `json:"apiRoutes"`
}

// BtnPermListResponse 按钮权限列表响应模型
// 用于分页返回按钮权限列表
type BtnPermListResponse struct {
	baseRes.PageResult
	List []BtnPermResponse `json:"list"`
}

// BtnPermForRole 角色关联的按钮权限响应模型
// 用于返回角色拥有的按钮权限信息
type BtnPermForRole struct {
	ID          string `json:"id"`
	MenuID      string `json:"menuId"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	IsAssigned  bool   `json:"isAssigned"`
}
