// 路由菜单结构，用于前端路由导航
export interface RouterMenuItem {
  id: string | number
  name: string
  path: string
  component: string | (() => Promise<{ default: import('vue').Component }>)
  meta: {
    title: string
    icon: string
    hidden?: boolean
    keepAlive?: boolean
    closeTab?: boolean
    btns?: string[]
    defaultMenu?: boolean
    transitionType?: string
  }
  hidden?: boolean
  children?: RouterMenuItem[]
}

// 菜单类型枚举
export enum MenuType {
  GROUP = 1, // 不可打开页面的菜单（父级菜单，用于分组）
  MENU = 2, // 可以打开页面的菜单（实际可访问的页面）
  BUTTON = 3, // 按钮权限数据（不是菜单，是页面内的操作按钮）
}

// 后端 API 返回的菜单数据结构（所有字段名统一使用小写，符合 JSON 规范）
export interface ApiMenuItem {
  id: string
  parentId: string // 必须使用 string，防止 int64 精度丢失
  path: string
  name: string
  component: string | (() => Promise<{ default: import('vue').Component }>)
  title: string
  icon: string
  hidden: boolean
  sort: number
  status: number
  isExt: boolean
  redirect: string
  permission: string
  keepAlive: boolean
  defaultMenu: boolean
  breadcrumb: boolean
  activeMenu: string
  affix: boolean
  type: MenuType | number // 1: 目录，2: 菜单，3: 按钮
  frameLoading: boolean
  apiIds?: string[]
  meta?: {
    title: string
    icon: string
    keepAlive: boolean
    defaultMenu: boolean
    breadcrumb: boolean
    activeMenu: string
    affix: boolean
    frameLoading: boolean
  }
  children?: ApiMenuItem[]
}

// 按钮权限数据结构
export interface ButtonPermission {
  id: string
  name: string // 按钮名称，如 "UserAdd"
  title: string // 按钮标题，如 "新增用户"
  permission: string // 权限标识，如 "system:user:add"
  icon: string // 按钮图标
  parentId: string // 所属页面菜单的 ID
}

// 菜单删除影响评估
export interface MenuDeleteImpact {
  childMenusCount: number
  btnPermsCount: number
  affectedRolesCount: number
  affectedApisCount: number
  level: 'LOW' | 'MEDIUM' | 'HIGH'
  warning: string
}
