<template>
  <el-page-header class="box-border relative z-[10] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
    @back="handleBack">
    <template #content>
      <el-breadcrumb separator=">" v-if="breadcrumbList.length > 0"
        class="h-full flex items-center px-[var(--space-sm)] rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:bg-[var(--bg-secondary)]">
        <el-breadcrumb-item v-for="(item, index) in breadcrumbList" :key="index" :to="item.path">
          {{ item.name }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </template>
    <template #extra>
      <!-- 角色切换 -->
      <el-dropdown v-if="userRoles.length > 1" :hide-on-click="true" trigger="click">
        <div
          class="flex items-center gap-2 cursor-pointer px-2 py-1.5 h-full rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:bg-[var(--bg-secondary)] text-[13px] text-[var(--text-secondary)]">
          <el-icon class="text-[14px]">
            <User />
          </el-icon>
          <span class="font-medium">{{ currentRoleName }}</span>
          <el-icon class="text-[12px] opacity-60">
            <ArrowDown />
          </el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-for="role in userRoles" :key="role.id" :class="{
              'is-active': String(role.id) === String(currentRoleId),
              'is-disabled': String(role.id) === String(currentRoleId),
            }" @click="handleSwitchRole(role.id)">
              <div class="flex items-center gap-2">
                <span class="flex-1">{{ role.name }}</span>
                <el-icon v-if="String(role.id) === String(currentRoleId)" class="text-[var(--success-color)]">
                  <Check />
                </el-icon>
              </div>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- 消息通知 -->
      <el-dropdown :hide-on-click="false" trigger="click">
        <el-button type="text"
          class="transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] px-[var(--space-sm)] py-[var(--space-md)] rounded-[var(--radius-md)] hover:bg-[var(--bg-secondary)] hover:text-[var(--primary-color)]">
          <el-icon>
            <Bell />
          </el-icon>
          <el-badge v-if="notificationCount > 0" :value="notificationCount" type="danger" />
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <div class="flex justify-between items-center">
                <span>消息通知</span>
                <el-button type="text" size="small" @click="markAllAsRead">全部已读</el-button>
              </div>
            </el-dropdown-item>
            <el-dropdown-item v-if="notifications.length === 0" disabled>
              <div class="text-center py-4 text-[var(--text-secondary)]">暂无新消息</div>
            </el-dropdown-item>
            <el-dropdown-item v-for="(notification, index) in notifications" :key="index">
              <el-card shadow="hover">
                <div>
                  <div class="flex justify-between items-start">
                    <div>{{ notification.title }}</div>
                    <el-tag size="small" type="success">新</el-tag>
                  </div>
                  <div class="text-[13px] text-[var(--text-secondary)] my-1">
                    {{ notification.desc }}
                  </div>
                  <div class="text-[12px] text-[var(--text-placeholder)]">
                    {{ notification.time }}
                  </div>
                </div>
              </el-card>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- 主题切换 -->
      <div
        class="flex items-center h-full px-[var(--space-sm)] rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:bg-[var(--bg-secondary)]">
        <el-switch v-model="appStore.isDark" @change="handleThemeChange"
          class="align-middle transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]" :active-action-icon="Moon"
          :inactive-action-icon="Sunny" />
      </div>

      <!-- 全屏切换 -->
      <div
        class="flex items-center h-full px-[var(--space-sm)] rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] cursor-pointer hover:bg-[var(--bg-secondary)]"
        @click="toggleFullscreen" :title="isFullscreen ? '退出全屏' : '全屏查看'">
        <el-button type="text"
          class="text-lg text-[var(--text-primary)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:text-[var(--primary-color)]">
          <el-icon :class="{ 'rotate-180': isFullscreen }">
            <FullScreen />
          </el-icon>
        </el-button>
      </div>
      <!-- 用户信息 -->
      <el-dropdown :hide-on-click="false" trigger="click">
        <div
          class="flex items-center gap-2 cursor-pointer px-2 py-1.5 h-full rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:bg-[var(--bg-secondary)]">
          <el-avatar :size="32" :src="userAvatar" />
          <div class="flex flex-col justify-center h-full overflow-hidden">
            <div class="text-[13px] leading-[1.4] whitespace-nowrap overflow-hidden text-ellipsis">
              {{ userName }}
            </div>
          </div>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="handleProfile">
              <el-icon>
                <User />
              </el-icon>
              个人资料
            </el-dropdown-item>
            <el-dropdown-item @click="handleSettings">
              <el-icon>
                <Setting />
              </el-icon>
              系统设置
            </el-dropdown-item>
            <el-dropdown-item divided @click="handleLogout">
              <el-icon>
                <SwitchButton />
              </el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </template>
  </el-page-header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'

