<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-2">
        <el-input
          v-model="searchForm.name"
          placeholder="请输入职位名称"
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
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" size="small" @click="loadPositionList">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
      <el-button
        type="primary"
        size="small"
        v-auth-btn="'system:position:add'"
        @click="handleAddPosition"
      >
        <el-icon><Plus /></el-icon>
        添加职位
      </el-button>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
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
        <el-table-column prop="status" label="状态" width="70">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <div class="flex gap-2">
              <el-button
                v-auth-btn="'system:position:edit'"
                type="primary"
                @click="handleEditPosition(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:position:delete'"
                type="danger"
                @click="handleDeletePosition(scope.row.id)"
              >
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

    <!-- 职位表单对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="positionFormRef" :model="positionForm" :rules="formRules" label-width="100px">
        <el-form-item label="职位名称" prop="name">
          <el-input v-model="positionForm.name" placeholder="请输入职位名称" />
        </el-form-item>
        <el-form-item label="职位编码" prop="code">
          <el-input v-model="positionForm.code" placeholder="请输入职位编码" />
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import {
  getPositionList,
  createPosition,
  updatePosition,
  deletePosition,
} from '@/api/modules/position'

defineOptions({
  name: 'PositionManagement',
})

interface Position {
  id: number
  name: string
  code: string
  description: string
  sort: number
  status: number
  createdAt: string
}

// 职位列表数据
const positionList = ref<Position[]>([])
// 加载状态
const loading = ref(false)
// 分页数据
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
// 搜索表单
const searchForm = reactive({
  name: '',
  status: undefined as number | undefined,
})
// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('')
// 表单引用
const positionFormRef = ref()
// 职位表单数据
const positionForm = reactive({
  id: 0,
  name: '',
  code: '',
  description: '',
  sort: 0,
  status: 1,
})
// 表单验证规则
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

// 获取职位列表
const loadPositionList = async () => {
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
    ElMessage.error('获取职位列表失败')
    console.error('获取职位列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 分页大小变化
const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadPositionList()
}

// 当前页码变化
const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadPositionList()
}

// 添加职位
const handleAddPosition = () => {
  dialogTitle.value = '添加职位'
  Object.assign(positionForm, {
    id: 0,
    name: '',
    code: '',
    description: '',
    sort: 0,
    status: 1,
  })
  dialogVisible.value = true
}

// 编辑职位
const handleEditPosition = (position: Position) => {
  dialogTitle.value = '编辑职位'
  Object.assign(positionForm, position)
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await positionFormRef.value.validate()
    if (positionForm.id) {
      // 编辑职位
      await updatePosition(positionForm.id, {
        name: positionForm.name,
        code: positionForm.code,
        description: positionForm.description,
        sort: positionForm.sort,
        status: positionForm.status,
      })
    } else {
      // 添加职位
      await createPosition({
        name: positionForm.name,
        code: positionForm.code,
        description: positionForm.description,
        sort: positionForm.sort,
        status: positionForm.status,
      })
    }
    ElMessage.success(positionForm.id ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadPositionList()
  } catch (error) {
    console.error('提交表单失败:', error)
  }
}

// 删除职位
const handleDeletePosition = async (id: number) => {
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
