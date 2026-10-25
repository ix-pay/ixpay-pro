<template>
  <div
    class="flex-1 bg-[var(--bg-primary)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] box-border flex flex-col h-[calc(100vh-64px)] overflow-hidden relative"
    :class="{
      'ml-[64px]': isSidebarCollapsed,
      'ml-[240px]': !isSidebarCollapsed,
      'md:ml-[64px]': true,
    }"
  >
    <div
      class="flex-1 p-[var(--space-lg)] overflow-y-auto overflow-x-hidden scrollbar-thin scrollbar-thumb-[var(--border-dark)] scrollbar-track-[var(--bg-dark)] transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] relative"
    >
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
// Props
defineProps({
  isSidebarCollapsed: {
    type: Boolean,
    default: false,
  },
})
</script>

<style scoped>
/* 使用 Tailwind CSS 重构后，仅保留滚动条样式和响应式调整 */

/* 响应式设计 - 平板设备 */
@media (max-width: 1024px) {
  .ml-\[240px\] {
    margin-left: 64px;
  }

  .ml-\[64px\] {
    margin-left: 240px;
  }
}

/* 响应式设计 - 移动设备 */
@media (max-width: 768px) {
  .ml-\[240px\],
  .ml-\[64px\] {
    margin-left: 0;
  }
}

/* 暗黑模式滚动条调整 */
html.dark {
  .scrollbar-thumb-\[var\(--border-dark\)\] {
    scrollbar-color: var(--border-primary) var(--bg-tertiary);
  }
}
</style>
