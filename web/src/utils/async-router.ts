// 静态导入所有页面组件，确保 Keep-Alive 和路由缓存正常工作
const staticComponents: Record<string, () => Promise<{ default: import('vue').Component }>> = {
  // 基础页面
  'views/base/login/index': () => import('@/views/base/login/index.vue'),
  'views/base/profile/index': () => import('@/views/base/profile/index.vue'),
  'views/base/setting/index': () => import('@/views/base/setting/index.vue'),
  'views/base/index/index': () => import('@/views/base/index/index.vue'),

  // 系统管理页面
  'views/system/index': () => import('@/views/system/index.vue'),
  'views/system/user/index': () => import('@/views/system/user/index.vue'),
  'views/system/role/index': () => import('@/views/system/role/index.vue'),
  'views/system/menu/index': () => import('@/views/system/menu/index.vue'),
  'views/system/department/index': () => import('@/views/system/department/index.vue'),
  'views/system/position/index': () => import('@/views/system/position/index.vue'),
  'views/system/dict/index': () => import('@/views/system/dict/index.vue'),
  'views/system/config/index': () => import('@/views/system/config/index.vue'),
  'views/system/notice/index': () => import('@/views/system/notice/index.vue'),
  'views/system/api-route/index': () => import('@/views/system/api-route/index.vue'),

  // 监控页面
  'views/monitor/monitor/index': () => import('@/views/monitor/monitor/index.vue'),
  'views/monitor/online-user/index': () => import('@/views/monitor/online-user/index.vue'),

  // 日志页面
  'views/log/operation-log/index': () => import('@/views/log/operation-log/index.vue'),
  'views/log/login-log/index': () => import('@/views/log/login-log/index.vue'),

  // 任务页面
  'views/task/task/index': () => import('@/views/task/task/index.vue'),
}

const viewModules = import.meta.glob('../views/**/*.vue')
const pluginModules = import.meta.glob('../plugin/**/*.vue')

import type { ExtendedRouteRecordRaw } from '@/stores/modules/router'

export const asyncRouterHandle = (asyncRouter: ExtendedRouteRecordRaw[]) => {
  // 只在开发环境输出详细日志
  if (import.meta.env.DEV) {
    console.log('异步路由处理 - 开始处理路由:', asyncRouter.length)
  }

  asyncRouter.forEach((item) => {
    // 只在开发环境输出详细日志
    if (import.meta.env.DEV) {
      console.log('异步路由处理 - 处理路由:', item.path, item.name)
      console.log('异步路由处理 - 路由组件类型:', typeof item.component)
    }

    // 处理组件加载 - 优先使用静态导入
    if (item.component && typeof item.component === 'string') {
      // 只在开发环境输出详细日志
      if (import.meta.env.DEV) {
        console.log('异步路由处理 - 处理字符串组件:', item.component)
      }

      item.meta.path = '/src/' + item.component

      try {
        let comp: (() => Promise<{ default: import('vue').Component }>) | undefined

        // 方法 1: 从静态导入映射表中查找
        comp = staticComponents[item.component]

        // 方法 2: 如果静态映射表中没有，使用动态查找作为后备
        if (!comp) {
          if (import.meta.env.DEV) {
            console.warn(
              '异步路由处理 - 组件不在静态映射中，使用动态导入:',
              item.component,
            )
          }

          if (item.component.split('/')[0] === 'view') {
            comp = dynamicImport(viewModules, item.component)
          } else if (item.component.split('/')[0] === 'plugin') {
            comp = dynamicImport(pluginModules, item.component)
          } else {
            comp =
              dynamicImport(viewModules, item.component) ||
              dynamicImport(pluginModules, item.component)
          }
        }

        if (comp) {
          item.component = comp
          // 只在开发环境输出详细日志
          if (import.meta.env.DEV) {
            console.log('异步路由处理 - 成功加载组件:', item.component)
          }
        } else {
          // 只在开发环境输出错误日志
          if (import.meta.env.DEV) {
            console.error('异步路由处理 - 加载组件失败:', item.component)
          }
          // 无论是否有子路由，都需要将 component 设置为有效的值
          // 对于有子路由的父路由，设置为 undefined，让 Vue Router 处理
          item.component = undefined
        }
      } catch (error) {
        // 只在开发环境输出错误日志
        if (import.meta.env.DEV) {
          console.error(`异步路由处理 - 加载组件 ${item.component} 出错:`, error)
        }
        // 无论是否有子路由，都需要将 component 设置为有效的值
        // 对于有子路由的父路由，设置为 undefined，让 Vue Router 处理
        item.component = undefined
      }
    } else if (item.component) {
      // 只在开发环境输出详细日志
      if (import.meta.env.DEV) {
        console.log('异步路由处理 - 组件已经是函数:', item.component)
      }
    } else {
      // 只在开发环境输出详细日志
      if (import.meta.env.DEV) {
        console.log('异步路由处理 - 路由没有组件:', item.path)
      }
    }

    if (item.children && item.children.length > 0) {
      // 只在开发环境输出详细日志
      if (import.meta.env.DEV) {
        console.log(
          '异步路由处理 - 处理子路由:',
          item.path,
          item.children.length,
        )
      }
      // 递归处理子路由
      asyncRouterHandle(item.children)

      // 如果父路由没有组件但有子路由，让 Vue Router 处理
      if (item.component === undefined && item.children.length > 0) {
        // 只在开发环境输出详细日志
        if (import.meta.env.DEV) {
          console.log(
            '异步路由处理 - 父路由没有组件但有子路由:',
            item.path,
          )
        }
        // 不设置默认组件，让 Vue Router 处理
      }
    }
  })

  // 只在开发环境输出详细日志
  if (import.meta.env.DEV) {
    console.log('异步路由处理 - 完成路由处理')
  }
}

function dynamicImport(
  dynamicViewsModules: Record<string, () => Promise<unknown>>,
  component: string,
) {
  // 只在开发环境输出详细日志
  if (import.meta.env.DEV) {
    console.log('正在查找组件:', component)
  }

  const keys = Object.keys(dynamicViewsModules)

  // 只在开发环境输出详细日志
  if (import.meta.env.DEV && keys.length > 0) {
    console.log('可用模块数量:', keys.length)
  }

  // 先尝试精确匹配
  let matchKey = keys.find((key) => {
    const k = key.replace('../', '').replace('.vue', '')
    return k === component
  })

  // 如果精确匹配失败，尝试不区分大小写的匹配
  if (!matchKey) {
    matchKey = keys.find((key) => {
      const k = key.replace('../', '').replace('.vue', '')
      return k.toLowerCase() === component.toLowerCase()
    })
  }

  // 如果还是失败，尝试前缀匹配
  if (!matchKey) {
    matchKey = keys.find((key) => {
      const k = key.replace('../', '').replace('.vue', '')
      return k.endsWith(component)
    })
  }

  if (!matchKey) {
    // 只在开发环境输出错误日志
    if (import.meta.env.DEV) {
      console.error(`组件未找到：${component}`)
    }
    // 检查是否是系统菜单，返回基础 layout 作为 fallback
    if (component.startsWith('views/base')) {
      return dynamicViewsModules['../views/base/index.vue'] as () => Promise<{
        default: import('vue').Component
      }>
    }
    // 不返回 null，而是返回 undefined
    return undefined
  }

  // 只在开发环境输出详细日志
  if (import.meta.env.DEV) {
    console.log('找到组件:', matchKey)
  }

  return dynamicViewsModules[matchKey] as () => Promise<{ default: import('vue').Component }>
}
