<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>登录</h1>
      </div>
      <div class="login-content">
        <el-form ref="formRef" :model="formData" :rules="rules" class="login-form">
          <el-form-item prop="username">
            <el-input v-model="formData.username" placeholder="请输入用户名" />
          </el-form-item>
          <el-form-item prop="password">
            <el-input v-model="formData.password" placeholder="请输入密码" type="password" />
          </el-form-item>
          <el-form-item v-if="formData.openCaptcha" prop="captcha" class="mb-6">
            <div class="flex w-full justify-between">
              <el-input v-model="formData.captcha" placeholder="请输入验证码" size="large" class="flex-1 mr-5" />
              <div class="w-1/3 h-11 bg-[#c3d4f2] rounded">
                <img v-if="picPath" class="w-full h-full" :src="picPath" alt="请输入验证码" @click="loginVerify()" />
              </div>
            </div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleLogin">登录</el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useUserStore } from '@/pinia/modules/user'
import { ElMessage } from 'element-plus'
import { captcha } from '@/api/user'

const formRef = ref(null)
const picPath = ref('')
const formData = reactive({
  username: 'admin',
  password: '',
  captcha: '',
  captchaId: '',
})

const rules = reactive({
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captcha: [
    {
      message: '验证码格式不正确',
      trigger: 'blur',
    },
  ],
})

// 获取验证码
const loginVerify = async () => {
  const ele = await captcha()
  rules.captcha.push({
    max: ele.data.captchaLength,
    min: ele.data.captchaLength,
    message: `请输入${ele.data.captchaLength}位验证码`,
    trigger: 'blur',
  })
  picPath.value = ele.data.picPath
  formData.captchaId = ele.data.captchaId
  formData.openCaptcha = ele.data.openCaptcha
}
loginVerify()

const userStore = useUserStore()
const login = async () => {
  return await userStore.LoginIn(formData)
}
const handleLogin = async () => {
  const v = await formRef.value.validate()
  if (!v) {
    // 未通过前端静态验证
    ElMessage({
      type: 'error',
      message: '请正确填写登录信息',
      showClose: true,
    })
    await loginVerify()
    return false
  }

  // 通过验证，请求登陆
  const flag = await login()

  // 登陆失败，刷新验证码
  if (!flag) {
    await loginVerify()
    return false
  }

  // 登陆成功
  return true
}
</script>
