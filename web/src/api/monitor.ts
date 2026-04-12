import request from '@/utils/request'

/**
 * 获取系统监控信息
 */
export function getSystemMonitor() {
  return request({
    url: '/monitor/system',
    method: 'get',
  })
}

/**
 * 获取缓存监控信息
 */
export function getCacheMonitor() {
  return request({
    url: '/monitor/cache',
    method: 'get',
  })
}

/**
 * 获取数据库监控信息
 */
export function getDatabaseMonitor() {
  return request({
    url: '/monitor/database',
    method: 'get',
  })
}

/**
 * 查询 Redis 键
 */
export function getRedisKeys(params?: { pattern?: string; limit?: number }) {
  return request({
    url: '/monitor/redis-keys',
    method: 'get',
    params,
  })
}

/**
 * 查询慢查询日志
 */
export function getSlowQueries() {
  return request({
    url: '/monitor/slow-queries',
    method: 'get',
  })
}
