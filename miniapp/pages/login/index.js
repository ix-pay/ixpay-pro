// 登录页：点击授权登录
// 流程：wx.login() 静默获取 js_code → POST /api/wx/auth/miniprogram-login → 后端
// jscode2session 换取 openid → 签发本系统 JWT → 保存登录态 → 进入待支付列表
const api = require('../../utils/api')
const config = require('../../utils/config')

Page({
  data: {
    logging: false,
  },

  onLoad() {
    // 启动时加载应用配置（空壳校验 app_key，确认对应小程序与商户渠道；失败不阻断，登录时后端仍会校验）
    // 注意：配置详情（appKey/appId）不输出日志，避免泄露实例标识
    api.getPublicAppConfig().catch((err) => {
        console.warn('应用配置加载失败（请检查管理平台应用配置与 APP_KEY）:', err.message)
      })
  },

  onShow() {
    // 已登录则直接进入待支付列表
    const token = wx.getStorageSync(config.storageKeys.token)
    if (token) {
      wx.reLaunch({ url: '/pages/pending-list/index' })
    }
  },

  /** 点击授权登录 */
  handleLogin() {
    if (this.data.logging) return
    this.setData({ logging: true })

    wx.login({
      success: (res) => {
        if (!res.code) {
          this.failLogin('获取微信登录凭证失败')
          return
        }
        // code 换取本系统登录凭据
        api
          .miniappLogin(res.code)
          .then((data) => {
            wx.setStorageSync(config.storageKeys.token, data.accessToken)
            wx.setStorageSync(config.storageKeys.refreshToken, data.refreshToken)
            wx.setStorageSync(config.storageKeys.user, data.user || {})
            wx.showToast({ title: '登录成功', icon: 'success' })
            setTimeout(() => {
              wx.reLaunch({ url: '/pages/pending-list/index' })
            }, 500)
          })
          .catch((err) => {
            this.failLogin(err.message || '登录失败')
          })
      },
      fail: () => {
        this.failLogin('微信登录调用失败')
      },
    })
  },

  failLogin(msg) {
    this.setData({ logging: false })
    wx.showToast({ title: msg, icon: 'none' })
  },
})
