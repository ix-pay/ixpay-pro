import { login, getUserInfo, logout, switchRole } from '@/api/modules/user'
import router from '@/app/router/index'
import { ElLoading, ElMessage } from 'element-plus'
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { UserInfo } from '@/types'
import { useRouterStore } from './router'

export const useUserStore = defineStore('user', () => {
  // ElLoading 是一个包含 service 方法的对象，不是构造函数
  const loadingInstance = ref<import('element-plus').LoadingInstance | null>(null)
  // 初始化时从localStorage恢复token
  const token = ref<string>(localStorage.getItem('token') || '')

  const userInfo = ref<UserInfo>({
    id: '',
    userName: '',
    nickname: '',
    email: '',
    phone: '',
    avatar: '',
    status: 1,
    roles: [],
    createdAt: '',
    updatedAt: '',
  })

  // 设置token
  const setToken = (newToken: string) => {
    token.value = newToken
    if (newToken) {
      localStorage.setItem('token', newToken)
    } else {
      localStorage.removeItem('token')
    }
  }

  const setUserInfo = (val: Partial<UserInfo>) => {
    console.log('userStore - setUserInfo 原始数据:', val)
    console.log('userStore - 原始 currentRoleId:', (val as Record<string, unknown>).currentRoleId)
    console.log(
      'userStore - 原始 current_role_id:',
      (val as Record<string, unknown>).current_role_id,
    )

    // 先处理 roles 数据
    const roles = processRoles(
      (val as Record<string, unknown>).roles || (val as Record<string, unknown>).Roles,
    )

    // 处理 current_role_id 字段
    let currentRoleIdValue = String(
      (val as Record<string, unknown>).current_role_id ||
        (val as Record<string, unknown>).currentRoleId ||
        '',
    )

    // ⚠️ 特殊处理：如果 currentRoleId 为 "0" 或空，使用第一个角色的 ID
    if (currentRoleIdValue === '0' || currentRoleIdValue === '') {
      if (roles && roles.length > 0) {
        currentRoleIdValue = roles[0].id
        console.log(
          'setUserInfo - currentRoleId 为 "0" 或空，使用第一个角色 ID:',
          currentRoleIdValue,
        )
      } else {
        console.warn('setUserInfo - 用户没有任何角色')
      }
    }

    // 处理后端返回的 user 对象，确保字段名称和类型正确匹配
    const normalizedUserInfo: Partial<UserInfo> = {
      // 处理字段名称大小写差异并确保是字符串类型
      nickname: String(
        (val as Record<string, unknown>).nickname ||
          (val as Record<string, unknown>).Nickname ||
          '',
      ),
      avatar: String(
        (val as Record<string, unknown>).avatar || (val as Record<string, unknown>).Avatar || '',
      ),
      // 处理 ID 字段类型转换
      id: (val as Record<string, unknown>).id
        ? String((val as Record<string, unknown>).id)
        : (val as Record<string, unknown>).ID
          ? String((val as Record<string, unknown>).ID)
          : '',
      // 使用处理后的 roles
      roles: roles,
      // 使用处理后的 currentRoleId
      currentRoleId: currentRoleIdValue,
      // 其他字段
      userName: String(
        (val as Record<string, unknown>).userName ||
          (val as Record<string, unknown>).Username ||
          '',
      ),
      email: String(
        (val as Record<string, unknown>).email || (val as Record<string, unknown>).Email || '',
      ),
      phone: String(
        (val as Record<string, unknown>).phone || (val as Record<string, unknown>).Phone || '',
      ),
      status: Number(
        (val as Record<string, unknown>).status || (val as Record<string, unknown>).Status || 1,
      ),
      createdAt: String(
        (val as Record<string, unknown>).created_at ||
          (val as Record<string, unknown>).CreatedAt ||
          '',
      ),
      updatedAt: String(
        (val as Record<string, unknown>).updated_at ||
          (val as Record<string, unknown>).UpdatedAt ||
          '',
      ),
    }

    console.log('userStore - normalizedUserInfo.currentRoleId:', normalizedUserInfo.currentRoleId)

    // 合并用户信息
    // 注意：先展开 val，再展开 normalizedUserInfo，确保处理后的字段不会被覆盖
    const userData: UserInfo = {
      ...userInfo.value,
      ...(val as Partial<UserInfo>),
      ...normalizedUserInfo,
    }

    console.log('userStore - 合并后的 currentRoleId:', userData.currentRoleId)
    console.log('userStore - 合并后的 userData:', userData)

    userInfo.value = userData

    console.log('userStore - 最终 userInfo.currentRoleId:', userInfo.value.currentRoleId)
    console.log('userStore - 最终 userInfo:', userInfo.value)
  }

  // 处理 roles 数据，只保留简单字段，避免循环引用
  const processRoles = (rolesData: unknown): import('@/types').RoleInfo[] => {
    if (!Array.isArray(rolesData)) {
      return []
    }
    return rolesData.map((role: unknown) => {
      const r = role as Record<string, unknown>
      return {
        id: String(r.id ?? r.ID ?? ''),
        name: String(r.name ?? r.Name ?? ''),
        code: String(r.code ?? r.Code ?? ''),
        description: String(r.description ?? r.Description ?? ''),
        type: Number(r.type ?? r.Type ?? 1),
        status: Number(r.status ?? r.Status ?? 1),
        is_system: Boolean(r.is_system ?? r.IsSystem ?? false),
        sort: Number(r.sort ?? r.Sort ?? 0),
      }
    })
  }

  const NeedInit = async () => {
    await ClearStorage()
    await router.push({ name: 'Init', replace: true })
  }

  const ResetUserInfo = (value: Partial<UserInfo> = {}) => {
    userInfo.value = {
      ...userInfo.value,
      ...value,
    }
  }
  /* 获取用户信息*/
  const GetUserInfo = async (): Promise<void> => {
    const res = await getUserInfo()
    console.log('GetUserInfo - API 响应:', res)
    console.log('GetUserInfo - res.data:', res.data)
    console.log('GetUserInfo - res.data.userInfo:', res.data?.userInfo)
    console.log(
      'GetUserInfo - res.data.currentRoleId:',
      (res.data as Record<string, unknown>)?.currentRoleId,
    )

    // 后端返回的数据结构是 { code: 0, data: {...} }，而不是 { code: 0, data: { userInfo: {...} } }
    // 所以需要直接传递 res.data，而不是 res.data.userInfo
    if (res.data) {
      setUserInfo(res.data as Partial<UserInfo>)
    }
  }
  /* 登录*/
  const LoginIn = async (loginInfo: { userName: string; password: string; captcha?: string }) => {
    try {
      loadingInstance.value = ElLoading.service({
        fullscreen: true,
        text: '登录中，请稍候...',
      })

      const res = await login(loginInfo)

      // 检查登录是否成功
      if (res.code === 0 && res.data?.user) {
        // 登陆成功，设置用户信息和权限相关信息
        setUserInfo(res.data.user)

        // 保存token到localStorage和pinia store
        if (res.data?.accessToken) {
          setToken(res.data.accessToken)
        }
        // 保存refreshToken（如果需要）
        if (res.data?.refreshToken) {
          localStorage.setItem('refreshToken', res.data.refreshToken)
        }
      } else {
        // 登录失败，抛出错误
        throw new Error(res.msg || '登录失败')
      }

      // 处理重定向
      const redirect = router.currentRoute.value.query.redirect
      if (redirect && typeof redirect === 'string') {
        await router.replace(redirect)
        return true
      }

      // 先尝试直接跳转到首页路径，确保使用正确的布局
      await router.replace('/index')

      const isWindows = /windows/i.test(navigator.userAgent)
      window.localStorage.setItem('osType', isWindows ? 'WIN' : 'MAC')

      // 全部操作均结束，关闭loading并返回
      return true
    } catch (error) {
      console.error('LoginIn error:', error)
      return false
    } finally {
      loadingInstance.value?.close()
    }
  }
  /* 登出*/
  const LoginOut = async (): Promise<void> => {
    try {
      await logout()
    } catch (error) {
      console.error('Logout error:', error)
      ElMessage.error('登出失败，请重试')
      return
    }

    await ClearStorage()

    // 把路由定向到登录页
    router.replace({ name: 'Login' })
  }

  /* 切换角色 */
  const SwitchRole = async (roleId: string) => {
    try {
      const res = await switchRole({ roleId })

      if (res.code === 0) {
        // 获取 routerStore
        const routerStore = useRouterStore()

        // 1. 清理所有 tabs
        routerStore.resetTabManager()

        // 2. 清理 keep-alive 缓存
        routerStore.clearAllKeepAlive()

        // 3. 重置路由标志，允许重新加载动态路由
        routerStore.setAsyncRouterFlag(0)

        // 4. 重新获取用户信息
        await GetUserInfo()

        // 5. 重新加载新角色的菜单数据
        await routerStore.SetAsyncRouter()

        ElMessage.success('切换角色成功')
        return true
      } else {
        throw new Error(res.msg || '切换失败')
      }
    } catch (error) {
      console.error('SwitchRole error:', error)
      ElMessage.error('切换角色失败')
      return false
    }
  }
  /* 清理数据 */
  const ClearStorage = async (): Promise<void> => {
    // 使用remove方法正确删除cookie
    sessionStorage.clear()
    // 清理所有相关的localStorage项
    localStorage.removeItem('originSetting')
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    // 清理其他可能存在的用户相关数据
    localStorage.removeItem('osType')

    // 重置 pinia store 中的数据
    token.value = ''
    userInfo.value = {
      id: '',
      userName: '',
      nickname: '',
      email: '',
      phone: '',
      avatar: '',
      status: 1,
      roles: [],
      createdAt: '',
      updatedAt: '',
    }
  }

  return {
    userInfo,
    NeedInit,
    ResetUserInfo,
    GetUserInfo,
    LoginIn,
    LoginOut,
    SwitchRole,
    loadingInstance,
    ClearStorage,
    token,
    setToken,
  }
})
