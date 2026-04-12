package response

// MenuResponse 菜单响应
// 用于返回菜单相关信息，字段名统一使用小写（符合 JSON 规范）
type MenuResponse struct {
	ID           string         `json:"id"`                 // 菜单 ID（字符串格式，使用 Snowflake ID）
	ParentID     string         `json:"parentId"`           // 父菜单 ID（字符串格式，防止 string 精度丢失）
	Path         string         `json:"path"`               // 路由路径
	Name         string         `json:"name"`               // 路由名称
	Component    string         `json:"component"`          // 组件路径
	Title        string         `json:"title"`              // 菜单标题
	Icon         string         `json:"icon"`               // 菜单图标
	Hidden       bool           `json:"hidden"`             // 是否隐藏
	Sort         int            `json:"sort"`               // 排序号
	Status       int            `json:"status"`             // 状态：0 禁用，1 启用
	IsExt        bool           `json:"isExt"`              // 是否外部链接
	Redirect     string         `json:"redirect"`           // 重定向地址
	Permission   string         `json:"permission"`         // 权限标识
	KeepAlive    bool           `json:"keepAlive"`          // 是否缓存
	DefaultMenu  bool           `json:"defaultMenu"`        // 是否默认菜单
	Breadcrumb   bool           `json:"breadcrumb"`         // 是否在面包屑显示
	ActiveMenu   string         `json:"activeMenu"`         // 当前激活的菜单
	Affix        bool           `json:"affix"`              // 是否固定在标签栏
	Type         int            `json:"type"`               // 菜单类型：1 目录，2 菜单，3 按钮，4iframe
	FrameLoading bool           `json:"frameLoading"`       // iframe 加载动画
	Meta         *MenuMetaResp  `json:"meta,omitempty"`     // 菜单元数据
	Children     []MenuResponse `json:"children,omitempty"` // 子菜单
}

// MenuMetaResp 菜单元数据响应
type MenuMetaResp struct {
	Title        string `json:"title"`        // 菜单标题
	Icon         string `json:"icon"`         // 菜单图标
	KeepAlive    bool   `json:"keepAlive"`    // 是否缓存（驼峰命名）
	DefaultMenu  bool   `json:"defaultMenu"`  // 是否默认菜单（驼峰命名）
	Breadcrumb   bool   `json:"breadcrumb"`   // 是否在面包屑显示（驼峰命名）
	ActiveMenu   string `json:"activeMenu"`   // 当前激活的菜单（驼峰命名）
	Affix        bool   `json:"affix"`        // 是否固定在标签栏（驼峰命名）
	FrameLoading bool   `json:"frameLoading"` // iframe 加载动画（驼峰命名）
}
