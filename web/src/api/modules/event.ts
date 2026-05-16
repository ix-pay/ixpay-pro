import service from '@/utils/request'
import type { ApiResponse } from '@/types'
import type {
  Event,
  EventDetail,
  DeadLetterEvent,
  Subscriber,
  PublishEventRequest,
  CreateSubscriberRequest,
  EventSearchParams,
  SubscriberSearchParams,
  EventDashboard,
} from '@/types/event'

// 获取事件列表（分页）
export const getEventList = (
  params?: EventSearchParams,
): Promise<ApiResponse<{ list: Event[]; total: number }>> => {
  return service({
    url: '/event',
    method: 'get',
    params,
  })
}

// 获取事件详情
export const getEventById = (id: string): Promise<ApiResponse<EventDetail>> => {
  return service({
    url: `/event/${id}`,
    method: 'get',
  })
}

// 发布事件
export const publishEvent = (data: PublishEventRequest): Promise<ApiResponse<Event>> => {
  return service({
    url: '/event',
    method: 'post',
    data,
  })
}

// 重试事件
export const retryEvent = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/event/${id}/retry`,
    method: 'post',
  })
}

// 获取死信队列列表
export const getDeadLetters = (params?: {
  page?: number
  pageSize?: number
}): Promise<ApiResponse<{ list: DeadLetterEvent[]; total: number }>> => {
  return service({
    url: '/event/dead-letters',
    method: 'get',
    params,
  })
}

// 重试死信事件
export const retryDeadLetter = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/event/dead-letters/${id}/retry`,
    method: 'post',
  })
}

// 获取订阅者列表
export const getSubscriberList = (
  params?: SubscriberSearchParams,
): Promise<ApiResponse<{ list: Subscriber[]; total: number }>> => {
  return service({
    url: '/subscriber',
    method: 'get',
    params,
  })
}

// 创建订阅者
export const createSubscriber = (
  data: CreateSubscriberRequest,
): Promise<ApiResponse<Subscriber>> => {
  return service({
    url: '/subscriber',
    method: 'post',
    data,
  })
}

// 删除订阅者
export const deleteSubscriber = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/subscriber/${id}`,
    method: 'delete',
  })
}

// 获取事件统计
export const getEventStatistics = (): Promise<ApiResponse<EventDashboard>> => {
  return service({
    url: '/event/statistics',
    method: 'get',
  })
}
