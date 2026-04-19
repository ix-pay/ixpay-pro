/**
 * 主题管理工具函数
 *
 * 提供主题切换、主题状态持久化等功能
 */

/**
 * 主题配置类型
 */
export interface ThemeConfig {
  /** 是否为暗黑模式 */
  isDark: boolean
  /** 主题名称 */
  themeName: string
  /** 最后更新时间 */
  lastUpdated: number
}

/**
 * 本地存储键名
 */
const THEME_STORAGE_KEY = 'ixpay-theme-config'

/**
 * 默认主题配置
 */
const DEFAULT_THEME_CONFIG: ThemeConfig = {
  isDark: false,
  themeName: 'default',
  lastUpdated: Date.now(),
}

/**
 * 从本地存储获取主题配置
 */
export function getThemeConfig(): ThemeConfig {
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    if (stored) {
      return JSON.parse(stored) as ThemeConfig
    }
  } catch (error) {
    console.error('获取主题配置失败:', error)
  }
  return DEFAULT_THEME_CONFIG
}

/**
 * 保存主题配置到本地存储
 */
export function saveThemeConfig(config: ThemeConfig): void {
  try {
    const configWithTimestamp = {
      ...config,
      lastUpdated: Date.now(),
    }
    localStorage.setItem(THEME_STORAGE_KEY, JSON.stringify(configWithTimestamp))
  } catch (error) {
    console.error('保存主题配置失败:', error)
  }
}

/**
 * 设置暗黑模式
 * @param isDark - 是否为暗黑模式
 */
export function setDarkMode(isDark: boolean): void {
  const html = document.documentElement
  const config = getThemeConfig()

  if (isDark) {
    html.classList.add('dark')
  } else {
    html.classList.remove('dark')
  }

  saveThemeConfig({
    ...config,
    isDark,
  })
}

/**
 * 切换暗黑模式
 */
export function toggleDarkMode(): void {
  const config = getThemeConfig()
  setDarkMode(!config.isDark)
}

/**
 * 初始化主题
 * 在应用启动时调用，恢复用户上次选择的主题
 */
export function initTheme(): void {
  const config = getThemeConfig()
  setDarkMode(config.isDark)
}

/**
 * 获取当前主题状态
 */
export function getCurrentTheme(): 'light' | 'dark' {
  const config = getThemeConfig()
  return config.isDark ? 'dark' : 'light'
}

/**
 * 监听系统主题变化
 * @param callback - 主题变化回调函数
 */
export function watchSystemTheme(callback: (isDark: boolean) => void): () => void {
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

  const handleChange = (event: MediaQueryListEvent) => {
    callback(event.matches)
  }

  mediaQuery.addEventListener('change', handleChange)

  // 返回取消监听函数
  return () => {
    mediaQuery.removeEventListener('change', handleChange)
  }
}

/**
 * 自动根据系统主题设置
 * 仅在用户未手动设置主题时使用
 */
export function autoSetThemeBySystem(): void {
  const config = getThemeConfig()

  // 如果用户已经手动设置过主题，则不自动设置
  if (config.themeName !== 'default') {
    return
  }

  const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  setDarkMode(isDark)
}

/**
 * 主题预设
 */
export const THEME_PRESETS = {
  /** 经典蓝 */
  classic: {
    name: 'classic',
    primaryColor: '#3b82f6',
  },
  /** 科技紫 */
  tech: {
    name: 'tech',
    primaryColor: '#667eea',
  },
  /** 翡翠绿 */
  emerald: {
    name: 'emerald',
    primaryColor: '#10b981',
  },
  /** 珊瑚红 */
  coral: {
    name: 'coral',
    primaryColor: '#ef4444',
  },
  /** 琥珀黄 */
  amber: {
    name: 'amber',
    primaryColor: '#f59e0b',
  },
  /** 青蓝色 */
  cyan: {
    name: 'cyan',
    primaryColor: '#06b6d4',
  },
} as const

/**
 * 主题预设类型
 */
export type ThemePreset = keyof typeof THEME_PRESETS

/**
 * 应用主题预设
 * @param preset - 主题预设名称
 */
export function applyThemePreset(preset: ThemePreset): void {
  const themePreset = THEME_PRESETS[preset]
  if (!themePreset) {
    console.error('无效的主题预设:', preset)
    return
  }

  const config = getThemeConfig()
  const root = document.documentElement

  // 设置主色调
  root.style.setProperty('--primary-color', themePreset.primaryColor)

  saveThemeConfig({
    ...config,
    themeName: preset,
  })
}

/**
 * 重置主题为默认
 */
export function resetTheme(): void {
  const root = document.documentElement

  // 移除所有自定义主题设置
  root.style.removeProperty('--primary-color')

  saveThemeConfig(DEFAULT_THEME_CONFIG)
  setDarkMode(false)
}

/**
 * 获取主题预设列表
 */
export function getThemePresets(): Array<{ key: ThemePreset; name: string }> {
  return Object.entries(THEME_PRESETS).map(([key, value]) => ({
    key: key as ThemePreset,
    name: value.name,
  }))
}
