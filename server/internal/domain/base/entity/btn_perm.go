package entity

import "time"

// BtnPerm 按钮级权限领域实体
// 用于细粒度控制模块内的按钮操作权限
// 纯业务模型，无 GORM 标签
type BtnPerm struct {
	ID          int64     // 按钮权限 ID
	MenuID      int64     // 所属菜单 ID
	Code        string    // 权限编码，如：user:create, user:edit
	Name        string    // 权限名称，如：创建用户，编辑用户
	Description string    // 权限描述
	Status      int       // 状态：1-启用，0-禁用
	APIRouteIds []int64   // 关联的 API 路由 ID 列表
	APIRoutes   []*API    // 关联的 API 路由对象列表
	RoleIds     []int64   // 关联的角色 ID 列表
	Roles       []*Role   // 关联的角色对象列表关联的角色列表（新增）
	Menu        *Menu     // 所属菜单（新增）
	CreatedBy   int64     // 创建人 ID
	CreatedAt   time.Time // 创建时间
	UpdatedBy   int64     // 更新人 ID
	UpdatedAt   time.Time // 更新时间
}

// IsActive 检查按钮权限是否启用
func (b *BtnPerm) IsActive() bool {
	return b.Status == 1
}
