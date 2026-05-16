package entity

import "time"

// Role 角色领域实体
// 系统角色管理，包含角色的基本信息、分类、继承关系和权限配置
// 纯业务模型，无 GORM 标签
type Role struct {
	ID                 int64              // 角色 ID
	Name               string             // 角色名称
	Code               string             // 角色编码
	Description        string             // 角色描述
	Type               int                // 角色类型：1-系统角色，2-业务角色，3-数据角色
	ParentID           int64              // 父角色 ID，支持角色继承
	Status             int                // 状态：1-启用，0-禁用
	IsSystem           bool               // 是否系统角色
	Sort               int                // 排序
	UserIds            []int64            // 角色关联的用户 ID 列表
	Users              []*User            // 角色关联的用户对象列表
	MenuIds            []int64            // 角色关联的菜单 ID 列表
	Menus              []*Menu            // 角色关联的菜单对象列表
	APIRouteIds        []int64            // 角色关联的接口路由 ID 列表
	APIRoutes          []*API             // 角色关联的接口路由对象列表
	BtnPermIds         []int64            // 角色关联的按钮权限 ID 列表
	BtnPerms           []*BtnPerm         // 角色关联的按钮权限对象列表
	PermissionGroupIds []int64            // 角色关联的权限组 ID 列表
	PermissionGroups   []*PermissionGroup // 角色关联的权限组对象列表
	Children           []*Role            // 子角色
	Parent             *Role              // 父角色
	CreatedBy          int64              // 创建人 ID
	CreatedAt          time.Time          // 创建时间
	UpdatedBy          int64              // 更新人 ID
	UpdatedAt          time.Time          // 更新时间
}

// HasUser 检查角色是否包含指定用户
func (r *Role) HasUser(userID int64) bool {
	for _, uid := range r.UserIds {
		if uid == userID {
			return true
		}
	}
	return false
}

// HasMenu 检查角色是否包含指定菜单
func (r *Role) HasMenu(menuID int64) bool {
	for _, mid := range r.MenuIds {
		if mid == menuID {
			return true
		}
	}
	return false
}

// HasAPIRoute 检查角色是否包含指定 API 路由
func (r *Role) HasAPIRoute(routeID int64) bool {
	for _, rid := range r.APIRouteIds {
		if rid == routeID {
			return true
		}
	}
	return false
}

// HasBtnPerm 检查角色是否包含指定按钮权限
func (r *Role) HasBtnPerm(btnPermID int64) bool {
	for _, bid := range r.BtnPermIds {
		if bid == btnPermID {
			return true
		}
	}
	return false
}

// IsActive 检查角色是否启用
func (r *Role) IsActive() bool {
	return r.Status == 1
}

// IsChildOf 检查当前角色是否是指定角色的子角色
func (r *Role) IsChildOf(parentRoleID int64) bool {
	return r.ParentID == parentRoleID
}
