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
  redirect: '/no-permission',
  meta: {
    title: '',
    hidden: false,
    keepAlive: true,
  },
  children: [
    // 无权限页面 - 所有登录用户都可访问，作为登录后的默认页面
    {
      path: 'no-permission',
      name: 'NoPermission',
      component: () => import('@/views/base/no-permission/index.vue'),
      meta: {
        title: '无权限',
        hidden: true,
        closeTab: true,
      },
    },
    // 首页路由 - 不再默认添加，改为从接口动态加载
    // 管理员登录后接口返回首页数据，首页路由自动添加
    // 普通用户登录后接口不返回首页数据，首页路由不存在
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

  // 未登录 - 只能访问登录页
  if (!token) {
    if (to.path === '/login') {
      next()
    } else {
      next('/login')
    }
    return
  }

  // 已登录 - 访问登录页则重定向到默认页面
  if (to.path === '/login') {
    next({ path: '/no-permission', replace: true })
    return
  }

  // 已登录 - 确保动态路由已加载（页面刷新或直接访问 URL 时进入此分支）
  if (!routerStore.dynamicRoutesLoaded) {
    try {
      await userStore.GetUserInfo()
      const dynamicRoutes = await routerStore.SetAsyncRouter()

      // 将动态路由添加到 layout 下（过滤掉已存在的路由）
      const validDynamicRoutes = dynamicRoutes.filter(
        (route) => route && route.name && !router.hasRoute(route.name),
      )
      validDynamicRoutes.forEach((route: ExtendedRouteRecordRaw) => {
        router.addRoute('layout', route as unknown as RouteRecordRaw)
      })

      // 添加 404 路由
      if (!router.hasRoute('NotFound')) {
        router.addRoute('layout', {
          path: '/:catchAll(.*)',
          name: 'NotFound',
          meta: { title: '404', closeTab: true, hidden: true },
          component: () => import('@/components/business/Error/index.vue'),
        })
      }

      routerStore.setDynamicRoutesLoaded(true)

      // 如果当前是根路径或默认页，跳转到第一个可用菜单页面
      if (to.path === '/' || to.path === '/no-permission') {
        const firstPage = routerStore.getFirstPagePath()
        if (firstPage) {
          next({ path: firstPage, replace: true })
          return
        }
      }

      next({ path: to.path, query: to.query, hash: to.hash, replace: true })
      return
    } catch {
      userStore.ClearStorage()
      next('/login')
      return
    }
  }

  // 动态路由已加载 - 检查路由是否存在
  const routeExists = router.getRoutes().some(
    (route) => route.name === to.name || route.path === to.path,
  )

  if (!routeExists) {
    next({ name: 'NotFound' })
    return
  }

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
