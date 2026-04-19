import axios from 'axios' // 引入axios
import type {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig,
  AxiosError,
} from 'axios'
import { useUserStore } from '@/stores/index'
import { ElLoading, ElMessage } from 'element-plus'
import { emitter } from '@/utils/bus'
import router from '@/app/router/index'
import type { ApiResponse } from '@/types'
import { refreshToken } from '@/api/modules/user'

// 自定义请求配置接口
interface ShowLoadingOption {
  target?: string | HTMLElement | undefined
}

interface CustomAxiosRequestConfig<T = unknown> extends AxiosRequestConfig<T> {
  donNotShowLoading?: boolean
  loadingOption?: ShowLoadingOption
  _retry?: boolean
  skipAuthorization?: boolean
  _isRefreshToken?: boolean // 标记是否是刷新 token 的请求
}

// Loading管理类
class LoadingManager {
  private activeAxios = 0
  private timer: number | null = null
  private loadingInstance: ReturnType<typeof ElLoading.service> | null = null
  private isLoadingVisible = false
  private forceCloseTimer: number | null = null

  show(option: ShowLoadingOption = {}): void {
    const loadDom = document.getElementById('gva-base-load-dom')
    this.activeAxios++

    // 清除之前的定时器
    this.clearTimers()

    this.timer = window.setTimeout(() => {
      // 再次检查activeAxios状态，防止竞态条件
      if (this.activeAxios > 0 && !this.isLoadingVisible) {
        if (option.target === undefined) option.target = loadDom || undefined
        this.loadingInstance = ElLoading.service(option)
        this.isLoadingVisible = true

        // 设置强制关闭定时器，防止 loading 永远不关闭（30 秒超时）
        this.forceCloseTimer = window.setTimeout(() => {
          if (this.isLoadingVisible && this.loadingInstance) {
            console.warn('Loading 强制关闭：超时 30 秒')
            this.close()
          }
        }, 30000)
      }
    }, 400)
  }

  close(): void {
    this.activeAxios--
    if (this.activeAxios <= 0) {
      this.activeAxios = 0 // 确保不会变成负数
      this.clearTimers()

      if (this.isLoadingVisible && this.loadingInstance) {
        try {
          this.loadingInstance.close()
        } catch (e) {
          console.warn('关闭loading时出错:', e)
        }
        this.isLoadingVisible = false
        this.loadingInstance = null
      }
    }
  }

  reset(): void {
    this.activeAxios = 0
    this.isLoadingVisible = false
    this.clearTimers()

    if (this.loadingInstance) {
      try {
        this.loadingInstance.close()
      } catch (e) {
        console.warn('关闭loading时出错:', e)
      }
      this.loadingInstance = null
    }
  }

  private clearTimers(): void {
    if (this.timer) {
      clearTimeout(this.timer)
      this.timer = null
    }

    if (this.forceCloseTimer) {
      clearTimeout(this.forceCloseTimer)
      this.forceCloseTimer = null
    }
  }
}

// Token管理类
class TokenManager {
  private isRefreshing = false
  private isCheckingExpiry = false
  private refreshSubscribers: ((token: string) => void)[] = []
  private lastCheckTime = 0 // 上次检查token过期状态的时间戳（毫秒）
  private checkInterval = 60000 // 检查间隔：1分钟（毫秒）

  // 解析JWT token
  parseJwt(token: string): { exp?: number } {
    try {
      const base64Url = token.split('.')[1]
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
      const jsonPayload = decodeURIComponent(
        window
          .atob(base64)
          .split('')
          .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join(''),
      )
      return JSON.parse(jsonPayload)
    } catch (error) {
      console.error('解析JWT失败:', error)
      return {}
    }
  }

  // 检查token是否快过期（默认提前10分钟）
  isTokenExpiringSoon(token: string, thresholdMinutes: number = 10): boolean {
    if (!token) return false

    const decoded = this.parseJwt(token)
    if (!decoded.exp) {
      console.error('Token缺少exp字段')
      return false
    }

    const now = Date.now() / 1000 // 当前时间戳（秒）
    const expTime = decoded.exp // token过期时间戳（秒）
    const timeLeft = expTime - now // 剩余时间（秒）
    const thresholdSeconds = thresholdMinutes * 60 // 阈值时间（秒）

    // 当阈值为0时，检查token是否已过期
    if (thresholdMinutes === 0) {
      const isExpired = timeLeft <= 0
      return isExpired
    }

    // 当阈值大于0时，检查token是否快过期但尚未过期
    const isExpiringSoon = timeLeft > 0 && timeLeft < thresholdSeconds
    return isExpiringSoon
  }

