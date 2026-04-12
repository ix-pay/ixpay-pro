package entity

import "time"

// API API 接口路由领域实体
// 纯业务模型，无 GORM 标签
type API struct {
	ID           string    // API ID
	Path         string    // 路由路径
	Method       string    // HTTP 方法
	Group        string    // 路由分组
	AuthRequired bool      // 是否需要认证
	AuthType     int       // 授权类型：0-不需要授权（只要登录），1-需要授权（需要角色权限）
	Description  string    // 描述
	Status       int       // 状态：1-启用，0-禁用
	RoleIds      []string  // 关联的角色 ID 列表
	MenuIds      []string  // 关联的菜单 ID 列表
	BtnPermIds   []string  // 关联的按钮权限 ID 列表
	CreatedBy    string    // 创建人 ID
	CreatedAt    time.Time // 创建时间
	UpdatedBy    string    // 更新人 ID
	UpdatedAt    time.Time // 更新时间
}

// IsActive 检查 API 是否启用
func (a *API) IsActive() bool {
	return a.Status == 1
}

// RequireAuth 检查 API 是否需要认证
func (a *API) RequireAuth() bool {
	return a.AuthRequired
}

// RequirePermission 检查 API 是否需要角色权限
func (a *API) RequirePermission() bool {
	return a.AuthType == 1
}
