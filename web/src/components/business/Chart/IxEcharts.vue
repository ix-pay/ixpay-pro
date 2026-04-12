<template>
  <div ref="chartRef" class="ix-echarts" :style="{ height: height, width: width }"></div>
</template>

<script setup lang="ts">
defineOptions({
  name: 'IxEcharts',
})

import { ref, onMounted, onBeforeUnmount, watch, shallowRef } from 'vue'
import { echarts } from '@/utils/echarts'
import type { EChartsOption } from 'echarts'

interface Props {
  // 图表配置项
  options: EChartsOption
  // 图表宽度
  width?: string
  // 图表高度
  height?: string
  // 是否自动调整大小
  autoResize?: boolean
  // 主题
  theme?: string
}

const props = withDefaults(defineProps<Props>(), {
  width: '100%',
  height: '400px',
  autoResize: true,
  theme: '',
})

// 图表实例
const chartRef = ref<HTMLElement | null>(null)
const chartInstance = shallowRef<echarts.ECharts | null>(null)

// 初始化图表
const initChart = (): void => {
  if (!chartRef.value) return

  // 如果已有实例，先销毁
  if (chartInstance.value) {
    chartInstance.value.dispose()
  }

  // 创建新实例
  chartInstance.value = echarts.init(chartRef.value, props.theme || undefined)

  // 设置配置项
  chartInstance.value.setOption(props.options, true)
}

// 更新图表配置
const updateOptions = (newOptions: EChartsOption): void => {
  if (!chartInstance.value) return
  chartInstance.value.setOption(newOptions, true)
}

// 监听配置变化
watch(
  () => props.options,
  (newOptions) => {
    if (chartInstance.value) {
      chartInstance.value.setOption(newOptions, true)
    }
  },
  { deep: true },
)

// 监听主题变化
watch(
  () => props.theme,
  () => {
    initChart()
  },
)

// 页面大小变化时重新调整图表
const handleResize = (): void => {
  if (chartInstance.value && props.autoResize) {
    chartInstance.value.resize()
  }
}

onMounted(() => {
  initChart()

  // 监听窗口大小变化
  if (props.autoResize) {
    window.addEventListener('resize', handleResize)
  }
})

onBeforeUnmount(() => {
  // 移除监听
  if (props.autoResize) {
    window.removeEventListener('resize', handleResize)
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
})
</script>

<style scoped>
.ix-echarts {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
