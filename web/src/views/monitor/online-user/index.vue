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
        <el-button type="primary" size="small" @click="loadOnlineUserList">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="onlineUserList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column prop="ip" label="IP 地址" width="130" />
        <el-table-column prop="location" label="登录地点" min-width="150" />
        <el-table-column prop="browser" label="浏览器" width="120" show-overflow-tooltip />
        <el-table-column prop="os" label="操作系统" width="120" show-overflow-tooltip />
        <el-table-column prop="loginTime" label="登录时间" width="160" />
        <el-table-column prop="lastActiveTime" label="最后活跃" width="160" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="scope">
            <el-button size="small" type="danger" link @click="handleForceLogout(scope.row.token)">
              强制下线
            </el-button>
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
import { Search } from '@element-plus/icons-vue'
import { getOnlineUserList, forceLogout } from '@/api/modules/online-user'

defineOptions({
  name: 'OnlineUserManagement',
})

interface OnlineUser {
  id: string
  userId: number
  username: string
  nickname: string
  ip: string
  location: string
  browser: string
  os: string
  loginTime: string
  lastActiveTime: string
  token: string
}

const onlineUserList = ref<OnlineUser[]>([])
const loading = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
const searchForm = reactive({
  username: '',
})

// 加载在线用户列表
const loadOnlineUserList = async () => {
  loading.value = true
  try {
    const response = await getOnlineUserList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.username ? { username: searchForm.username } : {}),
    })
    const pageData = response.data as Record<string, unknown>
    onlineUserList.value = (pageData?.list as OnlineUser[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    ElMessage.error('获取在线用户列表失败')
    console.error('获取在线用户列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadOnlineUserList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadOnlineUserList()
}

// 强制下线处理
const handleForceLogout = async (token: string) => {
  try {
    await ElMessageBox.confirm('确定要强制该用户下线吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await forceLogout(token)
    ElMessage.success('强制下线成功')
    loadOnlineUserList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('强制下线失败')
      console.error('强制下线失败:', error)
    }
  }
}

onMounted(() => {
  loadOnlineUserList()
})
</script>
