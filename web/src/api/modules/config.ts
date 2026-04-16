import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Config {
  id: number
  name: string
  key: string
  value: string
  type: string
  description: string
  status: number
  createdAt: string
  updatedAt: string
}

// 获取配置列表
export const getConfigList = (params?: {
  page?: number
  pageSize?: number
  name?: string
  key?: string
  status?: number
}): Promise<
  ApiResponse<{
    list: Config[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/config',
    method: 'get',
    params,
  })
}

// 创建配置
export const createConfig = (data: {
  name: string
  key: string
  value: string
  type: string
  description?: string
  status: number
}): Promise<ApiResponse<Config>> => {
  return service({
    url: '/config',
    method: 'post',
    data,
  })
}

// 更新配置
export const updateConfig = (
  id: number,
  data: {
    name?: string
    key?: string
    value?: string
    type?: string
    description?: string
    status?: number
  },
): Promise<ApiResponse<Config>> => {
  return service({
    url: `/config/${id}`,
    method: 'put',
    data,
  })
}

// 删除配置
export const deleteConfig = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/config/${id}`,
    method: 'delete',
  })
}

// 根据 key 获取配置
export const getConfigByKey = (key: string): Promise<ApiResponse<Config>> => {
  return service({
    url: `/config/key?config_key=${key}`,
    method: 'get',
  })
}
