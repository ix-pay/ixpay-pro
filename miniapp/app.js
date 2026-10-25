// ixpay 微信小程序入口
// 登录态管理：token 存储于本地存储，由 utils/request.js 统一注入与刷新
App({
  onLaunch() {
    // 启动时不做阻塞式登录校验，各页面在 onShow 中校验登录态
    // 未登录的页面会自动跳转到登录页（点击授权登录）
  },

  globalData: {},
})