defineOptions({
  name: 'LayoutHeader',
})
import { useUserStore } from '@/stores/modules/user'
import { useAppStore } from '@/stores/modules/app'
import {
  Bell,
  User,
  Setting,
  SwitchButton,
  Sunny,
  Moon,
  FullScreen,
  ArrowDown,
  Check,
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { switchRole } from '@/api/modules/user'
import type { RoleInfo } from '@/types'

// 面包屑项接口
interface BreadcrumbItem {
  name?: string
  path: string
}

const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()

// Props
const props = defineProps<{
  isSidebarCollapsed: boolean
  breadcrumbList: BreadcrumbItem[]
}>()

// 用户信息
const userName = computed(() => userStore.userInfo?.nickname || '管理员')
const userAvatar = ref('')

// 角色切换相关
const userRoles = ref<RoleInfo[]>([])
const currentRoleId = ref<string | number>('')
const currentRoleName = computed(() => {
  if (!currentRoleId.value) {
    return userStore.userInfo?.role || '管理员'
  }
  const role = userRoles.value.find((r) => r.id === currentRoleId.value)
  return role?.name || userStore.userInfo?.role || '管理员'
})

// 全屏状态
const isFullscreen = ref(false)

// 通知相关
const notificationCount = ref(3)
const notifications = ref([
  {
    title: '新订单通知',
    desc: '您有一笔新订单需要处理',
    time: '10 分钟前',
  },
  {
    title: '系统更新提醒',
    desc: '系统将于今晚 23:00 进行维护更新',
    time: '1 小时前',
  },
  {
    title: '安全提示',
    desc: '您的账户在新设备上登录',
    time: '3 小时前',
  },
])

// 个人资料
const handleProfile = () => {
  router.push('/profile')
}

// 系统设置
const handleSettings = () => {
  router.push('/settings')
}

// 切换角色
const handleSwitchRole = async (roleId: string | number) => {
  // 如果当前已经是这个角色，不执行切换操作
  if (String(roleId) === String(currentRoleId.value)) {
    ElMessage.info('当前已经是该角色')
    return
  }

  try {
    // 使用 userStore.SwitchRole 方法
    const success = await userStore.SwitchRole(String(roleId))

    if (success) {
      // userStore.SwitchRole 中已经显示了提示，不需要重复显示
      // 手动更新 currentRoleId，确保下拉选项立即激活
      currentRoleId.value = String(roleId)
    } else {
      ElMessage.error('角色切换失败')
    }
  } catch (error) {
    console.error('切换角色失败:', error)
    ElMessage.error('角色切换失败，请重试')
  }
}

// 全部已读
const markAllAsRead = () => {
  notificationCount.value = 0
  notifications.value = []
  ElMessage.success('所有消息已标记为已读')
}

// 切换主题
const handleThemeChange = (value: string | number | boolean) => {
  appStore.isDark = Boolean(value)
}

// 切换全屏
const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    // 进入全屏
    document.documentElement.requestFullscreen().catch((err) => {
      ElMessage.error('无法进入全屏模式')
      console.error('全屏模式错误:', err)
    })
    isFullscreen.value = true
  } else {
    // 退出全屏
    document.exitFullscreen().catch((err) => {
      ElMessage.error('无法退出全屏模式')
      console.error('退出全屏错误:', err)
    })
    isFullscreen.value = false
  }
}

