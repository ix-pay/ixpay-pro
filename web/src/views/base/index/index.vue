<template>
  <!-- Dashboard 首页 - 现代化数据可视化设计 -->
  <div class="min-h-screen bg-bg-secondary">
    <!-- 顶部欢迎区域 - 渐变背景设计 -->
    <div class="relative overflow-hidden bg-gradient-to-r from-primary-color to-purple-600">
      <!-- 装饰性光晕 -->
      <div class="absolute -right-16 -top-16 h-64 w-64 rounded-full bg-white opacity-10 blur-3xl"></div>
      <div class="absolute -bottom-16 -left-16 h-64 w-64 rounded-full bg-white opacity-10 blur-3xl"></div>

      <div class="relative mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <!-- 欢迎标题区 -->
        <div class="mb-6 flex items-center justify-between">
          <div>
            <h1 class="mb-2 text-3xl font-bold text-white">欢迎回来，{{ userName }}！</h1>
            <p class="text-primary-light text-sm">{{ curDate }}</p>
          </div>
          <el-button :icon="RefreshRight" circle size="small"
            class="border-none bg-white/20 text-white hover:bg-white/30" @click="refreshData" />
        </div>

        <!-- 总统计数据展示 -->
        <div class="grid grid-cols-1 gap-4 border-t border-white/20 pt-4 md:grid-cols-3">
          <div class="text-center">
            <div class="mb-1 text-2xl font-bold text-white">{{ totalUsers.toLocaleString() }}</div>
            <div class="text-primary-light text-xs">总用户数</div>
          </div>
          <div class="border-l border-white/20 text-center">
            <div class="mb-1 text-2xl font-bold text-white">{{ totalOrders.toLocaleString() }}</div>
            <div class="text-primary-light text-xs">总订单数</div>
          </div>
          <div class="border-l border-white/20 text-center">
            <div class="mb-1 text-2xl font-bold text-white">
              {{ formatCurrency(totalRevenue) }}
            </div>
            <div class="text-primary-light text-xs">总交易额</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="main-content mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <!-- 统计卡片区域 - 响应式网格布局 -->
      <div class="mb-6 grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        <IxStatCard title="今日订单" :value="todayOrders" change="+5.2%" change-type="positive" change-label="较昨日"
          icon="GoodsFilled" color="blue" :show-title-icon="true" />
        <IxStatCard title="今日交易额" :value="formatCurrency(todayAmount)" change="+12.8%" change-type="positive"
          change-label="较昨日" icon="Money" color="green" :show-title-icon="true" />
        <IxStatCard title="新增用户" :value="newUsers" change="+3.6%" change-type="positive" change-label="较昨日" icon="User"
          color="purple" :show-title-icon="true" />
        <IxStatCard title="退款订单" :value="refundOrders" change="-2.4%" change-type="negative" change-label="较昨日"
          icon="Warning" color="red" :show-title-icon="true" />
      </div>

      <!-- 图表区域 - 双列布局 -->
      <div class="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <!-- 销售趋势图表 -->
        <IxCard title="销售趋势" hover-effect>
          <template #header-actions>
            <el-select v-model="chartType" size="small" class="w-24" @change="updateSalesTrendData">
              <el-option label="日" value="day" />
              <el-option label="周" value="week" />
              <el-option label="月" value="month" />
            </el-select>
          </template>
          <IxEcharts :options="salesTrendOptions as EChartsOption" height="320px" :hover-effect="true" />
        </IxCard>

        <!-- 订单类型分布 -->
        <IxCard title="订单类型分布" hover-effect>
          <IxEcharts :options="orderDistributionOptions as EChartsOption" height="320px" :hover-effect="true" />
        </IxCard>
      </div>

      <!-- 最近订单与系统状态区 -->
      <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <!-- 最近订单表格 -->
        <IxCard title="最近订单" hover-effect>
          <template #header-actions>
            <el-button type="primary" size="small" :icon="Right" @click="navigateTo('all-orders')">
              查看全部
            </el-button>
          </template>
          <el-table :data="recentOrders" style="width: 100%" class="modern-table" :header-cell-style="headerCellStyle"
            :cell-style="cellStyle">
            <el-table-column prop="orderNo" label="订单号" min-width="180" />
            <el-table-column prop="amount" label="金额" min-width="120">
              <template #default="scope">
                <span class="font-medium text-text-primary">
                  {{ formatCurrency(scope.row.amount) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" min-width="100">
              <template #default="scope">
                <el-tag :type="getStatusTagType(scope.row.status)" size="small" effect="light">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createTime" label="创建时间" min-width="180" />
            <el-table-column label="操作" min-width="100" fixed="right">
              <template #default="scope">
                <el-button type="primary" size="small" :icon="View" @click="viewOrderDetails(scope.row.orderNo)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </IxCard>

        <!-- 右侧：系统状态和快速操作 -->
        <div class="flex flex-col gap-6">
          <!-- 系统状态监控 -->
          <IxCard title="系统状态" hover-effect>
            <div class="grid grid-cols-2 gap-4">
              <div
                class="flex items-center gap-3 rounded-xl bg-success-light/10 p-3 transition-all duration-300 hover:scale-105">
                <div class="text-success-color">
                  <el-icon :size="24">
                    <CircleCheckFilled />
                  </el-icon>
                </div>
                <div>
                  <div class="text-xs text-text-tertiary">服务器</div>
                  <div class="font-medium text-success-color">运行正常</div>
                </div>
              </div>
              <div
                class="flex items-center gap-3 rounded-xl bg-success-light/10 p-3 transition-all duration-300 hover:scale-105">
                <div class="text-success-color">
                  <el-icon :size="24">
                    <CircleCheckFilled />
                  </el-icon>
                </div>
                <div>
                  <div class="text-xs text-text-tertiary">数据库</div>
                  <div class="font-medium text-success-color">连接正常</div>
                </div>
              </div>
              <div
                class="flex items-center gap-3 rounded-xl bg-success-light/10 p-3 transition-all duration-300 hover:scale-105">
                <div class="text-success-color">
                  <el-icon :size="24">
                    <CircleCheckFilled />
                  </el-icon>
                </div>
                <div>
                  <div class="text-xs text-text-tertiary">API 接口</div>
                  <div class="font-medium text-success-color">响应正常</div>
                </div>
              </div>
              <div
                class="flex items-center gap-3 rounded-xl bg-warning-light/10 p-3 transition-all duration-300 hover:scale-105">
                <div class="text-warning-color">
                  <el-icon :size="24">
                    <Warning />
                  </el-icon>
                </div>
                <div>
                  <div class="text-xs text-text-tertiary">缓存</div>
                  <div class="font-medium text-warning-color">部分未命中</div>
                </div>
              </div>
            </div>
          </IxCard>

          <!-- 快速操作入口 -->
          <IxCard title="快速操作" hover-effect>
            <div class="grid grid-cols-2 gap-3">
              <el-button type="primary"
                class="flex h-20 flex-col items-center justify-center gap-2 transition-all duration-300 hover:scale-105">
                <el-icon :size="20">
                  <Plus />
                </el-icon>
                <span class="text-xs">创建订单</span>
              </el-button>
              <el-button type="success"
                class="flex h-20 flex-col items-center justify-center gap-2 transition-all duration-300 hover:scale-105">
                <el-icon :size="20">
                  <Download />
                </el-icon>
                <span class="text-xs">导出报表</span>
              </el-button>
              <el-button type="info"
                class="flex h-20 flex-col items-center justify-center gap-2 transition-all duration-300 hover:scale-105">
                <el-icon :size="20">
                  <Setting />
                </el-icon>
                <span class="text-xs">系统设置</span>
              </el-button>
              <el-button type="warning"
                class="flex h-20 flex-col items-center justify-center gap-2 transition-all duration-300 hover:scale-105">
                <el-icon :size="20">
                  <Message />
                </el-icon>
                <span class="text-xs">查看通知</span>
              </el-button>
            </div>
          </IxCard>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: 'DashboardIndex',
})

