import service from '@/utils/request'
import type { ApiResponse } from '@/types'
import type { MenuItem } from './menu'
import type { ApiRoute } from './api-route'
import type {
  Role as RoleType,
  RolePermissionRequest,
  RolePermissionResponse,
  DeleteImpact,
  PermissionLog,
  PermissionLogQuery,
} from '@/types/role'

export type { MenuItem, ApiRoute }
export type { RoleType as Role }

// 创建角色请求参数
export interface CreateRoleRequest {
  name: string
  description: string
  status: number
}

// 更新角色请求参数
export interface UpdateRoleRequest {
  id: string
  name: string
  description: string
  status: number
}

// 创建角色请求参数
export interface CreateRoleRequest {
  name: string
  description: string
  status: number
}

// 更新角色请求参数
export interface UpdateRoleRequest {
  id: string
  name: string
  description: string
  status: number
}

// 获取角色列表
export const getRoleList = (params?: {
  page: number
  pageSize: number
  name?: string
  status?: number
}): Promise<ApiResponse<{ list: RoleType[]; total: number; page: number; pageSize: number }>> => {
  return service({
    url: '/role',
    method: 'get',
    params,
  })
}

// 创建角色
export const createRole = (data: CreateRoleRequest): Promise<ApiResponse<RoleType>> => {
  return service({
    url: '/role',
    method: 'post',
    data,
  })
}

// 获取角色详情 (包含关联权限)
export const getRoleDetail = (id: string): Promise<ApiResponse<RoleType>> => {
  return service({
    url: '/role/detail',
    method: 'get',
    params: { id }, // 直接传递字符串，避免 Number 精度丢失
  })
}

// 更新角色
export const updateRole = (data: UpdateRoleRequest): Promise<ApiResponse> => {
  return service({
    url: '/role',
    method: 'put',
    data,
  })
}

// 删除角色
export const deleteRole = (id: string): Promise<ApiResponse> => {
  return service({
    url: '/role',
    method: 'delete',
    data: { id },
  })
}

// 分配用户到角色
export const assignUserToRole = (data: {
  roleId: number
  userIds: number[]
}): Promise<ApiResponse> => {
  return service({
    url: '/role/assign-users',
    method: 'post',
    data,
  })
}

// 分配菜单到角色
export const assignMenuToRole = (data: {
  roleId: number
  menuIds: number[]
}): Promise<ApiResponse> => {
  return service({
    url: '/role/assign-menus',
    method: 'post',
    data,
  })
}

// 分配 API 路由到角色
export const assignApiToRole = (data: { roleId: number; ids: number[] }): Promise<ApiResponse> => {
  return service({
    url: '/role/assign-api-routes',
    method: 'post',
    data,
  })
}

// 获取所有角色列表
export const getRolesList = (): Promise<ApiResponse<RoleType[]>> => {
  return service({
    url: '/role/all',
    method: 'get',
  })
}

// 获取角色可授权的菜单树（包含按钮）
// 注意：roleId 使用 string | number 类型，避免 Number() 转换导致精度丢失
export const getRoleAvailableMenus = (
  roleId: string | number,
): Promise<ApiResponse<MenuItem[]>> => {
  return service({
    url: `//roles/${roleId}/available-menus`,
    method: 'get',
  })
}

// 获取角色可授权的 API 列表（过滤已关联的 API）
// 注意：roleId 使用 string | number 类型，避免 Number() 转换导致精度丢失
export const getRoleAvailableApis = (roleId: string | number): Promise<ApiResponse<ApiRoute[]>> => {
  return service({
    url: `//roles/${roleId}/available-apis`,
    method: 'get',
  })
}

// 获取角色详情（包含所有权限）
// 注意：roleId 使用 string | number 类型，避免 Number() 转换导致精度丢失
export const getRolePermissionDetail = (
  roleId: string | number,
): Promise<ApiResponse<RolePermissionResponse>> => {
  return service({
    url: `//roles/${roleId}/detail`,
    method: 'get',
  })
}

// 保存角色权限
// 注意：roleId 使用 string | number 类型，避免 Number() 转换导致精度丢失
export const saveRolePermissions = (
  roleId: string | number,
  data: RolePermissionRequest,
): Promise<ApiResponse> => {
  return service({
    url: `//roles/${roleId}/permissions`,
    method: 'post',
    data,
  })
}

// 获取菜单删除影响评估
// 注意：menuId 使用 string | number 类型，避免 Number() 转换导致精度丢失
export const getMenuDeleteImpact = (
  menuId: string | number,
): Promise<ApiResponse<DeleteImpact>> => {
  return service({
    url: `//menu/${menuId}/delete-impact`,
    method: 'get',
  })
}

// 获取权限日志列表
export const getPermissionLogs = (
  params?: PermissionLogQuery,
): Promise<
  ApiResponse<{ list: PermissionLog[]; total: number; page: number; pageSize: number }>
> => {
  return service({
    url: '/permission-logs',
    method: 'get',
    params,
  })
}

// 获取角色权限日志
export const getRolePermissionLogs = (
  roleId: number,
  params?: { page?: number; pageSize?: number },
): Promise<
  ApiResponse<{ list: PermissionLog[]; total: number; page: number; pageSize: number }>
> => {
  return service({
    url: `//roles/${roleId}/permission-logs`,
    method: 'get',
    params,
  })
}
