<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md p-4 transition-colors duration-300"
  >
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-lg font-semibold">节点管理</h2>
      <div class="flex items-center gap-3">
        <span class="update-time" v-if="lastUpdateTime">最后更新：{{ lastUpdateTime }}</span>
        <el-select
          v-model="refreshInterval"
          size="small"
          class="interval-select"
          @change="changeRefreshInterval"
        >
          <el-option label="5 秒" :value="5000" />
          <el-option label="10 秒" :value="10000" />
          <el-option label="30 秒" :value="30000" />
        </el-select>
        <el-switch
          v-model="autoRefreshEnabled"
          active-text="自动刷新"
          class="refresh-switch"
          @change="toggleAutoRefresh"
        />
        <el-button type="primary" @click="refreshData" :loading="loading" circle>
          <el-icon><Refresh /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="flex gap-4 mb-4">
      <el-card class="flex-1" shadow="hover">
        <div class="flex items-center justify-center">
          <div class="text-3xl font-bold text-primary">{{ statistics.total }}</div>
          <div class="ml-3 text-gray-500 dark:text-gray-400">总节点数</div>
        </div>
      </el-card>
      <el-card class="flex-1" shadow="hover">
        <div class="flex items-center justify-center">
          <div class="text-3xl font-bold text-success">{{ statistics.online }}</div>
          <div class="ml-3 text-gray-500 dark:text-gray-400">在线节点</div>
        </div>
      </el-card>
      <el-card class="flex-1" shadow="hover">
        <div class="flex items-center justify-center">
          <div class="text-3xl font-bold text-danger">{{ statistics.offline }}</div>
          <div class="ml-3 text-gray-500 dark:text-gray-400">离线节点</div>
        </div>
      </el-card>
    </div>

    <div class="flex gap-4 mb-4">
      <el-select
        v-model="filterRole"
        placeholder="按角色筛选"
        clearable
        class="w-40"
        @change="handleFilter"
      >
        <el-option label="全部" value="" />
        <el-option label="API节点" value="api" />
        <el-option label="任务节点" value="task" />
        <el-option label="全功能节点" value="all" />
      </el-select>
      <el-select
        v-model="filterStatus"
        placeholder="按状态筛选"
        clearable
        class="w-40"
        @change="handleFilter"
      >
        <el-option label="全部" value="" />
        <el-option label="在线" value="online" />
        <el-option label="离线" value="offline" />
      </el-select>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <div
        v-for="node in filteredNodes"
        :key="node.nodeId"
        class="flex flex-col p-4 bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 hover:shadow-lg transition-shadow duration-300"
      >
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <el-tag :type="node.status === 'online' ? 'success' : 'danger'" size="small">
              {{ node.status === 'online' ? '在线' : '离线' }}
            </el-tag>
            <el-tag :type="getRoleTag(node.role)" size="small">
              {{ getRoleLabel(node.role) }}
            </el-tag>
          </div>
        </div>

        <div class="space-y-2 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">节点 ID</span>
            <span class="text-gray-900 dark:text-white font-mono">{{ node.nodeId }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">IP 地址</span>
            <span class="text-gray-900 dark:text-white">{{ node.ipAddress }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">端口</span>
            <span class="text-gray-900 dark:text-white">{{ node.port }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">运行任务数</span>
            <span class="text-gray-900 dark:text-white font-semibold">{{ node.runningTasks }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">最大并发</span>
            <span class="text-gray-900 dark:text-white font-semibold">{{ node.maxConcurrent }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">最后心跳</span>
            <span class="text-gray-900 dark:text-white text-xs">{{ node.lastHeartbeat }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">注册时间</span>
            <span class="text-gray-900 dark:text-white text-xs">{{ node.registeredAt }}</span>
          </div>
        </div>

        <div class="flex gap-2 mt-4 pt-3 border-t border-gray-200 dark:border-gray-700">
          <el-button
            type="primary"
            size="small"
            class="flex-1"
            @click="handleViewDetail(node)"
          >
            详情
          </el-button>
          <el-button
            v-if="node.status === 'online'"
            type="danger"
            size="small"
            class="flex-1"
            @click="handleOfflineNode(node)"
          >
            下线
          </el-button>
        </div>
      </div>
    </div>

    <el-alert
      v-if="filteredNodes.length === 0 && !loading"
      title="暂无节点"
      type="info"
      description="当前没有活跃的节点，请检查服务是否正常运行。"
      show-icon
      class="mt-4"
    />

    <el-dialog v-model="detailDialogVisible" title="节点详情" width="600px">
      <el-descriptions :column="2" border v-if="currentNode">
        <el-descriptions-item label="节点ID">{{ currentNode.nodeId }}</el-descriptions-item>
        <el-descriptions-item label="角色">{{
          getRoleLabel(currentNode.role)
        }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentNode.status === 'online' ? 'success' : 'danger'">
            {{ currentNode.status === 'online' ? '在线' : '离线' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentNode.ipAddress }}</el-descriptions-item>
        <el-descriptions-item label="端口">{{ currentNode.port }}</el-descriptions-item>
        <el-descriptions-item label="运行任务数">{{
          currentNode.runningTasks
        }}</el-descriptions-item>
        <el-descriptions-item label="最大并发">{{
          currentNode.maxConcurrent
        }}</el-descriptions-item>
        <el-descriptions-item label="最后心跳">{{
          currentNode.lastHeartbeat
        }}</el-descriptions-item>
        <el-descriptions-item label="注册时间" :span="2">{{
          currentNode.registeredAt
        }}</el-descriptions-item>
        <el-descriptions-item label="启动时间" :span="2">{{
          currentNode.startedAt
        }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { getNodeList, offlineNode as offlineNodeApi, getNodeStatistics } from '@/api/modules/node'
import type { NodeInfo, NodeStatistics } from '@/api/modules/node'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

defineOptions({
  name: 'TaskNodes',
})

const loading = ref(false)
const nodes = ref<NodeInfo[]>([])
const statistics = ref<NodeStatistics>({
  total: 0,
  online: 0,
  offline: 0,
})

const filterRole = ref('')
const filterStatus = ref('')

// 加载节点列表函数（需要在 useAutoRefresh 之前定义）
const loadNodes = async () => {
  loading.value = true
  try {
    const res = await getNodeList()
    if (res.code === 0) {
      nodes.value = res.data || []
      statistics.value.total = nodes.value.length
      statistics.value.online = nodes.value.filter((n) => n.status === 'online').length
      statistics.value.offline = nodes.value.filter((n) => n.status === 'offline').length
    }
  } catch (error) {
    ElMessage.error('获取节点列表失败')
    console.error('获取节点列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 定时刷新
const {
  autoRefresh: autoRefreshEnabled,
  refreshInterval,
  lastUpdateTime,
  changeRefreshInterval,
  toggleAutoRefresh,
  refreshData,
} = useAutoRefresh(loadNodes, true, 5000)

const filteredNodes = computed(() => {
  return nodes.value.filter((node) => {
    if (filterRole.value && node.role !== filterRole.value) {
      return false
    }
    if (filterStatus.value && node.status !== filterStatus.value) {
      return false
    }
    return true
  })
})

const detailDialogVisible = ref(false)
const currentNode = ref<NodeInfo | null>(null)

const handleFilter = () => {}

const handleViewDetail = (node: NodeInfo) => {
  currentNode.value = node
  detailDialogVisible.value = true
}

const handleOfflineNode = async (node: NodeInfo) => {
  try {
    await ElMessageBox.confirm(`确定要将节点 ${node.nodeId} 下线吗？`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    const res = await offlineNodeApi(node.nodeId)
    if (res.code === 0) {
      ElMessage.success('节点下线成功')
      loadNodes()
    } else {
      ElMessage.error(res.msg || '节点下线失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('节点下线失败')
      console.error('节点下线失败:', error)
    }
  }
}

const getRoleLabel = (role: string) => {
  const map: Record<string, string> = {
    api: 'API节点',
    task: '任务节点',
    all: '全功能节点',
  }
  return map[role] || role
}

const getRoleTag = (role: string): 'primary' | 'success' | 'warning' | 'info' | 'danger' => {
  const map: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
    api: 'primary',
    task: 'success',
    all: 'warning',
  }
  return map[role] || 'info'
}

onMounted(() => {
  loadNodes()
})
</script>

<style scoped>
:deep(.el-table) {
  --el-table-bg-color: transparent;
}

.update-time {
  font-size: 12px;
  color: #909399;
}

.interval-select {
  width: 90px;
}

.refresh-switch {
  :deep(.el-switch__label) {
    font-size: 12px;
  }
}
</style>
