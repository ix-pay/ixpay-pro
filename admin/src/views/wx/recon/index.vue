<template>
  <div class="flex flex-col gap-4 p-4">
    <!-- 汇总卡片：按功能色区分指标，暗黑模式经 CSS 变量自动适配 -->
    <div class="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-6">
      <el-card shadow="hover">
        <div class="flex items-center gap-3">
          <el-icon :size="28" color="var(--primary-color)"><Tickets /></el-icon>
          <el-statistic title="支付单总数" :value="summary.totalCount" />
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="flex items-center gap-3">
          <el-icon :size="28" color="var(--success-color)"><CircleCheck /></el-icon>
          <el-statistic title="支付成功笔数" :value="summary.successCount" />
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="flex items-center gap-3">
          <el-icon :size="28" color="var(--primary-color)"><Coin /></el-icon>
          <el-statistic title="成功金额(元)" :value="summary.successAmount" :precision="2" />
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="flex items-center gap-3">
          <el-icon :size="28" color="var(--info-color)"><RefreshLeft /></el-icon>
          <el-statistic title="退款笔数" :value="summary.refundCount" />
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="flex items-center gap-3">
          <el-icon :size="28" color="var(--info-color)"><Wallet /></el-icon>
          <el-statistic title="退款金额(元)" :value="summary.refundAmount" :precision="2" />
        </div>
      </el-card>
      <el-card shadow="hover">
        <div class="flex items-center gap-3">
          <el-icon :size="28" color="var(--warning-color)"><Clock /></el-icon>
          <el-statistic title="待支付笔数" :value="summary.pendingCount" />
        </div>
      </el-card>
    </div>

    <!-- 筛选 -->
    <el-form inline>
      <el-form-item label="租户">
        <TenantSwitcher @change="loadAll" />
      </el-form-item>
      <el-form-item label="日期范围">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          value-format="YYYY-MM-DD"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          @change="loadAll"
        />
      </el-form-item>
      <el-form-item label="状态">
        <el-select
          v-model="statusFilter"
          placeholder="全部"
          clearable
          style="width: 140px"
          @change="loadPayments"
        >
          <el-option label="待支付" value="pending" />
          <el-option label="已支付" value="success" />
          <el-option label="已退款" value="refunded" />
          <el-option label="已取消" value="cancelled" />
          <el-option label="失败" value="failed" />
        </el-select>
      </el-form-item>
    </el-form>

    <!-- 支付单明细 -->
    <el-table :data="payments" border stripe v-loading="loading">
      <el-table-column prop="id" label="支付ID" width="160" />
      <el-table-column prop="orderId" label="商户订单号" width="200" />
      <el-table-column prop="description" label="描述" min-width="140" show-overflow-tooltip />
      <el-table-column label="金额(元)" width="100">
        <template #default="{ row }">{{ (row.amount ?? 0).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="transactionId" label="微信交易号" width="220" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="创建时间" width="170" />
      <el-table-column prop="paidAt" label="支付时间" width="170" />
    </el-table>

    <!-- 分页 -->
    <div class="flex justify-end">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="loadPayments"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Tickets, CircleCheck, Coin, RefreshLeft, Wallet, Clock } from '@element-plus/icons-vue'
import { getWxReconSummary, getWxReconPayments } from '@/api/modules/wx-channel'
import TenantSwitcher from '@/components/TenantSwitcher/index.vue'

interface PaymentRow {
  id: string
  orderId: string
  description: string
  amount: number
  status: string
  transactionId: string
  createdAt: string
  paidAt: string
}

const defaultStart = () => {
  const d = new Date()
  d.setDate(d.getDate() - 7)
  return d.toISOString().slice(0, 10)
}

const dateRange = ref<[string, string] | null>([
  defaultStart(),
  new Date().toISOString().slice(0, 10),
])
const statusFilter = ref('')
const summary = ref({
  totalCount: 0,
  successCount: 0,
  successAmount: 0,
  refundCount: 0,
  refundAmount: 0,
  pendingCount: 0,
})
const payments = ref<PaymentRow[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

function statusText(s: string) {
  const map: Record<string, string> = {
    pending: '待支付',
    success: '已支付',
    refunded: '已退款',
    cancelled: '已取消',
    failed: '失败',
  }
  return map[s] || s
}

function statusTag(s: string): 'warning' | 'success' | 'info' | 'danger' {
  const map: Record<string, 'warning' | 'success' | 'info' | 'danger'> = {
    pending: 'warning',
    success: 'success',
    refunded: 'info',
    cancelled: 'info',
    failed: 'danger',
  }
  return map[s] || 'info'
}

async function loadSummary() {
  const [start, end] = dateRange.value || [undefined, undefined]
  // tenant_id 由 wx-channel API 层统一注入
  const res = await getWxReconSummary({ startDate: start, endDate: end })
  summary.value = res.data!
}

async function loadPayments() {
  loading.value = true
  try {
    const [start, end] = dateRange.value || [undefined, undefined]
    const res = await getWxReconPayments({
      startDate: start,
      endDate: end,
      status: statusFilter.value || undefined,
      page: page.value,
      pageSize: pageSize.value,
    })
    payments.value = res.data!.list as PaymentRow[]
    total.value = res.data!.total
  } finally {
    loading.value = false
  }
}

function loadAll() {
  page.value = 1
  loadSummary()
  loadPayments()
}

onMounted(loadAll)
</script>