import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/modules/user'
import { ElMessage } from 'element-plus'
import {
  Setting,
  RefreshRight,
  CircleCheckFilled,
  Plus,
  Download,
  Message,
  Right,
  View,
  Warning,
} from '@element-plus/icons-vue'
import { IxCard, IxStatCard, IxEcharts } from '@/components/business'
import type { EChartsOption } from 'echarts'
import type { CSSProperties } from 'vue'

const router = useRouter()
const userStore = useUserStore()

// 用户信息
const userName = computed(() => userStore.userInfo?.nickname || '管理员')

// 统计数据
const todayOrders = ref(128)
const todayAmount = ref(38524.5)
const newUsers = ref(24)
const refundOrders = ref(5)

// 总统计数据
const totalUsers = ref(12580)
const totalOrders = ref(89652)
const totalRevenue = ref(2856478.9)

// 图表类型
const chartType = ref('day')

// 销售趋势图表配置
const salesTrendOptions = ref<EChartsOption>({
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'shadow',
    },
  },
  legend: {
    data: ['订单数', '交易额'],
    top: '0%',
    right: '5%',
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    top: '15%',
    containLabel: true,
  },
  xAxis: {
    type: 'category',
    data: [] as string[],
    axisLabel: {
      interval: 0,
      rotate: 30,
    },
  },
  yAxis: [
    {
      type: 'value',
      name: '订单数',
      position: 'left',
      axisLabel: {
        formatter: '{value}',
      },
    },
    {
      type: 'value',
      name: '交易额',
      position: 'right',
      axisLabel: {
        formatter: '{value}元',
      },
    },
  ],
  series: [
    {
      name: '订单数',
      type: 'line',
      smooth: true,
      data: [] as number[],
      itemStyle: {
        color: '#3B82F6',
      },
      areaStyle: {
        x: 0,
        y: 0,
        x2: 0,
        y2: 1,
        colorStops: [
          { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
          { offset: 1, color: 'rgba(59, 130, 246, 0.01)' },
        ],
      },
    },
    {
      name: '交易额',
      type: 'bar',
      yAxisIndex: 1,
      data: [] as number[],
      itemStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(16, 185, 129, 0.8)' },
            { offset: 1, color: 'rgba(16, 185, 129, 0.3)' },
          ],
        },
      },
    },
  ] as unknown,
} as EChartsOption)

