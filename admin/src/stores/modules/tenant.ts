import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getGrantedTenants } from '@/api/modules/tenant'
import type { Tenant } from '@/api/modules/tenant'

// 租户切换器状态（管理端单租户切换查询模式，见设计文档 01 §3.4）
// currentTenantId 全局共享并持久化：对账/渠道配置等页面请求自动携带 ?tenant_id=
const STORAGE_KEY = 'ixpay_current_tenant_id'

export const useTenantStore = defineStore('tenant', () => {
  // 当前选中租户（0/空 = 默认租户，查主库存量数据）
  const currentTenantId = ref<string>(localStorage.getItem(STORAGE_KEY) || '')
  // 租户列表（切换器下拉数据源：admin 全量/普通用户按授权）
  const tenants = ref<Tenant[]>([])
  const loaded = ref(false)

  const setCurrentTenant = (tenantId: string) => {
    currentTenantId.value = tenantId || ''
    if (tenantId) {
      localStorage.setItem(STORAGE_KEY, tenantId)
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  }

  // 加载当前用户可见租户（granted 接口权限要求低，普通用户仅返回被授权租户）
  const loadTenants = async (force = false) => {
    if (loaded.value && !force) return
    try {
      const res = await getGrantedTenants()
      tenants.value = (res.data || []) as unknown as Tenant[]
      // 当前选中租户不在可见列表时重置（如授权被取消）
      if (currentTenantId.value && !tenants.value.some(t => t.id === currentTenantId.value)) {
        setCurrentTenant('')
      }
    } catch {
      // 获取失败保持现状（如网络抖动），不阻断页面使用
    }
    loaded.value = true
  }

  // 当前请求附加的 tenant_id 查询参数（未选择租户时不传 = 主库）
  const tenantIdParam = () => (currentTenantId.value ? { tenant_id: currentTenantId.value } : {})

  return { currentTenantId, tenants, loaded, setCurrentTenant, loadTenants, tenantIdParam }
})
