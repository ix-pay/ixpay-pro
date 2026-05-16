<template>
  <div class="flex flex-col h-full gap-4">
    <!-- 统计面板 -->
    <div class="grid grid-cols-5 gap-4">
      <div
        class="flex flex-col items-center justify-center p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg"
      >
        <div class="text-2xl font-bold text-blue-600 dark:text-blue-400">
          {{ dashboard.totalEvents }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">事件总数</div>
      </div>
      <div
        class="flex flex-col items-center justify-center p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg"
      >
        <div class="text-2xl font-bold text-yellow-600 dark:text-yellow-400">
          {{ dashboard.pendingEvents }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">待处理</div>
      </div>
      <div
        class="flex flex-col items-center justify-center p-4 bg-green-50 dark:bg-green-900/20 rounded-lg"
      >
        <div class="text-2xl font-bold text-green-600 dark:text-green-400">
          {{ dashboard.successEvents }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">成功数</div>
      </div>
      <div
        class="flex flex-col items-center justify-center p-4 bg-red-50 dark:bg-red-900/20 rounded-lg"
      >
        <div class="text-2xl font-bold text-red-600 dark:text-red-400">
          {{ dashboard.failedEvents }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">失败数</div>
      </div>
      <div
        class="flex flex-col items-center justify-center p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg"
      >
        <div class="text-2xl font-bold text-gray-600 dark:text-gray-400">
          {{ dashboard.deadLetterCount }}
        </div>
        <div class="text-sm text-gray-600 dark:text-gray-400 mt-1">死信队列</div>
      </div>
    </div>

    <!-- Tab 切换 -->
    <div class="flex-1 overflow-hidden bg-white dark:bg-gray-800 rounded-lg shadow-md">
      <el-tabs v-model="activeTab" class="h-full" @tab-click="handleTabClick">
        <!-- 事件列表 -->
        <el-tab-pane label="事件列表" name="events">
          <div class="flex flex-col h-full p-4">
            <!-- 搜索区域 -->
            <div class="flex flex-wrap items-center gap-3 mb-4">
              <el-input v-model="searchForm.name" placeholder="事件名称" style="width: 192px" />
              <el-select v-model="searchForm.status" placeholder="事件状态" style="width: 192px">
                <el-option label="全部" value="" />
                <el-option label="待处理" value="pending" />
                <el-option label="处理中" value="processing" />
                <el-option label="成功" value="success" />
                <el-option label="失败" value="failed" />
                <el-option label="死信" value="dead" />
              </el-select>
              <el-input v-model="searchForm.source" placeholder="事件来源" style="width: 192px" />
              <el-button type="primary" @click="handleSearch">搜索</el-button>
              <el-button @click="handleReset">重置</el-button>
            </div>

            <!-- 事件表格 -->
            <div class="flex-1 overflow-hidden">
              <el-table :data="eventList" stripe class="w-full h-full" v-loading="loading">
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="name" label="事件名称" min-width="150" />
                <el-table-column prop="type" label="事件类型" width="120">
                  <template #default="scope">
                    <el-tag type="primary" size="small">{{ scope.row.type }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="source" label="来源" width="120" />
                <el-table-column prop="status" label="状态" width="100">
                  <template #default="scope">
                    <el-tag :type="getStatusType(scope.row.status)" size="small">
                      {{ getStatusLabel(scope.row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="retryCount" label="重试次数" width="100" />
                <el-table-column prop="createdAt" label="创建时间" width="160" />
                <el-table-column label="操作" width="180" fixed="right">
                  <template #default="scope">
                    <div class="flex flex-wrap gap-2">
                      <el-button type="primary" size="small" @click="handleViewDetail(scope.row)">
                        详情
                      </el-button>
                      <el-button
                        v-if="scope.row.status === 'failed'"
                        type="warning"
                        size="small"
                        @click="handleRetryEvent(scope.row)"
                      >
                        重试
                      </el-button>
                    </div>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <!-- 分页 -->
            <div
              class="flex items-center justify-between mt-4 pt-3 border-t border-gray-200 dark:border-gray-700"
            >
              <span class="text-sm text-gray-600 dark:text-gray-400"
                >共 {{ eventPagination.total }} 条</span
              >
              <el-pagination
                v-model:current-page="eventPagination.page"
                v-model:page-size="eventPagination.pageSize"
                :page-sizes="[10, 20, 50, 100]"
                layout="sizes, prev, pager, next"
                :total="eventPagination.total"
                @size-change="handleEventSizeChange"
                @current-change="handleEventCurrentChange"
                small
              />
            </div>
          </div>
        </el-tab-pane>

        <!-- 死信队列 -->
        <el-tab-pane label="死信队列" name="dead-letters">
          <div class="flex flex-col h-full p-4">
            <div class="flex-1 overflow-hidden">
              <el-table
                :data="deadLetterList"
                stripe
                class="w-full h-full"
                v-loading="deadLetterLoading"
              >
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="eventName" label="事件名称" min-width="150" />
                <el-table-column prop="source" label="来源" width="120" />
                <el-table-column prop="errorMessage" label="错误信息" min-width="200" />
                <el-table-column prop="retryCount" label="重试次数" width="100" />
                <el-table-column prop="deadLetterAt" label="进入死信时间" width="160" />
                <el-table-column label="操作" width="120" fixed="right">
                  <template #default="scope">
                    <el-button
                      type="warning"
                      size="small"
                      @click="handleRetryDeadLetter(scope.row)"
                    >
                      重试
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <!-- 分页 -->
            <div
              class="flex items-center justify-between mt-4 pt-3 border-t border-gray-200 dark:border-gray-700"
            >
              <span class="text-sm text-gray-600 dark:text-gray-400"
                >共 {{ deadLetterPagination.total }} 条</span
              >
              <el-pagination
                v-model:current-page="deadLetterPagination.page"
                v-model:page-size="deadLetterPagination.pageSize"
                :page-sizes="[10, 20, 50, 100]"
                layout="sizes, prev, pager, next"
                :total="deadLetterPagination.total"
                @size-change="handleDeadLetterSizeChange"
                @current-change="handleDeadLetterCurrentChange"
                small
              />
            </div>
          </div>
        </el-tab-pane>

        <!-- 订阅者管理 -->
        <el-tab-pane label="订阅者管理" name="subscribers">
          <div class="flex flex-col h-full p-4">
            <!-- 操作按钮 -->
            <div class="flex flex-wrap items-center gap-3 mb-4">
              <el-button type="primary" @click="handleAddSubscriber">
                <el-icon><Plus /></el-icon>
                添加订阅者
              </el-button>
            </div>

            <!-- 订阅者表格 -->
            <div class="flex-1 overflow-hidden">
              <el-table
                :data="subscriberList"
                stripe
                class="w-full h-full"
                v-loading="subscriberLoading"
              >
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="name" label="订阅者名称" min-width="150" />
                <el-table-column prop="eventType" label="事件类型" width="120" />
                <el-table-column prop="endpoint" label="回调地址" min-width="200" />
                <el-table-column prop="status" label="状态" width="100">
                  <template #default="scope">
                    <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'" size="small">
                      {{ scope.row.status === 'active' ? '活跃' : '未激活' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="lastTriggeredAt" label="最后触发时间" width="160" />
                <el-table-column prop="createdAt" label="创建时间" width="160" />
                <el-table-column label="操作" width="120" fixed="right">
                  <template #default="scope">
                    <el-button
                      type="danger"
                      size="small"
                      @click="handleDeleteSubscriber(scope.row)"
                    >
                      删除
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <!-- 分页 -->
            <div
              class="flex items-center justify-between mt-4 pt-3 border-t border-gray-200 dark:border-gray-700"
            >
              <span class="text-sm text-gray-600 dark:text-gray-400"
                >共 {{ subscriberPagination.total }} 条</span
              >
              <el-pagination
                v-model:current-page="subscriberPagination.page"
                v-model:page-size="subscriberPagination.pageSize"
                :page-sizes="[10, 20, 50, 100]"
                layout="sizes, prev, pager, next"
                :total="subscriberPagination.total"
                @size-change="handleSubscriberSizeChange"
                @current-change="handleSubscriberCurrentChange"
                small
              />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 事件详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="事件详情" width="800px">
      <el-descriptions :column="2" border v-if="currentEvent">
        <el-descriptions-item label="事件ID">{{ currentEvent.id }}</el-descriptions-item>
        <el-descriptions-item label="事件名称">{{ currentEvent.name }}</el-descriptions-item>
        <el-descriptions-item label="事件类型">{{ currentEvent.type }}</el-descriptions-item>
        <el-descriptions-item label="事件来源">{{ currentEvent.source }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentEvent.status)" size="small">
            {{ getStatusLabel(currentEvent.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="重试次数"
          >{{ currentEvent.retryCount }}/{{ currentEvent.maxRetries }}</el-descriptions-item
        >
        <el-descriptions-item label="创建时间">{{ currentEvent.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ currentEvent.updatedAt }}</el-descriptions-item>
        <el-descriptions-item label="处理时间" v-if="currentEvent.processedAt">{{
          currentEvent.processedAt
        }}</el-descriptions-item>
        <el-descriptions-item label="错误信息" v-if="currentEvent.errorMessage">{{
          currentEvent.errorMessage
        }}</el-descriptions-item>
        <el-descriptions-item label="事件负载" :span="2">
          <pre class="bg-gray-100 dark:bg-gray-700 p-2 rounded text-sm">{{
            JSON.stringify(currentEvent.payload, null, 2)
          }}</pre>
        </el-descriptions-item>
      </el-descriptions>

      <div
        v-if="currentEvent && currentEvent.history && currentEvent.history.length > 0"
        class="mt-4"
      >
        <h3 class="text-base font-semibold mb-2">处理历史</h3>
        <el-table :data="currentEvent.history" stripe>
          <el-table-column prop="id" label="记录ID" width="80" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'success' ? 'success' : 'danger'" size="small">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="errorMessage" label="错误信息" />
          <el-table-column prop="retryCount" label="重试次数" width="100" />
          <el-table-column prop="createdAt" label="时间" width="160" />
        </el-table>
      </div>
    </el-dialog>

    <!-- 添加订阅者对话框 -->
    <el-dialog v-model="subscriberDialogVisible" title="添加订阅者" width="600px">
      <el-form
        :model="subscriberForm"
        :rules="subscriberRules"
        ref="subscriberFormRef"
        label-width="100px"
      >
        <el-form-item label="订阅者名称" prop="name">
          <el-input v-model="subscriberForm.name" placeholder="请输入订阅者名称" />
        </el-form-item>
        <el-form-item label="事件类型" prop="eventType">
          <el-input v-model="subscriberForm.eventType" placeholder="请输入事件类型" />
        </el-form-item>
        <el-form-item label="回调地址" prop="endpoint">
          <el-input v-model="subscriberForm.endpoint" placeholder="https://example.com/webhook" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="subscriberDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubscriberSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getEventList,
  getEventById,
  retryEvent,
  getDeadLetters,
  retryDeadLetter,
  getSubscriberList,
  createSubscriber,
  deleteSubscriber,
  getEventStatistics,
} from '@/api/modules/event'
import type { Event, EventDetail, DeadLetterEvent, Subscriber } from '@/types/event'
import type { TabsPaneContext } from 'element-plus'

defineOptions({
  name: 'EventMonitor',
})

// Tab 切换
const activeTab = ref('events')

// 加载状态
const loading = ref(false)
const deadLetterLoading = ref(false)
const subscriberLoading = ref(false)

// 事件列表相关
const eventList = ref<Event[]>([])
const searchForm = reactive({
  name: '',
  status: '',
  source: '',
})
const eventPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

// 死信队列相关
const deadLetterList = ref<DeadLetterEvent[]>([])
const deadLetterPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

// 订阅者相关
const subscriberList = ref<Subscriber[]>([])
const subscriberPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
const subscriberDialogVisible = ref(false)
const subscriberFormRef = ref<FormInstance>()
const subscriberForm = reactive({
  name: '',
  eventType: '',
  endpoint: '',
})
const subscriberRules = reactive({
  name: [{ required: true, message: '请输入订阅者名称', trigger: 'blur' }],
  eventType: [{ required: true, message: '请输入事件类型', trigger: 'blur' }],
  endpoint: [
    { required: true, message: '请输入回调地址', trigger: 'blur' },
    { pattern: /^https?:\/\/.+/, message: '请输入有效的 URL 地址', trigger: 'blur' },
  ],
})

// 事件详情对话框
const detailDialogVisible = ref(false)
const currentEvent = ref<EventDetail | null>(null)

// 统计面板数据
const dashboard = reactive({
  totalEvents: 0,
  pendingEvents: 0,
  successEvents: 0,
  failedEvents: 0,
  deadLetterCount: 0,
})

// 获取状态类型
const getStatusType = (status: string): 'success' | 'warning' | 'danger' | 'info' | 'primary' => {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'primary'> = {
    pending: 'info',
    processing: 'primary',
    success: 'success',
    failed: 'danger',
    dead: 'danger',
  }
  return map[status] || 'info'
}

// 获取状态标签
const getStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    pending: '待处理',
    processing: '处理中',
    success: '成功',
    failed: '失败',
    dead: '死信',
  }
  return map[status] || status
}

// 加载统计面板数据
const loadDashboard = async () => {
  try {
    const response = await getEventStatistics()
    if (response.data) {
      dashboard.totalEvents = response.data.totalEvents || 0
      dashboard.pendingEvents = response.data.pendingEvents || 0
      dashboard.successEvents = response.data.successEvents || 0
      dashboard.failedEvents = response.data.failedEvents || 0
      dashboard.deadLetterCount = response.data.deadLetterCount || 0
    }
  } catch (error) {
    console.error('获取事件统计面板数据失败:', error)
  }
}

// 加载事件列表
const loadEventList = async () => {
  loading.value = true
  try {
    const response = await getEventList({
      page: eventPagination.page,
      pageSize: eventPagination.pageSize,
      name: searchForm.name || undefined,
      status: searchForm.status || undefined,
      source: searchForm.source || undefined,
    })
    eventList.value = response.data?.list || []
    eventPagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('获取事件列表失败')
    console.error('获取事件列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 加载死信队列
const loadDeadLetters = async () => {
  deadLetterLoading.value = true
  try {
    const response = await getDeadLetters({
      page: deadLetterPagination.page,
      pageSize: deadLetterPagination.pageSize,
    })
    deadLetterList.value = response.data?.list || []
    deadLetterPagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('获取死信队列失败')
    console.error('获取死信队列失败:', error)
  } finally {
    deadLetterLoading.value = false
  }
}

// 加载订阅者列表
const loadSubscriberList = async () => {
  subscriberLoading.value = true
  try {
    const response = await getSubscriberList({
      page: subscriberPagination.page,
      pageSize: subscriberPagination.pageSize,
    })
    subscriberList.value = response.data?.list || []
    subscriberPagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('获取订阅者列表失败')
    console.error('获取订阅者列表失败:', error)
  } finally {
    subscriberLoading.value = false
  }
}

// 搜索
const handleSearch = () => {
  eventPagination.page = 1
  loadEventList()
}

// 重置
const handleReset = () => {
  searchForm.name = ''
  searchForm.status = ''
  searchForm.source = ''
  eventPagination.page = 1
  loadEventList()
}

// 事件列表分页
const handleEventSizeChange = (size: number) => {
  eventPagination.pageSize = size
  loadEventList()
}

const handleEventCurrentChange = (current: number) => {
  eventPagination.page = current
  loadEventList()
}

// 死信队列分页
const handleDeadLetterSizeChange = (size: number) => {
  deadLetterPagination.pageSize = size
  loadDeadLetters()
}

const handleDeadLetterCurrentChange = (current: number) => {
  deadLetterPagination.page = current
  loadDeadLetters()
}

// 订阅者分页
const handleSubscriberSizeChange = (size: number) => {
  subscriberPagination.pageSize = size
  loadSubscriberList()
}

const handleSubscriberCurrentChange = (current: number) => {
  subscriberPagination.page = current
  loadSubscriberList()
}

// 查看事件详情
const handleViewDetail = async (row: Event) => {
  try {
    const response = await getEventById(row.id)
    currentEvent.value = response.data || null
    detailDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取事件详情失败')
    console.error('获取事件详情失败:', error)
  }
}

// 重试事件
const handleRetryEvent = async (row: Event) => {
  try {
    await ElMessageBox.confirm('确定要重试该事件吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await retryEvent(row.id)
    ElMessage.success('重试请求已发送')
    loadEventList()
    loadDashboard()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重试失败')
      console.error('重试事件失败:', error)
    }
  }
}

// 重试死信
const handleRetryDeadLetter = async (row: DeadLetterEvent) => {
  try {
    await ElMessageBox.confirm('确定要重试该死信事件吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await retryDeadLetter(row.id)
    ElMessage.success('死信重试请求已发送')
    loadDeadLetters()
    loadDashboard()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('死信重试失败')
      console.error('死信重试失败:', error)
    }
  }
}

// 添加订阅者
const handleAddSubscriber = () => {
  subscriberForm.name = ''
  subscriberForm.eventType = ''
  subscriberForm.endpoint = ''
  subscriberDialogVisible.value = true
}

// 提交订阅者
const handleSubscriberSubmit = async () => {
  if (!subscriberFormRef.value) return
  try {
    await subscriberFormRef.value.validate()
    await createSubscriber({
      name: subscriberForm.name,
      eventType: subscriberForm.eventType,
      endpoint: subscriberForm.endpoint,
    })
    ElMessage.success('创建成功')
    subscriberDialogVisible.value = false
    loadSubscriberList()
    loadDashboard()
  } catch (error) {
    if (error !== 'Validation failed') {
      ElMessage.error('创建失败')
      console.error('创建订阅者失败:', error)
    }
  }
}

// 删除订阅者
const handleDeleteSubscriber = async (row: Subscriber) => {
  try {
    await ElMessageBox.confirm('确定要删除该订阅者吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteSubscriber(row.id)
    ElMessage.success('删除成功')
    loadSubscriberList()
    loadDashboard()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除订阅者失败:', error)
    }
  }
}

// Tab 切换时加载数据
const handleTabClick = (pane: TabsPaneContext) => {
  const tabName = String(pane.paneName)
  if (tabName === 'dead-letters' && deadLetterList.value.length === 0) {
    loadDeadLetters()
  } else if (tabName === 'subscribers' && subscriberList.value.length === 0) {
    loadSubscriberList()
  }
}

onMounted(() => {
  loadDashboard()
  loadEventList()
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

pre {
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>
