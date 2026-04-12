package response

// UserResponse 用户响应 DTO
// 所有 ID 字段使用 string 格式返回，避免前端精度丢失
type UserResponse struct {
	ID           string `json:"id,string"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Avatar       string `json:"avatar"`
	Status       int    `json:"status"`
	DepartmentID string `json:"departmentId,string"`
	PositionID   string `json:"positionId,string"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// UserSettingResponse 用户设置响应 DTO
type UserSettingResponse struct {
	ID               string `json:"id,string"`
	UserID           string `json:"userId,string"`
	ThemeColor       string `json:"themeColor"`
	SidebarColor     string `json:"sidebarColor"`
	NavbarColor      string `json:"navbarColor"`
	FontSize         int    `json:"fontSize"`
	Language         string `json:"language"`
	AutoLogin        bool   `json:"autoLogin"`
	RememberPassword bool   `json:"rememberPassword"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

// RoleInfo 角色信息响应结构
type RoleInfo struct {
	ID          string `json:"id,string"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	ParentId    string `json:"parentId,string"`
	Status      int    `json:"status"`
	IsSystem    bool   `json:"isSystem"`
	Sort        int    `json:"sort"`
}

// AuthorityInfo 权限信息结构
type AuthorityInfo struct {
	DefaultRouter string `json:"defaultRouter"`
}

// UserInfoResponse 用户信息响应
// 所有 ID 字段使用 string 格式返回，避免前端精度丢失
type UserInfoResponse struct {
	ID            string        `json:"id,string"`
	Username      string        `json:"username"`
	Nickname      string        `json:"nickname"`
	Email         string        `json:"email"`
	Phone         string        `json:"phone"`
	Avatar        string        `json:"avatar"`
	Status        int           `json:"status"`
	Roles         []*RoleInfo   `json:"roles"`
	CurrentRoleId string        `json:"currentRoleId,string"`
	Role          string        `json:"role"`
	Authority     AuthorityInfo `json:"authority"`
}
