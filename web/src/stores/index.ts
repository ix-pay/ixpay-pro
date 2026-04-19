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

        // 对于 user store，只保存简单字段，避免循环引用
        if (store.$id === 'user') {
          state = {
            token: state.token,
            userInfo: {
              id: state.userInfo?.id,
              userName: state.userInfo?.userName,
              nickname: state.userInfo?.nickname,
              email: state.userInfo?.email,
              phone: state.userInfo?.phone,
              avatar: state.userInfo?.avatar,
              status: state.userInfo?.status,
              // 只保留角色的简单信息，避免循环引用
              roles: (state.userInfo?.roles || []).map((r: { id: number | string; name: string; code: string }) => ({
                id: r.id,
                name: r.name,
                code: r.code,
              })),
              currentRoleId: state.userInfo?.currentRoleId,
            },
          }
        }

        // 简单的状态克隆，避免循环引用
        const cloneState = JSON.parse(
          JSON.stringify(state, (key, value) => {
            // 过滤函数和内部 Vue 属性
            if (typeof value === 'function' || key === '__ob__') {
              return undefined
            }
            // 过滤 parent、component 等属性，避免循环引用
            if (key === 'parent' || key === 'component' || key === 'meta') {
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
