// 权限按钮展示指令
import { useUserStore } from '@/stores/modules/user'
import type { App, DirectiveBinding } from 'vue'

export default {
  install: (app: App) => {
    const userStore = useUserStore()
    app.directive('auth', {
      // 当被绑定的元素插入到 DOM 中时……
      mounted: function (el: HTMLElement, binding: DirectiveBinding<string>) {
        const userInfo = userStore.userInfo
        if (!binding.value) {
          el.parentNode?.removeChild(el)
          return
        }
        const waitUse = binding.value.toString().split(',')
        let flag = waitUse.some(
          (item: string) => Number(item) === Number(userInfo.authorityId || 0),
        )
        if (binding.modifiers.not) {
          flag = !flag
        }
        if (!flag) {
          el.parentNode?.removeChild(el)
        }
      },
    })
  },
}
