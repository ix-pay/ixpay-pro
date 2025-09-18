<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="login-bg-pattern"></div>

    <!-- 登录卡片 -->
    <div class="login-box">
      <!-- 登录头部 -->
      <div class="login-header">
        <div class="login-logo">
          <img src="@/assets/ixpay.png" alt="IxPay Pro Logo" class="login-logo-image" />
        </div>
        <h1>欢迎登录</h1>
        <p class="login-subtitle">IxPay Pro 支付管理系统</p>
      </div>

      <!-- 登录表单 -->
      <div class="login-content">
        <el-form ref="formRef" :model="formData" :rules="rules" class="login-form" autocomplete="on">
          <!-- 用户名输入框 -->
          <el-form-item prop="username" class="login-form-item">
            <el-input v-model="formData.username" placeholder="请输入用户名" prefix-icon="User" size="large" clearable
              :validate-event="false" />
          </el-form-item>

          <!-- 密码输入框 -->
          <el-form-item prop="password" class="login-form-item">
            <el-input v-model="formData.password" placeholder="请输入密码" type="password" prefix-icon="Key" show-password
              size="large" clearable :validate-event="false" />
          </el-form-item>

          <!-- 验证码输入框 -->
          <el-form-item v-if="formData.openCaptcha" prop="captcha" class="login-form-item">
            <div class="captcha-container">
              <el-input v-model="formData.captcha" placeholder="请输入验证码" maxLength="4" prefix-icon="Lock" size="large"
                class="captcha-input" clearable :validate-event="false" />
              <div class="captcha-image-wrapper">
                <img v-if="picPath" class="captcha-image" :src="picPath" alt="验证码" @click="refreshCaptcha"
                  title="点击刷新验证码" :class="{ 'shaking': isRefreshing }" />
                <div v-else class="captcha-placeholder" @click="refreshCaptcha">
                  <el-icon class="refresh-icon">
                    <component is="Refresh" :class="{ 'rotating': isRefreshing }" />
                  </el-icon>
                </div>
              </div>
            </div>
          </el-form-item>

          <!-- 登录选项 -->
          <div class="login-options">
            <el-checkbox v-model="formData.rememberMe" size="small">记住密码</el-checkbox>
            <el-link type="primary" :underline="false" size="small" @click="handleForgotPassword">
              忘记密码？
            </el-link>
          </div>

          <!-- 登录按钮 -->
          <el-form-item class="login-form-item login-button-group">
            <el-button type="primary" @click="handleLogin" :loading="isLoading" :disabled="isLoading" size="large"
              class="w-full">
              登录
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch, nextTick } from 'vue'
import { useUserStore } from '@/pinia/modules/user'
import { ElMessage, ElMessageBox } from 'element-plus'
import { captcha } from '@/api/user'

// 表单引用和状态管理
const formRef = ref(null)
const picPath = ref('')
const isLoading = ref(false)
const isRefreshing = ref(false) // 验证码刷新动画状态

// 表单数据
const formData = reactive({
  username: 'admin',
  password: '',
  captcha: '',
  captchaId: '',
  openCaptcha: false,
  rememberMe: false
})

// 表单验证规则
const rules = reactive({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度应在2-20个字符之间', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 32, message: '密码长度应在6-32个字符之间', trigger: 'blur' }
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { message: '验证码格式不正确', trigger: 'blur' }
  ]
})

// 刷新验证码
const refreshCaptcha = async () => {
  // 防止重复点击
  if (isRefreshing.value) return

  isRefreshing.value = true

  try {
    // 触发刷新动画
    await nextTick()

    const res = await captcha()

    // 更新验证码规则
    rules.captcha = [
      { required: true, message: '请输入验证码', trigger: 'blur' },
      {
        max: res.data.captchaLength,
        min: res.data.captchaLength,
        message: `请输入${res.data.captchaLength}位验证码`,
        trigger: 'blur',
      }
    ]

    // 更新验证码图片和ID
    picPath.value = res.data.picPath
    formData.captchaId = res.data.captchaId
    formData.openCaptcha = res.data.openCaptcha
    formData.captcha = '' // 清空验证码输入框

    // 重置表单验证状态
    if (formRef.value) {
      formRef.value.clearValidate('captcha')
    }
  } catch (error) {
    ElMessage({
      type: 'error',
      message: '获取验证码失败，请重试',
      showClose: true,
      duration: 3000
    })
    console.error('获取验证码失败:', error)
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
  // 触发表单验证
  try {
    await formRef.value.validate()
  } catch (error) {
    ElMessage({
      type: 'error',
      message: '请正确填写登录信息',
      showClose: true,
      duration: 3000
    })
    return false
  }

  // 显示加载状态
  isLoading.value = true

  try {
    // 提交登录请求
    const flag = await userStore.LoginIn(formData)

    // 登录失败处理
    if (!flag) {
      // ElMessage({
      //   type: 'error',
      //   message: '用户名或密码错误，请重试',
      //   showClose: true,
      //   duration: 3000
      // })
      await refreshCaptcha()
      return false
    }

    // 登录成功处理
    ElMessage({
      type: 'success',
      message: '登录成功，正在跳转...',
      showClose: true,
      duration: 2000
    })

    return true
  } catch (error) {
    ElMessage({
      type: 'error',
      message: '登录失败，请重试',
      showClose: true,
      duration: 3000
    })
    console.error('登录失败:', error)
    await refreshCaptcha()
    return false
  } finally {
    // 隐藏加载状态
    isLoading.value = false
  }
}

// 处理忘记密码
const handleForgotPassword = () => {
  ElMessageBox.alert(
    '请联系系统管理员重置密码',
    '忘记密码',
    {
      confirmButtonText: '确定',
      type: 'info',
      customClass: 'forgot-password-dialog'
    }
  )
}

// 保存记住的密码
const saveRememberedPassword = () => {
  if (formData.rememberMe) {
    // 记住密码
    localStorage.setItem('rememberMe', 'true')
    localStorage.setItem('rememberedUsername', formData.username)
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
    formData.username = savedUsername
    formData.password = savedPassword
    formData.rememberMe = true
  }

  // 加载验证码
  refreshCaptcha()
})

