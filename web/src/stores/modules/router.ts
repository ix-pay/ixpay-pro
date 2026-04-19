import { asyncRouterHandle } from '@/utils/async-router'
import { emitter } from '@/utils/bus'
import { getMenuList } from '@/api/modules/menu'
import { defineStore } from 'pinia'
import { ref, watchEffect } from 'vue'
import pathInfo from '@/path-info.json'
import { useRoute } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { hasMenuPermission } from '@/utils/permission'
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

// 递归过滤路由，只保留用户有权限的路由
const filterRoutesByPermission = (routes: ExtendedRouteRecordRaw[]): ExtendedRouteRecordRaw[] => {
  return routes.filter((route) => {
    // 过滤掉 type: 3 的按钮数据（按钮不应该出现在菜单中）
    if ((route as ApiMenuItem).type === 3) {
      return false
    }

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
    // 过滤掉 type: 3 的按钮数据（按钮不应该出现在菜单中）
    if ((item as ApiMenuItem).type === 3) {
      console.log(
        'formatRouter - Skipping button type menu item:',
        item.name,
        (item as ApiMenuItem).title,
      )
      return
    }

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
      console.log('格式化路由 - 原始组件路径:', item.component)
      // 确保 component 路径以 views/或 plugin/开头，asyncRouterHandle 会处理
      if (!item.component.startsWith('views/') && !item.component.startsWith('plugin/')) {
        // 如果是基础组件（如 base/index.vue），添加 base 前缀
        if (item.component.startsWith('base/')) {
          item.component = `views/base/${item.component.replace('base/', '')}`
        } else {
          // 其他情况，添加 view 前缀
          item.component = `views/${item.component}`
        }
        console.log('格式化路由 - 更新后的组件路径:', item.component)
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

    // 递归处理子路由前，先过滤掉 type: 3 的按钮数据
    if (item.children && item.children.length > 0) {
      // 过滤掉 type: 3 的按钮
      item.children = item.children.filter((child) => (child as ApiMenuItem).type !== 3)

      if (item.children.length > 0) {
        formatRouter(item.children, routeMap, item, notLayoutRouterArr)
      }
    }
  })
}

export const useRouterStore = defineStore('router', () => {
  const keepAliveRouters = ref<string[]>([])
  // 每次初始化时重置为 0，确保刷新页面时能重新加载动态路由
  const asyncRouterFlag = ref<number>(0)

  const setAsyncRouterFlag = (value: number) => {
    asyncRouterFlag.value = value
  }
  const routeMap: Record<string, ExtendedRouteRecordRaw> = {}

  // 存储按钮权限列表
  const buttonPermissions = ref<string[]>([])

  // 从菜单数据中提取按钮权限（type: 3 的菜单项）
  const extractButtonPermissions = (menus: ApiMenuItem[]): string[] => {
    const buttons: string[] = []

    const traverse = (items: ApiMenuItem[]) => {
      for (const item of items) {
        // 提取 type: 3 的按钮权限
        if (item.type === MenuType.BUTTON && item.permission) {
          buttons.push(item.permission)
        }
        // 递归处理子菜单
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

  // 清理所有 keep-alive 缓存
  const clearAllKeepAlive = () => {
    keepAliveRouters.value = []
    keepAliveInfo.value = {
      keepAliveRouters: [],
      nameMap: {},
    }
  }

  // 通过事件通知 TabManager 重置
  const resetTabManager = () => {
    emitter.emit('resetTabManager')
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
    console.log('路由监听 - 开始')
    console.log('路由监听 - 当前路由:', route)

    // 登录页面不处理菜单逻辑，直接返回
    if (route.path === '/login') {
      console.log('路由监听 - 登录页面，跳过菜单处理')
      console.log('路由监听 - 结束')
      return
    }

    // 只在 asyncRouters 有有效内容时才处理菜单
    if (!asyncRouters.value || asyncRouters.value.length === 0) {
      console.log('路由监听 - 没有有效路由，跳过菜单处理')
      console.log('路由监听 - 结束')
      return
    }

    // 过滤掉无效路由和非顶层菜单路由（顶层菜单应该没有 parent）
    const validTopRoutes = asyncRouters.value.filter(
      (route) => route && typeof route === 'object' && route.name && !route.parent, // 顶层菜单没有 parent
    )

    if (validTopRoutes.length === 0) {
      console.log('路由监听 - 没有有效顶层路由，跳过菜单处理')
      console.log('路由监听 - 结束')
      return
    }

    console.log('路由监听 - 处理有效顶层路由:', validTopRoutes.length)

    // 初始化菜单内容
    topMenu.value = []
    Object.keys(menuMap).forEach((key) => delete menuMap[key]) // 重置 menuMap

    // 处理顶层菜单
    validTopRoutes.forEach((item) => {
      if (item && !item.hidden && item.name) {
        const routeName = typeof item.name === 'string' ? item.name : item.name.toString()
        menuMap[routeName] = item
        topMenu.value.push({ ...item, children: [] })
      }
    })

    console.log('路由监听 - 处理顶层菜单:', topMenu.value.length, '项')

    // 处理当前激活的菜单
    let topActive: string | null = sessionStorage.getItem('topActive')
    if (!topActive || topActive === 'undefined' || topActive === 'null') {
      topActive = findTopActive(menuMap, typeof route.name === 'string' ? route.name : undefined)
    }

    setLeftMenu(topActive || '')
    console.log('路由监听 - 结束')
  })

  // 从后台获取动态路由：仅处理登录后的页面路由信息
  // 基础路由已在 router/index.ts 中定义为 baseRoutes
  const SetAsyncRouter = async (): Promise<ExtendedRouteRecordRaw[]> => {
    console.log('设置异步路由 - 开始（仅登录后的路由）')

    // 使用局部变量存储非布局路由，避免模块级变量积累
    const notLayoutRouterArr: ExtendedRouteRecordRaw[] = []

    // 增加 asyncRouterFlag，触发 watchEffect 更新
    // 每次成功加载路由后都递增 flag，确保 watchEffect 能检测到变化
    asyncRouterFlag.value++
    console.log('设置异步路由 - 已更新 asyncRouterFlag:', asyncRouterFlag.value)

    try {
      // 调用 getMenuList 函数获取菜单数据（仅包含登录后的页面路由）
      console.log('设置异步路由 - 正在从 API 获取菜单数据...')
      const asyncRouterRes = await getMenuList()
      console.log('设置异步路由 - 菜单 API 响应:', asyncRouterRes)

      // 检查菜单数据是否有效
      if (!asyncRouterRes || !asyncRouterRes.data) {
        console.error('设置异步路由 - API 返回的菜单数据无效:', asyncRouterRes)
        // 如果数据无效，返回空数组
        asyncRouters.value = []
        return []
      }

      // 直接使用返回的菜单数组
      let dynamicRoutes: ExtendedRouteRecordRaw[] = []
      if (Array.isArray(asyncRouterRes?.data)) {
        dynamicRoutes = asyncRouterRes.data
      } else {
        console.error('设置异步路由 - 菜单 API 返回的数据格式不符合预期:', asyncRouterRes.data)
        // 如果数据格式不符合预期，使用空数组
        dynamicRoutes = []
      }

      // 过滤掉 null 元素和无效对象
      dynamicRoutes = dynamicRoutes.filter((route) => route && typeof route === 'object')
      console.log('设置异步路由 - 过滤后的动态路由:', dynamicRoutes)

      // 过滤掉已在前端 router/index.ts 中定义的路由，避免重复
      // 这些路由包括：profile（个人资料）、settings（系统设置）
      // 注意：index（首页）路由已从前端移除，现在从后端加载
      const predefinedRouteNames = ['UserProfile', 'SystemSetting']
      dynamicRoutes = dynamicRoutes.filter((route) => {
        if (route.name && predefinedRouteNames.includes(route.name.toString())) {
          console.log('设置异步路由 - 过滤掉已定义的路由:', route.name)
          return false
        }
        return true
      })
      console.log('设置异步路由 - 过滤已定义路由后的数量:', dynamicRoutes.length)

      // 先提取按钮权限（在过滤路由之前提取，因为过滤会移除 type: 3 的按钮数据）
      buttonPermissions.value = extractButtonPermissions(dynamicRoutes as ApiMenuItem[])
      console.log('设置异步路由 - 提取的按钮权限:', buttonPermissions.value)

      // 根据用户权限过滤路由
      dynamicRoutes = filterRoutesByPermission(dynamicRoutes)
      console.log('设置异步路由 - 权限过滤后的路由:', dynamicRoutes)

      console.log('设置异步路由 - 菜单 API 数据数量:', dynamicRoutes.length)

      // 格式化路由
      console.log('设置异步路由 - 正在格式化动态路由')
      console.log('设置异步路由 - 格式化前路由数量:', dynamicRoutes.length)
      console.log('设置异步路由 - 格式化前的第一个路由:', JSON.stringify(dynamicRoutes[0], null, 2))

      try {
        formatRouter(dynamicRoutes, routeMap, null, notLayoutRouterArr)
        console.log('设置异步路由 - 格式化后路由数量:', dynamicRoutes.length)
        console.log(
          '设置异步路由 - 格式化后的第一个路由:',
          JSON.stringify(dynamicRoutes[0], null, 2),
        )
      } catch (error) {
        console.error('设置异步路由 - 格式化路由时出错:', error)
        console.error('设置异步路由 - 错误详情:', JSON.stringify(error, null, 2))
        throw error
      }

      // 确保路由路径是相对路径（相对于 layout）
      const normalizedRoutes = dynamicRoutes.map((route) => {
        const normalizedRoute = { ...route }
        // 如果路径是绝对路径，将其转换为相对路径
        if (normalizedRoute.path && normalizedRoute.path.startsWith('/')) {
          console.log(
            '设置异步路由 - 将绝对路径转换为相对路径:',
            normalizedRoute.path,
            '->',
            normalizedRoute.path.slice(1),
          )
          normalizedRoute.path = normalizedRoute.path.slice(1)
        }
        return normalizedRoute
      })
      console.log('设置异步路由 - 标准化后的路由数量:', normalizedRoutes.length)
      console.log('设置异步路由 - 第一个标准化路由:', JSON.stringify(normalizedRoutes[0], null, 2))

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
          console.log('设置异步路由 - 添加非布局路由:', validNotLayoutRoutes.length)
          finalRoutes.push(...validNotLayoutRoutes)
        }
      }

      console.log(
        '设置异步路由 - asyncRouterHandle 前的 finalRoutes 数量:',
        finalRoutes.length,
      )
      console.log(
        '设置异步路由 - asyncRouterHandle 前的第一个路由:',
        JSON.stringify(finalRoutes[0], null, 2),
      )

      // 处理路由组件
      asyncRouterHandle(finalRoutes)

      console.log(
        '设置异步路由 - asyncRouterHandle 后的第一个路由:',
        JSON.stringify(finalRoutes[0], null, 2),
      )
      console.log('设置异步路由 - 最终过滤前检查路由...')
      finalRoutes.forEach((route, index) => {
        console.log(`设置异步路由 - 路由 ${index}:`, {
          name: route.name,
          path: route.path,
          hasComponent: !!route.component,
          componentType: typeof route.component,
          childrenCount: route.children?.length || 0,
        })
      })

      // 过滤掉无效路由（没有组件或没有 name 的路由）
      const validFinalRoutes = finalRoutes.filter((route) => {
        // 确保路由有名称
        if (!route.name) {
          console.log('设置异步路由 - 过滤掉没有名称的路由:', route.path)
          return false
        }

        // 确保有组件或有子路由（父路由可以没有组件但必须有子路由）
        if (!route.component && (!route.children || route.children.length === 0)) {
          console.log(
            '设置异步路由 - 过滤掉没有组件或子路由的路由:',
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
      console.log('设置异步路由 - 最终有效路由已存储:', asyncRouters.value)
      console.log('设置异步路由 - 有效路由数量:', asyncRouters.value.length)

      // 返回处理后的动态路由
      return validFinalRoutes
    } catch (error) {
      console.error('设置异步路由 - 错误:', error)
      return [] // 返回空数组而不是 false，保持类型一致性
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
    clearAllKeepAlive,
    resetTabManager,
    buttonPermissions,
  }
})
