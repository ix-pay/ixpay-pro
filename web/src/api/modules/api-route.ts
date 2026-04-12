import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface ApiRoute {
  id: string // 使用 string 类型，避免 int64 精度丢失
  path: string
  method: string
  name: string
  description: string
  group: string
  authRequired: boolean
  status: number
  createdAt: string
  updatedAt: string
}

export interface ApiRouteListParams {
  page?: number
  pageSize?: number
  group?: string
  authRequired?: boolean
}

// 获取 API 路由列表（分页）
export const getApiRouteList = (
  params?: ApiRouteListParams,
): Promise<
  ApiResponse<{
    list: ApiRoute[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/apis',
    method: 'get',
    params,
  })
}

// 获取所有 API 路由（已废弃，统一使用 getApiRouteList）
export const getAllApiRoutes = (): Promise<ApiResponse<ApiRoute[]>> => {
  return service({
    url: '/apis',
    method: 'get',
  })
}

// 根据 ID 获取 API 路由
export const getApiRouteById = (id: string): Promise<ApiResponse<ApiRoute>> => {
  return service({
    url: `//apis/${id}`,
    method: 'get',
  })
}

// 创建 API 路由
export const createApiRoute = (data: Partial<ApiRoute>): Promise<ApiResponse<ApiRoute>> => {
  return service({
    url: '/apis',
    method: 'post',
    data,
  })
}

// 更新 API 路由
export const updateApiRoute = (
  id: string,
  data: Partial<ApiRoute>,
): Promise<ApiResponse<ApiRoute>> => {
  return service({
    url: `//apis/${id}`,
    method: 'put',
    data,
  })
}

// 删除 API 路由
export const deleteApiRoute = (id: string): Promise<ApiResponse> => {
  return service({
    url: `//apis/${id}`,
    method: 'delete',
  })
}
