import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Position {
  id: number
  name: string
  code: string
  description: string
  sort: number
  status: number
  createdAt: string
  updatedAt: string
}

// 获取职位列表
export const getPositionList = (params?: {
  page?: number
  pageSize?: number
  name?: string
  status?: number
}): Promise<
  ApiResponse<{
    list: Position[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/positions',
    method: 'get',
    params,
  })
}

// 创建职位
export const createPosition = (data: {
  name: string
  code: string
  description?: string
  sort: number
  status: number
}): Promise<ApiResponse<Position>> => {
  return service({
    url: '/positions',
    method: 'post',
    data,
  })
}

// 更新职位
export const updatePosition = (
  id: number,
  data: {
    name?: string
    code?: string
    description?: string
    sort?: number
    status?: number
  },
): Promise<ApiResponse<Position>> => {
  return service({
    url: `//positions/${id}`,
    method: 'put',
    data,
  })
}

// 删除职位
export const deletePosition = (id: number): Promise<ApiResponse> => {
  return service({
    url: `//positions/${id}`,
    method: 'delete',
  })
}
