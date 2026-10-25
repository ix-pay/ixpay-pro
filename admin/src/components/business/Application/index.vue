<template>
  <div class="app-grid" :class="[`app-grid-cols-${cols}`, { dark: isDark }]">
    <div
      v-for="(app, index) in apps"
      :key="app.id || index"
      class="app-card group"
      @click="handleClick(app)"
      @mouseenter="handleMouseEnter"
      @mouseleave="handleMouseLeave"
    >
      <!-- 渐变背景层 -->
      <div class="app-card__bg"></div>

      <!-- 内容区域 -->
      <div class="app-card__content">
        <!-- 应用图标 -->
        <div class="app-card__icon-wrapper">
          <slot name="icon" :app="app">
            <div class="app-card__icon">
              <img v-if="app.icon" :src="app.icon" :alt="app.name" class="app-card__icon-img" />
              <span v-else class="app-card__icon-placeholder">
                {{ app.name.charAt(0).toUpperCase() }}
              </span>
            </div>
          </slot>
        </div>

        <!-- 应用名称 -->
        <h3 class="app-card__title">
          <slot name="title" :app="app">
            {{ app.name }}
          </slot>
        </h3>

        <!-- 应用描述 -->
        <p v-if="app.description" class="app-card__description">
          <slot name="description" :app="app">
            {{ app.description }}
          </slot>
        </p>

        <!-- 自定义内容插槽 -->
        <div class="app-card__extra">
          <slot name="extra" :app="app"></slot>
        </div>
      </div>

      <!-- 悬浮效果遮罩 -->
      <div class="app-card__overlay"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDark } from '@vueuse/core'

/**
 * 应用卡片数据接口
 */
export interface AppItem {
  /** 应用唯一标识 */
  id?: string | number
  /** 应用名称 */
  name: string
  /** 应用描述 */
  description?: string
  /** 应用图标 URL */
  icon?: string
  /** 自定义数据 */
  [key: string]: unknown
}

/**
 * 组件 Props 定义
 */
interface Props {
  /** 应用列表数据 */
  apps: AppItem[]
  /** 网格列数 (1-4)，支持响应式 */
  cols?: 1 | 2 | 3 | 4
  /** 是否启用暗黑模式 */
  dark?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  cols: 3,
  dark: false,
})

defineOptions({
  name: 'AppGrid',
})

/**
 * 组件 Emits 定义
 */
interface Emits {
  /** 点击应用卡片 */
  (e: 'click', app: AppItem): void
  /** 鼠标移入 */
  (e: 'mouseenter', app: AppItem): void
  /** 鼠标移出 */
  (e: 'mouseleave', app: AppItem): void
}

const emit = defineEmits<Emits>()

/**
 * 暗黑模式状态 - 在 setup 顶层调用 useDark
 */
const useDarkMode = useDark()
const isDark = computed(() => props.dark || useDarkMode.value)

/**
 * 处理点击事件
 */
const handleClick = (app: AppItem): void => {
  emit('click', app)
}

/**
 * 处理鼠标移入事件
 */
const handleMouseEnter = (event: MouseEvent): void => {
  const target = event.currentTarget as HTMLElement
  const app = props.apps[Array.from(target.parentNode?.children || []).indexOf(target)]
  if (app) {
    emit('mouseenter', app)
  }
}

/**
 * 处理鼠标移出事件
 */
const handleMouseLeave = (event: MouseEvent): void => {
  const target = event.currentTarget as HTMLElement
  const app = props.apps[Array.from(target.parentNode?.children || []).indexOf(target)]
  if (app) {
    emit('mouseleave', app)
  }
}
</script>

<style lang="scss" scoped>
/**
 * 设计令牌变量
 */
