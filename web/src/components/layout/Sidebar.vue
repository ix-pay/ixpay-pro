<template>
  <el-aside
    :width="isCollapsed ? '64px' : '240px'"
    class="h-screen overflow-visible bg-[var(--sidebar-bg)] transition-[width] duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] shadow-[var(--shadow-md)] relative z-[100]"
  >
    <!-- Logo 区域 -->
    <div
      class="flex items-center justify-center px-[var(--space-lg)] h-16 bg-[var(--sidebar-bg)] border-b border-[var(--border-color)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] relative z-[10] after:content-[''] after:absolute after:bottom-0 after:left-0 after:right-0 after:h-[1px] after:bg-gradient-to-r after:from-transparent after:via-[var(--primary-color)] after:to-transparent after:opacity-0 after:transition-opacity after:duration-300 after:ease-[cubic-bezier(0.4,0,0.2,1)] hover:after:opacity-100"
    >
      <img
        :src="ixpayLogo"
        alt="IxPay Pro Logo"
        class="w-9 h-9 rounded-[var(--radius-md)] shadow-[var(--shadow-sm)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:scale-105 hover:shadow-[var(--shadow-md)]"
      />
      <span
        v-if="!isCollapsed"
        class="ml-[var(--space-md)] text-lg font-semibold text-[var(--sidebar-text)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] overflow-hidden whitespace-nowrap flex-1 relative after:content-[''] after:absolute after:-bottom-[2px] after:left-0 after:w-0 after:h-[2px] after:bg-[var(--primary-color)] after:transition-[width] after:duration-300 after:ease-[cubic-bezier(0.4,0,0.2,1)] hover:after:w-full"
      >
        IxPay Pro
      </span>
    </div>

    <!-- 菜单 -->
    <el-menu
      :default-active="currentRoute"
      class="border-r-0 border-t-0 border-b-0 h-[calc(100vh-64px-60px)] flex flex-col bg-[var(--sidebar-bg)] [--el-menu-text-color:var(--sidebar-text)] [--el-menu-hover-text-color:var(--sidebar-text)] [--el-menu-bg-color:var(--sidebar-bg)] [--el-menu-hover-bg-color:var(--sidebar-hover-bg)] [--el-menu-active-color:var(--sidebar-active-text)] [--el-menu-active-bg-color:var(--sidebar-active-bg)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] py-[var(--space-sm)] overflow-y-auto overflow-x-visible relative z-[101] not-[:is(.el-menu--collapse)]:shadow-[4px_0_12px_rgba(0,0,0,0.15)] not-[:is(.el-menu--collapse)]:border-r"
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
          <el-sub-menu v-if="menu.children && menu.children.length > 0" :index="String(menu.name)">
            <template #title>
              <el-icon>
                <component :is="getIconComponent(menu.meta.icon)" />
              </el-icon>
              <span>{{ menu.meta.title }}</span>
            </template>
            <!-- 递归渲染子菜单 -->
            <template v-for="subMenu in menu.children" :key="subMenu.id">
              <el-menu-item
                v-if="!subMenu.children || subMenu.children.length === 0"
                :index="getFullPath(menu.path, subMenu.path)"
              >
                <template #default>
                  <el-icon>
                    <component :is="getIconComponent(subMenu.meta.icon)" />
                  </el-icon>
                  <span>{{ subMenu.meta.title }}</span>
                </template>
              </el-menu-item>
              <!-- 三级子菜单 -->
              <el-sub-menu v-else :index="String(subMenu.name)">
                <template #title>
                  <el-icon>
                    <component :is="getIconComponent(subMenu.meta.icon)" />
                  </el-icon>
                  <span>{{ subMenu.meta.title }}</span>
                </template>
                <el-menu-item
                  v-for="thirdMenu in subMenu.children"
                  :key="thirdMenu.id"
                  :index="getFullPath(menu.path, getFullPath(subMenu.path, thirdMenu.path))"
                >
                  <template #default>
                    <el-icon>
                      <component :is="getIconComponent(thirdMenu.meta.icon)" />
                    </el-icon>
                    <span>{{ thirdMenu.meta.title }}</span>
                  </template>
                </el-menu-item>
              </el-sub-menu>
            </template>
          </el-sub-menu>
          <!-- 没有子菜单的情况 -->
          <el-menu-item v-else :index="menu.path.startsWith('/') ? menu.path : '/' + menu.path">
            <template #default>
              <el-icon>
                <component :is="getIconComponent(menu.meta.icon)" />
              </el-icon>
              <span>{{ menu.meta.title }}</span>
            </template>
          </el-menu-item>
        </template>
      </template>
      <!-- 加载状态 -->
      <template v-else-if="loading">
        <el-menu-item disabled>
          <template #title>
            <el-icon class="is-loading">
              <Loading />
            </el-icon>
            <span>加载菜单中...</span>
          </template>
        </el-menu-item>
      </template>
      <!-- 无菜单数据状态 -->
      <template v-else>
        <el-menu-item disabled>
          <template #title>
            <el-icon>
              <DocumentRemove />
            </el-icon>
            <span>暂无菜单数据</span>
          </template>
        </el-menu-item>
      </template>

      <!-- 菜单底部折叠按钮 -->
      <div
        class="flex justify-center items-center py-[var(--space-sm)] mt-auto bg-[var(--sidebar-bg)] border-t border-[var(--border-color)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] relative h-[50px] flex-shrink-0 before:content-[''] before:absolute before:top-0 before:left-0 before:right-0 before:h-[1px] before:bg-gradient-to-r before:from-transparent before:via-[var(--primary-color)] before:to-transparent before:opacity-0 before:transition-opacity before:duration-300 before:ease-[cubic-bezier(0.4,0,0.2,1)] hover:before:opacity-100"
      >
        <el-button
          type="text"
          size="small"
          class="bg-transparent border-none rounded-[var(--radius-md)] text-[var(--sidebar-text)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] w-[calc(100%-var(--space-md))] h-11 px-[var(--space-md)] flex justify-center items-center mx-[var(--space-sm)] hover:bg-[var(--sidebar-hover-bg)] hover:text-[var(--primary-color)] hover:-translate-y-0.5 active:bg-[var(--sidebar-active-bg)] active:text-[var(--sidebar-active-text)] active:translate-y-0 [&_.el-button__icon]:m-0 [&_.el-button__icon]:transition-transform [&_.el-button__icon]:duration-300 [&_.el-button__icon]:ease-[cubic-bezier(0.4,0,0.2,1)] [&_.el-button__icon]:text-lg"
          @click="toggleCollapse"
          :title="isCollapsed ? '展开菜单' : '收起菜单'"
        >
          <template #icon>
            <ArrowRight
              v-if="isCollapsed"
              class="transition-transform duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
            />
            <ArrowLeft
              v-else
              class="transition-transform duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
            />
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

