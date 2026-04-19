<template>
  <el-aside :width="isCollapsed ? '64px' : '240px'" class="sidebar-container">
    <!-- Logo 区域 -->
    <div class="logo-section">
      <img :src="ixpayLogo" alt="IxPay Pro Logo" class="logo-image" />
      <span v-if="!isCollapsed" class="logo-text"> IxPay Pro </span>
    </div>

    <!-- 菜单 -->
    <el-menu
      :default-active="currentRoute"
      class="sidebar-menu"
      unique-opened
      :collapse="isCollapsed"
      :collapse-transition="true"
      @select="handleMenuSelect"
      router
    >
      <!-- 动态菜单 -->
      <template v-if="menuList.length > 0">
        <template v-for="menu in menuList" :key="menu.id">
          <!-- 有子菜单的情况 -->
          <el-sub-menu
            v-if="menu.children && menu.children.length > 0"
            :index="String(menu.name)"
            class="menu-item"
          >
            <template #title>
              <el-icon class="menu-icon">
                <component :is="getIconComponent(menu.meta.icon)" />
              </el-icon>
              <span class="menu-text">{{ menu.meta.title }}</span>
            </template>
            <!-- 递归渲染子菜单 -->
            <template v-for="subMenu in menu.children" :key="subMenu.id">
              <el-menu-item
                v-if="!subMenu.children || subMenu.children.length === 0"
                :index="getFullPath(menu.path, subMenu.path)"
                class="menu-item"
              >
                <template #default>
                  <el-icon class="menu-icon">
                    <component :is="getIconComponent(subMenu.meta.icon)" />
                  </el-icon>
                  <span class="menu-text">{{ subMenu.meta.title }}</span>
                </template>
              </el-menu-item>
              <!-- 三级子菜单 -->
              <el-sub-menu v-else :index="String(subMenu.name)" class="menu-item">
                <template #title>
                  <el-icon class="menu-icon">
                    <component :is="getIconComponent(subMenu.meta.icon)" />
                  </el-icon>
                  <span class="menu-text">{{ subMenu.meta.title }}</span>
                </template>
                <el-menu-item
                  v-for="thirdMenu in subMenu.children"
                  :key="thirdMenu.id"
                  :index="getFullPath(menu.path, getFullPath(subMenu.path, thirdMenu.path))"
                  class="menu-item"
                >
                  <template #default>
                    <el-icon class="menu-icon">
                      <component :is="getIconComponent(thirdMenu.meta.icon)" />
                    </el-icon>
                    <span class="menu-text">{{ thirdMenu.meta.title }}</span>
                  </template>
                </el-menu-item>
              </el-sub-menu>
            </template>
          </el-sub-menu>
          <!-- 没有子菜单的情况 -->
          <el-menu-item
            v-else
            :index="menu.path.startsWith('/') ? menu.path : '/' + menu.path"
            class="menu-item"
          >
            <template #default>
              <el-icon class="menu-icon">
                <component :is="getIconComponent(menu.meta.icon)" />
              </el-icon>
              <span class="menu-text">{{ menu.meta.title }}</span>
            </template>
          </el-menu-item>
        </template>
      </template>
      <!-- 加载状态 -->
      <template v-else-if="loading">
        <el-menu-item disabled class="menu-item">
          <template #title>
            <el-icon class="is-loading menu-icon">
              <Loading />
            </el-icon>
            <span class="menu-text">加载菜单中...</span>
          </template>
        </el-menu-item>
      </template>
      <!-- 无菜单数据状态 -->
      <template v-else>
        <el-menu-item disabled class="menu-item">
          <template #title>
            <el-icon class="menu-icon">
              <DocumentRemove />
            </el-icon>
            <span class="menu-text">暂无菜单数据</span>
          </template>
        </el-menu-item>
      </template>

      <!-- 菜单底部折叠按钮 -->
      <div class="menu-bottom">
        <el-button
          type="text"
          size="small"
          class="collapse-btn"
          @click="toggleCollapse"
          :title="isCollapsed ? '展开菜单' : '收起菜单'"
        >
          <template #icon>
            <ArrowRight v-if="isCollapsed" class="collapse-icon" />
            <ArrowLeft v-else class="collapse-icon" />
          </template>
        </el-button>
      </div>
    </el-menu>
  </el-aside>
