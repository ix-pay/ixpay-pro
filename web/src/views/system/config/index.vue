<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <div class="flex flex-col gap-3 p-4 border-b">
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.keyword"
          placeholder="请输入配置名称或键"
          clearable
          style="width: 192px"
          @keyup.enter="loadConfigList"
        >
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>
        <el-select
          v-model="searchForm.status"
          placeholder="选择状态"
          clearable
          style="width: 192px"
        >
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" @click="loadConfigList">
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

      <div class="flex flex-wrap items-center gap-2">
        <el-button type="primary" v-auth-btn="'system:config:add'" @click="handleAddConfig">
          <el-icon><Plus /></el-icon>
          添加配置
        </el-button>
      </div>
    </div>

    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500">
            <Setting />
          </el-icon>
          配置总数：<span class="font-medium">{{ pagination.total }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <SuccessFilled />
          </el-icon>
          启用：<span class="font-medium">{{
            configList.filter((c) => c.status === 1).length
          }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-orange-500">
            <CircleClose />
          </el-icon>
          禁用：<span class="font-medium">{{
            configList.filter((c) => c.status === 0).length
          }}</span>
        </span>
      </div>
    </div>

    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="configList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="configKey" label="配置键" width="180" />
        <el-table-column prop="configValue" label="配置值" min-width="200" />
        <el-table-column prop="configType" label="类型" width="100" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button
                v-auth-btn="'system:config:edit'"
                type="primary"
                size="small"
                @click="handleEditConfig(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:config:delete'"
                type="danger"
                size="small"
                @click="handleDeleteConfig(scope.row.id)"
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="configFormRef" :model="configForm" :rules="formRules" label-width="100px">
        <el-form-item label="配置键" prop="configKey">
          <el-input
            v-model="configForm.configKey"
            placeholder="请输入配置键"
            :disabled="!!configForm.id"
          />
        </el-form-item>
        <el-form-item label="配置值" prop="configValue">
          <el-input v-model="configForm.configValue" placeholder="请输入配置值" />
        </el-form-item>
        <el-form-item label="类型" prop="configType">
          <el-select v-model="configForm.configType" placeholder="请选择类型" class="w-full">
            <el-option label="文本" value="1" />
            <el-option label="数字" value="2" />
            <el-option label="布尔" value="3" />
            <el-option label="JSON" value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="configForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入配置描述"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="configForm.status"
            :active-value="1"
            :inactive-value="0"
            active-color="#13ce66"
            inactive-color="#ff4949"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus, Search, Refresh, Setting, SuccessFilled, CircleClose } from '@element-plus/icons-vue'
import { getConfigList, createConfig, updateConfig, deleteConfig } from '@/api/modules/config'
import { formatDate } from '@/utils/format'

defineOptions({
  name: 'ConfigManagement',
})

interface Config {
  id: string
  configKey: string
  configValue: string
  configType: string
  description: string
  status: number
  createdAt: string
}

const configList = ref<Config[]>([])
const loading = ref(false)
const isLoading = ref(false)

const searchForm = reactive({
  keyword: '',
  status: undefined as number | undefined,
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const dialogVisible = ref(false)
const dialogTitle = ref('')
const configFormRef = ref<FormInstance>()

const configForm = reactive({
  id: '',
  configKey: '',
  configValue: '',
  configType: 'string',
  description: '',
  status: 1,
})

const formRules = reactive({
  configKey: [{ required: true, message: '请输入配置键', trigger: 'blur' }],
  configValue: [{ required: true, message: '请输入配置值', trigger: 'blur' }],
  configType: [{ required: true, message: '请选择类型', trigger: 'change' }],
})

const loadConfigList = async () => {
  if (isLoading.value) return

  isLoading.value = true
  loading.value = true
  try {
    const response = await getConfigList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.keyword ? { keyword: searchForm.keyword } : {}),
      ...(searchForm.status !== undefined ? { status: searchForm.status } : {}),
    })
    const pageData = response.data as Record<string, unknown>
    configList.value = (pageData?.list as Config[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    console.error('获取配置列表失败:', error)
    configList.value = []
    pagination.total = 0
  } finally {
    loading.value = false
    isLoading.value = false
  }
}

const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.status = undefined
  pagination.page = 1
  loadConfigList()
}

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadConfigList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadConfigList()
}

const handleAddConfig = () => {
  dialogTitle.value = '添加配置'
  Object.assign(configForm, {
    id: '',
    configKey: '',
    configValue: '',
    configType: '1',
    description: '',
    status: 1,
  })
  dialogVisible.value = true
}

const handleEditConfig = (config: Config) => {
  dialogTitle.value = '编辑配置'
  Object.assign(configForm, config)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    await configFormRef.value?.validate()
    if (configForm.id) {
      await updateConfig(configForm.id, {
        configKey: configForm.configKey,
        configValue: configForm.configValue,
        configType: configForm.configType,
        description: configForm.description,
        status: configForm.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createConfig({
        configKey: configForm.configKey,
        configValue: configForm.configValue,
        configType: configForm.configType,
        description: configForm.description,
        status: configForm.status,
      })
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadConfigList()
  } catch {
    // 验证失败
  }
}

const handleDeleteConfig = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除该配置吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteConfig(id)
    ElMessage.success('删除成功')
    loadConfigList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除配置失败:', error)
    }
  }
}

onMounted(() => {
  loadConfigList()
})
</script>

<style scoped>
.flex-1 {
  min-height: 0;
}
</style>
