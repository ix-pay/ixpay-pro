<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 统计面板 -->
    <div class="grid grid-cols-4 gap-4 p-4 border-b border-gray-200 dark:border-gray-700">
      <div
        class="flex flex-col items-center justify-center p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg"
      >
        <div class="text-2xl font-bold text-blue-600 dark:text-blue-400">
          {{ dashboard.totalTasks }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">任务总数</div>
      </div>
      <div
        class="flex flex-col items-center justify-center p-4 bg-green-50 dark:bg-green-900/20 rounded-lg"
      >
        <div class="text-2xl font-bold text-green-600 dark:text-green-400">
          {{ dashboard.enabledTasks }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">启用数</div>
      </div>
      <div
        class="flex flex-col items-center justify-center p-4 bg-red-50 dark:bg-red-900/20 rounded-lg"
      >
        <div class="text-2xl font-bold text-red-600 dark:text-red-400">
          {{ dashboard.disabledTasks }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">禁用数</div>
      </div>
      <div
        class="flex flex-col items-center justify-center p-4 bg-orange-50 dark:bg-orange-900/20 rounded-lg"
      >
        <div class="text-2xl font-bold text-orange-600 dark:text-orange-400">
          {{ dashboard.todayExecutions }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">今日执行数</div>
      </div>
    </div>

    <div class="flex flex-col gap-3 p-4 border-b">
      <div class="flex flex-wrap items-center gap-3">
        <el-input v-model="searchForm.taskId" placeholder="任务ID" style="width: 192px" />
        <el-select v-model="searchForm.taskType" placeholder="任务类型" style="width: 192px">
          <el-option label="全部" value="" />
          <el-option
            v-for="item in taskTypeOptions"
            :key="item.id"
            :label="item.itemValue"
            :value="item.itemKey"
          />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <el-button
          type="info"
          v-auth-btn="'task:task:execute'"
          @click="(e) => handleRunTask(e as MouseEvent)"
        >
          <el-icon>
            <VideoPlay />
          </el-icon>
          执行任务
        </el-button>
        <el-button type="primary" v-auth-btn="'task:task:add'" @click="handleAddTask">
          <el-icon>
            <Plus />
          </el-icon>
          添加任务
        </el-button>
      </div>
    </div>

    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="text-gray-600 dark:text-gray-400">
          任务总数：<strong class="text-gray-900 dark:text-white">{{ pagination.total }}</strong>
        </span>
        <span class="text-gray-600 dark:text-gray-400">
          选中任务：<strong class="text-gray-900 dark:text-white">{{
            selectedTasks.length
          }}</strong>
        </span>
      </div>
    </div>

    <div class="flex-1 overflow-hidden">
      <el-table
        :data="taskList"
        stripe
        class="w-full h-full"
        :height="'100%'"
        v-loading="loading"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="taskId" label="任务ID" width="150" />
        <el-table-column prop="taskType" label="任务类型" width="120">
          <template #default="scope">
            <el-tag type="primary" size="small">
              {{ getTaskTypeLabel(scope.row.taskType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="expression" label="Cron表达式" min-width="160" />
        <el-table-column prop="description" label="描述" min-width="180" />
        <el-table-column prop="status" label="数据库状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'enabled' ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 'enabled' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="statusLabel" label="运行状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.statusLabel === 'running' ? 'primary' : 'info'" size="small">
              {{ scope.row.statusLabel === 'running' ? '运行中' : '已停止' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="concurrency" label="并发模式" width="100">
          <template #default="scope">
            {{ getConcurrencyLabel(scope.row.concurrency) }}
          </template>
        </el-table-column>
        <el-table-column prop="timeout" label="超时(秒)" width="90" />
        <el-table-column prop="lastRunAt" label="最后执行" width="160" />
        <el-table-column prop="nextRunAt" label="下次执行" width="160" />
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button
                v-auth-btn="'task:task:execute'"
                type="success"
                size="small"
                @click="(e) => handleRunTask(e as MouseEvent, scope.row.taskId)"
              >
                执行
              </el-button>
              <el-button
                v-if="scope.row.status === 'disabled'"
                v-auth-btn="'task:task:enable'"
                type="success"
                size="small"
                @click="handleEnableTask(scope.row)"
              >
                启用
              </el-button>
              <el-button
                v-if="scope.row.status === 'enabled'"
                v-auth-btn="'task:task:disable'"
                type="warning"
                size="small"
                @click="handleDisableTask(scope.row)"
              >
                禁用
              </el-button>
              <el-button type="primary" size="small" @click="handleEditTask(scope.row)">
                编辑
              </el-button>
              <el-button type="info" size="small" @click="handleViewLog(scope.row)">
                日志
              </el-button>
              <el-button
                v-auth-btn="'task:task:delete'"
                type="danger"
                size="small"
                @click="handleDeleteTask(scope.row.id)"
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
      width="600px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="任务ID" prop="taskId">
          <el-input v-model="formData.taskId" placeholder="请输入任务ID" :disabled="isEdit" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="任务类型" prop="taskType">
              <el-select v-model="formData.taskType" placeholder="请选择">
                <el-option
                  v-for="item in taskTypeOptions"
                  :key="item.id"
                  :label="item.itemValue"
                  :value="item.itemKey"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="调度类型" prop="type">
              <el-select v-model="formData.type" placeholder="请选择">
                <el-option label="定时任务" value="cron" />
                <el-option label="一次性任务" value="one_time" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="表达式" prop="expression">
          <el-input
            v-model="formData.expression"
            :placeholder="
              formData.type === 'cron'
                ? '请输入 Cron 表达式，如：0 0 2 * * ?'
                : '请输入执行时间 (RFC3339)'
            "
          />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input v-model="formData.description" type="textarea" placeholder="请输入任务描述" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="并发模式" prop="concurrency">
              <el-select v-model="formData.concurrency" placeholder="请选择">
                <el-option label="允许并发" value="allow" />
                <el-option label="跳过执行" value="skip" />
                <el-option label="等待执行" value="wait" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="超时(秒)" prop="timeout">
              <el-input-number
                v-model="formData.timeout"
                :min="1"
                :max="3600"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="重试次数" prop="retryCount">
          <el-input-number v-model="formData.retryCount" :min="0" :max="10" />
        </el-form-item>

        <template v-if="formData.taskType === 'http'">
          <el-divider content-position="left">HTTP 参数</el-divider>
          <el-form-item label="请求URL" prop="params.url">
            <el-input v-model="formData.params.url" placeholder="http://example.com/api" />
          </el-form-item>
          <el-form-item label="请求方法" prop="params.method">
            <el-select v-model="formData.params.method" style="width: 100%">
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
              <el-option label="DELETE" value="DELETE" />
            </el-select>
          </el-form-item>
          <el-form-item label="请求头" prop="params.headers">
            <el-input
              v-model="formData.params.headersJson"
              type="textarea"
              placeholder='{"Content-Type": "application/json"}'
              :rows="2"
            />
          </el-form-item>
          <el-form-item label="请求体" prop="params.body">
            <el-input
              v-model="formData.params.body"
              type="textarea"
              placeholder="请求体内容"
              :rows="3"
            />
          </el-form-item>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="超时(秒)" prop="params.timeout">
                <el-input-number
                  v-model="formData.params.timeout"
                  :min="1"
                  :max="300"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="期望状态码" prop="params.expected_status">
                <el-input-number
                  v-model="formData.params.expected_status"
                  :min="100"
                  :max="599"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </template>

        <template v-if="formData.taskType === 'database'">
          <el-divider content-position="left">数据库参数</el-divider>
          <el-form-item label="数据库类型" prop="params.db_type">
            <el-select v-model="formData.params.db_type" style="width: 100%">
              <el-option label="PostgreSQL" value="postgres" />
              <el-option label="MySQL" value="mysql" />
            </el-select>
          </el-form-item>
          <el-form-item label="SQL 查询" prop="params.query">
            <el-input
              v-model="formData.params.query"
              type="textarea"
              placeholder="SELECT * FROM table WHERE ..."
              :rows="4"
            />
          </el-form-item>
        </template>

        <template v-if="formData.taskType === 'cache'">
          <el-divider content-position="left">缓存参数</el-divider>
          <el-form-item label="操作类型" prop="params.action">
            <el-select v-model="formData.params.action" style="width: 100%">
              <el-option label="刷新" value="refresh" />
              <el-option label="清理" value="clear" />
            </el-select>
          </el-form-item>
          <el-form-item label="缓存键" prop="params.cache_keys">
            <el-input
              v-model="formData.params.cache_keys_str"
              placeholder="多个键用逗号分隔，如：key1,key2"
            />
          </el-form-item>
          <el-form-item label="通配符模式" prop="params.pattern">
            <el-input v-model="formData.params.pattern" placeholder="如：user:*" />
          </el-form-item>
          <el-form-item label="TTL(秒)" prop="params.ttl">
            <el-input-number
              v-model="formData.params.ttl"
              :min="0"
              :max="86400"
              style="width: 100%"
            />
          </el-form-item>
        </template>

        <template v-if="formData.taskType === 'script'">
          <el-divider content-position="left">脚本参数</el-divider>
          <el-form-item label="命令" prop="params.command">
            <el-input v-model="formData.params.command" placeholder="/usr/bin/python3" />
          </el-form-item>
          <el-form-item label="参数" prop="params.args">
            <el-input v-model="formData.params.args_str" placeholder="多个参数用逗号分隔" />
          </el-form-item>
          <el-form-item label="工作目录" prop="params.work_dir">
            <el-input v-model="formData.params.work_dir" placeholder="/opt/scripts" />
          </el-form-item>
          <el-form-item label="超时(秒)" prop="params.timeout">
            <el-input-number
              v-model="formData.params.timeout"
              :min="1"
              :max="3600"
              style="width: 100%"
            />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="logDialogVisible" title="任务日志" width="800px">
      <el-table :data="taskLogs" stripe class="w-full" v-loading="logLoading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="taskName" label="任务名称" width="120" />
        <el-table-column prop="result" label="执行结果" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.result === 'success' ? 'success' : 'danger'">
              {{ scope.row.result === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="triggerType" label="触发类型" width="100" />
        <el-table-column prop="duration" label="耗时 (ms)" width="100" />
        <el-table-column prop="errorInfo" label="错误信息" />
        <el-table-column prop="executeAt" label="执行时间" width="180" />
      </el-table>
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
import { ref, onMounted, reactive } from 'vue'
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
  enableTask,
  disableTask,
  getTaskById,
  getTaskDashboard,
} from '@/api/modules/task'
import type { Task, TaskLog } from '@/api/modules/task'
import { getDictItemsByCode } from '@/api/modules/dict'
import type { DictItem } from '@/api/modules/dict'

defineOptions({
  name: 'TaskManagement',
})

const taskList = ref<Task[]>([])
const loading = ref(false)
const logLoading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加任务')
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const selectedTasks = ref<Task[]>([])
const logDialogVisible = ref(false)
const currentTaskId = ref('')
const taskTypeOptions = ref<DictItem[]>([])

// 统计面板数据
const dashboard = reactive({
  totalTasks: 0,
  enabledTasks: 0,
  disabledTasks: 0,
  todayExecutions: 0,
})

const searchForm = reactive({
  taskId: '',
  taskType: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const logPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const formData = reactive({
  id: '',
  taskId: '',
  taskType: 'http',
  type: 'cron',
  expression: '',
  description: '',
  group: '',
  retryCount: 0,
  concurrency: 'allow',
  timeout: 30,
  params: {
    url: '',
    method: 'GET',
    headersJson: '',
    body: '',
    timeout: 30,
    expected_status: 200,
    db_type: 'postgres',
    query: '',
    action: 'refresh',
    cache_keys_str: '',
    pattern: '',
    ttl: 3600,
    command: '',
    args_str: '',
    work_dir: '',
  },
})

const taskLogs = ref<TaskLog[]>([])

const formRules = reactive({
  taskId: [{ required: true, message: '请输入任务ID', trigger: 'blur' }],
  taskType: [{ required: true, message: '请选择任务类型', trigger: 'change' }],
  type: [{ required: true, message: '请选择调度类型', trigger: 'change' }],
  expression: [{ required: true, message: '请输入表达式', trigger: 'blur' }],
})

const loadTaskTypes = async () => {
  try {
    const res = await getDictItemsByCode('task_type')
    taskTypeOptions.value = (res.data as DictItem[]) || []
  } catch (error) {
    console.error('获取任务类型字典失败:', error)
  }
}

const getTaskTypeLabel = (taskType: string) => {
  const item = taskTypeOptions.value.find((d) => d.itemKey === taskType)
  return item ? item.itemValue : taskType
}

const getConcurrencyLabel = (concurrency?: string) => {
  const map: Record<string, string> = {
    allow: '允许',
    skip: '跳过',
    wait: '等待',
  }
  return concurrency ? map[concurrency] || concurrency : '允许'
}

// 加载统计面板数据
const loadDashboard = async () => {
  try {
    const response = await getTaskDashboard()
    if (response.data) {
      dashboard.totalTasks = response.data.totalTasks || 0
      dashboard.enabledTasks = response.data.enabledTasks || 0
      dashboard.disabledTasks = response.data.disabledTasks || 0
      dashboard.todayExecutions = response.data.todayExecutions || 0
    }
  } catch (error) {
    console.error('获取任务统计面板数据失败:', error)
  }
}

const loadTaskList = async () => {
  loading.value = true
  try {
    const response = await getTaskList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      taskId: searchForm.taskId || undefined,
      taskType: searchForm.taskType || undefined,
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

const handleSearch = () => {
  pagination.page = 1
  loadTaskList()
}

const handleReset = () => {
  searchForm.taskId = ''
  searchForm.taskType = ''
  pagination.page = 1
  loadTaskList()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  loadTaskList()
}

const handleCurrentChange = (current: number) => {
  pagination.page = current
  loadTaskList()
}

const handleAddTask = () => {
  dialogTitle.value = '添加任务'
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEditTask = async (row: Task) => {
  dialogTitle.value = '编辑任务'
  isEdit.value = true
  try {
    const res = await getTaskById(row.id)
    const data = res.data
    if (!data) return
    Object.assign(formData, {
      id: data.id,
      taskId: data.taskId,
      taskType: data.taskType,
      type: data.type,
      expression: data.expression,
      description: data.description,
      group: data.group,
      retryCount: data.retryCount,
      concurrency: data.concurrency || 'allow',
      timeout: data.timeout || 30,
    })
    dialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取任务详情失败')
    console.error('获取任务详情失败:', error)
  }
}

const handleDeleteTask = async (id: string) => {
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

const handleEnableTask = async (row: Task) => {
  try {
    await enableTask(row.id)
    ElMessage.success('启用成功')
    loadTaskList()
  } catch (error) {
    ElMessage.error('启用失败')
    console.error('启用任务失败:', error)
  }
}

const handleDisableTask = async (row: Task) => {
  try {
    await disableTask(row.id)
    ElMessage.success('禁用成功')
    loadTaskList()
  } catch (error) {
    ElMessage.error('禁用失败')
    console.error('禁用任务失败:', error)
  }
}

const handleRunTask = async (_event?: MouseEvent, taskId?: string) => {
  try {
    let taskIds: string[] = []
    if (taskId) {
      taskIds = [taskId]
    } else if (selectedTasks.value.length > 0) {
      taskIds = selectedTasks.value.map((task) => task.taskId)
    } else {
      ElMessage.warning('请选择要执行的任务')
      return
    }

    for (const id of taskIds) {
      await runTask(id)
    }

    ElMessage.success('任务执行成功')
    loadTaskList()
  } catch (error) {
    ElMessage.error('任务执行失败')
    console.error('任务执行失败:', error)
  }
}

const handleViewLog = async (row: Task) => {
  currentTaskId.value = row.taskId
  logPagination.page = 1
  await loadTaskLogs()
  logDialogVisible.value = true
}

const loadTaskLogs = async () => {
  logLoading.value = true
  try {
    const response = await getTaskLogs(currentTaskId.value, {
      page: logPagination.page,
      pageSize: logPagination.pageSize,
    })
    taskLogs.value = response.data?.logs || []
    logPagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('获取任务日志失败')
    console.error('获取任务日志失败:', error)
  } finally {
    logLoading.value = false
  }
}

const handleLogSizeChange = (size: number) => {
  logPagination.pageSize = size
  loadTaskLogs()
}

const handleLogCurrentChange = (current: number) => {
  logPagination.page = current
  loadTaskLogs()
}

const handleSelectionChange = (selection: Task[]) => {
  selectedTasks.value = selection
}

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(formData, {
    id: '',
    taskId: '',
    taskType: 'http',
    type: 'cron',
    expression: '',
    description: '',
    group: '',
    retryCount: 0,
    concurrency: 'allow',
    timeout: 30,
    params: {
      url: '',
      method: 'GET',
      headersJson: '',
      body: '',
      timeout: 30,
      expected_status: 200,
      db_type: 'postgres',
      query: '',
      action: 'refresh',
      cache_keys_str: '',
      pattern: '',
      ttl: 3600,
      command: '',
      args_str: '',
      work_dir: '',
    },
  })
}

const handleDialogClose = () => {
  resetForm()
}

const buildTaskParams = () => {
  const baseParams: Record<string, unknown> = {}

  if (formData.taskType === 'http') {
    let headers: Record<string, string> = {}
    if (formData.params.headersJson) {
      try {
        headers = JSON.parse(formData.params.headersJson)
      } catch {
        headers = {}
      }
    }
    baseParams.url = formData.params.url
    baseParams.method = formData.params.method
    baseParams.headers = headers
    baseParams.body = formData.params.body
    baseParams.timeout = formData.params.timeout
    baseParams.expected_status = formData.params.expected_status
  } else if (formData.taskType === 'database') {
    baseParams.query = formData.params.query
    baseParams.db_type = formData.params.db_type
  } else if (formData.taskType === 'cache') {
    baseParams.action = formData.params.action
    baseParams.cache_keys = formData.params.cache_keys_str
      ? formData.params.cache_keys_str.split(',').map((k: string) => k.trim())
      : []
    baseParams.pattern = formData.params.pattern
    baseParams.ttl = formData.params.ttl
  } else if (formData.taskType === 'script') {
    baseParams.command = formData.params.command
    baseParams.args = formData.params.args_str
      ? formData.params.args_str.split(',').map((a: string) => a.trim())
      : []
    baseParams.work_dir = formData.params.work_dir
    baseParams.timeout = formData.params.timeout
  }

  return baseParams
}

const handleSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()

    const params = buildTaskParams()

    if (isEdit.value) {
      await updateTask(formData.id, {
        taskId: formData.taskId,
        taskType: formData.taskType,
        type: formData.type,
        expression: formData.expression,
        description: formData.description,
        group: formData.group,
        retryCount: formData.retryCount,
        params,
      })
      ElMessage.success('更新成功')
    } else {
      await createTask({
        taskId: formData.taskId,
        taskType: formData.taskType,
        type: formData.type,
        expression: formData.expression,
        description: formData.description,
        group: formData.group,
        retryCount: formData.retryCount,
        params,
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadTaskList()
  } catch (error) {
    if (error === 'Validation failed') {
      return
    }
    ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
    console.error(isEdit.value ? '更新任务失败:' : '创建任务失败:', error)
  }
}

onMounted(async () => {
  await loadTaskTypes()
  loadDashboard()
  loadTaskList()
})
</script>

<style scoped>
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
