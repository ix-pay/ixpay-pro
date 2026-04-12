<template>
  <div
    class="w-full h-screen flex items-center justify-center p-4 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800"
  >
    <el-card class="error-card" shadow="hover">
      <div class="flex flex-col items-center gap-8 p-6">
        <div class="error-image-wrapper relative">
          <img class="error-image" src="../../../assets/images/404.png" />
          <div
            class="absolute -top-4 -right-4 bg-primary-500 text-white text-2xl font-bold rounded-full w-12 h-12 flex items-center justify-center shadow-lg"
          >
            404
          </div>
        </div>
        <div class="text-center max-w-md">
          <h1
            class="error-title text-4xl font-bold bg-gradient-to-r from-primary-600 to-indigo-600 bg-clip-text text-transparent mb-4"
          >
            页面不存在
          </h1>
          <p
            class="error-description text-gray-600 dark:text-gray-300 text-lg leading-relaxed mb-6"
          >
            抱歉，您访问的页面可能已被移动、删除或不存在。
            常见原因包括：URL输入错误、页面已被移除或您的角色没有访问权限。
          </p>
          <div class="error-link mb-8">
            <span class="text-gray-500 dark:text-gray-400">项目地址：</span>
            <a
              href="https://github.com/flipped-aurora/gin-vue-admin"
              target="_blank"
              class="text-primary-600 hover:text-primary-800 dark:text-primary-400 dark:hover:text-primary-300 transition-all duration-300 font-medium"
            >
              https://github.com/flipped-aurora/gin-vue-admin
            </a>
          </div>
        </div>
        <div class="error-action flex flex-wrap gap-4 justify-center">
          <el-button
            type="primary"
            size="large"
            @click="toDashboard"
            class="px-8 py-2 rounded-lg shadow-md hover:shadow-lg transition-all duration-300"
          >
            <el-icon class="mr-2"><House /></el-icon>
            返回首页
          </el-button>
          <el-button
            size="large"
            @click="goBack"
            class="px-8 py-2 rounded-lg transition-all duration-300"
          >
            <el-icon class="mr-2"><ArrowLeft /></el-icon>
            返回上一页
          </el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores'
import { useRouter } from 'vue-router'
import { emitter } from '@/utils/bus'
import { House, ArrowLeft } from '@element-plus/icons-vue'

defineOptions({
  name: 'ErrorIndex',
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

<style scoped>
.error-card {
  max-width: 700px;
  width: 100%;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: white;
  border: none;
}

.error-card:hover {
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.15);
  transform: translateY(-5px);
}

.error-image-wrapper {
  width: 100%;
  text-align: center;
  padding: 20px 0;
  position: relative;
}

.error-image {
  max-width: 240px;
  height: auto;
  opacity: 0.9;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  filter: drop-shadow(0 10px 20px rgba(0, 0, 0, 0.1));
}

.error-image:hover {
  opacity: 1;
  transform: scale(1.08);
}

.error-title {
  font-size: 36px;
  font-weight: 800;
  margin-bottom: 16px;
  letter-spacing: -0.5px;
}

.error-description {
  font-size: 16px;
  line-height: 1.7;
  margin-bottom: 24px;
}

.error-link {
  font-size: 14px;
  margin-bottom: 32px;
}

.error-link a {
  text-decoration: none;
  transition: all 0.3s ease;
  position: relative;
}

.error-link a::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 0;
  height: 2px;
  background: currentColor;
  transition: width 0.3s ease;
}

.error-link a:hover::after {
  width: 100%;
}

.error-action {
  padding: 8px 0 0;
}

/* 暗黑模式支持 */
.dark .error-card {
  background: #1f2937;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.dark .error-card:hover {
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.4);
}

@media (max-width: 768px) {
  .error-card {
    margin: 0 16px;
    padding: 0 16px;
  }

  .error-title {
    font-size: 28px;
  }

  .error-description {
    font-size: 15px;
  }

  .error-image {
    max-width: 200px;
  }

  .error-action {
    flex-direction: column;
    align-items: center;
  }

  .error-action .el-button {
    width: 100%;
    max-width: 280px;
  }
}

@media (max-width: 480px) {
  .error-title {
    font-size: 24px;
  }

  .error-description {
    font-size: 14px;
  }

  .error-image {
    max-width: 180px;
  }
}
</style>
