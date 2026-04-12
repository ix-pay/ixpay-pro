<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-2">
        <el-input
          v-model="searchForm.username"
          placeholder="请输入用户名"
          clearable
          size="small"
          class="w-48"
        />
        <el-select
          v-model="searchForm.status"
          placeholder="选择状态"
          clearable
          size="small"
          class="w-32"
        >
          <el-option label="成功" :value="1" />
          <el-option label="失败" :value="0" />
        </el-select>
        <el-button type="primary" size="small" @click="loadLoginLogList">
          <el-icon><Search /></el-icon>
          搜索
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
        <el-table-column prop="username" label="用户名" width="120" />
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
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getLoginLogList } from '@/api/modules/login-log'

defineOptions({
  name: 'LoginLogManagement',
})

interface LoginLog {
  id: number
  userId: number
  username: string
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
  username: '',
  status: undefined,
})

// 加载登录日志列表
const loadLoginLogList = async () => {
  loading.value = true
  try {
    const response = await getLoginLogList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.username ? { username: searchForm.username } : {}),
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

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadLoginLogList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadLoginLogList()
}

onMounted(() => {
  loadLoginLogList()
})
</script>
