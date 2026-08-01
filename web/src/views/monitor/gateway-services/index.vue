<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <div class="flex flex-col gap-3 p-4 border-b">
      <div class="flex flex-wrap items-center gap-2">
        <el-button type="primary" @click="handleRefresh">
          <el-icon>
            <Refresh />
          </el-icon>
          刷新服务列表
        </el-button>
      </div>
    </div>

    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500">
            <Connection />
          </el-icon>
          服务总数：<span class="font-medium">{{ gatewayServiceList.length }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <SuccessFilled />
          </el-icon>
          健康服务：<span class="font-medium">{{
            gatewayServiceList.filter((s) => s.status === 'healthy').length
          }}</span>
        </span>
      </div>
    </div>

    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="gatewayServiceList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="id" label="实例ID" min-width="180" />
        <el-table-column prop="name" label="服务名称" width="150" />
        <el-table-column prop="address" label="地址" width="150" />
        <el-table-column prop="port" label="端口" width="90" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'healthy' ? 'success' : 'danger'">
              {{ scope.row.status === 'healthy' ? '健康' : '不健康' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="activeConnections" label="活跃连接" width="110" align="center" />
        <el-table-column prop="metadata" label="元数据" min-width="200">
          <template #default="scope">
            <div class="text-xs space-y-1">
              <div v-if="scope.row.metadata.version">
                <span class="text-gray-500">版本:</span> {{ scope.row.metadata.version }}
              </div>
              <div v-if="scope.row.metadata.node_role">
                <span class="text-gray-500">角色:</span> {{ scope.row.metadata.node_role }}
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="lastSeen" label="最后心跳" width="180">
          <template #default="scope">
            <span class="text-xs">{{ formatDate(scope.row.lastSeen) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Connection, SuccessFilled } from '@element-plus/icons-vue'
import { getGatewayServices, type GatewayService } from '@/api/modules/gateway'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'GatewayServices' })

const gatewayServiceList = ref<GatewayService[]>([])
const loading = ref(false)

const loadGatewayServices = async () => {
  loading.value = true
  try {
    const res = await getGatewayServices()
    gatewayServiceList.value = res.data || []
  } catch {
    ElMessage.error('获取网关服务列表失败')
    gatewayServiceList.value = []
  } finally {
    loading.value = false
  }
}

const handleRefresh = () => {
  loadGatewayServices()
}

onMounted(() => {
  loadGatewayServices()
})
</script>

<style scoped>
/* 网关服务页面样式 */
</style>
