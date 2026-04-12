<template>
  <!-- 角色管理页面 - 使用 Tailwind CSS + Element Plus 混合架构 -->
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-2">
        <el-input
          v-model="searchForm.name"
          placeholder="请输入角色名称"
          clearable
          size="small"
          class="w-48"
          @keyup.enter="loadRoleList"
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
        <el-button type="primary" size="small" @click="loadRoleList">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
      <el-button type="primary" size="small" @click="handleAddRole">
        <el-icon><Plus /></el-icon>
        添加角色
      </el-button>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table v-loading="loading" :data="roleList" stripe class="w-full h-full" :height="'100%'">
        <el-table-column prop="name" label="角色名称" width="160" />
        <el-table-column prop="description" label="角色描述" min-width="200" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <div class="flex gap-1">
              <el-button
                v-if="!scope.row.is_system"
                size="small"
                type="primary"
                link
                @click="handleEditRole(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-if="!scope.row.is_system"
                size="small"
                type="primary"
                link
                @click="handlePermission(scope.row)"
              >
                权限设置
              </el-button>
              <el-button
                v-if="!scope.row.is_system"
                size="small"
                type="danger"
                link
                @click="handleDeleteRole(scope.row.id)"
              >
                删除
              </el-button>
              <el-tag v-if="scope.row.is_system" type="info" size="small"> 系统角色 </el-tag>
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

    <!-- 角色表单对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="roleFormRef" :model="roleForm" :rules="formRules" label-width="100px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色描述" prop="description">
          <el-input
            v-model="roleForm.description"
            placeholder="请输入角色描述"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="roleForm.status"
            active-color="#13ce66"
            inactive-color="#ff4949"
            :active-value="1"
            :inactive-value="0"
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

    <!-- 权限设置对话框 -->
    <RolePermissionDialog
      v-model:visible="permissionDialogVisible"
      :role-id="currentRoleId"
      @success="handlePermissionSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getRoleList, deleteRole, createRole, updateRole } from '@/api/modules/role'
import RolePermissionDialog from './RolePermissionDialog.vue'

defineOptions({
  name: 'RoleManagement',
})

// 角色类型定义
interface Role {
  id: string
  name: string
  description: string
  status: number
  createdAt: string
  is_system?: boolean // 是否为系统角色
}

// 权限设置对话框状态
const permissionDialogVisible = ref(false)
const currentRoleId = ref('')

// 角色列表数据
const roleList = ref<Role[]>([])
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
// 对话框状态
const dialogVisible = ref(false)
const dialogTitle = ref('')
// 表单引用
const roleFormRef = ref<FormInstance | null>(null)
// 角色表单数据
const roleForm = reactive({
  id: '',
  name: '',
  description: '',
  status: 1,
})
// 表单验证规则
const formRules = reactive({
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' },
    { min: 2, max: 20, message: '角色名称长度在 2 到 20 个字符', trigger: 'blur' },
  ],
  description: [{ max: 100, message: '角色描述不能超过 100 个字符', trigger: 'blur' }],
})

// 获取角色列表
const loadRoleList = async () => {
  loading.value = true
  try {
    const response = await getRoleList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.name ? { name: searchForm.name } : {}),
      ...(searchForm.status !== undefined ? { status: searchForm.status } : {}),
    })
    const pageData = response.data as Record<string, unknown>
    roleList.value = (pageData?.list as Role[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    ElMessage.error('获取角色列表失败')
    console.error('获取角色列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 分页处理
const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadRoleList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadRoleList()
}

// 添加角色
const handleAddRole = () => {
  dialogTitle.value = '添加角色'
  // 重置表单
  Object.assign(roleForm, {
    id: '',
    name: '',
    description: '',
    status: 1,
  })
  dialogVisible.value = true
}

// 编辑角色
const handleEditRole = (role: Role) => {
  dialogTitle.value = '编辑角色'
  // 填充表单
  Object.assign(roleForm, role)
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await roleFormRef.value?.validate()
    if (roleForm.id) {
      // 编辑角色，只提交必要字段
      await updateRole({
        id: roleForm.id,
        name: roleForm.name,
        description: roleForm.description,
        status: roleForm.status,
      })
    } else {
      // 添加角色，只提交必要字段
      await createRole({
        name: roleForm.name,
        description: roleForm.description,
        status: roleForm.status,
      })
    }
    ElMessage.success(roleForm.id ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadRoleList()
  } catch (error) {
    console.error('提交表单失败:', error)
  }
}

// 删除角色
const handleDeleteRole = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除该角色吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteRole(id)
    ElMessage.success('删除成功')
    loadRoleList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除角色失败:', error)
    }
  }
}

// 权限设置
const handlePermission = (row: Role) => {
  currentRoleId.value = row.id
  permissionDialogVisible.value = true
}

// 权限设置成功回调
const handlePermissionSuccess = () => {
  loadRoleList()
}

onMounted(() => {
  loadRoleList()
})
</script>

<style scoped>
/* 
 * 角色管理页面样式说明：
 * - 布局和样式主要使用 Tailwind CSS
 * - 只需保留少量必要的自定义样式
 * - Element Plus 组件使用默认样式，通过 :deep() 进行暗黑模式适配
 */

/* 现代化表格样式 */
.modern-table :deep(.el-table) {
  border-radius: var(--radius-lg);
  overflow: hidden;
}

/* 暗黑模式下表格样式适配 */
html.dark :deep(.el-table) {
  background-color: var(--bg-color);
  color: var(--text-primary);

  .el-table__header-wrapper {
    th {
      background-color: var(--bg-dark);
      color: var(--text-primary);
      border-bottom-color: var(--border-color);
    }
  }

  .el-table__body-wrapper {
    tr {
      background-color: var(--bg-color);
      color: var(--text-primary);

      &:hover > td {
        background-color: var(--bg-light);
      }

      &.el-table__row--striped {
        background-color: var(--bg-dark);

        &:hover > td {
          background-color: var(--bg-light);
        }
      }

      td {
        border-bottom-color: var(--border-color);
      }
    }
  }
}
</style>
