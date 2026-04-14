package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// RoleDTO 角色 DTO（用于用户详情中的角色信息）
type RoleDTO struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// RoleResponse 角色响应模型
// 所有 ID 字段使用 int64 格式，通过 json:",string" 标签自动序列化为字符串
type RoleResponse struct {
	ID          int64  `json:"id,string"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	ParentId    int64  `json:"parentId,string"`
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
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// MenuInfo 菜单简略信息
type MenuInfo struct {
	ID       int64  `json:"id,string"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ParentID int64  `json:"parentId,string"`
}

// RouteInfo API 路由简略信息
type RouteInfo struct {
	ID     int64  `json:"id,string"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Group  string `json:"group"`
}