</template>

<script setup lang="ts">
import { computed, onMounted, watch, type Component } from 'vue'

defineOptions({
  name: 'LayoutSidebar',
})
import { useRoute } from 'vue-router'
import { ArrowRight, ArrowLeft, Loading, DocumentRemove, House } from '@element-plus/icons-vue'
import ixpayLogo from '@/assets/images/ixpay.png'
import type { ExtendedRouteRecordRaw } from '@/stores/modules/router'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { useRouterStore } from '@/stores/modules/router'

const route = useRoute()
const routerStore = useRouterStore()

const _props = defineProps({
  isCollapsed: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['toggle', 'menu-select'])

// 从 routerStore 获取菜单数据
const menuList = computed<ExtendedRouteRecordRaw[]>(() => {
  routerStore.asyncRouterFlag

  const defaultIndexMenu: ExtendedRouteRecordRaw = {
    id: 1,
    name: 'index',
    path: '/index',
    component: () => import('@/views/base/index/index.vue'),
    meta: {
      title: '首页',
      icon: 'House',
      hidden: false,
      keepAlive: true,
    },
  }

  const allMenus: ExtendedRouteRecordRaw[] = []
  const indexMenuPaths = new Set<string>()

  if (routerStore.asyncRouters.length > 0) {
    const filteredMenus = routerStore.asyncRouters.filter((menu) => {
      if (!menu || typeof menu !== 'object') {
        return false
      }

      if (!menu.meta) {
        menu.meta = { title: '' }
      }

      if (menu.hidden || menu.meta.hidden) {
        return false
      }

      if (!menu.name || !menu.meta.title) {
        return false
      }

      return true
    })

    filteredMenus.forEach((menu) => {
      const isIndexMenu =
        menu.path === 'index' ||
        menu.path === '/index' ||
        menu.name === 'index' ||
        menu.meta?.title === '首页'

      if (isIndexMenu) {
        if (indexMenuPaths.has('/index')) {
          return
        }
        if (menu.path === 'index') {
          menu.path = '/index'
        }
        indexMenuPaths.add('/index')
      }

      allMenus.push(menu)
    })
  }

  if (allMenus.length === 0) {
    allMenus.push(defaultIndexMenu)
  }

  return allMenus
})

const loading = computed(() => {
  return routerStore.asyncRouterFlag === 0
})

const _isMobile = computed(() => {
  return window.innerWidth <= 768
})

const currentRoute = computed(() => {
  return route.path
})

const iconMap: Record<string, Component> = {
  default: House,
  Dashboard: ElementPlusIconsVue.Grid,
  Setting: ElementPlusIconsVue.Setting,
  User: ElementPlusIconsVue.User,
  UserFilled: ElementPlusIconsVue.UserFilled,
  Clock: ElementPlusIconsVue.Clock,
  Menu: ElementPlusIconsVue.Menu,
  Folder: ElementPlusIconsVue.Folder,
  Document: ElementPlusIconsVue.Document,
  Bell: ElementPlusIconsVue.Bell,
  Connection: ElementPlusIconsVue.Connection,
  Monitor: ElementPlusIconsVue.Monitor,
  Tickets: ElementPlusIconsVue.Tickets,
  DocumentCopy: ElementPlusIconsVue.DocumentCopy,
  dashboard: ElementPlusIconsVue.Grid,
  setting: ElementPlusIconsVue.Setting,
  user: ElementPlusIconsVue.User,
  role: ElementPlusIconsVue.UserFilled,
  menu: ElementPlusIconsVue.Menu,
  tree: ElementPlusIconsVue.Folder,
  document: ElementPlusIconsVue.Document,
  bell: ElementPlusIconsVue.Bell,
  monitor: ElementPlusIconsVue.Monitor,
  clock: ElementPlusIconsVue.Clock,
  log: ElementPlusIconsVue.Tickets,
}

const getIconComponent = (iconName?: string): Component => {
  if (!iconName) return iconMap.default

  if (iconMap[iconName]) return iconMap[iconName]

  const iconComponent = ElementPlusIconsVue[iconName as keyof typeof ElementPlusIconsVue]
  if (iconComponent) return iconComponent as Component

  return iconMap.default
}

const getFullPath = (parentPath: string, childPath: string): string => {
  const cleanParentPath = parentPath.endsWith('/') ? parentPath.slice(0, -1) : parentPath
  const cleanChildPath = childPath.startsWith('/') ? childPath.slice(1) : childPath
  return `/${cleanParentPath}/${cleanChildPath}`
}

watch(
  () => routerStore.asyncRouters,
  () => {},
  { deep: true },
)

const toggleCollapse = () => {
  emit('toggle')
}

const handleMenuSelect = (index: string) => {
  if (index.startsWith('/')) {
    emit('menu-select', index)
  }
}

onMounted(() => {})
</script>

<style scoped>
/* ===== 侧边栏容器 ===== */
.sidebar-container {
  height: 100vh;
  overflow: visible;
  background: linear-gradient(180deg, var(--sidebar-bg) 0%, var(--sidebar-bg-secondary) 100%);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: var(--shadow-md);
  position: relative;
  z-index: 100;
  display: flex;
  flex-direction: column;
}

/* ===== Logo 区域 ===== */
.logo-section {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 64px;
  padding: 0 var(--space-lg);
  background: linear-gradient(135deg, var(--sidebar-bg) 0%, var(--sidebar-bg-secondary) 100%);
  border-bottom: 1px solid var(--border-color);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  z-index: 10;
}

.logo-section::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--primary-color), transparent);
  opacity: 0;
  transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-section:hover::after {
  opacity: 1;
}

