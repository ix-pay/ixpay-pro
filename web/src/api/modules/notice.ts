import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Notice {
  id: number
  title: string
  content: string
  type: string
  status: number
  publishTime: string
  createdAt: string
  updatedAt: string
}

// 获取公告列表
export const getNoticeList = (params?: {
  page?: number
  pageSize?: number
  title?: string
  type?: string
  status?: number
}): Promise<
  ApiResponse<{
    list: Notice[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/notices',
    method: 'get',
    params,
  })
}

// 创建公告
export const createNotice = (data: {
  title: string
  content: string
  type: string
  status: number
  publishTime?: string
}): Promise<ApiResponse<Notice>> => {
  return service({
    url: '/notices',
    method: 'post',
    data,
  })
}

// 更新公告
export const updateNotice = (
  id: number,
  data: {
    title?: string
    content?: string
    type?: string
    status?: number
    publishTime?: string
  },
): Promise<ApiResponse<Notice>> => {
  return service({
    url: `/notices/${id}`,
    method: 'put',
    data,
  })
}

// 删除公告
export const deleteNotice = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/notices/${id}`,
    method: 'delete',
  })
}

// 发布/取消发布公告
export const publishNotice = (id: number, publish: boolean): Promise<ApiResponse> => {
  return service({
    url: `/notices/${id}/publish`,
    method: 'post',
    data: { publish },
  })
}

// 获取公告详情
export const getNoticeById = (id: number): Promise<ApiResponse<Notice>> => {
  return service({
    url: `/notices/${id}`,
    method: 'get',
  })
}

// 标记公告已读
export const markNoticeAsRead = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/notices/${id}/read`,
    method: 'post',
  })
}

// 检查公告是否已读
export const isNoticeRead = (id: number): Promise<ApiResponse<{ isRead: boolean }>> => {
  return service({
    url: `/notices/${id}/is-read`,
    method: 'get',
  })
}

// 获取公告统计
export const getNoticeStatistics = (): Promise<
  ApiResponse<{ total: number; read: number; unread: number }>
> => {
  return service({
    url: '/notices/statistics',
    method: 'get',
  })
}