  // 主动刷新token
  async proactiveRefresh(): Promise<void> {
    // 如果正在刷新或检查，直接返回
    if (this.isRefreshing || this.isCheckingExpiry) return

    // 检查时间间隔，避免频繁检查
    const now = Date.now()
    if (now - this.lastCheckTime < this.checkInterval) {
      return
    }

    this.isCheckingExpiry = true
    this.lastCheckTime = now // 更新上次检查时间

    try {
      const token = this.getToken()
      const refreshTokenFromStorage = localStorage.getItem('refreshToken')

      // 检查token是否快过期
      if (token && refreshTokenFromStorage && this.isTokenExpiringSoon(token)) {
        await this.refresh(refreshTokenFromStorage)
      }
    } catch (error) {
      console.error('检查token过期状态失败:', error)
    } finally {
      this.isCheckingExpiry = false
    }
  }

  // 刷新token
  async refresh(refreshTokenValue: string): Promise<string | null> {
    if (this.isRefreshing) {
      return new Promise((resolve) => {
        this.refreshSubscribers.push((token) => {
          resolve(token)
        })
      })
    }

    this.isRefreshing = true

    try {
      const res = await refreshToken({ refreshToken: refreshTokenValue })

      if (res.code === 0 && res.data?.accessToken) {
        const userStore = useUserStore()
        // 更新accessToken
        userStore.setToken(res.data.accessToken)

        // 如果返回了新的refreshToken，也更新它
        if (res.data.refreshToken) {
          localStorage.setItem('refreshToken', res.data.refreshToken)
        }

        // 通知所有等待的请求
        this.notifySubscribers(res.data.accessToken)
        return res.data.accessToken
      }

      return null
    } catch (error) {
      console.error('刷新token失败:', error)
      throw error
    } finally {
      this.isRefreshing = false
    }
  }

  // 获取token（优先从store获取，其次从localStorage获取）
  getToken(): string {
    // 优先从localStorage直接获取token，确保在任何情况下都能获取到
    let token = localStorage.getItem('token') || ''

    // 同时尝试使用store获取最新的token，确保数据同步
    try {
      const userStore = useUserStore()
      if (userStore.token) {
        token = userStore.token
      }
    } catch (error) {
      console.warn('获取用户信息失败:', error)
    }

    return token
  }

  // 通知所有等待的请求
  notifySubscribers(token: string): void {
    this.refreshSubscribers.forEach((callback) => callback(token))
    this.refreshSubscribers = []
  }

  // 处理未授权错误
  async handleUnauthorizedError(config: CustomAxiosRequestConfig): Promise<AxiosResponse> {
    const refreshTokenFromStorage = localStorage.getItem('refreshToken')
    const token = this.getToken()

    // 检查 token 是否已过期
    const isExpired = token ? this.isTokenExpiringSoon(token, 0) : true

    if (!refreshTokenFromStorage || isExpired) {
      // 没有 refreshToken 或 token 已过期，直接重新登录
      // 关闭加载动画
      if (!config.donNotShowLoading) {
        loadingManager.close()
      }
      this.redirectToLogin()
      throw new Error(isExpired ? 'Token 已过期' : '没有可用的 refresh token')
    }

    try {
      const newToken = await this.refresh(refreshTokenFromStorage)

      if (newToken) {
        // 重新设置请求头
        if (!config.headers) config.headers = {}
        config.headers['Authorization'] = `Bearer ${newToken}`

        // 重试请求 - 使用 axiosInstance 而不是 service，因为需要返回 AxiosResponse
        return axiosInstance(config)
      } else {
        // 刷新 token 失败，需要重新登录
        // 关闭加载动画
        if (!config.donNotShowLoading) {
          loadingManager.close()
        }
        this.redirectToLogin()
        throw new Error('刷新 token 失败')
      }
    } catch (error) {
      // 刷新 token 过程中出现错误，需要重新登录
      // 关闭加载动画
      if (!config.donNotShowLoading) {
        loadingManager.close()
      }
      this.redirectToLogin()
      throw error
    }
  }

