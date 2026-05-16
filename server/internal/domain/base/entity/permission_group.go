package entity

import "time"

// PermissionGroup 权限组领域实体
// 实现权限的分组管理功能
// 纯业务模型，无 GORM 标签
type PermissionGroup struct {
	ID          int64              // 权限组 ID
	Name        string             // 权限组名称
	Description string             // 权限组描述
	Status      int                // 状态：1-启用，0-禁用
	Sort        int                // 排序
	APIRouteIds []int64            // 权限组关联的接口路由 ID 列表
	APIRoutes   []*API             // 权限组关联的接口路由对象列表
	BtnPermIds  []int64            // 权限组关联的按钮权限 ID 列表
	BtnPerms    []*BtnPerm         // 权限组关联的按钮权限对象列表
	RoleIds     []int64            // 权限组关联的角色 ID 列表
	Roles       []*PermissionGroup // 权限组关联的角色对象列表
	CreatedBy   int64              // 创建人 ID
	CreatedAt   time.Time          // 创建时间
	UpdatedBy   int64              // 更新人 ID
	UpdatedAt   time.Time          // 更新时间
}

// IsActive 检查权限组是否启用
func (p *PermissionGroup) IsActive() bool {
	return p.Status == 1
}

// HasAPI 检查权限组是否包含指定 API
func (p *PermissionGroup) HasAPI(apiID int64) bool {
	for _, rid := range p.APIRouteIds {
		if rid == apiID {
			return true
		}
	}
	return false
}

// HasBtnPerm 检查权限组是否包含指定按钮权限
func (p *PermissionGroup) HasBtnPerm(btnPermID int64) bool {
	for _, bid := range p.BtnPermIds {
		if bid == btnPermID {
			return true
		}
	}
	return false
}

// HasRole 检查权限组是否包含指定角色
func (p *PermissionGroup) HasRole(roleID int64) bool {
	for _, rid := range p.RoleIds {
		if rid == roleID {
			return true
		}
	}
	return false
}
