// 岗位相关类型定义
// 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题

// 岗位信息
export interface Position {
  id: string // 岗位 ID
  name: string // 岗位名称
  code: string // 岗位编码
  sort?: number // 排序
  status?: number // 状态：1-启用 0-禁用
  description?: string // 描述
  createdAt?: string // 创建时间
  updatedAt?: string // 更新时间
}

// 岗位列表响应
export interface PositionListResponse {
  list: Position[]
  total: number
  page: number
  pageSize: number
}

// 创建岗位请求
export interface CreatePositionRequest {
  name: string
  code: string
  sort?: number
  status?: number
  description?: string
}

// 更新岗位请求
export interface UpdatePositionRequest {
  id: string
  name?: string
  code?: string
  sort?: number
  status?: number
  description?: string
}
