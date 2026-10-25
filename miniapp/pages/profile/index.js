// 个人中心：用户信息展示与退出登录（退出后回登录页重新授权）
const config = require('../../utils/config')

Page({
  data: {
    nickname: '',
    userId: '',
  },

  onShow() {
    if (!wx.getStorageSync(config.storageKeys.token)) {
      wx.reLaunch({ url: '/pages/login/index' })
      return
    }
    const user = wx.getStorageSync(config.storageKeys.user) || {}
    this.setData({
      nickname: user.nickname || '',
      userId: user.id || '',
    })
  },

  goPending() {
    wx.navigateTo({ url: '/pages/pending-list/index' })
  },

  onLogout() {
    wx.showModal({
      title: '退出登录',
      content: '确定退出登录吗？退出后将重新授权登录。',
      success: (res) => {
        if (!res.confirm) return
        wx.removeStorageSync(config.storageKeys.token)
        wx.removeStorageSync(config.storageKeys.refreshToken)
        wx.removeStorageSync(config.storageKeys.user)
        wx.reLaunch({ url: '/pages/login/index' })
      },
    })
  },
})
