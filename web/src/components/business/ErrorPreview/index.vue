<template>
  <el-dialog
    v-model="dialogVisible"
    :title="''"
    :close-on-click-modal="true"
    :close-on-press-escape="true"
    :width="600"
    center
    :show-close="false"
    header-position="none"
    @close="closeModal"
  >
    <!-- 弹窗内容 -->
    <div class="error-preview-container">
      <!-- 错误图标 -->
      <el-card
        :body-style="{
          padding: '30px',
          borderRadius: '50%',
          width: '120px',
          height: '120px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          margin: '0 auto 24px',
        }"
        :shadow="'hover'"
        :class="`error-icon-card-${displayData.icon}`"
      >
        <el-icon
          :size="64"
          :color="getIconColor()"
          :class="props.errorData.code === 401 ? 'pulse-animation' : ''"
        >
          <Lock v-if="displayData.icon === 'lock'" />
          <Warning v-else-if="displayData.icon === 'warn'" />
          <CircleCloseFilled v-else-if="displayData.icon === 'server'" />
          <QuestionFilled v-else />
        </el-icon>
      </el-card>

      <!-- 错误标题和类型 -->
      <el-card :body-style="{ padding: '24px' }" shadow="hover">
        <div class="error-content">
          <h2 class="error-title">{{ displayData.title }}</h2>
          <el-tag :type="getTagType()" size="large" style="margin: 12px 0">{{
            displayData.type
          }}</el-tag>

          <!-- 错误信息 -->
          <el-alert
            :title="displayData.message"
            type="info"
            :closable="false"
            style="margin: 16px 0"
            effect="light"
          />

          <!-- 提示信息 -->
          <el-alert
            v-if="displayData.tips"
            :title="displayData.tips"
            type="info"
            :closable="false"
            effect="light"
          >
            <template #icon>
              <el-icon :size="16" color="#1890ff">
                <InfoFilled />
              </el-icon>
            </template>
          </el-alert>
        </div>
      </el-card>

      <!-- 弹窗底部 -->
      <div style="margin-top: 24px; display: flex; justify-content: center">
        <el-button type="primary" size="large" @click="handleConfirm" round style="width: 160px">
          确定
        </el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  Lock,
  Warning,
  QuestionFilled,
  CircleCloseFilled,
  InfoFilled,
} from '@element-plus/icons-vue'

defineOptions({
  name: 'ErrorPreview',
})

interface ErrorData {
  code: number | string
  message?: string
}

const props = defineProps({
  errorData: {
    type: Object as () => ErrorData,
    required: true,
  },
})

const emits = defineEmits(['close', 'confirm'])

// 控制弹窗显示
const dialogVisible = ref(true)

// 监听props变化
watch(
  () => props.errorData,
  () => {
    dialogVisible.value = true
  },
  { deep: true },
)

interface PresetError {
  title: string
  type: string
  icon: string
  color: string
  tips: string
}

const presetErrors: Record<number | string, PresetError> = {
  500: {
    title: '检测到接口错误',
    type: '服务器发生内部错误',
    icon: 'server',
    color: 'var(--error-color)',
    tips: '此类错误内容常见于后台panic，请先查看后台日志，如果影响您正常使用可强制登出清理缓存',
  },
  404: {
    title: '资源未找到',
    type: 'Not Found',
    icon: 'warn',
    color: 'var(--warning-color)',
    tips: '此类错误多为接口未注册（或未重启）或者请求路径（方法）与api路径（方法）不符--如果为自动化代码请检查是否存在空格',
  },
  401: {
    title: '身份认证失败',
    type: '身份令牌无效',
    icon: 'lock',
    color: 'var(--error-color)',
    tips: '您的身份认证已过期或无效，请重新登录。',
  },
  network: {
    title: '网络错误',
    type: 'Network Error',
    icon: 'server',
    color: 'var(--info-color)',
    tips: '无法连接到服务器，请检查您的网络连接。',
  },
}

const displayData = computed(() => {
  const preset = presetErrors[props.errorData.code]
  if (preset) {
    return {
      ...preset,
      message: props.errorData.message || '没有提供额外信息。',
    }
  }

  return {
    title: '未知错误',
    type: '检测到请求错误',
    icon: 'question',
    color: 'el-color-info',
    message: props.errorData.message || '发生了一个未知错误。',
    tips: '请检查控制台获取更多信息。',
  }
})

// 根据错误类型获取图标颜色
const getIconColor = () => {
  switch (displayData.value.icon) {
    case 'lock':
      return 'var(--error-color)'
    case 'warn':
      return 'var(--warning-color)'
    case 'server':
      return 'var(--error-color)'
    default:
      return 'var(--primary-color)'
  }
}

// 根据错误类型获取标签类型
const getTagType = () => {
  switch (displayData.value.icon) {
    case 'lock':
      return 'danger'
    case 'warn':
      return 'warning'
    case 'server':
      return 'danger'
    default:
      return 'info'
  }
}

const closeModal = () => {
  dialogVisible.value = false
  // 延迟触发close事件，确保弹窗动画完成
  setTimeout(() => {
    emits('close')
  }, 300)
}

const handleConfirm = () => {
  emits('confirm', props.errorData.code)
  closeModal()
}
</script>

<style lang="scss" scoped>
// 错误预览容器
.error-preview-container {
  text-align: center;
}

// 错误标题
.error-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 8px;
}

// 错误内容区域
.error-content {
  text-align: center;
}

// 不同错误类型的图标卡片样式
.error-icon-card-lock {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.1), rgba(239, 68, 68, 0.05));
  border-color: rgba(239, 68, 68, 0.2);
}

.error-icon-card-warn {
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.1), rgba(251, 191, 36, 0.05));
  border-color: rgba(251, 191, 36, 0.2);
}

.error-icon-card-server {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.1), rgba(239, 68, 68, 0.05));
  border-color: rgba(239, 68, 68, 0.2);
}

.error-icon-card-question {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(59, 130, 246, 0.05));
  border-color: rgba(59, 130, 246, 0.2);
}

// 脉冲动画效果
.pulse-animation {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

// 暗黑模式下的样式调整
html.dark :deep(.error-preview-container) {
  .error-content {
    color: var(--text-primary);
  }

  // 调整卡片样式
  .el-card {
    background-color: var(--bg-secondary);
    border-color: var(--border-primary);
  }

  // 调整 alert 组件样式
  .el-alert {
    background-color: var(--bg-primary);
    border-color: var(--border-primary);
    color: var(--text-primary);
  }

  // 调整图标卡片透明度
  .error-icon-card-lock,
  .error-icon-card-warn,
  .error-icon-card-server,
  .error-icon-card-question {
    background: rgba(0, 0, 0, 0.2);
    border-color: rgba(255, 255, 255, 0.1);
  }
}
</style>
