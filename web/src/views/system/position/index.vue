<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <div class="flex flex-col gap-3 p-4 border-b">
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.name"
          placeholder="请输入职位名称"
          clearable
          style="width: 192px"
          @keyup.enter="loadPositionList"
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
        <el-button type="primary" @click="loadPositionList">
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
        <el-button type="primary" v-auth-btn="'system:position:add'" @click="handleAddPosition">
          <el-icon>
            <Plus />
          </el-icon>
          添加职位
        </el-button>
      </div>
    </div>

    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500">
            <Briefcase />
          </el-icon>
          岗位总数：<span class="font-medium">{{ pagination.total }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <SuccessFilled />
          </el-icon>
          启用：<span class="font-medium">{{
            positionList.filter((p) => p.status === 1).length
          }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-orange-500">
            <CircleClose />
          </el-icon>
          禁用：<span class="font-medium">{{
            positionList.filter((p) => p.status === 0).length
          }}</span>
        </span>
      </div>
    </div>

    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="positionList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="name" label="职位名称" width="160" />
        <el-table-column prop="code" label="职位编码" width="140" />
        <el-table-column prop="description" label="职位描述" min-width="200" />
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              size="small"
              @change="handleStatusChange(scope.row)"
            />
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
                v-auth-btn="'system:position:edit'"
                type="primary"
                size="small"
                @click="handleEditPosition(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:position:delete'"
                type="danger"
                size="small"
                @click="handleDeletePosition(scope.row.id)"
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
      <el-form ref="positionFormRef" :model="positionForm" :rules="formRules" label-width="100px">
        <el-form-item label="职位名称" prop="name">
          <el-input v-model="positionForm.name" placeholder="请输入职位名称" />
        </el-form-item>
        <el-form-item label="职位编码" prop="code">
          <el-input
            v-model="positionForm.code"
            placeholder="请输入职位编码"
            :disabled="!!positionForm.id"
          />
        </el-form-item>
        <el-form-item label="职位描述" prop="description">
          <el-input
            v-model="positionForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入职位描述"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="positionForm.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="positionForm.status"
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
import {
  Plus,
  Search,
  Refresh,
  Briefcase,
  SuccessFilled,
  CircleClose,
} from '@element-plus/icons-vue'
import {
  getPositionList,
  createPosition,
  updatePosition,
  deletePosition,
} from '@/api/modules/position'
import { formatDate } from '@/utils/format'

defineOptions({
  name: 'PositionManagement',
})

interface Position {
  id: string
  name: string
  code: string
  description: string
  sort: number
  status: number
  createdAt: string
}

const positionList = ref<Position[]>([])
const loading = ref(false)
const isLoading = ref(false)

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const searchForm = reactive({
  name: '',
  status: undefined as number | undefined,
})

const dialogVisible = ref(false)
const dialogTitle = ref('')
const positionFormRef = ref<FormInstance>()

const positionForm = reactive({
  id: '',
  name: '',
  code: '',
  description: '',
  sort: 0,
  status: 1,
})

const formRules = reactive({
  name: [
    { required: true, message: '请输入职位名称', trigger: 'blur' },
    { min: 1, max: 50, message: '职位名称长度在 1 到 50 个字符', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入职位编码', trigger: 'blur' },
    { min: 1, max: 50, message: '职位编码长度在 1 到 50 个字符', trigger: 'blur' },
  ],
})

const loadPositionList = async () => {
  if (isLoading.value) return

  isLoading.value = true
  loading.value = true
  try {
    const response = await getPositionList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.name ? { name: searchForm.name } : {}),
      ...(searchForm.status !== undefined ? { status: searchForm.status } : {}),
    })
    const pageData = response.data as Record<string, unknown>
    positionList.value = (pageData?.list as Position[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    console.error('获取职位列表失败:', error)
    positionList.value = []
    pagination.total = 0
  } finally {
    loading.value = false
    isLoading.value = false
  }
}

const resetSearch = () => {
  searchForm.name = ''
  searchForm.status = undefined
  pagination.page = 1
  loadPositionList()
}

// 状态变更
const handleStatusChange = async (position: Position) => {
  try {
    await updatePosition({
      id: position.id,
      name: position.name,
      code: position.code,
      description: position.description,
      sort: position.sort,
      status: position.status,
    })
    ElMessage.success('状态更新成功')
  } catch (error) {
    ElMessage.error('状态更新失败')
    console.error('状态更新失败:', error)
    // 恢复原状态
    position.status = position.status === 1 ? 0 : 1
  }
}

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadPositionList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadPositionList()
}

const handleAddPosition = () => {
  dialogTitle.value = '添加职位'
  Object.assign(positionForm, {
    id: '',
    name: '',
    code: '',
    description: '',
    sort: 0,
    status: 1,
  })
  dialogVisible.value = true
}

const handleEditPosition = (position: Position) => {
  dialogTitle.value = '编辑职位'
  Object.assign(positionForm, position)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    await positionFormRef.value?.validate()
    if (positionForm.id) {
      await updatePosition({
        id: positionForm.id,
        name: positionForm.name,
        code: positionForm.code,
        description: positionForm.description,
        sort: positionForm.sort,
        status: positionForm.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createPosition({
        name: positionForm.name,
        code: positionForm.code,
        description: positionForm.description,
        sort: positionForm.sort,
        status: positionForm.status,
      })
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadPositionList()
  } catch {
    // 验证失败
  }
}

const handleDeletePosition = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除该职位吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deletePosition(id)
    ElMessage.success('删除成功')
    loadPositionList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除职位失败:', error)
    }
  }
}

onMounted(() => {
  loadPositionList()
})
</script>

<style scoped>
/* 职位管理页面样式 */
</style>
