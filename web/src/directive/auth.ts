// 权限按钮展示指令（已废弃，请使用 v-auth-btn 指令）
import type { App, DirectiveBinding } from 'vue'

export default {
  install: (app: App) => {
    app.directive('auth', {
      mounted: function (_el: HTMLElement, _binding: DirectiveBinding<string>) {
        console.warn('[v-auth] 指令已废弃，不再执行权限检查，请使用 v-auth-btn 指令替代')
      },
    })
  },
}
