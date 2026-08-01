<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <div class="flex flex-col gap-3 p-4 border-b">
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.group"
          placeholder="路由分组"
          clearable
          style="width: 192px"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>
        <el-select
          v-model="searchForm.authRequired"
          placeholder="认证"
          clearable
          style="width: 192px"
        >
          <el-option label="是" :value="true" />
          <el-option label="否" :value="false" />
        </el-select>
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 192px">
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" @click="handleSearch">
          <el-icon>
            <Search />
          </el-icon>
          搜索
        </el-button>
        <el-button @click="handleReset">
          <el-icon>
            <Refresh />
          </el-icon>
          重置
        </el-button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <el-button type="info" @click="handleRefresh">
          <el-icon>
            <Refresh />
          </el-icon>
          刷新
        </el-button>
        <el-button type="primary" v-auth-btn="'system:api-route:add'" @click="handleAddApiRoute">
          <el-icon>
            <Plus />
          </el-icon>
          添加 API
        </el-button>
      </div>
    </div>

    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500">
            <Connection />
          </el-icon>
          路由总数：<span class="font-medium">{{ pagination.total }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <Lock />
          </el-icon>
          需认证：<span class="font-medium">{{
            apiRouteList.filter((r) => r.authRequired).length
          }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-orange-500">
            <SuccessFilled />
          </el-icon>
          已启用：<span class="font-medium">{{
            apiRouteList.filter((r) => r.status === 1).length
          }}</span>
        </span>
      </div>
    </div>

    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="apiRouteList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="group" label="路由分组" width="120" />
        <el-table-column prop="path" label="路由路径" min-width="200" />
        <el-table-column prop="method" label="方法" width="90">
          <template #default="scope">
            <el-tag :type="getMethodType(scope.row.method)" size="small">
              {{ scope.row.method }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="路由名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column prop="authRequired" label="认证" width="70">
          <template #default="scope">
            <el-tag :type="scope.row.authRequired ? 'success' : 'info'" size="small">
              {{ scope.row.authRequired ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              size="small"
              @change="handleStatusChange(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button
                v-auth-btn="'system:api-route:edit'"
                type="primary"
                size="small"
                @click="handleEditApiRoute(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:api-route:delete'"
                type="danger"
                size="small"
                @click="handleDeleteApiRoute(scope.row.id)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div
      class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700"
    >
      <span class="text-sm text-gray-600 dark:text-gray-400">共 {{ pagination.total }} 条</span>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        layout="sizes, prev, pager, next"
        :total="pagination.total"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        small
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="110px">
        <el-form-item label="路由路径" prop="path">
          <el-input v-model="formData.path" placeholder="如：/api/admin/user" />
        </el-form-item>
        <el-form-item label="请求方法" prop="method">
          <el-select v-model="formData.method" placeholder="请选择请求方法" class="w-full">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item label="路由名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入路由名称" />
        </el-form-item>
        <el-form-item label="路由分组" prop="group">
          <el-input v-model="formData.group" placeholder="请输入路由分组" />
        </el-form-item>
        <el-form-item label="路由描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            placeholder="请输入路由描述"
            :rows="3"
          />
        </el-form-item>
        <el-form-item label="需要认证" prop="authRequired">
          <el-switch v-model="formData.authRequired" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="formData.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus, Search, Refresh, Connection, Lock, SuccessFilled } from '@element-plus/icons-vue'
import {
  getApiRouteList,
  deleteApiRoute,
  createApiRoute,
  updateApiRoute,
} from '@/api/modules/api-route'
import type { ApiRoute } from '@/api/modules/api-route'
import { formatDate } from '@/utils/format'

defineOptions({
  name: 'ApiRouteManagement',
})

const apiRouteList = ref<ApiRoute[]>([])
const loading = ref(false)
const isLoading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加 API 路由')
const formRef = ref<FormInstance>()
const isEditMode = ref(false)

const searchForm = reactive({
  group: '',
  authRequired: undefined as boolean | undefined,
  status: undefined as number | undefined,
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const formData = reactive({
  id: '',
  path: '',
  method: 'GET',
  name: '',
  group: '',
  description: '',
  authRequired: false,
  status: 1,
})

const formRules = reactive({
  path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }],
  method: [{ required: true, message: '请选择请求方法', trigger: 'change' }],
  name: [{ required: true, message: '请输入路由名称', trigger: 'blur' }],
  group: [{ required: true, message: '请输入路由分组', trigger: 'blur' }],
})

const getMethodType = (method: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
  const methodMap: Record<string, 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    GET: 'primary',
    POST: 'success',
    PUT: 'warning',
    DELETE: 'danger',
    PATCH: 'info',
  }
  return methodMap[method] || 'info'
}

const loadApiRouteList = async () => {
  if (isLoading.value) return

  isLoading.value = true
  loading.value = true
  try {
    const response = await getApiRouteList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.group ? { group: searchForm.group } : {}),
      ...(searchForm.authRequired !== undefined ? { authRequired: searchForm.authRequired } : {}),
      ...(searchForm.status !== undefined ? { status: searchForm.status } : {}),
    })
    apiRouteList.value = response.data?.list || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    console.error('获取 API 路由列表失败:', error)
    apiRouteList.value = []
    pagination.total = 0
  } finally {
    loading.value = false
    isLoading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadApiRouteList()
}

const handleReset = () => {
  searchForm.group = ''
  searchForm.authRequired = undefined
  searchForm.status = undefined
  pagination.page = 1
  loadApiRouteList()
}

// 状态变更
const handleStatusChange = async (apiRoute: ApiRoute) => {
  try {
    await updateApiRoute(apiRoute.id, {
      path: apiRoute.path,
      method: apiRoute.method,
      name: apiRoute.name,
      group: apiRoute.group,
      description: apiRoute.description,
      authRequired: apiRoute.authRequired,
      status: apiRoute.status,
    })
    ElMessage.success('状态更新成功')
  } catch (error) {
    ElMessage.error('状态更新失败')
    console.error('状态更新失败:', error)
    // 恢复原状态
    apiRoute.status = apiRoute.status === 1 ? 0 : 1
  }
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  loadApiRouteList()
}

const handleCurrentChange = (current: number) => {
  pagination.page = current
  loadApiRouteList()
}

const handleRefresh = () => {
  loadApiRouteList()
}

const handleAddApiRoute = () => {
  dialogTitle.value = '添加 API 路由'
  isEditMode.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEditApiRoute = (row: ApiRoute) => {
  dialogTitle.value = '编辑 API 路由'
  isEditMode.value = true
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleDeleteApiRoute = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除该 API 路由吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteApiRoute(id)
    ElMessage.success('删除成功')
    loadApiRouteList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除 API 路由失败:', error)
    }
  }
}

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(formData, {
    id: '',
    path: '',
    method: 'GET',
    name: '',
    group: '',
    description: '',
    authRequired: true,
    status: 1,
  })
}

