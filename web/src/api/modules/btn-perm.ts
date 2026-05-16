import service from '@/utils/request'
import type { ApiResponse } from '@/types'
import type { ApiRoute } from './api-route'

/**
 * 按钮权限信息
 * 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题
 */
export interface BtnPerm {
  id: string
  menuId: string
  code: string
  name: string
  description: string
  status: number // 0-禁用，1-启用
  createdAt?: string
  updatedAt?: string
}

/**
 * 按钮权限详情（包含关联的菜单和 API 路由）
 */
export interface BtnPermDetail extends BtnPerm {
  menu?: {
    id: string
    name: string
    title: string
    path: string
  }
  apiRoutes?: ApiRoute[]
}

/**
 * 按钮权限列表响应
 */
export interface BtnPermListResponse {
  list: BtnPerm[]
  total: number
  page: number
  size: number
}

/**
 * 创建按钮权限请求参数
 */
export interface CreateBtnPermRequest {
  menuId: string
  code: string
  name: string
  description?: string
  status: number
}

/**
 * 更新按钮权限请求参数
 */
export interface UpdateBtnPermRequest {
  id: string
  menuId: string
  code: string
  name: string
  description?: string
  status: number
}

/**
 * 获取按钮权限列表请求参数
 */
export interface GetBtnPermListParams {
  page: number
  pageSize: number
  menuId?: string
  code?: string
  name?: string
  status?: number // -1-全部，0-禁用，1-启用
}

/**
 * 分配 API 路由到按钮权限请求参数
 */
export interface AssignApiRoutesRequest {
  btnPermId: string
  ids: string[]
}

/**
 * 撤销按钮权限的 API 路由请求参数
 */
export interface RevokeApiRouteRequest {
  btnPermId: string
  id: string
}

/**
 * 分配按钮权限到角色请求参数
 */
export interface AssignBtnPermToRoleRequest {
  roleId: string
  btnPermIds: string[]
}

/**
 * 从角色撤销按钮权限请求参数
 */
export interface RevokeBtnPermFromRoleRequest {
  roleId: string
  btnPermId: string
}

/**
 * 角色关联的按钮权限信息
 */
export interface BtnPermForRole {
  id: string
  menuId: string
  code: string
  name: string
  description: string
  status: number
  isAssigned: boolean
}

// ==================== 按钮权限 CRUD 接口 ====================

/**
 * 创建按钮权限
 * @param data 按钮权限信息
 * @returns 创建结果
 */
export const createBtnPerm = (data: CreateBtnPermRequest): Promise<ApiResponse> => {
  return service({
    url: '/btn-perms',
    method: 'post',
    data,
  })
}

/**
 * 获取按钮权限详情
 * @param id 按钮权限 ID
 * @returns 按钮权限详情
 */
export const getBtnPermDetail = (id: string): Promise<ApiResponse<BtnPermDetail>> => {
  return service({
    url: '/btn-perms/detail',
    method: 'get',
    params: { id },
  })
}

/**
 * 更新按钮权限
 * @param data 按钮权限信息
 * @returns 更新结果
 */
export const updateBtnPerm = (data: UpdateBtnPermRequest): Promise<ApiResponse> => {
  return service({
    url: '/btn-perms',
    method: 'put',
    data,
  })
}

/**
 * 删除按钮权限
 * @param id 按钮权限 ID
 * @returns 删除结果
 */
export const deleteBtnPerm = (id: number): Promise<ApiResponse> => {
  return service({
    url: '/btn-perms',
    method: 'delete',
    data: { id },
  })
}

/**
 * 获取按钮权限列表（分页）
 * @param params 查询参数
 * @returns 按钮权限列表
 */
export const getBtnPermList = (
  params: GetBtnPermListParams,
): Promise<ApiResponse<BtnPermListResponse>> => {
  return service({
    url: '/btn-perms',
    method: 'get',
    params,
  })
}

// ==================== 按钮权限与 API 路由关联接口 ====================

/**
 * 分配 API 路由到按钮权限
 * @param data 请求参数
 * @returns 分配结果
 */
export const assignApiRoutesToBtnPerm = (data: AssignApiRoutesRequest): Promise<ApiResponse> => {
  return service({
    url: '/btn-perms/assign-api-routes',
    method: 'post',
    data,
  })
}

/**
 * 撤销按钮权限的 API 路由
 * @param data 请求参数
 * @returns 撤销结果
 */
export const revokeApiRouteFromBtnPerm = (data: RevokeApiRouteRequest): Promise<ApiResponse> => {
  return service({
    url: '/btn-perms/revoke-api-route',
    method: 'post',
    data,
  })
}

/**
 * 获取按钮权限关联的 API 路由
 * @param btnPermId 按钮权限 ID
 * @returns API 路由列表
 */
export const getApiRoutesByBtnPerm = (btnPermId: number): Promise<ApiResponse<ApiRoute[]>> => {
  return service({
    url: '/btn-perms/api-routes',
    method: 'get',
    params: { btnPermId },
  })
}

/**
 * 获取 API 路由关联的按钮权限
 * @param routeId API 路由 ID
 * @returns 按钮权限列表
 */
export const getBtnPermsByApiRoute = (routeId: number): Promise<ApiResponse<BtnPerm[]>> => {
  return service({
    url: '/btn-perms/for-route',
    method: 'get',
    params: { routeId },
  })
}

// ==================== 按钮权限与角色关联接口 ====================

/**
 * 分配按钮权限到角色
 * @param data 请求参数
 * @returns 分配结果
 */
export const assignBtnPermToRole = (data: AssignBtnPermToRoleRequest): Promise<ApiResponse> => {
  return service({
    url: '/btn-perms/assign-to-role',
    method: 'post',
    data,
  })
}

/**
 * 从角色撤销按钮权限
 * @param data 请求参数
 * @returns 撤销结果
 */
export const revokeBtnPermFromRole = (data: RevokeBtnPermFromRoleRequest): Promise<ApiResponse> => {
  return service({
    url: '/btn-perms/revoke-from-role',
    method: 'post',
    data,
  })
}

/**
 * 获取角色的按钮权限
 * @param roleId 角色 ID
 * @returns 按钮权限列表
 */
export const getBtnPermsByRole = (roleId: number): Promise<ApiResponse<BtnPermForRole[]>> => {
  return service({
    url: '/btn-perms/by-role',
    method: 'get',
    params: { roleId },
  })
}

// ==================== 按钮权限与菜单关联接口 ====================

/**
 * 获取菜单的按钮权限
 * @param menuId 菜单 ID
 * @returns 按钮权限列表
 */
export const getBtnPermsByMenu = (menuId: number): Promise<ApiResponse<BtnPerm[]>> => {
  return service({
    url: '/btn-perms/by-menu',
    method: 'get',
    params: { menuId },
  })
}
