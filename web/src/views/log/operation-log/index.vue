<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-2 flex-wrap">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始"
          end-placeholder="结束"
          value-format="YYYY-MM-DD"
          size="small"
          class="w-64"
        />
        <el-input
          v-model="searchForm.username"
          placeholder="用户名"
          clearable
          size="small"
          class="w-36"
        />
        <el-input
          v-model="searchForm.module"
          placeholder="模块"
          clearable
          size="small"
          class="w-36"
        />
        <el-select
          v-model="searchForm.operationType"
          placeholder="操作类型"
          clearable
          size="small"
          class="w-32"
        >
          <el-option label="登录" :value="1" />
          <el-option label="登出" :value="2" />
          <el-option label="新增" :value="3" />
          <el-option label="修改" :value="4" />
          <el-option label="删除" :value="5" />
          <el-option label="查询" :value="6" />
        </el-select>
        <el-select
          v-model="searchForm.isSuccess"
          placeholder="结果"
          clearable
          size="small"
          class="w-28"
        >
          <el-option label="成功" :value="true" />
          <el-option label="失败" :value="false" />
        </el-select>
        <el-button type="primary" size="small" @click="loadLogList">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button size="small" @click="handleReset">重置</el-button>
      </div>
      <div class="flex items-center gap-2">
        <el-button
          type="danger"
          size="small"
          @click="handleBatchDelete"
          :disabled="selectedIds.length === 0"
        >
          <el-icon><Delete /></el-icon>
          批量删除
        </el-button>
        <el-button type="warning" size="small" @click="handleClearLogs">
          <el-icon><Delete /></el-icon>
          清空日志
        </el-button>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="logList"
        stripe
        class="w-full h-full"
        :height="'100%'"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="username" label="用户名" width="100" />
        <el-table-column prop="module" label="模块" width="120" />
        <el-table-column
          prop="description"
          label="操作内容"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column label="类型" width="80">
          <template #default="scope">
            <el-tag :type="getOperationTypeTag(scope.row.operationType)" size="small">
              {{ getOperationTypeName(scope.row.operationType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="结果" width="70">
          <template #default="scope">
            <el-tag :type="scope.row.isSuccess ? 'success' : 'danger'" size="small">
              {{ scope.row.isSuccess ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="clientIp" label="IP 地址" width="120" />
        <el-table-column label="操作时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <div class="flex gap-1">
              <el-button size="small" type="primary" link @click="handleViewDetail(scope.row)">
                详情
              </el-button>
              <el-button size="small" type="danger" link @click="handleDeleteLog(scope.row.id)">
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

    <!-- 日志详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="日志详情" width="700px">
      <el-descriptions :column="2" border v-if="currentLog">
        <el-descriptions-item label="日志 ID">{{ currentLog.id }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ currentLog.username }}</el-descriptions-item>
        <el-descriptions-item label="操作模块">{{ currentLog.module }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">
          <el-tag :type="getOperationTypeTag(currentLog.operationType)">
            {{ getOperationTypeName(currentLog.operationType) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作结果">
          <el-tag :type="currentLog.isSuccess ? 'success' : 'danger'">
            {{ currentLog.isSuccess ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="IP 地址">{{ currentLog.clientIp }}</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{
          formatDate(currentLog.created_at)
        }}</el-descriptions-item>
        <el-descriptions-item label="用户代理" :span="2">
          <div class="max-h-28 overflow-y-auto">{{ currentLog.userAgent }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="操作内容" :span="2">
          <div class="max-h-52 overflow-y-auto">{{ currentLog.description }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="请求参数" :span="2" v-if="currentLog.params">
          <div class="max-h-52 overflow-y-auto">
            <pre>{{ currentLog.params }}</pre>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="响应结果" :span="2" v-if="currentLog.result">
          <div class="max-h-52 overflow-y-auto">
            <pre>{{ currentLog.result }}</pre>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="错误信息" :span="2" v-if="currentLog.errorMessage">
          <div class="max-h-40 overflow-y-auto text-red-500">
            {{ currentLog.errorMessage }}
          </div>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 清空日志对话框 -->
    <el-dialog v-model="clearDialogVisible" title="清空日志" width="500px">
      <el-form :model="clearForm" label-width="100px">
        <el-form-item label="开始日期" required>
          <el-date-picker
            v-model="clearForm.startTime"
            type="date"
            placeholder="选择开始日期"
            value-format="YYYY-MM-DD"
            class="w-full"
          />
        </el-form-item>
        <el-form-item label="结束日期" required>
          <el-date-picker
            v-model="clearForm.endTime"
            type="date"
            placeholder="选择结束日期"
            value-format="YYYY-MM-DD"
            class="w-full"
          />
        </el-form-item>
      </el-form>
      <el-alert
        title="此操作将清空指定时间范围内的所有操作日志，请谨慎操作！"
        type="warning"
        :closable="false"
        class="mb-5"
      />
      <template #footer>
        <el-button @click="clearDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="handleClearSubmit">确定清空</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Delete } from '@element-plus/icons-vue'
import {
  getLogList,
  deleteLogByID,
  batchDeleteLog,
  clearLogByTimeRange,
} from '@/api/modules/operation-log'

defineOptions({
  name: 'OperationLogManagement',
})

interface OperationLog {
  id: string
  username: string
  module: string
  description: string
  operationType: number
  isSuccess: boolean
  clientIp: string
  userAgent: string
  params?: string
  result?: string
  errorMessage: string
  created_at: string
}

// 日志列表数据
const logList = ref<OperationLog[]>([])
// 加载状态
const loading = ref(false)
// 分页数据
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
// 搜索表单
const searchForm = reactive({
  username: '',
  module: '',
  operationType: undefined as number | undefined,
  isSuccess: undefined as boolean | undefined,
})
// 日期范围
const dateRange = ref<[string, string] | null>(null)
// 选中的 ID
const selectedIds = ref<string[]>([])
// 详情对话框
const detailDialogVisible = ref(false)
const currentLog = ref<OperationLog | null>(null)
// 清空对话框
const clearDialogVisible = ref(false)
const clearForm = reactive({
  startTime: '',
  endTime: '',
})

// 获取操作类型名称
const getOperationTypeName = (type: number): string => {
  const typeMap: Record<number, string> = {
    1: '登录',
    2: '登出',
    3: '新增',
    4: '修改',
    5: '删除',
    6: '查询',
  }
  return typeMap[type] || '其他'
}

// 获取操作类型标签颜色
const getOperationTypeTag = (
  type: number,
): 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
  const tagMap: Record<number, 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    1: 'primary',
    2: 'info',
    3: 'success',
    4: 'warning',
    5: 'danger',
    6: 'info',
  }
  return tagMap[type] || 'info'
}

// 获取日志列表
const loadLogList = async () => {
  loading.value = true
  try {
    const params: {
      page: number
      pageSize: number
      startTime?: string
      endTime?: string
      username?: string
      module?: string
      operationType?: number
      isSuccess?: boolean
    } = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.username ? { username: searchForm.username } : {}),
      ...(searchForm.module ? { module: searchForm.module } : {}),
      ...(searchForm.operationType !== undefined
        ? { operationType: searchForm.operationType }
        : {}),
      ...(searchForm.isSuccess !== undefined ? { isSuccess: searchForm.isSuccess } : {}),
    }

    if (dateRange.value && dateRange.value.length === 2) {
      params.startTime = dateRange.value[0]
      params.endTime = dateRange.value[1]
    }

    const response = await getLogList(params)
    const pageData = response.data as Record<string, unknown>
    logList.value = (pageData?.list as OperationLog[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    ElMessage.error('获取日志列表失败')
    console.error('获取日志列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 重置搜索条件
const handleReset = () => {
  searchForm.username = ''
  searchForm.module = ''
  searchForm.operationType = undefined
  searchForm.isSuccess = undefined
  dateRange.value = null
  pagination.page = 1
  loadLogList()
}

// 分页处理
const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadLogList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadLogList()
}

// 表格选择变化
const handleSelectionChange = (selection: OperationLog[]) => {
  selectedIds.value = selection.map((item) => item.id)
}

// 查看详情
const handleViewDetail = (log: OperationLog) => {
  currentLog.value = log
  detailDialogVisible.value = true
}

// 删除单条日志
const handleDeleteLog = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除这条日志吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteLogByID(id)
    ElMessage.success('删除成功')
    loadLogList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除日志失败:', error)
    }
  }
}

// 批量删除
const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 条日志吗？`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await batchDeleteLog({ ids: selectedIds.value })
    ElMessage.success('批量删除成功')
    loadLogList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
      console.error('批量删除日志失败:', error)
    }
  }
}

// 清空日志
const handleClearLogs = () => {
  clearForm.startTime = ''
  clearForm.endTime = ''
  clearDialogVisible.value = true
}

// 提交清空
const handleClearSubmit = async () => {
  if (!clearForm.startTime || !clearForm.endTime) {
    ElMessage.warning('请选择日期范围')
    return
  }

  try {
    await ElMessageBox.confirm('此操作不可逆，确定要清空该时间范围内的所有日志吗？', '严重警告', {
      confirmButtonText: '确定清空',
      cancelButtonText: '取消',
      type: 'error',
    })

    await clearLogByTimeRange(clearForm)
    ElMessage.success('清空成功')
    clearDialogVisible.value = false
    loadLogList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('清空失败')
      console.error('清空日志失败:', error)
    }
  }
}

// 格式化日期
const formatDate = (date: string | null | undefined) => {
  if (!date) return '-'
  try {
    return new Date(date).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch {
    return '-'
  }
}

onMounted(() => {
  loadLogList()
})
</script>
