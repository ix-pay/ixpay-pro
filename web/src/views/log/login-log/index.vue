<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 顶部操作栏 -->
    <div class="flex flex-col gap-3 p-4 border-b">
      <!-- 第一行：搜索条件 -->
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.userName"
          placeholder="请输入用户名"
          clearable
          style="width: 192px"
        />
        <el-select
          v-model="searchForm.status"
          placeholder="选择状态"
          clearable
          style="width: 192px"
        >
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
        <el-button type="danger" @click="handleClearLog">
          <el-icon>
            <Delete />
          </el-icon>
          清空日志
        </el-button>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="loginLogList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="userName" label="用户名" width="120" />
        <el-table-column prop="ip" label="IP 地址" width="130" />
        <el-table-column prop="location" label="登录地点" min-width="150" />
        <el-table-column prop="browser" label="浏览器" width="120" show-overflow-tooltip />
        <el-table-column prop="os" label="操作系统" width="120" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="70">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="消息" width="150" show-overflow-tooltip />
        <el-table-column prop="loginTime" label="登录时间" width="160" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button type="primary" size="small" @click="handleViewDetail(scope.row)">
                详情
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Delete, Refresh } from '@element-plus/icons-vue'
import { getLoginLogList, clearLoginLogs } from '@/api/modules/login-log'

defineOptions({
  name: 'LoginLogManagement',
})

interface LoginLog {
  id: number
  userId: number
  userName: string
  ip: string
  location: string
  browser: string
  os: string
  status: number
  message: string
  loginTime: string
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
  status: undefined,
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
      ...(searchForm.status !== undefined ? { status: searchForm.status } : {}),
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
  searchForm.status = undefined
  loadLoginLogList()
}

// 清空日志
const handleClearLog = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有登录日志吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await clearLoginLogs({
      startTime: searchForm.startTime,
      endTime: searchForm.endTime,
    })
    ElMessage.success('清空日志成功')
    loadLoginLogList()
  } catch (error: unknown) {
    if (error !== 'cancel') {
      ElMessage.error('清空日志失败')
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
const handleViewDetail = (row: LoginLog) => {
  ElMessage.info(`查看日志详情：${row.userName}`)
}

onMounted(() => {
  loadLoginLogList()
})
</script>
