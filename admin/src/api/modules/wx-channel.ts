import service from '@/utils/request'
import type { ApiResponse } from '@/types'
import { useTenantStore } from '@/stores/modules/tenant'

// 微信多渠道配置 API（应用/商户/关联/对账）
// 说明：路由已迁移至 base 管理端 /api/admin/wx/*，鉴权统一由 base 负责，
// 直接使用主请求实例（baseURL = /api/admin）

// 渠道配置数据按租户库隔离：所有接口统一自动携带当前选中租户（?tenant_id=），
// 未选中租户时查/写主库。写操作此前仅列表查询带租户参数，导致切换租户后
// 新增应用/商户落库到主库或报错——现收敛到本 helper 统一注入，杜绝遗漏
function tenantParams(extra?: Record<string, unknown>): Record<string, unknown> {
  const tenant = useTenantStore()
  return { ...tenant.tenantIdParam(), ...extra }
}

export interface WxAppConfig {
  id: string
  appKey: string
  appType: 'mp' | 'miniapp'
  appId: string
  appSecret: string
  /** 默认商户（只读，服务端按商户关联表 is_default 标记组装返回；0=未设置） */
  defaultMerchantId: string
  /** 默认商户号（展示用，空=未设置） */
  defaultMerchantMchId?: string
  /** 默认商户名称（展示用，空=未设置） */
  defaultMerchantName?: string
  authScope: string
  oauthCallbackUrl: string
  frontendUrl: string
  description: string
  status: number
  authUrl?: string
  createdAt?: string
}

export interface WxMerchantConfig {
  id: string
  mchId: string
  mchName: string
  mchSerialNo: string
  apiv3Key: string
  privateKey: string
  publicKeyId: string
  publicKey: string
  payNotifyUrl: string
  refundNotifyUrl: string
  status: number
  createdAt?: string
}

export interface WxAppMerchant {
  id: string
  merchantId?: string
  appConfigId: string
  merchantConfigId: string
  mchId?: string
  mchName?: string
  isDefault: boolean
}

export interface WxReconSummary {
  totalCount: number
  successCount: number
  successAmount: number
  refundCount: number
  refundAmount: number
  pendingCount: number
}

// ---------- 应用配置 ----------

export const getWxAppList = (params?: { page?: number; pageSize?: number; tenant_id?: string }) =>
  service({
    url: '/wx/apps',
    method: 'get',
    params: tenantParams(params),
  }) as Promise<ApiResponse<{ list: WxAppConfig[]; total: number; page: number; pageSize: number }>>

export const createWxApp = (data: Partial<WxAppConfig>) =>
  service({
    url: '/wx/apps',
    method: 'post',
    data,
    params: tenantParams(),
  }) as Promise<ApiResponse<WxAppConfig>>

export const updateWxApp = (id: string, data: Partial<WxAppConfig>) =>
  service({
    url: `/wx/apps/${id}`,
    method: 'put',
    data,
    params: tenantParams(),
  }) as Promise<ApiResponse<WxAppConfig>>

export const deleteWxApp = (id: string) =>
  service({
    url: `/wx/apps/${id}`,
    method: 'delete',
    params: tenantParams(),
  }) as Promise<ApiResponse>

export const getWxAppAuthURL = (id: string, redirectUri?: string) =>
  service({
    url: `/wx/apps/${id}/auth-url`,
    method: 'get',
    params: tenantParams({ redirect_uri: redirectUri }),
  }) as Promise<ApiResponse<{ authUrl: string; appKey: string }>>

// 小程序入口码：返回 PNG 二进制 Blob（小程序类型应用专属；响应为图片流，经 request 封装后 data 字段即 Blob）
export const getWxAppMiniappQrcode = async (id: string): Promise<Blob> => {
  const res = (await service({
    url: `/wx/apps/${id}/miniapp-qrcode`,
    method: 'get',
    params: tenantParams(),
    responseType: 'blob',
    donNotShowLoading: false,
  })) as unknown as ApiResponse<Blob>
  return res.data as Blob
}

// 服务号授权入口二维码：返回 PNG 二进制 Blob（服务号类型应用专属）
export const getWxAppAuthQrcode = async (id: string): Promise<Blob> => {
  const res = (await service({
    url: `/wx/apps/${id}/auth-qrcode`,
    method: 'get',
    params: tenantParams(),
    responseType: 'blob',
    donNotShowLoading: false,
  })) as unknown as ApiResponse<Blob>
  return res.data as Blob
}

// ---------- 商户配置 ----------

export const getWxMerchantList = (params?: { page?: number; pageSize?: number }) =>
  service({
    url: '/wx/merchants',
    method: 'get',
    params: tenantParams(params),
  }) as Promise<
    ApiResponse<{ list: WxMerchantConfig[]; total: number; page: number; pageSize: number }>
  >

export const createWxMerchant = (data: Partial<WxMerchantConfig>) =>
  service({
    url: '/wx/merchants',
    method: 'post',
    data,
    params: tenantParams(),
  }) as Promise<ApiResponse<WxMerchantConfig>>

export const updateWxMerchant = (id: string, data: Partial<WxMerchantConfig>) =>
  service({
    url: `/wx/merchants/${id}`,
    method: 'put',
    data,
    params: tenantParams(),
  }) as Promise<ApiResponse<WxMerchantConfig>>

export const deleteWxMerchant = (id: string) =>
  service({
    url: `/wx/merchants/${id}`,
    method: 'delete',
    params: tenantParams(),
  }) as Promise<ApiResponse>

// ---------- 应用-商户关联 ----------

export const getWxAppMerchants = (appId: string) =>
  service({
    url: `/wx/apps/${appId}/merchants`,
    method: 'get',
    params: tenantParams(),
  }) as Promise<ApiResponse<{ list: WxAppMerchant[]; total: number }>>

export const bindWxAppMerchant = (
  appId: string,
  data: { merchantId: string; isDefault: boolean },
) =>
  service({
    url: `/wx/apps/${appId}/merchants`,
    method: 'post',
    data,
    params: tenantParams(),
  }) as Promise<ApiResponse>

export const unbindWxAppMerchant = (appId: string, merchantId: string) =>
  service({
    url: `/wx/apps/${appId}/merchants/${merchantId}`,
    method: 'delete',
    params: tenantParams(),
  }) as Promise<ApiResponse>

// ---------- 租户 UI 配置（菜单/图片/主题） ----------

export const getWxAppUI = (appId: string) =>
  service({
    url: `/wx/apps/${appId}/ui`,
    method: 'get',
    params: tenantParams(),
  }) as Promise<ApiResponse<{ uiJson: string }>>

export const saveWxAppUI = (appId: string, data: { uiJson: string }) =>
  service({
    url: `/wx/apps/${appId}/ui`,
    method: 'put',
    data,
    params: tenantParams(),
  }) as Promise<ApiResponse>

// ---------- 对账 ----------

export const getWxReconSummary = (params?: { startDate?: string; endDate?: string }) =>
  service({
    url: '/wx/recon/summary',
    method: 'get',
    params: tenantParams(params),
  }) as Promise<ApiResponse<WxReconSummary>>

export const getWxReconPayments = (params?: {
  startDate?: string
  endDate?: string
  status?: string
  page?: number
  pageSize?: number
}) =>
  service({
    url: '/wx/recon/payments',
    method: 'get',
    params: tenantParams(params),
  }) as Promise<ApiResponse<{ list: unknown[]; total: number; page: number; pageSize: number }>>
