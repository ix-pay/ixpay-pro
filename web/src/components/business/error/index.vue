<template>
  <div class="ix-error-page">
    <div class="ix-error-card">
      <div class="ix-error-content">
        <!-- 错误图标容器 -->
        <div class="ix-error-image-wrapper">
          <img class="ix-error-image" src="../../../assets/images/404.png" />
          <div class="ix-error-badge">404</div>
        </div>

        <!-- 错误信息 -->
        <div class="text-center max-w-md">
          <h1 class="ix-error-title">页面不存在</h1>
          <p class="ix-error-description">
            抱歉，您访问的页面可能已被移动、删除或不存在。 常见原因包括：URL
            输入错误、页面已被移除或您的角色没有访问权限。
          </p>
          <div class="ix-error-link">
            <span class="text-gray-500 dark:text-gray-400">项目地址：</span>
            <a href="https://github.com/flipped-aurora/gin-vue-admin" target="_blank">
              https://github.com/flipped-aurora/gin-vue-admin
            </a>
          </div>
        </div>

        <!-- 操作按钮组 -->
        <div class="ix-error-actions">
          <el-button type="primary" size="large" @click="toDashboard" class="px-8 py-2">
            <el-icon class="mr-2"><House /></el-icon>
            返回首页
          </el-button>
          <el-button size="large" @click="goBack" class="px-8 py-2">
            <el-icon class="mr-2"><ArrowLeft /></el-icon>
            返回上一页
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores'
import { useRouter } from 'vue-router'
import { emitter } from '@/utils/bus'
import { House, ArrowLeft } from '@element-plus/icons-vue'

defineOptions({
  name: 'IxError',
})

const userStore = useUserStore()
const router = useRouter()

// 返回上一页方法
const goBack = () => {
  router.go(-1)
}

const toDashboard = () => {
  try {
    // 检查 defaultRouter 是否存在，如果不存在则使用默认路由
    const defaultRouteName = userStore.userInfo.authority?.defaultRouter || 'index'
    router.push({ name: defaultRouteName })
  } catch {
    emitter.emit('show-error', {
      code: '401',
      message: '检测到其他用户修改了路由权限，请重新登录',
      fn: () => {
        userStore.ClearStorage()
        router.push({ name: 'Login', replace: true })
      },
    })
  }
}
</script>
