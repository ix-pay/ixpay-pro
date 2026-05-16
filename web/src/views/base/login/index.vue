<template>
  <div
    class="login-container min-h-screen flex items-center justify-center relative overflow-hidden bg-gradient-to-br from-primary via-purple-500 to-pink-500 dark:from-primary-dark dark:via-purple-900 dark:to-pink-900"
  >
    <!-- 背景动画装饰 -->
    <div class="absolute inset-0 overflow-hidden">
      <div
        class="floating-shape absolute top-1/4 left-1/4 w-96 h-96 bg-white/10 dark:bg-white/5 rounded-full blur-3xl animate-float"
      ></div>
      <div
        class="floating-shape absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-500/20 dark:bg-blue-600/10 rounded-full blur-3xl animate-float delay-1000"
      ></div>
      <div
        class="floating-shape absolute top-1/3 right-1/3 w-64 h-64 bg-purple-500/20 dark:bg-purple-600/10 rounded-full blur-2xl animate-float delay-500"
      ></div>
    </div>

    <!-- 登录卡片 -->
    <div class="relative z-10 w-full max-w-md mx-4 animate-fade-in-up">
      <div
        class="login-card backdrop-blur-xl bg-white/90 dark:bg-gray-800/90 rounded-3xl shadow-2xl p-8 border border-white/20 dark:border-gray-700/50 hover:shadow-primary-glow transition-all duration-500"
      >
        <!-- Logo 和标题 -->
        <div class="text-center mb-8">
          <div class="flex justify-center mb-6">
            <div
              class="logo-wrapper relative w-20 h-20 rounded-2xl shadow-lg overflow-hidden hover:scale-110 transition-transform duration-300 group"
            >
              <img
                :src="logoImage"
                alt="IxPay Pro Logo"
                class="w-full h-full object-cover group-hover:rotate-6 transition-transform duration-300"
              />
              <div
                class="absolute inset-0 bg-gradient-to-br from-primary/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"
              ></div>
            </div>
          </div>
          <h1
            class="text-3xl font-bold bg-gradient-to-r from-primary to-purple-600 bg-clip-text text-transparent mb-2 animate-fade-in"
          >
            欢迎登录{{ $IXPAY_PRO.appName }}
          </h1>
          <p class="text-text-secondary dark:text-gray-400 text-sm">企业级支付管理系统</p>
        </div>

        <!-- 登录表单 -->
        <el-form
          ref="formRef"
          :model="formData"
          :rules="rules"
          class="space-y-5"
          autocomplete="on"
          @submit.prevent="handleLogin"
        >
          <!-- 用户名输入 -->
          <el-form-item prop="userName">
            <el-input
              v-model="formData.userName"
              placeholder="请输入用户名"
              :prefix-icon="User"
              size="large"
              clearable
              :validate-event="false"
              class="input-field rounded-xl"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <!-- 密码输入 -->
          <el-form-item prop="password">
            <el-input
              v-model="formData.password"
              placeholder="请输入密码"
              type="password"
              :prefix-icon="Key"
              show-password
              size="large"
              clearable
              :validate-event="false"
              class="input-field rounded-xl"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <!-- 验证码输入 -->
          <el-form-item v-if="formData.openCaptcha" prop="captcha">
            <div class="flex items-center gap-2">
              <el-input
                v-model="formData.captcha"
                placeholder="请输入验证码"
                maxLength="4"
                :prefix-icon="Lock"
                size="large"
                class="flex-1 input-field rounded-xl"
                clearable
                :validate-event="false"
                @keyup.enter="handleLogin"
              />
              <div
                class="captcha-wrapper w-28 h-11 cursor-pointer rounded-xl overflow-hidden transition-all duration-300 hover:scale-105 hover:shadow-md"
                @click="refreshCaptcha"
              >
                <img
                  v-if="picPath"
                  class="w-full h-full object-cover"
                  :src="picPath"
                  alt="验证码"
                  title="点击刷新验证码"
                  :class="{ shaking: isRefreshing }"
                />
                <div
                  v-else
                  class="w-full h-full flex items-center justify-center bg-bg-tertiary dark:bg-gray-700"
                >
                  <el-icon class="refresh-icon" :class="{ rotating: isRefreshing }">
                    <Refresh />
                  </el-icon>
                </div>
              </div>
            </div>
          </el-form-item>

          <!-- 登录选项 -->
          <div class="flex items-center justify-between">
            <el-checkbox v-model="formData.rememberMe" size="small" class="text-text-secondary">
              记住密码
            </el-checkbox>
            <el-link
              type="primary"
              :underline="false"
              size="small"
              class="hover:text-primary-light transition-colors"
              @click="handleForgotPassword"
            >
              忘记密码？
            </el-link>
          </div>

          <!-- 登录按钮 -->
          <el-form-item>
            <el-button
              type="primary"
              @click="handleLogin"
              :loading="isLoading"
              :disabled="isLoading"
              size="large"
              class="login-button w-full rounded-xl shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-0.5"
            >
              <span v-if="!isLoading">登录系统</span>
              <span v-else>登录中...</span>
            </el-button>
          </el-form-item>
        </el-form>

        <!-- 页脚 -->
        <div class="mt-6 text-center text-xs text-text-tertiary dark:text-gray-400">
          <span>{{ $IXPAY_PRO.appName }} © 2023 - 2025 | 企业级支付管理系统</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted, watch, nextTick } from 'vue'
