// 待支付列表：status=pending 过滤查询
// 点击列表项 → 待支付项目明细页；明细页内下单支付
const api = require('../../utils/api')
const config = require('../../utils/config')

Page({
  data: {
    list: [],
    loading: false,
    page: 1,
    pageSize: 10,
    total: 0,
    // 新增弹窗
    showModal: false,
    creating: false,
    form: { description: '', amount: '' },
  },

  onShow() {
    // 登录态校验：未登录跳登录页（点击授权登录）
    const token = wx.getStorageSync(config.storageKeys.token)
    if (!token) {
      wx.reLaunch({ url: '/pages/login/index' })
      return
    }
    this.loadList()
  },

  /** 下拉刷新 */
  onPullDownRefresh() {
    this.loadList().then(() => wx.stopPullDownRefresh())
  },

  /** 加载待支付列表 */
  loadList() {
    this.setData({ loading: true })
    return api
      .getPaymentList({ page: this.data.page, pageSize: this.data.pageSize, status: 'pending' })
      .then((data) => {
        this.setData({ list: data.list || [], total: data.total || 0, loading: false })
      })
      .catch((err) => {
        this.setData({ loading: false, list: [] })
        wx.showToast({ title: err.message || '加载失败', icon: 'none' })
      })
  },

  /** 进入明细页 */
  goDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/payment-detail/index?id=' + id })
  },

  /** 列表内快捷支付：跳明细页支付 */
  payItem(e) {
    this.goDetail(e)
  },

  /** 取消待支付项目 */
  cancelItem(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '取消支付',
      content: '确定取消该待支付项目吗？',
      success: (res) => {
        if (!res.confirm) return
        api
          .cancelPayment(id)
          .then(() => {
            wx.showToast({ title: '已取消', icon: 'success' })
            this.loadList()
          })
          .catch((err) => {
            wx.showToast({ title: err.message || '取消失败', icon: 'none' })
          })
      },
    })
  },

  /** 新增弹窗 */
  showCreateModal() {
    this.setData({ showModal: true, form: { description: '', amount: '' } })
  },

  hideCreateModal() {
    this.setData({ showModal: false })
  },

  noop() {},

  onDescInput(e) {
    this.setData({ 'form.description': e.detail.value })
  },

  onAmountInput(e) {
    this.setData({ 'form.amount': e.detail.value })
  },

  /** 创建待支付记录 */
  submitCreate() {
    const desc = (this.data.form.description || '').trim()
    const amount = parseFloat(this.data.form.amount)
    if (!desc) {
      wx.showToast({ title: '请输入项目名称', icon: 'none' })
      return
    }
    if (!amount || amount <= 0) {
      wx.showToast({ title: '请输入有效金额', icon: 'none' })
      return
    }

    this.setData({ creating: true })
    api
      .createPayment({ description: desc, amount: amount })
      .then(() => {
        this.setData({ creating: false, showModal: false })
        wx.showToast({ title: '创建成功', icon: 'success' })
        this.loadList()
      })
      .catch((err) => {
        this.setData({ creating: false })
        wx.showToast({ title: err.message || '创建失败', icon: 'none' })
      })
  },
})
