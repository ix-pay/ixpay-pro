// 租户 UI 配置全局状态（模块级 reactive + localStorage 持久化）
// 数据来源：管理后台「渠道配置 → UI配置」按租户写入，wxapp 启动经 /api/wx/config/app 按 app_key 加载
// Schema：
//   brand:   { title, logo_url, primary_color }
//   home:    { notice, banners: [{ image_url, link? }], banner_url(旧版单张，兼容), grid: [{ icon, title, path, visible? }] }
//   tabbar:  { items: [{ icon, title, path }] }（底部导航，默认 首页/个人中心）
//   profile: { menu: [{ title, path }] }（个人中心菜单，默认 待支付列表/订单支付）
import { reactive } from 'vue'

/** 轮播图项 */
export interface UIBanner {
  image_url: string
  link?: string
}

/** 通用菜单项（首页宫格/tab栏/个人中心菜单共用） */
export interface UIMenuItem {
  icon?: string
  title: string
  path: string
  visible?: boolean
}

/** 租户 UI 配置结构 */
export interface TenantUI {
  brand?: { title?: string; logo_url?: string; primary_color?: string }
  home?: {
    notice?: string
    banners?: UIBanner[]
    banner_url?: string
    grid?: UIMenuItem[]
  }
  tabbar?: { items?: UIMenuItem[] }
  profile?: { menu?: UIMenuItem[] }
}

const STORAGE_KEY = 'wx_tenant_ui'

function loadFromStorage(): TenantUI | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? (JSON.parse(raw) as TenantUI) : null
  } catch {
    return null
  }
}

export const uiState = reactive<{ ui: TenantUI | null }>({ ui: loadFromStorage() })

/** 写入租户 UI 配置（首页加载公开配置后调用；null 时清除） */
export function setTenantUI(ui: unknown) {
  uiState.ui = (ui && typeof ui === 'object' ? (ui as TenantUI) : null)
  try {
    if (uiState.ui) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(uiState.ui))
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  } catch {
    // 存储失败不阻断渲染（内存态仍可用）
  }
}

/** 底部 tab 默认项（未配置时生效） */
export const defaultTabbar: UIMenuItem[] = [
  { icon: '🏠', title: '首页', path: '/' },
  { icon: '👤', title: '个人中心', path: '/profile' },
]

/** 底部 tab 配置（空配置走默认） */
export function getTabbarItems(): UIMenuItem[] {
  const items = uiState.ui?.tabbar?.items
  if (Array.isArray(items) && items.length > 0) {
    return items.filter((t) => t && t.path && t.title)
  }
  return defaultTabbar
}

/** 个人中心菜单（空配置走默认） */
export function getProfileMenu(): UIMenuItem[] {
  const menu = uiState.ui?.profile?.menu
  if (Array.isArray(menu) && menu.length > 0) {
    return menu.filter((m) => m && m.path && m.title)
  }
  return [
    { title: '待支付列表', path: '/pending-list' },
    { title: '订单支付', path: '/payment' },
  ]
}

/** 首页宫格菜单（空配置走默认；visible=false 过滤） */
export function getHomeGrid(): UIMenuItem[] {
  const grid = uiState.ui?.home?.grid
  if (Array.isArray(grid) && grid.length > 0) {
    return grid.filter((g) => g && g.visible !== false)
  }
  return [
    { icon: '📋', title: '待支付列表', path: '/pending-list' },
    { icon: '💳', title: '订单支付', path: '/payment' },
    { icon: '👤', title: '个人中心', path: '/profile' },
  ]
}
