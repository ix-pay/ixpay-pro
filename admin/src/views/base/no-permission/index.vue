<template>
  <div class="flex flex-col h-full items-center justify-center bg-[var(--bg-color)] rounded-lg p-8">
    <div class="text-center">
      <div
        class="mb-6 flex h-32 w-full items-center justify-center rounded-full bg-warning-light/20 text-warning-color">
        <el-icon :size="64">
          <Lock />
        </el-icon>
      </div>
      <h1 class="mb-3 text-2xl font-bold text-text-primary">暂无访问权限</h1>
      <p class="mb-8 max-w-md text-text-secondary">
        您当前的账户尚未分配任何菜单权限，请联系管理员为您分配相应的权限后再试。
      </p>
      <div class="flex items-center justify-center gap-4">
        <el-button type="primary" @click="logoutHandler">退出登录</el-button>
        <el-button @click="goToProfile">查看个人资料</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { useRouterStore } from '@/stores/modules/router'
import { Lock } from '@element-plus/icons-vue'

defineOptions({
  name: 'NoPermission',
})

const router = useRouter()
const userStore = useUserStore()
const routerStore = useRouterStore()

// 挂载时检查：如果用户有菜单权限，跳转到第一个可用页面
onMounted(() => {
  if (routerStore.asyncRouters && routerStore.asyncRouters.length > 0) {
    const findFirstPage = (menus: any[]): string | null => {
      for (const menu of menus) {
        if (menu.type === 2 && menu.path) {
          return menu.path.startsWith('/') ? menu.path : `/${menu.path}`
        }
        if (menu.children) {
          const result = findFirstPage(menu.children)
          if (result) return result
        }
      }
      return null
    }
    const firstPage = findFirstPage(routerStore.asyncRouters)
    if (firstPage) {
      router.replace(firstPage)
    }
  }
})

const logoutHandler = async () => {
  await userStore.LoginOut()
  router.push('/login')
}

const goToProfile = () => {
  router.push('/profile')
}
</script>