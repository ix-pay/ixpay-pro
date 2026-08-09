import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/pages/home/index.vue'),
    meta: {
      title: 'ixpay',
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