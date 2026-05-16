package entity

import "time"

// MenuType 菜单类型枚举
type MenuType int

const (
	MenuTypeDirectory MenuType = 1 // 目录
	MenuTypeMenu      MenuType = 2 // 菜单
	MenuTypeButton    MenuType = 3 // 按钮
	MenuTypeIframe    MenuType = 4 // 内嵌 iframe
)

// MenuMeta 菜单元数据
// 用于存储菜单的元信息，如标题、图标等
type MenuMeta struct {
	Title        string // 菜单标题
	Icon         string // 菜单图标
	KeepAlive    bool   // 是否缓存
	DefaultMenu  bool   // 是否默认菜单
	Breadcrumb   bool   // 是否在面包屑显示
	ActiveMenu   string // 当前激活的菜单
	Affix        bool   // 是否固定在标签栏
	FrameSrc     string // iframe 地址（如果是 iframe 类型）
	FrameLoading bool   // iframe 加载动画
}

// Menu 菜单领域实体
// 对应前端路由配置，用于构建前端动态路由
// 包含菜单的基本信息、路由信息和权限信息
// 纯业务模型，无 GORM 标签
type Menu struct {
	ID           int64      // 菜单 ID
	ParentID     int64      // 父菜单 ID，0 表示顶级菜单
	Path         string     // 路由路径
	Name         string     // 路由名称，前端组件名
	Component    string     // 组件路径
	Title        string     // 菜单标题
	Icon         string     // 菜单图标
	Hidden       bool       // 是否隐藏
	Sort         int        // 排序号
	Status       int        // 状态：0 禁用，1 启用
	IsExt        bool       // 是否外部链接
	Redirect     string     // 重定向地址
	Permission   string     // 权限标识
	KeepAlive    bool       // 是否缓存
	DefaultMenu  bool       // 是否默认菜单
	Breadcrumb   bool       // 是否在面包屑显示
	ActiveMenu   string     // 当前激活的菜单
	Affix        bool       // 是否固定在标签栏
	Type         MenuType   // 菜单类型：1 目录，2 菜单，3 按钮，4iframe
	FrameSrc     string     // iframe 地址（如果是 iframe 类型）
	FrameLoading bool       // iframe 加载动画
	Meta         MenuMeta   // 菜单元数据
	Children     []*Menu    // 子菜单
	Parent       *Menu      // 父菜单（新增）
	BtnPermIds   []int64    // 菜单下的按钮权限 ID 列表
	BtnPerms     []*BtnPerm // 菜单下的按钮权限列表（新增）
	RoleIds      []int64    // 关联的角色 ID 列表
	Roles        []*Role    // 关联的角色列表（新增）
	APIRouteIds  []int64    // 关联的 API 路由 ID 列表
	APIRoutes    []*API     // 关联的 API 路由列表（新增）
	CreatedBy    int64      // 创建人 ID
	CreatedAt    time.Time  // 创建时间
	UpdatedBy    int64      // 更新人 ID
	UpdatedAt    time.Time  // 更新时间
}

// IsActive 检查菜单是否启用
func (m *Menu) IsActive() bool {
	return m.Status == 1
}

// IsDirectory 检查菜单是否是目录类型
func (m *Menu) IsDirectory() bool {
	return m.Type == MenuTypeDirectory
}

// IsMenu 检查菜单是否是菜单类型
func (m *Menu) IsMenu() bool {
	return m.Type == MenuTypeMenu
}

// IsButton 检查菜单是否是按钮类型
func (m *Menu) IsButton() bool {
	return m.Type == MenuTypeButton
}

// IsIframe 检查菜单是否是 iframe 类型
func (m *Menu) IsIframe() bool {
	return m.Type == MenuTypeIframe
}

// HasPermission 检查菜单是否有指定权限
func (m *Menu) HasPermission(permission string) bool {
	return m.Permission == permission
}

// HasChild 检查菜单是否有子菜单
func (m *Menu) HasChild() bool {
	return len(m.Children) > 0
}
