<template>
  <!-- 用户管理页面 - 使用 Tailwind CSS + Element Plus 混合架构 -->
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 顶部操作栏 -->
    <div
      class="flex flex-wrap items-center justify-between gap-3 p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.userName"
          placeholder="请输入用户名"
          clearable
          size="default"
          class="w-64"
          @keyup.enter="loadUserList"
        >
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="loadUserList">
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
        <el-button type="primary" v-auth-btn="'system:user:add'" @click="handleAddUser">
          <el-icon>
            <Plus />
          </el-icon>
          添加用户
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
            <User />
          </el-icon>
          用户总数：<span class="font-medium">{{ pagination.total }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <SuccessFilled />
          </el-icon>
          启用：<span class="font-medium">{{ userList.filter((u) => u.status === 1).length }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-orange-500">
            <CircleClose />
          </el-icon>
          禁用：<span class="font-medium">{{ userList.filter((u) => u.status === 0).length }}</span>
        </span>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table v-loading="loading" :data="userList" stripe class="w-full h-full" :height="'100%'">
        <el-table-column prop="userName" label="用户名" width="160" />
        <el-table-column prop="nickname" label="昵称" min-width="100" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="phone" label="电话" min-width="120" />
        <el-table-column label="角色" min-width="120">
          <template #default="scope">
            {{ scope.row.roles?.map((role: RoleType) => role.name).join(', ') || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <div v-if="!isAdminUser(scope.row)" class="flex gap-2">
              <el-button
                v-auth-btn="'system:user:edit'"
                type="primary"
                @click="handleEditUser(scope.row)"
              >
                编辑
              </el-button>
              <!-- 管理员账户不允许删除 -->
              <el-button
                v-auth-btn="'system:user:delete'"
                type="danger"
                @click="handleDeleteUser(scope.row.id)"
              >
                删除
              </el-button>
              <el-button
                v-auth-btn="'system:user:view'"
                type="warning"
                @click="handleResetPassword(scope.row.id)"
              >
                重置密码
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

    <!-- 用户表单对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="userFormRef" :model="userForm" :rules="formRules" label-width="100px">
        <el-form-item label="用户名" prop="userName">
          <el-input v-model="userForm.userName" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!userForm.id">
          <el-input type="password" v-model="userForm.password" placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="userForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="电话" prop="phone">
          <el-input v-model="userForm.phone" placeholder="请输入电话" />
        </el-form-item>
        <el-form-item label="角色" prop="roles">
          <el-select v-model="userForm.roles" multiple placeholder="请选择角色" class="w-full">
            <el-option
              v-for="role in roleList"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="userForm.status"
            active-color="#13ce66"
            inactive-color="#ff4949"
            :active-value="1"
            :inactive-value="0"
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

    <!-- 重置密码对话框 -->
    <el-dialog v-model="resetPasswordVisible" title="重置密码" width="400px">
      <el-empty description="确定要重置该用户的密码吗？" />
      <template #footer>
        <div class="flex justify-end gap-2">
          <el-button @click="resetPasswordVisible = false">取消</el-button>
          <el-button type="primary" @click="handleResetPasswordSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus, Search, Refresh, User, SuccessFilled, CircleClose } from '@element-plus/icons-vue'
import {
  getUserList,
  deleteUser,
  createUser,
  updateUserInfo,
  resetPassword,
} from '@/api/modules/user'
import { getRolesList, type Role as RoleType } from '@/api/modules/role'

defineOptions({
  name: 'UserManagement',
})

// 使用统一的 Role 类型定义（来自 api/modules/role）

// 用户类型定义
interface User {
  id: string
  userName: string
  nickname: string
  email: string
  phone: string
  roles: RoleType[]
  status: number
  created_at?: string
}

// 判断用户是否为管理员
const isAdminUser = (user: User): boolean => {
  // 用户名为 admin 或者拥有 admin 角色的用户都是管理员
  return user.userName === 'admin' || user.roles?.some((role) => role.code === 'admin') || false
}

// 表单验证规则类型
interface UserFormRules {
  userName: Array<
    | { required: boolean; message: string; trigger: string }
    | { min: number; max: number; message: string; trigger: string }
  >
  password: Array<
    | { required: boolean; message: string; trigger: string }
    | { min: number; max: number; message: string; trigger: string }
  >
  nickname: Array<{ required: boolean; message: string; trigger: string }>
  email: Array<
    | { required: boolean; message: string; trigger: string }
    | {
        validator: (rule: unknown, value: string, callback: (arg0?: Error) => void) => void
        trigger: string
      }
  >
  phone: Array<{
    validator: (rule: unknown, value: string, callback: (arg0?: Error) => void) => void
    trigger: string
  }>
  roles: Array<{ required: boolean; message: string; trigger: string }>
}

// 用户列表数据
const userList = ref<User[]>([])
// 加载状态
const loading = ref(false)
// 防止重复加载的标志
const isLoading = ref(false)
// 分页数据
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
// 搜索表单
const searchForm = reactive({
  userName: '',
})
// 角色列表
const roleList = ref<RoleType[]>([])
// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('')
// 表单引用
const userFormRef = ref<FormInstance | null>(null)
// 用户表单数据
const userForm = reactive({
  id: '',
  userName: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  roles: [] as string[],
  status: 1,
})
// 重置密码对话框
const resetPasswordVisible = ref(false)
const resetPasswordForm = reactive({
  userId: '',
})

// 获取角色列表
const loadRolesList = async () => {
  try {
    const response = await getRolesList()
    roleList.value = response.data || []
  } catch (error) {
    ElMessage.error('获取角色列表失败')
    console.error('获取角色列表失败:', error)
  }
}

// 获取用户列表
const loadUserList = async () => {
  if (isLoading.value) {
    return
  }

  isLoading.value = true
  loading.value = true
  try {
    const response = await getUserList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.userName ? { userName: searchForm.userName } : {}),
    })
    const pageData = response.data as Record<string, unknown>
    userList.value = (pageData?.list as User[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    // 静默处理错误，不显示错误弹窗
    // 当搜索条件无匹配数据时，后端可能返回 400 错误，此时只需清空列表即可
    console.error('获取用户列表失败:', error)
    userList.value = []
    pagination.total = 0
  } finally {
    loading.value = false
    isLoading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  searchForm.userName = ''
  pagination.page = 1
  loadUserList()
}

// 分页处理
const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadUserList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadUserList()
}

// 添加用户
const handleAddUser = () => {
  dialogTitle.value = '添加用户'
  Object.assign(userForm, {
    id: '',
    userName: '',
    password: '',
    nickname: '',
    email: '',
    phone: '',
    roles: [],
    status: 1,
  })
  dialogVisible.value = true
}

// 编辑用户
const handleEditUser = (user: User) => {
  dialogTitle.value = '编辑用户'
  Object.assign(userForm, {
    id: user.id,
    userName: user.userName,
    nickname: user.nickname,
    email: user.email,
    phone: user.phone,
    roles: user.roles?.map((role) => role.id) || [],
    status: user.status,
  })
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await userFormRef.value?.validate()
    if (userForm.id) {
      await updateUserInfo({
        id: userForm.id,
        nickname: userForm.nickname,
        email: userForm.email,
        phone: userForm.phone,
        status: userForm.status,
        roles: userForm.roles,
      })
      ElMessage.success('更新成功')
    } else {
      await createUser(userForm)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadUserList()
  } catch {
    // 验证失败或提交失败
  }
}

// 删除用户
const handleDeleteUser = async (id: string) => {
  try {
    // 查找要删除的用户
    const userToDelete = userList.value.find((user) => user.id === id)
    if (!userToDelete) {
      ElMessage.error('用户不存在')
      return
    }

    // 检查是否为管理员
    if (isAdminUser(userToDelete)) {
      ElMessage.error('管理员账户不允许删除')
      return
    }

    await ElMessageBox.confirm('确定要删除该用户吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteUser(id)
    ElMessage.success('删除成功')
    loadUserList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除用户失败:', error)
    }
  }
}

// 重置密码
const handleResetPassword = (id: string) => {
  resetPasswordForm.userId = id
  resetPasswordVisible.value = true
}

// 提交重置密码
const handleResetPasswordSubmit = async () => {
  try {
    await resetPassword({
      userId: resetPasswordForm.userId,
    })
    ElMessage.success('密码重置成功')
    resetPasswordVisible.value = false
  } catch (error) {
    console.error('重置密码失败:', error)
  }
}

// 格式化日期
const formatDate = (date: string | null | undefined) => {
  if (!date) return '-'
  try {
    return new Date(date).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch {
    return '-'
  }
}

// 表单验证规则
const formRules = reactive<UserFormRules>({
  userName: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' },
  ],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    {
      validator: (_rule: unknown, value: string, callback: (arg0?: Error) => void) => {
        if (value && !/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(value)) {
          callback(new Error('请输入正确的邮箱地址'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  phone: [
    {
      validator: (_rule: unknown, value: string, callback: (arg0?: Error) => void) => {
        if (value && !/^1[3-9]\d{9}$/.test(value)) {
          callback(new Error('请输入正确的手机号码'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  roles: [{ required: true, message: '请选择角色', trigger: 'change' }],
})

onMounted(() => {
  loadRolesList()
  loadUserList()
})
</script>

<style scoped>
/* 
 * 用户管理页面样式说明：
 * - 布局使用 Tailwind CSS
 * - Element Plus 组件使用原生样式 + 必要的 Tailwind 辅助类
 */

/* 表格容器高度 */
.flex-1 {
  min-height: 0;
  /* 允许 flex 子项滚动 */
}

/* 表格样式修正 */
:deep(.el-table) {
  font-size: 14px;
}

/* 固定列背景色 - 使用项目主题变量 */
:deep(.el-table__header th) {
  background-color: var(--bg-dark) !important;
  color: var(--text-primary) !important;
  font-weight: 600 !important;
}

/* 表格单元格内容 */
:deep(.el-table .cell) {
  white-space: normal;
  word-wrap: break-word;
}

/* 修复表格滚动条 */
:deep(.el-table__body-wrapper) {
  overflow: auto !important;
}
</style>