import { useUserStore } from '@/stores/index'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { captcha } from '@/api/modules/user'
import { User, Key, Lock, Refresh } from '@element-plus/icons-vue'
import logoImage from '@/assets/images/ixpay.png'

defineOptions({
  name: 'BaseLogin',
})

// 表单引用和状态管理
const formRef = ref<FormInstance | null>(null)
const picPath = ref('')
const isLoading = ref(false)
const isRefreshing = ref(false)

// 表单数据
const formData = reactive({
  userName: 'admin',
  password: '',
  captcha: '',
  captchaId: '',
  openCaptcha: false,
  rememberMe: false,
})

// 表单验证规则
const rules = reactive({
  userName: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度应在 2-20 个字符之间', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 32, message: '密码长度应在 6-32 个字符之间', trigger: 'blur' },
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { message: '验证码格式不正确', trigger: 'blur' },
  ],
})

// 刷新验证码
const refreshCaptcha = async () => {
  if (isRefreshing.value) return

  isRefreshing.value = true

  try {
    await nextTick()

    const captchaData = await captcha()
    console.log('验证码数据:', captchaData)

    rules.captcha = [{ required: true, message: '请输入验证码', trigger: 'blur' }]

    picPath.value = captchaData?.data?.picPath || ''
    formData.captchaId = captchaData?.data?.captchaId || ''
    formData.openCaptcha =
      captchaData?.data?.openCaptcha === undefined ? true : captchaData.data.openCaptcha
    formData.captcha = ''

    if (formRef.value) {
      formRef.value.clearValidate('captcha')
    }
  } catch (err) {
    ElMessage({
      type: 'error',
      message: '获取验证码失败，请重试',
      showClose: true,
      duration: 3000,
    })
    console.error('获取验证码失败:', err)
  } finally {
    setTimeout(() => {
      isRefreshing.value = false
    }, 500)
  }
}

