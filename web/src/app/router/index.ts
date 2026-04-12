import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { useRouterStore } from '@/stores/modules/router'
import type { ExtendedRouteRecordRaw } from '@/stores/modules/router'
import { store } from '@/stores'

// 路由类型定义
type AppRouteRecordRaw = RouteRecordRaw & {
  meta?: {
    title?: string
    icon?: string
    closeTab?: boolean
    hidden?: boolean
    keepAlive?: boolean
    [key: string]: unknown
  }
}

// 固定路由 - 不需要登录即可访问的路由
const fixedRoutes: AppRouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/base/login/index.vue'),
    meta: {
      title: '登录',
      hidden: true,
      closeTab: true,
    },
  },
]

// 基础布局路由 - 需要登录才能访问的基础路由
const baseLayoutRoute: AppRouteRecordRaw = {
  path: '/',
  name: 'layout',
  component: () => import('@/components/layout/BaseLayout.vue'),
  redirect: '/index',
  meta: {
    title: '',
    hidden: false,
    keepAlive: true,
  },
  children: [
    // 个人资料和系统设置路由作为 Layout 的子路由添加
    {
      path: 'profile',
      name: 'UserProfile',
      component: () => import('@/views/base/profile/index.vue'),
      meta: {
        title: '个人资料',
        hidden: false,
      },
    },
    {
      path: 'settings',
      name: 'SystemSetting',
      component: () => import('@/views/base/setting/index.vue'),
      meta: {
        title: '系统设置',
        hidden: false,
      },
    },
  ],
}

// 创建路由实例
const router = createRouter({
  history: createWebHashHistory(import.meta.env.VITE_BASE_URL),
  routes: [...fixedRoutes, baseLayoutRoute],
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore(store)
  const routerStore = useRouterStore(store)

  const token = userStore.token

  // 未登录状态下只能访问固定路由
  if (!token) {
    if (to.path === '/login') {
      next()
    } else {
      next('/login')
    }
    return
  }

  // 已登录状态下访问login页面，重定向到首页
  if (to.path === '/login') {
    next({ path: '/index', replace: true })
    return
  }

  // 已登录状态下，确保动态路由已加载
  if (!routerStore.asyncRouterFlag) {
    try {
      // 加载用户信息
      await userStore.GetUserInfo()

      // 从接口获取并加载动态路由
      const dynamicRoutes = await routerStore.SetAsyncRouter()

      // 将有效动态路由添加到 layout 下（过滤掉已存在的路由）
      const validDynamicRoutes = dynamicRoutes.filter(
        (route) => route && route.name && !router.hasRoute(route.name),
      )
      validDynamicRoutes.forEach((route: ExtendedRouteRecordRaw) => {
        router.addRoute('layout', route as unknown as RouteRecordRaw)
      })

      // 添加 404 路由，在动态路由加载完成后添加（只添加一次）
      if (!router.hasRoute('NotFound')) {
        router.addRoute('layout', {
          path: '/:catchAll(.*)',
          name: 'NotFound',
          meta: {
            title: '404',
            closeTab: true,
            hidden: true,
          },
          component: () => import('@/components/business/error/index.vue'),
        })
      }

      // 重新导航以应用新路由
      // 确保导航到正确的路径，而不是默认的首页
      next({ path: to.path, replace: true })
      return
    } catch {
      userStore.ClearStorage()
      next('/login')
      return
    }
  }

  // 认证路由已加载，直接放行
  next()
})

// 添加全局错误处理
router.onError(() => {
  // 如果是组件加载错误，可能是因为动态路由还没有加载完成
  const userStore = useUserStore(store)
  const routerStore = useRouterStore(store)
  if (userStore.token && !routerStore.asyncRouterFlag) {
    userStore.ClearStorage()
    router.push('/login')
  }
})

export default router
export type { AppRouteRecordRaw }
