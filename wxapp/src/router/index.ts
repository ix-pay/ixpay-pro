import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/pages/home/index.vue'),
    meta: {
      title: 'ixpay',
      tabbar: true, // 底部 tab 页
    },
  },
  {
    // 服务号授权回跳落地页：fragment 取 token 后回跳授权前页面
    path: '/oauth/callback',
    name: 'OAuthCallback',
    component: () => import('@/pages/oauth/callback.vue'),
    meta: {
      title: '授权登录',
    },
  },
  {
    path: '/pending-list',
    name: 'PendingList',
    component: () => import('@/pages/pending-list/index.vue'),
    meta: {
      title: '待支付列表',
    },
  },
  {
    path: '/payment',
    name: 'Payment',
    component: () => import('@/pages/payment/index.vue'),
    meta: {
      title: '微信支付',
    },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/pages/profile/index.vue'),
    meta: {
      title: '个人中心',
      tabbar: true, // 底部 tab 页
    },
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// 路由守卫：设置页面标题
router.beforeEach((to, _from, next) => {
  if (to.meta.title) {
    document.title = to.meta.title as string
  }
  next()
})

export default router
