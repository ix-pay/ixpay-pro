<template>
  <el-container class="h-screen overflow-hidden bg-gray-50 dark:bg-gray-900">
    <!-- 水印 -->
    <el-watermark
      v-if="config.show_watermark"
      :font="font"
      :z-index="9999"
      :gap="[180, 150]"
      :content="userStore.userInfo.nickname"
    />

    <!-- 侧边栏 -->
    <el-aside
      :width="isSidebarCollapsed ? '64px' : '240px'"
      :class="{
        'mobile-sidebar': isMobile,
        'transition-all duration-300': true,
      }"
      class="relative z-100 h-full overflow-hidden"
    >
      <gva-aside
        :is-collapsed="isSidebarCollapsed"
        @toggle="toggleSidebar"
        @menu-select="handleMenuSelect"
      />
    </el-aside>

    <!-- 遮罩层（仅移动设备显示） -->
    <div
      v-if="isMobile && !isSidebarCollapsed"
      class="fixed inset-0 bg-black/50 z-999 animate-fade-in"
      @click="toggleSidebar"
    />

    <!-- 右侧区域：上中下布局 -->
    <el-container class="flex flex-col h-full relative z-1">
      <!-- 页头 -->
      <el-header
        class="!p-0 !h-16 flex-shrink-0 overflow-hidden bg-[var(--bg-color)] transition-colors duration-300"
      >
        <gva-header
          :breadcrumb-list="breadcrumbList"
          :is-sidebar-collapsed="isSidebarCollapsed"
          @toggle-sidebar="toggleSidebar"
        />
      </el-header>

      <!-- 内容区域 -->
      <el-main class="!p-0 flex-1 overflow-hidden bg-gray-50 dark:bg-gray-900">
        <tab-manager ref="tabManagerRef" />
      </el-main>

      <!-- 页脚 -->
      <el-footer
        class="!p-0 h-auto min-h-[40px] bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 border-t border-gray-200 dark:border-gray-700 flex-shrink-0 transition-colors duration-300"
      >
        <BottomInfo />
      </el-footer>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import GvaAside from '@/components/layout/Sidebar.vue'
import GvaHeader from '@/components/layout/Header.vue'
import TabManager from '@/components/layout/TabManager.vue'
import useResponsive from '@/hooks/responsive'
import BottomInfo from '@/components/business/bottomInfo/bottomInfo.vue'
import { ref, reactive, watchEffect, computed, watch, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { useAppStore } from '@/stores'
import { storeToRefs } from 'pinia'

const appStore = useAppStore()
const { config, isDark } = storeToRefs(appStore)

defineOptions({
  name: 'BaseLayout',
})

// 响应式
const { screenWidth } = useResponsive(true)
const font = reactive({
  color: 'rgba(0, 0, 0, .15)',
})

watchEffect(() => {
  font.color = isDark.value ? 'rgba(255,255,255, .15)' : 'rgba(0, 0, 0, .15)'
})

const router = useRouter()
const route = useRoute()

// 判断是否为移动设备（小于 768px）
const isMobile = computed(() => screenWidth.value < 768)

// 侧边栏收起/展开状态管理
const isSidebarCollapsed = ref(false)

// 监听屏幕宽度变化，自动切换侧边栏状态
watch(
  () => screenWidth.value,
  (newWidth) => {
    if (newWidth < 768) {
      isSidebarCollapsed.value = true
    } else {
      isSidebarCollapsed.value = false
    }
  },
  { immediate: true },
)

// 切换侧边栏状态
const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

// 获取用户存储
const userStore = useUserStore()

// TabManager 引用
const tabManagerRef = ref<InstanceType<typeof TabManager> | null>(null)

// 面包屑项接口
interface BreadcrumbItem {
  name?: string
  path: string
}

// 面包屑数据
const breadcrumbList = computed<BreadcrumbItem[]>(() => {
  const matched = route.matched
  return matched
    .filter((item) => {
      return !(item.path === '/' && !item.meta.title)
    })
    .map((item) => ({
      name: String(item.meta.title || ''),
      path: item.path.startsWith('/') ? item.path : `/${item.path}`,
    }))
})

// 菜单选择处理
const handleMenuSelect = (path: string) => {
  router.push(path)
}

onMounted(() => {
  if (userStore.loadingInstance && typeof userStore.loadingInstance.close === 'function') {
    userStore.loadingInstance.close()
    userStore.loadingInstance = null
  }
})
</script>

<style lang="scss" scoped>
// 侧边栏移动端样式
:deep(.mobile-sidebar) {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 1000;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
}

// 淡入动画
@keyframes fadeIn {
  from {
    opacity: 0;
  }

  to {
    opacity: 1;
  }
}

.animate-fade-in {
  animation: fadeIn 0.3s ease;
}
</style>
