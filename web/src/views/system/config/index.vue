<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <span class="text-sm text-gray-600 dark:text-gray-400">系统配置</span>
      <el-button type="primary" size="small" @click="handleAddConfig">
        <el-icon><Plus /></el-icon>
        添加配置
      </el-button>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="configList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="name" label="配置名称" width="160" />
        <el-table-column prop="key" label="配置键" width="180" />
        <el-table-column prop="value" label="配置值" min-width="200" />
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <div class="flex gap-1">
              <el-button size="small" type="primary" link @click="handleEditConfig(scope.row)">
                编辑
              </el-button>
              <el-button size="small" type="danger" link @click="handleDeleteConfig(scope.row.id)">
                删除
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

    <!-- 配置表单对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="configFormRef" :model="configForm" :rules="formRules" label-width="100px">
        <el-form-item label="配置名称" prop="name">
          <el-input v-model="configForm.name" placeholder="请输入配置名称" />
        </el-form-item>
        <el-form-item label="配置键" prop="key">
          <el-input v-model="configForm.key" placeholder="请输入配置键" />
        </el-form-item>
        <el-form-item label="配置值" prop="value">
          <el-input v-model="configForm.value" placeholder="请输入配置值" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-input v-model="configForm.type" placeholder="请输入类型" />
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getConfigList, createConfig, updateConfig, deleteConfig } from '@/api/modules/config'

defineOptions({
  name: 'ConfigManagement',
})

interface Config {
  id: number
  name: string
  key: string
  value: string
  type: string
  description: string
  status: number
  createdAt: string
}

const configList = ref<Config[]>([])
const loading = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
const dialogVisible = ref(false)
const dialogTitle = ref('')
const configFormRef = ref()
const configForm = reactive({
  id: 0,
  name: '',
  key: '',
  value: '',
  type: '',
  description: '',
  status: 1,
})
const formRules = reactive({
  name: [{ required: true, message: '请输入配置名称', trigger: 'blur' }],
  key: [{ required: true, message: '请输入配置键', trigger: 'blur' }],
  value: [{ required: true, message: '请输入配置值', trigger: 'blur' }],
  type: [{ required: true, message: '请输入类型', trigger: 'blur' }],
})

// 加载配置列表
const loadConfigList = async () => {
  loading.value = true
  try {
    const response = await getConfigList({
      page: pagination.page,
      pageSize: pagination.pageSize,
    })
    const pageData = response.data as Record<string, unknown>
    configList.value = (pageData?.list as Config[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    ElMessage.error('获取配置列表失败')
    console.error('获取配置列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 分页大小变化
const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadConfigList()
}

// 当前页码变化
const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadConfigList()
}

// 添加配置
const handleAddConfig = () => {
  dialogTitle.value = '添加配置'
  Object.assign(configForm, {
    id: 0,
    name: '',
    key: '',
    value: '',
    type: '',
    description: '',
    status: 1,
  })
  dialogVisible.value = true
}

// 编辑配置
const handleEditConfig = (config: Config) => {
  dialogTitle.value = '编辑配置'
  Object.assign(configForm, config)
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await configFormRef.value.validate()
    if (configForm.id) {
      await updateConfig(configForm.id, {
        name: configForm.name,
        key: configForm.key,
        value: configForm.value,
        type: configForm.type,
        description: configForm.description,
        status: configForm.status,
      })
    } else {
      await createConfig({
        name: configForm.name,
        key: configForm.key,
        value: configForm.value,
        type: configForm.type,
        description: configForm.description,
        status: configForm.status,
      })
    }
    ElMessage.success(configForm.id ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadConfigList()
  } catch (error) {
    console.error('提交表单失败:', error)
  }
}

// 删除配置
const handleDeleteConfig = async (id: number) => {
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
