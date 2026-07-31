<template>
  <div class="flex flex-col h-full bg-[var(--bg-secondary)] rounded-lg shadow-md transition-colors duration-300">
    <!-- 顶部操作栏 -->
    <div class="flex flex-col gap-3 p-4 border-b">
      <!-- 第一行：搜索条件 -->
      <div class="flex flex-wrap items-center gap-3">
        <el-input v-model="searchForm.userName" placeholder="请输入用户名" clearable style="width: 192px" />
        <el-select v-model="searchForm.result" placeholder="选择状态" clearable style="width: 192px">
          <el-option label="成功" :value="1" />
          <el-option label="失败" :value="0" />
        </el-select>
        <el-button type="primary" @click="loadLoginLogList">
          <el-icon>
            <Search />
          </el-icon>
          搜索
        </el-button>
        <el-button @click="resetSearch">
          <el-icon>
            <Refresh />
          </el-icon>
          重置
        </el-button>
      </div>

      <!-- 第二行：功能按钮 -->
      <div class="flex flex-wrap items-center gap-2">
        <el-button v-auth-btn="'log:login:clear'" type="danger" @click="handleClearLog">
          <el-icon>
            <Delete />
          </el-icon>
          清空日志
        </el-button>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table v-loading="loading" :data="loginLogList" stripe class="w-full h-full" :height="'100%'">
        <el-table-column prop="userName" label="用户名" width="120" />
        <el-table-column prop="ip" label="IP 地址" width="130" />
        <el-table-column prop="place" label="登录地点" min-width="150" />
        <el-table-column prop="result" label="状态" width="70">
          <template #default="scope">
            <el-tag :type="scope.row.result === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.result === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="登录时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button v-auth-btn="'log:login:view'" type="primary" size="small" @click="handleViewDetail(scope.row)">
                详情
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页区域 - 紧凑布局 -->
    <div class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700">
      <span class="text-sm text-gray-600 dark:text-gray-400">共 {{ pagination.total }} 条</span>
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]" layout="sizes, prev, pager, next" :total="pagination.total"
        @size-change="handleSizeChange" @current-change="handleCurrentChange" small />
    </div>

    <el-dialog v-model="detailVisible" title="登录日志详情" width="600px">
      <el-descriptions v-if="detailData" :column="1" border>
        <el-descriptions-item label="用户名">{{ detailData.userName }}</el-descriptions-item>
        <el-descriptions-item label="用户 ID">{{ detailData.userId }}</el-descriptions-item>
        <el-descriptions-item label="IP 地址">{{ detailData.ip }}</el-descriptions-item>
        <el-descriptions-item label="登录地点">{{ detailData.place }}</el-descriptions-item>
        <el-descriptions-item label="设备信息">{{ detailData.device || '-' }}</el-descriptions-item>
        <el-descriptions-item label="浏览器">{{ detailData.browser || '-' }}</el-descriptions-item>
        <el-descriptions-item label="操作系统">{{ detailData.os || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detailData.result === 1 ? 'success' : 'danger'" size="small">
            {{ detailData.result === 1 ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="登录时间">{{
          formatDate(detailData.createdAt)
          }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="clearDialogVisible" title="清空日志" width="500px">
      <el-form :model="clearForm" label-width="100px">
        <el-form-item label="开始日期" required>
          <el-date-picker v-model="clearForm.startTime" type="date" placeholder="选择开始日期" value-format="YYYY-MM-DD"
            class="w-full" />
        </el-form-item>
        <el-form-item label="结束日期" required>
          <el-date-picker v-model="clearForm.endTime" type="date" placeholder="选择结束日期" value-format="YYYY-MM-DD"
            class="w-full" />
        </el-form-item>
      </el-form>
      <el-alert title="此操作将清空指定时间范围内的所有登录日志，请谨慎操作！" type="warning" :closable="false" class="mb-5" />
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
import { Search, Delete, Refresh } from '@element-plus/icons-vue'
import {
  getLoginLogList,
  clearLoginLogs,
  getLoginLogById,
  type LoginLogDetail,
} from '@/api/modules/login-log'
import { formatDate } from '@/utils/format'

defineOptions({
  name: 'LoginLogManagement',
})

interface LoginLog {
  id: number
  userName: string
  ip: string
  place: string
  result: number
  createdAt: string
}

const loginLogList = ref<LoginLog[]>([])
const loading = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
const searchForm = reactive({
  userName: '',
  result: undefined,
})

const detailVisible = ref(false)
const detailData = ref<LoginLogDetail | null>(null)
const clearDialogVisible = ref(false)
const clearForm = reactive({
  startTime: '',
  endTime: '',
})

// 加载登录日志列表
const loadLoginLogList = async () => {
  loading.value = true
  try {
    const response = await getLoginLogList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.userName ? { userName: searchForm.userName } : {}),
      ...(searchForm.result !== undefined ? { result: searchForm.result } : {}),
    })
    const pageData = response.data as Record<string, unknown>
    loginLogList.value = (pageData?.list as LoginLog[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    ElMessage.error('获取登录日志列表失败')
    console.error('获取登录日志列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  searchForm.userName = ''
  searchForm.result = undefined
  loadLoginLogList()
}

// 清空日志
const handleClearLog = () => {
  clearForm.startTime = ''
  clearForm.endTime = ''
  clearDialogVisible.value = true
}

// 提交清空日志
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

    await clearLoginLogs({
      startTime: clearForm.startTime + ' 00:00:00',
      endTime: clearForm.endTime + ' 23:59:59',
    })
    ElMessage.success('清空成功')
    clearDialogVisible.value = false
    loadLoginLogList()
  } catch (error: unknown) {
    if (error !== 'cancel') {
      ElMessage.error('清空失败')
      console.error('清空日志失败:', error)
    }
  }
}

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadLoginLogList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadLoginLogList()
}

// 查看详情
const handleViewDetail = async (row: LoginLog) => {
  try {
    const response = await getLoginLogById(row.id)
    if (response.data) {
      detailData.value = response.data
      detailVisible.value = true
    }
  } catch (error) {
    ElMessage.error('获取登录日志详情失败')
    console.error('获取登录日志详情失败:', error)
  }
}

onMounted(() => {
  loadLoginLogList()
})
</script>
