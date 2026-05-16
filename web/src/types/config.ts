// 配置相关类型定义
// 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题

// 配置信息
export interface Config {
  id: string // 配置 ID
  configKey: string // 配置键
  configValue: string // 配置值
  configType: string // 配置类型
  description?: string // 描述
  status?: number // 状态：1-启用 0-禁用
  createdAt?: string // 创建时间
  updatedAt?: string // 更新时间
}

// 配置列表响应
export interface ConfigListResponse {
  list: Config[]
  total: number
  page: number
  pageSize: number
}

// 创建配置请求
export interface CreateConfigRequest {
  configKey: string
  configValue: string
  configType: string
  description?: string
  status?: number
}

// 更新配置请求
export interface UpdateConfigRequest {
  id: string
  configKey?: string
  configValue?: string
  configType?: string
  description?: string
  status?: number
}