.app-grid {
  // 间距令牌
  --app-spacing-sm: 0.5rem;
  --app-spacing-md: 1rem;
  --app-spacing-lg: 1.5rem;
  --app-spacing-xl: 2rem;

  // 圆角令牌
  --app-radius-sm: 0.5rem;
  --app-radius-md: 0.75rem;
  --app-radius-lg: 1rem;
  --app-radius-xl: 1.25rem;

  // 阴影令牌
  --app-shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --app-shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  --app-shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  --app-shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);

  // 颜色令牌 - 亮色模式
  --app-bg-primary: #ffffff;
  --app-bg-secondary: #f9fafb;
  --app-text-primary: #111827;
  --app-text-secondary: #6b7280;
  --app-border-color: #e5e7eb;
  --app-accent-color: #3b82f6;
  --app-gradient-from: #eff6ff;
  --app-gradient-to: #dbeafe;

  // 过渡效果
  --app-transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

  display: grid;
  gap: var(--app-spacing-lg);
  width: 100%;

  // 响应式网格布局
  &.app-grid-cols-1 {
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }

  &.app-grid-cols-2 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  &.app-grid-cols-3 {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  &.app-grid-cols-4 {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  // 平板响应式
  @media (max-width: 1024px) {
    &.app-grid-cols-4 {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
  }

  // 小屏响应式
  @media (max-width: 768px) {
    &.app-grid-cols-3,
    &.app-grid-cols-4 {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  // 手机响应式
  @media (max-width: 640px) {
    gap: var(--app-spacing-md);

    &.app-grid-cols-2,
    &.app-grid-cols-3,
    &.app-grid-cols-4 {
      grid-template-columns: repeat(1, minmax(0, 1fr));
    }
  }
}

/**
 * 应用卡片样式
 */
.app-card {
  position: relative;
  overflow: hidden;
  cursor: pointer;
  border-radius: var(--app-radius-xl);
  background: var(--app-bg-primary);
  box-shadow: var(--app-shadow-md);
  transition: var(--app-transition);
  min-height: 180px;

  // 悬浮效果
  &:hover {
    transform: translateY(-8px);
    box-shadow: var(--app-shadow-xl);

    .app-card__bg {
      opacity: 1;
    }

    .app-card__overlay {
      opacity: 1;
    }

    .app-card__icon {
      transform: scale(1.1) rotate(5deg);
    }
  }

  // 背景渐变层
  &__bg {
    position: absolute;
    inset: 0;
    background: linear-gradient(135deg, var(--app-gradient-from) 0%, var(--app-gradient-to) 100%);
    opacity: 0.6;
    transition: var(--app-transition);
    z-index: 0;
  }

  // 内容区域
  &__content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: var(--app-spacing-xl);
    height: 100%;
  }

  // 图标容器
  &__icon-wrapper {
    margin-bottom: var(--app-spacing-md);
  }

  // 图标样式
  &__icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 64px;
    height: 64px;
    border-radius: var(--app-radius-lg);
    background: linear-gradient(135deg, var(--app-accent-color) 0%, #2563eb 100%);
    box-shadow: var(--app-shadow-md);
    transition: var(--app-transition);
    overflow: hidden;

    &-img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    &-placeholder {
      font-size: 1.75rem;
      font-weight: 700;
      color: #ffffff;
      text-shadow: 0 2px 4px rgb(0 0 0 / 0.2);
    }
  }

  // 标题样式
  &__title {
    margin: 0 0 var(--app-spacing-sm) 0;
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--app-text-primary);
    text-align: center;
    line-height: 1.5;
    transition: var(--app-transition);
  }

  // 描述样式
  &__description {
    margin: 0 0 var(--app-spacing-md) 0;
    font-size: 0.875rem;
    color: var(--app-text-secondary);
    text-align: center;
    line-height: 1.6;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    transition: var(--app-transition);
  }

  // 额外内容区域
  &__extra {
    margin-top: auto;
    width: 100%;
    transition: var(--app-transition);
  }

  // 悬浮遮罩
  &__overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(to bottom, transparent 0%, rgb(59 130 246 / 0.1) 100%);
    opacity: 0;
    transition: var(--app-transition);
    pointer-events: none;
  }
}

/**
 * 暗黑模式适配
 */
.app-grid.dark {
  // 暗黑模式颜色令牌
  --app-bg-primary: #1f2937;
  --app-bg-secondary: #111827;
  --app-text-primary: #f9fafb;
  --app-text-secondary: #9ca3af;
  --app-border-color: #374151;
  --app-accent-color: #60a5fa;
  --app-gradient-from: #1e3a5f;
  --app-gradient-to: #1e40af;

  .app-card {
    background: var(--app-bg-primary);
    border: 1px solid var(--app-border-color);

    &:hover {
      border-color: var(--app-accent-color);
    }

    &__title {
      color: var(--app-text-primary);
    }

    &__description {
      color: var(--app-text-secondary);
    }

    &__icon {
      background: linear-gradient(135deg, var(--app-accent-color) 0%, #3b82f6 100%);
    }
  }
}

// Tailwind CSS 工具类兼容
// 如果需要使用 Tailwind 的工具类，可以在这里添加
</style>