// 订单类型分布图表配置
const orderDistributionOptions = ref<EChartsOption>({
  tooltip: {
    trigger: 'item',
    formatter: '{a} <br/>{b}: {c} ({d}%)',
  },
  legend: {
    orient: 'vertical',
    right: '5%',
    top: 'middle',
  },
  series: [
    {
      name: '订单类型',
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      padAngle: 2,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2,
      },
      label: {
        show: false,
        position: 'center',
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 20,
          fontWeight: 'bold',
        },
      },
      labelLine: {
        show: false,
      },
      data: [] as Array<{ name: string; value: number }>,
    },
  ],
})

// 当前日期
const curDate = ref('')

// 最近订单数据
const recentOrders = ref([
  {
    orderNo: 'IXP202401150001',
    amount: 256.0,
    status: '已完成',
    paymentMethod: '支付宝',
    createTime: '2024-01-15 14:30:25',
  },
  {
    orderNo: 'IXP202401150002',
    amount: 158.5,
    status: '已完成',
    paymentMethod: '微信支付',
    createTime: '2024-01-15 13:45:12',
  },
  {
    orderNo: 'IXP202401150003',
    amount: 420.0,
    status: '待支付',
    paymentMethod: '银联支付',
    createTime: '2024-01-15 12:18:30',
  },
  {
    orderNo: 'IXP202401150004',
    amount: 89.9,
    status: '已退款',
    paymentMethod: '支付宝',
    createTime: '2024-01-15 11:20:45',
  },
  {
    orderNo: 'IXP202401150005',
    amount: 315.0,
    status: '已完成',
    paymentMethod: '微信支付',
    createTime: '2024-01-15 10:05:18',
  },
])

// 格式化金额
const formatCurrency = (value: number): string => {
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: 'CNY',
    minimumFractionDigits: 2,
  }).format(value)
}

// 获取状态标签类型
const getStatusTagType = (
  status: '已完成' | '处理中' | '已取消' | '待支付' | '已退款',
): 'success' | 'primary' | 'danger' | 'warning' | 'info' => {
  const statusMap: Record<string, 'success' | 'primary' | 'danger' | 'warning' | 'info'> = {
    已完成: 'success',
    处理中: 'primary',
    已取消: 'danger',
    待支付: 'warning',
    已退款: 'danger',
  }
  return statusMap[status] || 'info'
}

// 页面跳转
const navigateTo = (path: string): void => {
  router.push(`/${path}`)
  ElMessage.success(`跳转到${path}页面`)
}

// 查看订单详情
const viewOrderDetails = (orderNo: string): void => {
  router.push(`/orders/detail/${orderNo}`)
}

// 表头样式
const headerCellStyle = (): CSSProperties => {
  return {
    textAlign: 'center',
    fontWeight: 'bold',
    backgroundColor: 'var(--bg-tertiary)',
    color: 'var(--text-primary)',
  }
}

// 单元格样式
const cellStyle = (): CSSProperties => {
  return {
    textAlign: 'center',
  }
}

