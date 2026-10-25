// 首页：租户 UI 配置驱动的宫格导航（菜单/图片/主题均来自管理平台配置）
const api = require('../../utils/api')
const config = require('../../utils/config')

// 默认宫格（无 UI 配置时兜底）
const DEFAULT_GRID = [
  { icon: '📋', title: '待支付列表', path: '/pages/pending-list/index' },
  { icon: '👤', title: '个人中心', path: '/pages/profile/index' },
]

Page({
  data: {
    loggedIn: false,
    logging: false,
    nickname: '',
    brandTitle: 'ixpay 支付',
    primaryColor: '#07c160',
    logoUrl: '',
    notice: '',
    gridItems: DEFAULT_GRID,
  },

  onLoad() {
    // 启动第一步：加载公开应用配置（ui 驱动首页渲染；无感校验 app_key）
    api
      .getPublicAppConfig()
      .then((app) => {
        if (app.ui) this.applyUI(app.ui)
      })
      .catch(() => {})
  },

  onShow() {
    const token = wx.getStorageSync(config.storageKeys.token)
    this.setData({
      loggedIn: !!token,
      nickname: (wx.getStorageSync(config.storageKeys.user) || {}).nickname || '',
    })
  },

  /** 未登录时自动尝试静默登录（wx.login 无感，失败停留首页可手动登录） */
  onReady() {
    if (!this.data.loggedIn) {
      // 静默重登一次（失败不提示，保持一键登录按钮）
      wx.login({
        success: (res) => {
          if (!res.code) return
          api
            .miniappLogin(res.code)
            .then((data) => {
              wx.setStorageSync(config.storageKeys.token, data.accessToken)
              wx.setStorageSync(config.storageKeys.refreshToken, data.refreshToken)
              wx.setStorageSync(config.storageKeys.user, data.user || {})
              this.setData({ loggedIn: true, nickname: (data.user || {}).nickname || '' })
            })
            .catch(() => {})
        },
      })
    }
  },

  /** 应用租户 UI 配置（管理平台可配，缺省走默认） */
  applyUI(ui) {
    const patch = {}
    if (ui.brand) {
      if (ui.brand.title) patch.brandTitle = ui.brand.title
      if (ui.brand.primary_color) patch.primaryColor = ui.brand.primary_color
      if (ui.brand.logo_url) patch.logoUrl = ui.brand.logo_url
    }
    if (ui.home) {
      if (ui.home.notice) patch.notice = ui.home.notice
      if (Array.isArray(ui.home.grid) && ui.home.grid.length > 0) {
        patch.gridItems = ui.home.grid.filter((g) => g.visible !== false)
      }
    }
    if (Object.keys(patch).length > 0) this.setData(patch)
  },

  /** 宫格跳转 */
  goGrid(e) {
    const path = e.currentTarget.dataset.path
    if (path) wx.navigateTo({ url: path, fail: () => wx.switchTab({ url: path, fail: () => {} }) })
  },

  /** 未登录时点击一键登录 */
  doLogin() {
    if (this.data.logging) return
    this.setData({ logging: true })
    wx.login({
      success: (res) => {
        if (!res.code) {
          this.setData({ logging: false })
          wx.showToast({ title: '登录失败', icon: 'none' })
          return
        }
        api
          .miniappLogin(res.code)
          .then((data) => {
            wx.setStorageSync(config.storageKeys.token, data.accessToken)
            wx.setStorageSync(config.storageKeys.refreshToken, data.refreshToken)
            wx.setStorageSync(config.storageKeys.user, data.user || {})
            this.setData({ logging: false, loggedIn: true, nickname: (data.user || {}).nickname || '' })
          })
          .catch((err) => {
            this.setData({ logging: false })
            wx.showToast({ title: err.message || '登录失败', icon: 'none' })
          })
      },
      fail: () => this.setData({ logging: false }),
    })
  },
})
