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
  type: number
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

// 菜单删除影响评估
export interface MenuDeleteImpact {
  childMenusCount: number
  btnPermsCount: number
  affectedRolesCount: number
  affectedApisCount: number
  level: 'LOW' | 'MEDIUM' | 'HIGH'
  warning: string
}
