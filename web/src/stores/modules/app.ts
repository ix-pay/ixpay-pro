import { defineStore } from 'pinia'
import { ref, watchEffect, reactive, computed } from 'vue'
import { setBodyPrimaryColor } from '@/utils/format.ts'

import { useDark, usePreferredDark } from '@vueuse/core'

// 配置类型定义
interface ConfigType {
  weakness: boolean
  grey: boolean
  primaryColor: string
  showTabs: boolean
  darkMode: string
  layout_side_width: number
  layout_side_collapsed_width: number
  layout_side_item_height: number
  show_watermark: boolean
  side_mode: string
  transition_type: string
}

// 默认配置
const DEFAULT_CONFIG: ConfigType = {
  weakness: false,
  grey: false,
  primaryColor: '#3b82f6',
  showTabs: true,
  darkMode: 'auto',
  layout_side_width: 256,
  layout_side_collapsed_width: 80,
  layout_side_item_height: 48,
  show_watermark: true,
  side_mode: 'normal',
  transition_type: 'slide',
}

// 从 localStorage 加载存储的配置
const loadConfigFromStorage = (): Partial<ConfigType> => {
  try {
    const savedConfig = localStorage.getItem('ixpay-app-config')
    return savedConfig ? JSON.parse(savedConfig) : {}
  } catch (error) {
    console.error('从存储加载配置失败:', error)
    return {}
  }
}

// 保存配置到 localStorage
const saveConfigToStorage = (config: ConfigType): void => {
  try {
    localStorage.setItem('ixpay-app-config', JSON.stringify(config))
  } catch (error) {
    console.error('保存配置到存储失败:', error)
  }
}

export const useAppStore = defineStore('app', () => {
  // 设备相关状态
  const device = ref<string>('')
  const drawerSize = ref<string>('')
  const operateMinWith = ref<string>('240')

  // 侧边栏折叠状态
  const isSidebarCollapsed = ref<boolean>(false)

  // 主题相关状态
  const isDark = useDark({
    selector: 'html',
    attribute: 'class',
    valueDark: 'dark',
    valueLight: 'light',
  })

  const preferredDark = usePreferredDark()

  // 应用配置 - 合并默认配置和存储的配置
  const savedConfig = loadConfigFromStorage()
  const config = reactive<ConfigType>({
    ...DEFAULT_CONFIG,
    ...savedConfig,
  })

  // 计算属性：当前主题模式
  const currentTheme = computed<'light' | 'dark'>(() => {
    if (config.darkMode === 'auto') {
      return preferredDark.value ? 'dark' : 'light'
    }
    return config.darkMode === 'dark' ? 'dark' : 'light'
  })

  // 设备切换
  const toggleDevice = (e: string) => {
    if (e === 'mobile') {
      drawerSize.value = '100%'
      operateMinWith.value = '80'
    } else {
      drawerSize.value = '800'
      operateMinWith.value = '240'
    }
    device.value = e
  }

  // 主题切换
  const toggleDarkMode = (e: string) => {
    config.darkMode = e
  }

  // 切换浅色模式
  const toggleLightMode = () => {
    toggleDarkMode('light')
  }

  // 切换深色模式
  const toggleDarkModeForce = () => {
    toggleDarkMode('dark')
  }

  // 切换自动主题
  const toggleAutoTheme = () => {
    toggleDarkMode('auto')
  }

  // 色弱模式切换
  const toggleWeakness = (e: boolean) => {
    config.weakness = e
  }

  // 灰度模式切换
  const toggleGrey = (e: boolean) => {
    config.grey = e
  }

  // 主题色切换
  const togglePrimaryColor = (e: string) => {
    config.primaryColor = e
  }

  // 标签栏显示切换
  const toggleTabs = (e: boolean) => {
    config.showTabs = e
  }

  // 侧边栏宽度设置
  const toggleConfigSideWidth = (e: number) => {
    config.layout_side_width = e
  }

  // 侧边栏折叠宽度设置
  const toggleConfigSideCollapsedWidth = (e: number) => {
    config.layout_side_collapsed_width = e
  }

  // 侧边栏菜单项高度设置
  const toggleConfigSideItemHeight = (e: number) => {
    config.layout_side_item_height = e
  }

  // 水印显示切换
  const toggleConfigWatermark = (e: boolean) => {
    config.show_watermark = e
  }

  // 侧边栏模式切换
  const toggleSideMode = (e: string) => {
    config.side_mode = e
  }

  // 页面过渡动画切换
  const toggleTransition = (e: string) => {
    config.transition_type = e
  }

  // 重置配置
  const resetConfig = () => {
    Object.assign(config, DEFAULT_CONFIG)
  }

  // 监听系统主题变化
  watchEffect(() => {
    if (config.darkMode === 'auto') {
      isDark.value = preferredDark.value
    } else {
      isDark.value = config.darkMode === 'dark'
    }
  })

  // 监听配置变化，保存到localStorage
  watchEffect(() => {
    saveConfigToStorage(config)
  })

  // 监听色弱模式和灰色模式
  watchEffect(() => {
    document.documentElement.classList.toggle('html-weakenss', config.weakness)
    document.documentElement.classList.toggle('html-grey', config.grey)
  })

  // 监听主题色
  watchEffect(() => {
    setBodyPrimaryColor(config.primaryColor, currentTheme.value)
  })

  return {
    // 状态
    device,
    drawerSize,
    operateMinWith,
    isDark,
    config,
    currentTheme,
    isSidebarCollapsed,

    // 方法
    toggleDevice,
    toggleDarkMode,
    toggleLightMode,
    toggleDarkModeForce,
    toggleAutoTheme,
    toggleWeakness,
    toggleGrey,
    togglePrimaryColor,
    toggleTabs,
    toggleConfigSideWidth,
    toggleConfigSideCollapsedWidth,
    toggleConfigSideItemHeight,
    toggleConfigWatermark,
    toggleSideMode,
    toggleTransition,
    resetConfig,
  }
})
