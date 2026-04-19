<template>
  <div class="w-full h-full flex flex-col overflow-hidden bg-[var(--bg-color)]">
    <!-- 标签栏容器 - 固定不滚动 -->
    <div class="flex-shrink-0 bg-[var(--bg-color)] border-b border-[var(--border-color)]">
      <div ref="scrollContainerRef" class="tab-scroll-container">
        <el-tabs
          v-model="activeTab"
          type="card"
          :closable="tabs.length > 1"
          @tab-click="handleTabClick"
          @tab-remove="handleTabRemove"
          @contextmenu.prevent="handleContextMenu"
          class="custom-tabs"
        >
          <template #default>
            <el-tab-pane
              v-for="tab in tabs"
              :key="tab.path"
              :label="tab.label"
              :name="tab.path"
              class="custom-tab-pane"
            >
            </el-tab-pane>
          </template>
        </el-tabs>

        <!-- 滚动控制按钮 -->
        <div class="scroll-controls flex items-center gap-1 px-2" v-if="showScrollButtons">
          <button
            @click="scrollLeft"
            class="scroll-btn p-1.5 rounded-md hover:bg-[var(--bg-hover-color)] transition-all duration-200 disabled:opacity-30 disabled:cursor-not-allowed"
            :disabled="scrollLeftDisabled"
          >
            <el-icon class="text-[var(--text-secondary)]">
              <ArrowLeft />
            </el-icon>
          </button>
          <button
            @click="scrollRight"
            class="scroll-btn p-1.5 rounded-md hover:bg-[var(--bg-hover-color)] transition-all duration-200 disabled:opacity-30 disabled:cursor-not-allowed"
            :disabled="scrollRightDisabled"
          >
            <el-icon class="text-[var(--text-secondary)]">
              <ArrowRight />
            </el-icon>
          </button>
        </div>
      </div>
    </div>

    <!-- 标签页内容区域 - 可滚动 -->
    <div class="flex-1 overflow-y-auto bg-[var(--bg-page)] transition-colors duration-300">
      <keep-alive :include="keepAliveNames">
        <component :is="getComponentByPath(activeTab)" v-show="true" />
      </keep-alive>
    </div>

    <!-- 右键菜单 -->
    <div
      v-show="contextMenuVisible"
      :style="{ left: contextMenuLeft + 'px', top: contextMenuTop + 'px' }"
      class="fixed z-[9999] bg-[var(--bg-menu)] border border-[var(--border-color)] rounded-lg shadow-2xl py-1.5 min-w-[180px] backdrop-blur-sm animate-fade-in"
    >
      <div class="context-menu-list">
        <button
          class="context-menu-item px-4 py-2.5 w-full flex items-center gap-3 hover:bg-[var(--bg-hover-color)] transition-all duration-150 disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent"
          :disabled="currentIndex === 0"
          @click="handleContextMenuAction('closeLeft')"
        >
          <el-icon class="text-base">
            <ArrowLeft />
          </el-icon>
          <span class="text-sm font-medium text-[var(--text-primary)]">关闭左侧</span>
        </button>

        <button
          class="context-menu-item px-4 py-2.5 w-full flex items-center gap-3 hover:bg-[var(--bg-hover-color)] transition-all duration-150 disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent"
          :disabled="currentIndex === tabs.length - 1"
          @click="handleContextMenuAction('closeRight')"
        >
          <el-icon class="text-base">
            <ArrowRight />
          </el-icon>
          <span class="text-sm font-medium text-[var(--text-primary)]">关闭右侧</span>
        </button>

        <button
          class="context-menu-item px-4 py-2.5 w-full flex items-center gap-3 hover:bg-[var(--bg-hover-color)] transition-all duration-150 disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent"
          :disabled="tabs.length <= 2"
          @click="handleContextMenuAction('closeOther')"
        >
          <el-icon class="text-base">
            <CloseBold />
          </el-icon>
          <span class="text-sm font-medium text-[var(--text-primary)]">关闭其他</span>
        </button>

        <div class="border-t border-[var(--border-color)] my-1"></div>

        <button
          class="context-menu-item px-4 py-2.5 w-full flex items-center gap-3 hover:bg-[var(--bg-hover-color)] transition-all duration-150 disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent text-red-500 hover:text-red-600"
          :disabled="tabs.length <= 1"
          @click="handleContextMenuAction('closeAll')"
        >
          <el-icon class="text-base">
            <FolderDelete />
          </el-icon>
          <span class="text-sm font-medium">关闭全部</span>
        </button>
      </div>
    </div>

    <!-- 遮罩层，点击关闭右键菜单 -->
    <div v-show="contextMenuVisible" class="fixed inset-0 z-[9998]" @click="closeContextMenu"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick, computed, type Component, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useRouterStore } from '@/stores/modules/router'
