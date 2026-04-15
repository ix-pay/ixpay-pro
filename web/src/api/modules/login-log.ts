import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface LoginLog {
  id: number
  userId: number
  userName: string
  ip: string
  location: string
  browser: string
  os: string
  status: number
  message: string
  loginTime: string
}

// 获取登录日志列表
export const getLoginLogList = (params?: {
  page?: number
  pageSize?: number
  userName?: string
  status?: number
  startTime?: string
  endTime?: string
}): Promise<
  ApiResponse<{
    list: LoginLog[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/login-log',
    method: 'get',
    params,
  })
}

// 根据 ID 获取登录日志
export const getLoginLogById = (id: number): Promise<ApiResponse<LoginLog>> => {
  return service({
    url: `//login-log/${id}`,
    method: 'get',
  })
}

// 删除登录日志
export const deleteLoginLog = (id: number): Promise<ApiResponse> => {
  return service({
    url: `//login-log/${id}`,
    method: 'delete',
  })
}

// 批量删除登录日志
export const batchDeleteLoginLogs = (data: { ids: number[] }): Promise<ApiResponse> => {
  return service({
    url: '/login-log/batch-delete',
    method: 'post',
    data,
  })
}

// 清空登录日志
export const clearLoginLogs = (data: {
  startTime: string
  endTime: string
}): Promise<ApiResponse> => {
  return service({
    url: '/login-log/clear',
    method: 'post',
    data,
  })
}
