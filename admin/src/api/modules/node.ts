import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface NodeInfo {
  nodeId: string
  role: string
  status: string
  ipAddress: string
  port: number
  runningTasks: number
  maxConcurrent: number
  lastHeartbeat: string
  registeredAt: string
  startedAt: string
}

export interface NodeStatistics {
  total: number
  online: number
  offline: number
}

export const getNodeList = (): Promise<ApiResponse<NodeInfo[]>> => {
  return service({
    url: '/nodes',
    method: 'get',
  })
}

export const getNodeById = (id: string): Promise<ApiResponse<NodeInfo>> => {
  return service({
    url: `/nodes/${id}`,
    method: 'get',
  })
}

export const offlineNode = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/nodes/${id}/offline`,
    method: 'post',
  })
}

export const getNodeStatistics = (): Promise<ApiResponse<NodeStatistics>> => {
  return service({
    url: '/nodes/statistics',
    method: 'get',
  })
}