  // 重定向到登录页
  redirectToLogin(): void {
    const userStore = useUserStore()
    userStore.ClearStorage()
    router.push({ name: 'Login', replace: true })
  }
}

// API响应处理工具
const responseUtils = {
  getErrorMessage(error: AxiosError): string {
    return ((error.response?.data as Record<string, unknown>)?.msg as string) || '请求失败'
  },

  handleResponse<T = unknown>(response: AxiosResponse): AxiosResponse<ApiResponse<T>> {
    const userStore = useUserStore()
    const customConfig = response.config as CustomAxiosRequestConfig
    if (!customConfig.donNotShowLoading) {
      loadingManager.close()
    }

    // 更新 token（如果响应头中包含新 token）
    if (response.headers && response.headers['new-token']) {
      userStore.setToken(response.headers['new-token'])
    }

    // 标准化响应格式：确保 response.data 始终是 ApiResponse 类型
    if (typeof (response.data as ApiResponse).code === 'undefined') {
      // 如果不是标准 API 响应格式，将其包装成 ApiResponse
      response.data = {
        code: 0,
        data: response.data as T,
        msg: '请求成功',
      } as ApiResponse<T>
    } else {
      // 处理 API 响应头中的消息
      if (response.headers && response.headers.msg) {
        ;(response.data as ApiResponse).msg = decodeURI(response.headers.msg)
      }
      // 确保 response.data 的类型正确
      response.data = response.data as ApiResponse<T>
    }

    // 处理响应：成功时返回数据内容，失败时根据情况抛出错误
    const apiResponse = response.data as ApiResponse<T>
    if (apiResponse.code !== 0) {
      // 请求失败，显示错误消息
      ElMessage({
        showClose: true,
        message:
          apiResponse.msg || (response.headers ? decodeURI(response.headers.msg) : '请求失败'),
        type: 'error',
      })

      // 对于特定的错误码（如登录失败或未授权），抛出错误以中断后续操作
      // code=1 通常表示登录失败或未授权
      if (apiResponse.code === 1) {
        throw new Error(apiResponse.msg || '请求失败')
      }
    }

    // 返回完整的 AxiosResponse 对象，但类型为 AxiosResponse<ApiResponse<T>>
    return response as AxiosResponse<ApiResponse<T>>
  },

  handleError(error: AxiosError): Promise<never> {
    if (error.config && !(error.config as CustomAxiosRequestConfig).donNotShowLoading) {
      loadingManager.close()
    }

    if (!error.response) {
      // 网络错误
      loadingManager.reset()
      emitter.emit('show-error', {
        code: 'network',
        message: this.getErrorMessage(error),
      })
      return Promise.reject(error)
    }

    return Promise.reject(error)
  },
}

// 创建axios实例
const axiosInstance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_BASE_API,
  timeout: 99999,
})

// 初始化管理类实例
const loadingManager = new LoadingManager()
const tokenManager = new TokenManager()

// 创建一个代理，将AxiosResponse转换为ApiResponse
type ApiService = Omit<
  AxiosInstance,
  'request' | 'get' | 'delete' | 'head' | 'options' | 'post' | 'put' | 'patch'
> & {
  <T = unknown>(config: CustomAxiosRequestConfig): Promise<ApiResponse<T>>
  request<T = unknown>(config: CustomAxiosRequestConfig): Promise<ApiResponse<T>>
  get<T = unknown>(url: string, config?: CustomAxiosRequestConfig): Promise<ApiResponse<T>>
  delete<T = unknown>(url: string, config?: CustomAxiosRequestConfig): Promise<ApiResponse<T>>
  head<T = unknown>(url: string, config?: CustomAxiosRequestConfig): Promise<ApiResponse<T>>
  options<T = unknown>(url: string, config?: CustomAxiosRequestConfig): Promise<ApiResponse<T>>
  post<T = unknown>(
    url: string,
    data?: unknown,
    config?: CustomAxiosRequestConfig,
  ): Promise<ApiResponse<T>>
  put<T = unknown>(
    url: string,
    data?: unknown,
    config?: CustomAxiosRequestConfig,
  ): Promise<ApiResponse<T>>
  patch<T = unknown>(
    url: string,
    data?: unknown,
    config?: CustomAxiosRequestConfig,
  ): Promise<ApiResponse<T>>
}

