<template>
  <div class="w-full h-full flex flex-col overflow-hidden">
    <!-- 标签栏容器 - 固定不滚动 -->
    <div class="flex-shrink-0 bg-[var(--bg-color)]">
      <el-tabs
        v-model="activeTab"
        type="card"
        :closable="tabs.length > 1"
        @tab-click="handleTabClick"
        @tab-remove="handleTabRemove"
        @contextmenu.prevent="handleContextMenu"
        class="[&_.el-tabs__header]:border-b [&_.el-tabs__header]:border-[var(--border-color)] [&_.el-tabs__nav-wrap]:overflow-x-auto [&_.el-tabs__nav-wrap::-webkit-scrollbar]:h-1 [&_.el-tabs__nav-wrap::-webkit-scrollbar-thumb]:bg-[var(--border-color)] [&_.el-tabs__nav-wrap::-webkit-scrollbar-thumb]:rounded-[2px]"
      >
        <template #default>
          <el-tab-pane v-for="tab in tabs" :key="tab.path" :label="tab.label" :name="tab.path">
          </el-tab-pane>
        </template>
      </el-tabs>
    </div>
    <!-- 标签页内容区域 - 可滚动 -->
    <div class="flex-1 overflow-y-auto bg-gray-50 dark:bg-gray-900">
      <keep-alive :include="keepAliveNames">
        <component :is="getComponentByPath(activeTab)" v-show="true" />
      </keep-alive>
    </div>

    <!-- 右键菜单 -->
    <div
      v-show="contextMenuVisible"
      :style="{ left: contextMenuLeft + 'px', top: contextMenuTop + 'px' }"
      class="fixed z-[9999] bg-[var(--bg-color)] border border-[var(--border-color)] rounded-[var(--radius-md)] shadow-[var(--shadow-lg)] py-2 min-w-[160px]"
    >
      <el-menu :ellipsis="false" class="border-none bg-transparent" style="max-width: 100%">
        <el-menu-item
          index="closeLeft"
          class="px-4 py-2 h-auto hover:bg-[var(--el-menu-hover-bg-color)]"
          :disabled="currentIndex === 0"
          @click="handleContextMenuAction('closeLeft')"
        >
          <el-icon class="mr-2">
            <DArrowLeft />
          </el-icon>
          <span>关闭左侧</span>
        </el-menu-item>

        <el-menu-item
          index="closeRight"
          class="px-4 py-2 h-auto hover:bg-[var(--el-menu-hover-bg-color)]"
          :disabled="currentIndex === tabs.length - 1"
          @click="handleContextMenuAction('closeRight')"
        >
          <el-icon class="mr-2">
            <DArrowRight />
          </el-icon>
          <span>关闭右侧</span>
        </el-menu-item>

        <el-menu-item
          index="closeOther"
          class="px-4 py-2 h-auto hover:bg-[var(--el-menu-hover-bg-color)]"
          :disabled="tabs.length <= 2"
          @click="handleContextMenuAction('closeOther')"
        >
          <el-icon class="mr-2">
            <CloseBold />
          </el-icon>
          <span>关闭其他</span>
        </el-menu-item>

        <el-menu-item
          index="closeAll"
          class="px-4 py-2 h-auto hover:bg-[var(--el-menu-hover-bg-color)]"
          :disabled="tabs.length <= 1"
          @click="handleContextMenuAction('closeAll')"
        >
          <el-icon class="mr-2">
            <FolderDelete />
          </el-icon>
          <span>关闭全部</span>
        </el-menu-item>
      </el-menu>
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
import { DArrowLeft, DArrowRight, CloseBold, FolderDelete } from '@element-plus/icons-vue'
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
/* 使用 Tailwind CSS 重构后，样式已通过 class 定义 */

/* 激活标签页下边框效果 - 类似菜单激活效果 */
:deep(.el-tabs__item.is-active) {
  position: relative;
}

:deep(.el-tabs__item.is-active::after) {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background-color: var(--primary-color);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.4);
  animation: tab-active-border 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes tab-active-border {
  from {
    transform: scaleX(0);
    opacity: 0;
  }

  to {
    transform: scaleX(1);
    opacity: 1;
  }
}

/* 右键菜单项禁用状态 */
:deep(.el-menu-item.is-disabled) {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 右键菜单项图标 */
:deep(.el-menu-item .el-icon) {
  font-size: 16px;
}

/* 确保右键菜单在暗黑模式下正常工作 */
html.dark :deep(.el-menu-item) {
  background-color: transparent;
  color: var(--el-text-color-primary);
}

html.dark :deep(.el-menu-item:hover) {
  background-color: var(--el-menu-hover-bg-color);
}
</style>
