import axios from 'axios'
import type { AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'

// 扩展 AxiosRequestConfig 类型，添加 _retry 标记
declare module 'axios' {
  interface InternalAxiosRequestConfig {
    _retry?: boolean
  }
}

// 创建 axios 实例
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL as string,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// ========== 令牌刷新状态管理 ==========
let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach((callback) => callback(token))
  refreshSubscribers = []
}

function addRefreshSubscriber(callback: (token: string) => void) {
  refreshSubscribers.push(callback)
}

// 请求拦截器：添加 token
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('wx_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

// 响应拦截器：统一错误处理，自动解包 data 字段，自动刷新令牌
request.interceptors.response.use(
  (response) => {
    const body = response.data
    // 后端返回 {code, data, msg} 结构，自动解包 data
    if (body.code === 0) {
      return body.data
    }
    // 业务错误，直接拒绝
    return Promise.reject(new Error(body.msg || '请求失败'))
  },
  async (error) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // 如果是 401 且不是刷新令牌请求本身，尝试刷新令牌
    if (error.response?.status === 401 && !originalRequest._retry) {
      const storedRefreshToken = localStorage.getItem('wx_refresh_token')
      if (!storedRefreshToken) {
        // 没有 refresh token：记录当前路由 → 无感重新授权（snsapi_base 微信内静默）
        silentReAuth()
        return Promise.reject(error)
      }

      // 防止并发刷新：如果正在刷新中，将请求排队等待
      if (isRefreshing) {
        return new Promise((resolve) => {
          addRefreshSubscriber((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(request(originalRequest))
          })
        })
      }

      isRefreshing = true
      originalRequest._retry = true

      try {
        // 使用 axios 直接发送刷新请求，避免拦截器循环
        const refreshResponse = await axios({
          baseURL: import.meta.env.VITE_API_BASE_URL as string,
          url: '/api/wx/auth/refresh-token',
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          data: { refreshToken: storedRefreshToken },
        })

        const body = refreshResponse.data
        if (body.code !== 0) {
          throw new Error(body.msg || '刷新令牌失败')
        }

        const result = body.data
        // 更新存储的令牌
        localStorage.setItem('wx_token', result.accessToken)
        localStorage.setItem('wx_refresh_token', result.refreshToken)

        // 通知所有等待的请求
        onTokenRefreshed(result.accessToken)

        // 重试原始请求
        originalRequest.headers.Authorization = `Bearer ${result.accessToken}`
        return request(originalRequest)
      } catch (refreshError) {
        // 刷新失败：清登录态 → 无感重新授权
        silentReAuth()
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    if (error.response) {
      const { status } = error.response
      switch (status) {
        case 500:
          console.error('服务器错误')
          break
        default:
          console.error(`请求错误: ${status}`)
      }
    } else {
      console.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  },
)

// ========== 无感授权（token 失效时 snsapi_base 静默重授权） ==========
// 记录当前路由，授权回调页(/oauth/callback)取回后恢复原页面
function silentReAuth() {
  localStorage.removeItem('wx_token')
  localStorage.removeItem('wx_refresh_token')

  const current = window.location.hash.startsWith('#')
    ? window.location.hash.slice(1)
    : '/'
  if (current && current !== '/oauth/callback') {
    localStorage.setItem('wx_auth_redirect', current)
  }

  // 优先使用公开配置缓存的授权链接（含对应服务号 appid）
  const cached = localStorage.getItem('wx_auth_url')
  if (cached) {
    window.location.href = appendRedirectURI(cached, '/oauth/callback')
    return
  }
  const apiBase = import.meta.env.VITE_API_BASE_URL as string
  window.location.href = `${apiBase}/api/wx/auth/authorize?redirect_uri=${encodeURIComponent('/oauth/callback')}`
}

function appendRedirectURI(authURL: string, redirectURI: string) {
  const sep = authURL.includes('?') ? '&' : '?'
  return `${authURL}${sep}redirect_uri=${encodeURIComponent(redirectURI)}`
}

// ========== API 接口定义 ==========

/** 微信登录 - 获取授权 code 后调用后端登录接口 */
export interface LoginParams {
  code: string
}

/** 用户信息 */
export interface UserInfo {
  id: string
  userName: string
  nickname: string
  email: string
  phone: string
  avatar: string
  role: string
  status: number
}

/** 登录响应 */
export interface LoginResult {
  user: UserInfo
  accessToken: string
  refreshToken: string
  accessExpire: number
}

/** 登录接口：POST /api/wx/auth/login */
export function wxLogin(params: LoginParams) {
  return request.post<unknown, LoginResult>('/api/wx/auth/login', params)
}

/** 公开应用配置（前端空壳启动时按 app_key 加载，无敏感信息） */
export interface WxPublicApp {
  appKey: string
  appType: 'mp' | 'miniapp'
  appId: string
  description: string
  authUrl?: string // 服务号授权登录入口链接（后端生成，含对应 appid）
  ui?: Record<string, unknown> | null // 租户个性化 UI 配置（菜单/图片/主题）
  status: number
}

/** 获取公开应用配置：GET /api/wx/config/app?app_key=xxx */
export function getPublicAppConfig(appKey: string) {
  return request.get<unknown, WxPublicApp>('/api/wx/config/app', {
    params: { app_key: appKey },
  })
}

/** 支付单（列表项） */
export interface PaymentRow {
  id: string
  orderId: string
  userId: string
  amount: number
  currency: string
  paymentMethod: string
  status: string
  transactionId: string
  description: string
  createdAt: string
  updatedAt: string
}

/** 待支付列表：GET /api/wx/payment?status=pending */
export function getPaymentList(params: { page?: number; pageSize?: number; status?: string }) {
  return request.get<unknown, { list: PaymentRow[]; total: number; page: number; pageSize: number }>(
    '/api/wx/payment',
    { params },
  )
}

/** 支付详情：GET /api/wx/payment/:id（sync=true 时向微信侧查单同步状态） */
export function getPaymentDetail(id: string, sync?: boolean) {
  return request.get<unknown, PaymentRow>(`/api/wx/payment/${id}`, {
    params: sync ? { sync: true } : undefined,
  })
}

/** 取消支付：PUT /api/wx/payment/:id/cancel */
export function cancelPayment(id: string) {
  return request.put<unknown, PaymentRow>(`/api/wx/payment/${id}/cancel`)
}

/** 创建待支付记录：POST /api/wx/payment */
export function createPayment(params: { orderId?: string; amount: number; description: string }) {
  return request.post<unknown, PaymentRow>('/api/wx/payment', params)
}

/** 刷新令牌响应 */
export interface RefreshTokenResult {
  accessToken: string
  refreshToken: string
  accessExpire: number
}

/** 刷新令牌：POST /api/wx/auth/refresh-token */
export function refreshToken(refreshTokenValue: string) {
  return request.post<unknown, RefreshTokenResult>('/api/wx/auth/refresh-token', {
    refreshToken: refreshTokenValue,
  })
}

/** 获取微信 OAuth 授权链接参数 */
export interface OAuthURLParams {
  redirect_uri: string
  state?: string
}

/** OAuth 授权链接响应 */
export interface OAuthURLResult {
  oauthUrl: string
  state?: string
}

/** 获取微信 OAuth 静默授权链接：GET /api/wx/auth/oauth-url */
export function getOAuthURL(params: OAuthURLParams) {
  return request.get<unknown, OAuthURLResult>('/api/wx/auth/oauth-url', {
    params,
  })
}

/** 创建支付订单参数 */
export interface UnifiedOrderParams {
  amount: number
  description: string
  orderId: string
}

/** JSAPI 支付参数（后端返回的嵌套对象） */
export interface JSAPIParams {
  appId: string
  timeStamp: string
  nonceStr: string
  package: string
  signType: string
  paySign: string
}

/** 支付订单结果（后端返回的统一下单响应结构） */
export interface UnifiedOrderResult {
  prepayId: string
  jsapiParams: JSAPIParams
}

/** 创建支付订单：POST /api/wx/pay/unified-order */
export function createUnifiedOrder(params: UnifiedOrderParams) {
  return request.post<unknown, UnifiedOrderResult>(
    '/api/wx/pay/unified-order',
    params,
  )
}

/** 微信 JS-SDK 配置结果 */
export interface JSAPIConfigResult {
  appId: string
  timestamp: string
  nonceStr: string
  signature: string
}

/** 获取微信 JS-SDK 配置：GET /api/wx/pay/jsapi-config（url 为当前页面地址，不含 # 及其后部分） */
export function getJSAPIConfig(url: string) {
  return request.get<unknown, JSAPIConfigResult>('/api/wx/pay/jsapi-config', {
    params: { url },
  })
}

/** 支付通知回调参数 */
export interface NotifyParams {
  outTradeNo: string
  transactionId: string
  totalFee: number
  timeEnd: string
  openid: string
}

/** 支付通知回调：POST /api/wx/pay/notify */
export function notifyPayment(params: NotifyParams) {
  return request.post<unknown, { code: number; message: string }>(
    '/api/wx/pay/notify',
    params,
  )
}

export default request