.logo-image {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-section:hover .logo-image {
  transform: scale(1.05);
  box-shadow: var(--shadow-md);
}

.logo-text {
  margin-left: var(--space-md);
  font-size: 18px;
  font-weight: 600;
  color: var(--sidebar-text);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  white-space: nowrap;
  flex: 1;
  position: relative;
}

.logo-text::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 0;
  height: 2px;
  background: var(--primary-color);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-text:hover::after {
  width: 100%;
}

/* ===== 菜单容器 ===== */
.sidebar-menu {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px - 60px);
  background: transparent !important;
  border: none !important;
  padding: var(--space-sm) 0;
  overflow-y: auto;
  overflow-x: hidden;
  position: relative;
  z-index: 101;
}

.sidebar-menu:not(.el-menu--collapse) {
  box-shadow: 4px 0 12px rgba(0, 0, 0, 0.1);
}

/* ===== 菜单项通用样式 ===== */
.menu-item {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin: var(--space-xs) var(--space-sm);
  border-radius: var(--radius-md);
  position: relative;
  overflow: hidden;
}

.menu-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, var(--sidebar-hover-bg) 0%, transparent 100%);
  opacity: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  transform: translateY(10px);
}

.menu-item:hover::before {
  opacity: 1;
  transform: translateY(0);
}

.menu-item:hover {
  transform: translateX(4px) translateY(-2px);
}

.menu-icon {
  font-size: 18px !important;
  margin-right: var(--space-md) !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
  color: var(--sidebar-text);
}

.menu-item:hover .menu-icon {
  transform: scale(1.1);
  color: var(--primary-color);
}

.menu-text {
  font-weight: 500;
  color: var(--sidebar-text);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.menu-item:hover .menu-text {
  color: var(--primary-color);
}

/* ===== 激活状态的菜单项 ===== */
.menu-item.is-active {
  background: linear-gradient(
    135deg,
    var(--sidebar-active-bg) 0%,
    var(--sidebar-active-bg-secondary) 100%
  ) !important;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
  transform: translateX(4px);
}

.menu-item.is-active::before {
  opacity: 1;
  transform: translateY(0);
}

.menu-item.is-active .menu-icon {
  color: var(--sidebar-active-text);
  transform: scale(1.1);
}

.menu-item.is-active .menu-text {
  color: var(--sidebar-active-text);
  font-weight: 600;
}

/* ===== 子菜单样式 ===== */
:deep(.el-sub-menu) {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

:deep(.el-sub-menu .el-sub-menu__title) {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.el-sub-menu .el-sub-menu__title:hover) {
  transform: translateX(4px) translateY(-2px);
}

:deep(.el-sub-menu .el-menu) {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.05) 0%, transparent 100%) !important;
  border-radius: 0 0 var(--radius-md) var(--radius-md) !important;
  overflow: hidden;
}

