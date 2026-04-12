package entity

// RolePermissions 角色权限缓存结构
type RolePermissions struct {
	MenuIds     []string        // 角色菜单权限 ID 列表
	BtnPermIds  []string        // 角色按钮权限 ID 列表
	APIRouteIds []string        // 角色通用 API 权限 ID 列表
	APIs        []*API          // 角色通用 API 权限
	Menus       []*Menu         // 角色菜单权限
	BtnPerms    []*BtnPerm      // 角色按钮权限
	ApiSet      map[string]bool // API 权限集合（method:path -> true）
}

// HasMenu 检查角色是否有指定菜单权限
func (r *RolePermissions) HasMenu(menuID string) bool {
	for _, mid := range r.MenuIds {
		if mid == menuID {
			return true
		}
	}
	return false
}

// HasBtnPerm 检查角色是否有指定按钮权限
func (r *RolePermissions) HasBtnPerm(btnPermID string) bool {
	for _, bid := range r.BtnPermIds {
		if bid == btnPermID {
			return true
		}
	}
	return false
}

// HasAPI 检查角色是否有指定 API 权限
func (r *RolePermissions) HasAPI(apiID string) bool {
	for _, rid := range r.APIRouteIds {
		if rid == apiID {
			return true
		}
	}
	return false
}

// HasAPIAccess 检查角色是否有指定 API 访问权限（通过 API 集合检查）
func (r *RolePermissions) HasAPIAccess(method, path string) bool {
	key := method + ":" + path
	return r.ApiSet[key]
}

// DeleteImpact 删除影响评估结果
type DeleteImpact struct {
	ChildMenusCount    int64  // 子菜单数量
	BtnPermsCount      int64  // 按钮权限数量
	AffectedRolesCount int64  // 受影响的角色数量
	AffectedApisCount  int64  // 影响的 API 数量
	Level              string // 影响等级：LOW, MEDIUM, HIGH
	Warning            string // 警告信息
}

// IsHighImpact 检查是否是高影响
func (d *DeleteImpact) IsHighImpact() bool {
	return d.Level == "HIGH"
}

// IsMediumImpact 检查是否是中等影响
func (d *DeleteImpact) IsMediumImpact() bool {
	return d.Level == "MEDIUM"
}

// IsLowImpact 检查是否是低影响
func (d *DeleteImpact) IsLowImpact() bool {
	return d.Level == "LOW"
}
