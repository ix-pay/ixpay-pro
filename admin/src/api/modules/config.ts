import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Config {
  id: string
  configKey: string
  configValue: string
  configType: string
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
  configKey: string
  configValue: string
  configType: string
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
  id: string,
  data: {
    configKey?: string
    configValue?: string
    configType?: string
    description?: string
    status?: number
  },
): Promise<ApiResponse<Config>> => {
  return service({
    url: '/config',
    method: 'put',
    data: { id, ...data },
  })
}

// 删除配置
export const deleteConfig = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/config/${id}`,
    method: 'delete',
  })
}

// 根据 key 获取配置
export const getConfigByKey = (key: string): Promise<ApiResponse<Config>> => {
  return service({
    url: `/config/key?configKey=${key}`,
    method: 'get',
  })
}

// 根据 ID 获取配置
export const getConfigById = (id: string): Promise<ApiResponse<Config>> => {
  return service({
    url: `/config/${id}`,
    method: 'get',
  })
}

// 获取所有启用的配置
export const getActiveConfigs = (): Promise<ApiResponse<Config[]>> => {
  return service({
    url: '/config/active',
    method: 'get',
  })
}