// 监听记住密码状态变化
watch(() => formData.rememberMe, () => {
  saveRememberedPassword()
})

// 监听用户名和密码变化，当记住密码开启时更新存储
watch(() => [formData.username, formData.password], () => {
  if (formData.rememberMe) {
    saveRememberedPassword()
  }
})
</script>

<style lang="scss" scoped>
// 登录容器样式
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
  animation: gradientBG 15s ease infinite;
}

// 背景图案
.login-bg-pattern {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
  z-index: 0;
}

// 登录卡片
.login-box {
  background-color: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.18);
  width: 100%;
  max-width: 420px;
  padding: 2.5rem;
  z-index: 1;
  transition: all 0.3s ease;
  animation: fadeIn 0.6s ease-out;

  &:hover {
    box-shadow: 0 12px 48px rgba(0, 0, 0, 0.18);
    transform: translateY(-2px);
  }
}

// 登录头部
.login-header {
  text-align: center;
  margin-bottom: 2.5rem;
}

// 登录Logo
.login-logo {
  margin-bottom: 1.5rem;
  display: flex;
  justify-content: center;
}

.login-logo-image {
  width: 80px;
  height: 80px;
  object-fit: contain;
  animation: pulse 2s infinite;
}

// 登录标题
.login-header h1 {
  margin: 0 0 0.75rem 0;
  font-size: 2rem;
  font-weight: 600;
  color: #333;
  background: linear-gradient(45deg, #667eea, #764ba2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

// 登录副标题
.login-subtitle {
  color: #666;
  font-size: 1rem;
  margin: 0;
  opacity: 0.8;
}

// 登录内容
.login-content {
  width: 100%;
}

// 登录表单
.login-form {
  width: 100%;
}

// 表单项
.login-form-item {
  margin-bottom: 1.25rem;
  position: relative;
}

// 登录选项
.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.75rem;
}

// 验证码容器
.captcha-container {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.captcha-input {
  flex: 1;
}

// 验证码图片
.captcha-image-wrapper {
  width: 110px;
  height: 44px;
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s ease;
  background-color: #f8f9fa;

  &:hover {
    transform: scale(1.02);
  }
}

.captcha-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

// 验证码占位符
.captcha-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  background-color: #f8f9fa;

  &:hover {
    background-color: #e9ecef;
  }
}

// 刷新图标
.refresh-icon {
  transition: transform 0.3s ease;
}

// 按钮组
.login-button-group {
  margin-bottom: 0;

  .el-button {
    width: 100%;
  }
}

// 暗黑模式适配
:deep(.dark) {
  .login-box {
    background-color: rgba(30, 41, 59, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .login-header h1 {
    color: #e2e8f0;
    background: linear-gradient(45deg, #94a3b8, #cbd5e1);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .login-subtitle {
    color: #94a3b8;
  }

  .captcha-image-wrapper,
  .captcha-placeholder {
    background-color: #334155;
  }

  .captcha-placeholder:hover {
    background-color: #475569;
  }
}

// 响应式设计
@media (max-width: 480px) {
  .login-box {
    margin: 1.5rem;
    padding: 2rem;
    border-radius: 12px;
  }

  .login-header h1 {
    font-size: 1.75rem;
  }

  .captcha-container {
    flex-direction: column;
    align-items: stretch;
  }

  .captcha-image-wrapper {
    width: 100%;
    height: 40px;
  }
}

// 加载动画
:deep(.el-loading-mask) {
  background-color: rgba(255, 255, 255, 0.8);
}

:deep(.dark) .el-loading-mask {
  background-color: rgba(30, 41, 59, 0.8);
}

// 自定义对话框样式
:deep(.forgot-password-dialog) {
  border-radius: 8px !important;
}

// 动画效果
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes gradientBG {
  0% {
    background-position: 0% 50%;
  }

  50% {
    background-position: 100% 50%;
  }

  100% {
    background-position: 0% 50%;
  }
}

@keyframes pulse {
  0% {
    transform: scale(1);
  }

  50% {
    transform: scale(1.05);
  }

  100% {
    transform: scale(1);
  }
}

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

// 验证码刷新动画
.shaking {
  animation: shake 0.5s ease-in-out;
}

.rotating {
  animation: rotate 1s linear;
}
</style>