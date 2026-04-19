<template>
  <div class="theme-panel fixed right-0 top-1/2 -translate-y-1/2 z-50">
    <!-- 切换按钮 -->
    <button
      @click="togglePanel"
      class="theme-toggle-btn flex items-center justify-center w-12 h-12 rounded-l-lg shadow-lg transition-all duration-300 hover:scale-105"
      :class="
        isOpen
          ? 'bg-blue-500 text-white'
          : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200'
      "
    >
      <el-icon :size="20">
        <component :is="isOpen ? Close : Setting" />
      </el-icon>
    </button>

    <!-- 配置面板 -->
    <div
      v-show="isOpen"
      class="theme-config-panel bg-white dark:bg-gray-800 shadow-2xl rounded-l-lg p-6 w-80 transition-all duration-300"
    >
      <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-4">主题配置</h3>

      <!-- 主题模式切换 -->
      <div class="config-section mb-6">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3 block">
          主题模式
        </label>
        <div class="flex gap-2">
          <button
            v-for="mode in themeModes"
            :key="mode.value"
            @click="setThemeMode(mode.value)"
            class="flex-1 py-2 px-3 rounded-lg text-sm font-medium transition-all duration-200 border-2"
            :class="[
              config.darkMode === mode.value
                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400'
                : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-500',
            ]"
          >
            <div class="flex flex-col items-center gap-1">
              <el-icon :size="20">
                <component :is="mode.icon" />
              </el-icon>
              <span>{{ mode.label }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- 主题色选择 -->
      <div class="config-section mb-6">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3 block">
          主题颜色
        </label>
        <div class="grid grid-cols-4 gap-2">
          <button
            v-for="color in themeColors"
            :key="color.value"
            @click="setPrimaryColor(color.value)"
            class="w-full aspect-square rounded-lg transition-all duration-200 hover:scale-110 border-2"
            :class="[
              config.primaryColor === color.value
                ? 'border-gray-900 dark:border-white scale-110 shadow-md'
                : 'border-transparent',
            ]"
            :style="{ backgroundColor: color.value }"
            :title="color.name"
          />
        </div>
      </div>

      <!-- 特殊模式 -->
      <div class="config-section mb-6">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3 block">
          特殊模式
        </label>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-600 dark:text-gray-400">色弱模式</span>
            <el-switch
              v-model="config.weakness"
              @change="(val) => appStore.toggleWeakness(val as boolean)"
              size="small"
            />
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-600 dark:text-gray-400">灰度模式</span>
            <el-switch
              v-model="config.grey"
              @change="(val) => appStore.toggleGrey(val as boolean)"
              size="small"
            />
          </div>
        </div>
      </div>

      <!-- 布局设置 -->
      <div class="config-section mb-6">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3 block">
          布局设置
        </label>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-600 dark:text-gray-400">显示标签栏</span>
            <el-switch
              v-model="config.showTabs"
              @change="(val) => appStore.toggleTabs(val as boolean)"
              size="small"
            />
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-600 dark:text-gray-400">显示水印</span>
            <el-switch
              v-model="config.show_watermark"
              @change="(val) => appStore.toggleConfigWatermark(val as boolean)"
              size="small"
            />
          </div>
        </div>
      </div>

      <!-- 侧边栏设置 -->
      <div class="config-section mb-6">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3 block">
          侧边栏宽度：{{ config.layout_side_width }}px
        </label>
        <el-slider
          v-model="config.layout_side_width"
          @change="
            (val: number | number[]) =>
              appStore.toggleConfigSideWidth(Array.isArray(val) ? val[0] : val)
          "
          :min="180"
          :max="300"
          :step="10"
          :marks="{
            180: '窄',
            240: '标准',
            300: '宽',
          }"
        />
      </div>

      <!-- 重置按钮 -->
      <div class="config-section">
        <el-button @click="handleReset" class="w-full" type="info" :icon="RefreshLeft">
          重置所有配置
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAppStore } from '@/stores/modules/app'
import { Setting, Close, Sunny, Moon, Monitor, RefreshLeft } from '@element-plus/icons-vue'

defineOptions({
  name: 'ThemePanel',
})

const appStore = useAppStore()
const { config } = appStore

// 面板开关状态
const isOpen = ref(false)

// 切换面板
const togglePanel = () => {
  isOpen.value = !isOpen.value
}

// 主题模式选项
const themeModes = [
  { value: 'light', label: '浅色', icon: Sunny },
  { value: 'dark', label: '深色', icon: Moon },
  { value: 'auto', label: '自动', icon: Monitor },
]

// 主题色选项
const themeColors = [
  { name: '蓝色', value: '#3b82f6' },
  { name: '紫色', value: '#722ED1' },
  { name: '绿色', value: '#13C2C2' },
  { name: '橙色', value: '#FAAD14' },
  { name: '红色', value: '#F56C6C' },
  { name: '青色', value: '#10B981' },
  { name: '粉色', value: '#EC4899' },
  { name: '靛蓝', value: '#6366F1' },
]

// 设置主题模式
const setThemeMode = (mode: string) => {
  if (mode === 'light') {
    appStore.toggleLightMode()
  } else if (mode === 'dark') {
    appStore.toggleDarkModeForce()
  } else if (mode === 'auto') {
    appStore.toggleAutoTheme()
  }
}

// 设置主题色
const setPrimaryColor = (color: string) => {
  appStore.togglePrimaryColor(color)
}

// 重置配置
const handleReset = () => {
  appStore.resetConfig()
}
</script>

<style scoped lang="scss">
.theme-panel {
  display: flex;
  flex-direction: row;
}

.theme-toggle-btn {
  border: 1px solid var(--border-primary);
  border-right: none;
}

.theme-config-panel {
  border: 1px solid var(--border-primary);
  border-right: none;
  max-height: 80vh;
  overflow-y: auto;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-track {
    background-color: var(--bg-tertiary);
  }

  &::-webkit-scrollbar-thumb {
    background-color: var(--border-primary);
    border-radius: 3px;

    &:hover {
      background-color: var(--info-light);
    }
  }
}

.config-section {
  .el-slider {
    :deep(.el-slider__bar) {
      background-color: var(--primary-color);
    }

    :deep(.el-slider__button) {
      border-color: var(--primary-color);
    }
  }
}

// 暗黑模式适配
html.dark {
  .theme-toggle-btn {
    border-color: var(--border-primary);
  }

  .theme-config-panel {
    border-color: var(--border-primary);
  }
}
</style>