// 监听全屏状态变化
const handleFullscreenChange = () => {
  isFullscreen.value = !!document.fullscreenElement
}

// 返回上一页
const handleBack = () => {
  if (props.breadcrumbList.length > 1) {
    // 如果有面包屑，返回到上一个面包屑路径
    const lastIndex = props.breadcrumbList.length - 2
    router.push(props.breadcrumbList[lastIndex].path)
  } else {
    // 否则返回上一页
    router.back()
  }
}

// 退出登录
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '退出登录', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await userStore.LoginOut()
    router.push('/login')
    ElMessage.success('退出登录成功')
  } catch {
    // 用户取消操作
  }
}

// 初始化角色数据
const initUserRoles = () => {
  const roles = userStore.userInfo?.roles || []
  userRoles.value = roles

  console.log('Header - initUserRoles 开始执行')
  console.log('Header - userStore.userInfo:', userStore.userInfo)
  console.log('Header - userStore.userInfo.currentRoleId:', userStore.userInfo?.currentRoleId)
  console.log('Header - roles:', roles)
  console.log('Header - roles[0]?.id:', roles[0]?.id)

  // 优先使用后端返回的 currentRoleId
  // 注意：需要检查字段是否存在且不为空字符串、null、undefined
  const backendCurrentRoleId = userStore.userInfo?.currentRoleId
  console.log('Header - backendCurrentRoleId:', backendCurrentRoleId)
  console.log('Header - backendCurrentRoleId != null:', backendCurrentRoleId != null)
  console.log('Header - backendCurrentRoleId !== "":', backendCurrentRoleId !== '')

  if (backendCurrentRoleId != null && backendCurrentRoleId !== '') {
    currentRoleId.value = String(backendCurrentRoleId)
    console.log('Header - 使用后端 currentRoleId:', currentRoleId.value)
  } else if (roles.length > 0) {
    // 如果没有 currentRoleId，使用第一个角色的 id
    currentRoleId.value = roles[0].id
    console.log('Header - 使用第一个角色 ID:', currentRoleId.value)
  } else if (userStore.userInfo?.role) {
    // 如果只有 role 字符串，创建一个临时角色对象
    userRoles.value = [
      {
        id: 'default',
        name: userStore.userInfo.role,
        code: 'default',
      },
    ]
    currentRoleId.value = 'default'
    console.log('Header - 使用默认角色:', currentRoleId.value)
  }

  console.log('Header - 最终 currentRoleId:', currentRoleId.value)
  console.log('Header - 最终 userRoles:', userRoles.value)
  console.log(
    'Header - 激活状态检查：String(roles[0].id) === String(currentRoleId):',
    String(roles[0]?.id) === String(currentRoleId.value),
  )
}

// 页面加载时执行
onMounted(() => {
  // 初始化用户头像
  userAvatar.value = `https://api.dicebear.com/7.x/avataaars/svg?seed=${userName.value}`
  // 初始化用户角色
  initUserRoles()
  // 监听全屏状态变化
  document.addEventListener('fullscreenchange', handleFullscreenChange)
})

// 监听用户信息变化，自动更新角色数据
watch(
  () => userStore.userInfo,
  () => {
    initUserRoles()
  },
  { deep: true },
)
</script>

<style scoped>
/* 使用 Tailwind CSS 重构后，仅保留必要的 Element Plus 组件深度样式和响应式 */

