<template>
  <el-page-header class="box-border relative z-[10] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
    @back="handleBack">
    <template #content>
      <el-breadcrumb separator=">" v-if="breadcrumbList.length > 0"
        class="h-full flex items-center px-[var(--space-sm)] rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:bg-[var(--bg-hover)]">
        <el-breadcrumb-item v-for="(item, index) in breadcrumbList" :key="index" :to="item.path">
          {{ item.name }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </template>
    <template #extra>
      <!-- 角色切换 -->
      <el-dropdown v-if="userRoles.length > 1" :hide-on-click="true" trigger="click">
        <div
          class="flex items-center gap-2 cursor-pointer px-2 py-1.5 h-full rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:bg-[var(--bg-hover)] text-[13px] text-[var(--text-regular)]">
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
            <el-dropdown-item v-for="role in userRoles" :key="role.id"
              :class="{ 'is-active': String(role.id) === String(currentRoleId) }" @click="handleSwitchRole(role.id)">
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
          class="transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] px-[var(--space-sm)] py-[var(--space-md)] rounded-[var(--radius-md)] hover:bg-[var(--bg-hover)] hover:text-[var(--primary-color)]">
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
                  <div class="text-[13px] text-[var(--text-regular)] my-1">
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
        class="flex items-center h-full px-[var(--space-sm)] rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:bg-[var(--bg-hover)]">
        <el-switch v-model="appStore.isDark" @change="handleThemeChange"
          class="align-middle transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]" :active-action-icon="Moon"
          :inactive-action-icon="Sunny" />
      </div>

      <!-- 全屏切换 -->
      <div
        class="flex items-center h-full px-[var(--space-sm)] rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] cursor-pointer hover:bg-[var(--bg-hover)]"
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
          class="flex items-center gap-2 cursor-pointer px-2 py-1.5 h-full rounded-[var(--radius-md)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] hover:bg-[var(--bg-hover)]">
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
import { ref, computed, onMounted } from 'vue'

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
    return
  }

  try {
    // 使用 userStore.SwitchRole 方法
    const success = await userStore.SwitchRole(String(roleId))

    if (success) {
      ElMessage.success('角色切换成功')
      // 不需要刷新页面，userInfo 已自动更新
      // 不需要手动修改 userInfo.role，SwitchRole 已调用 GetUserInfo()
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

  // 优先使用后端返回的 currentRoleId
  const backendCurrentRoleId = userStore.userInfo?.currentRoleId
  if (backendCurrentRoleId && typeof backendCurrentRoleId !== 'boolean') {
    currentRoleId.value = String(backendCurrentRoleId)
  } else if (roles.length > 0) {
    // 如果没有 currentRoleId，使用第一个角色的 id
    currentRoleId.value = roles[0].id
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
  }
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
</script>

<style scoped>
/* 使用 Tailwind CSS 重构后，仅保留必要的 Element Plus 组件深度样式和响应式 */

:deep(.el-page-header) {
  @apply bg-[var(--bg-light)] shadow-[var(--shadow-md)] px-[var(--space-lg)] border-b border-[var(--border-color)] h-full leading-[84px] box-border text-[var(--text-primary)] overflow-hidden rounded-none;
  transition:
    background-color 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    color 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    border-bottom-color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.el-page-header__left) {
  @apply flex items-center gap-[var(--space-lg)] text-[var(--text-primary)] flex-1;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.el-page-header__extra) {
  @apply flex items-center gap-[var(--space-sm)] text-[var(--text-primary)] h-full overflow-hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 角色切换下拉框激活样式 */
:deep(.el-dropdown-menu__item.is-active) {
  @apply bg-[var(--primary-color)] text-white;
}

:deep(.el-dropdown-menu__item.is-active:hover) {
  @apply bg-[var(--primary-color)] text-white;
}

/* 响应式设计 */
@media (max-width: 768px) {
  :deep(.el-page-header) {
    @apply px-[var(--space-md)];
  }

  :deep(.el-page-header__left) {
    gap: var(--space-md);
  }

  :deep(.el-page-header__extra) {
    gap: var(--space-md);
  }

  :deep(.el-breadcrumb) {
    @apply hidden;
  }
}

@media (max-width: 480px) {
  :deep(.el-page-header__extra) {
    gap: var(--space-sm);
  }

  :deep(.el-dropdown:nth-last-child(1))>div {
    @apply px-[var(--space-xs)] py-[var(--space-sm)];
  }

  :deep(.el-dropdown:nth-last-child(1)) .el-avatar {
    @apply w-7 h-7;
  }

  :deep(.el-dropdown:nth-last-child(1)) .el-dropdown-menu {
    min-width: 160px;
  }
}
</style>
