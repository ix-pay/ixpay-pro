/// <reference types="vite/client" />

// 为环境变量添加类型声明
interface ImportMetaEnv {
  readonly VITE_CLI_PORT: string
  readonly VITE_SERVER_PORT: string
  readonly VITE_BASE_API: string
  readonly VITE_BASE_PATH: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

// 为Vue文件添加类型声明
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>
  export default component
}

// 为图片文件添加类型声明
declare module '*.png'
declare module '*.jpg'
declare module '*.jpeg'
declare module '*.gif'
declare module '*.svg'

// 为element-plus语言包添加类型声明
declare module 'element-plus/es/locale/lang/zh-cn'
declare module 'element-plus/es/locale/lang/en'
