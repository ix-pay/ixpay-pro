<template>
  <div class="profile-page">
    <h2>个人资料</h2>
    <div class="profile-content">
      <el-card shadow="hover">
        <template #header>
          <div class="card-header">
            <span>基本信息</span>
          </div>
        </template>
        <el-form
          ref="profileFormRef"
          :model="profileForm"
          :rules="formRules"
          label-width="120px"
          class="profile-form"
        >
          <el-form-item label="用户名" prop="username">
            <el-input v-model="profileForm.username" placeholder="请输入用户名" disabled />
          </el-form-item>
          <el-form-item label="昵称" prop="nickname">
            <el-input v-model="profileForm.nickname" placeholder="请输入昵称" />
          </el-form-item>
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="profileForm.email" placeholder="请输入邮箱" />
          </el-form-item>
          <el-form-item label="电话" prop="phone">
            <el-input v-model="profileForm.phone" placeholder="请输入电话" />
          </el-form-item>
          <el-form-item label="角色" prop="role">
            <el-input v-model="profileForm.role" placeholder="请输入角色" disabled />
          </el-form-item>
          <el-form-item label="状态" prop="status">
            <el-input v-model="statusLabel" placeholder="请输入状态" disabled />
          </el-form-item>
          <el-form-item label="创建时间" prop="createdAt">
            <el-input v-model="profileForm.createdAt" placeholder="请输入创建时间" disabled />
          </el-form-item>
        </el-form>
        <div class="form-actions">
          <el-button type="primary" @click="handleUpdateProfile">保存修改</el-button>
        </div>
      </el-card>

      <!-- 修改密码卡片 -->
      <el-card shadow="hover" style="margin-top: 20px">
        <template #header>
          <div class="card-header">
            <span>修改密码</span>
          </div>
        </template>
        <el-form
          ref="passwordFormRef"
          :model="passwordForm"
          :rules="passwordRules"
          label-width="120px"
          class="password-form"
        >
          <el-form-item label="原密码" prop="oldPassword">
            <el-input
              type="password"
              v-model="passwordForm.oldPassword"
              placeholder="请输入原密码"
            />
          </el-form-item>
          <el-form-item label="新密码" prop="newPassword">
            <el-input
              type="password"
              v-model="passwordForm.newPassword"
              placeholder="请输入新密码"
            />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              type="password"
              v-model="passwordForm.confirmPassword"
              placeholder="请确认新密码"
            />
          </el-form-item>
        </el-form>
        <div class="form-actions">
          <el-button type="primary" @click="handleUpdatePassword">修改密码</el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/modules/user'
import { getUserInfo, setSelfInfo, changePassword } from '@/api/modules/user'

defineOptions({
  name: 'ProfilePage',
})

interface UserProfile {
  id: string
  username: string
  nickname: string
  email: string
  phone: string
  role: string
  status: number
  createdAt: string
}

const userStore = useUserStore()

// 加载状态
const loading = ref(false)
// 表单引用
const profileFormRef = ref()
const passwordFormRef = ref()

// 个人资料表单数据
const profileForm = reactive<UserProfile>({
  id: '',
  username: '',
  nickname: '',
  email: '',
  phone: '',
  role: '',
  status: 1,
  createdAt: '',
})

// 状态标签
const statusLabel = computed(() => {
  return profileForm.status === 1 ? '启用' : '禁用'
})

// 表单验证规则
const formRules = reactive({
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' } as const,
  ],
  phone: [{ pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号码', trigger: 'blur' }],
})

// 密码表单数据
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// 密码验证规则
const passwordRules = reactive({
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule: unknown, value: string, callback: (error?: string | Error) => void) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
})

// 获取用户信息
const loadUserProfile = async () => {
  loading.value = true
  try {
    const res = await getUserInfo()
    if (res.data?.userInfo) {
      const userInfo = res.data.userInfo
      Object.assign(profileForm, {
        id: String(userInfo.id || ''),
        username: userInfo.username || '',
        nickname: userInfo.nickname || '',
        email: userInfo.email || '',
        phone: userInfo.phone || '',
        role: userInfo.role || '',
        status: userInfo.status || 1,
        createdAt: userInfo.createdAt || '',
      })
    }
  } catch (error) {
    ElMessage.error('获取用户信息失败')
    console.error('获取用户信息失败:', error)
  } finally {
    loading.value = false
  }
}

// 更新个人资料
const handleUpdateProfile = async () => {
  try {
    await profileFormRef.value.validate()
    const updateData = {
      nickname: profileForm.nickname,
      email: profileForm.email,
      phone: profileForm.phone,
    }
    await setSelfInfo(updateData)
    ElMessage.success('个人资料更新成功')
    // 更新用户存储
    await userStore.GetUserInfo()
  } catch (error) {
    console.error('更新个人资料失败:', error)
  }
}

// 更新密码
const handleUpdatePassword = async () => {
  try {
    await passwordFormRef.value.validate()
    const updateData = {
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    }
    await changePassword(updateData)
    ElMessage.success('密码更新成功')
    // 清空密码表单
    Object.assign(passwordForm, {
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
    })
  } catch (error) {
    console.error('更新密码失败:', error)
  }
}

onMounted(() => {
  loadUserProfile()
})
</script>

<style scoped>
.profile-page {
  background-color: var(--bg-color);
  padding: 20px;
  min-height: calc(100vh - 60px);
}

.profile-content {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  font-weight: bold;
  font-size: 16px;
  color: var(--text-primary);
}

.profile-form,
.password-form {
  margin-top: 20px;
}

.form-actions {
  margin-top: 30px;
  text-align: right;
}

/* 暗黑模式适配 */
html.dark :deep(.profile-page) {
  .el-card {
    background-color: var(--bg-dark);
    color: var(--text-primary);

    .el-card__header {
      border-bottom-color: var(--border-color);
    }
  }

  .el-form {
    .el-form-item__label {
      color: var(--text-primary);
    }

    .el-input__wrapper {
      background-color: var(--bg-color);
      border-color: var(--border-color);

      .el-input__inner {
        color: var(--text-primary);
      }

      &:hover {
        border-color: var(--border-hover);
      }

      &.is-focus {
        border-color: var(--primary-color);
        box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
      }
    }
  }
}
</style>
