// 按钮权限展示指令
import { hasButtonPermission } from '@/utils/permission'
import type { App, DirectiveBinding } from 'vue'

export default {
  install: (app: App) => {
    app.directive('auth-btn', {
      // 当被绑定的元素插入到 DOM 中时
      mounted: function (el: HTMLElement, binding: DirectiveBinding<string>) {
        // 如果没有绑定值，移除元素
        if (!binding.value) {
          el.parentNode?.removeChild(el)
          return
        }

        // 检查按钮权限
        const hasPermission = hasButtonPermission(binding.value.toString())

        // 如果没有权限，移除元素
        if (!hasPermission) {
          el.parentNode?.removeChild(el)
        }
      },
    })
  },
}
