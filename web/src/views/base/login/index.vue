<template>
  <!-- 使用 Tailwind CSS 重构背景和布局 -->
  <div
    class="min-h-screen flex items-center justify-center relative overflow-hidden bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500"
  >
    <!-- 背景动画装饰 - 使用 Tailwind 实现 -->
    <div class="absolute inset-0">
      <div
        class="absolute top-1/4 left-1/4 w-96 h-96 bg-white/10 rounded-full blur-3xl animate-pulse"
      ></div>
      <div
        class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-500/20 rounded-full blur-3xl animate-pulse delay-1000"
      ></div>
      <div
        class="absolute top-1/3 right-1/3 w-64 h-64 bg-purple-500/20 rounded-full blur-2xl animate-pulse delay-500"
      ></div>
    </div>

    <!-- 登录卡片 - 使用 Tailwind 实现毛玻璃效果 -->
    <div class="relative z-10 w-full max-w-md mx-4">
      <div
        class="backdrop-blur-xl bg-white/90 dark:bg-gray-800/90 rounded-3xl shadow-2xl p-8 border border-white/20"
      >
        <!-- 登录头部 -->
        <div class="text-center mb-8">
          <div class="flex justify-center mb-4">
            <img
              :src="logoImage"
              alt="IxPay Pro Logo"
              class="w-20 h-20 object-contain rounded-xl shadow-lg hover:scale-105 transition-transform duration-300"
            />
          </div>
          <h1
            class="text-3xl font-bold bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent mb-2"
          >
            欢迎登录{{ $IXPAY_PRO.appName }}
          </h1>
          <p class="text-gray-500 dark:text-gray-400 text-sm">企业级支付管理系统</p>
        </div>

        <!-- 登录表单 -->
        <el-form
          ref="formRef"
          :model="formData"
          :rules="rules"
          class="space-y-4"
          autocomplete="on"
          @submit.prevent="handleLogin"
        >
          <!-- 用户名输入框 -->
          <el-form-item prop="userName">
            <el-input
              v-model="formData.userName"
              placeholder="请输入用户名"
              :prefix-icon="User"
              size="large"
              clearable
              :validate-event="false"
              class="rounded-lg"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <!-- 密码输入框 -->
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
              class="rounded-lg"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <!-- 验证码输入框 -->
          <el-form-item v-if="formData.openCaptcha" prop="captcha">
            <div class="flex items-center gap-2">
              <el-input
                v-model="formData.captcha"
                placeholder="请输入验证码"
                maxLength="4"
                :prefix-icon="Lock"
                size="large"
                class="flex-1 rounded-lg"
                clearable
                :validate-event="false"
                @keyup.enter="handleLogin"
              />
              <div
                class="w-28 h-11 cursor-pointer rounded-lg overflow-hidden transition-transform hover:scale-105"
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
                  class="w-full h-full flex items-center justify-center bg-gray-100 dark:bg-gray-700"
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
            <el-checkbox v-model="formData.rememberMe" size="small">记住密码</el-checkbox>
            <el-link type="primary" :underline="false" size="small" @click="handleForgotPassword">
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
              class="w-full rounded-lg shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-0.5"
            >
              登录系统
            </el-button>
          </el-form-item>
        </el-form>

        <!-- 页脚 -->
        <div class="mt-6 text-center text-xs text-gray-500 dark:text-gray-400">
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
import { User, Key, Lock } from '@element-plus/icons-vue'
// 直接在组件中导入logo图片
import logoImage from '@/assets/images/ixpay.png'

defineOptions({
  name: 'BaseLogin',
})

// 表单引用和状态管理
const formRef = ref<FormInstance | null>(null)
const picPath = ref('')
const isLoading = ref(false)
const isRefreshing = ref(false) // 验证码刷新动画状态

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
    { min: 2, max: 20, message: '用户名长度应在2-20个字符之间', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 32, message: '密码长度应在6-32个字符之间', trigger: 'blur' },
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { message: '验证码格式不正确', trigger: 'blur' },
  ],
})

// 刷新验证码
const refreshCaptcha = async () => {
  // 防止重复点击
  if (isRefreshing.value) return

  isRefreshing.value = true

  try {
    // 触发刷新动画
    await nextTick()

    const captchaData = await captcha()
    console.log('验证码数据:', captchaData)

    // 更新验证码规则
    rules.captcha = [{ required: true, message: '请输入验证码', trigger: 'blur' }]

    // 更新验证码图片和ID
    console.log('验证码数据:', captchaData)
    console.log('验证码图片路径:', captchaData?.data?.picPath)

    // 后端已经返回完整的base64图片数据（包含data:image/png;base64,前缀）
    // 直接设置图片路径
    picPath.value = captchaData?.data?.picPath || ''
    console.log('最终验证码图片路径:', picPath.value)
    formData.captchaId = captchaData?.data?.captchaId || ''
    // 根据后端返回值决定是否显示验证码组件
    formData.openCaptcha =
      captchaData?.data?.openCaptcha === undefined ? true : captchaData.data.openCaptcha
    formData.captcha = '' // 清空验证码输入框

    // 重置表单验证状态
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
    // 延迟关闭动画，确保动画效果可见
    setTimeout(() => {
      isRefreshing.value = false
    }, 500)
  }
}

// 处理登录
const userStore = useUserStore()
const handleLogin = async () => {
  console.log('Login Component - handleLogin function called')
  // 触发表单验证
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

  // 显示加载状态
  isLoading.value = true

  try {
    // 提交登录请求
    console.log('Login Component - Calling userStore.LoginIn with formData:', formData)
    const flag = await userStore.LoginIn(formData)
    console.log('Login Component - userStore.LoginIn returned:', flag)

    // 登录失败处理
    if (!flag) {
      console.log('Login Component - Login failed, refreshing captcha')
      await refreshCaptcha()
      return false
    }

    // 登录成功处理
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
    // 隐藏加载状态
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
    // 记住密码
    localStorage.setItem('rememberMe', 'true')
    localStorage.setItem('rememberedUsername', formData.userName)
    localStorage.setItem('rememberedPassword', formData.password)
  } else {
    // 清除记住的密码
    localStorage.removeItem('rememberMe')
    localStorage.removeItem('rememberedUsername')
    localStorage.removeItem('rememberedPassword')
  }
}

// 页面加载时执行
onMounted(() => {
  // 尝试从本地存储恢复记住的密码
  const savedUsername = localStorage.getItem('rememberedUsername')
  const savedPassword = localStorage.getItem('rememberedPassword')
  const rememberMe = localStorage.getItem('rememberMe') === 'true'

  if (rememberMe && savedUsername && savedPassword) {
    formData.userName = savedUsername
    formData.password = savedPassword
    formData.rememberMe = true
  }

  // 加载验证码
  refreshCaptcha()
})

// 监听记住密码状态变化
watch(
  () => formData.rememberMe,
  () => {
    saveRememberedPassword()
  },
)

// 监听用户名和密码变化，当记住密码开启时更新存储
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
/* 
 * 使用 Tailwind CSS 重构后，只需保留少量定制样式
 * 大部分样式已通过 Tailwind 类名实现
 */

/* 验证码刷新动画 */
.shaking {
  animation: shake 0.5s ease-in-out;
}

.rotating {
  animation: rotate 1s linear;
}

/* 自定义对话框样式 */
:deep(.forgot-password-dialog) {
  border-radius: var(--radius-lg);
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