// Props - 未使用，但保留以备后续扩展

const _props = defineProps({
  isCollapsed: {
    type: Boolean,
    default: false,
  },
})

// Emits
const emit = defineEmits(['toggle', 'menu-select'])

// 从routerStore获取菜单数据
const menuList = computed<ExtendedRouteRecordRaw[]>(() => {
  // 添加对asyncRouterFlag的依赖，确保菜单数据加载完成后会重新渲染
  // eslint-disable-next-line @typescript-eslint/no-unused-expressions
  routerStore.asyncRouterFlag

  // 创建一个包含默认首页的菜单列表
  const defaultIndexMenu: ExtendedRouteRecordRaw = {
    id: 1,
    name: 'index',
    path: '/index', // 使用绝对路径，避免导航错误
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

  // 直接使用asyncRouters中的数据，不再假设第一个元素是layout路由
  if (routerStore.asyncRouters.length > 0) {
    // 过滤掉隐藏的菜单和无效菜单
    const filteredMenus = routerStore.asyncRouters.filter((menu) => {
      // 确保菜单对象有效
      if (!menu || typeof menu !== 'object') {
        return false
      }

      // 确保meta对象存在
      if (!menu.meta) {
        menu.meta = { title: '' }
      }

      // 过滤掉隐藏的菜单
      if (menu.hidden || menu.meta.hidden) {
        return false
      }

      // 确保菜单有名称和标题
      if (!menu.name || !menu.meta.title) {
        return false
      }

      return true
    })

    // 处理首页菜单去重和路径修正
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
        // 确保首页菜单使用绝对路径
        if (menu.path === 'index') {
          menu.path = '/index'
        }
        indexMenuPaths.add('/index')
      }

      allMenus.push(menu)
    })
  }

  // 如果没有菜单数据，确保至少有首页菜单
  if (allMenus.length === 0) {
    allMenus.push(defaultIndexMenu)
  }

  return allMenus
})

