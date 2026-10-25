// 网络请求封装：自动携带 token、401 自动刷新/无感重登并重试、统一错误处理
const config = require('./config')

// 防止并发刷新令牌/静默重登
let isRefreshing = false
let waitingQueue = []

/**
 * 将等待刷新的请求放入队列，刷新完成后统一放行
 */
function enqueueWaiting(callback) {
  waitingQueue.push(callback)
}

/**
 * 清空本地登录态并跳转到登录页
 */
function logoutAndRedirect() {
  wx.removeStorageSync(config.storageKeys.token)
  wx.removeStorageSync(config.storageKeys.refreshToken)
  wx.removeStorageSync(config.storageKeys.user)
  wx.reLaunch({ url: '/pages/login/index' })
}

/**
 * 刷新令牌（refresh_token 换新）
 */
function refreshAccessToken() {
  const refreshToken = wx.getStorageSync(config.storageKeys.refreshToken)
  if (!refreshToken) {
    return Promise.reject(new Error('缺少 refresh token'))
  }
  return new Promise((resolve, reject) => {
    wx.request({
      url: config.baseURL + '/api/wx/auth/refresh-token',
      method: 'POST',
      data: { refreshToken: refreshToken },
      header: { 'Content-Type': 'application/json' },
      success(res) {
        const body = res.data || {}
        if (res.statusCode === 200 && body.code === 0 && body.data) {
          wx.setStorageSync(config.storageKeys.token, body.data.accessToken)
          wx.setStorageSync(config.storageKeys.refreshToken, body.data.refreshToken)
          resolve(body.data.accessToken)
        } else {
          reject(new Error(body.msg || '刷新令牌失败'))
        }
      },
      fail(err) {
        reject(err)
      },
    })
  })
}

/**
 * 静默重登（无感授权）：wx.login 取 code → 后端 jscode2session 换新令牌
 * 微信能力静默无弹窗，token 失效时对用户完全无感
 */
function silentRelogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success(res) {
        if (!res.code) {
          reject(new Error('wx.login 未返回 code'))
          return
        }
        wx.request({
          url: config.baseURL + '/api/wx/auth/miniprogram-login',
          method: 'POST',
          data: { code: res.code, app_key: config.APP_KEY },
          header: { 'Content-Type': 'application/json' },
          success(r) {
            const body = r.data || {}
            if (r.statusCode === 200 && body.code === 0 && body.data) {
              wx.setStorageSync(config.storageKeys.token, body.data.accessToken)
              wx.setStorageSync(config.storageKeys.refreshToken, body.data.refreshToken)
              wx.setStorageSync(config.storageKeys.user, body.data.user || {})
              resolve(body.data.accessToken)
            } else {
              reject(new Error(body.msg || '静默重登失败'))
            }
          },
          fail(err) {
            reject(err)
          },
        })
      },
      fail(err) {
        reject(err)
      },
    })
  })
}

/**
 * 恢复登录态：先尝试 refresh_token 刷新，失败则静默重登兜底
 */
function restoreAuth() {
  return refreshAccessToken().catch(() => silentRelogin())
}

/**
 * 基础请求方法
 * @param {object} options { url, method, data, auth(是否需要登录态，默认 true) }
 * @returns {Promise<any>} 返回后端 data 字段
 */
function request(options) {
  const { url, method = 'GET', data = {}, auth = true } = options

  return new Promise((resolve, reject) => {
    const header = { 'Content-Type': 'application/json' }
    if (auth) {
      const token = wx.getStorageSync(config.storageKeys.token)
      if (token) {
        header.Authorization = 'Bearer ' + token
      }
    }

    wx.request({
      url: config.baseURL + url,
      method,
      data,
      header,
      success(res) {
        const body = res.data || {}

        // 401：恢复登录态（刷新/静默重登）后重试一次
        if (res.statusCode === 401 && auth) {
          if (isRefreshing) {
            enqueueWaiting((newToken) => {
              options.header = { 'Content-Type': 'application/json', Authorization: 'Bearer ' + newToken }
              request(options).then(resolve).catch(reject)
            })
            return
          }
          isRefreshing = true
          restoreAuth()
            .then((newToken) => {
              isRefreshing = false
              const queue = waitingQueue
              waitingQueue = []
              queue.forEach((cb) => cb(newToken))
              options.header = { 'Content-Type': 'application/json', Authorization: 'Bearer ' + newToken }
              return request(options).then(resolve).catch(reject)
            })
            .catch((err) => {
              isRefreshing = false
              waitingQueue = []
              logoutAndRedirect()
              reject(err)
            })
          return
        }

        // 业务响应：{ code, data, msg }
        if (res.statusCode >= 200 && res.statusCode < 300) {
          if (body.code === 0) {
            resolve(body.data)
          } else {
            reject(new Error(body.msg || '请求失败'))
          }
          return
        }

        reject(new Error((body && (body.msg || body.error)) || '请求错误(' + res.statusCode + ')'))
      },
      fail() {
        reject(new Error('网络错误，请检查网络连接'))
      },
    })
  })
}

module.exports = { request, logoutAndRedirect, restoreAuth }
