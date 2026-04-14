package response

// UserResponse 用户响应 DTO
// 所有 ID 字段使用 int64 格式返回，通过 json:",string" 标签自动序列化为字符串
type UserResponse struct {
	ID           int64  `json:"id,string"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Avatar       string `json:"avatar"`
	Status       int    `json:"status"`
	DepartmentID int64  `json:"departmentId,string"`
	PositionID   int64  `json:"positionId,string"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// UserSettingResponse 用户设置响应 DTO
type UserSettingResponse struct {
	ID               int64  `json:"id,string"`
	UserID           int64  `json:"userId,string"`
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
	ID          int64  `json:"id,string"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	ParentId    int64  `json:"parentId,string"`
	Status      int    `json:"status"`
	IsSystem    bool   `json:"isSystem"`
	Sort        int    `json:"sort"`
}

// AuthorityInfo 权限信息结构
type AuthorityInfo struct {
	DefaultRouter string `json:"defaultRouter"`
}

// UserInfoResponse 用户信息响应
// 所有 ID 字段使用 int64 格式返回，通过 json:",string" 标签自动序列化为字符串
type UserInfoResponse struct {
	ID            int64         `json:"id,string"`
	Username      string        `json:"username"`
	Nickname      string        `json:"nickname"`
	Email         string        `json:"email"`
	Phone         string        `json:"phone"`
	Avatar        string        `json:"avatar"`
	Status        int           `json:"status"`
	Roles         []*RoleInfo   `json:"roles"`
	CurrentRoleId int64         `json:"currentRoleId,string"`
	Role          string        `json:"role"`
	Authority     AuthorityInfo `json:"authority"`
}
