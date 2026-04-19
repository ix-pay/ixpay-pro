import { formatTimeToStr } from '@/utils/date'

/**
 * 格式化布尔值为中文文本
 * @param bool 布尔值或null
 * @returns 格式化后的中文文本（"是"、"否"或空字符串）
 */
export const formatBoolean = (bool: boolean | null): string => {
  return bool !== null ? (bool ? '是' : '否') : ''
}

/**
 * 格式化日期为指定格式字符串
 * @param time 时间（字符串、时间戳、Date对象或null）
 * @returns 格式化后的日期字符串
 */
export const formatDate = (time: string | number | Date | null | undefined): string => {
  if (!time && time !== 0) {
    return ''
  }

  try {
    const date = new Date(time)
    return formatTimeToStr(date, 'yyyy-MM-dd hh:mm:ss')
  } catch {
    console.warn('无效的日期输入:', time)
    return ''
  }
}

/**
 * 字典项接口
 */
export interface DictItem<T = unknown> {
  value: T
  label: string
}

/**
 * 根据值从字典中过滤出对应的标签
 * @param value 要查找的值
 * @param options 字典选项数组
 * @returns 对应的标签或undefined
 */
export const filterDict = <T extends DictItem>(
  value: T['value'],
  options: T[] | undefined | null,
): string | undefined => {
  if (!options || !Array.isArray(options)) {
    return undefined
  }

  const found = options.find((item) => item.value === value)
  return found?.label
}

/**
 * 根据数据源过滤出对应的标签（支持单个值或值数组）
 * @param dataSource 数据源数组
 * @param value 要查找的值或值数组
 * @returns 对应的标签、标签数组或undefined
 */
export const filterDataSource = <T extends DictItem>(
  dataSource: T[] | undefined | null,
  value: T['value'] | T['value'][],
): string | string[] | undefined => {
  if (!dataSource || !Array.isArray(dataSource)) {
    return undefined
  }

  if (Array.isArray(value)) {
    const result = value
      .map((item) => {
        const found = dataSource.find((i) => i.value === item)
        return found?.label
      })
      .filter((label): label is string => label !== undefined)

    return result.length > 0 ? result : []
  }

  const found = dataSource.find((item) => item.value === value)
  return found?.label
}

/**
 * 构建完整的服务器路径
 * @returns 完整的服务器路径
 */
const getServerPath = (): string => {
  const basePath = import.meta.env.VITE_BASE_PATH || ''
  const serverPort = import.meta.env.VITE_SERVER_PORT
  return serverPort ? `${basePath}:${serverPort}/` : basePath
}

/**
 * 格式化图片路径为完整URL（支持单个路径或路径数组）
 * @param paths 图片路径（单个字符串或字符串数组）
 * @returns 完整的图片URL数组
 */
export const formatImageUrls = (paths: string | string[] | undefined | null): string[] => {
  const imgUrls: string[] = []
  if (!paths) return imgUrls

  const serverPath = getServerPath()
  const processPath = (path: string): string => {
    return path.startsWith('http') ? path : `${serverPath}${path}`
  }

  if (Array.isArray(paths)) {
    paths.forEach((imgPath) => {
      if (imgPath) {
        imgUrls.push(processPath(imgPath))
      }
    })
  } else if (typeof paths === 'string') {
    imgUrls.push(processPath(paths))
  }

  return imgUrls
}

// 保持向后兼容
export const ReturnArrImg = formatImageUrls
export const returnArrImg = formatImageUrls

/**
 * 下载文件
 * @param url 文件路径
 */
export const downloadFile = (url: string): void => {
  if (!url) {
    console.warn('无效的文件 URL')
    return
  }

  const serverPath = getServerPath()
  const fullUrl = url.startsWith('http') ? url : `${serverPath}${url}`
  window.open(fullUrl, '_blank')
}

// 保持向后兼容
export const onDownloadFile = downloadFile

/**
 * 将十六进制颜色转换为RGB数组
 * @param hexColor 十六进制颜色字符串
 * @returns RGB数组 [r, g, b]
 */
const hexToRgb = (hexColor: string): [number, number, number] => {
  const hex = hexColor.replace('#', '')
  const rgb = hex.match(/../g)

  if (!rgb || rgb.length < 3) {
    return [0, 0, 0]
  }

  return [parseInt(rgb[0], 16), parseInt(rgb[1], 16), parseInt(rgb[2], 16)]
}

/**
 * 将RGB数组转换为十六进制颜色字符串
 * @param r 红色分量 (0-255)
 * @param g 绿色分量 (0-255)
 * @param b 蓝色分量 (0-255)
 * @returns 十六进制颜色字符串
 */
