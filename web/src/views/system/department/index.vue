<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <span class="text-sm text-gray-600 dark:text-gray-400">部门列表</span>
      <el-button
        type="primary"
        size="small"
        v-auth-btn="'system:department:add'"
        @click="handleAddDepartment"
      >
        <el-icon><Plus /></el-icon>
        添加部门
      </el-button>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="departmentList"
        stripe
        row-key="id"
        :tree-props="{ children: 'children' }"
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="name" label="部门名称" width="200" />
        <el-table-column prop="sort" label="排序" width="80" />
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
              <el-button
                v-auth-btn="'system:department:edit'"
                size="small"
                type="primary"
                link
                @click="handleEditDepartment(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:department:delete'"
                size="small"
                type="danger"
                link
                @click="handleDeleteDepartment(scope.row.id)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 部门表单对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form
        ref="departmentFormRef"
        :model="departmentForm"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="departmentForm.name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="父部门" prop="parentId">
          <el-tree-select
            v-model="departmentForm.parentId"
            :data="departmentTreeData"
            check-strictly
            placeholder="请选择父部门"
            :render-after-expand="false"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="departmentForm.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="departmentForm.status"
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
import {
  getDepartmentList,
  createDepartment,
  updateDepartment,
  deleteDepartment,
} from '@/api/modules/department'

defineOptions({
  name: 'DepartmentManagement',
})

interface Department {
  id: number
  name: string
  parentId: number
  sort: number
  status: number
  createdAt: string
  children?: Department[]
}

// 部门列表数据
const departmentList = ref<Department[]>([])
// 加载状态
const loading = ref(false)
// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('')
// 表单引用
const departmentFormRef = ref()
// 部门表单数据
const departmentForm = reactive({
  id: 0,
  name: '',
  parentId: 0,
  sort: 0,
  status: 1,
})
// 表单验证规则
const formRules = reactive({
  name: [
    { required: true, message: '请输入部门名称', trigger: 'blur' },
    { min: 1, max: 50, message: '部门名称长度在 1 到 50 个字符', trigger: 'blur' },
  ],
})
// 部门树形数据
const departmentTreeData = ref<Department[]>([])

// 获取部门列表
const loadDepartmentList = async () => {
  loading.value = true
  try {
    const response = await getDepartmentList()
    departmentList.value = response.data || []
    departmentTreeData.value = response.data || []
  } catch (error) {
    ElMessage.error('获取部门列表失败')
    console.error('获取部门列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 添加部门
const handleAddDepartment = () => {
  dialogTitle.value = '添加部门'
  // 重置表单
  Object.assign(departmentForm, {
    id: 0,
    name: '',
    parentId: 0,
    sort: 0,
    status: 1,
  })
  dialogVisible.value = true
}

// 编辑部门
const handleEditDepartment = (department: Department) => {
  dialogTitle.value = '编辑部门'
  // 填充表单
  Object.assign(departmentForm, department)
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await departmentFormRef.value.validate()
    if (departmentForm.id) {
      // 编辑部门
      await updateDepartment(departmentForm.id, {
        name: departmentForm.name,
        parentId: departmentForm.parentId,
        sort: departmentForm.sort,
        status: departmentForm.status,
      })
    } else {
      // 添加部门
      await createDepartment({
        name: departmentForm.name,
        parentId: departmentForm.parentId,
        sort: departmentForm.sort,
        status: departmentForm.status,
      })
    }
    ElMessage.success(departmentForm.id ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadDepartmentList()
  } catch (error) {
    console.error('提交表单失败:', error)
  }
}

// 删除部门
const handleDeleteDepartment = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该部门吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteDepartment(id)
    ElMessage.success('删除成功')
    loadDepartmentList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除部门失败:', error)
    }
  }
}

onMounted(() => {
  loadDepartmentList()
})
</script>
