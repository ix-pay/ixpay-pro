import { asyncRouterHandle } from '@/utils/asyncRouter'
import { emitter } from '@/utils/bus'
import { asyncMenu } from '@/api/modules/menu'
import { defineStore } from 'pinia'
import { ref, watchEffect } from 'vue'
import pathInfo from '@/pathInfo.json'
import { useRoute } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { hasMenuPermission } from '@/utils/permission'
import type { ApiMenuItem } from '@/types/menu'

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

// 递归过滤路由，只保留用户有权限的路由
const filterRoutesByPermission = (routes: ExtendedRouteRecordRaw[]): ExtendedRouteRecordRaw[] => {
  return routes.filter((route) => {
    // 检查当前路由是否有权限
    // 由于 route 是 ExtendedRouteRecordRaw 类型，我们需要创建一个兼容的对象
    const menuForPermission = {
      permission: route.btns?.[0] || '',
    } as unknown as ApiMenuItem

    const hasPermission = hasMenuPermission(menuForPermission)

    // 如果有权限，继续检查子路由
    if (hasPermission && route.children && route.children.length > 0) {
      route.children = filterRoutesByPermission(route.children)
      // 如果子路由过滤后为空，仍然保留父路由
    }

    return hasPermission
  })
}

const formatRouter = (
  routes: ExtendedRouteRecordRaw[],
  routeMap: Record<string, ExtendedRouteRecordRaw>,
  parent: ExtendedRouteRecordRaw | null,
  notLayoutRouterArr: ExtendedRouteRecordRaw[],
) => {
  routes.forEach((item) => {
    item.parent = parent
    // 确保 meta 对象存在
    if (!item.meta) {
      item.meta = { title: '' }
    }
    // 将 title 字段合并到 meta.title 中（优先使用 item.title）
    item.meta.title = (item as { title?: string }).title || item.meta.title || ''
    // 将 icon 字段合并到 meta.icon 中（优先使用 item.icon）
    item.meta.icon = (item as { icon?: string }).icon || item.meta.icon || ''
    item.meta.btns = item.btns
    item.meta.hidden = item.hidden || false

    // 保持 component 为字符串格式，由 asyncRouterHandle 统一处理
    if (item.component && typeof item.component === 'string') {
      console.log('formatRouter - Original component path:', item.component)
      // 确保 component 路径以 views/或 plugin/开头，asyncRouterHandle 会处理
      if (!item.component.startsWith('views/') && !item.component.startsWith('plugin/')) {
        // 如果是基础组件（如 base/index.vue），添加 base 前缀
        if (item.component.startsWith('base/')) {
          item.component = `views/base/${item.component.replace('base/', '')}`
        } else {
          // 其他情况，添加 view 前缀
          item.component = `views/${item.component}`
        }
        console.log('formatRouter - Updated component path:', item.component)
      }
    }

    if (item.meta.defaultMenu === true && !parent) {
      // 避免将已经以/开头的路径再次添加/
      const newPath = item.path.startsWith('/') ? item.path : `/${item.path}`
      item = { ...item, path: newPath }
      notLayoutRouterArr.push(item)
    }

    if (typeof item.name === 'string') {
      routeMap[item.name] = item
    }

    if (item.children && item.children.length > 0) {
      formatRouter(item.children, routeMap, item, notLayoutRouterArr)
    }
  })
}

