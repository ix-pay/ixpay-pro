<template>
  <div class="profile-page">
    <!-- 上：页面标题（tab 页，无返回按钮） -->
    <div class="page-header">
      <h1 class="title">个人中心</h1>
    </div>

    <!-- 上：用户信息卡 -->
    <div class="card user-card">
      <div class="avatar">{{ avatarText }}</div>
      <div class="user-info">
        <div class="nickname">{{ nickname || '微信用户' }}</div>
        <div class="openid">ID: {{ userId || '-' }}</div>
      </div>
    </div>

    <!-- 中：功能菜单（ui.profile.menu 驱动，空配置用默认项） -->
    <div class="card menu-card">
      <div
        v-for="(item, idx) in menuItems"
        :key="idx"
        class="menu-item"
        @click="router.push(item.path)"
      >
        <span>
          <span v-if="item.icon" class="menu-icon">{{ item.icon }}</span>
          {{ item.title }}
        </span>
        <span class="arrow">›</span>
      </div>
    </div>

    <!-- 下：退出登录 -->
    <div class="btn-area">
      <button class="logout-btn" @click="onLogout">退出登录</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getPublicAppConfig } from '@/api'
import { getProfileMenu } from '@/store/ui'

const router = useRouter()
const APP_KEY_STORAGE = 'wx_app_key'

const nickname = ref(localStorage.getItem('wx_nickname') || '')
const userId = ref('')

// 个人中心菜单（租户 UI 配置 ui.profile.menu 驱动，空配置走默认）
const menuItems = computed(() => getProfileMenu())

const avatarText = computed(() => (nickname.value || '微').slice(0, 1))

async function onLogout() {
  if (!confirm('确定退出登录吗？退出后将重新进行微信授权。')) return
  // 清登录态 → 回首页触发静默重新授权（snsapi_base 无感）
  localStorage.removeItem('wx_token')
  localStorage.removeItem('wx_refresh_token')
  localStorage.removeItem('wx_nickname')

  // 重新走授权（优先缓存的 auth_url，失败回退全局 authorize）
  const appKey = localStorage.getItem(APP_KEY_STORAGE) || ''
  try {
    if (appKey) {
      const app = await getPublicAppConfig(appKey)
      if (app.authUrl) {
        localStorage.setItem('wx_auth_redirect', '/')
        window.location.href = app.authUrl
        return
      }
    }
  } catch {
    /* 回退 */
  }
  const apiBase = import.meta.env.VITE_API_BASE_URL as string
  window.location.href = `${apiBase}/api/wx/auth/authorize?redirect_uri=${encodeURIComponent('/')}`
}

onMounted(() => {
  // 登录接口返回的 user.id 存于本地（登录成功时写入）
  userId.value = localStorage.getItem('wx_user_id') || ''
  if (!nickname.value) nickname.value = localStorage.getItem('wx_nickname') || ''
  // 兜底：无 token 回首页重新授权
  if (!localStorage.getItem('wx_token')) router.replace('/')
})
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: #f5f5f5;
  /* 底部 tab 栏占位 */
  padding-bottom: 64px;
}
.page-header {
  display: flex;
  align-items: center;
  padding: 20px 16px;
}
.title {
  font-size: 18px;
  font-weight: 600;
}
.card {
  background: #fff;
  border-radius: 12px;
  margin: 0 16px 12px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}
.user-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 20px;
}
.avatar {
  width: 56px;
  height: 56px;
  line-height: 56px;
  text-align: center;
  border-radius: 50%;
  background: linear-gradient(135deg, #07c160, #06ad56);
  color: #fff;
  font-size: 24px;
  font-weight: 600;
}
.nickname {
  font-size: 17px;
  font-weight: 600;
}
.openid {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
.menu-card {
  padding: 0 16px;
}
.menu-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  font-size: 14px;
  border-bottom: 1px solid #f5f5f5;
}
.menu-item:last-child {
  border-bottom: none;
}
.menu-icon {
  margin-right: 6px;
}
.arrow {
  color: #ccc;
}
.btn-area {
  padding: 24px 16px;
}
.logout-btn {
  width: 100%;
  height: 44px;
  border: none;
  border-radius: 22px;
  background: #fff;
  color: #e64545;
  font-size: 15px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}
</style>