import type { ExtendedRouteRecordRaw } from '@/stores/modules/router'
import { storeToRefs } from 'pinia'
import { ArrowLeft, ArrowRight, CloseBold, FolderDelete } from '@element-plus/icons-vue'
import { emitter } from '@/utils/bus'

const router = useRouter()
const route = useRoute()
const routerStore = useRouterStore()
const { asyncRouters } = storeToRefs(routerStore)

// 标签页列表
const tabs = ref<TabItem[]>([])
// 当前激活的标签页
const activeTab = ref<string>('')
// 右键菜单相关
const contextMenuVisible = ref(false)
const contextMenuLeft = ref(0)
const contextMenuTop = ref(0)
const currentContextMenuPath = ref<string>('')
// keep-alive 缓存的路由名称列表
const keepAliveNames = ref<string[]>([])
// 滚动控制相关
const scrollContainerRef = ref<HTMLElement | null>(null)
const showScrollButtons = ref(false)
const scrollLeftDisabled = ref(true)
const scrollRightDisabled = ref(true)

// 标签页类型定义
interface TabItem {
  path: string
  label: string
  keepAlive: boolean
  title?: string
}

// 计算当前右键菜单对应的标签页索引
const currentIndex = computed(() => {
  return tabs.value.findIndex((tab) => tab.path === currentContextMenuPath.value)
})

// 从本地存储加载标签页
const loadTabsFromStorage = () => {
  const storedTabs = localStorage.getItem('tabManagerTabs')
  const storedActiveTab = localStorage.getItem('tabManagerActiveTab')

  if (storedTabs) {
    tabs.value = JSON.parse(storedTabs)
  }

  if (storedActiveTab) {
    activeTab.value = storedActiveTab
  }
}

// 保存标签页到本地存储
const saveTabsToStorage = () => {
  localStorage.setItem('tabManagerTabs', JSON.stringify(tabs.value))
  localStorage.setItem('tabManagerActiveTab', activeTab.value)
}

// 根据路径获取组件和组件名
const getComponentInfoByPath = (path: string): { component: Component; name?: string } => {
  // 使用 router.resolve 获取路由信息
  const resolved = router.resolve(path)
  const matched = resolved.matched[resolved.matched.length - 1]

  if (matched && matched.components && matched.components.default) {
    const component = matched.components.default as Component
    // 获取组件的 name 属性（从 options 或 __name）
    const componentName =
      ((component as Record<string, unknown>).name as string | undefined) ||
      ((component as Record<string, unknown>).__name as string | undefined)

    return {
      component,
      name: componentName,
    }
  }

  // 如果找不到组件，返回一个空组件
  return {
    component: { render: () => null } as Component,
    name: undefined,
  }
}

// 根据路径获取组件
const getComponentByPath = (path: string): Component => {
  return getComponentInfoByPath(path).component
}

// 根据路径获取组件名（用于 keep-alive 的 include）
// 优先从路由 meta 中获取组件名，确保 keep-alive 能正确缓存
const getComponentName = (path: string): string | undefined => {
  // 方法 1: 从路由 matched 中获取组件名
  const resolved = router.resolve(path)
  const matched = resolved.matched[resolved.matched.length - 1]

  if (matched && matched.components && matched.components.default) {
    const component = matched.components.default as Component
    const componentName =
      ((component as Record<string, unknown>).name as string | undefined) ||
      ((component as Record<string, unknown>).__name as string | undefined)

    if (componentName) {
      return componentName
    }
  }

  // 方法 2: 从路由 name 获取 (更可靠)
  if (matched && matched.name) {
    return matched.name as string
  }

  return undefined
}

