// using ES6 modules
import mitt from 'mitt'

// 定义事件类型接口
export interface Events {
  'show-error': { code: number | string; message?: string; fn?: (code: string | number) => void }
  setKeepAlive: { name: string }[]
  [key: string]: unknown
  [key: symbol]: unknown
}

export const emitter = mitt<Events>()
