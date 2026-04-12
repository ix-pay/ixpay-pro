import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Task {
  id: number
  name: string
  type: string
  cronExpression: string
  params: Record<string, unknown>
  status: number
  lastRunTime: string
  nextRunTime: string
  createdAt: string
  updatedAt: string
}

// 获取任务列表（分页）
export const getTaskList = (params?: {
  page?: number
  pageSize?: number
  name?: string
  type?: string
  status?: number
}): Promise<
  ApiResponse<{
    list: Task[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/task/list',
    method: 'get',
    params,
  })
}

// 获取所有任务
export const getAllTasks = (): Promise<ApiResponse<Task[]>> => {
  return service({
    url: '/task',
    method: 'get',
  })
}

// 根据 ID 获取任务
export const getTaskById = (id: number): Promise<ApiResponse<Task>> => {
  return service({
    url: `/task/${id}`,
    method: 'get',
  })
}

// 创建任务
export const createTask = (data: {
  name: string
  type: string
  cronExpression: string
  params: Record<string, unknown>
  status: number
}): Promise<ApiResponse<Task>> => {
  return service({
    url: '/task',
    method: 'post',
    data,
  })
}

// 更新任务
export const updateTask = (
  id: number,
  data: {
    name?: string
    type?: string
    cronExpression?: string
    params?: Record<string, unknown>
    status?: number
  },
): Promise<ApiResponse<Task>> => {
  return service({
    url: `/task/${id}`,
    method: 'put',
    data,
  })
}

// 删除任务
export const deleteTask = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}`,
    method: 'delete',
  })
}

// 执行任务
export const runTask = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}/run`,
    method: 'post',
  })
}

// 获取任务日志
export const getTaskLogs = (
  id: number,
  params?: {
    page?: number
    pageSize?: number
  },
): Promise<
  ApiResponse<{
    list: Array<{
      id: number
      taskId: number
      status: number
      message: string
      executedAt: string
      executionTime: number
    }>
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: `/task/${id}/execution-logs`,
    method: 'get',
    params,
  })
}
