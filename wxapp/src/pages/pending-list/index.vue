<template>
  <div class="pending-page">
    <div class="page-header">
      <h1 class="title">待支付列表</h1>
      <button class="back-btn" @click="goBack">返回</button>
    </div>

    <div v-if="list.length" class="list">
      <div v-for="item in list" :key="item.id" class="card" @click="goPay(item)">
        <div class="row-top">
          <span class="desc">{{ item.description || '待支付项目' }}</span>
          <span class="tag">待支付</span>
        </div>
        <div class="row-mid">
          <span class="order">单号 {{ item.orderId }}</span>
          <span class="amount">¥{{ (item.amount ?? 0).toFixed(2) }}</span>
        </div>
        <div class="row-bottom">
          <span class="time">{{ item.createdAt }}</span>
          <button class="btn-cancel" @click.stop="onCancel(item)">取消</button>
          <button class="btn-pay" @click.stop="goPay(item)">去支付</button>
        </div>
      </div>
    </div>
    <div v-else class="empty">
      <div class="empty-icon">🧾</div>
      <div>暂无待支付项目</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getPaymentList, cancelPayment, type PaymentRow } from '@/api'

const router = useRouter()
const list = ref<PaymentRow[]>([])

async function load() {
  try {
    const data = await getPaymentList({ page: 1, pageSize: 20, status: 'pending' })
    list.value = data.list || []
  } catch (e) {
    console.error('加载待支付列表失败:', e)
  }
}

function goPay(item: PaymentRow) {
  // 带订单号进确认支付页，确认页对该订单下单
  router.push({ path: '/payment', query: { order_id: item.orderId, amount: String(item.amount) } })
}

async function onCancel(item: PaymentRow) {
  if (!confirm(`确定取消订单 ${item.orderId} 吗？`)) return
  try {
    await cancelPayment(item.id)
    load()
  } catch (e) {
    alert((e as Error).message || '取消失败')
  }
}

function goBack() {
  router.back()
}

onMounted(load)
</script>

<style scoped>
.pending-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 32px;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 16px;
}
.title {
  font-size: 18px;
  font-weight: 600;
}
.back-btn {
  background: transparent;
  border: none;
  color: #666;
  font-size: 14px;
}
.list {
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.card {
  background: #fff;
  border-radius: 12px;
  padding: 14px 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}
.card:active {
  opacity: 0.85;
}
.row-top {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}
.desc {
  font-size: 15px;
  font-weight: 600;
}
.tag {
  font-size: 11px;
  color: #fa8c16;
  background: #fff7e6;
  border-radius: 4px;
  padding: 2px 8px;
}
.row-mid {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}
.order {
  font-size: 12px;
  color: #999;
  font-family: monospace;
}
.amount {
  font-size: 17px;
  font-weight: 700;
  color: #e64545;
}
.row-bottom {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-top: 10px;
  border-top: 1px solid #f0f0f0;
}
.time {
  flex: 1;
  font-size: 11px;
  color: #bbb;
}
.btn-cancel {
  border: none;
  background: #f0f0f0;
  color: #666;
  font-size: 12px;
  padding: 6px 14px;
  border-radius: 14px;
}
.btn-pay {
  border: none;
  background: linear-gradient(135deg, #07c160, #06ad56);
  color: #fff;
  font-size: 12px;
  padding: 6px 16px;
  border-radius: 14px;
}
.empty {
  padding-top: 120px;
  text-align: center;
  color: #999;
  font-size: 14px;
}
.empty-icon {
  font-size: 56px;
  margin-bottom: 12px;
}
</style>
