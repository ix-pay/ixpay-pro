// 角色相关类型定义
// 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题

// 菜单信息（简化版，用于角色权限）
// 注意：所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题
export interface Menu {
  id: string
  parentId: string
  title: string
  path: string
  component?: string
  icon?: string
  sort?: number
  type?: number // 1-目录，2-菜单，3-按钮
  children?: Menu[]
}

// 角色信息
export interface Role {
  id: string // 角色 ID
  name: string // 角色名称
  code: string // 角色编码
  description?: string // 描述
  type?: number // 角色类型：1-系统角色 2-普通角色
  parentId?: string // 父角色 ID
  status?: number // 状态：1-启用 0-禁用
  isSystem?: boolean // 是否系统角色
  sort?: number // 排序
  createdAt?: string // 创建时间
  updatedAt?: string // 更新时间
  // 可选字段（详情接口返回）
  users?: SimpleUserInfo[] // 用户列表
  menus?: Menu[] // 菜单列表
  routes?: ApiRoute[] // API 路由列表
}

// 简单用户信息（用于角色详情中的用户列表）
export interface SimpleUserInfo {
  id: string
  userName: string
  nickname: string
  email?: string
  phone?: string
  avatar?: string
  status?: number
}

// 角色列表响应
export interface RoleListResponse {
  list: Role[]
  total: number
  page: number
  pageSize: number
}

// 角色详情响应
export interface RoleDetailResponse extends Role {
  users?: SimpleUserInfo[]
  menus?: Menu[]
  routes?: ApiRoute[]
}

// 菜单信息（简化版，用于角色权限）
// 注意：所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题
export interface Menu {
  id: string
  parentId: string
  title: string
  path: string
  component?: string
  icon?: string
  sort?: number
  type?: number // 1-目录，2-菜单，3-按钮
  children?: Menu[]
}

// 按钮权限信息
// 注意：所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题
export interface BtnPerm {
  id: string
  menuId: string
  name: string
  code: string
  icon?: string
  sort?: number
  apis?: ApiRoute[]
}

// API 路由信息
// 注意：所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题
export interface ApiRoute {
  id: string
  path: string
  method: string
  group?: string
  description?: string
  status?: number
}

// 角色权限设置请求
// 注意：所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题
export interface RolePermissionRequest {
  menuIds: string[]
  btnPermIds: string[]
  apiRouteIds: string[]
}

// 角色权限设置响应
export interface RolePermissionResponse {
  menus: Menu[]
  btnPerms: BtnPerm[]
  apiRoutes: ApiRoute[]
}

// 删除影响评估
export interface DeleteImpact {
  childMenusCount: number
  btnPermsCount: number
  affectedRolesCount: number
  affectedApisCount: number
  level: 'LOW' | 'MEDIUM' | 'HIGH'
  warning: string
}

// 权限日志
export interface PermissionLog {
  id: number
  operatorId: number
  operatorName: string
  actionType: string // SAVE_ROLE_PERMS, DELETE_MENU, DELETE_BTN_PERM, DELETE_API
  targetType: string // ROLE, MENU, BTN_PERM, API
  targetId: number
  beforeData?: Record<string, unknown>
  afterData?: Record<string, unknown>
  ipAddress?: string
  userAgent?: string
  createdAt: string
}

// 权限日志查询参数
export interface PermissionLogQuery {
  page?: number
  pageSize?: number
  actionType?: string
  targetType?: string
  targetId?: number
  operatorId?: number
}

// 树形节点数据（用于菜单树）
export interface TreeNode {
  id: number
  label: string
  type?: number // 1-目录，2-菜单，3-按钮
  children?: TreeNode[]
  isLeaf?: boolean
  disabled?: boolean
}

// 树形组件勾选事件
export interface TreeCheckEvent {
  checkedKeys: number[]
  halfCheckedKeys: number[]
}