// 加载状态
const loading = computed(() => {
  // 当asyncRouterFlag为0时，表示路由未加载
  return routerStore.asyncRouterFlag === 0
})

// 判断是否为移动设备（未使用，但保留以备后续扩展）

const _isMobile = computed(() => {
  return window.innerWidth <= 768
})

// 当前路由
const currentRoute = computed(() => {
  return route.path
})

// 图标映射表，将后端返回的图标字符串转换为 Element Plus 图标组件
// 后端统一使用驼峰命名（与 Element Plus 组件名一致）
const iconMap: Record<string, Component> = {
  // 默认图标
  default: House,
  // 常用图标（驼峰命名）
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
  // 兼容旧数据（小写格式）
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

// 获取图标组件
const getIconComponent = (iconName?: string): Component => {
  if (!iconName) return iconMap.default

  // 1. 尝试直接从图标映射表中获取
  if (iconMap[iconName]) return iconMap[iconName]

  // 2. 尝试直接从 ElementPlusIconsVue 中获取图标组件
  const iconComponent = ElementPlusIconsVue[iconName as keyof typeof ElementPlusIconsVue]
  if (iconComponent) return iconComponent as Component

  // 默认图标
  return iconMap.default
}

// 获取完整路径（拼接父路径和子路径）
const getFullPath = (parentPath: string, childPath: string): string => {
  // 确保父路径不以 / 结尾
  const cleanParentPath = parentPath.endsWith('/') ? parentPath.slice(0, -1) : parentPath
  // 确保子路径不以 / 开头
  const cleanChildPath = childPath.startsWith('/') ? childPath.slice(1) : childPath
  // 拼接路径
  return `/${cleanParentPath}/${cleanChildPath}`
}

// 监听菜单数据变化
watch(
  () => routerStore.asyncRouters,
  () => {
    // 菜单数据变化时无需特殊处理，computed会自动重新计算
  },
  { deep: true },
)

// 切换侧边栏
const toggleCollapse = () => {
  emit('toggle')
}

// 菜单选择处理
const handleMenuSelect = (index: string) => {
  if (index.startsWith('/')) {
    emit('menu-select', index)
  }
}

// 组件挂载时无需手动获取菜单数据，会从routerStore自动获取
onMounted(() => {
  // 组件挂载时无需特殊处理
})
</script>

<style scoped>
/* 基础样式 */
.el-aside {
  height: 100vh;
  overflow: visible;
  /* 允许菜单溢出 */
  background-color: var(--sidebar-bg);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: var(--shadow-md);
  z-index: 100;
  position: relative;
}

/* Logo容器 */
.logo-container {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  padding: 0 var(--space-lg);
  height: 64px;
  background-color: var(--sidebar-bg);
  border-bottom: 1px solid var(--border-color);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  z-index: 10;

  &::after {
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

  &:hover::after {
    opacity: 1;
  }
}

/* 菜单底部区域 */
.menu-bottom {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--space-sm) 0;
  margin-top: auto;
  background-color: var(--sidebar-bg);
  border-top: 1px solid var(--border-color);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  height: 50px;
  /* 固定高度 */
  flex-shrink: 0;
  /* 防止被压缩 */

  &::before {
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

  &:hover::before {
    opacity: 1;
  }
}

/* 折叠按钮 */
.collapse-btn {
  background-color: transparent;
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

  &:hover {
    background-color: var(--sidebar-hover-bg);
    color: var(--primary-color);
    transform: translateY(-1px);
  }

  &:active {
    background-color: var(--sidebar-active-bg);
    color: var(--sidebar-active-text);
    transform: translateY(0);
  }

  /* 去掉Element Plus默认的图标容器样式 */
  :deep(.el-button__icon) {
    margin: 0;
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    font-size: 18px;
  }
}

/* 图标过渡动画 */
.transition-icon {
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-image {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: var(--shadow-sm);

  &:hover {
    transform: scale(1.05);
    box-shadow: var(--shadow-md);
  }
}

/* Logo文本 */
.logo-text {
  margin-left: var(--space-md);
  font-size: 18px;
  font-weight: 600;
  color: var(--sidebar-text);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  white-space: nowrap;
  width: auto;
  flex: 1;
  position: relative;

  &::after {
    content: '';
    position: absolute;
    bottom: -2px;
    left: 0;
    width: 0;
    height: 2px;
    background-color: var(--primary-color);
    transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  &:hover::after {
    width: 100%;
  }
}

/* 菜单 */
:deep(.el-menu-vertical-demo) {
  border-right: none;
  border-top: none;
  border-bottom: none;
  height: calc(100vh - 64px - 60px);
  /* 减去 logo 区域和底部按钮区域 */
  display: flex;
  flex-direction: column;
  background-color: var(--sidebar-bg);
  --el-menu-text-color: var(--sidebar-text);
  --el-menu-hover-text-color: var(--sidebar-text);
  --el-menu-bg-color: var(--sidebar-bg);
  --el-menu-hover-bg-color: var(--sidebar-hover-bg);
  --el-menu-active-color: var(--sidebar-active-text);
  --el-menu-active-bg-color: var(--sidebar-active-bg);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  padding: var(--space-sm) 0;
  overflow-y: auto;
  overflow-x: visible;
  /* 允许横向溢出 */
  position: relative;
  z-index: 101;

  /* 当菜单展开时，添加浮动效果 */
  &:not(.el-menu--collapse) {
    box-shadow: 4px 0 12px rgba(0, 0, 0, 0.15);
    border-right: 1px solid rgba(0, 0, 0, 0.1);
  }
}

/* 自定义滚动条样式 */
:deep(.el-menu-vertical-demo::-webkit-scrollbar) {
  width: 6px;
}

:deep(.el-menu-vertical-demo::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.el-menu-vertical-demo::-webkit-scrollbar-thumb) {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 3px;

  &:hover {
    background: rgba(0, 0, 0, 0.3);
  }
}

/* 菜单项 */
:deep(.el-menu-item) {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

  &:hover {
    transform: translateX(4px);
  }

  &.is-active {
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
  }

  .el-icon {
    font-size: 18px !important;
    margin-right: var(--space-md) !important;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    flex-shrink: 0;
  }

  .el-menu-item__label {
    font-weight: 500;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

/* 子菜单项 */
:deep(.el-sub-menu) {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;

  .el-sub-menu__title {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

    &:hover {
      transform: translateX(4px);
    }

    .el-icon {
      font-size: 18px !important;
      margin-right: var(--space-md) !important;
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      flex-shrink: 0;
    }

    .el-sub-menu__title-text {
      font-weight: 500;
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .el-sub-menu__icon-arrow {
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      flex-shrink: 0;
    }
  }

  .el-menu {
    background-color: rgba(255, 255, 255, 0.05) !important;
    border-radius: 0 0 var(--radius-md) var(--radius-md) !important;
    overflow: hidden;

    .el-menu-item {
      padding-left: var(--space-xl) !important;

      &:hover {
        transform: translateX(2px);
      }
    }
  }
}

/* 菜单文本标签 */
:deep(.el-menu-item__label),
:deep(.el-sub-menu__title .el-sub-menu__title-text) {
  transition:
    opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  opacity: 1;
  transform: translateX(0);
}

/* 折叠状态下的菜单文本 */
:deep(.el-menu--collapse) .el-menu-item__label,
:deep(.el-menu--collapse) .el-sub-menu__title .el-sub-menu__title-text {
  transition:
    opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  opacity: 0;
  transform: translateX(-10px);
}

/* 确保折叠状态下图标仍然可见 */
:deep(.el-menu--collapse) .el-menu-item .el-icon,
:deep(.el-menu--collapse) .el-sub-menu__title .el-icon {
  display: flex !important;
  justify-content: center;
  align-items: center;
  margin: 0 !important;
  width: auto;
  height: auto;
  opacity: 1 !important;
  font-size: 20px !important;
}

/* 确保折叠状态下菜单项有足够的空间显示图标 */
:deep(.el-menu--collapse) .el-menu-item,
:deep(.el-menu--collapse) .el-sub-menu__title {
  justify-content: center !important;
  padding: 0 var(--space-md) !important;
  width: 64px;
  border-radius: var(--radius-md) !important;
}

/* 确保折叠状态下标题插槽内容正确布局 */
:deep(.el-menu--collapse) .el-menu-item__title,
:deep(.el-menu--collapse) .el-sub-menu__title {
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
  width: 100%;
}

/* 确保折叠状态下菜单项内容容器显示图标 */
:deep(.el-menu--collapse) .el-menu-item__content {
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
  width: 100%;
}

/* 确保折叠状态下菜单项的图标容器不被隐藏 */
:deep(.el-menu--collapse) .el-menu-item .el-menu-item__icon {
  display: flex !important;
  margin-right: 0 !important;
  width: auto;
  height: auto;
  opacity: 1 !important;
  font-size: 20px !important;
}

/* 确保折叠状态下工具提示触发器显示图标 */
:deep(.el-menu--collapse) .el-menu-tooltip__trigger {
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
}

/* 确保折叠状态下工具提示触发器中的图标可见 */
:deep(.el-menu--collapse) .el-menu-tooltip__trigger .el-icon {
  display: flex !important;
  margin-right: 0 !important;
  opacity: 1 !important;
  font-size: 20px !important;
}

/* 确保折叠状态下工具提示触发器中的标题内容可见 */
:deep(.el-menu--collapse) .el-menu-tooltip__trigger .el-menu-item__title,
:deep(.el-menu--collapse) .el-menu-tooltip__trigger .el-sub-menu__title {
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
}

/* 隐藏折叠状态下的子菜单箭头图标 */
:deep(.el-menu--collapse .el-sub-menu__icon-arrow) {
  display: none !important;
  width: 0 !important;
  height: 0 !important;
  padding: 0 !important;
  margin: 0 !important;
  visibility: hidden !important;
}

/* 暗黑模式样式 */
html.dark :deep(.el-aside) {
  background-color: var(--sidebar-bg);
  box-shadow: var(--shadow-md);
}

html.dark :deep(.logo-container) {
  background-color: var(--sidebar-bg);
  border-bottom-color: var(--border-color);
}

html.dark :deep(.menu-bottom) {
  border-top-color: var(--border-color);
}

html.dark :deep(.el-menu-vertical-demo) {
  background-color: var(--sidebar-bg);
  --el-menu-text-color: var(--sidebar-text);
  --el-menu-hover-text-color: var(--sidebar-text);
  --el-menu-bg-color: var(--sidebar-bg);
  --el-menu-hover-bg-color: var(--sidebar-hover-bg);
  --el-menu-active-color: var(--sidebar-active-text);
  --el-menu-active-bg-color: var(--sidebar-active-bg);
}

html.dark :deep(.el-menu-item:hover),
html.dark :deep(.el-sub-menu__title:hover) {
  background-color: var(--sidebar-hover-bg);
}

html.dark :deep(.el-menu-item.is-active) {
  background-color: var(--sidebar-active-bg);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

/* 暗黑模式样式 */
html.dark :deep(.el-aside) {
  background-color: var(--sidebar-bg);
  box-shadow: var(--shadow-md);
}

html.dark :deep(.logo-container) {
  background-color: var(--sidebar-bg);
  border-bottom-color: var(--border-color);
}

html.dark :deep(.menu-bottom) {
  border-top-color: var(--border-color);
}

html.dark :deep(.el-menu-vertical-demo) {
  background-color: var(--sidebar-bg);
  --el-menu-text-color: var(--sidebar-text);
  --el-menu-hover-text-color: var(--sidebar-text);
  --el-menu-bg-color: var(--sidebar-bg);
  --el-menu-hover-bg-color: var(--sidebar-hover-bg);
  --el-menu-active-color: var(--sidebar-active-text);
  --el-menu-active-bg-color: var(--sidebar-active-bg);
}

html.dark :deep(.el-menu-item:hover),
html.dark :deep(.el-sub-menu__title:hover) {
  background-color: var(--sidebar-hover-bg);
}

html.dark :deep(.el-menu-item.is-active) {
  background-color: var(--sidebar-active-bg);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}
</style>
