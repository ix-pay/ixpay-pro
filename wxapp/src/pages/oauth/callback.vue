<template>
  <div class="callback-page">
    <div class="loading-spinner"></div>
    <p class="text">{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

// 授权回跳落地页：承接服务号网页授权回调（fragment 携带 token），回跳授权前页面
const router = useRouter()
const route = useRoute()
const message = ref('登录中...')

onMounted(() => {
  // URL 格式： /oauth/callback#token=xxx&refresh_token=yyy&expire=zzz
  const hash = window.location.hash
  const fragmentPart = hash.startsWith('#') ? hash.slice(1) : hash
  const params = new URLSearchParams(fragmentPart)
  const token = params.get('token')
  const refreshToken = params.get('refresh_token')

  if (token) {
    localStorage.setItem('wx_token', token)
    if (refreshToken) localStorage.setItem('wx_refresh_token', refreshToken)
    const user = params.get('user')
    if (user) localStorage.setItem('wx_nickname', user)
    const uid = params.get('uid')
    if (uid) localStorage.setItem('wx_user_id', uid)
    // 后端回传应用实例标志：扫码/公众号菜单直达的新用户凭此归属租户并加载 UI 配置
    const appKey = params.get('app_key')
    if (appKey) localStorage.setItem('wx_app_key', appKey)
  }

  // 回跳授权前页面（wx_auth_redirect 由 401 拦截器/首页写入）
  const redirect = localStorage.getItem('wx_auth_redirect') || (route.query.redirect as string) || '/'
  localStorage.removeItem('wx_auth_redirect')

  if (!token) {
    message.value = '登录失败，正在重试...'
    setTimeout(() => {
      window.location.href = `${import.meta.env.VITE_API_BASE_URL as string}/api/wx/auth/authorize?redirect_uri=${encodeURIComponent('/oauth/callback')}`
    }, 1000)
    return
  }

  message.value = '登录成功，正在跳转...'
  router.replace(redirect || '/')
})
</script>

<style scoped>
.callback-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  flex-direction: column;
  background: #f5f5f5;
}
.loading-spinner {
  width: 40px;
  height: 40px;
  margin-bottom: 16px;
  border: 3px solid #e0e0e0;
  border-top-color: #07c160;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.text {
  font-size: 14px;
  color: #999;
}
</style>
