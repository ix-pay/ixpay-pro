<template>
  <div class="home-page">
    <div class="loading-container">
      <div class="loading-spinner"></div>
      <p class="loading-text">{{ loadingMessage }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const loadingMessage = ref('正在加载...')

onMounted(() => {
  // 检查是否已登录
  const token = localStorage.getItem('wx_token')
  if (token) {
    // 已登录，直接跳转到支付页面
    router.replace('/payment')
    return
  }

  // 未登录，直接跳转到后端授权端点
  // 后端会生成微信静默授权链接并重定向到微信
  // 授权后微信回调到后端，后端处理 code 后重定向回前端（使用 fragment 携带 token）
  redirectToWechatAuth()
})

/**
 * 跳转到微信 OAuth 静默授权页面
 * 直接跳转到后端 /api/wx/auth/authorize 端点，减少一次 API 请求
 * 后端将生成授权链接并 302 重定向到微信
 */
function redirectToWechatAuth() {
  loadingMessage.value = '正在跳转至微信授权...'

  const apiBase = import.meta.env.VITE_API_BASE_URL as string
  const authorizeURL = `${apiBase}/api/wx/auth/authorize?redirect_uri=/payment`

  window.location.href = authorizeURL
}
</script>

<style scoped>
.home-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f5f5f5;
}

.loading-container {
  text-align: center;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 16px;
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

.loading-text {
  font-size: 14px;
  color: #999;
}
</style>