<template><!-- 水印 -->
  <el-watermark v-if="config.show_watermark" :font="watermarkFont" :gap="[180, 150]"
    :content="userStore.userInfo.nickname">
    <el-container class="relative h-screen w-screen overflow-hidden bg-[var(--bg-secondary)]">
      <!-- 侧边栏（左侧布局） -->
      <div v-if="!isTopMenuLayout" :class="[
        'relative h-full overflow-hidden transition-all',
        isMobile ? 'fixed inset-y-0 left-0 z-[1000]' : '',
        mobileSidebarClass,
      ]">
        <gva-aside :is-collapsed="isSidebarCollapsed" @toggle="toggleSidebar" @menu-select="handleMenuSelect" />
      </div>

      <!-- 移动端遮罩层 -->
      <Transition enter-active-class="transition-opacity duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
        enter-from-class="opacity-0" enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
        leave-from-class="opacity-100" leave-to-class="opacity-0">
        <div v-if="isMobile && !isSidebarCollapsed" class="fixed inset-0 bg-black/50 z-[999]" @click="toggleSidebar" />
      </Transition>

      <!-- 右侧区域：上中下布局 -->
      <el-container class="flex flex-col h-full relative z-[1] overflow-hidden">
        <!-- 页头 -->
        <el-header :class="[
          'flex-shrink-0 overflow-hidden bg-[var(--bg-primary)]',
          'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
          'border-b border-[var(--border-primary)]',
        ]">
          <gva-header :breadcrumb-list="breadcrumbList" :is-sidebar-collapsed="isSidebarCollapsed"
            @toggle-sidebar="toggleSidebar" />
        </el-header>

        <!-- 顶部菜单栏（顶部布局） -->
        <div v-if="isTopMenuLayout" :class="[
          'flex-shrink-0 bg-[var(--bg-primary)] border-b border-[var(--border-primary)]',
          'px-4 overflow-x-auto',
        ]">
          <div class="flex items-center h-12 gap-1">
            <template v-for="menu in topMenuList" :key="menu.id">
              <el-popover v-if="menu.children && menu.children.length > 0" :key="menu.id" trigger="hover" :width="200"
                placement="bottom-start" :hide-after="100">
                <div class="flex flex-col gap-0.5 p-1">
                  <div v-for="child in menu.children" :key="child.id"
                    class="px-3 py-2 rounded-md cursor-pointer hover:bg-[var(--el-color-primary-light-9)] text-sm text-[var(--text-primary)] transition-colors"
                    @click="handleMenuSelect(getFullPath(menu.path, child.path))">
                    {{ child.meta?.title || child.name }}
                  </div>
                </div>
                <template #reference>
                  <div
                    class="flex items-center gap-1.5 px-3 py-1.5 rounded-md cursor-pointer hover:bg-[var(--el-color-primary-light-9)] text-sm text-[var(--text-primary)] transition-colors whitespace-nowrap"
                    :class="{ 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]': isTopMenuActive(menu) }">
                    <span>{{ menu.meta?.title || menu.name }}</span>
                    <el-icon class="text-xs">
                      <ArrowDown />
                    </el-icon>
                  </div>
                </template>
              </el-popover>
              <div v-else
                class="flex items-center px-3 py-1.5 rounded-md cursor-pointer hover:bg-[var(--el-color-primary-light-9)] text-sm text-[var(--text-primary)] transition-colors whitespace-nowrap"
                :class="{ 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]': isTopMenuActive(menu) }"
                @click="handleMenuSelect(getMenuPath(menu.path))">
                {{ menu.meta?.title || menu.name }}
              </div>
            </template>
          </div>
        </div>

        <!-- 内容区域 -->
        <el-main :class="[
          'flex-1 overflow-hidden bg-[var(--bg-secondary)]',
          'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
        ]">
          <tab-manager ref="tabManagerRef" />
        </el-main>

        <!-- 页脚 -->
        <el-footer :class="[
          'h-auto min-h-[40px] flex-shrink-0',
          'bg-[var(--bg-primary)] text-[var(--text-primary)]',
          'border-t border-[var(--border-primary)]',
          'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
        ]">
          <BottomInfo />
        </el-footer>
      </el-container>
    </el-container>
  </el-watermark>
  <el-container v-else class="relative h-screen w-screen overflow-hidden bg-[var(--bg-secondary)]">
    <!-- 侧边栏（左侧布局） -->
    <div v-if="!isTopMenuLayout" :class="[
      'relative h-full overflow-hidden transition-all',
      isMobile ? 'fixed inset-y-0 left-0 z-[1000]' : '',
      mobileSidebarClass,
    ]">
      <gva-aside :is-collapsed="isSidebarCollapsed" @toggle="toggleSidebar" @menu-select="handleMenuSelect" />
    </div>

    <!-- 移动端遮罩层 -->
    <Transition enter-active-class="transition-opacity duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
      enter-from-class="opacity-0" enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
      leave-from-class="opacity-100" leave-to-class="opacity-0">
      <div v-if="isMobile && !isSidebarCollapsed" class="fixed inset-0 bg-black/50 z-[999]" @click="toggleSidebar" />
    </Transition>

    <!-- 右侧区域：上中下布局 -->
    <el-container class="flex flex-col h-full relative z-[1] overflow-hidden">
      <!-- 页头 -->
      <el-header :class="[
        'flex-shrink-0 overflow-hidden bg-[var(--bg-primary)]',
        'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
        'border-b border-[var(--border-primary)]',
      ]">
        <gva-header :breadcrumb-list="breadcrumbList" :is-sidebar-collapsed="isSidebarCollapsed"
          @toggle-sidebar="toggleSidebar" />
      </el-header>

      <!-- 顶部菜单栏（顶部布局） -->
      <div v-if="isTopMenuLayout" :class="[
        'flex-shrink-0 bg-[var(--bg-primary)] border-b border-[var(--border-primary)]',
        'px-4 overflow-x-auto',
      ]">
        <div class="flex items-center h-12 gap-1">
          <template v-for="menu in topMenuList" :key="menu.id">
            <el-popover v-if="menu.children && menu.children.length > 0" :key="menu.id" trigger="hover" :width="200"
              placement="bottom-start" :hide-after="100">
              <div class="flex flex-col gap-0.5 p-1">
                <div v-for="child in menu.children" :key="child.id"
                  class="px-3 py-2 rounded-md cursor-pointer hover:bg-[var(--el-color-primary-light-9)] text-sm text-[var(--text-primary)] transition-colors"
                  @click="handleMenuSelect(getFullPath(menu.path, child.path))">
                  {{ child.meta?.title || child.name }}
                </div>
              </div>
              <template #reference>
                <div
                  class="flex items-center gap-1.5 px-3 py-1.5 rounded-md cursor-pointer hover:bg-[var(--el-color-primary-light-9)] text-sm text-[var(--text-primary)] transition-colors whitespace-nowrap"
                  :class="{ 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]': isTopMenuActive(menu) }">
                  <span>{{ menu.meta?.title || menu.name }}</span>
                  <el-icon class="text-xs">
                    <ArrowDown />
                  </el-icon>
                </div>
              </template>
            </el-popover>
            <div v-else
              class="flex items-center px-3 py-1.5 rounded-md cursor-pointer hover:bg-[var(--el-color-primary-light-9)] text-sm text-[var(--text-primary)] transition-colors whitespace-nowrap"
              :class="{ 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]': isTopMenuActive(menu) }"
              @click="handleMenuSelect(getMenuPath(menu.path))">
              {{ menu.meta?.title || menu.name }}
            </div>
          </template>
        </div>
      </div>

      <!-- 内容区域 -->
      <el-main :class="[
        'flex-1 overflow-hidden bg-[var(--bg-secondary)]',
        'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
      ]">
        <tab-manager ref="tabManagerRef" />
      </el-main>

      <!-- 页脚 -->
      <el-footer :class="[
        'h-auto min-h-[40px] flex-shrink-0',
        'bg-[var(--bg-primary)] text-[var(--text-primary)]',
        'border-t border-[var(--border-primary)]',
        'transition-all duration-[var(--duration-normal)] ease-[cubic-bezier(0.4,0,0.2,1)]',
      ]">
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
import { useRouterStore } from '@/stores/modules/router'
import type { ExtendedRouteRecordRaw } from '@/stores/modules/router'
import { storeToRefs } from 'pinia'
import { ArrowDown } from '@element-plus/icons-vue'
import { getMenuPath, getFullPath } from '@/utils/menu'

defineOptions({
  name: 'BaseLayout',
})

const appStore = useAppStore()
const { config, isDark } = storeToRefs(appStore)
const routerStore = useRouterStore()
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

// 菜单选择处理（左侧和顶部菜单路径格式统一，均带 / 前缀）
const handleMenuSelect = (path: string) => {
  router.push(path)
}

// 是否为顶部菜单布局
const isTopMenuLayout = computed(() => {
  return config.value.menuLayout === 'top'
})

// 顶部菜单列表（从 routerStore 获取）
const topMenuList = computed(() => {
  return routerStore.asyncRouters || []
})

// 判断顶部菜单项是否激活
const isTopMenuActive = (menu: ExtendedRouteRecordRaw): boolean => {
  const currentPath = route.path
  if (menu.path && currentPath.includes('/' + menu.path)) {
    return true
  }
  if (menu.children) {
    return menu.children.some((child) => child.path && currentPath.includes('/' + child.path))
  }
  return false
}

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
