package request

// GetAPIPageRequest 获取 API 路由分页列表请求
type GetAPIPageRequest struct {
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
	Keyword  string `json:"keyword" form:"keyword"`
	Group    string `json:"group" form:"group"`
}

// CreateAPIRequest 创建 API 路由请求
type CreateAPIRequest struct {
	Path         string   `json:"path" binding:"required,max=255"`
	Method       string   `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	Group        string   `json:"group" binding:"max=100"`
	AuthRequired bool     `json:"authRequired"`
	AuthType     int      `json:"authType" binding:"oneof=0 1"`
	Description  string   `json:"description" binding:"max=500"`
	Status       int      `json:"status" binding:"oneof=0 1"`
	RoleIds      []string `json:"roleIds"`
	MenuIds      []string `json:"menuIds"`
	BtnPermIds   []string `json:"btnPermIds"`
}

// UpdateAPIRequest 更新 API 路由请求
type UpdateAPIRequest struct {
	ID           int64    `json:"id" binding:"required"`
	Path         string   `json:"path" binding:"required,max=255"`
	Method       string   `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	Group        string   `json:"group" binding:"max=100"`
	AuthRequired bool     `json:"authRequired"`
	AuthType     int      `json:"authType" binding:"oneof=0 1"`
	Description  string   `json:"description" binding:"max=500"`
	Status       int      `json:"status" binding:"oneof=0 1"`
	RoleIds      []string `json:"roleIds"`
	MenuIds      []string `json:"menuIds"`
	BtnPermIds   []string `json:"btnPermIds"`
}
