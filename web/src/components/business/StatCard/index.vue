<template>
  <div
    class="ix-stat-card rounded-2xl p-6 text-white shadow-lg hover:shadow-2xl transition-all duration-300 hover:-translate-y-1"
    :class="gradientClass"
  >
    <div class="flex items-center justify-between">
      <!-- 左侧信息区 -->
      <div class="flex-1">
        <p class="text-white/80 text-sm font-medium mb-2">
          {{ title }}
        </p>
        <h3 class="text-4xl font-bold mb-2">
          {{ formattedValue }}
        </h3>

        <!-- 变化趋势 -->
        <div v-if="change" class="flex items-center text-sm">
          <el-icon v-if="changeIcon" class="mr-1">
            <component :is="changeIcon" />
          </el-icon>
          <span :class="changeColorClass">
            {{ change }}
          </span>
          <span v-if="changeLabel" class="ml-2 text-white/60">
            {{ changeLabel }}
          </span>
        </div>

        <!-- 副标题 -->
        <p v-if="subtitle" class="text-white/60 text-xs mt-2">
          {{ subtitle }}
        </p>
      </div>

      <!-- 右侧图标区 -->
      <div v-if="icon || $slots.icon" class="bg-white/20 p-4 rounded-xl backdrop-blur-sm">
        <slot name="icon">
          <el-icon v-if="icon" :size="32">
            <component :is="icon" />
          </el-icon>
        </slot>
      </div>
    </div>

    <!-- 底部进度条 -->
    <div v-if="progress !== undefined" class="mt-4">
      <div class="flex items-center justify-between text-xs text-white/60 mb-1">
        <span>{{ progressLabel }}</span>
        <span>{{ progress }}%</span>
      </div>
      <div class="w-full bg-white/20 rounded-full h-2 overflow-hidden">
        <div
          class="bg-white h-full rounded-full transition-all duration-500"
          :style="{ width: `${progress}%` }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'
import { ArrowUp, ArrowDown } from '@element-plus/icons-vue'

defineOptions({
  name: 'IxStatCard',
})

interface Props {
  title: string
  value: string | number
  change?: string
  changeType?: 'positive' | 'negative' | 'neutral'
  changeLabel?: string
  subtitle?: string
  icon?: string | Component
  color?: 'blue' | 'green' | 'purple' | 'red' | 'orange' | 'cyan'
  progress?: number
  progressLabel?: string
}

const props = withDefaults(defineProps<Props>(), {
  change: '',
  changeType: 'positive',
  changeLabel: '',
  subtitle: '',
  icon: '',
  color: 'blue',
  progress: undefined,
  progressLabel: '进度',
})

// 渐变色背景
const gradientClass = computed(() => {
  const gradients: Record<string, string> = {
    blue: 'bg-gradient-to-br from-blue-500 to-blue-600',
    green: 'bg-gradient-to-br from-green-500 to-green-600',
    purple: 'bg-gradient-to-br from-purple-500 to-purple-600',
    red: 'bg-gradient-to-br from-red-500 to-red-600',
    orange: 'bg-gradient-to-br from-orange-500 to-orange-600',
    cyan: 'bg-gradient-to-br from-cyan-500 to-cyan-600',
  }
  return gradients[props.color] || gradients.blue
})

// 格式化数值
const formattedValue = computed(() => {
  if (typeof props.value === 'number') {
    // 大数字格式化
    if (props.value >= 100000000) {
      return (props.value / 100000000).toFixed(2) + ' 亿'
    }
    if (props.value >= 10000) {
      return (props.value / 10000).toFixed(2) + ' 万'
    }
    return props.value.toLocaleString()
  }
  return props.value
})

// 变化图标
const changeIcon = computed(() => {
  if (!props.change) return undefined
  if (props.changeType === 'positive') return ArrowUp
  if (props.changeType === 'negative') return ArrowDown
  return undefined
})

// 变化颜色
const changeColorClass = computed(() => {
  if (props.changeType === 'positive') {
    return 'text-green-300 font-medium'
  }
  if (props.changeType === 'negative') {
    return 'text-red-300 font-medium'
  }
  return 'text-white/80'
})
</script>

<style scoped>
.ix-stat-card {
  position: relative;
  overflow: hidden;
}

/* 添加光晕效果 */
.ix-stat-card::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.1) 0%, transparent 70%);
  pointer-events: none;
}
</style>
