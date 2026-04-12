package response

// APIResponse API 路由响应
type APIResponse struct {
	ID           string   `json:"id"`
	Path         string   `json:"path"`
	Method       string   `json:"method"`
	Group        string   `json:"group"`
	AuthRequired bool     `json:"authRequired"`
	AuthType     int      `json:"authType"`
	Description  string   `json:"description"`
	Status       int      `json:"status"`
	RoleIds      []string `json:"roleIds"`
	MenuIds      []string `json:"menuIds"`
	BtnPermIds   []string `json:"btnPermIds"`
	CreatedBy    string   `json:"createdBy"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedBy    string   `json:"updatedBy"`
	UpdatedAt    string   `json:"updatedAt"`
}
