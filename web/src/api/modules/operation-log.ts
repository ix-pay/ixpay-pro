import service from '@/utils/request'
import type { ApiResponse } from '@/types'

// @Summary 获取操作日志列表
// @Description 获取系统操作日志列表，支持分页和多条件过滤
// @Tags 系统管理
// @Security BearerAuth
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param startTime query string false "开始时间 (格式:2006-01-02)"
// @Param endTime query string false "结束时间 (格式:2006-01-02)"
// @Param userName query string false "用户名"
// @Param module query string false "操作模块"
// @Param operationType query int false "操作类型"
// @Param isSuccess query bool false "操作结果"
// @Router /api/admin/logs [get]
export const getLogList = (params: {
  page: number
  pageSize: number
  startTime?: string
  endTime?: string
  userName?: string
  module?: string
  operationType?: number
  isSuccess?: boolean
}): Promise<ApiResponse> => {
  return service({
    url: '/logs',
    method: 'get',
    params: params,
  })
}

// @Summary 根据 ID 获取操作日志
// @Description 根据日志 ID 获取详细的操作日志信息
// @Tags 系统管理
// @Security BearerAuth
// @Param id path int true "日志 ID"
// @Router /api/admin/logs/{id} [get]
export const getLogByID = (id: string): Promise<ApiResponse> => {
  return service({
    url: `//logs/${id}`,
    method: 'get',
  })
}

// @Summary 根据 ID 删除操作日志
// @Description 根据日志 ID 删除指定的操作日志
// @Tags 系统管理
// @Security BearerAuth
// @Param id path int true "日志 ID"
// @Router /api/admin/logs/{id} [delete]
export const deleteLogByID = (id: string): Promise<ApiResponse> => {
  return service({
    url: `//logs/${id}`,
    method: 'delete',
  })
}

// @Summary 批量删除操作日志
// @Description 批量删除指定 ID 列表的操作日志
// @Tags 系统管理
// @Security BearerAuth
// @Param request body map[string][]int64 true "批量删除请求参数"
// @Router /api/admin/logs/batch-delete [post]
export const batchDeleteLog = (data: { ids: string[] }): Promise<ApiResponse> => {
  return service({
    url: '/logs/batch-delete',
    method: 'post',
    data: data,
  })
}

// @Summary 获取操作日志统计信息
// @Description 获取操作日志的统计信息，包括操作类型分布等
// @Tags 系统管理
// @Security BearerAuth
// @Param startTime query string false "开始时间 (格式:2006-01-02)"
// @Param endTime query string false "结束时间 (格式:2006-01-02)"
// @Router /api/admin/logs/statistics [get]
export const getLogStatistics = (params?: {
  startTime?: string
  endTime?: string
}): Promise<ApiResponse> => {
  return service({
    url: '/logs/statistics',
    method: 'get',
    params: params,
  })
}

// @Summary 根据时间范围清空操作日志
// @Description 根据时间范围清空操作日志
// @Tags 系统管理
// @Security BearerAuth
// @Param request body map[string]string true "时间范围参数"
// @Router /api/admin/logs/clear [post]
export const clearLogByTimeRange = (data: {
  startTime: string
  endTime: string
}): Promise<ApiResponse> => {
  return service({
    url: '/logs/clear',
    method: 'post',
    data: data,
  })
}
