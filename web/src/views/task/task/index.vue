<template>
  <!-- 任务管理主容器 -->
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-2">
        <el-input v-model="searchForm.name" placeholder="任务名称" size="small" class="w-40" />
        <el-select v-model="searchForm.type" placeholder="类型" size="small" class="w-32">
          <el-option label="全部" :value="''" />
          <el-option label="HTTP" value="HTTP" />
          <el-option label="SCRIPT" value="SCRIPT" />
          <el-option label="EMAIL" value="EMAIL" />
          <el-option label="NOTIFICATION" value="NOTIFICATION" />
        </el-select>
        <el-select v-model="searchForm.status" placeholder="状态" size="small" class="w-28">
          <el-option label="全部" :value="-1" />
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" size="small" @click="handleSearch">搜索</el-button>
        <el-button size="small" @click="handleReset">重置</el-button>
      </div>
      <div class="flex items-center gap-2">
        <el-button
          type="info"
          size="small"
          v-auth-btn="'task:task:execute'"
          @click="(e) => handleRunTask(e as MouseEvent)"
        >
          <el-icon><VideoPlay /></el-icon>
          执行任务
        </el-button>
        <el-button type="primary" size="small" v-auth-btn="'task:task:add'" @click="handleAddTask">
          <el-icon><Plus /></el-icon>
          添加任务
        </el-button>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        :data="taskList"
        stripe
        class="w-full h-full"
        :height="'100%'"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="任务名称" min-width="180" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="scope">
            <el-tag :type="getTaskType(scope.row.type)" size="small">
              {{ scope.row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cronExpression" label="Cron 表达式" width="160" />
        <el-table-column prop="status" label="状态" width="70">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastRunTime" label="最后执行" width="160" />
        <el-table-column prop="nextRunTime" label="下次执行" width="160" />
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <div class="flex gap-1">
              <el-button
                v-auth-btn="'task:task:edit'"
                size="small"
                type="primary"
                link
                @click="handleEditTask(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'task:task:execute'"
                size="small"
                type="success"
                link
                @click="(e) => handleRunTask(e as MouseEvent, scope.row.id)"
              >
                执行
              </el-button>
              <el-button size="small" type="primary" link @click="handleViewLog(scope.row)">
                日志
              </el-button>
              <el-button
                v-auth-btn="'task:task:delete'"
                size="small"
                type="danger"
                link
                @click="handleDeleteTask(scope.row.id)"
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

    <!-- 任务表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="任务类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择任务类型">
            <el-option label="HTTP" value="HTTP" />
            <el-option label="SCRIPT" value="SCRIPT" />
            <el-option label="EMAIL" value="EMAIL" />
            <el-option label="NOTIFICATION" value="NOTIFICATION" />
          </el-select>
        </el-form-item>
        <el-form-item label="Cron 表达式" prop="cronExpression">
          <el-input v-model="formData.cronExpression" placeholder="请输入 Cron 表达式" />
        </el-form-item>
        <el-form-item label="任务参数" prop="params">
          <el-input
            v-model="formData.paramsJson"
            type="textarea"
            placeholder="请输入 JSON 格式的任务参数"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="formData.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 任务日志对话框 -->
    <el-dialog v-model="logDialogVisible" title="任务日志" width="800px">
      <el-table :data="taskLogs" stripe class="w-full">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="status" label="执行状态" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">
              {{ scope.row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="执行信息" />
        <el-table-column prop="executionTime" label="执行时间 (ms)" width="120" />
        <el-table-column prop="executedAt" label="执行时间" width="200" />
      </el-table>
      <!-- 日志分页 -->
      <div class="flex justify-center mt-4">
        <el-pagination
          v-model:current-page="logPagination.page"
          v-model:page-size="logPagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="logPagination.total"
          @size-change="handleLogSizeChange"
          @current-change="handleLogCurrentChange"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { Plus, VideoPlay } from '@element-plus/icons-vue'
import {
  getTaskList,
  deleteTask,
  createTask,
  updateTask,
  runTask,
  getTaskLogs,
} from '@/api/modules/task'
import type { Task } from '@/api/modules/task'

defineOptions({
  name: 'TaskManagement',
})

const taskList = ref<Task[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加任务')
const formRef = ref<FormInstance>()
const isEditMode = ref(false)
const selectedTasks = ref<Task[]>([])
const logDialogVisible = ref(false)
const currentTaskId = ref(0)

// 搜索表单
const searchForm = reactive({
  name: '',
  type: '',
  status: -1,
})

// 分页信息
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

// 日志分页信息
const logPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

// 表单数据
const formData = reactive({
  id: 0,
  name: '',
  type: 'HTTP',
  cronExpression: '',
  params: {} as Record<string, unknown>,
  paramsJson: '{}',
  status: 1,
})

// 任务日志
const taskLogs = ref<
  Array<{
    id: number
    taskId: number
    status: number
    message: string
    executedAt: string
    executionTime: number
  }>
>([])

// 表单验证规则
const formRules = reactive({
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择任务类型', trigger: 'change' }],
  cronExpression: [{ required: true, message: '请输入 Cron 表达式', trigger: 'blur' }],
  paramsJson: [
    { required: true, message: '请输入任务参数', trigger: 'blur' },
    {
      validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
        try {
          JSON.parse(value)
          callback()
        } catch {
          callback(new Error('请输入有效的 JSON 格式'))
        }
      },
      trigger: 'blur',
    } as const,
  ],
})

// 监听paramsJson变化，更新params对象
watch(
  () => formData.paramsJson,
  (newValue) => {
    try {
      formData.params = JSON.parse(newValue)
    } catch {
      // 忽略JSON解析错误，验证规则会处理
    }
  },
)

// 获取任务类型对应的标签类型
const getTaskType = (type: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
  const typeMap: Record<string, 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    HTTP: 'primary',
    SCRIPT: 'success',
    EMAIL: 'warning',
    NOTIFICATION: 'info',
  }
  return typeMap[type] || 'info'
}

// 获取任务列表
const loadTaskList = async () => {
  loading.value = true
  try {
    const response = await getTaskList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.name ? { name: searchForm.name } : {}),
      ...(searchForm.type ? { type: searchForm.type } : {}),
      ...(searchForm.status !== -1 ? { status: searchForm.status } : {}),
    })
    taskList.value = response.data?.list || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('获取任务列表失败')
    console.error('获取任务列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadTaskList()
}

// 重置
const handleReset = () => {
  searchForm.name = ''
  searchForm.type = ''
  searchForm.status = -1
  pagination.page = 1
  loadTaskList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  loadTaskList()
}

// 当前页码变化
const handleCurrentChange = (current: number) => {
  pagination.page = current
  loadTaskList()
}

// 添加任务
const handleAddTask = () => {
  dialogTitle.value = '添加任务'
  isEditMode.value = false
  resetForm()
  dialogVisible.value = true
}

// 编辑任务
const handleEditTask = (row: Task) => {
  dialogTitle.value = '编辑任务'
  isEditMode.value = true
  // 复制数据到表单
  Object.assign(formData, {
    ...row,
    paramsJson: JSON.stringify(row.params, null, 2),
  })
  dialogVisible.value = true
}

// 删除任务
const handleDeleteTask = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该任务吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteTask(id)
    ElMessage.success('删除成功')
    loadTaskList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除任务失败:', error)
    }
  }
}