export const useRouterStore = defineStore('router', () => {
  const keepAliveRouters = ref<string[]>([])
  // 每次初始化时重置为0，确保刷新页面时能重新加载动态路由
  const asyncRouterFlag = ref<number>(0)

  const setAsyncRouterFlag = (value: number) => {
    asyncRouterFlag.value = value
  }
  const routeMap: Record<string, ExtendedRouteRecordRaw> = {}

  // 存储keep-alive相关的信息
  const keepAliveInfo = ref<{ keepAliveRouters: string[]; nameMap: Record<string, string> }>({
    keepAliveRouters: [],
    nameMap: {},
  })

  const setKeepAliveRouters = (history: HistoryItem[]) => {
    const keepArrTemp: string[] = []

    // 1. 首先添加原有的 keepAlive 配置
    keepArrTemp.push(...keepAliveInfo.value.keepAliveRouters)

    history.forEach((item) => {
      // 2. 为所有 history 中的路由强制启用 keep-alive
      // 通过 routeMap 获取路由信息，然后通过 pathInfo 获取组件名
      const routeInfo = routeMap[item.name]
      if (routeInfo && routeInfo.meta && routeInfo.meta.path) {
        const componentName = (pathInfo as Record<string, string>)[routeInfo.meta.path]
        if (componentName) {
          keepArrTemp.push(componentName)
        }
      }

      // 3. 如果子路由在 tabs 中打开，父路由也需要 keepAlive
      if (keepAliveInfo.value.nameMap[item.name]) {
        keepArrTemp.push(keepAliveInfo.value.nameMap[item.name])
      }
    })

    keepAliveRouters.value = Array.from(new Set(keepArrTemp))
  }

  // 移除指定路由的 keep-alive 缓存
  const removeKeepAliveRouter = (path: string) => {
    const componentName = (pathInfo as Record<string, string>)[path]
    if (componentName) {
      const index = keepAliveRouters.value.indexOf(componentName)
      if (index > -1) {
        keepAliveRouters.value.splice(index, 1)
      }
    }
  }

  const route = useRoute()

  emitter.on('setKeepAlive', setKeepAliveRouters)

  const asyncRouters = ref<ExtendedRouteRecordRaw[]>([])
  const topMenu = ref<ExtendedRouteRecordRaw[]>([])
  const leftMenu = ref<ExtendedRouteRecordRaw[]>([])
  const menuMap: Record<string, ExtendedRouteRecordRaw> = {}
  const topActive = ref<string>('')

  const setLeftMenu = (name: string) => {
    sessionStorage.setItem('topActive', name)
    topActive.value = name
    leftMenu.value = menuMap[name]?.children || []
  }

  const findTopActive = (
    menuMap: Record<string, ExtendedRouteRecordRaw>,
    routeName: string | undefined,
  ): string | null => {
    if (!routeName) return null

    for (const topName in menuMap) {
      const topItem = menuMap[topName]
      if (topItem.children?.some((item) => item.name === routeName)) {
        return topName
      }
      const foundName = findTopActive(
        topItem.children?.reduce(
          (acc, cur) => {
            if (typeof cur.name === 'string') acc[cur.name] = cur
            return acc
          },
          {} as Record<string, ExtendedRouteRecordRaw>,
        ) || {},
        routeName,
      )
      if (foundName) {
        return topName
      }
    }
    return null
  }

  watchEffect(() => {
    console.log('Router WatchEffect - Start')
    console.log('Router WatchEffect - Current route:', route)

    // 登录页面不处理菜单逻辑，直接返回
    if (route.path === '/login') {
      console.log('Router WatchEffect - Login page, skipping menu processing')
      console.log('Router WatchEffect - End')
      return
    }

    // 只在asyncRouters有有效内容时才处理菜单
    if (!asyncRouters.value || asyncRouters.value.length === 0) {
      console.log('Router WatchEffect - No valid routes, skipping menu processing')
      console.log('Router WatchEffect - End')
      return
    }

    // 过滤掉无效路由和非顶层菜单路由（顶层菜单应该没有parent）
    const validTopRoutes = asyncRouters.value.filter(
      (route) => route && typeof route === 'object' && route.name && !route.parent, // 顶层菜单没有parent
    )

    if (validTopRoutes.length === 0) {
      console.log('Router WatchEffect - No valid top routes, skipping menu processing')
      console.log('Router WatchEffect - End')
      return
    }

    console.log('Router WatchEffect - Processing valid top routes:', validTopRoutes.length)

    // 初始化菜单内容
    topMenu.value = []
    Object.keys(menuMap).forEach((key) => delete menuMap[key]) // 重置menuMap

    // 处理顶层菜单
    validTopRoutes.forEach((item) => {
      if (item && !item.hidden && item.name) {
        const routeName = typeof item.name === 'string' ? item.name : item.name.toString()
        menuMap[routeName] = item
        topMenu.value.push({ ...item, children: [] })
      }
    })

    console.log('Router WatchEffect - Processed topMenu:', topMenu.value.length, 'items')

    // 处理当前激活的菜单
    let topActive: string | null = sessionStorage.getItem('topActive')
    if (!topActive || topActive === 'undefined' || topActive === 'null') {
      topActive = findTopActive(menuMap, typeof route.name === 'string' ? route.name : undefined)
    }

    setLeftMenu(topActive || '')
    console.log('Router WatchEffect - End')
  })

  // 从后台获取动态路由：仅处理登录后的页面路由信息
  // 基础路由已在 router/index.ts 中定义为 baseRoutes
  const SetAsyncRouter = async (): Promise<ExtendedRouteRecordRaw[]> => {
    console.log('SetAsyncRouter - Start (Only login after routes)')

    // 使用局部变量存储非布局路由，避免模块级变量积累
    const notLayoutRouterArr: ExtendedRouteRecordRaw[] = []

    // 增加asyncRouterFlag，触发watchEffect更新
    // 每次成功加载路由后都递增flag，确保watchEffect能检测到变化
    asyncRouterFlag.value++
    console.log('SetAsyncRouter - Updated asyncRouterFlag:', asyncRouterFlag.value)

    try {
      // 调用asyncMenu函数获取菜单数据（仅包含登录后的页面路由）
      console.log('SetAsyncRouter - Getting menu data from API...')
      const asyncRouterRes = await asyncMenu()
      console.log('SetAsyncRouter - Menu API response:', asyncRouterRes)

      // 检查菜单数据是否有效
      if (!asyncRouterRes || !asyncRouterRes.data) {
        console.error('SetAsyncRouter - Invalid menu data from API:', asyncRouterRes)
        // 如果数据无效，返回空数组
        asyncRouters.value = []
        return []
      }

      // 由于 asyncMenu 函数已经处理了响应格式，直接使用返回的菜单数组
      let dynamicRoutes: ExtendedRouteRecordRaw[] = []
      if (Array.isArray(asyncRouterRes?.data)) {
        dynamicRoutes = asyncRouterRes.data
      } else {
        console.error('SetAsyncRouter - Unexpected data format from menu API:', asyncRouterRes.data)
        // 如果数据格式不符合预期，使用空数组
        dynamicRoutes = []
      }

      // 过滤掉 null 元素和无效对象
      dynamicRoutes = dynamicRoutes.filter((route) => route && typeof route === 'object')
      console.log('SetAsyncRouter - Filtered dynamic routes:', dynamicRoutes)

      // 过滤掉已在前端 router/index.ts 中定义的路由，避免重复
      // 这些路由包括：profile（个人资料）、settings（系统设置）
      // 注意：index（首页）路由已从前端移除，现在从后端加载
      const predefinedRouteNames = ['UserProfile', 'SystemSetting']
      dynamicRoutes = dynamicRoutes.filter((route) => {
        if (route.name && predefinedRouteNames.includes(route.name.toString())) {
          console.log('SetAsyncRouter - Filtering out predefined route:', route.name)
          return false
        }
        return true
      })
      console.log('SetAsyncRouter - After filtering predefined routes:', dynamicRoutes.length)

      // 根据用户权限过滤路由
      dynamicRoutes = filterRoutesByPermission(dynamicRoutes)
      console.log('SetAsyncRouter - Permission filtered routes:', dynamicRoutes)

      console.log('SetAsyncRouter - Menu API data count:', dynamicRoutes.length)

      // 格式化路由
      console.log('SetAsyncRouter - Formatting dynamic routes')
      console.log('SetAsyncRouter - Before format - routes count:', dynamicRoutes.length)

      try {
        formatRouter(dynamicRoutes, routeMap, null, notLayoutRouterArr)
        console.log('SetAsyncRouter - After format - routes count:', dynamicRoutes.length)
      } catch (error) {
        console.error('SetAsyncRouter - Error formatting routes:', error)
        console.error('SetAsyncRouter - Error details:', JSON.stringify(error, null, 2))
        throw error
      }

      // 确保路由路径是相对路径（相对于layout）
      const normalizedRoutes = dynamicRoutes.map((route) => {
        const normalizedRoute = { ...route }
        // 如果路径是绝对路径，将其转换为相对路径
        if (normalizedRoute.path && normalizedRoute.path.startsWith('/')) {
          console.log(
            'SetAsyncRouter - Converting absolute path to relative:',
            normalizedRoute.path,
            '->',
            normalizedRoute.path.slice(1),
          )
          normalizedRoute.path = normalizedRoute.path.slice(1)
        }
        return normalizedRoute
      })
      console.log('SetAsyncRouter - Normalized routes:', normalizedRoutes)

      // 更新父属性引用
      normalizedRoutes.forEach((route) => {
        if (route.children) {
          route.children.forEach((child) => {
            child.parent = route
          })
        }
      })

      // 处理非布局路由（如果有）
      const finalRoutes: ExtendedRouteRecordRaw[] = [...normalizedRoutes]
      if (notLayoutRouterArr.length > 0) {
        const validNotLayoutRoutes = notLayoutRouterArr.filter(
          (route) => route && typeof route === 'object',
        )
        if (validNotLayoutRoutes.length > 0) {
          console.log('SetAsyncRouter - Adding non-layout routes:', validNotLayoutRoutes)
          finalRoutes.push(...validNotLayoutRoutes)
        }
      }

      // 处理路由组件
      asyncRouterHandle(finalRoutes)

      // 过滤掉无效路由（没有组件或没有 name 的路由）
      const validFinalRoutes = finalRoutes.filter((route) => {
        // 确保路由有名称
        if (!route.name) {
          console.log('SetAsyncRouter - Filtering out route without name:', route.path)
          return false
        }

        // 确保有组件或有子路由（父路由可以没有组件但必须有子路由）
        if (!route.component && (!route.children || route.children.length === 0)) {
          console.log(
            'SetAsyncRouter - Filtering out route without component or children:',
            route.path,
          )
          return false
        }

        // 如果有子路由，递归检查子路由
        if (route.children && route.children.length > 0) {
          const validChildren = route.children.filter((child) => {
            // 子路由必须有名称和组件
            return child.name && (child.component || (child.children && child.children.length > 0))
          })

          // 更新子路由数组
          route.children = validChildren

          // 如果父路由没有组件且没有有效子路由，过滤掉
          if (!route.component && validChildren.length === 0) {
            console.log(
              'SetAsyncRouter - Filtering out parent route without valid children:',
              route.path,
            )
            return false
          }
        }

        return true
      })

      // 存储处理后的有效路由
      asyncRouters.value = validFinalRoutes
      console.log('SetAsyncRouter - Final valid routes stored:', asyncRouters.value)
      console.log('SetAsyncRouter - Valid routes count:', asyncRouters.value.length)

      // 返回处理后的动态路由
      return validFinalRoutes
    } catch (error) {
      console.error('SetAsyncRouter - Error:', error)
      return [] // 返回空数组而不是false，保持类型一致性
    }
  }

  return {
    topActive,
    setLeftMenu,
    topMenu,
    leftMenu,
    asyncRouters,
    keepAliveRouters,
    asyncRouterFlag,
    setAsyncRouterFlag,
    SetAsyncRouter,
    routeMap,
    removeKeepAliveRouter,
  }
})
