<template>
  <el-container class="h-screen w-screen overflow-hidden bg-[var(--bg-secondary)]">
    <!-- 水印 -->
    <el-watermark
      v-if="config.show_watermark"
      :font="watermarkFont"
      :z-index="9999"
      :gap="[180, 150]"
      :content="userStore.userInfo.nickname"
      class="pointer-events-none"
    />

    <!-- 侧边栏 -->
    <el-aside
      :width="sidebarWidth"
      :class="[
        'relative h-full overflow-hidden transition-all',
        isMobile ? 'fixed inset-y-0 left-0 z-[1000]' : '',
        sidebarWidthClass,
        mobileSidebarClass,
      ]"
    >
      <gva-aside
        :is-collapsed="isSidebarCollapsed"
        @toggle="toggleSidebar"
        @menu-select="handleMenuSelect"
      />
    </el-aside>

    <!-- 移动端遮罩层 -->
    <Transition
      enter-active-class="transition-opacity duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isMobile && !isSidebarCollapsed"
        class="fixed inset-0 bg-black/50 z-[999]"
        @click="toggleSidebar"
      />
    </Transition>

    <!-- 右侧区域：上中下布局 -->
    <el-container class="flex flex-col h-full relative z-[1] overflow-hidden">
      <!-- 页头 -->
      <el-header
        :class="[
          'flex-shrink-0 overflow-hidden bg-[var(--bg-primary)]',
          'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
          'border-b border-[var(--border-primary)]',
        ]"
      >
        <gva-header
          :breadcrumb-list="breadcrumbList"
          :is-sidebar-collapsed="isSidebarCollapsed"
          @toggle-sidebar="toggleSidebar"
        />
      </el-header>

      <!-- 内容区域 -->
      <el-main
        :class="[
          'flex-1 overflow-hidden bg-[var(--bg-secondary)]',
          'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
        ]"
      >
        <tab-manager ref="tabManagerRef" />
      </el-main>

      <!-- 页脚 -->
      <el-footer
        :class="[
          'h-auto min-h-[40px] flex-shrink-0',
          'bg-[var(--bg-primary)] text-[var(--text-primary)]',
          'border-t border-[var(--border-primary)]',
          'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
        ]"
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
import BottomInfo from '@/components/business/BottomInfo/index.vue'
import { ref, reactive, watchEffect, computed, watch, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { useAppStore } from '@/stores'
import { storeToRefs } from 'pinia'

defineOptions({
  name: 'BaseLayout',
})

const appStore = useAppStore()
const { config, isDark } = storeToRefs(appStore)
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

// 响应式
const { screenWidth } = useResponsive(true)
const watermarkFont = reactive({
  color: 'rgba(0, 0, 0, .15)',
})

watchEffect(() => {
  watermarkFont.color = isDark.value ? 'rgba(255,255,255, .15)' : 'rgba(0, 0, 0, .15)'
})

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

// 侧边栏宽度计算
const sidebarWidth = computed(() => {
  return isSidebarCollapsed.value ? '64px' : '240px'
})

// 侧边栏宽度 class
const sidebarWidthClass = computed(() => {
  return isSidebarCollapsed.value ? 'w-[64px]' : 'w-[240px]'
})

// 移动端侧边栏 class
const mobileSidebarClass = computed(() => {
  if (!isMobile.value) return ''
  return isSidebarCollapsed.value ? '-translate-x-full' : 'translate-x-0'
})

onMounted(() => {
  if (userStore.loadingInstance && typeof userStore.loadingInstance.close === 'function') {
    userStore.loadingInstance.close()
    userStore.loadingInstance = null
  }
})
</script>

<style lang="scss" scoped>
// 移动端侧边栏样式
:deep(.mobile-sidebar) {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 1000;
  box-shadow: var(--shadow-lg);
}

// 侧边栏过渡动画优化
.el-aside {
  transition:
    width var(--duration-normal) cubic-bezier(0.4, 0, 0.2, 1),
    transform var(--duration-normal) cubic-bezier(0.4, 0, 0.2, 1);
}

// 内容区自适应过渡
.el-main {
  transition:
    padding var(--duration-normal) cubic-bezier(0.4, 0, 0.2, 1),
    margin var(--duration-normal) cubic-bezier(0.4, 0, 0.2, 1);
}

// 暗黑模式下的阴影增强
html.dark :deep(.mobile-sidebar) {
  box-shadow: var(--shadow-2xl);
}
</style>