const rgbToHex = (r: number, g: number, b: number): string => {
  const clamp = (value: number): number => Math.max(0, Math.min(255, Math.round(value)))

  const toHex = (value: number): string => {
    const hex = clamp(value).toString(16)
    return hex.padStart(2, '0')
  }

  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}

/**
 * 生成颜色的变体
 * @param baseColor 基础颜色
 * @param factor 变化因子 (0-1)
 * @param targetColor 目标颜色
 * @returns 生成的颜色
 */
const generateColorVariant = (
  baseColor: string,
  factor: number,
  targetColor: [number, number, number],
): string => {
  const baseRgb = hexToRgb(baseColor)
  const clampedFactor = Math.max(0, Math.min(1, factor))

  const interpolate = (base: number, target: number) => {
    return base * (1 - clampedFactor) + target * clampedFactor
  }

  const newR = interpolate(baseRgb[0], targetColor[0])
  const newG = interpolate(baseRgb[1], targetColor[1])
  const newB = interpolate(baseRgb[2], targetColor[2])

  return rgbToHex(newR, newG, newB)
}

/**
 * 生成深色变体
 * @param baseColor 基础颜色
 * @param factor 变化因子 (0-1)
 * @returns 生成的深色
 */
const generateDarkColor = (baseColor: string, factor: number): string => {
  return generateColorVariant(baseColor, factor, [10, 10, 30])
}

/**
 * 生成浅色变体
 * @param baseColor 基础颜色
 * @param factor 变化因子 (0-1)
 * @returns 生成的浅色
 */
const generateLightColor = (baseColor: string, factor: number): string => {
  return generateColorVariant(baseColor, factor, [240, 248, 255]) // 蓝白色的 RGB 值
}

/**
 * 为颜色添加透明度
 * @param hexColor 十六进制颜色字符串
 * @param opacity 透明度 (0-1)
 * @returns RGBA颜色字符串
 */
const addOpacityToColor = (hexColor: string, opacity: number): string => {
  const [r, g, b] = hexToRgb(hexColor)
  const clampedOpacity = Math.max(0, Math.min(1, opacity))
  return `rgba(${r}, ${g}, ${b}, ${clampedOpacity})`
}

/**
 * 设置主题主颜色
 * @param primaryColor 主颜色
 * @param darkMode 深色模式状态
 */
export const setBodyPrimaryColor = (primaryColor: string, darkMode: 'light' | 'dark'): void => {
  const fmtColorFunc = darkMode === 'dark' ? generateDarkColor : generateLightColor

  document.documentElement.style.setProperty('--el-color-primary', primaryColor)
  document.documentElement.style.setProperty(
    '--el-color-primary-bg',
    addOpacityToColor(primaryColor, 0.4),
  )

  // 设置深色变体
  for (let times = 1; times <= 2; times++) {
    document.documentElement.style.setProperty(
      `--el-color-primary-dark-${times}`,
      fmtColorFunc(primaryColor, times / 10),
    )
  }

  // 设置浅色变体
  for (let times = 1; times <= 10; times++) {
    document.documentElement.style.setProperty(
      `--el-color-primary-light-${times}`,
      fmtColorFunc(primaryColor, times / 10),
    )
  }

  document.documentElement.style.setProperty(
    `--el-menu-hover-bg-color`,
    addOpacityToColor(primaryColor, 0.2),
  )
}

/**
 * 获取API基础URL
 * @returns API基础URL
 */
export const getBaseUrl = (): string => {
  const baseUrl = import.meta.env.VITE_BASE_API
  return baseUrl === '/' ? '' : baseUrl
}

/**
 * 生成UUID
 * @returns UUID字符串
 */
export const generateUUID = (): string => {
  let timestamp = new Date().getTime()

  if (
    typeof window !== 'undefined' &&
    window.performance &&
    typeof window.performance.now === 'function'
  ) {
    timestamp += performance.now()
  }

  const uuidTemplate = '00000000-0000-0000-0000-000000000000'

  return uuidTemplate.replace(/[018]/g, (char) => {
    const random = (timestamp + Math.random() * 16) % 16 | 0
    timestamp = Math.floor(timestamp / 16)

    switch (char) {
      case '0':
        return random.toString(16)
      case '1':
        return ((random & 0x3) | 0x8).toString(16)
      case '8':
        return ((random & 0x3) | 0x8).toString(16)
      default:
        return char
    }
  })
}

// 保持向后兼容
export const CreateUUID = generateUUID

/**
 * 格式化工具导出
 */
export const formatUtils = {
  formatBoolean,
  formatDate,
  filterDict,
  filterDataSource,
  formatImageUrls,
  downloadFile,
  setBodyPrimaryColor,
  getBaseUrl,
  generateUUID,
}
