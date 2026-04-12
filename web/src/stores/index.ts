import { createPinia } from 'pinia'
import type { PiniaPluginContext } from 'pinia'
import { useAppStore } from '@/stores/modules/app'
import { useUserStore } from '@/stores/modules/user'
import { useRouterStore } from '@/stores/modules/router'

// 简化的持久化插件
const createPersistPlugin = () => {
  return (context: PiniaPluginContext) => {
    const { store } = context
    const storage = localStorage
    const key = `pinia_${store.$id}`

    // 加载持久化数据
    const loadPersistedData = () => {
      try {
        const persisted = storage.getItem(key)
        if (persisted) {
          const data = JSON.parse(persisted)
          store.$patch(data)
        }
      } catch (error) {
        console.error(`Failed to load persisted state for ${store.$id}:`, error)
      }
    }

    // 保存持久化数据
    const savePersistedData = () => {
      try {
        // 对于 router store，只保存特定字段
        let state = store.$state
        if (store.$id === 'router') {
          state = {
            keepAliveRouters: state.keepAliveRouters,
            asyncRouters: state.asyncRouters,
            topMenu: state.topMenu,
            leftMenu: state.leftMenu,
            topActive: state.topActive,
          }
        }

        // 简单的状态克隆，避免循环引用
        const cloneState = JSON.parse(
          JSON.stringify(state, (key, value) => {
            // 过滤函数和内部 Vue 属性
            if (typeof value === 'function' || key === '__ob__') {
              return undefined
            }
            // 过滤 parent 属性，避免循环引用（parent -> children -> parent）
            if (key === 'parent') {
              return undefined
            }
            return value
          }),
        )

        storage.setItem(key, JSON.stringify(cloneState))
      } catch (error) {
        console.error(`Failed to save persisted state for ${store.$id}:`, error)
      }
    }

    // 初始化加载
    loadPersistedData()

    // 监听状态变化
    store.$subscribe(() => {
      savePersistedData()
    })
  }
}

const store = createPinia()

// 注册持久化插件
store.use(createPersistPlugin())

export { store, useAppStore, useUserStore, useRouterStore }
