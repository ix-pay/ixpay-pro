package request

// CreateRoleRequest 创建角色请求模型
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=255"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

// UpdateRoleRequest 更新角色请求模型
type UpdateRoleRequest struct {
	ID          int64  `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=255"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

// DeleteRoleRequest 删除角色请求模型
type DeleteRoleRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// GetRoleByIDRequest 根据 ID 获取角色请求模型
type GetRoleByIDRequest struct {
	ID int64 `form:"id" binding:"required"`
}

// GetRoleListRequest 获取角色列表请求模型
type GetRoleListRequest struct {
	Page     int    `form:"page" binding:"gte=1"`
	PageSize int    `form:"pageSize" binding:"gte=1,lte=100"`
	Name     string `form:"name" binding:"max=50"`
	Status   *int   `form:"status" binding:"omitempty,oneof=-1 0 1"`
}

// AssignUserToRoleRequest 分配用户到角色请求模型
type AssignUserToRoleRequest struct {
	RoleID  int64    `json:"roleId" binding:"required,gte=1"`
	UserIDs []string `json:"userIds" binding:"required,min=1"`
}

// AssignMenuToRoleRequest 分配菜单到角色请求模型
type AssignMenuToRoleRequest struct {
	RoleID  int64    `json:"roleId" binding:"required,gte=1"`
	MenuIDs []string `json:"menuIds" binding:"required,min=1"`
}

// AssignToRoleRequest 分配 API 路由到角色请求模型
type AssignToRoleRequest struct {
	RoleID int64    `json:"roleId" binding:"required,gte=1"`
	IDs    []string `json:"Ids" binding:"required,min=1"`
}

// SaveRolePermissionsRequest 保存角色权限请求模型
type SaveRolePermissionsRequest struct {
	MenuIds     []string `json:"menuIds" binding:"required"`
	BtnPermIds  []string `json:"btnPermIds" binding:"required"`
	ApiRouteIds []string `json:"apiRouteIds" binding:"required"`
}
