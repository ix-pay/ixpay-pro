import { asyncRouterHandle } from '@/utils/async-router'
import { emitter } from '@/utils/bus'
import { getMenuList } from '@/api/modules/menu'
import { defineStore } from 'pinia'
import { ref } from 'vue'
import pathInfo from '@/path-info.json'
import type { RouteRecordRaw } from 'vue-router'
import type { ApiMenuItem } from '@/types/menu'
import { MenuType } from '@/types/menu'

// 定义路由元信息接口
export interface RouteMeta {
  title: string
  hidden?: boolean
  keepAlive?: boolean
  closeTab?: boolean
  btns?: string[]
  defaultMenu?: boolean
  path?: string
  transitionType?: string
  icon?: string
}

// 定义扩展的路由记录接口
export interface ExtendedRouteRecordRaw
  extends Omit<RouteRecordRaw, 'meta' | 'children' | 'component'> {
  id?: string | number
  parent?: ExtendedRouteRecordRaw | null
  hidden?: boolean
  meta: RouteMeta
  component?: string | (() => Promise<{ default: import('vue').Component }>)
  children?: ExtendedRouteRecordRaw[]
  btns?: string[]
  redirect?: RouteRecordRaw['redirect']
}

// 定义历史记录项接口
export interface HistoryItem {
  name: string
}

