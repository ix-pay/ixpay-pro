<template>
  <div class="payment-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="header-title">订单确认</h1>
      <p class="header-desc">请确认订单信息后进行支付</p>
    </div>

    <!-- 订单信息卡片 -->
    <div class="order-card">
      <div class="order-header">
        <span class="order-label">订单编号</span>
        <span class="order-id">{{ order.orderId }}</span>
      </div>

      <div class="order-body">
        <div class="order-item">
          <span class="item-label">商品名称</span>
          <span class="item-value">{{ order.description }}</span>
        </div>
        <div class="order-item">
          <span class="item-label">订单金额</span>
          <span class="item-value price">¥{{ (order.amount / 100).toFixed(2) }}</span>
        </div>
        <div class="order-item">
          <span class="item-label">订单状态</span>
          <span class="item-value status" :class="order.status">
            {{ statusMap[order.status] }}
          </span>
        </div>
      </div>
    </div>

    <!-- 支付方式 -->
    <div class="payment-method">
      <div class="method-header">支付方式</div>
      <div class="method-item">
        <div class="method-left">
          <svg class="method-icon" viewBox="0 0 24 24" width="24" height="24">
            <path
              fill="#07c160"
              d="M8.5 11.5c-.83 0-1.5.67-1.5 1.5s.67 1.5 1.5 1.5S10 13.83 10 13s-.67-1.5-1.5-1.5zm7 0c-.83 0-1.5.67-1.5 1.5s.67 1.5 1.5 1.5S17 13.83 17 13s-.67-1.5-1.5-1.5zM12 2C6.48 2 2 6.48 2 12c0 1.82.49 3.53 1.34 5L2 22l5.34-1.34C8.82 21.49 10.35 22 12 22c5.52 0 10-4.48 10-10S17.52 2 12 2zm0 18c-1.46 0-2.86-.35-4.1-.97l-.62-.31-2.4.63.63-2.36-.33-.64A7.87 7.87 0 014 12c0-4.41 3.59-8 8-8s8 3.59 8 8-3.59 8-8 8z"
            />
          </svg>
          <span>微信支付</span>
        </div>
        <span class="method-checked">✓</span>
      </div>
    </div>

    <!-- 支付按钮 -->
    <div class="btn-area">
      <button
        class="pay-btn"
        :disabled="paying || paid"
        @click="handlePayment"
      >
        {{ paying ? '支付处理中...' : paid ? '已支付' : '确认支付' }}
      </button>
    </div>

    <!-- 底部提示 -->
    <p class="footer-tips">
      支付即表示同意 <a href="#" class="link">《支付服务协议》</a>
    </p>

    <!-- 支付结果弹窗 -->
    <transition name="fade">
      <div v-if="showResult" class="result-overlay" @click.self="closeResult">
        <div class="result-card" :class="paySuccess ? 'success' : 'fail'">
          <div class="result-icon">
            {{ paySuccess ? '✓' : '✕' }}
          </div>
          <h3 class="result-title">
            {{ paySuccess ? '支付成功' : '支付失败' }}
          </h3>
          <p class="result-desc">
            {{ paySuccess ? '您的订单已支付成功' : resultMessage }}
          </p>
          <button class="result-btn" @click="closeResult">
            {{ paySuccess ? '完成' : '关闭' }}
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { createUnifiedOrder, getJSAPIConfig, type JSAPIParams } from '@/api'
import wx from 'weixin-js-sdk'

const router = useRouter()
const route = useRoute()

// 订单状态映射
const statusMap: Record<string, string> = {
  pending: '待支付',
  paid: '已支付',
  cancelled: '已取消',
}

// 订单信息
const order = reactive({
  orderId: 'ORD' + Date.now().toString().slice(-8),
  description: 'ixpay 商品购买',
  amount: 100, // 分
  status: 'pending' as string,
})

const paying = ref(false)
const paid = ref(false)
const showResult = ref(false)
const paySuccess = ref(false)
const resultMessage = ref('')
// 存储后端返回的 JSAPI 支付参数，供 JS-SDK 和 WeixinJSBridge 兜底使用
const currentJSAPIParams = ref<JSAPIParams | null>(null)

