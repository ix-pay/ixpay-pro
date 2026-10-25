import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Task {
  id: string
  taskId: string
  taskType: string
  type: string
  expression: string
  description: string
  group: string
  status: string
  statusLabel: string
  createdAt: string
  lastRunAt: string
  nextRunAt: string
  retryCount: number
  maxRetries: number
  concurrency?: string
  timeout?: number
  assignedNode?: string
  nodeGroup?: string
}

export interface TaskLog {
  id: string
  taskId: string
  taskName: string
  group: string
  executeAt: string
  duration: number
  result: string
  errorInfo: string
  retryCount: number
  cronExpr: string
  triggerType: string
  operatorId: string
}

export interface TaskStatistics {
  taskId: string
  taskName: string
  group: string
  totalExecutes: number
  successCount: number
  failedCount: number
  successRate: number
  avgDuration: number
  lastExecuteAt: string
  nextExecuteAt: string
}

export const getTaskList = (params?: {
  page?: number
  pageSize?: number
  taskId?: string
  taskType?: string
}): Promise<ApiResponse<{ list: Task[]; total: number }>> => {
  return service({
    url: '/task',
    method: 'get',
    params,
  })
}

export const getAllTasks = (): Promise<ApiResponse<Task[]>> => {
  return service({
    url: '/task',
    method: 'get',
  })
}

export const getTaskById = (id: string): Promise<ApiResponse<Task>> => {
  return service({
    url: `/task/${id}`,
    method: 'get',
  })
}

export const createTask = (data: {
  taskId: string
  taskType: string
  type: string
  expression: string
  description: string
  group?: string
  retryCount: number
  params: Record<string, unknown>
}): Promise<ApiResponse<Task>> => {
  return service({
    url: '/task',
    method: 'post',
    data,
  })
}

export const updateTask = (
  id: string,
  data: {
    taskId: string
    taskType: string
    type: string
    expression: string
    description: string
    group?: string
    retryCount: number
    params: Record<string, unknown>
  },
): Promise<ApiResponse<Task>> => {
  return service({
    url: `/task/${id}`,
    method: 'put',
    data,
  })
}

export const deleteTask = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}`,
    method: 'delete',
  })
}

export const enableTask = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}/enable`,
    method: 'post',
  })
}

export const disableTask = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}/disable`,
    method: 'post',
  })
}

export const runTask = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}/start`,
    method: 'post',
  })
}

export const getTaskLogs = (
  id: string,
  params?: {
    page?: number
    pageSize?: number
  },
): Promise<ApiResponse<{ logs: TaskLog[]; total: number }>> => {
  return service({
    url: `/task/${id}/execution-logs`,
    method: 'get',
    params,
  })
}

export const searchTaskLogs = (params?: {
  taskId?: string
  result?: string
  startDate?: string
  endDate?: string
  page?: number
  pageSize?: number
}): Promise<ApiResponse<{ logs: TaskLog[]; total: number }>> => {
  return service({
    url: '/task/execution-logs',
    method: 'get',
    params,
  })
}

export const startTask = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}/start`,
    method: 'post',
  })
}

export const stopTask = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}/stop`,
    method: 'post',
  })
}

export const retryTask = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}/retry`,
    method: 'post',
  })
}

export const getTaskStatistics = (): Promise<ApiResponse<TaskStatistics[]>> => {
  return service({
    url: '/task/statistics',
    method: 'get',
  })
}

export const setTaskGroup = (id: string, group: string): Promise<ApiResponse> => {
  return service({
    url: `/task/${id}/group`,
    method: 'post',
    data: { group },
  })
}

// 获取任务统计面板数据
export const getTaskDashboard = (): Promise<
  ApiResponse<{
    totalTasks: number
    enabledTasks: number
    disabledTasks: number
    todayExecutions: number
  }>
> => {
  return service({
    url: '/task/dashboard',
    method: 'get',
  })
}