// 刷新数据
const refreshData = () => {
  ElMessage.info('数据刷新中...')
  // 模拟刷新数据
  setTimeout(() => {
    // 随机更新一些数据
    todayOrders.value += Math.floor(Math.random() * 10) - 5
    todayAmount.value += Math.random() * 1000 - 500
    newUsers.value += Math.floor(Math.random() * 5) - 2
    refundOrders.value += Math.floor(Math.random() * 3) - 1

    // 更新图表数据
    updateSalesTrendData()
    updateOrderDistributionData()

    ElMessage.success('数据刷新成功')
  }, 1000)
}

// 更新销售趋势数据
const updateSalesTrendData = () => {
  const now = new Date()
  const dates: string[] = []
  const orderData: number[] = []
  const amountData: number[] = []

  // 根据图表类型生成不同的数据
  if (chartType.value === 'day') {
    // 日视图：显示 24 小时
    for (let i = 0; i < 24; i++) {
      dates.push(`${i.toString().padStart(2, '0')}:00`)
      orderData.push(Math.floor(Math.random() * 200) + 50)
      amountData.push(Math.floor(Math.random() * 50000) + 10000)
    }
  } else if (chartType.value === 'week') {
    // 周视图：显示 7 天
    const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
    for (let i = 6; i >= 0; i--) {
      const date = new Date(now)
      date.setDate(date.getDate() - i)
      dates.push(weekDays[date.getDay()])
      orderData.push(Math.floor(Math.random() * 500) + 200)
      amountData.push(Math.floor(Math.random() * 100000) + 50000)
    }
  } else {
    // 月视图：显示 30 天
    for (let i = 29; i >= 0; i--) {
      const date = new Date(now)
      date.setDate(date.getDate() - i)
      dates.push(`${date.getMonth() + 1}/${date.getDate()}`)
      orderData.push(Math.floor(Math.random() * 1000) + 500)
      amountData.push(Math.floor(Math.random() * 200000) + 100000)
    }
  }

  const series = salesTrendOptions.value.series
  if (series && Array.isArray(series) && series.length >= 2) {
    salesTrendOptions.value = {
      ...salesTrendOptions.value,
      xAxis: {
        ...salesTrendOptions.value.xAxis,
        data: dates,
      },
      series: [
        {
          ...series[0],
          data: orderData,
        },
        {
          ...series[1],
          data: amountData,
        },
      ] as unknown as echarts.SeriesOption[],
    } as EChartsOption
  }
}

// 更新订单类型分布数据
const updateOrderDistributionData = () => {
  const distributionData = [
    { name: '线上支付', value: Math.floor(Math.random() * 1000) + 500 },
    { name: '线下支付', value: Math.floor(Math.random() * 800) + 300 },
    { name: '扫码支付', value: Math.floor(Math.random() * 600) + 200 },
    { name: 'NFC 支付', value: Math.floor(Math.random() * 400) + 100 },
    { name: '刷脸支付', value: Math.floor(Math.random() * 300) + 50 },
  ]

  const series = orderDistributionOptions.value.series
  if (series && Array.isArray(series) && series.length >= 1) {
    orderDistributionOptions.value = {
      ...orderDistributionOptions.value,
      series: [
        {
          ...series[0],
          data: distributionData,
        },
      ] as unknown as echarts.SeriesOption[],
    } as EChartsOption
  }
}

// 页面加载时执行
onMounted(() => {
  // 验证用户是否已登录
  if (!userStore.token) {
    router.push('/login')
    return
  }

  // 获取当前日期
  const now = new Date()
  curDate.value = now.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  })

  // 初始化图表数据
  updateSalesTrendData()
  updateOrderDistributionData()
})
</script>

<style lang="scss" scoped>
// 现代化表格样式
.modern-table {
  :deep(.el-table) {
    --el-table-header-bg-color: var(--bg-tertiary);
    --el-table-header-text-color: var(--text-primary);
    --el-table-row-hover-bg-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    overflow: hidden;
    background-color: transparent;

    th.el-table__cell {
      background-color: var(--bg-tertiary);
      color: var(--text-primary);
      font-weight: 600;
    }

    td.el-table__cell {
      border-bottom-color: var(--border-primary);
    }

    .el-table__body tr:hover>td {
      background-color: var(--bg-secondary) !important;
    }
  }
}

// 响应式调整
@media (max-width: 768px) {
  .main-content {
    padding: 12px;
  }

  .grid {
    gap: 0.75rem;
  }
}
</style>
