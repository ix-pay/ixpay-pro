<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <span class="text-sm text-gray-600 dark:text-gray-400">字典列表</span>
      <el-button type="primary" size="small" @click="handleAddDict">
        <el-icon><Plus /></el-icon>
        添加字典
      </el-button>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table v-loading="loading" :data="dictList" stripe class="w-full h-full" :height="'100%'">
        <el-table-column prop="name" label="字典名称" width="160" />
        <el-table-column prop="code" label="字典编码" width="160" />
        <el-table-column prop="type" label="字典类型" width="120" />
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
              <el-button size="small" type="primary" link @click="handleEditDict(scope.row)">
                编辑
              </el-button>
              <el-button size="small" type="danger" link @click="handleDeleteDict(scope.row.id)">
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

    <!-- 字典表单对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="dictFormRef" :model="dictForm" :rules="formRules" label-width="100px">
        <el-form-item label="字典名称" prop="name">
          <el-input v-model="dictForm.name" placeholder="请输入字典名称" />
        </el-form-item>
        <el-form-item label="字典编码" prop="code">
          <el-input v-model="dictForm.code" placeholder="请输入字典编码" />
        </el-form-item>
        <el-form-item label="字典类型" prop="type">
          <el-input v-model="dictForm.type" placeholder="请输入字典类型" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="dictForm.status"
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
import { getDictList, createDict, updateDict, deleteDict } from '@/api/modules/dict'

defineOptions({
  name: 'DictManagement',
})

interface Dict {
  id: number
  name: string
  code: string
  type: string
  status: number
  createdAt: string
}

const dictList = ref<Dict[]>([])
const loading = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
const dialogVisible = ref(false)
const dialogTitle = ref('')
const dictFormRef = ref()
const dictForm = reactive({
  id: 0,
  name: '',
  code: '',
  type: '',
  status: 1,
})
const formRules = reactive({
  name: [{ required: true, message: '请输入字典名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入字典编码', trigger: 'blur' }],
  type: [{ required: true, message: '请输入字典类型', trigger: 'blur' }],
})

// 加载字典列表
const loadDictList = async () => {
  loading.value = true
  try {
    const response = await getDictList({
      page: pagination.page,
      pageSize: pagination.pageSize,
    })
    const pageData = response.data as Record<string, unknown>
    dictList.value = (pageData?.list as Dict[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    ElMessage.error('获取字典列表失败')
    console.error('获取字典列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 分页大小变化
const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadDictList()
}

// 当前页码变化
const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadDictList()
}

// 添加字典
const handleAddDict = () => {
  dialogTitle.value = '添加字典'
  Object.assign(dictForm, { id: 0, name: '', code: '', type: '', status: 1 })
  dialogVisible.value = true
}

// 编辑字典
const handleEditDict = (dict: Dict) => {
  dialogTitle.value = '编辑字典'
  Object.assign(dictForm, dict)
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await dictFormRef.value.validate()
    if (dictForm.id) {
      await updateDict(dictForm.id, {
        name: dictForm.name,
        code: dictForm.code,
        type: dictForm.type,
        status: dictForm.status,
      })
    } else {
      await createDict({
        name: dictForm.name,
        code: dictForm.code,
        type: dictForm.type,
        status: dictForm.status,
      })
    }
    ElMessage.success(dictForm.id ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadDictList()
  } catch (error) {
    console.error('提交表单失败:', error)
  }
}

// 删除字典
const handleDeleteDict = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该字典吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteDict(id)
    ElMessage.success('删除成功')
    loadDictList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除字典失败:', error)
    }
  }
}

onMounted(() => {
  loadDictList()
})
</script>
