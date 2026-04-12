<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-2">
        <el-input v-model="searchForm.group" placeholder="路由分组" size="small" class="w-40" />
        <el-select
          v-model="searchForm.authRequired"
          placeholder="认证"
          size="small"
          class="w-28"
          clearable
        >
          <el-option label="是" :value="true" />
          <el-option label="否" :value="false" />
        </el-select>
        <el-button type="primary" size="small" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button size="small" @click="handleReset">重置</el-button>
      </div>
      <div class="flex items-center gap-2">
        <el-button type="info" size="small" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" size="small" @click="handleAddApiRoute">
          <el-icon><Plus /></el-icon>
          添加 API
        </el-button>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table :data="apiRouteList" stripe class="w-full h-full" :height="'100%'">
        <el-table-column prop="id" label="ID" width="70" />
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
        <el-table-column prop="status" label="状态" width="70">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <div class="flex gap-1">
              <el-button size="small" type="primary" link @click="handleEditApiRoute(scope.row)">
                编辑
              </el-button>
              <el-button
                size="small"
                type="danger"
                link
                @click="handleDeleteApiRoute(scope.row.id)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页区域 - 紧凑布局 -->
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

    <!-- API 路由表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="路由路径" prop="path">
          <el-input v-model="formData.path" placeholder="请输入路由路径" />
        </el-form-item>
        <el-form-item label="请求方法" prop="method">
          <el-select v-model="formData.method" placeholder="请选择请求方法">
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
          <el-input v-model="formData.description" type="textarea" placeholder="请输入路由描述" />
        </el-form-item>
        <el-form-item label="是否需要认证" prop="authRequired">
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
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import {
  getApiRouteList,
  deleteApiRoute,
  createApiRoute,
  updateApiRoute,
} from '@/api/modules/api-route'
import type { ApiRoute } from '@/api/modules/api-route'

defineOptions({
  name: 'ApiRouteManagement',
})

const apiRouteList = ref<ApiRoute[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加 API 路由')
const formRef = ref<FormInstance>()
const isEditMode = ref(false)

// 搜索表单
const searchForm = reactive({
  group: '',
  authRequired: undefined as boolean | undefined,
})

// 分页信息
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

// 表单数据
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

// 表单验证规则
const formRules = reactive({
  path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }],
  method: [{ required: true, message: '请选择请求方法', trigger: 'change' }],
  name: [{ required: true, message: '请输入路由名称', trigger: 'blur' }],
  group: [{ required: true, message: '请输入路由分组', trigger: 'blur' }],
})

// 获取请求方法对应的标签类型
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

// 获取 API 路由列表
const loadApiRouteList = async () => {
  loading.value = true
  try {
    const response = await getApiRouteList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.group ? { group: searchForm.group } : {}),
      authRequired: searchForm.authRequired,
    })
    apiRouteList.value = response.data?.list || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('获取 API 路由列表失败')
    console.error('获取 API 路由列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadApiRouteList()
}

// 重置
const handleReset = () => {
  searchForm.group = ''
  searchForm.authRequired = undefined
  pagination.page = 1
  loadApiRouteList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  loadApiRouteList()
}

// 当前页码变化
const handleCurrentChange = (current: number) => {
  pagination.page = current
  loadApiRouteList()
}

// 刷新
const handleRefresh = () => {
  loadApiRouteList()
}

// 添加 API 路由
const handleAddApiRoute = () => {
  dialogTitle.value = '添加 API 路由'
  isEditMode.value = false
  resetForm()
  dialogVisible.value = true
}

// 编辑 API 路由
const handleEditApiRoute = (row: ApiRoute) => {
  dialogTitle.value = '编辑 API 路由'
  isEditMode.value = true
  Object.assign(formData, row)
  dialogVisible.value = true
}

// 删除 API 路由
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

// 重置表单
const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(formData, {
    id: 0,
    path: '',
    method: 'GET',
    name: '',
    group: '',
    description: '',
    authRequired: true,
    status: 1,
  })
}

// 关闭对话框
const handleDialogClose = () => {
  resetForm()
}

// 提交表单
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
