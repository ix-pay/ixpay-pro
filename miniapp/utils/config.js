// 全局配置
// 说明：
// 1. 开发时在微信开发者工具中勾选「不校验合法域名」，并把 baseURL 改为后端实际地址
// 2. 线上环境必须是 HTTPS 域名，且需在小程序后台配置 request 合法域名
module.exports = {
  // 应用实例标志（唯一写死项）：管理平台「微信应用配置」中配置的 app_key
  // 不同标志请求同一后端接口即得到对应应用与商户的支付渠道（前端空壳模式）
  APP_KEY: 'mini-demo',

  // 后端 API 基础地址（项目域名，/api 反代到后端；本地调试可改为 http://localhost:8081）
  baseURL: 'https://ixpay.zcyxiaokeai.live',

  // 本地存储 key（与 wxapp 命名区分，小程序独立存储）
  storageKeys: {
    token: 'wx_token',
    refreshToken: 'wx_refresh_token',
    user: 'wx_user',
  },
}