// 处理登录
const userStore = useUserStore()
const handleLogin = async () => {
  console.log('Login Component - handleLogin function called')

  try {
    if (!formRef.value) {
      console.log('Login Component - formRef.value is null')
      return false
    }
    console.log('Login Component - Calling form validation')
    await formRef.value.validate()
    console.log('Login Component - Form validation passed')
  } catch (error) {
    console.log('Login Component - Form validation failed:', error)
    ElMessage({
      type: 'error',
      message: '请正确填写登录信息',
      showClose: true,
      duration: 3000,
    })
    return false
  }

  isLoading.value = true

  try {
    console.log('Login Component - Calling userStore.LoginIn with formData:', formData)
    const flag = await userStore.LoginIn(formData)
    console.log('Login Component - userStore.LoginIn returned:', flag)

    if (!flag) {
      console.log('Login Component - Login failed, refreshing captcha')
      await refreshCaptcha()
      return false
    }

    console.log('Login Component - Login succeeded, showing success message')
    ElMessage({
      type: 'success',
      message: '登录成功，正在跳转...',
      showClose: true,
      duration: 2000,
    })

    return true
  } catch (error) {
    console.log('Login Component - Exception in handleLogin:', error)
    ElMessage({
      type: 'error',
      message: '登录失败，请重试',
      showClose: true,
      duration: 3000,
    })
    console.error('登录失败:', error)
    await refreshCaptcha()
    return false
  } finally {
    console.log('Login Component - Setting isLoading to false')
    isLoading.value = false
  }
}

// 处理忘记密码
const handleForgotPassword = () => {
  ElMessageBox.alert('请联系系统管理员重置密码', '忘记密码', {
    confirmButtonText: '确定',
    type: 'info',
    customClass: 'forgot-password-dialog',
  })
}

// 保存记住的密码
const saveRememberedPassword = () => {
  if (formData.rememberMe) {
    localStorage.setItem('rememberMe', 'true')
    localStorage.setItem('rememberedUsername', formData.userName)
    localStorage.setItem('rememberedPassword', formData.password)
  } else {
    localStorage.removeItem('rememberMe')
    localStorage.removeItem('rememberedUsername')
    localStorage.removeItem('rememberedPassword')
  }
}

// 页面加载时执行
onMounted(() => {
  const savedUsername = localStorage.getItem('rememberedUsername')
  const savedPassword = localStorage.getItem('rememberedPassword')
  const rememberMe = localStorage.getItem('rememberMe') === 'true'

  if (rememberMe && savedUsername && savedPassword) {
    formData.userName = savedUsername
    formData.password = savedPassword
    formData.rememberMe = true
  }

  refreshCaptcha()
})

// 监听记住密码状态变化
watch(
  () => formData.rememberMe,
  () => {
    saveRememberedPassword()
  },
)

// 监听用户名和密码变化
watch(
  () => [formData.userName, formData.password],
  () => {
    if (formData.rememberMe) {
      saveRememberedPassword()
    }
  },
)
</script>

<style scoped>
/* 登录容器 */
.login-container {
  --floating-animation: float 3s ease-in-out infinite;
}

/* 输入框样式增强 */
.input-field :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--border-primary) inset !important;
  transition: all var(--duration-normal) var(--ease-in-out);
  background-color: var(--bg-primary);
}

.input-field :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--primary-color) inset !important;
}

.input-field :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 2px var(--primary-glow),
    0 0 0 1px var(--primary-color) inset !important;
}

/* 登录按钮样式 */
.login-button {
  background: var(--primary-gradient) !important;
  border: none !important;
  font-weight: 600;
  letter-spacing: 1px;
}

.login-button:hover {
  opacity: 0.9;
  box-shadow: var(--shadow-primary-glow) !important;
}

.login-button:active {
  opacity: 0.8;
  transform: translateY(0) !important;
}

/* 验证码动画 */
.shaking {
  animation: shake 0.5s ease-in-out;
}

.rotating {
  animation: rotate 1s linear;
}

/* 对话框样式 */
:deep(.forgot-password-dialog) {
  border-radius: var(--radius-lg);
  background: var(--bg-primary);
  backdrop-filter: blur(10px);
}

/* 动画定义 */
@keyframes shake {
  0%,
  100% {
    transform: translateX(0);
  }
  20%,
  60% {
    transform: translateX(-4px);
  }
  40%,
  80% {
    transform: translateX(4px);
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
