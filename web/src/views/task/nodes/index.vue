<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md p-4 transition-colors duration-300"
  >
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-lg font-semibold">节点管理</h2>
      <div class="flex gap-2">
        <el-button type="primary" @click="loadNodes" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
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

    <el-table :data="filteredNodes" stripe v-loading="loading">
      <el-table-column prop="nodeId" label="节点ID" width="180" />
      <el-table-column prop="role" label="角色" width="120">
        <template #default="scope">
          <el-tag :type="getRoleTag(scope.row.role)">
            {{ getRoleLabel(scope.row.role) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'online' ? 'success' : 'danger'">
            {{ scope.row.status === 'online' ? '在线' : '离线' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="ipAddress" label="IP地址" width="150" />
      <el-table-column prop="port" label="端口" width="80" />
      <el-table-column prop="runningTasks" label="运行任务数" width="120" />
      <el-table-column prop="maxConcurrent" label="最大并发" width="100" />
      <el-table-column prop="lastHeartbeat" label="最后心跳" width="180" />
      <el-table-column prop="registeredAt" label="注册时间" width="180" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="scope">
          <el-button type="primary" size="small" @click="handleViewDetail(scope.row)">
            详情
          </el-button>
          <el-button
            v-if="scope.row.status === 'online'"
            type="danger"
            size="small"
            @click="handleOfflineNode(scope.row)"
          >
            下线
          </el-button>
        </template>
      </el-table-column>
    </el-table>

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

const loadNodes = async () => {
  loading.value = true
  try {
    const [nodesRes, statsRes] = await Promise.all([getNodeList(), getNodeStatistics()])

    if (nodesRes.code === 0) {
      nodes.value = nodesRes.data || []
    }

    if (statsRes.code === 0) {
      statistics.value = statsRes.data || { total: 0, online: 0, offline: 0 }
    }
  } catch (error) {
    ElMessage.error('获取节点列表失败')
    console.error('获取节点列表失败:', error)
  } finally {
    loading.value = false
  }
}

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
</style>