onMounted(() => {
  // 处理从微信 OAuth 回调重定向带来的 token（URL fragment 方式传递）
  // URL 格式: /#/payment#token=xxx&refresh_token=yyy&expire=zzz
  const hash = window.location.hash
  if (hash.includes('#token=')) {
    // 提取第二个 # 后面的参数部分
    const fragmentPart = hash.split('#')[2] || ''
    const fragmentParams = new URLSearchParams(fragmentPart)
    const urlToken = fragmentParams.get('token')
    const urlRefreshToken = fragmentParams.get('refresh_token')
    if (urlToken) {
      localStorage.setItem('wx_token', urlToken)
    }
    if (urlRefreshToken) {
      localStorage.setItem('wx_refresh_token', urlRefreshToken)
    }
    // 清除 URL 中的 fragment 参数，保留路由路径
    // 提取第一段 hash（路由路径），如 /payment
    const routePath = hash.split('#')[1]?.split('?')[0] || '/payment'
    window.history.replaceState({}, document.title, window.location.pathname + '#' + routePath)
  }

  // 检查是否已登录
  const token = localStorage.getItem('wx_token')
  if (!token) {
    // 未登录，跳转到首页重新授权
    router.replace('/')
    return
  }

  // 从后端获取微信 JS-SDK 配置
  loadJSAPIConfig()
})

/** 从后端获取微信 JS-SDK 配置并初始化 wx.config() */
async function loadJSAPIConfig() {
  try {
    const config = await getJSAPIConfig()
    wx.config({
      debug: false,
      appId: config.appId,
      timestamp: config.timestamp,
      nonceStr: config.nonceStr,
      signature: config.signature,
      jsApiList: ['chooseWXPay'],
    })

    wx.ready(() => {
      console.log('微信 JS-SDK 就绪')
    })

    wx.error((err: any) => {
      console.error('微信 JS-SDK 配置错误:', err)
    })
  } catch (error) {
    console.error('获取微信 JS-SDK 配置失败:', error)
  }
}

/** 处理支付 */
async function handlePayment() {
  if (paying.value || paid.value) return

  paying.value = true

  try {
    // 调用后端创建支付订单，后端从 JWT 认证中获取用户 openid
    const payResult = await createUnifiedOrder({
      amount: order.amount,
      description: order.description,
      order_id: order.orderId,
    })

    // 使用后端返回的 JSAPI 参数调起微信支付（公众号内）
    currentJSAPIParams.value = payResult.jsapiParams
    await invokeWxPay(payResult.jsapiParams)
  } catch (error: any) {
    // 方式二：如果 JS-SDK 不可用，尝试 WeixinJSBridge（H5 支付）
    const fallback = await tryWeixinJSBridge()
    if (!fallback) {
      const errMsg = error?.response?.data?.message || error?.message || '支付失败'
      showPaymentResult(false, errMsg)
    }
  } finally {
    paying.value = false
  }
}

/** 使用微信 JS-SDK 调起支付 */
function invokeWxPay(params: JSAPIParams): Promise<void> {
  return new Promise((resolve, reject) => {
    wx.chooseWXPay({
      timestamp: params.timeStamp,
      nonceStr: params.nonceStr,
      package: params.package,
      signType: params.signType || 'MD5',
      paySign: params.paySign,
      success: () => {
        order.status = 'paid'
        paid.value = true
        showPaymentResult(true)
        resolve()
      },
      fail: (err: any) => {
        reject(new Error(err?.errMsg || '支付失败'))
      },
      cancel: () => {
        reject(new Error('用户取消支付'))
      },
    })
  })
}

/** 使用 WeixinJSBridge 调起支付（H5 支付兜底方案） */
function tryWeixinJSBridge(): Promise<boolean> {
  return new Promise((resolve) => {
    if (typeof window.WeixinJSBridge === 'undefined') {
      // WeixinJSBridge 未就绪，监听事件
      document.addEventListener(
        'WeixinJSBridgeReady',
        () => {
          doWeixinJSBridgePay(resolve)
        },
        false,
      )
      // 超时处理
      setTimeout(() => resolve(false), 3000)
    } else {
      doWeixinJSBridgePay(resolve)
    }
  })
}

