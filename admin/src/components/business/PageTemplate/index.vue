<template>
  <div class="ix-page min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- 页头区域 -->
    <header v-if="showHeader" class="bg-white dark:bg-gray-800 shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <!-- 左侧：标题和面包屑 -->
          <div class="flex items-center gap-4">
            <!-- 面包屑 -->
            <el-breadcrumb v-if="showBreadcrumb && breadcrumbItems.length > 0" separator="/">
              <el-breadcrumb-item
                v-for="(item, index) in breadcrumbItems"
                :key="index"
                :to="item.to"
              >
                {{ item.name }}
              </el-breadcrumb-item>
            </el-breadcrumb>

            <!-- 页面标题 -->
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ title }}
            </h1>
          </div>

          <!-- 右侧：操作区 -->
          <div class="flex items-center gap-3">
            <!-- 刷新按钮 -->
            <el-button
              v-if="showRefresh"
              circle
              :icon="Refresh"
              @click="handleRefresh"
              :loading="refreshing"
            />

            <!-- 自定义操作区 -->
            <slot name="header-actions"></slot>
          </div>
        </div>

        <!-- 页面描述 -->
        <p v-if="description" class="mt-2 text-sm text-gray-500 dark:text-gray-400">
          {{ description }}
        </p>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <slot></slot>
    </main>

    <!-- 页脚 -->
    <footer v-if="showFooter" class="border-t dark:border-gray-700 mt-8">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <slot name="footer">
          <p class="text-center text-sm text-gray-500 dark:text-gray-400">
            {{ footerText }}
          </p>
        </slot>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Refresh } from '@element-plus/icons-vue'

defineOptions({
  name: 'IxPageTemplate',
})

interface BreadcrumbItem {
  name: string
  to?: string
}

interface Props {
  title: string
  description?: string
  showHeader?: boolean
  showBreadcrumb?: boolean
  breadcrumbItems?: BreadcrumbItem[]
  showRefresh?: boolean
  showFooter?: boolean
  footerText?: string
}

withDefaults(defineProps<Props>(), {
  description: '',
  showHeader: true,
  showBreadcrumb: false,
  breadcrumbItems: () => [],
  showRefresh: false,
  showFooter: false,
  footerText: '© 2024 IxPay Pro. All rights reserved.',
})

const emit = defineEmits<{
  refresh: []
}>()

const refreshing = ref(false)

// 处理刷新
const handleRefresh = async () => {
  refreshing.value = true
  try {
    emit('refresh')
  } finally {
    setTimeout(() => {
      refreshing.value = false
    }, 500)
  }
}
</script>

<style scoped>
.ix-page {
  background-color: var(--bg-primary);
}

header {
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-primary);
}

footer {
  border-top-color: var(--border-primary);
}

/* 暗黑模式适配 */
html.dark .ix-page {
  background-color: var(--bg-primary);
}

html.dark header {
  background-color: var(--bg-secondary);
  border-bottom-color: var(--border-primary);
}

html.dark footer {
  border-top-color: var(--border-primary);
}
</style>
