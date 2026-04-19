import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface OnlineUser {
  id: string
  userId: number
  userName: string
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
  userName?: string
}): Promise<
  ApiResponse<{
    list: OnlineUser[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/online-user',
    method: 'get',
    params,
  })
}

// 强制下线
export const forceLogout = (token: string): Promise<ApiResponse> => {
  return service({
    url: `/online-user/${token}/logout`,
    method: 'post',
  })
}

// 批量强制下线
export const batchForceLogout = (data: { tokens: string[] }): Promise<ApiResponse> => {
  return service({
    url: '/online-user/batch-logout',
    method: 'post',
    data,
  })
}

// 获取在线用户详情
export const getOnlineUserById = (userId: number): Promise<ApiResponse<OnlineUser>> => {
  return service({
    url: `/online-user/${userId}`,
    method: 'get',
  })
}

// 获取在线用户数量
export const getOnlineCount = (): Promise<ApiResponse<{ count: number }>> => {
  return service({
    url: '/online-user/count',
    method: 'get',
  })
}

// 检查用户是否在线
export const isUserOnline = (userId: number): Promise<ApiResponse<{ isOnline: boolean }>> => {
  return service({
    url: '/online-user/online',
    method: 'get',
    params: { userId },
  })
}

// 强制用户下线（使用 DELETE 方法）
export const forceOffline = (userId: number): Promise<ApiResponse> => {
  return service({
    url: `/online-user/${userId}`,
    method: 'delete',
  })
}

// 批量强制用户下线
export const batchForceOffline = (data: { userIds: number[] }): Promise<ApiResponse> => {
  return service({
    url: '/online-user/batch',
    method: 'post',
    data,
  })
}
