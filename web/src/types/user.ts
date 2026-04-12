// 用户相关类型定义
// 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题

import type { Role } from './role'

// 用户信息响应（GET /api/admin/user/info）
export interface UserInfo {
  id: string // 用户 ID，string 格式
  username: string // 用户名
  nickname: string // 昵称
  email: string // 邮箱
  phone: string // 手机号
  avatar: string // 头像
  status: number // 状态：1-启用 0-禁用
  roles: Role[] // 角色列表
  currentRoleId: string // 当前角色 ID，string 格式（原 currentRoleId）
  role: string // 当前角色名称
  authority: {
    defaultRouter: string // 默认路由
  }
}

// 用户信息响应（用于登录等场景）
export interface UserResponse {
  user: UserInfo
  accessToken: string
  refreshToken: string
}

// 用户列表项
export interface UserListItem {
  id: string
  username: string
  nickname: string
  email: string
  phone: string
  avatar: string
  status: number
  roleIds: string[]
  createdAt: string
  updatedAt: string
}

// 用户列表响应
export interface UserListResponse {
  items: UserListItem[]
  total: number
  page: number
  pageSize: number
}

// 创建用户请求
export interface CreateUserRequest {
  username: string
  password: string
  nickname?: string
  email?: string
  phone?: string
  avatar?: string
  roleIds?: string[]
  status?: number
}

// 更新用户请求
export interface UpdateUserRequest {
  id: string
  username?: string
  nickname?: string
  email?: string
  phone?: string
  avatar?: string
  roleIds?: string[]
  status?: number
}

// 重置密码请求
export interface ResetPasswordRequest {
  id: string
  password: string
}
