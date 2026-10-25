<template>
  <div class="home-page">
    <!-- 未登录：静默授权跳转中 -->
    <div v-if="!loggedIn" class="loading-container">
      <div class="loading-spinner"></div>
      <p class="loading-text">{{ loadingMessage }}</p>
    </div>

    <!-- 已登录：首页（租户 UI 配置驱动） -->
    <div v-else class="home-main">
      <div class="brand" :style="{ background: primaryColor }">
        <img v-if="logoUrl" :src="logoUrl" class="brand-logo" alt="logo" />
        <div class="brand-title">{{ brandTitle }}</div>
        <div class="brand-user" v-if="nickname">您好，{{ nickname }}</div>
      </div>

      <div v-if="notice" class="notice">{{ notice }}</div>

      <!-- 图片轮播（ui.home.banners 驱动，兼容旧版单张 banner_url） -->
      <div
        v-if="banners.length"
        ref="bannerRef"
        class="banners"
        @scroll.passive="onBannerScroll"
        @touchstart.passive="stopBannerAutoplay"
        @touchend.passive="startBannerAutoplay"
      >
        <img
          v-for="(b, i) in banners"
          :key="i"
          :src="b.image_url"
          class="banner"
          alt="banner"
          @click="onBannerTap(b)"
        />
      </div>
      <div v-if="banners.length > 1" class="banner-dots">
        <span
          v-for="(b, i) in banners"
          :key="i"
          class="dot"
          :class="{ on: i === bannerActive }"
          :style="i === bannerActive ? { background: primaryColor } : undefined"
          @click="scrollToBanner(i)"
        ></span>
      </div>

      <!-- 宫格导航（ui.home.grid 驱动，空配置用默认项） -->
      <div class="grid">
        <div
          v-for="(item, idx) in gridItems"
          :key="idx"
          class="grid-item"
          @click="go(item.path)"
        >
          <div class="grid-icon">{{ item.icon }}</div>
          <div class="grid-title">{{ item.title }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { getPublicAppConfig, getPaymentList } from '@/api'
import { setTenantUI, getHomeGrid, type UIBanner } from '@/store/ui'

const router = useRouter()

// 应用实例标志：URL ?app_key= 优先，其次本地缓存（wxapp 空壳只配置此标志）
const APP_KEY_STORAGE = 'wx_app_key'
const AUTH_URL_STORAGE = 'wx_auth_url'
const AUTH_REDIRECT_STORAGE = 'wx_auth_redirect'

function resolveAppKey(): string {
  const fromURL = new URLSearchParams(window.location.search).get('app_key')
  if (fromURL) {
    localStorage.setItem(APP_KEY_STORAGE, fromURL)
    return fromURL
  }
  return localStorage.getItem(APP_KEY_STORAGE) || ''
}

const loggedIn = ref(false)
const loadingMessage = ref('正在加载...')
const appUI = ref<Record<string, unknown> | null>(null)
const nickname = ref(localStorage.getItem('wx_nickname') || '')
const pendingCount = ref(0)

// ---------- UI 配置读取（空值走默认） ----------
const uiHome = computed<Record<string, unknown>>(() => {
  return ((appUI.value?.home as Record<string, unknown>) || {}) as Record<string, unknown>
})
const brand = computed<Record<string, unknown>>(() => {
  return ((appUI.value?.brand as Record<string, unknown>) || {}) as Record<string, unknown>
})
const brandTitle = computed(() => (brand.value.title as string) || 'ixpay 支付')
const primaryColor = computed(() => (brand.value.primary_color as string) || '#07c160')
const logoUrl = computed(() => (brand.value.logo_url as string) || '')
const notice = computed(() => (uiHome.value.notice as string) || '')

// ---------- 轮播图（ui.home.banners 数组；兼容旧版单张 banner_url） ----------
const banners = computed<UIBanner[]>(() => {
  const list = uiHome.value.banners
  if (Array.isArray(list) && list.length > 0) {
    return list.filter((b) => b && (b as UIBanner).image_url) as UIBanner[]
  }
  const legacy = uiHome.value.banner_url as string | undefined
  return legacy ? [{ image_url: legacy }] : []
})
const bannerActive = ref(0)
const bannerRef = ref<HTMLElement | null>(null)
let bannerTimer: ReturnType<typeof setInterval> | null = null

// 自动轮播（单张不播；触摸时暂停，松手恢复）
function startBannerAutoplay() {
  stopBannerAutoplay()
  if (banners.value.length <= 1) return
  bannerTimer = setInterval(() => scrollToBanner((bannerActive.value + 1) % banners.value.length), 3000)
}
function stopBannerAutoplay() {
  if (bannerTimer) {
    clearInterval(bannerTimer)
    bannerTimer = null
  }
}
function scrollToBanner(i: number) {
  const el = bannerRef.value
  if (!el) return
  el.scrollTo({ left: i * el.clientWidth, behavior: 'smooth' })
}
// 滚动同步圆点指示
function onBannerScroll() {
  const el = bannerRef.value
  if (!el || el.clientWidth === 0) return
  bannerActive.value = Math.round(el.scrollLeft / el.clientWidth)
}
// 轮播图点击跳转（可选 link）
function onBannerTap(item: UIBanner) {
  if (item.link) router.push(item.link)
}

// ---------- 宫格导航（store 统一读取：ui.home.grid，空配置走默认） ----------
const gridItems = computed(() => getHomeGrid())

function go(path: string) {
  if (!path) return
  router.push(path)
}

// ---------- 启动逻辑 ----------
onMounted(async () => {
  const appKey = resolveAppKey()

  // 1. 加载公开应用配置（ui 驱动 + auth_url 缓存供无感授权使用）
  try {
    const app = await getPublicAppConfig(appKey)
    // 同步全局 UI 状态（含 ui 为空时清空旧缓存，TabBar/个人中心与首页保持一致）
    appUI.value = app.ui || null
    setTenantUI(app.ui || null)
    if (app.authUrl) localStorage.setItem(AUTH_URL_STORAGE, app.authUrl)
  } catch (e) {
    console.error('获取应用配置失败:', e)
  }

  // 2. 登录态检查
  const token = localStorage.getItem('wx_token')
  if (token) {
    // 已登录：渲染首页 + 拉待支付角标数
    loggedIn.value = true
    startBannerAutoplay()
    getPaymentList({ page: 1, pageSize: 1, status: 'pending' })
      .then((data) => {
        pendingCount.value = data.total
      })
      .catch(() => {})
    return
  }

  // 3. 未登录：跳转无感授权（snsapi_base，微信内无感；回调页回跳本页）
  await redirectToWechatAuth(appKey)
})

// 离开页面停止轮播定时器
onUnmounted(stopBannerAutoplay)

/**
 * 跳转无感授权（授权链接由后端按 app_key 生成）
 */
async function redirectToWechatAuth(appKey: string) {
  loadingMessage.value = '正在跳转至微信授权...'
  // 记录授权后回跳目标（授权落地页 /oauth/callback 读取）
  localStorage.setItem(AUTH_REDIRECT_STORAGE, '/')

  // 优先使用缓存的 auth_url（公开配置返回，含对应服务号 appid）
  const cached = localStorage.getItem(AUTH_URL_STORAGE)
  if (cached) {
    window.location.href = appendRedirect(cached, '/oauth/callback')
    return
  }

  try {
    if (appKey) {
      const app = await getPublicAppConfig(appKey)
      if (app.authUrl) {
        localStorage.setItem(AUTH_URL_STORAGE, app.authUrl)
        window.location.href = appendRedirect(app.authUrl, '/oauth/callback')
        return
      }
    }
  } catch (e) {
    console.error('获取应用配置失败，回退到默认授权入口:', e)
  }
  // 回退：无 app_key 或获取失败时走全局 authorize 端点
  const apiBase = import.meta.env.VITE_API_BASE_URL as string
  window.location.href = `${apiBase}/api/wx/auth/authorize?redirect_uri=${encodeURIComponent('/oauth/callback')}`
}

function appendRedirect(authURL: string, redirectURI: string) {
  const sep = authURL.includes('?') ? '&' : '?'
  return `${authURL}${sep}redirect_uri=${encodeURIComponent(redirectURI)}`
}
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: #f5f5f5;
  /* 底部 tab 栏占位 */
  padding-bottom: 64px;
}
.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  flex-direction: column;
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
.loading-text {
  font-size: 14px;
  color: #999;
}
.brand {
  padding: 40px 24px 32px;
  color: #fff;
}
.brand-logo {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  margin-bottom: 12px;
}
.brand-title {
  font-size: 22px;
  font-weight: 600;
}
.brand-user {
  margin-top: 6px;
  font-size: 13px;
  opacity: 0.85;
}
.notice {
  margin: 12px 16px 0;
  padding: 10px 14px;
  background: #fffbe6;
  border: 1px solid #ffe58f;
  border-radius: 8px;
  font-size: 13px;
  color: #ad6800;
}
/* 图片轮播：scroll-snap 横滑 + 自动播放 */
.banners {
  display: flex;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
  margin: 12px 16px 0;
  border-radius: 12px;
  scrollbar-width: none; /* Firefox 隐藏滚动条 */
}
.banners::-webkit-scrollbar {
  display: none; /* WebKit 隐藏滚动条 */
}
.banner {
  display: block;
  width: 100%;
  flex-shrink: 0;
  scroll-snap-align: center;
  object-fit: cover;
}
.banner-dots {
  display: flex;
  justify-content: center;
  gap: 6px;
  margin-top: 8px;
}
.dot {
  width: 6px;
  height: 6px;
  border-radius: 3px;
  background: #ddd;
  transition: all 0.2s;
}
.dot.on {
  width: 16px;
}
.grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  padding: 16px;
}
.grid-item {
  background: #fff;
  border-radius: 12px;
  padding: 20px 8px;
  text-align: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}
.grid-item:active {
  opacity: 0.8;
}
.grid-icon {
  font-size: 30px;
  margin-bottom: 8px;
}
.grid-title {
  font-size: 13px;
  color: #333;
}
</style>
