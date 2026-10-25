// 后端 API 接口定义（对应 server/internal/app/wx 模块）
const { request } = require('./request')
const config = require('./config')

/**
 * 获取公开应用配置：GET /api/wx/config/app?app_key=xxx
 * 小程序空壳启动时加载（无敏感信息），校验 app_key 并确认对应渠道
 * @param {string} appKey 应用实例标志（默认取 config.APP_KEY）
 */
function getPublicAppConfig(appKey) {
  return request({
    url: '/api/wx/config/app?app_key=' + encodeURIComponent(appKey || config.APP_KEY),
    auth: false,
  })
}

/**
 * 小程序授权登录：POST /api/wx/auth/miniprogram-login
 * @param {string} code wx.login() 获取的 js_code
 * @returns {Promise<{user:object, accessToken:string, refreshToken:string, accessExpire:number}>}
 */
function miniappLogin(code) {
  return request({
    url: '/api/wx/auth/miniprogram-login',
    method: 'POST',
    data: { code, app_key: config.APP_KEY },
    auth: false,
  })
}

/**
 * 创建待支付记录：POST /api/wx/payment
 * @param {{orderId?: string, amount: number, description: string}} params 金额单位：元
 */
function createPayment(params) {
  return request({
    url: '/api/wx/payment',
    method: 'POST',
    data: params,
  })
}

/**
 * 待支付列表：GET /api/wx/payment?status=pending
 * @param {{page?: number, pageSize?: number, status?: string}} params
 * @returns {Promise<{list: Array, total: number, page: number, pageSize: number}>}
 */
function getPaymentList(params) {
  const query = params || {}
  return request({
    url:
      '/api/wx/payment?page=' + (query.page || 1) +
      '&pageSize=' + (query.pageSize || 10) +
      '&status=' + (query.status || ''),
  })
}

/**
 * 支付详情：GET /api/wx/payment/:id
 * @param {string} id 支付 ID
 * @param {boolean} sync 是否同步微信侧订单状态（支付完成后确认结果用）
 */
function getPaymentDetail(id, sync) {
  return request({
    url: '/api/wx/payment/' + id + (sync ? '?sync=true' : ''),
  })
}

/**
 * 下单支付（V3 JSAPI Prepay）：POST /api/wx/payment/unified-order
 * @param {string} orderId 商户订单号
 * @returns {Promise<{prepayId: string, jsapiParams: {appId, timeStamp, nonceStr, package, signType, paySign}}>}
 */
function unifiedOrder(orderId) {
  return request({
    url: '/api/wx/payment/unified-order',
    method: 'POST',
    data: { orderId: orderId },
  })
}

/**
 * 取消支付：PUT /api/wx/payment/:id/cancel
 * @param {string} id 支付 ID
 */
function cancelPayment(id) {
  return request({
    url: '/api/wx/payment/' + id + '/cancel',
    method: 'PUT',
  })
}

module.exports = {
  getPublicAppConfig,
  miniappLogin,
  createPayment,
  getPaymentList,
  getPaymentDetail,
  unifiedOrder,
  cancelPayment,
}