const handleDialogClose = () => {
  resetForm()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    if (isEditMode.value) {
      await updateApiRoute(formData.id, formData)
      ElMessage.success('更新成功')
    } else {
      await createApiRoute(formData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadApiRouteList()
  } catch (error) {
    if (error === 'Validation failed') {
      return
    }
    ElMessage.error(isEditMode.value ? '更新失败' : '创建失败')
    console.error(`${isEditMode.value ? '更新' : '创建'}API 路由失败:`, error)
  }
}

onMounted(() => {
  loadApiRouteList()
})
</script>

<style scoped>
.flex-1 {
  min-height: 0;
}

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-table__header th) {
  background-color: var(--bg-tertiary) !important;
  color: var(--text-primary) !important;
  font-weight: 600 !important;
}

:deep(.el-table__fixed-header-wrapper th),
:deep(.el-table__fixed-right-header-wrapper th) {
  background-color: var(--bg-tertiary) !important;
}

:deep(.el-table__fixed-body-wrapper),
:deep(.el-table__fixed-right-body-wrapper) {
  background-color: var(--bg-primary);
}

:deep(.el-table .cell) {
  white-space: normal;
  word-wrap: break-word;
}

:deep(.el-table__body-wrapper) {
  overflow: auto !important;
}
</style>
