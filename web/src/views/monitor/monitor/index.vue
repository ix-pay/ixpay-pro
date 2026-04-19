<template>
  <div class="monitor-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>系统监控</span>
          <el-button @click="refreshData" :loading="loading" circle>
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
      </template>
      <div class="monitor-content">
        <!-- 系统资源概览 -->
        <el-row :gutter="20" class="mb-6">
          <el-col :span="6">
            <el-card shadow="hover">
              <div class="resource-card">
                <div class="resource-header">
                  <el-icon class="header-icon"><Monitor /></el-icon>
                  <span class="resource-title">CPU 使用率</span>
                </div>
                <div class="resource-value">
                  {{ systemData?.cpu?.usage_percent?.toFixed(2) ?? 0 }}%
                </div>
                <el-progress
                  :percentage="systemData?.cpu?.usage_percent ?? 0"
                  :color="getProgressColor(systemData?.cpu?.usage_percent ?? 0)"
                />
              </div>
            </el-card>
          </el-col>

          <el-col :span="6">
            <el-card shadow="hover">
              <div class="resource-card">
                <div class="resource-header">
                  <el-icon class="header-icon"><Coin /></el-icon>
                  <span class="resource-title">内存使用率</span>
                </div>
                <div class="resource-value">
                  {{ systemData?.memory?.usage_percent?.toFixed(2) ?? 0 }}%
                </div>
                <div class="resource-detail">
                  {{ formatBytes(systemData?.memory?.used ?? 0) }} /
                  {{ formatBytes(systemData?.memory?.total ?? 0) }}
                </div>
                <el-progress
                  :percentage="systemData?.memory?.usage_percent ?? 0"
                  :color="getProgressColor(systemData?.memory?.usage_percent ?? 0)"
                />
              </div>
            </el-card>
          </el-col>

          <el-col :span="6">
            <el-card shadow="hover">
              <div class="resource-card">
                <div class="resource-header">
                  <el-icon class="header-icon"><Folder /></el-icon>
                  <span class="resource-title">磁盘使用率</span>
                </div>
                <div class="resource-value">
                  {{ systemData?.disk?.usage_percent?.toFixed(2) ?? 0 }}%
                </div>
                <div class="resource-detail">
                  {{ formatBytes(systemData?.disk?.used ?? 0) }} /
                  {{ formatBytes(systemData?.disk?.total ?? 0) }}
                </div>
                <el-progress
                  :percentage="systemData?.disk?.usage_percent ?? 0"
                  :color="getProgressColor(systemData?.disk?.usage_percent ?? 0)"
                />
              </div>
            </el-card>
          </el-col>

          <el-col :span="6">
            <el-card shadow="hover">
              <div class="resource-card">
                <div class="resource-header">
                  <el-icon class="header-icon"><Grid /></el-icon>
                  <span class="resource-title">系统负载</span>
                </div>
                <div class="resource-value">
                  {{ systemData?.load?.load1?.toFixed(2) ?? 0 }}
                </div>
                <div class="resource-detail">
                  1 分钟：{{ systemData?.load?.load1?.toFixed(2) ?? 0 }} | 5 分钟：{{
                    systemData?.load?.load5?.toFixed(2) ?? 0
                  }}
                  | 15 分钟：{{ systemData?.load?.load15?.toFixed(2) ?? 0 }}
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 详细信息 -->
      <el-row :gutter="20">
        <!-- CPU 详情 -->
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>CPU 详情</span>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="CPU 核心数">{{
                systemData.cpu.cores
              }}</el-descriptions-item>
              <el-descriptions-item label="CPU 使用率"
                >{{ systemData.cpu.usage_percent }}%</el-descriptions-item
              >
            </el-descriptions>
            <div
              v-if="systemData.cpu.per_cpu_usage && systemData.cpu.per_cpu_usage.length > 0"
              class="mt-4"
            >
              <div class="sub-title">各核心使用率</div>
              <el-space direction="vertical" :size="8" style="width: 100%">
                <div
                  v-for="(usage, index) in systemData.cpu.per_cpu_usage"
                  :key="index"
                  class="core-item"
                >
                  <span>核心 {{ index + 1 }}</span>
                  <el-progress
                    :percentage="usage"
                    :stroke-width="8"
                    :color="getProgressColor(usage)"
                  />
                </div>
              </el-space>
            </div>
          </el-card>
        </el-col>

        <!-- 内存详情 -->
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>内存详情</span>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="总内存">{{
                formatSize(systemData.memory.total_mb)
              }}</el-descriptions-item>
              <el-descriptions-item label="已使用">{{
                formatSize(systemData.memory.used_mb)
              }}</el-descriptions-item>
              <el-descriptions-item label="空闲">{{
                formatSize(systemData.memory.free_mb)
              }}</el-descriptions-item>
              <el-descriptions-item label="可用">{{
                formatSize(systemData.memory.available_mb)
              }}</el-descriptions-item>
              <el-descriptions-item label="缓冲区">{{
                formatSize(systemData.memory.buffers_mb)
              }}</el-descriptions-item>
              <el-descriptions-item label="缓存">{{
                formatSize(systemData.memory.cached_mb)
              }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>
      </el-row>

      <!-- 磁盘和数据库监控 -->
      <el-row :gutter="20" class="mt-4">
        <!-- 磁盘详情 -->
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>磁盘详情</span>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="总空间">{{
                formatSizeGB(systemData.disk.total_gb)
              }}</el-descriptions-item>
              <el-descriptions-item label="已使用">{{
                formatSizeGB(systemData.disk.used_gb)
              }}</el-descriptions-item>
              <el-descriptions-item label="空闲">{{
                formatSizeGB(systemData.disk.free_gb)
              }}</el-descriptions-item>
              <el-descriptions-item label="使用率"
                >{{ systemData.disk.usage_percent }}%</el-descriptions-item
              >
              <el-descriptions-item label="读取次数">{{
                systemData.disk.read_count
              }}</el-descriptions-item>
              <el-descriptions-item label="写入次数">{{
                systemData.disk.write_count
              }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>

        <!-- 数据库连接监控 -->
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>数据库连接</span>
            </template>
            <div v-if="databaseData" class="db-stats">
              <el-space direction="vertical" :size="16" style="width: 100%">
                <div class="stat-item">
                  <div class="stat-label">活跃连接数</div>
                  <el-progress
                    :percentage="getDBConnectionPercent(databaseData.active)"
                    :format="() => databaseData.active"
                  />
                </div>
                <div class="stat-item">
                  <div class="stat-label">空闲连接数</div>
                  <div class="stat-value">{{ databaseData.idle }}</div>
                </div>
                <div class="stat-item">
                  <div class="stat-label">最大连接数</div>
                  <div class="stat-value">{{ databaseData.max }}</div>
                </div>
                <div class="stat-item">
                  <div class="stat-label">等待次数</div>
                  <div class="stat-value">{{ databaseData.wait_count }}</div>
                </div>
              </el-space>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 缓存监控 -->
      <el-row :gutter="20" class="mt-4">
        <el-col :span="24">
          <el-card>
            <template #header>
              <span>缓存监控</span>
            </template>
            <div v-if="cacheData" class="cache-stats">
              <el-space
                direction="horizontal"
                :size="40"
                style="width: 100%; justify-content: space-around"
              >
                <div class="stat-item-center">
                  <div class="stat-label">命中率</div>
                  <div
                    class="stat-big-value"
                    :style="{ color: getCacheHitRateColor(cacheData.hit_rate) }"
                  >
                    {{ cacheData.hit_rate }}%
                  </div>
                </div>
                <div class="stat-item-center">
                  <div class="stat-label">命中次数</div>
                  <div class="stat-value">{{ cacheData.hits }}</div>
                </div>
                <div class="stat-item-center">
                  <div class="stat-label">未命中次数</div>
                  <div class="stat-value">{{ cacheData.misses }}</div>
                </div>
                <div class="stat-item-center">
                  <div class="stat-label">键数量</div>
                  <div class="stat-value">{{ cacheData.keys }}</div>
                </div>
                <div class="stat-item-center">
                  <div class="stat-label">占用内存</div>
                  <div class="stat-value">{{ formatSize(cacheData.memory_used) }}</div>
                </div>
              </el-space>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Monitor, Coin, Folder, Grid } from '@element-plus/icons-vue'
import { getSystemMonitor, getCacheMonitor, getDatabaseMonitor } from '@/api'

defineOptions({
  name: 'MonitorDashboard',
})

// 加载状态
const loading = ref(false)

// 系统数据
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const systemData = ref<any>({
  cpu: {
    usage_percent: 0,
    cores: 0,
    per_cpu_usage: [],
  },
  memory: {
    total_mb: 0,
    used_mb: 0,
    free_mb: 0,
    usage_percent: 0,
    available_mb: 0,
    buffers_mb: 0,
    cached_mb: 0,
    used_percent: 0,
  },
  disk: {
    total_gb: 0,
    used_gb: 0,
    free_gb: 0,
    usage_percent: 0,
    read_count: 0,
    write_count: 0,
  },
  load: {
    load1: 0,
    load5: 0,
    load15: 0,
  },
})

// 数据库数据
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const databaseData = ref<any>(null)

// 缓存数据
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const cacheData = ref<any>(null)

// 定时刷新
let refreshTimer: ReturnType<typeof setInterval> | null = null

// 格式化字节数
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// 获取系统监控数据
const fetchSystemMonitor = async () => {
  try {
    const res = await getSystemMonitor()
    if (res.code === 0) {
      systemData.value = res.data
    }
  } catch (error) {
    console.error('获取系统监控数据失败:', error)
  }
}

// 获取数据库监控数据
const fetchDatabaseMonitor = async () => {
  try {
    const res = await getDatabaseMonitor()
    if (res.code === 0) {
      databaseData.value = res.data
    }
  } catch (error) {
    console.error('获取数据库监控数据失败:', error)
  }
}

// 获取缓存监控数据
const fetchCacheMonitor = async () => {
  try {
    const res = await getCacheMonitor()
    if (res.code === 0) {
      cacheData.value = res.data
    }
  } catch (error) {
    console.error('获取缓存监控数据失败:', error)
  }
}

// 刷新所有数据
const refreshData = async () => {
  loading.value = true
  try {
    await Promise.all([fetchSystemMonitor(), fetchDatabaseMonitor(), fetchCacheMonitor()])
  } catch {
    ElMessage.error('刷新数据失败')
  } finally {
    loading.value = false
  }
}

// 获取进度条颜色
const getProgressColor = (percent: number) => {
  if (percent < 60) return '#67C23A'
  if (percent < 80) return '#E6A23C'
  return '#F56C6C'
}

// 获取缓存命中率颜色
const getCacheHitRateColor = (hitRate: number) => {
  if (hitRate >= 90) return '#67C23A'
  if (hitRate >= 70) return '#E6A23C'
  return '#F56C6C'
}

// 获取数据库连接百分比
const getDBConnectionPercent = (active: number) => {
  if (!databaseData.value || !databaseData.value.max) return 0
  return (active / databaseData.value.max) * 100
}

// 格式化大小（MB）
const formatSize = (mb: number) => {
  if (mb >= 1024) {
    return `${(mb / 1024).toFixed(2)} GB`
  }
  return `${mb.toFixed(2)} MB`
}

// 格式化大小（GB）
const formatSizeGB = (gb: number) => {
  return `${gb.toFixed(2)} GB`
}

// 启动定时刷新
const startRefreshTimer = () => {
  refreshTimer = setInterval(() => {
    refreshData()
  }, 5000) // 每 5 秒刷新一次
}

// 停止定时刷新
const stopRefreshTimer = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 组件挂载时
onMounted(() => {
  refreshData()
  startRefreshTimer()
})

// 组件卸载时
onUnmounted(() => {
  stopRefreshTimer()
})
</script>

<style scoped lang="scss">
.monitor-container {
  padding: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .mb-4 {
    margin-bottom: 20px;
  }

  .mt-4 {
    margin-top: 20px;
  }

  .metric-card {
    .metric-content {
      display: flex;
      align-items: center;
      gap: 16px;

      .metric-icon {
        width: 60px;
        height: 60px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 28px;
        color: white;

        &.cpu {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }

        &.memory {
          background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
        }

        &.disk {
          background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
        }

        &.load {
          background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
        }
      }

      .metric-info {
        flex: 1;

        .metric-label {
          font-size: 14px;
          color: #666;
          margin-bottom: 8px;
        }

        .metric-value {
          font-size: 24px;
          font-weight: bold;
          color: #333;
          margin-bottom: 8px;

          &.load-value {
            font-size: 18px;
          }
        }
      }
    }
  }

  .sub-title {
    font-size: 14px;
    font-weight: bold;
    margin-bottom: 12px;
    color: #333;
  }

  .core-item {
    display: flex;
    align-items: center;
    gap: 12px;

    span {
      width: 60px;
      font-size: 13px;
      color: #666;
    }
  }

  .db-stats,
  .cache-stats {
    padding: 10px 0;

    .stat-item {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .stat-label {
        font-size: 14px;
        color: #666;
      }

      .stat-value {
        font-size: 18px;
        font-weight: bold;
        color: #333;
      }
    }

    .stat-item-center {
      text-align: center;

      .stat-label {
        font-size: 14px;
        color: #666;
        margin-bottom: 8px;
      }

      .stat-big-value {
        font-size: 32px;
        font-weight: bold;
      }

      .stat-value {
        font-size: 24px;
        font-weight: bold;
        color: #333;
      }
    }
  }
}
</style>
