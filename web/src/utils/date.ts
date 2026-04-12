/**
 * 日期工具类
 * 提供日期格式化、解析等功能
 */

/**
 * 日期格式化配置接口
 */
interface FormatOptions {
  year?: 'numeric' | '2-digit'
  month?: 'numeric' | '2-digit' | 'long' | 'short' | 'narrow'
  day?: 'numeric' | '2-digit'
  hour?: 'numeric' | '2-digit'
  minute?: 'numeric' | '2-digit'
  second?: 'numeric' | '2-digit'
  hour12?: boolean
}

/**
 * 检查日期是否有效
 * @param date 日期对象
 * @returns 是否有效
 */
function isValidDate(date: Date): boolean {
  return date instanceof Date && !isNaN(date.getTime())
}

/**
 * 将输入转换为Date对象
 * @param input 日期输入（时间戳、日期字符串或Date对象）
 * @returns Date对象或null（如果输入无效）
 */
function toDate(input: number | string | Date): Date | null {
  if (input instanceof Date) {
    return input
  }

  const date = new Date(input)
  return isValidDate(date) ? date : null
}

/**
 * 日期格式化函数
 * @param date 日期对象
 * @param fmt 格式化字符串
 * @returns 格式化后的日期字符串
 */
function formatDate(date: Date, fmt: string): string {
  if (!isValidDate(date)) {
    throw new Error('Invalid date')
  }

  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const hours = date.getHours()
  const minutes = date.getMinutes()
  const seconds = date.getSeconds()
  const milliseconds = date.getMilliseconds()
  const quarter = Math.floor((month + 2) / 3)

  // 替换年份
  fmt = fmt.replace(/(y+)/g, (match) => {
    const len = match.length
    return String(year).slice(-len)
  })

  // 替换其他时间单位
  const timeUnits: Record<string, number> = {
    'M+': month,
    'd+': day,
    'h+': hours,
    'm+': minutes,
    's+': seconds,
    'q+': quarter,
    S: milliseconds,
  }

  for (const [key, value] of Object.entries(timeUnits)) {
    const reg = new RegExp(`(${key})`)
    fmt = fmt.replace(reg, (match) => {
      if (match.length === 1) {
        return String(value)
      }
      return String(value).padStart(match.length, '0')
    })
  }

  return fmt
}

/**
 * 格式化时间为字符串
 * @param input 日期输入（时间戳、日期字符串或Date对象）
 * @param pattern 格式化模式
 * @returns 格式化后的日期字符串
 */
export function formatTimeToStr(
  input: number | string | Date,
  pattern: string = 'yyyy-MM-dd hh:mm:ss',
): string {
  const date = toDate(input)
  if (!date) {
    console.warn('Invalid date input:', input)
    return ''
  }
  return formatDate(date, pattern)
}

/**
 * 使用Intl API格式化日期
 * @param input 日期输入
 * @param options 格式化选项
 * @param locale 区域设置
 * @returns 格式化后的日期字符串
 */
export function formatDateWithIntl(
  input: number | string | Date,
  options: FormatOptions = {},
  locale: string = navigator.language,
): string {
  const date = toDate(input)
  if (!date) {
    console.warn('Invalid date input:', input)
    return ''
  }

  return new Intl.DateTimeFormat(locale, options).format(date)
}

/**
 * 获取相对时间（如3天前）
 * @param input 日期输入
 * @returns 相对时间字符串
 */
export function getRelativeTime(input: number | string | Date): string {
  const date = toDate(input)
  if (!date) {
    console.warn('Invalid date input:', input)
    return ''
  }

  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSeconds = Math.floor(diffMs / 1000)
  const diffMinutes = Math.floor(diffSeconds / 60)
  const diffHours = Math.floor(diffMinutes / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffDays > 30) {
    return formatTimeToStr(date, 'yyyy-MM-dd')
  } else if (diffDays > 0) {
    return `${diffDays}天前`
  } else if (diffHours > 0) {
    return `${diffHours}小时前`
  } else if (diffMinutes > 0) {
    return `${diffMinutes}分钟前`
  } else {
    return '刚刚'
  }
}

/**
 * 日期工具导出
 */
export const dateUtils = {
  isValidDate,
  toDate,
  formatDate,
  formatTimeToStr,
  formatDateWithIntl,
  getRelativeTime,
}
