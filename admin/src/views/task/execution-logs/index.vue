<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 顶部操作栏 -->
    <div class="flex flex-col gap-3 p-4 border-b">
      <!-- 第一行：搜索条件 -->
      <div class="flex flex-wrap items-center gap-3">
        <el-select
          v-model="searchForm.taskId"
          placeholder="选择任务"
          clearable
          filterable
          style="width: 220px"
        >
          <el-option
            v-for="task in taskOptions"
            :key="task.id"
            :label="task.taskId"
            :value="task.taskId"
          />
        </el-select>
        <el-select
          v-model="searchForm.result"
          placeholder="执行结果"
          clearable
          style="width: 150px"
        >
          <el-option label="全部" value="" />
          <el-option label="成功" value="success" />
          <el-option label="失败" value="failed" />
        </el-select>
        <el-date-picker
          v-model="searchForm.dateRange"
          type="daterange"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 260px"
        />
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
    </div>

    <!-- 统计信息 -->
    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500">
            <Document />
          </el-icon>
          日志总数：<span class="font-medium">{{ pagination.total }}</span>
        </span>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table v-loading="loading" :data="logList" stripe class="w-full h-full" :height="'100%'">
        <el-table-column prop="taskId" label="任务ID" width="150" />
        <el-table-column prop="taskName" label="任务名称" width="150" />
        <el-table-column prop="group" label="分组" width="120" />
        <el-table-column prop="result" label="执行结果" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.result === 'success' ? 'success' : 'danger'" size="small">
              {{ scope.row.result === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="triggerType" label="触发类型" width="120" />
        <el-table-column prop="duration" label="耗时 (ms)" width="100" />
        <el-table-column prop="retryCount" label="重试次数" width="100" />
        <el-table-column prop="errorInfo" label="错误信息" min-width="200">
          <template #default="scope">
            <el-tooltip v-if="scope.row.errorInfo" :content="scope.row.errorInfo" placement="top">
              <span class="truncate block max-w-[200px]">{{ scope.row.errorInfo }}</span>
            </el-tooltip>
            <span v-else class="text-gray-400">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="executeAt" label="执行时间" width="180" />
        <el-table-column prop="cronExpr" label="Cron表达式" width="150" />
      </el-table>
    </div>

    <!-- 分页区域 -->
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh, Document } from '@element-plus/icons-vue'
import { searchTaskLogs, getTaskList } from '@/api/modules/task'
import type { TaskLog, Task } from '@/api/modules/task'

defineOptions({
  name: 'TaskExecutionLogs',
})

const loading = ref(false)
const logList = ref<TaskLog[]>([])
const taskOptions = ref<Task[]>([])

const searchForm = reactive({
  taskId: '',
  result: '',
  dateRange: [] as string[],
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const loadLogs = async () => {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      pageSize: pagination.pageSize,
    }

    if (searchForm.taskId) {
      params.taskId = searchForm.taskId
    }
    if (searchForm.result) {
      params.result = searchForm.result
    }
    if (searchForm.dateRange && searchForm.dateRange.length === 2) {
      params.startDate = `${searchForm.dateRange[0]}T00:00:00Z`
      params.endDate = `${searchForm.dateRange[1]}T23:59:59Z`
    }

    const res = await searchTaskLogs(params)
    logList.value = res.data?.logs || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    ElMessage.error('获取日志失败')
    console.error('获取日志失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadLogs()
}

const handleReset = () => {
  searchForm.taskId = ''
  searchForm.result = ''
  searchForm.dateRange = []
  pagination.page = 1
  loadLogs()
}

const loadTasks = async () => {
  try {
    const res = await getTaskList({ page: 1, pageSize: 1000 })
    taskOptions.value = res.data?.list || []
  } catch (error) {
    console.error('获取任务列表失败:', error)
  }
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  loadLogs()
}

const handleCurrentChange = (current: number) => {
  pagination.page = current
  loadLogs()
}

onMounted(() => {
  loadTasks()
  loadLogs()
})
</script>

<style scoped></style>
