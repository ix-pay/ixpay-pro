// 部门相关类型定义
// 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题

// 部门信息
export interface Department {
  id: string // 部门 ID
  parentId: string // 父部门 ID
  name: string // 部门名称
  sort?: number // 排序
  status?: number // 状态：1-启用 0-禁用
  leader?: string // 负责人
  phone?: string // 联系电话
  email?: string // 邮箱
  createdAt?: string // 创建时间
  updatedAt?: string // 更新时间
  children?: Department[] // 子部门列表
}

// 部门列表响应
export interface DepartmentListResponse {
  list: Department[]
  total: number
  page: number
  pageSize: number
}

// 创建部门请求
export interface CreateDepartmentRequest {
  parentId?: string
  name: string
  sort?: number
  status?: number
  leader?: string
  phone?: string
  email?: string
}

// 更新部门请求
export interface UpdateDepartmentRequest {
  id: string
  parentId?: string
  name?: string
  sort?: number
  status?: number
  leader?: string
  phone?: string
  email?: string
}
