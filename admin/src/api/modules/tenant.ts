import service from '@/utils/request'
import type { ApiResponse } from '@/types'

// 多租户管理 API（租户生命周期/应用路由/授权/审计，设计文档 01 §4）
// 数据传输规范：JSON 字段 camelCase，雪花 ID 为字符串（后端 json:",string" 序列化）

export interface Tenant {
  id: string
  tenantCode: string
  tenantName: string
  dbName: string
  dbHost: string
  dbPort: number
  dbUser: string
  hasDbPass: boolean
  status: 'init' | 'active' | 'disabled' | 'archived' | 'failed'
  initError: string
  description: string
  createdAt: string
  updatedAt: string
}

export interface TenantRoute {
  id: string
  appKey: string
  tenantId: string
  moduleCode: string
  createdAt: string
}

export interface TenantGrant {
  id: string
  userId: string
  tenantId: string
  createdAt: string
}

export interface TenantOpLog {
  id: string
  tenantId: string
  action: string
  detail: string
  operatorId: string
  createdAt: string
}

export interface GrantedTenant {
  id: string
  tenantCode: string
  tenantName: string
  dbName: string
  status: string
}

// ---------- 租户 ----------

// 当前用户可见租户（admin 全量/普通用户按授权；租户切换器数据源，权限要求低）
export const getGrantedTenants = () =>
  service({ url: '/tenant/granted', method: 'get' }) as Promise<ApiResponse<GrantedTenant[]>>

export const getTenantList = (params?: { page?: number; pageSize?: number; status?: string }) =>
  service({
    url: '/tenant',
    method: 'get',
    params,
  }) as Promise<ApiResponse<{ list: Tenant[]; total: number; page: number; pageSize: number }>>

export const getTenant = (id: string | number) =>
  service({ url: `/tenant/${id}`, method: 'get' }) as Promise<ApiResponse<Tenant>>

export const createTenant = (data: {
  tenantCode: string
  tenantName: string
  dbHost?: string
  dbPort?: number
  dbUser?: string
  dbPassword?: string
  description?: string
}) => service({ url: '/tenant', method: 'post', data }) as Promise<ApiResponse<Tenant>>

export const updateTenant = (
  id: string | number,
  data: {
    tenantName?: string
    dbHost?: string
    dbPort?: number
    dbUser?: string
    dbPassword?: string
    description?: string
  },
) => service({ url: `/tenant/${id}`, method: 'put', data }) as Promise<ApiResponse<Tenant>>

export const retryTenantInit = (id: string | number) =>
  service({ url: `/tenant/${id}/retry`, method: 'post' }) as Promise<ApiResponse>

export const disableTenant = (id: string | number) =>
  service({ url: `/tenant/${id}/disable`, method: 'post' }) as Promise<ApiResponse>

export const enableTenant = (id: string | number) =>
  service({ url: `/tenant/${id}/enable`, method: 'post' }) as Promise<ApiResponse>

export const archiveTenant = (id: string | number) =>
  service({ url: `/tenant/${id}/archive`, method: 'post' }) as Promise<ApiResponse>

// ---------- 应用路由（app_key → 租户） ----------

export const getTenantRoutes = (tenantId: string | number) =>
  service({ url: `/tenant/${tenantId}/routes`, method: 'get' }) as Promise<ApiResponse<TenantRoute[]>>

export const createTenantRoute = (tenantId: string | number, data: { appKey: string; moduleCode?: string }) =>
  service({ url: `/tenant/${tenantId}/routes`, method: 'post', data }) as Promise<ApiResponse<TenantRoute>>

export const deleteTenantRoute = (routeId: string | number) =>
  service({ url: `/tenant/routes/${routeId}`, method: 'delete' }) as Promise<ApiResponse>

// ---------- 用户授权 ----------

export const getTenantGrants = (tenantId: string | number) =>
  service({ url: `/tenant/${tenantId}/grants`, method: 'get' }) as Promise<ApiResponse<TenantGrant[]>>

export const createTenantGrant = (tenantId: string | number, data: { userId: string | number }) =>
  service({ url: `/tenant/${tenantId}/grants`, method: 'post', data }) as Promise<ApiResponse<TenantGrant>>

export const deleteTenantGrant = (tenantId: string | number, userId: string | number) =>
  service({ url: `/tenant/${tenantId}/grants/${userId}`, method: 'delete' }) as Promise<ApiResponse>

// ---------- 运维审计 ----------

export const getTenantOpLogs = (tenantId: string | number, params?: { limit?: number }) =>
  service({
    url: `/tenant/${tenantId}/ops-logs`,
    method: 'get',
    params,
  }) as Promise<ApiResponse<TenantOpLog[]>>
