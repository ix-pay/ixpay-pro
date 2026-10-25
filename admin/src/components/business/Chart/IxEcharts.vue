<template>
  <div
    class="ix-echarts-container relative overflow-hidden rounded-2xl border border-border-primary bg-bg-primary transition-all duration-300"
    :class="[containerClass, { 'ix-echarts--hover': hoverEffect }]"
  >
    <!-- 加载状态 -->
    <div
      v-if="loading"
      class="ix-echarts-loading absolute inset-0 z-10 flex flex-col items-center justify-center bg-bg-primary/80 backdrop-blur-sm"
    >
      <div class="ix-echarts-loading-spinner mb-3">
        <el-icon class="is-loading" :size="32">
          <Loading />
        </el-icon>
      </div>
      <p class="ix-echarts-loading-text text-sm text-text-secondary opacity-70">
        {{ loadingText }}
      </p>
    </div>

    <!-- 空数据状态 -->
    <div
      v-else-if="isEmpty"
      class="ix-echarts-empty absolute inset-0 z-10 flex flex-col items-center justify-center bg-bg-primary/80 backdrop-blur-sm"
    >
      <div class="ix-echarts-empty-icon mb-3 text-text-secondary opacity-50">
        <el-icon :size="48">
          <DataAnalysis />
        </el-icon>
      </div>
      <p class="ix-echarts-empty-text text-sm text-text-secondary opacity-70">
        {{ emptyText }}
      </p>
    </div>

    <!-- 图表容器 -->
    <div ref="chartRef" class="ix-echarts-chart" :style="{ height: computedHeight }"></div>

    <!-- 右下角装饰元素 -->
    <div
      class="ix-echarts-decoration absolute -bottom-8 -right-8 h-16 w-16 rounded-full bg-gradient-to-br from-primary/10 to-transparent blur-xl"
    ></div>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  onMounted,
  onBeforeUnmount,
  watch,
  shallowRef,
  onActivated,
  nextTick,
} from 'vue'
import { echarts } from '@/utils/echarts'
import type { EChartsOption } from 'echarts'
import { Loading, DataAnalysis } from '@element-plus/icons-vue'

defineOptions({
  name: 'IxEcharts',
})

interface Props {
  /** 图表配置项 */
  options: EChartsOption
  /** 图表高度 */
  height?: string | number
  /** 是否自动调整大小 */
  autoResize?: boolean
  /** 主题 */
  theme?: string
  /** 是否显示加载状态 */
  loading?: boolean
  /** 加载提示文字 */
  loadingText?: string
  /** 是否为空数据状态 */
  empty?: boolean
  /** 空数据提示文字 */
  emptyText?: string
  /** 是否启用悬浮效果 */
  hoverEffect?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  height: '400px',
  autoResize: true,
  theme: '',
  loading: false,
  loadingText: '加载中...',
  empty: false,
  emptyText: '暂无数据',
  hoverEffect: true,
})

// 计算高度
const computedHeight = computed(() => {
  if (typeof props.height === 'number') {
    return `${props.height}px`
  }
  return props.height
})

// 判断是否为空
const isEmpty = computed(() => {
  return (
    props.empty ||
    !props.options ||
    (Array.isArray(props.options.series) && props.options.series.length === 0)
  )
})

// 容器样式类
const containerClass = computed(() => {
  const classes: string[] = []
  if (props.loading || isEmpty.value) {
    classes.push('ix-echarts--overlay')
  }
  return classes
})

// 图表实例引用
const chartRef = ref<HTMLElement | null>(null)
const chartInstance = shallowRef<echarts.ECharts | null>(null)

// 暗黑模式检测
const isDarkMode = computed(() => {
  if (typeof window !== 'undefined') {
    return (
      document.documentElement.classList.contains('dark') ||
      window.matchMedia('(prefers-color-scheme: dark)').matches
    )
  }
  return false
})

// 获取图表主题
const getChartTheme = () => {
  if (props.theme) {
    return props.theme
  }
  // 根据暗黑模式自动选择主题
  return isDarkMode.value ? 'dark' : undefined
}

// 初始化图表
const initChart = (): void => {
  if (!chartRef.value) return

  // 如果已有实例，先销毁
  if (chartInstance.value) {
    chartInstance.value.dispose()
  }

  // 创建新实例
  chartInstance.value = echarts.init(chartRef.value, getChartTheme())

  // 设置配置项
  if (!isEmpty.value && props.options) {
    chartInstance.value.setOption(props.options, true)
  }
}

// 更新图表配置
const updateOptions = (newOptions: EChartsOption, notMerge = true): void => {
  if (!chartInstance.value) return
  chartInstance.value.setOption(newOptions, notMerge)
}

