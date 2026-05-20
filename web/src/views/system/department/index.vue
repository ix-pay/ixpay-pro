<template>
  <!-- 部门管理页面 - 树形表格设计 -->
  <div class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md">
    <!-- 顶部操作栏 -->
    <div class="flex flex-col gap-3 p-4 border-b">
      <!-- 第一行：搜索条件 -->
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.keyword"
          placeholder="搜索部门名称"
          clearable
          style="width: 192px"
          @keyup.enter="loadDepartmentList"
        >
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="loadDepartmentList">
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

      <!-- 第二行：功能按钮 -->
      <div class="flex flex-wrap items-center gap-2">
        <el-button type="primary" v-auth-btn="'system:department:add'" @click="handleAddDepartment">
          <el-icon>
            <Plus />
          </el-icon>
          添加部门
        </el-button>
      </div>
    </div>

    <!-- 统计信息 -->
    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500">
            <OfficeBuilding />
          </el-icon>
          部门总数：<span class="font-medium">{{ statistics.totalCount }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <CircleCheck />
          </el-icon>
          启用：<span class="font-medium">{{ statistics.enabledCount }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-red-500">
            <CircleClose />
          </el-icon>
          禁用：<span class="font-medium">{{ statistics.disabledCount }}</span>
        </span>
      </div>
    </div>

    <!-- 表格区域 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="departmentList"
        stripe
        class="w-full h-full"
        :height="'100%'"
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="expandAll"
        :indent="48"
      >
        <!-- 部门名称列 -->
        <el-table-column prop="name" label="部门名称" fixed="left">
          <template #default="scope">
            <div style="display: inline-flex; align-items: center; gap: 8px; flex-wrap: nowrap">
              <el-icon :size="16" style="flex-shrink: 0; color: var(--el-color-primary)">
                <OfficeBuilding />
              </el-icon>
              <span
                :class="{ 'font-medium': !scope.row.parentId || scope.row.parentId === '0' }"
                style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis"
              >
                {{ scope.row.name }}
              </span>
            </div>
          </template>
        </el-table-column>

        <!-- 排序 -->
        <el-table-column prop="sort" label="排序" width="80" align="center">
          <template #default="scope">
            <span>{{ scope.row.sort }}</span>
          </template>
        </el-table-column>

        <!-- 状态 -->
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

        <!-- 创建时间 -->
        <el-table-column prop="createdAt" label="创建时间" width="160">
          <template #default="scope">
            <span style="color: var(--text-secondary)">
              {{ formatDate(scope.row.createdAt) }}
            </span>
          </template>
        </el-table-column>

        <!-- 操作列 -->
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button
                v-auth-btn="'system:department:edit'"
                type="primary"
                size="small"
                @click="handleEditDepartment(scope.row)"
              >
                编辑
              </el-button>
              <el-popconfirm title="确定要删除吗？" @confirm="handleDeleteDepartment(scope.row.id)">
                <template #reference>
                  <el-button v-auth-btn="'system:department:delete'" type="danger" size="small">
                    删除
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 部门表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      :close-on-click-modal="false"
    >
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
            :data="departmentTreeOptions"
            :props="{ value: 'id', label: 'name', children: 'children' }"
            placeholder="请选择父部门（不选则为顶级部门）"
            clearable
            check-strictly
            value-key="id"
            :render-after-expand="false"
            class="w-full"
          />
          <div v-if="!departmentForm.parentId" class="text-xs text-gray-500 mt-1">
            不选择则为顶级部门
          </div>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number
            v-model="departmentForm.sort"
            :min="0"
            :max="999"
            controls-position="right"
            class="w-full"
          />
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
        <div class="flex justify-end gap-3">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'
import {
  Plus,
  Search,
  Refresh,
  OfficeBuilding,
  CircleCheck,
  CircleClose,
} from '@element-plus/icons-vue'
import {
  getDepartmentList,
  createDepartment,
  updateDepartment,
  deleteDepartment,
} from '@/api/modules/department'
import { formatDate } from '@/utils/format'

defineOptions({
  name: 'DepartmentManagement',
})

interface Department {
  id: string
  name: string
  parentId: string
  sort: number
  status: number
  createdAt: string
  children?: Department[]
}

// 部门列表数据
const departmentList = ref<Department[]>([])
// 加载状态
const loading = ref(false)
// 是否展开所有层级
const expandAll = ref(true)
// 搜索表单
const searchForm = reactive({
  keyword: '',
})
// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('')
// 表单引用
const departmentFormRef = ref<FormInstance | null>(null)
// 部门表单数据
const departmentForm = reactive({
  id: '',
  name: '',
  parentId: undefined as string | undefined,
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

// 统计数据
const statistics = computed(() => {
  let totalCount = 0
  let enabledCount = 0
  let disabledCount = 0

  const count = (depts: Department[]) => {
    for (const dept of depts) {
      totalCount++
      if (dept.status === 1) {
        enabledCount++
      } else {
        disabledCount++
      }
      if (dept.children && dept.children.length > 0) {
        count(dept.children)
      }
    }
  }
  count(departmentList.value)

  return {
    totalCount,
    enabledCount,
    disabledCount,
  }
})

// 部门树形选项（用于 el-tree-select）
const departmentTreeOptions = computed(() => {
  // 过滤掉当前编辑的部门，避免选择自己作为父部门
  const filterDept = (depts: Department[], excludeId?: string): Department[] => {
    return depts
      .filter((dept) => dept.id !== excludeId)
      .map((dept) => ({
        ...dept,
        children: dept.children ? filterDept(dept.children, excludeId) : [],
      }))
  }

  return filterDept(departmentList.value, departmentForm.id || undefined)
})

// 获取部门列表
const loadDepartmentList = async () => {
  loading.value = true
  try {
    const response = await getDepartmentList({ page: 1, pageSize: 1000 })

    // 处理后端返回的数据格式：{ code: 0, data: [...], msg: "..." }
    let list: Department[] = []
    if (Array.isArray(response.data)) {
      // 直接是数组格式
      list = response.data as Department[]
    } else if ((response.data as Record<string, unknown>)?.list && Array.isArray((response.data as Record<string, unknown>).list)) {
      // 分页格式 { list: [], total: 0 }
      list = (response.data as Record<string, unknown>).list as Department[]
    } else if ((response.data as Record<string, unknown>)?.data && Array.isArray((response.data as Record<string, unknown>).data)) {
      // 嵌套格式 { data: { data: [] } }
      list = (response.data as Record<string, unknown>).data as Department[]
    }

    // 如果有搜索条件，进行过滤
    if (searchForm.keyword) {
      departmentList.value = filterDepartments(list, searchForm.keyword)
    } else {
      departmentList.value = list
    }
  } catch (error) {
    ElMessage.error('获取部门列表失败')
    console.error('获取部门列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 递归过滤部门树
const filterDepartments = (depts: Department[], keyword: string): Department[] => {
  const result: Department[] = []
  for (const dept of depts) {
    // 如果当前部门匹配
    if (dept.name?.toLowerCase().includes(keyword.toLowerCase())) {
      // 复制部门并保留子部门
      const clonedDept: Department = { ...dept }
      if (dept.children && dept.children.length > 0) {
        clonedDept.children = filterDepartments(dept.children, keyword)
      }
      result.push(clonedDept)
    } else if (dept.children && dept.children.length > 0) {
      // 递归过滤子部门
      const filteredChildren = filterDepartments(dept.children, keyword)
      if (filteredChildren.length > 0) {
        // 如果子部门有匹配的，保留父部门
        result.push({ ...dept, children: filteredChildren })
      }
    }
  }
  return result
}

// 重置搜索
const resetSearch = () => {
  searchForm.keyword = ''
  loadDepartmentList()
}

// 状态变更
const handleStatusChange = async (department: Department) => {
  try {
    await updateDepartment({
      id: department.id,
      name: department.name,
      parentId: department.parentId,
      sort: department.sort,
      status: department.status,
    })
    ElMessage.success('状态更新成功')
  } catch (error) {
    ElMessage.error('状态更新失败')
    console.error('状态更新失败:', error)
    // 恢复原状态
    department.status = department.status === 1 ? 0 : 1
  }
}

// 添加部门
const handleAddDepartment = () => {
  dialogTitle.value = '添加部门'
  // 重置表单
  Object.assign(departmentForm, {
    id: '',
    name: '',
    parentId: undefined,
    sort: 0,
    status: 1,
  })
  dialogVisible.value = true
}

// 编辑部门
const handleEditDepartment = (department: Department) => {
  dialogTitle.value = '编辑部门'
  // 填充表单
  Object.assign(departmentForm, {
    ...department,
    parentId: department.parentId === '0' ? undefined : department.parentId,
  })
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await departmentFormRef.value?.validate()

    // 构建提交数据
    const submitData = {
      name: departmentForm.name,
      parentId: departmentForm.parentId ?? '0',
      sort: departmentForm.sort,
      status: departmentForm.status,
    }

    if (departmentForm.id) {
      // 编辑部门 - 需要包含 ID 字段
      await updateDepartment({
        id: departmentForm.id,
        ...submitData,
      })
    } else {
      // 添加部门
      await createDepartment(submitData)
    }
    ElMessage.success(departmentForm.id ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadDepartmentList()
  } catch (error) {
    console.error('提交表单失败:', error)
  }
}

// 删除部门
const handleDeleteDepartment = async (id: string) => {
  try {
    await deleteDepartment(id)
    ElMessage.success('删除成功')
    loadDepartmentList()
  } catch (error) {
    ElMessage.error('删除失败')
    console.error('删除部门失败:', error)
  }
}

onMounted(() => {
  loadDepartmentList()
})
</script>

<style scoped>
/* 
 * 部门管理页面样式说明：
 * - 布局使用 Tailwind CSS
 * - Element Plus 组件使用原生样式 + 必要的 Tailwind 辅助类
 */

/* 表格容器高度 */
.flex-1 {
  min-height: 0;
  /* 允许 flex 子项滚动 */
}

/* 表格样式 */
:deep(.el-table) {
  font-size: 14px;
}

/* 表头样式 */
:deep(.el-table__header th) {
  background-color: var(--bg-tertiary) !important;
  color: var(--text-primary) !important;
  font-weight: 600 !important;
}

:deep(.el-table__fixed-header-wrapper th),
:deep(.el-table__fixed-right-header-wrapper th) {
  background-color: var(--bg-tertiary) !important;
}

:deep(.el-table__fixed-body-wrapper),
:deep(.el-table__fixed-right-body-wrapper) {
  background-color: var(--bg-primary);
}

/* 树形表格展开图标垂直居中 */
:deep(.el-table__expand-icon) {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  height: 100% !important;
  vertical-align: middle !important;
  margin-right: 4px !important;
}

/* 表格单元格内容 */
:deep(.el-table .cell) {
  white-space: normal;
  word-wrap: break-word;
}

/* 表格滚动条 */
:deep(.el-table__body-wrapper) {
  overflow: auto !important;
}
</style>