// 获取路由对应的菜单项信息
const getMenuInfoByPath = (
  path: string,
): { label: string; keepAlive: boolean; title?: string } | null => {
  const findMenu = (menus: ExtendedRouteRecordRaw[]): ExtendedRouteRecordRaw | null => {
    for (const menu of menus) {
      if (menu.path === path || menu.path === path.replace(/^\//, '')) {
        return menu
      }
      if (menu.children && menu.children.length > 0) {
        const found = findMenu(menu.children)
        if (found) {
          return found
        }
      }
    }
    return null
  }

  // 从路由匹配中获取更准确的信息
  const matchedRoute = router.resolve(path).matched[router.resolve(path).matched.length - 1]

  if (matchedRoute && matchedRoute.meta) {
    return {
      label: String(matchedRoute.meta.title || '未命名页面'),
      keepAlive: Boolean(matchedRoute.meta.keepAlive || false),
      title: matchedRoute.meta.title as string | undefined,
    }
  }

  // 如果路由匹配中没有，再从 asyncRouters 中查找
  const menuItem = findMenu(asyncRouters.value)

  if (menuItem && menuItem.meta) {
    return {
      label: String(menuItem.meta.title || '未命名页面'),
      keepAlive: Boolean(menuItem.meta.keepAlive || false),
      title: menuItem.meta.title,
    }
  }

  return null
}

// 添加标签页
const addTab = (path: string) => {
  // 检查标签页是否已存在
  const existingTab = tabs.value.find((tab) => tab.path === path)
  if (existingTab) {
    return
  }

  // 获取菜单项信息
  const menuInfo = getMenuInfoByPath(path)
  if (!menuInfo) {
    console.warn(`未找到路径 ${path} 对应的菜单信息`)
    return
  }

  // 获取组件名 (优先用于 keep-alive)
  const componentName = getComponentName(path)

  // 如果组件需要 keep-alive，添加到缓存列表
  if (menuInfo.keepAlive) {
    // 优先使用 getComponentName 获取的名称
    if (componentName && !keepAliveNames.value.includes(componentName)) {
      keepAliveNames.value.push(componentName)
    }
  }

  // 添加新标签页
  const newTab: TabItem = {
    path,
    label: menuInfo.label,
    keepAlive: menuInfo.keepAlive,
  }

  tabs.value.push(newTab)

  // 保存到本地存储
  saveTabsToStorage()
}

// 处理标签页点击
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const handleTabClick = (tab: any) => {
  const path = String(tab.props.name)
  if (path !== route.path) {
    // 确保点击标签页时也更新 keepAliveNames
    const menuInfo = getMenuInfoByPath(path)
    if (menuInfo && menuInfo.keepAlive) {
      const componentName = getComponentName(path)
      if (componentName && !keepAliveNames.value.includes(componentName)) {
        keepAliveNames.value.push(componentName)
      }
    }
    // 直接更新路由
    router.push(path)
  }
}

// 处理标签页关闭
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const handleTabRemove = (path: any) => {
  const pathStr = String(path)
  // 如果只有一个标签页，不允许关闭
  if (tabs.value.length <= 1) return

  const index = tabs.value.findIndex((tab) => tab.path === pathStr)
  if (index === -1) return

  // 获取组件名并从 keep-alive 缓存中移除
  const componentName = getComponentName(pathStr)
  if (componentName) {
    const keepAliveIndex = keepAliveNames.value.indexOf(componentName)
    if (keepAliveIndex > -1) {
      keepAliveNames.value.splice(keepAliveIndex, 1)
    }
  }

  tabs.value.splice(index, 1)

  // 如果关闭的是当前激活的标签页，激活上一个标签页
  if (path === activeTab.value) {
    const newActiveIndex = Math.max(0, index - 1)
    activeTab.value = tabs.value[newActiveIndex]?.path || ''
    router.push(activeTab.value)
  }

  // 保存到本地存储
  saveTabsToStorage()
}

// 处理右键菜单
const handleContextMenu = (e: MouseEvent) => {
  e.preventDefault()
  const tabElement = (e.target as HTMLElement).closest('.el-tabs__item') as HTMLElement
  if (!tabElement) return

  const tabName = tabElement.getAttribute('aria-controls')?.replace('pane-', '')
  if (!tabName) return

  currentContextMenuPath.value = tabName
  contextMenuLeft.value = e.clientX
  contextMenuTop.value = e.clientY
  contextMenuVisible.value = true
}

// 关闭右键菜单
const closeContextMenu = () => {
  contextMenuVisible.value = false
}

// 处理右键菜单操作
const handleContextMenuAction = (action: string) => {
  closeContextMenu()

  switch (action) {
    case 'closeLeft':
      closeLeftTabs()
      break
    case 'closeRight':
      closeRightTabs()
      break
    case 'closeOther':
      closeOtherTabs()
      break
    case 'closeAll':
      closeAllTabs()
      break
  }
}

// 关闭左侧标签页
const closeLeftTabs = () => {
  const currentIndex = tabs.value.findIndex((tab) => tab.path === currentContextMenuPath.value)
  if (currentIndex <= 0) return

  const tabsToKeep = tabs.value.slice(currentIndex)
  const tabsToRemove = tabs.value.slice(0, currentIndex)

  // 移除被关闭标签页的 keep-alive 缓存
  tabsToRemove.forEach((tab) => {
    routerStore.removeKeepAliveRouter(tab.path)
  })

  tabs.value = tabsToKeep

  // 如果当前激活的标签页被关闭，激活第一个保留的标签页
  if (!tabsToKeep.find((tab) => tab.path === activeTab.value)) {
    activeTab.value = tabsToKeep[0]?.path || ''
    router.push(activeTab.value)
  }

  saveTabsToStorage()
}

// 关闭右侧标签页
const closeRightTabs = () => {
  const currentIndex = tabs.value.findIndex((tab) => tab.path === currentContextMenuPath.value)
  if (currentIndex === -1 || currentIndex === tabs.value.length - 1) return

  const tabsToKeep = tabs.value.slice(0, currentIndex + 1)
  const tabsToRemove = tabs.value.slice(currentIndex + 1)

  // 移除被关闭标签页的 keep-alive 缓存
  tabsToRemove.forEach((tab) => {
    routerStore.removeKeepAliveRouter(tab.path)
  })

  tabs.value = tabsToKeep

  // 如果当前激活的标签页被关闭，激活最后一个保留的标签页
  if (!tabsToKeep.find((tab) => tab.path === activeTab.value)) {
    activeTab.value = tabsToKeep[tabsToKeep.length - 1]?.path || ''
    router.push(activeTab.value)
  }

  saveTabsToStorage()
}

// 关闭其他标签页
const closeOtherTabs = () => {
  const currentTab = tabs.value.find((tab) => tab.path === currentContextMenuPath.value)
  if (!currentTab) return

  const tabsToRemove = tabs.value.filter((tab) => tab.path !== currentContextMenuPath.value)

  // 移除被关闭标签页的 keep-alive 缓存
  tabsToRemove.forEach((tab) => {
    routerStore.removeKeepAliveRouter(tab.path)
  })

  tabs.value = [currentTab]

  // 激活当前标签页
  activeTab.value = currentTab.path
  router.push(activeTab.value)

  saveTabsToStorage()
}

// 关闭全部标签页
const closeAllTabs = () => {
  if (tabs.value.length <= 1) return

  // 移除所有标签页的 keep-alive 缓存
  tabs.value.forEach((tab) => {
    routerStore.removeKeepAliveRouter(tab.path)
  })

  // 保留第一个标签页（首页）
  const firstTab = tabs.value[0]
  tabs.value = [firstTab]
  activeTab.value = firstTab.path
  router.push(activeTab.value)

  saveTabsToStorage()
}

// 滚动控制方法
const scrollLeft = () => {
  const container = scrollContainerRef.value?.querySelector('.el-tabs__nav-wrap') as HTMLElement
  if (container) {
    container.scrollBy({ left: -300, behavior: 'smooth' })
  }
}

const scrollRight = () => {
  const container = scrollContainerRef.value?.querySelector('.el-tabs__nav-wrap') as HTMLElement
  if (container) {
    container.scrollBy({ left: 300, behavior: 'smooth' })
  }
}

const checkScrollButtons = () => {
  const container = scrollContainerRef.value?.querySelector('.el-tabs__nav-wrap') as HTMLElement
  if (!container) return

  const scrollWidth = container.scrollWidth
  const clientWidth = container.clientWidth
  const scrollLeft = container.scrollLeft

  showScrollButtons.value = scrollWidth > clientWidth
  scrollLeftDisabled.value = scrollLeft <= 0
  scrollRightDisabled.value = scrollLeft + clientWidth >= scrollWidth - 1
}

// 重置所有 tabs（用于角色切换等场景）
const resetTabs = () => {
  // 保留首页 tab
  const homeTab = tabs.value.find((tab) => tab.path === '/index')

  // 清空所有 tabs
  tabs.value = []

  // 清空 keep-alive 缓存
  routerStore.clearAllKeepAlive()

  // 清空 localStorage
  localStorage.removeItem('tabManagerTabs')
  localStorage.removeItem('tabManagerActiveTab')

  // 如果之前有首页 tab，重新添加首页
  if (homeTab) {
    tabs.value.push(homeTab)
  } else {
    // 否则创建新的首页 tab
    tabs.value.push({
      path: '/index',
      label: '首页',
      keepAlive: true,
    })
  }

  // 重定向到首页
  activeTab.value = '/index'
  router.push('/index')
}

// 监听路由变化，自动添加标签页并同步 activeTab
watch(
  () => route.path,
  (newPath) => {
    // 不处理刷新路由和 404 路由
    if (newPath === '/reload' || newPath === '/404') return

    nextTick(() => {
      // 添加标签页
      addTab(newPath)

      // 同步 activeTab 到当前路由
      activeTab.value = newPath
    })
  },
  { immediate: false }, // 初始化
)

// 初始化
onMounted(() => {
  // 延迟加载标签页，确保动态路由已经加载完成
  setTimeout(() => {
    // 从本地存储加载标签页
    loadTabsFromStorage()

    // 过滤掉不存在的路由
    tabs.value = tabs.value.filter((tab) => {
      try {
        router.resolve(tab.path)
        return true
      } catch {
        return false
      }
    })

    // 确保当前路由在标签页列表中
    const currentPath = route.path
    const hasCurrentPath = tabs.value.some((tab) => tab.path === currentPath)

    if (!hasCurrentPath) {
      addTab(currentPath)
    } else {
      // 确保激活的标签页与当前路由匹配
      activeTab.value = currentPath
    }

    // 保存更新后的标签页状态
    saveTabsToStorage()

    // 初始化滚动按钮状态
    nextTick(() => {
      checkScrollButtons()

      // 监听滚动容器的滚动事件
      const container = scrollContainerRef.value?.querySelector('.el-tabs__nav-wrap') as HTMLElement
      if (container) {
        container.addEventListener('scroll', checkScrollButtons)
      }
    })
  }, 100)
})

// 暴露方法给父组件
defineExpose({
  resetTabs,
})

// 监听重置 tabs 的事件
onMounted(() => {
  emitter.on('resetTabManager', resetTabs)
})

// 组件卸载时移除监听器
onUnmounted(() => {
  emitter.off('resetTabManager', resetTabs)
})
</script>

<style scoped>
/* ========================================
   设计令牌变量定义
   ======================================== */
:root {
  /* 背景色 */
  --bg-color: #ffffff;
  --bg-page: #f5f7fa;
  --bg-hover-color: #f0f2f5;
  --bg-menu: #ffffff;

  /* 边框色 */
  --border-color: #e5e7eb;

  /* 文本色 */
  --text-primary: #1f2937;
  --text-secondary: #6b7280;
  --text-tertiary: #9ca3af;

  /* 主题色 */
  --primary-color: #3b82f6;
  --primary-hover: #2563eb;

  /* 阴影 */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);

  /* 圆角 */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
}

