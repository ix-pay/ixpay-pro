<template>
  <div class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md">
    <!-- 顶部操作栏（两行：筛选 / 功能按钮） -->
    <div class="flex flex-col gap-3 p-4 border-b">
      <div class="flex flex-wrap items-center gap-3">
        <el-select
          v-model="statusFilter"
          placeholder="全部状态"
          clearable
          style="width: 160px"
          @change="onStatusFilterChange"
        >
          <el-option label="初始化中" value="init" />
          <el-option label="正常" value="active" />
          <el-option label="已禁用" value="disabled" />
          <el-option label="已归档" value="archived" />
          <el-option label="初始化失败" value="failed" />
        </el-select>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <el-button v-if="canCreate" type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新增租户
        </el-button>
        <el-button @click="load">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <span class="text-xs text-gray-500 dark:text-gray-400">
          接入流程：创建租户 → 配置微信应用（自动绑定路由）→ 授权用户；点「详情」可查看进度与下一步操作
        </span>
      </div>
    </div>

    <!-- 统计信息（基于当前页数据） -->
    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon color="var(--primary-color)"><OfficeBuilding /></el-icon>
          租户总数：<span class="font-medium">{{ total }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon color="var(--success-color)"><CircleCheck /></el-icon>
          正常：<span class="font-medium">{{ statusCounts.active }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon color="var(--warning-color)"><Loading /></el-icon>
          初始化中：<span class="font-medium">{{ statusCounts.init }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon color="var(--danger-color)"><Warning /></el-icon>
          失败：<span class="font-medium">{{ statusCounts.failed }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon color="var(--text-secondary)"><CircleClose /></el-icon>
          禁用/归档：<span class="font-medium">{{ statusCounts.offline }}</span>
        </span>
      </div>
    </div>

    <!-- 表格区域（占满剩余空间） -->
    <div class="flex-1 overflow-hidden">
      <el-table :data="tenants" :height="'100%'" stripe v-loading="loading">
        <el-table-column type="index" label="序号" width="70" />
        <el-table-column prop="tenantCode" label="租户编码" width="140" />
        <el-table-column prop="tenantName" label="租户名称" width="160" />
        <el-table-column prop="dbName" label="业务库名" min-width="150">
          <template #default="{ row }">
            <span>{{ row.dbName || '主库(默认租户)' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="独立实例" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.dbHost" type="warning">独立实例</el-tag>
            <el-tag v-else type="info">共用主实例</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="130">
          <template #default="{ row }">
            <el-tooltip
              :disabled="!row.initError"
              :content="row.initError"
              placement="top"
              max-width="360"
            >
              <span class="flex items-center gap-1">
                <el-icon
                  v-if="row.status === 'init'"
                  class="is-loading"
                  color="var(--el-color-warning)"
                >
                  <Loading />
                </el-icon>
                <el-tag :type="statusTag(row.status)">{{ statusText(row.status) }}</el-tag>
              </span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="130" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-2">
              <el-button size="small" type="primary" @click="openDetail(row)">详情</el-button>
              <el-button
                v-if="canLifecycle && (row.status === 'failed' || row.status === 'active')"
                size="small"
                type="warning"
                @click="doRetry(row)"
              >
                重试初始化
              </el-button>
              <el-button
                v-if="canLifecycle && row.status === 'active'"
                size="small"
                type="danger"
                @click="doDisable(row)"
              >
                禁用
              </el-button>
              <el-button
                v-if="canLifecycle && row.status === 'disabled'"
                size="small"
                type="success"
                @click="doEnable(row)"
              >
                启用
              </el-button>
              <el-button
                v-if="canLifecycle && row.status === 'active'"
                size="small"
                @click="doArchive(row)"
              >
                归档
              </el-button>
              <el-button v-if="canEdit && row.id !== '1'" size="small" @click="openEditDialog(row)">
                编辑
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页区域 -->
    <div class="flex items-center justify-between px-4 py-3 border-t">
      <span class="text-sm text-gray-600 dark:text-gray-400">共 {{ total }} 条</span>
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="load"
      />
    </div>

    <!-- 创建/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑租户' : '新增租户'" width="720px">
      <el-alert
        v-if="!form.id"
        title="只需填写编码与名称；建库/迁移/初始化自动执行（异步任务），失败可一键重试。独立实例连接信息可选。"
        type="info"
        :closable="false"
        class="mb-4"
      />
      <el-form :model="form" label-width="130px">
        <el-divider content-position="left">基础信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="租户编码" required>
              <el-input
                v-model="form.tenantCode"
                placeholder="小写字母开头，3-31位字母数字下划线"
                :disabled="!!form.id"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="租户名称" required>
              <el-input v-model="form.tenantName" placeholder="如 子公司A" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="form.description" placeholder="仅用于平台内备注识别" />
        </el-form-item>

        <el-divider content-position="left">独立实例连接（可选，留空沿用主实例）</el-divider>
        <el-row :gutter="16">
          <el-col :span="16">
            <el-form-item label="实例Host">
              <el-input v-model="form.dbHost" placeholder="如 192.168.1.100；留空沿用主实例" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="实例端口">
              <el-input-number v-model="form.dbPort" :min="0" :max="65535" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="实例用户">
              <el-input v-model="form.dbUser" placeholder="如 ixpay_tenant_a" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="实例密码">
              <el-input
                v-model="form.dbPassword"
                type="password"
                show-password
                :placeholder="form.hasDbPass ? '留空表示保留原密码' : 'AES-GCM 加密存储'"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label-width="0">
          <div
            class="-mt-3.5 text-xs leading-relaxed text-gray-500 dark:text-gray-400"
          >
            独立实例连接信息仅租户创建后用于建库；密码 AES-GCM 加密落库，编辑时留空表示保留原密码
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 详情抽屉：接入进度 + 应用接入 / 用户授权 / 操作审计 -->
    <el-drawer
      v-model="detailVisible"
      :title="`租户详情 - ${detail?.tenantName || ''}`"
      size="720px"
    >
      <template v-if="detail">
        <div class="flex flex-col gap-4">
          <!-- 头部摘要卡：核心信息一屏可读 -->
          <div class="rounded-lg bg-gray-50 dark:bg-gray-800 p-4">
            <div class="flex flex-wrap items-center gap-2">
              <el-icon :size="20" color="var(--primary-color)"><OfficeBuilding /></el-icon>
              <span class="text-base font-medium">{{ detail.tenantName }}</span>
              <el-tag :type="statusTag(detail.status)" size="small">{{
                statusText(detail.status)
              }}</el-tag>
              <el-tag v-if="detail.dbHost" type="warning" size="small">独立实例</el-tag>
            </div>
            <div class="mt-2 grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
              <div class="flex items-center gap-1">
                <span class="text-gray-500 dark:text-gray-400">编码：</span>
                <span class="font-mono">{{ detail.tenantCode }}</span>
              </div>
              <div class="flex items-center gap-1">
                <span class="text-gray-500 dark:text-gray-400">业务库：</span>
                <span class="font-mono">{{ detail.dbName || '主库(默认租户)' }}</span>
              </div>
              <div class="col-span-2 flex items-center gap-1">
                <span class="text-gray-500 dark:text-gray-400">描述：</span>
                <span>{{ detail.description || '—' }}</span>
              </div>
              <div v-if="detail.initError" class="col-span-2 flex items-center gap-1">
                <span class="text-gray-500 dark:text-gray-400">初始化错误：</span>
                <span class="text-red-500">{{ detail.initError }}</span>
              </div>
            </div>
          </div>

          <!-- 接入流程步骤条：标示当前进度（初始化 → 配置微信应用 → 绑定路由 → 授权用户） -->
          <div class="rounded-lg border border-gray-200 dark:border-gray-700 p-4">
            <el-steps :active="stepActive" align-center>
              <el-step
                title="租户初始化"
                description="自动建库/迁移"
                :status="detail.status === 'failed' ? 'error' : undefined"
              />
              <el-step title="配置微信应用" description="渠道配置中新增（自动绑路由）" />
              <el-step title="应用路由" description="自动绑定，可手动修正" />
              <el-step title="授权用户" description="可选" />
            </el-steps>
            <!-- 下一步操作引导：按当前进度提示在哪一步、做什么 -->
            <el-alert
              v-if="nextAction"
              :title="nextAction.text"
              :type="nextAction.type"
              :closable="false"
              class="mt-3"
            />
          </div>

          <el-tabs v-model="detailTab">
            <!-- 应用接入：微信应用（渠道配置）+ 应用路由绑定 一体化操作 -->
            <el-tab-pane label="应用接入" name="routes">
              <div class="flex flex-col gap-3">
                <el-alert
                  title="应用路由决定前端实例归属：wxapp 运行时凭 app_key 经主库路由表定位本租户（登录/支付均按此路由）。在渠道配置中新增应用时会自动绑定，此处仅用于手动修正/补绑"
                  type="info"
                  :closable="false"
                />

                <!-- 步骤 2：该租户已配置的微信应用（来自渠道配置，按租户库查询） -->
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <span class="flex items-center gap-1 text-sm font-medium">
                    <el-icon color="var(--primary-color)"><Connection /></el-icon>
                    微信应用（步骤 2 · 在渠道配置中维护）
                  </span>
                  <el-button size="small" type="primary" @click="goChannel">
                    前往渠道配置（服务号 / 商户号）→
                  </el-button>
                </div>
                <el-alert
                  v-if="tenantAppsError"
                  :title="`微信应用读取失败（租户初始化未完成或库不可达）：${tenantAppsError}`"
                  type="warning"
                  :closable="false"
                />
                <template v-else>
                  <el-table :data="tenantApps" size="small" v-loading="tenantAppsLoading">
                    <el-table-column prop="appKey" label="实例标志(app_key)" min-width="140" />
                    <el-table-column prop="appType" label="类型" width="90">
                      <template #default="{ row }">
                        <el-tag :type="row.appType === 'mp' ? 'success' : 'warning'" size="small">
                          {{ row.appType === 'mp' ? '服务号' : '小程序' }}
                        </el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column
                      prop="appId"
                      label="AppID"
                      min-width="160"
                      show-overflow-tooltip
                    />
                    <el-table-column label="路由状态" width="100">
                      <template #default="{ row }">
                        <el-tag v-if="isRouteBound(row.appKey)" type="success" size="small"
                          >已绑定</el-tag
                        >
                        <el-tag v-else type="info" size="small">未绑定</el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column label="操作" width="110">
                      <template #default="{ row }">
                        <el-button
                          v-if="canManage && !isRouteBound(row.appKey)"
                          size="small"
                          type="primary"
                          @click="bindRoute(row.appKey)"
                        >
                          绑定路由
                        </el-button>
                        <span v-else class="text-xs text-gray-400">—</span>
                      </template>
                    </el-table-column>
                  </el-table>
                  <el-empty
                    v-if="!tenantAppsLoading && tenantApps.length === 0"
                    description="该租户暂未配置微信应用，请先「前往渠道配置」新增应用"
                    :image-size="72"
                  />
                </template>

                <!-- 步骤 3：应用路由（创建应用时自动绑定；仅存在未绑定项时展示修正操作） -->
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <span class="flex items-center gap-1 text-sm font-medium">
                    <el-icon color="var(--success-color)"><Link /></el-icon>
                    应用路由（创建应用时自动绑定）
                  </span>
                  <el-tag v-if="!showRouteBinding" type="success" size="small">
                    全部应用已绑定
                  </el-tag>
                </div>
                <!-- 手动修正/补绑（仅有未绑定应用或路由为空时展示，正常流程无需操作） -->
                <div v-if="showRouteBinding" class="flex flex-wrap items-center gap-2">
                  <el-select
                    v-model="newRouteKey"
                    filterable
                    allow-create
                    default-first-option
                    clearable
                    placeholder="选择未绑定的应用，或手动输入 app_key"
                    style="width: 300px"
                    :disabled="!canManage"
                  >
                    <el-option
                      v-for="a in tenantApps"
                      :key="a.id"
                      :label="`${a.appKey}（${a.appType === 'mp' ? '服务号' : '小程序'}）`"
                      :value="a.appKey"
                    />
                  </el-select>
                  <el-button v-if="canManage" type="primary" @click="addRoute">
                    补绑/手动绑定
                  </el-button>
                  <span class="text-xs text-gray-500 dark:text-gray-400">
                    正常流程无需此操作；仅用于旧数据补绑或前端实例先于应用配置创建的场景
                  </span>
                </div>

                <!-- 已绑定路由列表 -->
                <el-table :data="routes" border size="small">
                  <el-table-column prop="appKey" label="app_key" min-width="140" />
                  <el-table-column prop="moduleCode" label="模块" width="80" />
                  <el-table-column prop="createdAt" label="绑定时间" width="160" />
                  <el-table-column label="操作" width="80">
                    <template #default="{ row }">
                      <el-button
                        v-if="canGrant"
                        size="small"
                        type="danger"
                        @click="removeRoute(row)"
                      >
                        解绑
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-tab-pane>

            <!-- 用户授权 -->
            <el-tab-pane label="用户授权" name="grants">
              <div class="flex flex-col gap-3">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <span class="flex items-center gap-1 text-sm font-medium">
                    <el-icon color="var(--warning-color)"><User /></el-icon>
                    授权用户（步骤 4 · 可选）
                  </span>
                  <div class="flex flex-wrap items-center gap-2">
                    <el-select
                      v-model="newGrantUserID"
                      filterable
                      remote
                      clearable
                      :remote-method="searchUsers"
                      :loading="userSearching"
                      placeholder="搜索用户名添加授权"
                      style="width: 240px"
                      :disabled="!canManage"
                    >
                      <el-option
                        v-for="u in userOptions"
                        :key="u.id"
                        :label="`${u.nickname || u.userName}(${u.userName})`"
                        :value="u.id"
                      />
                    </el-select>
                    <el-button v-if="canManage" type="primary" @click="addGrant">
                      添加授权
                    </el-button>
                  </div>
                </div>
                <el-alert
                  title="admin 角色天然可查全部租户；此处授权普通用户查看本租户数据（对账/渠道/支付单，经租户切换器）"
                  type="info"
                  :closable="false"
                />
                <el-table :data="grants" border size="small">
                  <el-table-column label="用户" min-width="160">
                    <template #default="{ row }">{{ userLabel(row.userId) }}</template>
                  </el-table-column>
                  <el-table-column prop="userId" label="用户ID" width="150" show-overflow-tooltip />
                  <el-table-column prop="createdAt" label="授权时间" width="160" />
                  <el-table-column label="操作" width="80">
                    <template #default="{ row }">
                      <el-button
                        v-if="canGrant"
                        size="small"
                        type="danger"
                        @click="removeGrant(row)"
                      >
                        取消
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-tab-pane>

            <!-- 操作审计 -->
            <el-tab-pane label="操作审计" name="logs">
              <el-table :data="opLogs" border size="small">
                <el-table-column prop="action" label="动作" width="110" />
                <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
                <el-table-column label="操作人" width="150" show-overflow-tooltip>
                  <template #default="{ row }">{{ userLabel(row.operatorId) }}</template>
                </el-table-column>
                <el-table-column prop="createdAt" label="时间" width="160" />
              </el-table>
            </el-tab-pane>
          </el-tabs>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import {
  Loading,
  Plus,
  Refresh,
  OfficeBuilding,
  CircleCheck,
  CircleClose,
  Warning,
  Connection,
  Link,
  User,
} from '@element-plus/icons-vue'
import {
  getTenantList,
  createTenant,
  updateTenant,
  retryTenantInit,
  disableTenant,
  enableTenant,
  archiveTenant,
  getTenantRoutes,
  createTenantRoute,
  deleteTenantRoute,
  getTenantGrants,
  createTenantGrant,
  deleteTenantGrant,
  getTenantOpLogs,
  type Tenant,
  type TenantRoute,
  type TenantGrant,
  type TenantOpLog,
} from '@/api/modules/tenant'
import { getUserList } from '@/api/modules/user'
import { getWxAppList, type WxAppConfig } from '@/api/modules/wx-channel'
import { hasAnyButtonPermission, hasRole } from '@/utils/permission'

const router = useRouter()

// 前往渠道配置页：携带 tenant_id 跳转，渠道页自动选中本租户（免二次切换），
// 渠道页的应用/商户数据即读写该租户库
function goChannel() {
  router.push({
    path: '/wx/channel',
    query: detail.value?.id ? { tenant_id: detail.value.id } : {},
  })
}

// 列表
const tenants = ref<Tenant[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const statusFilter = ref('')

// 新增入口：admin 角色直通（角色来自登录信息，不依赖菜单按钮种子），普通用户按按钮权限
const canCreate = computed(() => hasRole('admin') || hasAnyButtonPermission(['tenant:tenant:add']))

// 各操作入口（admin 直通，普通用户按对应按钮权限；后端仍有 isAdmin 强校验兜底）
const canManage = computed(
  () =>
    hasRole('admin') ||
    hasAnyButtonPermission([
      'tenant:tenant:add',
      'tenant:tenant:update',
      'tenant:tenant:lifecycle',
      'tenant:tenant:grant',
    ]),
)
const canLifecycle = computed(
  () => hasRole('admin') || hasAnyButtonPermission(['tenant:tenant:lifecycle']),
)
const canEdit = computed(() => hasRole('admin') || hasAnyButtonPermission(['tenant:tenant:update']))
const canGrant = computed(() => hasRole('admin') || hasAnyButtonPermission(['tenant:tenant:grant']))

// 统计信息（基于当前页数据；offline 含禁用与归档）
const statusCounts = computed(() => {
  const counts = { active: 0, init: 0, failed: 0, offline: 0 }
  for (const t of tenants.value) {
    if (t.status === 'active') counts.active++
    else if (t.status === 'init') counts.init++
    else if (t.status === 'failed') counts.failed++
    else counts.offline++
  }
  return counts
})

// 状态筛选变更：重置到第一页再查询
function onStatusFilterChange() {
  page.value = 1
  load()
}

// 弹窗表单（独立实例连接信息可选；dbPassword 明文输入，服务端加密落库）
interface TenantForm {
  id?: string
  tenantCode: string
  tenantName: string
  dbHost: string
  dbPort: number
  dbUser: string
  dbPassword: string
  description: string
  hasDbPass?: boolean
}
const dialogVisible = ref(false)
const form = ref<TenantForm>({
  tenantCode: '',
  tenantName: '',
  dbHost: '',
  dbPort: 0,
  dbUser: '',
  dbPassword: '',
  description: '',
})

// 详情
const detailVisible = ref(false)
const detail = ref<Tenant | null>(null)
const detailTab = ref('routes')
const routes = ref<TenantRoute[]>([])
const grants = ref<TenantGrant[]>([])
const opLogs = ref<TenantOpLog[]>([])
const newRouteKey = ref('')

// 该租户已配置的微信应用（渠道配置数据，按租户库查询；应用接入 tab 展示与一键绑定路由）
const tenantApps = ref<WxAppConfig[]>([])
const tenantAppsLoading = ref(false)
const tenantAppsError = ref('')

async function loadTenantApps() {
  if (!detail.value) return
  tenantAppsLoading.value = true
  tenantAppsError.value = ''
  try {
    const res = await getWxAppList({ page: 1, pageSize: 100, tenant_id: detail.value.id })
    tenantApps.value = res.data!.list
  } catch (e) {
    // 租户初始化未完成/库不可达时读取失败，降级为提示而非阻断详情页
    tenantApps.value = []
    tenantAppsError.value = (e as Error).message || '未知错误'
  } finally {
    tenantAppsLoading.value = false
  }
}

// app_key 是否已绑定到本租户路由
function isRouteBound(appKey: string): boolean {
  return routes.value.some((r) => r.appKey === appKey)
}

// 是否展示手动绑定/补绑操作：路由为空（允许先绑前端实例标志兜底）或存在未绑定应用；
// 全部绑定时隐藏操作区（正常流程创建应用已自动绑定，无需手动操作）
const showRouteBinding = computed(
  () =>
    routes.value.length === 0 || tenantApps.value.some((a) => !isRouteBound(a.appKey)),
)

// 从已配置的微信应用一键绑定路由（正常由创建应用自动完成，此处为补绑入口）
async function bindRoute(appKey: string) {
  if (!detail.value) return
  await createTenantRoute(detail.value.id, { appKey })
  ElMessage.success(`应用 ${appKey} 已绑定到本租户`)
  await loadRoutes()
}

// 接入流程进度（步骤条 active：0=初始化中 1=应用未配置 2=路由未绑定 3=未授权 4=全部完成）
const stepActive = computed(() => {
  if (!detail.value) return 0
  const s = detail.value.status
  if (s !== 'active' && s !== 'disabled' && s !== 'archived') return 0
  let done = 1
  if (tenantApps.value.length > 0) done = 2
  if (done >= 2 && routes.value.length > 0) done = 3
  if (done >= 3 && grants.value.length > 0) done = 4
  return done
})

// 下一步操作引导：按当前进度提示在哪一步、修改/操作什么
const nextAction = computed<{
  text: string
  type: 'info' | 'warning' | 'error' | 'success'
} | null>(() => {
  if (!detail.value) return null
  const s = detail.value.status
  if (s === 'init') {
    return {
      text: '步骤 1：租户初始化中（自动建库/迁移），完成后列表状态变为「正常」；期间渠道配置暂不可用',
      type: 'warning',
    }
  }
  if (s === 'failed') {
    return {
      text: '步骤 1：租户初始化失败，请在列表点击「重试初始化」；失败原因见上方「初始化错误」',
      type: 'error',
    }
  }
  if (s === 'disabled') {
    return { text: '租户已禁用，请先在列表「启用」后再进行接入操作', type: 'warning' }
  }
  if (s === 'archived') {
    return { text: '租户已归档（只读保留），不可再进行接入操作', type: 'info' }
  }
  if (tenantApps.value.length === 0) {
    return {
      text: '步骤 2：该租户还没有微信应用。点击下方「前往渠道配置」新增应用（新增时已自动路由到本租户库）；若无商户号请先在该页「商户号配置」新增',
      type: 'info',
    }
  }
  if (routes.value.length === 0 || tenantApps.value.some((a) => !isRouteBound(a.appKey))) {
    return {
      text: '步骤 3：正常情况下新增应用时会自动绑定路由；此处显示未绑定多为旧数据或自动绑定失败，点击「绑定路由」一键补绑即可',
      type: 'info',
    }
  }
  if (grants.value.length === 0) {
    return {
      text: '步骤 4（可选）：接入已完成。如需普通用户查看本租户数据，请切换到「用户授权」标签添加授权',
      type: 'success',
    }
  }
  return {
    text: '接入流程已全部完成：租户正常、微信应用已配置并绑定路由、用户已授权',
    type: 'success',
  }
})

// 用户授权下拉选中值
const newGrantUserID = ref('')

// 授权用户远程搜索（替代手输用户ID，降低误操作）
// 字段与后端 UserResponse JSON 一致：userName/nickname（驼峰）
interface UserOption {
  id: string
  userName: string
  nickname: string
}
const userOptions = ref<UserOption[]>([])
const userSearching = ref(false)

// 用户ID → 展示名映射（授权列表/操作审计显示昵称/用户名，替代裸 ID；打开抽屉时批量拉取）
const userLabelMap = ref<Record<string, string>>({})

function userLabel(userId: string): string {
  return userLabelMap.value[userId] || userId
}

async function loadUserLabelMap() {
  try {
    // pageSize 上限为后端校验的 100
    const res = await getUserList({ page: 1, pageSize: 100 })
    const list = (res.data as unknown as { list?: UserOption[] })?.list || []
    const map: Record<string, string> = {}
    for (const u of list) {
      map[u.id] = u.nickname || u.userName
    }
    userLabelMap.value = map
  } catch {
    // 拉取失败退化为显示原始 ID，不阻断详情页
    userLabelMap.value = {}
  }
}

async function searchUsers(keyword: string) {
  if (!keyword || keyword.length < 3) {
    userOptions.value = []
    return
  }
  userSearching.value = true
  try {
    const res = await getUserList({ page: 1, pageSize: 20, userName: keyword })
    const list = (res.data as unknown as { list?: UserOption[] })?.list || []
    userOptions.value = list
  } finally {
    userSearching.value = false
  }
}

function statusText(s: string): string {
  const map: Record<string, string> = {
    init: '初始化中',
    active: '正常',
    disabled: '已禁用',
    archived: '已归档',
    failed: '初始化失败',
  }
  return map[s] || s
}

function statusTag(s: string): 'warning' | 'success' | 'info' | 'danger' {
  const map: Record<string, 'warning' | 'success' | 'info' | 'danger'> = {
    init: 'warning',
    active: 'success',
    disabled: 'info',
    archived: 'info',
    failed: 'danger',
  }
  return map[s] || 'info'
}

async function load() {
  loading.value = true
  try {
    const res = await getTenantList({
      page: page.value,
      pageSize: pageSize.value,
      status: statusFilter.value || undefined,
    })
    tenants.value = res.data!.list
    total.value = res.data!.total
  } finally {
    loading.value = false
  }
}

// ---- 初始化状态轮询（创建/重试后异步任务在服务端执行，前端轮询直至终态） ----
let pollTimer: ReturnType<typeof setInterval> | null = null
const POLL_INTERVAL = 3000
const POLL_MAX_TIMES = 30 // 最多 90 秒

function startInitPolling(code: string) {
  stopInitPolling()
  let times = 0
  pollTimer = setInterval(async () => {
    times++
    try {
      const res = await getTenantList({ page: 1, pageSize: 100 })
      tenants.value = res.data!.list
      total.value = res.data!.total
      const target = tenants.value.find((t) => t.tenantCode === code)
      if (target && (target.status === 'active' || target.status === 'failed')) {
        stopInitPolling()
        if (target.status === 'active') {
          ElNotification.success({
            title: '租户初始化完成',
            message: `租户 ${code} 建库与模块迁移已完成，可绑定应用路由`,
          })
        } else {
          ElNotification.error({
            title: '租户初始化失败',
            message: target.initError || '未知错误，可点击「重试初始化」重新执行',
            duration: 0,
          })
        }
        return
      }
    } catch {
      // 轮询失败忽略（网络抖动），到达上限自动停止
    }
    if (times >= POLL_MAX_TIMES) {
      stopInitPolling()
      ElNotification.warning({
        title: '初始化仍在进行',
        message: `租户 ${code} 超过 90 秒未到终态，请稍后手动刷新查看`,
      })
    }
  }, POLL_INTERVAL)
}

function stopInitPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onUnmounted(stopInitPolling)

function openCreateDialog() {
  form.value = {
    tenantCode: '',
    tenantName: '',
    dbHost: '',
    dbPort: 0,
    dbUser: '',
    dbPassword: '',
    description: '',
  }
  dialogVisible.value = true
}

function openEditDialog(row: Tenant) {
  form.value = {
    id: row.id,
    tenantCode: row.tenantCode,
    tenantName: row.tenantName,
    dbHost: row.dbHost,
    dbPort: row.dbPort,
    dbUser: row.dbUser,
    dbPassword: '',
    description: row.description,
    hasDbPass: row.hasDbPass,
  }
  dialogVisible.value = true
}

async function save() {
  try {
    if (form.value.id) {
      const { id, hasDbPass: _has, ...data } = form.value
      void _has
      await updateTenant(id as string, data)
      ElMessage.success('保存成功')
    } else {
      const code = form.value.tenantCode
      await createTenant(form.value)
      ElMessage.success(`租户 ${code} 已创建，初始化任务已投递（建库/迁移自动执行）`)
      startInitPolling(code)
    }
    dialogVisible.value = false
    if (!pollTimer) load()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function doRetry(row: Tenant) {
  await ElMessageBox.confirm(`确定重新执行租户 ${row.tenantCode} 的建库/迁移初始化吗？`, '提示', {
    type: 'warning',
  })
  await retryTenantInit(row.id)
  ElMessage.success('初始化任务已重新投递')
  startInitPolling(row.tenantCode)
}

async function doDisable(row: Tenant) {
  await ElMessageBox.confirm(
    `禁用后该租户的登录/支付/管理端查询将被拒绝，确定禁用 ${row.tenantCode} 吗？`,
    '提示',
    {
      type: 'warning',
    },
  )
  await disableTenant(row.id)
  ElMessage.success('已禁用')
  load()
}

async function doEnable(row: Tenant) {
  await enableTenant(row.id)
  ElMessage.success('已启用')
  load()
}

async function doArchive(row: Tenant) {
  await ElMessageBox.confirm(
    `归档为只读保留（物理删库仅运维 CLI），确定归档 ${row.tenantCode} 吗？`,
    '提示',
    {
      type: 'warning',
    },
  )
  await archiveTenant(row.id)
  ElMessage.success('已归档')
  load()
}

// ---- 详情抽屉 ----

async function openDetail(row: Tenant) {
  detail.value = row
  detailVisible.value = true
  detailTab.value = 'routes'
  await Promise.all([
    loadRoutes(),
    loadGrants(),
    loadOpLogs(),
    loadTenantApps(),
    loadUserLabelMap(),
  ])
}

async function loadRoutes() {
  if (!detail.value) return
  const res = await getTenantRoutes(detail.value.id)
  routes.value = res.data || []
}

async function loadGrants() {
  if (!detail.value) return
  const res = await getTenantGrants(detail.value.id)
  grants.value = res.data || []
}

async function loadOpLogs() {
  if (!detail.value) return
  const res = await getTenantOpLogs(detail.value.id, { limit: 50 })
  opLogs.value = res.data || []
}

async function addRoute() {
  if (!detail.value || !newRouteKey.value) {
    ElMessage.warning('请选择或输入 app_key')
    return
  }
  if (isRouteBound(newRouteKey.value)) {
    ElMessage.warning(`应用 ${newRouteKey.value} 已绑定到本租户`)
    return
  }
  await createTenantRoute(detail.value.id, { appKey: newRouteKey.value })
  ElMessage.success(`应用 ${newRouteKey.value} 已绑定到本租户`)
  newRouteKey.value = ''
  await loadRoutes()
}

async function removeRoute(row: TenantRoute) {
  await ElMessageBox.confirm(`确定解绑应用 ${row.appKey} 吗？`, '提示', { type: 'warning' })
  await deleteTenantRoute(row.id)
  ElMessage.success('已解绑')
  loadRoutes()
}

async function addGrant() {
  if (!detail.value || !newGrantUserID.value) {
    ElMessage.warning('请搜索并选择要授权的用户')
    return
  }
  await createTenantGrant(detail.value.id, { userId: newGrantUserID.value })
  ElMessage.success('授权成功')
  newGrantUserID.value = ''
  loadGrants()
}

async function removeGrant(row: TenantGrant) {
  if (!detail.value) return
  await deleteTenantGrant(detail.value.id, row.userId)
  ElMessage.success('已取消授权')
  loadGrants()
}

onMounted(load)
</script>
