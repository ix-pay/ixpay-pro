package response

// APIResponse API 路由响应
type APIResponse struct {
	ID           int64   `json:"id,string"`
	Path         string  `json:"path"`
	Method       string  `json:"method"`
	Group        string  `json:"group"`
	AuthRequired bool    `json:"authRequired"`
	AuthType     int     `json:"authType"`
	Description  string  `json:"description"`
	Status       int     `json:"status"`
	RoleIds      []int64 `json:"roleIds"`
	MenuIds      []int64 `json:"menuIds"`
	BtnPermIds   []int64 `json:"btnPermIds"`
	CreatedBy    int64   `json:"createdBy,string"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedBy    int64   `json:"updatedBy,string"`
	UpdatedAt    string  `json:"updatedAt"`
}