:deep(.el-sub-menu .el-menu .el-menu-item) {
  padding-left: var(--space-12) !important;
}

:deep(.el-sub-menu .el-menu .el-menu-item:hover) {
  transform: translateX(2px);
}

/* ===== 菜单底部折叠按钮 ===== */
.menu-bottom {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--space-sm) 0;
  background: linear-gradient(0deg, var(--sidebar-bg) 0%, var(--sidebar-bg-secondary) 100%);
  border-top: 1px solid var(--border-color);
  position: relative;
  height: 50px;
  flex-shrink: 0;
  margin-top: auto;
}

.menu-bottom::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--primary-color), transparent);
  opacity: 0;
  transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.menu-bottom:hover::before {
  opacity: 1;
}

.collapse-btn {
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--sidebar-text);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  width: calc(100% - var(--space-md));
  height: 44px;
  padding: 0 var(--space-md);
  display: flex;
  justify-content: center;
  align-items: center;
  margin: 0 var(--space-sm);
}

.collapse-btn:hover {
  background: var(--sidebar-hover-bg);
  color: var(--primary-color);
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.collapse-btn:active {
  background: var(--sidebar-active-bg);
  color: var(--sidebar-active-text);
  transform: translateY(0);
}

.collapse-btn :deep(.el-button__icon) {
  margin: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-size: 18px;
}

.collapse-icon {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.collapse-btn:hover .collapse-icon {
  transform: scale(1.1);
}

/* ===== 折叠状态特殊处理 ===== */
:deep(.el-menu--collapse) .menu-item {
  justify-content: center !important;
  padding: 0 var(--space-md) !important;
  width: 64px !important;
}

:deep(.el-menu--collapse) .menu-text {
  opacity: 0;
  transform: translateX(-10px);
  display: none;
}

:deep(.el-menu--collapse) .menu-icon {
  margin-right: 0 !important;
  font-size: 20px !important;
}

:deep(.el-menu--collapse .el-sub-menu__icon-arrow) {
  display: none !important;
}

/* ===== 滚动条样式 ===== */
.sidebar-menu::-webkit-scrollbar {
  width: 6px;
}

.sidebar-menu::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-menu::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

.sidebar-menu::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.3);
}

/* ===== 暗黑模式 ===== */
html.dark .sidebar-container {
  background: linear-gradient(180deg, var(--sidebar-bg) 0%, var(--sidebar-bg-secondary) 100%);
  box-shadow: var(--shadow-md);
}

html.dark .logo-section {
  background: linear-gradient(135deg, var(--sidebar-bg) 0%, var(--sidebar-bg-secondary) 100%);
  border-bottom-color: var(--border-color);
}

html.dark .menu-bottom {
  background: linear-gradient(0deg, var(--sidebar-bg) 0%, var(--sidebar-bg-secondary) 100%);
  border-top-color: var(--border-color);
}

html.dark .menu-item:hover {
  background: var(--sidebar-hover-bg);
}

html.dark .menu-item.is-active {
  background: linear-gradient(
    135deg,
    var(--sidebar-active-bg) 0%,
    var(--sidebar-active-bg-secondary) 100%
  ) !important;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.4);
}

html.dark :deep(.el-sub-menu .el-menu) {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.08) 0%, transparent 100%) !important;
}

/* ===== 响应式设计 - 移动端 ===== */
@media (max-width: 768px) {
  .sidebar-container {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    transform: translateX(-100%);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .sidebar-container.show {
    transform: translateX(0);
  }

  .sidebar-container:not(.show) {
    width: 0 !important;
    overflow: hidden;
  }
}
</style>
