<template>
  <nav v-if="visible" class="tabbar">
    <div
      v-for="item in items"
      :key="item.path"
      class="tab-item"
      :class="{ active: isActive(item.path) }"
      :style="isActive(item.path) ? { color: primaryColor } : undefined"
      @click="go(item.path)"
    >
      <div class="tab-icon">{{ item.icon || '•' }}</div>
      <div class="tab-title">{{ item.title }}</div>
    </div>
  </nav>
</template>

<script setup lang="ts">
// 底部 tab 导航（首页/个人中心）：tab 项由租户 UI 配置驱动（ui.tabbar.items），空配置走默认
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTabbarItems, uiState } from '@/store/ui'

const route = useRoute()
const router = useRouter()

const items = computed(() => getTabbarItems())
const primaryColor = computed(
  () => uiState.ui?.brand?.primary_color || '#07c160',
)

// 仅带 tab 的页面（meta.tabbar）且已登录时显示
const visible = computed(
  () => route.meta.tabbar === true && !!localStorage.getItem('wx_token'),
)

function isActive(path: string): boolean {
  return route.path === path
}

function go(path: string) {
  if (route.path === path) return
  router.push(path)
}
</script>

<style scoped>
.tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  background: #fff;
  border-top: 1px solid #eee;
  /* 适配 iPhone 底部安全区 */
  padding-bottom: env(safe-area-inset-bottom);
  z-index: 100;
}
.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 0 4px;
  color: #999;
}
.tab-item.active {
  font-weight: 600;
}
.tab-icon {
  font-size: 22px;
  line-height: 1.2;
}
.tab-title {
  font-size: 11px;
  margin-top: 1px;
}
</style>
