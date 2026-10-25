// 待支付项目明细页 → 下单支付完整流程
// 流程：
//  1. 加载明细：GET /api/wx/payment/:id
//  2. 点击「立即支付」：POST /api/wx/payment/unified-order {orderId} → 后端 V3 JSAPI 下单
//  3. wx.requestPayment 透传后端返回的调起参数（appId/timeStamp/nonceStr/package/signType=RSA/paySign）
//  4. 支付完成 → GET /api/wx/payment/:id?sync=true 向微信侧查单确认最终状态
const api = require('../../utils/api')
const config = require('../../utils/config')

Page({
  data: {
    id: '',
    detail: {},
    statusMap: {
      pending: '待支付',
      success: '已支付',
      failed: '支付失败',
      refunded: '已退款',
      cancelled: '已取消',
    },
    paying: false,
    // 支付结果
    showResult: false,
    paySuccess: false,
    resultMessage: '',
  },

  onLoad(options) {
    this.setData({ id: options.id || '' })
  },

  onShow() {
    const token = wx.getStorageSync(config.storageKeys.token)
    if (!token) {
      wx.reLaunch({ url: '/pages/login/index' })
      return
    }
    if (this.data.id) {
      this.loadDetail(false)
    }
  },

  /** 加载支付明细（sync=true 时向微信侧查单同步状态） */
  loadDetail(sync) {
    return api
      .getPaymentDetail(this.data.id, sync)
      .then((data) => {
        this.setData({ detail: data || {} })
      })
      .catch((err) => {
        wx.showToast({ title: err.message || '加载明细失败', icon: 'none' })
      })
  },

  /** 下单支付 */
  handlePay() {
    if (this.data.paying) return
    this.setData({ paying: true })

    // 1. V3 JSAPI 下单，获取调起支付参数
    api
      .unifiedOrder(this.data.detail.orderId)
      .then((data) => {
        const params = data.jsapiParams
        // 2. 小程序调起支付（参数由后端签名，signType=RSA，appId 与下单 appid 一致）
        return new Promise((resolve, reject) => {
          wx.requestPayment({
            appId: params.appId,
            timeStamp: params.timeStamp,
            nonceStr: params.nonceStr,
            package: params.package,
            signType: params.signType || 'RSA',
            paySign: params.paySign,
            success: () => resolve(),
            fail: (err) => reject(new Error(this.payErrMsg(err))),
          })
        })
      })
      .then(() => {
        // 3. 支付成功 → 查单确认最终状态
        return this.loadDetail(true)
      })
      .then(() => {
        const paid = this.data.detail.status === 'success'
        this.setData({
          paying: false,
          showResult: true,
          paySuccess: paid,
          resultMessage: paid ? '您的订单已支付成功' : '支付结果确认中，请稍后在列表中查看',
        })
      })
      .catch((err) => {
        this.setData({ paying: false })
        const msg = err.message || '支付失败'
        this.setData({
          showResult: true,
          paySuccess: false,
          resultMessage: msg,
        })
      })
  },

  /** 解析 wx.requestPayment 失败信息 */
  payErrMsg(err) {
    const errMsg = (err && err.errMsg) || ''
    if (errMsg.indexOf('cancel') > -1) {
      return '用户取消支付'
    }
    if (errMsg.indexOf('fail') > -1) {
      return '支付失败：' + errMsg
    }
    return '支付失败'
  },

  goBack() {
    wx.navigateBack({
      fail: () => wx.reLaunch({ url: '/pages/pending-list/index' }),
    })
  },

  closeResult() {
    const success = this.data.paySuccess
    this.setData({ showResult: false })
    if (success) {
      // 支付成功返回列表刷新
      setTimeout(() => {
        wx.navigateBack({
          fail: () => wx.reLaunch({ url: '/pages/pending-list/index' }),
        })
      }, 300)
    }
  },

  noop() {},
})