:deep(.el-page-header) {
  background-color: var(--bg-primary);
  box-shadow: var(--shadow-md);
  padding: 0 var(--space-lg);
  border-bottom: 1px solid var(--border-primary);
  height: 100%;
  line-height: 84px;
  box-sizing: border-box;
  color: var(--text-primary);
  overflow: hidden;
  border-radius: 0;
  transition:
    background-color 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    color 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    border-bottom-color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.el-page-header__left) {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
  color: var(--text-primary);
  flex: 1;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.el-page-header__extra) {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  /* 8px - 使用固定值而非未定义的 CSS 变量 */
  color: var(--text-primary);
  height: 100%;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 角色切换下拉框激活样式 */
:deep(.el-dropdown-menu__item.is-active) {
  background-color: var(--primary-color);
  color: white;
}

:deep(.el-dropdown-menu__item.is-active:hover) {
  background-color: var(--primary-color);
  color: white;
}

/* 角色切换下拉框禁用样式（当前角色） */
:deep(.el-dropdown-menu__item.is-disabled) {
  cursor: not-allowed;
  opacity: 0.6;
}

:deep(.el-dropdown-menu__item.is-disabled:hover) {
  background-color: transparent;
}

/* 响应式设计 */
@media (max-width: 768px) {
  :deep(.el-page-header) {
    padding: 0 var(--space-md);
  }

  :deep(.el-page-header__left) {
    gap: var(--space-md);
  }

  :deep(.el-page-header__extra) {
    gap: 0.5rem;
    /* 8px */
  }

  :deep(.el-breadcrumb) {
    display: none;
  }
}

@media (max-width: 480px) {
  :deep(.el-page-header__extra) {
    gap: 0.25rem;
    /* 4px */
  }

  :deep(.el-dropdown:nth-last-child(1))>div {
    padding: 0 var(--space-xs);
    padding-top: var(--space-sm);
    padding-bottom: var(--space-sm);
  }

  :deep(.el-dropdown:nth-last-child(1)) .el-avatar {
    width: 1.75rem;
    height: 1.75rem;
  }

  :deep(.el-dropdown:nth-last-child(1)) .el-dropdown-menu {
    min-width: 160px;
  }
}

/* ==================== 角色切换 ==================== */
.role-switch {
  gap: 6px;
  padding: 6px 12px;
  background: transparent;
  border: none;

  .action-text {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
    transition: var(--transition);
  }

  .arrow-icon {
    font-size: 14px;
    opacity: 0.6;
    transition: var(--transition);
  }

  &:hover {
    .arrow-icon {
      opacity: 1;
      transform: translateY(2px);
    }
  }
}

/* ==================== 消息通知 ==================== */
.notification-btn {
  width: var(--action-btn-size);
  height: var(--action-btn-size);
  padding: 0;
  background: transparent;
  border: none;
  position: relative;

  .notification-badge {
    position: absolute;
    top: 4px;
    right: 4px;
  }
}

.notification-menu {
  min-width: 320px;
  max-width: 400px;
  padding: 0;

  .notification-header {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-secondary);

    .notification-header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .notification-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--text-primary);
    }

    .mark-read-btn {
      font-size: 13px;
      color: var(--primary-color);

      &:hover {
        color: var(--primary-hover);
      }
    }
  }

  .empty-notification {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 0;
    color: var(--text-secondary);

    .empty-icon {
      font-size: 48px;
      margin-bottom: 12px;
      opacity: 0.3;
    }

    .empty-text {
      font-size: 14px;
    }
  }

  .notification-item {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
    transition: var(--transition);

    &:last-child {
      border-bottom: none;
    }

    &:hover {
      background-color: var(--hover-bg);
    }

    .notification-item-content {
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .notification-item-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .notification-item-title {
      font-size: 14px;
      font-weight: 500;
      color: var(--text-primary);
    }

    .notification-tag {
      font-size: 12px;
    }

    .notification-item-desc {
      font-size: 13px;
      color: var(--text-secondary);
      margin: 0;
      line-height: 1.5;
    }

    .notification-item-time {
      font-size: 12px;
      color: var(--text-placeholder);
    }
  }
}

/* ==================== 主题切换 ==================== */
.theme-toggle {
  display: flex;
  align-items: center;
  justify-content: center;

  .theme-switch {
    --el-switch-on-color: #141414;
    --el-switch-off-color: #f5f5f5;

    :deep(.el-switch__core) {
      width: 50px;
      height: 24px;
      border-radius: 12px;
    }

    :deep(.el-switch__action) {
      width: 18px;
      height: 18px;
    }

    :deep(.el-switch__label) {
      font-size: 12px;
      color: var(--text-secondary);

      &.is-active {
        color: var(--primary-color);
      }
    }
  }
}