export const useRouterStore = defineStore('router', () => {
  const keepAliveRouters = ref<string[]>([])
  const asyncRouterFlag = ref<number>(0)
  const dynamicRoutesLoaded = ref<boolean>(false)

  const setAsyncRouterFlag = (value: number) => {
    asyncRouterFlag.value = value
  }
  const setDynamicRoutesLoaded = (value: boolean) => {
    dynamicRoutesLoaded.value = value
  }
  const routeMap: Record<string, ExtendedRouteRecordRaw> = {}

  // 存储按钮权限列表
  const buttonPermissions = ref<string[]>([])

  // 从菜单数据中提取按钮权限（type: 3 的菜单项）
  const extractButtonPermissions = (menus: ApiMenuItem[]): string[] => {
    const buttons: string[] = []

    const traverse = (items: ApiMenuItem[]) => {
      for (const item of items) {
        if (item.type === MenuType.BUTTON && item.permission) {
          buttons.push(item.permission)
        }
        if (item.children && item.children.length > 0) {
          traverse(item.children)
        }
      }
    }

    traverse(menus)
    return buttons
  }

  // 存储keep-alive相关的信息
  const keepAliveInfo = ref<{ keepAliveRouters: string[]; nameMap: Record<string, string> }>({
    keepAliveRouters: [],
    nameMap: {},
  })

  const setKeepAliveRouters = (history: HistoryItem[]) => {
    const keepArrTemp: string[] = []

    keepArrTemp.push(...keepAliveInfo.value.keepAliveRouters)

    history.forEach((item) => {
      const routeInfo = routeMap[item.name]
      if (routeInfo && routeInfo.meta && routeInfo.meta.path) {
        const componentName = (pathInfo as Record<string, string>)[routeInfo.meta.path]
        if (componentName) {
          keepArrTemp.push(componentName)
        }
      }

      if (keepAliveInfo.value.nameMap[item.name]) {
        keepArrTemp.push(keepAliveInfo.value.nameMap[item.name])
      }
    })

    keepAliveRouters.value = Array.from(new Set(keepArrTemp))
  }

  const removeKeepAliveRouter = (path: string) => {
    const componentName = (pathInfo as Record<string, string>)[path]
    if (componentName) {
      const index = keepAliveRouters.value.indexOf(componentName)
      if (index > -1) {
        keepAliveRouters.value.splice(index, 1)
      }
    }
  }

  const clearAllKeepAlive = () => {
    keepAliveRouters.value = []
    keepAliveInfo.value = {
      keepAliveRouters: [],
      nameMap: {},
    }
  }

  const resetTabManager = () => {
    emitter.emit('resetTabManager')
  }

  emitter.on('setKeepAlive', setKeepAliveRouters)

  const asyncRouters = ref<ExtendedRouteRecordRaw[]>([])

  // 为路由添加meta信息
  const enrichRouteMeta = (routes: ExtendedRouteRecordRaw[]) => {
    routes.forEach((item) => {
      if (!item.meta) {
        item.meta = { title: '' }
      }
      item.meta.title = (item as { title?: string }).title || item.meta.title || ''
      item.meta.icon = (item as { icon?: string }).icon || item.meta.icon || ''
      item.meta.hidden = item.hidden || false

      if (item.component && typeof item.component === 'string') {
        item.meta.path = '/src/' + item.component
        if (!item.component.startsWith('views/') && !item.component.startsWith('plugin/')) {
          if (item.component.startsWith('base/')) {
            item.component = `views/base/${item.component.replace('base/', '')}`
          } else {
            item.component = `views/${item.component}`
          }
        }
      }

      if (typeof item.name === 'string') {
        routeMap[item.name] = item
      }

      if (item.children && item.children.length > 0) {
        enrichRouteMeta(item.children)
      }
    })
  }

  // 过滤有效路由：type:1必须有子路由，type:2必须有组件，type:3直接移除
  const filterValidRoutes = (routes: ExtendedRouteRecordRaw[]): ExtendedRouteRecordRaw[] => {
    return routes.filter((route) => {
      if (!route.name) return false

      const menuItem = route as ApiMenuItem

      // type:3 按钮权限 - 直接移除
      if (menuItem.type === MenuType.BUTTON) return false

      // type:1 目录 - 必须有子路由
      if (menuItem.type === MenuType.GROUP) {
        if (!route.children || route.children.length === 0) return false
        route.children = filterValidRoutes(route.children)
        return route.children.length > 0
      }

      // type:2 页面 - 必须有组件
      if (menuItem.type === MenuType.MENU) {
        if (!route.component) return false
        if (route.children) {
          route.children = filterValidRoutes(route.children)
        }
        return true
      }

      // 其他类型：有组件或子路由
      if (!route.component && (!route.children || route.children.length === 0)) return false
      if (route.children) {
        route.children = filterValidRoutes(route.children)
      }
      return true
    })
  }

  // 标准化路径
  const normalizePath = (routes: ExtendedRouteRecordRaw[]) => {
    routes.forEach((route) => {
      if (route.path && route.path.startsWith('/')) {
        route.path = route.path.slice(1)
      }
      if (route.children) {
        route.path = route.path || ''
        normalizePath(route.children)
      }
    })
  }

  // 清除所有路由缓存数据
  const clearAllRouterCache = () => {
    asyncRouters.value = []
    buttonPermissions.value = []
    clearAllKeepAlive()
    resetTabManager()
    // 清空 routeMap
    Object.keys(routeMap).forEach((key) => delete routeMap[key])
    // 清空 localStorage 中的持久化缓存
    localStorage.removeItem('pinia_router')
    // 清空 TabManager 的本地缓存（组件可能未挂载，事件方式不可靠）
    localStorage.removeItem('tabManagerTabs')
    localStorage.removeItem('tabManagerActiveTab')
  }

  // 从后台获取动态路由
  const SetAsyncRouter = async (): Promise<ExtendedRouteRecordRaw[]> => {
    asyncRouterFlag.value++

    // 每次获取菜单数据前，先清除所有旧缓存，避免因接口数据变化导致残留
    clearAllRouterCache()

    try {
      const asyncRouterRes = await getMenuList()

      if (!asyncRouterRes || !asyncRouterRes.data || !Array.isArray(asyncRouterRes.data)) {
        return []
      }

      let dynamicRoutes: ExtendedRouteRecordRaw[] = asyncRouterRes.data.filter(
        (route) => route && typeof route === 'object',
      )

      // 过滤掉前端已定义的路由（profile、settings）
      const predefinedRouteNames = ['UserProfile', 'SystemSetting']
      dynamicRoutes = dynamicRoutes.filter((route) => {
        if (route.name && predefinedRouteNames.includes(route.name.toString())) {
          return false
        }
        return true
      })

      // 提取按钮权限（type:3）
      buttonPermissions.value = extractButtonPermissions(dynamicRoutes as ApiMenuItem[])

      // 添加meta信息
      enrichRouteMeta(dynamicRoutes)

      // 处理组件加载
      asyncRouterHandle(dynamicRoutes)

      // 过滤有效路由
      const validRoutes = filterValidRoutes(dynamicRoutes)

      // 标准化路径
      normalizePath(validRoutes)

      // 存储到asyncRouters（供侧边栏渲染）
      asyncRouters.value = validRoutes

      return validRoutes
    } catch (error) {
      console.error('设置异步路由失败:', error)
      return []
    }
  }

  // 从 asyncRouters 中查找第一个可用的页面路径
  const getFirstPagePath = (): string | null => {
    const findFirst = (menus: ExtendedRouteRecordRaw[], parentPath = ''): string | null => {
      for (const menu of menus) {
        const menuItem = menu as any
        const currentPath = parentPath + (menu.path || '')
        if (menuItem.type === 2 && menu.path) {
          return `/${currentPath}`
        }
        if (menu.children && menu.children.length > 0) {
          const childResult = findFirst(menu.children, currentPath + '/')
          if (childResult) return childResult
        }
      }
      return null
    }
    return findFirst(asyncRouters.value)
  }

  return {
    asyncRouters,
    keepAliveRouters,
    asyncRouterFlag,
    setAsyncRouterFlag,
    dynamicRoutesLoaded,
    setDynamicRoutesLoaded,
    SetAsyncRouter,
    getFirstPagePath,
    clearAllRouterCache,
    routeMap,
    removeKeepAliveRouter,
    clearAllKeepAlive,
    resetTabManager,
    buttonPermissions,
  }
})