/* 暗黑模式 */
html.dark {
  --bg-color: #1f2937;
  --bg-page: #111827;
  --bg-hover-color: #374151;
  --bg-menu: #1f2937;

  --border-color: #374151;

  --text-primary: #f9fafb;
  --text-secondary: #9ca3af;
  --text-tertiary: #6b7280;

  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.3);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.4);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.5);
}

/* ========================================
   标签滚动容器
   ======================================== */
.tab-scroll-container {
  display: flex;
  align-items: center;
  width: 100%;
  overflow: hidden;
}

/* ========================================
   自定义标签页样式
   ======================================== */
.custom-tabs {
  flex: 1;
  width: auto;
}

/* 标签栏头部 */
.custom-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-color);
  transition: all 0.3s ease;
}

/* 标签导航包装器 */
.custom-tabs :deep(.el-tabs__nav-wrap) {
  overflow-x: auto;
  overflow-y: hidden;
  scroll-behavior: smooth;
  scrollbar-width: thin;
  scrollbar-color: var(--border-color) transparent;
}

/* Webkit 滚动条样式 */
.custom-tabs :deep(.el-tabs__nav-wrap::-webkit-scrollbar) {
  height: 4px;
}

.custom-tabs :deep(.el-tabs__nav-wrap::-webkit-scrollbar-track) {
  background: transparent;
}