/* ==================== 全屏切换 ==================== */
.fullscreen-toggle {
  .fullscreen-icon {
    font-size: 20px;
    transition: transform 0.3s ease;

    &.is-fullscreen {
      transform: rotate(180deg);
    }
  }
}

/* ==================== 分隔线 ==================== */
.divider {
  width: 1px;
  height: 24px;
  margin: 0 8px;
  background-color: var(--border-color);
}

/* ==================== 用户信息 ==================== */
.user-dropdown {
  margin-left: 4px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: var(--radius);
  cursor: pointer;
  transition: var(--transition);

  &:hover {
    background-color: var(--hover-bg);

    .user-arrow {
      opacity: 1;
      transform: translateY(2px);
    }
  }

  .user-avatar {
    border: 2px solid var(--border-color);
    transition: var(--transition);

    &:hover {
      border-color: var(--primary-color);
    }
  }

  .user-details {
    display: flex;
    align-items: center;
    gap: 4px;

    .user-name {
      font-size: 14px;
      font-weight: 500;
      color: var(--text-primary);
      max-width: 120px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .user-arrow {
      font-size: 14px;
      opacity: 0.6;
      transition: var(--transition);
    }
  }
}

.user-menu {
  min-width: 160px;
  padding: 8px 0;

  .menu-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    transition: var(--transition);

    &:hover {
      background-color: var(--hover-bg);
    }

    .menu-icon {
      font-size: 18px;
      color: var(--text-secondary);
      transition: var(--transition);
    }

    &:hover .menu-icon {
      color: var(--primary-color);
    }

    &.logout-item {
      margin-top: 8px;
      padding-top: 10px;
      border-top: 1px solid var(--border-color);

      .menu-icon {
        color: var(--error-color, #f56c6c);
      }

      &:hover {
        background-color: var(--error-bg, #fef0f0);

        .menu-icon,
        span {
          color: var(--error-color, #f56c6c);
        }
      }
    }
  }
}

/* ==================== 下拉框通用样式 ==================== */
.action-dropdown {
  .dropdown-menu {
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    border: 1px solid var(--border-color);
    background-color: var(--bg-primary);
  }

  .dropdown-item-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .dropdown-item-text {
    flex: 1;
    font-size: 14px;
    color: var(--text-primary);
  }

  .check-icon {
    color: var(--success-color, #67c23a);
    font-size: 16px;
  }

  .is-active {
    background-color: var(--active-bg);
    color: #fff;

    .dropdown-item-text,
    .check-icon {
      color: #fff;
    }

    &:hover {
      background-color: var(--primary-hover);
    }
  }

  .is-disabled {
    cursor: not-allowed;
    opacity: 0.6;

    &:hover {
      background-color: transparent;
    }
  }
}

/* ==================== 暗黑模式适配 ==================== */
:deep(.dark) {
  .header-wrapper {
    --hover-bg: rgba(255, 255, 255, 0.05);
    --border-color: rgba(255, 255, 255, 0.1);
  }

  .notification-menu {
    background-color: var(--bg-secondary);
  }

  .user-menu {
    background-color: var(--bg-secondary);
  }
}

/* ==================== 响应式设计 ==================== */
@media (max-width: 768px) {
  .header-wrapper {
    --header-padding: 12px 16px;
  }

  .header-left {
    .breadcrumb {
      display: none;
    }
  }

  .header-right {
    gap: 2px;
  }

  .role-switch {
    .action-text {
      display: none;
    }
  }

  .user-info {
    .user-details {
      display: none;
    }
  }

  .notification-menu {
    min-width: 280px;
    max-width: 320px;
  }
}

@media (max-width: 480px) {
  .header-wrapper {
    --header-padding: 8px 12px;
    --action-btn-size: 36px;
  }

  .header-right {
    gap: 0;
  }

  .action-btn {
    padding: 6px;
  }

  .divider {
    margin: 0 4px;
  }

  .theme-toggle {
    .theme-switch {
      :deep(.el-switch__core) {
        width: 44px;
        height: 22px;
      }

      :deep(.el-switch__label) {
        display: none;
      }
    }
  }
}
</style>
