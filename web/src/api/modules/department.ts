import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Department {
  id: number
  name: string
  parentId: number
  sort: number
  status: number
  createdAt: string
  updatedAt: string
  children?: Department[]
}

// 获取部门列表
export const getDepartmentList = (params?: {
  parentId?: number
}): Promise<ApiResponse<Department[]>> => {
  return service({
    url: '/dept',
    method: 'get',
    params,
  })
}

// 创建部门
export const createDepartment = (data: {
  name: string
  parentId: number
  sort: number
  status: number
}): Promise<ApiResponse<Department>> => {
  return service({
    url: '/dept',
    method: 'post',
    data,
  })
}

// 更新部门
export const updateDepartment = (
  id: number,
  data: {
    name?: string
    parentId?: number
    sort?: number
    status?: number
  },
): Promise<ApiResponse<Department>> => {
  return service({
    url: `/dept/${id}`,
    method: 'put',
    data,
  })
}

// 删除部门
export const deleteDepartment = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/dept/${id}`,
    method: 'delete',
  })
}

// 获取部门树形结构
export const getDepartmentTree = (): Promise<ApiResponse<Department[]>> => {
  return service({
    url: '/dept/tree',
    method: 'get',
  })
}

// 获取部门详情
export const getDepartmentById = (id: number): Promise<ApiResponse<Department>> => {
  return service({
    url: `/dept/${id}`,
    method: 'get',
  })
}

// 更新部门负责人
export const updateDepartmentLeader = (
  id: number,
  data: {
    leaderId: number
  },
): Promise<ApiResponse> => {
  return service({
    url: `/dept/${id}/leader`,
    method: 'put',
    data,
  })
}