.custom-tabs :deep(.el-tabs__nav-wrap::-webkit-scrollbar-thumb) {
  background: var(--border-color);
  border-radius: 2px;
  transition: background 0.2s ease;
}

.custom-tabs :deep(.el-tabs__nav-wrap::-webkit-scrollbar-thumb:hover) {
  background: var(--text-secondary);
}

/* 标签导航 */
.custom-tabs :deep(.el-tabs__nav) {
  display: flex;
  gap: 2px;
  padding: 8px 8px 0;
}

/* 单个标签项 */
.custom-tabs :deep(.el-tabs__item) {
  padding: 8px 16px;
  height: auto;
  line-height: 1.5;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md) var(--radius-md) 0 0;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
  position: relative;
  overflow: hidden;
  /* 增强边界可见性 */
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

/* 标签项悬浮效果 */
.custom-tabs :deep(.el-tabs__item:not(.is-active):hover) {
  color: var(--primary-color);
  background: var(--bg-hover-color);
  border-color: var(--primary-color);
  transform: translateY(-1px);
  box-shadow:
    0 2px 4px rgba(59, 130, 246, 0.15),
    0 1px 2px rgba(0, 0, 0, 0.06);
}

/* 激活的标签项 */
.custom-tabs :deep(.el-tabs__item.is-active) {
  color: var(--primary-color);
  background: var(--bg-color);
  border-color: var(--primary-color);
  border-bottom-color: transparent;
  z-index: 10;
  /* 增强激活状态的边界和阴影效果 */
  box-shadow:
    0 -2px 8px rgba(59, 130, 246, 0.15),
    0 1px 3px rgba(0, 0, 0, 0.08);
}

