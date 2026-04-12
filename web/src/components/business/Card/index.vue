<template>
  <div
    class="bg-white dark:bg-gray-800 rounded-xl shadow-md hover:shadow-xl transition-all duration-300 p-6"
    :class="[hoverEffect ? 'hover:-translate-y-1' : '']"
  >
    <!-- 卡片头部 -->
    <div
      v-if="title || $slots.header"
      class="flex items-center justify-between mb-4 pb-4 border-b dark:border-gray-700"
    >
      <slot name="header">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ title }}</h3>
      </slot>
    </div>

    <!-- 卡片内容 -->
    <div class="text-gray-600 dark:text-gray-300">
      <slot />
    </div>

    <!-- 卡片底部 -->
    <div v-if="$slots.footer" class="mt-4 pt-4 border-t dark:border-gray-700">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: 'IxCard',
})

interface Props {
  title?: string
  hoverEffect?: boolean
}

withDefaults(defineProps<Props>(), {
  title: '',
  hoverEffect: true,
})
</script>
