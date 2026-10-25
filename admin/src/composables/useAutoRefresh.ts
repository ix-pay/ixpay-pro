import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'

/**
 * 定时刷新数据
 * @param refreshFn 刷新数据的函数
 * @param autoRefreshInitial 是否默认开启自动刷新
 * @param intervalInitial 初始刷新间隔（毫秒）
 */
export function useAutoRefresh(
  refreshFn: () => Promise<void>,
  autoRefreshInitial = true,
  intervalInitial = 5000,
) {
  const autoRefresh = ref(autoRefreshInitial)
  const refreshInterval = ref(intervalInitial)
  const lastUpdateTime = ref('')
  let refreshTimer: ReturnType<typeof setInterval> | null = null

  const getCurrentTime = () => {
    const now = new Date()
    return `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
  }

  const refreshData = async () => {
    try {
      await refreshFn()
      lastUpdateTime.value = getCurrentTime()
    } catch (error) {
      console.error('刷新数据失败:', error)
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

  const manualRefresh = async () => {
    await refreshData()
  }

  onMounted(() => {
    if (autoRefresh.value) {
      startRefreshTimer()
    }
  })

  onUnmounted(() => {
    stopRefreshTimer()
  })

  return {
    autoRefresh,
    refreshInterval,
    lastUpdateTime,
    changeRefreshInterval,
    toggleAutoRefresh,
    refreshData: manualRefresh,
  }
}