/* 激活标签项底部指示器 */
.custom-tabs :deep(.el-tabs__item.is-active::after) {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--primary-color), var(--primary-hover));
  box-shadow: 0 -2px 8px rgba(59, 130, 246, 0.4);
  animation: slideIn 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes slideIn {
  from {
    transform: scaleX(0);
    opacity: 0;
  }
  to {
    transform: scaleX(1);
    opacity: 1;
  }
}

/* 标签项波纹效果 */
.custom-tabs :deep(.el-tabs__item::before) {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 0;
  height: 0;
  border-radius: 50%;
  background: rgba(59, 130, 246, 0.1);
  transform: translate(-50%, -50%);
  transition:
    width 0.6s ease,
    height 0.6s ease;
}

.custom-tabs :deep(.el-tabs__item:active::before) {
  width: 200px;
  height: 200px;
}

/* 标签内容 */
.custom-tabs :deep(.el-tabs__content) {
  padding: 0;
  overflow: visible;
}

/* ========================================
   关闭按钮样式
   ======================================== */
.custom-tabs :deep(.el-tabs__item.is-closable .el-icon-close) {
  width: 16px;
  height: 16px;
  margin-left: 8px;
  padding: 2px;
  font-size: 12px;
  color: var(--text-tertiary);
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
}

.custom-tabs :deep(.el-tabs__item.is-closable .el-icon-close:hover) {
  color: #ffffff;
  background: #ef4444;
  transform: scale(1.1);
  box-shadow: 0 2px 4px rgba(239, 68, 68, 0.3);
}

/* ========================================
   滚动控制按钮
   ======================================== */
.scroll-controls {
  flex-shrink: 0;
}

.scroll-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid var(--border-color);
  background: var(--bg-color);
  cursor: pointer;
  transition: all 0.2s ease;
}

.scroll-btn:hover:not(:disabled) {
  background: var(--bg-hover-color);
  border-color: var(--primary-color);
  transform: scale(1.05);
}

.scroll-btn:active:not(:disabled) {
  transform: scale(0.95);
}

.scroll-btn:disabled {
  cursor: not-allowed;
  opacity: 0.3;
}

/* ========================================
   右键菜单样式
   ======================================== */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.animate-fade-in {
  animation: fadeIn 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}

.context-menu-list {
  display: flex;
  flex-direction: column;
}

.context-menu-item {
  text-align: left;
  border: none;
  background: transparent;
  cursor: pointer;
  outline: none;
}

.context-menu-item:focus {
  background: var(--bg-hover-color);
}

/* ========================================
   暗黑模式特殊适配
   ======================================== */
html.dark .custom-tabs :deep(.el-tabs__item) {
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

html.dark .custom-tabs :deep(.el-tabs__item.is-active) {
  background: var(--bg-color);
  box-shadow:
    0 -2px 8px rgba(59, 130, 246, 0.3),
    0 1px 3px rgba(0, 0, 0, 0.3);
}

html.dark .custom-tabs :deep(.el-tabs__item:not(.is-active):hover) {
  box-shadow:
    0 2px 4px rgba(59, 130, 246, 0.25),
    0 1px 2px rgba(0, 0, 0, 0.2);
}

html.dark .scroll-btn {
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

html.dark .context-menu-item:hover {
  box-shadow: inset 2px 0 0 var(--primary-color);
}
</style>
