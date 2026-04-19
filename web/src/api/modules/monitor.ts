import service from '@/utils/request'
import type { ApiResponse } from '@/types'

/**
 * 系统监控相关接口
 */

// @Summary 获取系统监控信息
// @Produce application/json
// @Router /monitor/system [get]
export const getSystemMonitor = (): Promise<ApiResponse> => {
  return service({
    url: '/monitor/system',
    method: 'get',
  })
}

// @Summary 获取缓存监控信息
// @Produce application/json
// @Router /monitor/cache [get]
export const getCacheMonitor = (): Promise<ApiResponse> => {
  return service({
    url: '/monitor/cache',
    method: 'get',
  })
}

// @Summary 获取数据库监控信息
// @Produce application/json
// @Router /monitor/database [get]
export const getDatabaseMonitor = (): Promise<ApiResponse> => {
  return service({
    url: '/monitor/database',
    method: 'get',
  })
}

// @Summary 查询 Redis 键
// @Produce application/json
// @Param pattern query string false "键名模式（支持通配符）"
// @Param limit query number false "返回数量限制"
// @Router /monitor/redis-keys [get]
export const getRedisKeys = (params?: {
  pattern?: string
  limit?: number
}): Promise<ApiResponse> => {
  return service({
    url: '/monitor/redis-keys',
    method: 'get',
    params: params,
  })
}

// @Summary 查询慢查询日志
// @Produce application/json
// @Router /monitor/slow-queries [get]
export const getSlowQueries = (): Promise<ApiResponse> => {
  return service({
    url: '/monitor/slow-queries',
    method: 'get',
  })
}
