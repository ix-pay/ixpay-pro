<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md p-4 transition-colors duration-300"
  >
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-lg font-semibold">任务统计</h2>
      <div class="flex items-center gap-3">
          <span class="update-time" v-if="lastUpdateTime">最后更新：{{ lastUpdateTime }}</span>
          <el-select
            v-model="refreshInterval"
            size="small"
            class="interval-select"
            @change="changeRefreshInterval"
          >
            <el-option label="5 秒" :value="5000" />
            <el-option label="10 秒" :value="10000" />
            <el-option label="30 秒" :value="30000" />
          </el-select>
          <el-switch
            v-model="autoRefreshEnabled"
            active-text="自动刷新"
            class="refresh-switch"
            @change="toggleAutoRefresh"
          />
          <el-button type="primary" @click="refreshData" :loading="loading" circle>
            <el-icon>
              <Refresh />
            </el-icon>
          </el-button>
        </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <el-card shadow="hover">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-500 text-sm">总任务数</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ statistics.totalTasks }}
            </p>
          </div>
          <el-icon class="text-blue-500 text-4xl">
            <List />
          </el-icon>
        </div>
      </el-card>

      <el-card shadow="hover">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-500 text-sm">已启用</p>
            <p class="text-2xl font-bold text-green-600">{{ statistics.enabledTasks }}</p>
          </div>
          <el-icon class="text-green-500 text-4xl">
            <CircleCheck />
          </el-icon>
        </div>
      </el-card>

      <el-card shadow="hover">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-500 text-sm">运行中</p>
            <p class="text-2xl font-bold text-blue-600">{{ statistics.runningTasks }}</p>
          </div>
          <el-icon class="text-blue-500 text-4xl">
            <VideoPlay />
          </el-icon>
        </div>
      </el-card>

      <el-card shadow="hover">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-500 text-sm">平均成功率</p>
            <p class="text-2xl font-bold" :class="getSuccessRateColor(statistics.avgSuccessRate)">
              {{ statistics.avgSuccessRate }}%
            </p>
          </div>
          <el-icon class="text-yellow-500 text-4xl">
            <TrendCharts />
          </el-icon>
        </div>
      </el-card>
    </div>

    <div>
      <div class="flex justify-between items-center mb-4">
        <h3 class="text-lg font-semibold">任务执行统计</h3>
      </div>

      <div v-loading="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <el-card v-for="item in statisticsList" :key="item.taskId" shadow="hover" class="task-card">
          <div class="flex flex-col gap-3">
            <div class="flex items-center justify-between">
              <div class="flex-1 min-w-0">
                <p class="text-base font-semibold text-gray-900 dark:text-white truncate">
                  {{ item.taskName || item.taskId }}
                </p>
                <p class="text-sm text-gray-500 mt-1">{{ item.taskId }}</p>
              </div>
              <el-tag v-if="item.group" size="small" type="info">
                {{ item.group }}
              </el-tag>
            </div>

            <div class="grid grid-cols-2 gap-3 pt-2 border-t border-gray-200 dark:border-gray-700">
              <div>
                <p class="text-xs text-gray-500">总执行次数</p>
                <p class="text-lg font-bold text-gray-900 dark:text-white">
                  {{ item.totalExecutes }}
                </p>
              </div>
              <div>
                <p class="text-xs text-gray-500">平均耗时</p>
                <p class="text-lg font-bold text-gray-900 dark:text-white">
                  {{ item.avgDuration }}ms
                </p>
              </div>
              <div>
                <p class="text-xs text-gray-500">成功次数</p>
                <p class="text-lg font-bold text-green-600">{{ item.successCount }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500">失败次数</p>
                <p class="text-lg font-bold text-red-600">{{ item.failedCount }}</p>
              </div>
            </div>

            <div class="flex items-center justify-between pt-2">
              <div class="flex items-center gap-2">
                <span class="text-xs text-gray-500">成功率</span>
                <el-tag :type="getSuccessRateTag(item.successRate)" size="small">
                  {{ item.successRate }}%
                </el-tag>
              </div>
            </div>

            <div
              class="space-y-2 pt-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-500"
            >
              <div class="flex justify-between">
                <span>最后执行</span>
                <span class="text-gray-900 dark:text-white">{{ item.lastExecuteAt || '-' }}</span>
              </div>
              <div class="flex justify-between">
                <span>下次执行</span>
                <span class="text-gray-900 dark:text-white">{{ item.nextExecuteAt || '-' }}</span>
              </div>
            </div>
          </div>
        </el-card>

        <el-empty
          v-if="!loading && statisticsList.length === 0"
          description="暂无任务统计数据"
          :image-size="120"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, List, CircleCheck, VideoPlay, TrendCharts } from '@element-plus/icons-vue'
import { getTaskStatistics, getTaskList } from '@/api/modules/task'
import type { TaskStatistics } from '@/api/modules/task'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

defineOptions({
  name: 'TaskStatistics',
})

const loading = ref(false)
const statisticsList = ref<TaskStatistics[]>([])

const statistics = ref({
  totalTasks: 0,
  enabledTasks: 0,
  runningTasks: 0,
  avgSuccessRate: 0,
})

// 加载统计数据函数（需要在 useAutoRefresh 之前定义）
const loadStatistics = async () => {
  loading.value = true
  try {
    const [statsRes, tasksRes] = await Promise.all([
      getTaskStatistics(),
      getTaskList({ page: 1, pageSize: 1000 }),
    ])

    statisticsList.value = (statsRes.data as TaskStatistics[]) || []

    const tasks = tasksRes.data?.list || []
    statistics.value.totalTasks = tasks.length
    statistics.value.enabledTasks = tasks.filter((t: any) => t.status === 'enabled').length
    statistics.value.runningTasks = tasks.filter((t: any) => t.statusLabel === 'running').length

    if (statisticsList.value.length > 0) {
      const totalRate = statisticsList.value.reduce((sum, s) => sum + s.successRate, 0)
      statistics.value.avgSuccessRate = Math.round(totalRate / statisticsList.value.length)
    }
  } catch (error) {
    ElMessage.error('获取统计数据失败')
    console.error('获取统计数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 定时刷新
const {
  autoRefresh: autoRefreshEnabled,
  refreshInterval,
  lastUpdateTime,
  changeRefreshInterval,
  toggleAutoRefresh,
  refreshData,
} = useAutoRefresh(loadStatistics, true, 5000)

const getSuccessRateTag = (rate: number) => {
  if (rate >= 90) return 'success'
  if (rate >= 70) return 'warning'
  return 'danger'
}

const getSuccessRateColor = (rate: number) => {
  if (rate >= 90) return 'text-green-600'
  if (rate >= 70) return 'text-yellow-600'
  return 'text-red-600'
}

onMounted(() => {
  loadStatistics()
})
</script>

<style scoped>
:deep(.el-card) {
  border-radius: 0.5rem;
}

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
</style>
