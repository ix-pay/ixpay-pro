package request

// CreateBtnPermRequest 创建按钮权限请求模型
type CreateBtnPermRequest struct {
	MenuID      string `json:"menuId" binding:"required,gte=1"`
	Code        string `json:"code" binding:"required,min=2,max=100"`
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=255"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

// UpdateBtnPermRequest 更新按钮权限请求模型
type UpdateBtnPermRequest struct {
	ID          string `json:"id" binding:"required,gte=1"`
	MenuID      string `json:"menuId" binding:"required,gte=1"`
	Code        string `json:"code" binding:"required,min=2,max=100"`
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=255"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

// DeleteBtnPermRequest 删除按钮权限请求模型
type DeleteBtnPermRequest struct {
	ID string `json:"id" binding:"required,gte=1"`
}

// GetBtnPermByIDRequest 根据ID获取按钮权限请求模型
type GetBtnPermByIDRequest struct {
	ID string `form:"id" binding:"required,gte=1"`
}

// GetBtnPermListRequest 获取按钮权限列表请求模型
type GetBtnPermListRequest struct {
	Page     int    `form:"page" binding:"gte=1"`
	PageSize int    `form:"pageSize" binding:"gte=1,lte=100"`
	MenuID   string `form:"menuId" binding:"gte=0"`
	Code     string `form:"code" binding:"max=100"`
	Name     string `form:"name" binding:"max=50"`
	Status   int    `form:"status" binding:"oneof=-1 0 1"`
}

// AssignToBtnPermRequest 为按钮权限分配API路由请求模型
type AssignToBtnPermRequest struct {
	BtnPermID string   `json:"btnPermId" binding:"required,gte=1"`
	IDs       []string `json:"Ids" binding:"required,min=1"`
}

// RevokeFromBtnPermRequest 从按钮权限撤销API路由请求模型
type RevokeFromBtnPermRequest struct {
	BtnPermID string `json:"btnPermId" binding:"required,gte=1"`
	ID        string `json:"Id" binding:"required,gte=1"`
}

// AssignBtnPermToRoleRequest 为角色分配按钮权限请求模型
type AssignBtnPermToRoleRequest struct {
	RoleID     string   `json:"roleId" binding:"required,gte=1"`
	BtnPermIDs []string `json:"btnPermIds" binding:"required,min=1"`
}

// RevokeBtnPermFromRoleRequest 从角色撤销按钮权限请求模型
type RevokeBtnPermFromRoleRequest struct {
	RoleID    string `json:"roleId" binding:"required,gte=1"`
	BtnPermID string `json:"btnPermId" binding:"required,gte=1"`
}
