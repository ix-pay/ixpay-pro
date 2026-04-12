import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface OnlineUser {
  id: string
  userId: number
  username: string
  nickname: string
  ip: string
  location: string
  browser: string
  os: string
  loginTime: string
  lastActiveTime: string
  token: string
}

// 获取在线用户列表
export const getOnlineUserList = (params?: {
  page?: number
  pageSize?: number
  username?: string
}): Promise<
  ApiResponse<{
    list: OnlineUser[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/online-users',
    method: 'get',
    params,
  })
}

// 强制下线
export const forceLogout = (token: string): Promise<ApiResponse> => {
  return service({
    url: `//online-users/${token}/logout`,
    method: 'post',
  })
}

// 批量强制下线
export const batchForceLogout = (data: { tokens: string[] }): Promise<ApiResponse> => {
  return service({
    url: '/online-users/batch-logout',
    method: 'post',
    data,
  })
}
