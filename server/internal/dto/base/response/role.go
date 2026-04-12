package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// RoleResponse 角色响应模型
// 所有 ID 字段使用 string 格式，避免前端精度丢失
type RoleResponse struct {
	ID          string `json:"id,string"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	ParentId    string `json:"parentId,string"`
	Status      int    `json:"status"`
	IsSystem    bool   `json:"isSystem"`
	Sort        int    `json:"sort"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// RoleListResponse 角色列表响应模型
type RoleListResponse struct {
	baseRes.PageResult
	List []RoleResponse `json:"list"`
}

// RoleDetailResponse 角色详情响应模型
type RoleDetailResponse struct {
	RoleResponse
	Users  []UserInfo  `json:"users"`
	Menus  []MenuInfo  `json:"menus"`
	Routes []RouteInfo `json:"routes"`
}

// UserInfo 用户简略信息
type UserInfo struct {
	ID       string `json:"id,string"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// MenuInfo 菜单简略信息
type MenuInfo struct {
	ID       string `json:"id,string"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ParentID string `json:"parentId,string"`
}

// RouteInfo API 路由简略信息
type RouteInfo struct {
	ID     string `json:"id,string"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Group  string `json:"group"`
}