// 监听配置变化
watch(
  () => props.options,
  (newOptions) => {
    if (chartInstance.value && !isEmpty.value && newOptions) {
      chartInstance.value.setOption(newOptions, true)
    }
  },
  { deep: true },
)

// 监听主题变化
watch(
  () => [props.theme, isDarkMode.value],
  () => {
    // 使用防抖避免频繁初始化
    if (resizeTimer) {
      clearTimeout(resizeTimer)
    }
    resizeTimer = setTimeout(() => {
      initChart()
    }, 100)
  },
)

// 监听 loading 和 empty 状态变化
watch(
  () => [props.loading, props.empty],
  () => {
    // 状态变化时重新调整图表大小
    nextTick(() => {
      handleResize()
    })
  },
)

// 页面大小变化时重新调整图表
let resizeTimer: ReturnType<typeof setTimeout> | null = null

const handleResize = (): void => {
  if (chartInstance.value && props.autoResize) {
    // 使用防抖避免频繁 resize
    if (resizeTimer) {
      clearTimeout(resizeTimer)
    }
    resizeTimer = setTimeout(() => {
      chartInstance.value?.resize({
        width: chartRef.value?.clientWidth,
        height: chartRef.value?.clientHeight,
      })
    }, 100)
  }
}

onMounted(() => {
  initChart()

  // 监听窗口大小变化
  if (props.autoResize) {
    window.addEventListener('resize', handleResize)
  }
})

// keep-alive 激活时重新调整图表
onActivated(() => {
  nextTick(() => {
    handleResize()
  })
})

onBeforeUnmount(() => {
  // 移除监听
  if (props.autoResize) {
    window.removeEventListener('resize', handleResize)
  }

  // 清除定时器
  if (resizeTimer) {
    clearTimeout(resizeTimer)
    resizeTimer = null
  }

  // 销毁图表实例
  if (chartInstance.value) {
    chartInstance.value.dispose()
    chartInstance.value = null
  }
})

// 暴露方法和属性
defineExpose({
  chartInstance,
  updateOptions,
  resize: handleResize,
  refresh: initChart,
})
</script>

<style lang="scss" scoped>
.ix-echarts {
  &-container {
    position: relative;
    overflow: hidden;
    border-radius: var(--radius-xl, 1rem);
    border: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
    background-color: var(--bg-primary, #ffffff);
    transition: all var(--duration-normal, 300ms) var(--ease-in-out, cubic-bezier(0.4, 0, 0.2, 1));

    // 悬浮效果
    &.ix-echarts--hover {
      &:hover {
        box-shadow: var(
          --shadow-xl,
          0 20px 25px -5px rgba(0, 0, 0, 0.1),
          0 10px 10px -5px rgba(0, 0, 0, 0.04)
        );
        border-color: var(--primary, #3b82f6);

        .ix-echarts-decoration {
          opacity: 1;
        }
      }
    }

    // 覆盖层状态（loading/empty）
    &.ix-echarts--overlay {
      .ix-echarts-chart {
        opacity: 0.3;
      }
    }
  }

  &-chart {
    width: 100%;
    height: 100%;
    min-height: 200px;
    transition: opacity var(--duration-normal, 300ms);
  }

  &-loading,
  &-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background-color: var(--bg-primary, rgba(255, 255, 255, 0.8));
    backdrop-filter: blur(8px);
    z-index: 10;
  }

  &-loading-spinner {
    color: var(--primary, #3b82f6);
    animation: spin 1s linear infinite;
  }

  &-empty-icon {
    color: var(--text-secondary, #6b7280);
    opacity: 0.5;
  }

  &-loading-text,
  &-empty-text {
    font-size: var(--text-sm, 0.875rem);
    color: var(--text-secondary, #6b7280);
    opacity: 0.7;
  }

  &-decoration {
    border-radius: 50%;
    background: linear-gradient(135deg, var(--primary, #3b82f6) 0%, transparent 100%);
    opacity: 0.1;
    filter: blur(20px);
    transition: opacity var(--duration-normal, 300ms);
    pointer-events: none;
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

// 暗黑模式适配
:deep(.dark) {
  .ix-echarts {
    &-container {
      background-color: var(--bg-secondary, rgba(30, 41, 59, 0.8));
      border-color: var(--border-secondary, rgba(59, 130, 246, 0.2));
      backdrop-filter: blur(8px);
    }

    &-loading,
    &-empty {
      background-color: var(--bg-secondary, rgba(30, 41, 59, 0.8));
    }
  }
}
</style>
