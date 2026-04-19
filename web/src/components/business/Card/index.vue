<template>
  <div
    class="ix-card relative bg-[var(--bg-primary)] rounded-[var(--radius-xl)] border border-[var(--border-primary)] p-6 overflow-hidden transition-all duration-[var(--duration-normal)] ease-[var(--ease-in-out)]"
    :class="[
      hoverEffect
        ? 'hover:-translate-y-1 hover:shadow-[var(--shadow-xl)] hover:border-transparent'
        : '',
      gradientBorder ? 'group' : '',
    ]"
  >
    <!-- 渐变边框层 -->
    <div
      v-if="gradientBorder"
      class="ix-card__gradient-border absolute inset-0 rounded-[var(--radius-xl)] p-[1px] bg-[var(--primary-gradient)] opacity-0 transition-opacity duration-[var(--duration-normal)] ease-[var(--ease-in-out)] pointer-events-none"
      style="
        -webkit-mask:
          linear-gradient(#fff 0 0) content-box,
          linear-gradient(#fff 0 0);
        -webkit-mask-composite: xor;
        mask-composite: exclude;
      "
    />

    <!-- 卡片头部 -->
    <div
      v-if="title || $slots.header"
      class="ix-card__header mb-4 pb-4 border-b border-[var(--border-primary)]"
    >
      <slot name="header">
        <h3
          class="ix-card__title text-[var(--text-lg)] font-semibold text-[var(--text-primary)] m-0 leading-normal"
        >
          {{ title }}
        </h3>
      </slot>
    </div>

    <!-- 卡片内容 -->
    <div class="ix-card__content text-[var(--text-secondary)] leading-relaxed">
      <slot />
    </div>

    <!-- 卡片底部 -->
    <div
      v-if="$slots.footer"
      class="ix-card__footer mt-4 pt-4 border-t border-[var(--border-primary)]"
    >
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
  gradientBorder?: boolean
}

withDefaults(defineProps<Props>(), {
  title: '',
  hoverEffect: true,
  gradientBorder: false,
})
</script>

<style lang="scss" scoped>
// 暗黑模式下的渐变边框增强
html.dark .ix-card__gradient-border {
  background: var(--primary-gradient);
}

// 确保渐变边框在 hover 时显示
.ix-card:hover .ix-card__gradient-border {
  opacity: 1;
}
</style>
