// 事件监控相关类型定义
// 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题

// 事件信息
export interface Event {
  id: string // 事件 ID
  name: string // 事件名称
  type: string // 事件类型
  source: string // 事件来源
  status: string // 状态（pending/processing/success/failed/dead）
  payload: string // 事件负载（JSON 字符串）
  retryCount: number // 重试次数
  maxRetries: number // 最大重试次数
  createdAt: string // 创建时间
  updatedAt: string // 更新时间
  processedAt?: string // 处理时间
  errorMessage?: string // 错误信息
}

// 死信队列事件
export interface DeadLetterEvent {
  id: string // 事件 ID
  eventId: string // 原始事件 ID
  eventName: string // 事件名称
  source: string // 事件来源
  payload: string // 事件负载
  errorMessage: string // 错误信息
  retryCount: number // 重试次数
  createdAt: string // 创建时间
  deadLetterAt: string // 进入死信队列时间
}

// 订阅者信息
export interface Subscriber {
  id: string // 订阅者 ID
  name: string // 订阅者名称
  eventType: string // 订阅的事件类型
  endpoint: string // 回调地址
  status: string // 状态（active/inactive）
  createdAt: string // 创建时间
  updatedAt: string // 更新时间
  lastTriggeredAt?: string // 最后触发时间
}

// 事件详情
export interface EventDetail {
  id: string // 事件 ID
  name: string // 事件名称
  type: string // 事件类型
  source: string // 事件来源
  status: string // 状态
  payload: Record<string, unknown> // 事件负载（解析后的对象）
  retryCount: number // 重试次数
  maxRetries: number // 最大重试次数
  createdAt: string // 创建时间
  updatedAt: string // 更新时间
  processedAt?: string // 处理时间
  errorMessage?: string // 错误信息
  history: EventHistoryItem[] // 处理历史
}

// 事件历史记录
export interface EventHistoryItem {
  id: string // 记录 ID
  eventId: string // 事件 ID
  status: string // 状态
  errorMessage?: string // 错误信息
  retryCount: number // 重试次数
  createdAt: string // 创建时间
}

// 发布事件请求
export interface PublishEventRequest {
  name: string // 事件名称
  type: string // 事件类型
  source: string // 事件来源
  payload: Record<string, unknown> // 事件负载
}

// 创建订阅者请求
export interface CreateSubscriberRequest {
  name: string // 订阅者名称
  eventType: string // 订阅的事件类型
  endpoint: string // 回调地址
}

// 事件搜索参数
export interface EventSearchParams {
  page?: number // 页码
  pageSize?: number // 每页条数
  name?: string // 事件名称
  status?: string // 状态
  source?: string // 来源
}

// 订阅者搜索参数
export interface SubscriberSearchParams {
  page?: number // 页码
  pageSize?: number // 每页条数
  name?: string // 订阅者名称
  eventType?: string // 事件类型
}

// 事件统计面板
export interface EventDashboard {
  totalEvents: number // 事件总数
  pendingEvents: number // 待处理数
  successEvents: number // 成功数
  failedEvents: number // 失败数
  deadLetterCount: number // 死信队列数
}