// 执行任务
const handleRunTask = async (event?: MouseEvent, id?: number) => {
  try {
    let taskIds: number[] = []
    if (id) {
      taskIds = [id]
    } else if (selectedTasks.value.length > 0) {
      taskIds = selectedTasks.value.map((task) => task.id)
    } else {
      ElMessage.warning('请选择要执行的任务')
      return
    }

    for (const taskId of taskIds) {
      await runTask(taskId)
    }

    ElMessage.success('任务执行成功')
    loadTaskList()
  } catch (error) {
    ElMessage.error('任务执行失败')
    console.error('任务执行失败:', error)
  }
}

// 查看任务日志
const handleViewLog = async (row: Task) => {
  currentTaskId.value = row.id
  logPagination.page = 1
  await loadTaskLogs()
  logDialogVisible.value = true
}

// 获取任务日志
const loadTaskLogs = async () => {
  try {
    const response = await getTaskLogs(currentTaskId.value, {
      page: logPagination.page,
      pageSize: logPagination.pageSize,
    })
    taskLogs.value = response.data?.list || []
    logPagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('获取任务日志失败')
    console.error('获取任务日志失败:', error)
  }
}

// 日志分页大小变化
const handleLogSizeChange = (size: number) => {
  logPagination.pageSize = size
  loadTaskLogs()
}

// 日志当前页码变化
const handleLogCurrentChange = (current: number) => {
  logPagination.page = current
  loadTaskLogs()
}

// 选择任务
const handleSelectionChange = (selection: Task[]) => {
  selectedTasks.value = selection
}

// 重置表单
const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(formData, {
    id: 0,
    name: '',
    type: 'HTTP',
    cronExpression: '',
    params: {},
    paramsJson: '{}',
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

    // 解析JSON参数
    try {
      formData.params = JSON.parse(formData.paramsJson)
    } catch {
      ElMessage.error('任务参数格式错误')
      return
    }

    if (isEditMode.value) {
      await updateTask(formData.id, {
        name: formData.name,
        type: formData.type,
        cronExpression: formData.cronExpression,
        params: formData.params,
        status: formData.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createTask({
        name: formData.name,
        type: formData.type,
        cronExpression: formData.cronExpression,
        params: formData.params,
        status: formData.status,
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadTaskList()
  } catch (error) {
    if (error === 'Validation failed') {
      return
    }
    ElMessage.error(isEditMode.value ? '更新失败' : '创建失败')
    console.error(`${isEditMode.value ? '更新' : '创建'}任务失败:`, error)
  }
}

onMounted(() => {
  loadTaskList()
})
</script>

<style scoped>
/* 暗黑模式支持 */
:deep(.el-table) {
  --el-table-bg-color: transparent;
}

:deep(.el-dialog) {
  border-radius: 0.5rem;
}

:deep(.el-pagination) {
  --el-pagination-text-color: theme('colors.gray.600');
}

:deep(.dark .el-pagination) {
  --el-pagination-text-color: theme('colors.gray.400');
}
</style>