// 创建代理函数
function createApiService(instance: AxiosInstance): ApiService {
  const apiService = function apiService<T = unknown>(
    config: CustomAxiosRequestConfig,
  ): Promise<ApiResponse<T>> {
    return instance(config).then((response) => response.data as ApiResponse<T>)
  } as ApiService

  // 复制axios实例的所有属性到apiService
  Object.assign(apiService, instance)

  // 重写所有HTTP方法
  const methods: Array<
    keyof Pick<
      AxiosInstance,
      'request' | 'get' | 'delete' | 'head' | 'options' | 'post' | 'put' | 'patch'
    >
  > = ['request', 'get', 'delete', 'head', 'options', 'post', 'put', 'patch']

  methods.forEach((method) => {
    apiService[method] = function <T = unknown>(
      ...args: Parameters<AxiosInstance[typeof method]>
    ): Promise<ApiResponse<T>> {
      return (
        instance[method] as (
          ...args: Parameters<AxiosInstance[typeof method]>
        ) => Promise<AxiosResponse>
      )(...args).then((response: AxiosResponse) => response.data as ApiResponse<T>)
    }
  })

  return apiService
}

// 创建service实例
const service = createApiService(axiosInstance)

// http request 拦截器
axiosInstance.interceptors.request.use(
  async (config: InternalAxiosRequestConfig<unknown>) => {
    const customConfig = config as CustomAxiosRequestConfig

    // 显示loading
    if (!customConfig.donNotShowLoading) {
      loadingManager.show(customConfig.loadingOption)
    }

    // 获取token
    const token = tokenManager.getToken()

    // 主动检查并刷新快过期的token
    // 只在没有检查或刷新时才触发，避免多次触发
    if (token && !tokenManager['isCheckingExpiry'] && !tokenManager['isRefreshing']) {
      // 直接调用，不使用setTimeout，因为proactiveRefresh内部已经有异步处理
      tokenManager.proactiveRefresh()
    }

    // 设置请求头
    config.headers = {
      'Content-Type': 'application/json',
      // 根据后端中间件要求，使用Authorization头传递token，格式为Bearer {token}
      ...(!customConfig.skipAuthorization && token ? { Authorization: `Bearer ${token}` } : {}),
      ...config.headers,
    } as typeof config.headers

    return config
  },
  (error) => {
    if (error.config && !(error.config as CustomAxiosRequestConfig).donNotShowLoading) {
      loadingManager.close()
    }
    emitter.emit('show-error', {
      code: 'request',
      message: error.message || '请求发送失败',
    })
    return Promise.reject(error)
  },
)

// http response 拦截器
axiosInstance.interceptors.response.use(
  (response) => {
    // 处理响应并直接返回结果
    return responseUtils.handleResponse(response)
  },
  async (error) => {
    // 处理 401 未授权错误
    if (error.response && error.response.status === 401) {
      const config = error.config as CustomAxiosRequestConfig

      // 如果是刷新 token 的请求返回 401，说明 refreshToken 也过期了，直接跳转到登录页
      if (config._isRefreshToken) {
        // 关闭加载动画
        if (!config.donNotShowLoading) {
          loadingManager.close()
        }
        tokenManager.redirectToLogin()
        return Promise.reject(error)
      }

      // 如果已经重试过，直接跳转到登录页
      if (config._retry) {
        // 关闭加载动画
        if (!config.donNotShowLoading) {
          loadingManager.close()
        }
        tokenManager.redirectToLogin()
        return Promise.reject(error)
      }

      // 标记请求已重试
      config._retry = true

      try {
        return await tokenManager.handleUnauthorizedError(config)
      } catch (refreshError) {
        // 关闭加载动画
        if (!config.donNotShowLoading) {
          loadingManager.close()
        }
        return Promise.reject(refreshError)
      }
    }

    return responseUtils.handleError(error)
  },
)

// 监听页面卸载事件，确保loading被正确清理
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', () => loadingManager.reset())
  window.addEventListener('unload', () => loadingManager.reset())
}

// 导出原始service和包装后的request函数
export default service
export const resetLoading = () => loadingManager.reset()
export const notifyRefreshSubscribers = (token: string) => tokenManager.notifySubscribers(token)