/** 执行 WeixinJSBridge 支付 */
function doWeixinJSBridgePay(resolve: (result: boolean) => void) {
  if (!window.WeixinJSBridge) {
    resolve(false)
    return
  }

  // 使用从后端获取的 JSAPI 参数
  const params = currentJSAPIParams.value
  if (!params) {
    resolve(false)
    return
  }

  window.WeixinJSBridge.invoke(
    'getBrandWCPayRequest',
    {
      appId: params.appId,
      timeStamp: params.timeStamp,
      nonceStr: params.nonceStr,
      package: params.package,
      signType: params.signType,
      paySign: params.paySign,
    },
    (res) => {
      if (res.err_msg === 'get_brand_wcpay_request:ok') {
        order.status = 'paid'
        paid.value = true
        showPaymentResult(true)
        resolve(true)
      } else if (res.err_msg === 'get_brand_wcpay_request:cancel') {
        showPaymentResult(false, '用户取消支付')
        resolve(false)
      } else {
        showPaymentResult(false, '支付失败: ' + res.err_msg)
        resolve(false)
      }
    },
  )
}

/** 显示支付结果 */
function showPaymentResult(success: boolean, message?: string) {
  paySuccess.value = success
  resultMessage.value = message || ''
  showResult.value = true
}

/** 关闭结果弹窗 */
function closeResult() {
  showResult.value = false
  if (paySuccess.value) {
    // 支付成功后跳转
    setTimeout(() => {
      router.push('/')
    }, 500)
  }
}
</script>

<style scoped>
.payment-page {
  min-height: 100vh;
  padding: 0 16px 40px;
  background: #f5f5f5;
}

.page-header {
  padding: 24px 0 20px;
  text-align: center;
}

.header-title {
  font-size: 20px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 6px;
}

.header-desc {
  font-size: 13px;
  color: #999;
}

.order-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 12px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 14px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 14px;
}

.order-label {
  font-size: 13px;
  color: #999;
}

.order-id {
  font-size: 13px;
  color: #666;
  font-family: monospace;
}

.order-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.order-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item-label {
  font-size: 14px;
  color: #666;
}

.item-value {
  font-size: 14px;
  color: #1a1a1a;
  font-weight: 500;
}

.item-value.price {
  font-size: 18px;
  font-weight: 700;
  color: #e64545;
}

.item-value.status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.item-value.status.pending {
  background: #fff7e6;
  color: #fa8c16;
}

.item-value.status.paid {
  background: #e8f8ee;
  color: #07c160;
}

.payment-method {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.method-header {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 12px;
}

.method-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
}

.method-left {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #333;
}

.method-icon {
  flex-shrink: 0;
}

.method-checked {
  color: #07c160;
  font-size: 16px;
  font-weight: 700;
}

.btn-area {
  margin-bottom: 16px;
}

.pay-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  color: #fff;
  background: linear-gradient(135deg, #07c160 0%, #06ad56 100%);
  border: none;
  border-radius: 24px;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(7, 193, 96, 0.3);
  transition: transform 0.1s, opacity 0.2s;
}

.pay-btn:active {
  transform: scale(0.98);
}

.pay-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.footer-tips {
  text-align: center;
  font-size: 12px;
  color: #bbb;
}

.link {
  color: #667eea;
}

/* 结果弹窗 */
.result-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.result-card {
  width: 280px;
  padding: 32px 24px 24px;
  background: #fff;
  border-radius: 16px;
  text-align: center;
}

.result-card.success {
  /* 默认样式 */
}

.result-card.fail {
  /* 默认样式 */
}

.result-icon {
  width: 56px;
  height: 56px;
  margin: 0 auto 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
  color: #fff;
}

.result-card.success .result-icon {
  background: #07c160;
}

.result-card.fail .result-icon {
  background: #e64545;
}

.result-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8px;
}

.result-desc {
  font-size: 13px;
  color: #999;
  margin-bottom: 24px;
}

.result-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  font-weight: 500;
  border: none;
  border-radius: 22px;
  cursor: pointer;
}

.result-card.success .result-btn {
  color: #fff;
  background: #07c160;
}

.result-card.fail .result-btn {
  color: #666;
  background: #f0f0f0;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>