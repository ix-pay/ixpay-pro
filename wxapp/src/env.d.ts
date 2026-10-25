/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module 'weixin-js-sdk' {
  // 微信 JS-SDK 类型声明
  interface WxConfig {
    debug?: boolean
    appId: string
    timestamp: string
    nonceStr: string
    signature: string
    jsApiList: string[]
  }

  interface WxChooseWXPay {
    timestamp: string
    nonceStr: string
    package: string
    signType: string
    paySign: string
    success?: (res: any) => void
    fail?: (err: any) => void
    cancel?: (res: any) => void
    complete?: (res: any) => void
  }

  function config(config: WxConfig): void
  function ready(callback: () => void): void
  function error(callback: (err: any) => void): void
  function chooseWXPay(params: WxChooseWXPay): void

  export { WxConfig, WxChooseWXPay }
  export default { config, ready, error, chooseWXPay }
}

// 微信 H5 支付 WeixinJSBridge 类型声明
interface WeixinJSBridge {
  invoke(
    apiName: string,
    params: Record<string, any>,
    callback: (res: { err_msg: string; [key: string]: any }) => void
  ): void
}

interface Window {
  WeixinJSBridge?: WeixinJSBridge
}