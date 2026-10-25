// 字典相关类型定义
// 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题

// 字典项信息
export interface DictItem {
  id: string // 字典项 ID
  dictId: string // 字典 ID
  itemKey: string // 字典项键
  itemValue: string // 字典项值
  sort?: number // 排序
  description?: string // 描述
  status?: number // 状态：1-启用 0-禁用
  createdAt?: string // 创建时间
  updatedAt?: string // 更新时间
}

// 字典信息
export interface Dict {
  id: string // 字典 ID
  dictName: string // 字典名称
  dictCode: string // 字典编码
  description?: string // 描述
  status?: number // 状态：1-启用 0-禁用
  createdAt?: string // 创建时间
  updatedAt?: string // 更新时间
  dictItems?: DictItem[] // 字典项列表
}

// 字典列表响应
export interface DictListResponse {
  list: Dict[]
  total: number
  page: number
  pageSize: number
}

// 字典项列表响应
export interface DictItemListResponse {
  list: DictItem[]
  total: number
  page: number
  pageSize: number
}

// 创建字典请求
export interface CreateDictRequest {
  dictName: string
  dictCode: string
  description?: string
  status?: number
}

// 更新字典请求
export interface UpdateDictRequest {
  id: string
  dictName?: string
  dictCode?: string
  description?: string
  status?: number
}

// 创建字典项请求
export interface CreateDictItemRequest {
  dictId: string
  itemKey: string
  itemValue: string
  sort?: number
  description?: string
  status?: number
}

// 更新字典项请求
export interface UpdateDictItemRequest {
  id: string
  itemKey?: string
  itemValue?: string
  sort?: number
  description?: string
  status?: number
}
