<template>
  <div class="monitor-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="title">系统监控</span>
            <el-tag v-if="systemData.os_info" :type="osTagType" size="small" class="os-tag">
              {{ osDisplay }}
            </el-tag>
          </div>
          <div class="header-right">
            <span class="update-time" v-if="lastUpdateTime">最后更新：{{ lastUpdateTime }}</span>
            <el-select
              v-model="refreshInterval"
              size="small"
              class="interval-select"
              @change="changeRefreshInterval"
            >
              <el-option label="5秒" :value="5000" />
              <el-option label="10秒" :value="10000" />
              <el-option label="30秒" :value="30000" />
            </el-select>
            <el-switch
              v-model="autoRefresh"
              active-text="自动刷新"
              class="refresh-switch"
              @change="toggleAutoRefresh"
            />
            <el-button @click="refreshData" circle>
              <el-icon>
                <Refresh />
              </el-icon>
            </el-button>
          </div>
        </div>
      </template>

      <div class="monitor-content">
        <el-row :gutter="20" class="mb-6">
          <el-col :xs="24" :sm="12" :md="8">
            <el-card shadow="hover">
              <div class="resource-card">
                <div class="resource-header">
                  <el-icon class="header-icon">
                    <Monitor />
                  </el-icon>
                  <span class="resource-title">CPU 使用率</span>
                </div>
                <div class="resource-value">
                  {{ systemData.cpu?.usage_percent?.toFixed(2) ?? 0 }}%
                </div>
                <el-progress
                  :percentage="systemData.cpu?.usage_percent ?? 0"
                  :color="getProgressColor(systemData.cpu?.usage_percent ?? 0)"
                  :duration="0"
                />
                <div class="resource-detail">
                  核心数：{{ systemData.cpu?.cores ?? 0 }} | 频率：{{
                    formatFrequency(systemData.cpu?.frequency ?? 0)
                  }}
                </div>
              </div>
            </el-card>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-card shadow="hover">
              <div class="resource-card">
                <div class="resource-header">
                  <el-icon class="header-icon">
                    <Coin />
                  </el-icon>
                  <span class="resource-title">内存使用率</span>
                </div>
                <div class="resource-value">
                  {{ systemData.memory?.usage_percent?.toFixed(2) ?? 0 }}%
                </div>
                <div class="resource-detail">
                  {{ formatBytes(systemData.memory?.used ?? 0) }} /
                  {{ formatBytes(systemData.memory?.total ?? 0) }}
                </div>
                <el-progress
                  :percentage="systemData.memory?.usage_percent ?? 0"
                  :color="getProgressColor(systemData.memory?.usage_percent ?? 0)"
                  :duration="0"
                />
              </div>
            </el-card>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-card shadow="hover">
              <div class="resource-card">
                <div class="resource-header">
                  <el-icon class="header-icon">
                    <Folder />
                  </el-icon>
                  <span class="resource-title">磁盘使用率</span>
                </div>
                <div class="resource-value">{{ mainDiskUsage?.toFixed(2) ?? 0 }}%</div>
                <div class="resource-detail">
                  {{ formatBytes(mainDisk?.used ?? 0) }} / {{ formatBytes(mainDisk?.total ?? 0) }}
                </div>
                <el-progress
                  :percentage="mainDiskUsage ?? 0"
                  :color="getProgressColor(mainDiskUsage ?? 0)"
                  :duration="0"
                />
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <el-divider>趋势图表</el-divider>

      <el-row :gutter="20" class="chart-row">
        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>CPU 使用率趋势</span></template>
            <div ref="cpuChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>内存使用趋势</span></template>
            <div ref="memoryChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="chart-row">
        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>网络流量趋势</span></template>
            <div ref="networkChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>磁盘使用分布</span></template>
            <div ref="diskChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-divider>详细信息</el-divider>

      <el-row :gutter="20" class="detail-row">
        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>CPU 详情</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="CPU 核心数">{{
                systemData.cpu?.cores ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="CPU 使用率"
                >{{ systemData.cpu?.usage_percent?.toFixed(2) ?? 0 }}%</el-descriptions-item
              >
              <el-descriptions-item label="基准频率">{{
                formatFrequency(systemData.cpu?.frequency ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item v-if="!systemData.load?.is_simulated" label="系统负载">
                {{ systemData.load?.load1?.toFixed(2) ?? 0 }} /
                {{ systemData.load?.load5?.toFixed(2) ?? 0 }} /
                {{ systemData.load?.load15?.toFixed(2) ?? 0 }}
              </el-descriptions-item>
            </el-descriptions>
            <div v-if="systemData.cpu?.per_cpu_usage?.length > 0" class="mt-4">
              <div class="sub-title">各核心使用率</div>
              <div class="cpu-grid">
                <div
                  v-for="(usage, index) in systemData.cpu.per_cpu_usage"
                  :key="index"
                  class="cpu-core-card"
                >
                  <div class="cpu-core-header">
                    <span class="cpu-core-label">核心 {{ index + 1 }}</span>
                    <span class="cpu-core-percent" :style="{ color: getProgressColor(usage) }"
                      >{{ usage?.toFixed(1) }}%</span
                    >
                  </div>
                  <el-progress
                    :percentage="usage"
                    :stroke-width="6"
                    :color="getProgressColor(usage)"
                    :show-text="false"
                    :duration="0"
                  />
                </div>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>内存详情</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="总内存">{{
                formatBytes(systemData.memory?.total ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item label="已使用">{{
                formatBytes(systemData.memory?.used ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item label="空闲">{{
                formatBytes(systemData.memory?.free ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item label="可用">{{
                formatBytes(systemData.memory?.available ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item label="缓冲区">{{
                formatBytes(systemData.memory?.buffers ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item label="缓存">{{
                formatBytes(systemData.memory?.cached ?? 0)
              }}</el-descriptions-item>
            </el-descriptions>
            <div v-if="systemData.memory?.swap_total > 0" class="mt-4">
              <div class="sub-title">Swap 内存</div>
              <el-descriptions :column="3" border>
                <el-descriptions-item label="总 Swap">{{
                  formatBytes(systemData.memory.swap_total)
                }}</el-descriptions-item>
                <el-descriptions-item label="已使用">{{
                  formatBytes(systemData.memory.swap_used ?? 0)
                }}</el-descriptions-item>
                <el-descriptions-item label="空闲">{{
                  formatBytes(systemData.memory.swap_free ?? 0)
                }}</el-descriptions-item>
              </el-descriptions>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="detail-row">
        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>磁盘详情</span></template>
            <el-table :data="systemData.disk" style="width: 100%" :show-header="true" size="small">
              <el-table-column prop="mountpoint" label="挂载点" width="120" />
              <el-table-column prop="fstype" label="类型" width="80" />
              <el-table-column label="总空间">
                <template #default="{ row }">{{ formatBytes(row.total) }}</template>
              </el-table-column>
              <el-table-column label="已使用">
                <template #default="{ row }">{{ formatBytes(row.used) }}</template>
              </el-table-column>
              <el-table-column label="使用率" width="150">
                <template #default="{ row }">
                  <el-progress
                    :percentage="row.usage_percent"
                    :color="getProgressColor(row.usage_percent)"
                    :stroke-width="12"
                  />
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>网络统计</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="接收字节">{{
                formatBytes(systemData.network?.bytes_recv ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item label="发送字节">{{
                formatBytes(systemData.network?.bytes_sent ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item label="接收数据包">{{
                systemData.network?.packets_recv?.toLocaleString() ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="发送数据包">{{
                systemData.network?.packets_sent?.toLocaleString() ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="接收错误">{{
                systemData.network?.errin ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="发送错误">{{
                systemData.network?.errout ?? 0
              }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="detail-row">
        <el-col :xs="24" :md="8">
          <el-card>
            <template #header><span>系统信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="操作系统">{{
                systemData.os_info?.platform ?? '-'
              }}</el-descriptions-item>
              <el-descriptions-item label="主机名">{{
                systemData.os_info?.hostname ?? '-'
              }}</el-descriptions-item>
              <el-descriptions-item label="内核版本">{{
                systemData.os_info?.kernel ?? '-'
              }}</el-descriptions-item>
              <el-descriptions-item label="运行时间">{{
                formatUptime(systemData.os_info?.uptime ?? 0)
              }}</el-descriptions-item>
              <el-descriptions-item label="进程总数">{{
                systemData.processes?.total ?? 0
              }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="8">
          <el-card>
            <template #header><span>CPU 占用 Top 5</span></template>
            <el-table :data="systemData.processes?.top_cpu ?? []" style="width: 100%" size="small">
              <el-table-column prop="name" label="进程名" show-overflow-tooltip />
              <el-table-column label="CPU%" width="80">
                <template #default="{ row }">{{ row.cpu_percent?.toFixed(1) }}%</template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="8">
          <el-card>
            <template #header><span>内存占用 Top 5</span></template>
            <el-table
              :data="systemData.processes?.top_memory ?? []"
              style="width: 100%"
              size="small"
            >
              <el-table-column prop="name" label="进程名" show-overflow-tooltip />
              <el-table-column label="MEM%" width="80">
                <template #default="{ row }">{{ row.mem_percent?.toFixed(1) }}%</template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>

      <el-divider>其他监控</el-divider>

      <el-row :gutter="20" class="detail-row">
        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>数据库连接</span></template>
            <el-descriptions v-if="databaseData" :column="2" border>
              <el-descriptions-item label="活跃连接">{{
                databaseData.pool_stats?.in_use ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="空闲连接">{{
                databaseData.pool_stats?.idle ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="最大连接">{{
                databaseData.pool_stats?.max_open_conns ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="等待次数">{{
                databaseData.pool_stats?.wait_count ?? 0
              }}</el-descriptions-item>
            </el-descriptions>
            <el-empty v-else description="暂无数据" :image-size="60" />
          </el-card>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-card>
            <template #header><span>缓存监控</span></template>
            <el-descriptions v-if="cacheData" :column="2" border>
              <el-descriptions-item label="命中率"
                >{{ cacheData.hit_rate?.toFixed(2) ?? 0 }}%</el-descriptions-item
              >
              <el-descriptions-item label="命中次数">{{
                cacheData.redis_info?.keyspace_hits?.toLocaleString() ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="未命中次数">{{
                cacheData.redis_info?.keyspace_misses?.toLocaleString() ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="键数量">{{
                cacheData.key_stats?.total_keys?.toLocaleString() ?? 0
              }}</el-descriptions-item>
              <el-descriptions-item label="占用内存">{{
                formatBytes((cacheData.redis_info?.used_memory_mb ?? 0) * 1024 * 1024)
              }}</el-descriptions-item>
            </el-descriptions>
            <el-empty v-else description="暂无数据" :image-size="60" />
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Monitor, Coin, Folder } from '@element-plus/icons-vue'
import { getSystemMonitor, getCacheMonitor, getDatabaseMonitor } from '@/api'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

defineOptions({
  name: 'MonitorDashboard',
})

const autoRefresh = ref(true)
const refreshInterval = ref(5000)
const lastUpdateTime = ref('')

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const systemData = ref<any>({
  cpu: { usage_percent: 0, cores: 0, per_cpu_usage: [], frequency: 0 },
  memory: {
    total: 0,
    used: 0,
    free: 0,
    usage_percent: 0,
    available: 0,
    buffers: 0,
    cached: 0,
    swap_total: 0,
    swap_used: 0,
    swap_free: 0,
  },
  disk: [],
  network: { bytes_recv: 0, bytes_sent: 0, packets_recv: 0, packets_sent: 0, errin: 0, errout: 0 },
  load: { load1: 0, load5: 0, load15: 0, is_simulated: false, cpu_percent: 0 },
  os_info: { os: '', hostname: '', platform: '', kernel: '', uptime: 0 },
  processes: { total: 0, top_cpu: [], top_memory: [] },
})

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const databaseData = ref<any>(null)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const cacheData = ref<any>(null)

let refreshTimer: ReturnType<typeof setInterval> | null = null

// Chart refs
const cpuChartRef = ref<HTMLElement>()
const memoryChartRef = ref<HTMLElement>()
const networkChartRef = ref<HTMLElement>()
const diskChartRef = ref<HTMLElement>()

// Chart instances
let cpuChart: echarts.ECharts | null = null
let memoryChart: echarts.ECharts | null = null
let networkChart: echarts.ECharts | null = null
let diskChart: echarts.ECharts | null = null

// Data history for charts
const cpuHistory = ref<number[]>([])
const memoryHistory = ref<number[]>([])
const networkRecvHistory = ref<number[]>([])
const networkSentHistory = ref<number[]>([])
const timeHistory = ref<string[]>([])
const maxHistoryLength = 20

// Computed
const osDisplay = computed(() => {
  const os = systemData.value.os_info?.os || ''
  return os ? os.charAt(0).toUpperCase() + os.slice(1) : ''
})

const osTagType = computed(() => {
  const os = systemData.value.os_info?.os || ''
  if (os === 'linux') return 'success'
  if (os === 'windows') return 'primary'
  if (os === 'darwin') return 'warning'
  return 'info'
})

const mainDisk = computed(() => {
  const disks = systemData.value.disk
  if (!disks || disks.length === 0) return null
  const os = systemData.value.os_info?.os || ''
  if (os === 'windows') {
    return disks.find((d: { mountpoint: string }) => d.mountpoint === 'C:\\') || disks[0]
  }
  return disks.find((d: { mountpoint: string }) => d.mountpoint === '/') || disks[0]
})

const mainDiskUsage = computed(() => {
  return mainDisk.value?.usage_percent ?? 0
})

// Format functions
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

const formatFrequency = (mhz: number): string => {
  if (mhz === 0) return '-'
  if (mhz >= 1000) return `${(mhz / 1000).toFixed(2)} GHz`
  return `${mhz.toFixed(0)} MHz`
}

const formatUptime = (seconds: number): string => {
  if (seconds === 0) return '-'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}天 ${hours}小时 ${minutes}分钟`
  if (hours > 0) return `${hours}小时 ${minutes}分钟`
  return `${minutes}分钟`
}

const getProgressColor = (percent: number) => {
  if (percent < 60) return '#67C23A'
  if (percent < 80) return '#E6A23C'
  return '#F56C6C'
}

const getCurrentTime = () => {
  const now = new Date()
  return `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
}

// Chart initialization
const initCharts = () => {
  if (cpuChartRef.value) {
    cpuChart = echarts.init(cpuChartRef.value)
  }
  if (memoryChartRef.value) {
    memoryChart = echarts.init(memoryChartRef.value)
  }
  if (networkChartRef.value) {
    networkChart = echarts.init(networkChartRef.value)
  }
  if (diskChartRef.value) {
    diskChart = echarts.init(diskChartRef.value)
  }
}

const updateCpuChart = () => {
  if (!cpuChart) return
  const option: EChartsOption = {
    tooltip: { trigger: 'axis', formatter: '{b}: {c}%' },
    grid: { left: '10%', right: '5%', bottom: '15%', top: '10%' },
    xAxis: { type: 'category', data: timeHistory.value, axisLabel: { fontSize: 10 } },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [
      {
        name: 'CPU%',
        type: 'line',
        data: cpuHistory.value,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#409EFF', width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64,158,255,0.3)' },
            { offset: 1, color: 'rgba(64,158,255,0.05)' },
          ]),
        },
      },
    ],
  }
  cpuChart.setOption(option, { notMerge: false })
}

const updateMemoryChart = () => {
  if (!memoryChart) return
  const option: EChartsOption = {
    tooltip: { trigger: 'axis', formatter: '{b}: {c}%' },
    grid: { left: '10%', right: '5%', bottom: '15%', top: '10%' },
    xAxis: { type: 'category', data: timeHistory.value, axisLabel: { fontSize: 10 } },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [
      {
        name: '内存%',
        type: 'line',
        data: memoryHistory.value,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#67C23A', width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(103,194,58,0.3)' },
            { offset: 1, color: 'rgba(103,194,58,0.05)' },
          ]),
        },
      },
    ],
  }
  memoryChart.setOption(option)
}

const updateNetworkChart = () => {
  if (!networkChart) return
  const option: EChartsOption = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['接收', '发送'], bottom: 0 },
    grid: { left: '10%', right: '5%', bottom: '20%', top: '10%' },
    xAxis: { type: 'category', data: timeHistory.value, axisLabel: { fontSize: 10 } },
    yAxis: { type: 'value', axisLabel: { formatter: (v: number) => formatBytes(v), fontSize: 10 } },
    series: [
      {
        name: '接收',
        type: 'line',
        data: networkRecvHistory.value,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#409EFF' },
        areaStyle: { opacity: 0.2 },
      },
      {
        name: '发送',
        type: 'line',
        data: networkSentHistory.value,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#E6A23C' },
        areaStyle: { opacity: 0.2 },
      },
    ],
  }
  networkChart.setOption(option)
}

const updateDiskChart = () => {
  if (!diskChart) return
  const disks = systemData.value.disk
  if (!disks || disks.length === 0) return
  const option: EChartsOption = {
    tooltip: { trigger: 'item', formatter: '{b}: {c}%' },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        label: { show: true, formatter: '{b}: {d}%' },
        data: disks.map((d: { mountpoint: string; usage_percent: number }) => ({
          name: d.mountpoint,
          value: d.usage_percent,
          itemStyle: { color: getProgressColor(d.usage_percent) },
        })),
      },
    ],
  }
  diskChart.setOption(option)
}

const updateChartHistory = () => {
  const time = getCurrentTime()
  timeHistory.value.push(time)
  cpuHistory.value.push(systemData.value.cpu?.usage_percent ?? 0)
  memoryHistory.value.push(systemData.value.memory?.usage_percent ?? 0)
  networkRecvHistory.value.push(systemData.value.network?.bytes_recv ?? 0)
  networkSentHistory.value.push(systemData.value.network?.bytes_sent ?? 0)

  if (timeHistory.value.length > maxHistoryLength) {
    timeHistory.value.shift()
    cpuHistory.value.shift()
    memoryHistory.value.shift()
    networkRecvHistory.value.shift()
    networkSentHistory.value.shift()
  }

  updateCpuChart()
  updateMemoryChart()
  updateNetworkChart()
  updateDiskChart()
}

// Data fetching
const fetchSystemMonitor = async () => {
  try {
    const res = await getSystemMonitor()
    if (res.code === 0 && res.data) {
      Object.assign(systemData.value, res.data)
      lastUpdateTime.value = new Date().toLocaleTimeString()
    }
  } catch (error) {
    console.error('获取系统监控数据失败:', error)
  }
}

const fetchDatabaseMonitor = async () => {
  try {
    const res = await getDatabaseMonitor()
    if (res.code === 0 && res.data) {
      if (databaseData.value) {
        Object.assign(databaseData.value, res.data)
      } else {
        databaseData.value = res.data
      }
    }
  } catch (error) {
    console.error('获取数据库监控数据失败:', error)
  }
}

const fetchCacheMonitor = async () => {
  try {
    const res = await getCacheMonitor()
    if (res.code === 0 && res.data) {
      if (cacheData.value) {
        Object.assign(cacheData.value, res.data)
      } else {
        cacheData.value = res.data
      }
    }
  } catch (error) {
    console.error('获取缓存监控数据失败:', error)
  }
}

const refreshData = async () => {
  try {
    await Promise.all([fetchSystemMonitor(), fetchDatabaseMonitor(), fetchCacheMonitor()])
    updateChartHistory()
  } catch {
    ElMessage.error('刷新数据失败')
  }
}

const changeRefreshInterval = () => {
  stopRefreshTimer()
  if (autoRefresh.value) {
    startRefreshTimer()
  }
}

const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    startRefreshTimer()
  } else {
    stopRefreshTimer()
  }
}

const startRefreshTimer = () => {
  refreshTimer = setInterval(() => {
    refreshData()
  }, refreshInterval.value)
}

const stopRefreshTimer = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

const handleResize = () => {
  cpuChart?.resize()
  memoryChart?.resize()
  networkChart?.resize()
  diskChart?.resize()
}

onMounted(async () => {
  await nextTick()
  initCharts()
  await refreshData()
  if (autoRefresh.value) {
    startRefreshTimer()
  }
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  stopRefreshTimer()
  cpuChart?.dispose()
  memoryChart?.dispose()
  networkChart?.dispose()
  diskChart?.dispose()
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
.monitor-container {
  padding: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .title {
        font-size: 18px;
        font-weight: bold;
      }

      .os-tag {
        margin-left: 4px;
      }
    }

    .header-right {
      display: flex;
      align-items: center;
      gap: 12px;

      .update-time {
        font-size: 12px;
        color: #909399;
      }

      .interval-select {
        width: 90px;
      }

      .refresh-switch {
        :deep(.el-switch__label) {
          font-size: 12px;
        }
      }
    }
  }

  .mb-6 {
    margin-bottom: 20px;
  }

  .mt-4 {
    margin-top: 16px;
  }

  .resource-card {
    .resource-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 12px;

      .header-icon {
        font-size: 20px;
        color: #409eff;
      }

      .resource-title {
        font-size: 14px;
        color: #606266;
      }
    }

    .resource-value {
      font-size: 24px;
      font-weight: bold;
      color: #303133;
      margin-bottom: 8px;
    }

    .resource-detail {
      font-size: 12px;
      color: #909399;
      margin-top: 8px;
    }
  }

  .chart-row {
    margin-bottom: 20px;
  }

  .chart-container {
    width: 100%;
    height: 280px;
    background: var(--el-bg-color);
  }

  .detail-row {
    margin-bottom: 20px;
  }

  .sub-title {
    font-size: 14px;
    font-weight: bold;
    margin-bottom: 12px;
    color: #303133;
  }

  .cpu-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
    margin-top: 12px;

    .cpu-core-card {
      background: var(--el-fill-color-light);
      border-radius: 8px;
      padding: 10px 12px;
      transition: all 0.2s;

      &:hover {
        background: var(--el-fill-color);
        transform: translateY(-2px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
      }

      .cpu-core-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 6px;

        .cpu-core-label {
          font-size: 12px;
          color: var(--el-text-color-regular);
          font-weight: 500;
        }

        .cpu-core-percent {
          font-size: 12px;
          font-weight: bold;
        }
      }
    }
  }

  .db-stats,
  .cache-stats {
    padding: 16px 0;

    .db-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 16px;

      .db-stat-card {
        display: flex;
        align-items: center;
        gap: 16px;
        padding: 16px;
        background: var(--el-fill-color-light);
        border-radius: 12px;
        transition: all 0.3s;
        border: 1px solid transparent;

        &:hover {
          background: var(--el-fill-color);
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
          border-color: var(--el-border-color-light);
        }

        .db-stat-icon {
          width: 48px;
          height: 48px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 24px;
          flex-shrink: 0;

          &.active {
            background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
            color: white;
          }

          &.idle {
            background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
            color: white;
          }

          &.max {
            background: linear-gradient(135deg, #e6a23c 0%, #ebb563 100%);
            color: white;
          }

          &.wait {
            background: linear-gradient(135deg, #f56c6c 0%, #f78989 100%);
            color: white;
          }
        }

        .db-stat-content {
          flex: 1;
          min-width: 0;

          .db-stat-label {
            font-size: 12px;
            color: var(--el-text-color-secondary);
            margin-bottom: 6px;
          }

          .db-stat-value {
            font-size: 24px;
            font-weight: bold;
            color: var(--el-text-color-primary);
            margin-bottom: 8px;

            &.has-wait {
              color: #f56c6c;
            }
          }
        }
      }
    }

    .cache-grid {
      display: grid;
      grid-template-columns: repeat(5, 1fr);
      gap: 16px;

      .cache-stat-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 20px 12px;
        background: var(--el-fill-color-light);
        border-radius: 12px;
        transition: all 0.3s;
        border: 1px solid transparent;
        text-align: center;

        &:hover {
          background: var(--el-fill-color);
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
          border-color: var(--el-border-color-light);
        }

        &.hit-rate {
          background: linear-gradient(
            135deg,
            var(--el-fill-color-light) 0%,
            var(--el-fill-color) 100%
          );
          border: 2px solid var(--el-border-color-light);
        }

        .cache-stat-icon {
          width: 40px;
          height: 40px;
          border-radius: 10px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 20px;
          margin-bottom: 12px;
          background: var(--el-color-primary-light-9);
          color: var(--el-color-primary);
        }

        .cache-stat-content {
          .cache-stat-label {
            font-size: 12px;
            color: var(--el-text-color-secondary);
            margin-bottom: 8px;
          }

          .cache-stat-big-value {
            font-size: 28px;
            font-weight: bold;
            line-height: 1.2;
          }

          .cache-stat-value {
            font-size: 20px;
            font-weight: bold;
            color: var(--el-text-color-primary);
            line-height: 1.2;
          }
        }
      }
    }
  }

  @media (max-width: 768px) {
    .db-stats,
    .cache-stats {
      .db-grid {
        grid-template-columns: 1fr;
      }

      .cache-grid {
        grid-template-columns: repeat(2, 1fr);
      }
    }
  }
}
</style>
