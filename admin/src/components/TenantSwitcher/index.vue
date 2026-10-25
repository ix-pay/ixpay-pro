<template>
  <el-select
    :model-value="tenant.currentTenantId"
    placeholder="默认租户(主库)"
    clearable
    filterable
    style="width: 220px"
    @change="onChange"
    @visible-change="(visible: boolean) => visible && tenant.loadTenants()"
  >
    <el-option v-for="t in tenant.tenants" :key="t.id" :label="`${t.tenantName}(${t.tenantCode})`" :value="t.id" />
  </el-select>
</template>

<script setup lang="ts">
// 租户切换器（管理端单租户切换查询，见设计文档 01 §3.4）
// 选中后全局持久化，对账/渠道配置等页面请求自动携带 ?tenant_id=；清空 = 默认租户(主库)
import { useTenantStore } from '@/stores/modules/tenant'

const tenant = useTenantStore()

const onChange = (val: string) => {
  tenant.setCurrentTenant(val)
}
</script>
