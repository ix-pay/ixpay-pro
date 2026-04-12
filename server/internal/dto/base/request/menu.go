package request

// GetMenuPageRequest 获取菜单分页列表请求
// 用于接收分页查询菜单的参数
type GetMenuPageRequest struct {
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
	Title    string `json:"title" form:"title"`
	Status   int    `json:"status" form:"status"`
}

// AddMenuRequest 创建菜单请求
// 用于接收创建新菜单的参数
type AddMenuRequest struct {
	ParentID     string   `json:"parentId" binding:"required"` // 使用 string 接收，手动转换
	Title        string   `json:"title" binding:"required,min=1,max=50"`
	Name         string   `json:"name" binding:"required,min=1,max=50"`
	Path         string   `json:"path" binding:"max=100"` // 按钮类型可以为空
	Component    string   `json:"component" binding:"max=255"`
	Icon         string   `json:"icon" binding:"max=50"`
	Sort         int      `json:"sort" binding:"min=0,max=9999"`
	Status       string   `json:"status" binding:"required,oneof=0 1"`   // 使用 string 接收，手动转换
	Type         int      `json:"type" binding:"required,oneof=1 2 3 4"` // 菜单类型：1 目录，2 菜单，3 按钮，4 iframe
	Hidden       bool     `json:"hidden"`
	IsExt        bool     `json:"isExt"`
	Redirect     string   `json:"redirect" binding:"max=100"`
	Permission   string   `json:"permission" binding:"max=100"`
	KeepAlive    bool     `json:"keepAlive"`
	DefaultMenu  bool     `json:"defaultMenu"`
	Breadcrumb   bool     `json:"breadcrumb"`
	ActiveMenu   string   `json:"activeMenu" binding:"max=100"`
	Affix        bool     `json:"affix"`
	FrameLoading bool     `json:"frameLoading"`
	ApiIds       []string `json:"apiIds"` // 关联的 API 路由 ID 列表（使用 string 数组接收）
}

// UpdateMenuRequest 更新菜单请求
// 用于接收更新菜单的参数
type UpdateMenuRequest struct {
	ID           string   `json:"id" binding:"required"`       // 使用 string 接收，手动转换
	ParentID     string   `json:"parentId" binding:"required"` // 使用 string 接收，手动转换
	Title        string   `json:"title" binding:"required,min=1,max=50"`
	Name         string   `json:"name" binding:"required,min=1,max=50"`
	Path         string   `json:"path" binding:"max=100"` // 按钮类型可以为空
	Component    string   `json:"component" binding:"max=255"`
	Icon         string   `json:"icon" binding:"max=50"`
	Sort         int      `json:"sort" binding:"min=0,max=9999"`
	Status       string   `json:"status" binding:"required,oneof=0 1"`   // 使用 string 接收，手动转换
	Type         int      `json:"type" binding:"required,oneof=1 2 3 4"` // 菜单类型：1 目录，2 菜单，3 按钮，4 iframe
	Hidden       bool     `json:"hidden"`
	IsExt        bool     `json:"isExt"`
	Redirect     string   `json:"redirect" binding:"max=100"`
	Permission   string   `json:"permission" binding:"max=100"`
	KeepAlive    bool     `json:"keepAlive"`
	DefaultMenu  bool     `json:"defaultMenu"`
	Breadcrumb   bool     `json:"breadcrumb"`
	ActiveMenu   string   `json:"activeMenu" binding:"max=100"`
	Affix        bool     `json:"affix"`
	FrameLoading bool     `json:"frameLoading"`
	ApiIds       []string `json:"apiIds"` // 关联的 API 路由 ID 列表（使用 string 数组接收）
}

// MenuResetPasswordRequest 菜单重置密码请求
// 用于接收重置用户密码的参数
type MenuResetPasswordRequest struct {
	UserID      string `json:"userId" binding:"required,numeric"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=20"`
}
