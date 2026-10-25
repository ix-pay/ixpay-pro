# Application 应用卡片组件使用示例

## 基础用法

```vue
<template>
  <div class="p-6">
    <h2 class="text-2xl font-bold mb-6">应用中心</h2>

    <!-- 基础用法：3 列布局 -->
    <Application :apps="appList" @click="handleAppClick" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Application, { type AppItem } from './index.vue'

const appList = ref<AppItem[]>([
  {
    id: '1',
    name: '用户管理',
    description: '管理系统用户、角色和权限配置',
    icon: '/images/icons/user.svg',
  },
  {
    id: '2',
    name: '支付中心',
    description: '处理支付订单、退款和交易记录',
    icon: '/images/icons/payment.svg',
  },
  {
    id: '3',
    name: '数据统计',
    description: '可视化数据报表和业务分析',
    icon: '/images/icons/analytics.svg',
  },
  {
    id: '4',
    name: '系统设置',
    description: '系统参数配置和基础数据管理',
    icon: '/images/icons/settings.svg',
  },
])

const handleAppClick = (app: AppItem): void => {
  console.log('点击了应用:', app.name)
  // 路由跳转等逻辑
}
</script>
```

## 响应式列数

```vue
<template>
  <div class="space-y-8">
    <!-- 1 列布局 -->
    <Application :apps="appList" :cols="1" @click="handleAppClick" />

    <!-- 2 列布局 -->
    <Application :apps="appList" :cols="2" @click="handleAppClick" />

    <!-- 3 列布局（默认） -->
    <Application :apps="appList" :cols="3" @click="handleAppClick" />

    <!-- 4 列布局 -->
    <Application :apps="appList" :cols="4" @click="handleAppClick" />
  </div>
</template>
```

## 暗黑模式

```vue
<template>
  <div class="p-6">
    <!-- 启用暗黑模式 -->
    <Application :apps="appList" :cols="3" dark @click="handleAppClick" />
  </div>
</template>
```

## 自定义插槽内容

```vue
<template>
  <div class="p-6">
    <Application :apps="appList" :cols="3" @click="handleAppClick">
      <!-- 自定义图标 -->
      <template #icon="{ app }">
        <div class="app-card__icon">
          <img :src="app.icon" :alt="app.name" />
        </div>
      </template>

      <!-- 自定义标题 -->
      <template #title="{ app }">
        <h3 class="text-lg font-semibold">
          {{ app.name }}
          <span class="text-xs text-blue-500 ml-1">NEW</span>
        </h3>
      </template>

      <!-- 自定义描述 -->
      <template #description="{ app }">
        <p class="text-sm text-gray-600">
          {{ app.description }}
        </p>
      </template>

      <!-- 额外内容 -->
      <template #extra="{ app }">
        <button class="mt-4 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600">
          进入应用
        </button>
      </template>
    </Application>
  </div>
</template>
```

## 完整示例

```vue
<template>
  <div class="min-h-screen bg-gray-50 p-8">
    <div class="max-w-7xl mx-auto">
      <!-- 页面标题 -->
      <div class="text-center mb-12">
        <h1 class="text-4xl font-bold text-gray-900 mb-4">应用中心</h1>
        <p class="text-lg text-gray-600">选择您需要使用的应用</p>
      </div>

      <!-- 应用网格 -->
      <Application
        :apps="appList"
        :cols="4"
        dark
        @click="handleAppClick"
        @mouseenter="handleMouseEnter"
        @mouseleave="handleMouseLeave"
      >
        <!-- 自定义额外内容：快捷操作 -->
        <template #extra="{ app }">
          <div class="flex justify-center space-x-2 mt-4">
            <button
              class="px-3 py-1.5 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition"
              @click.stop="handleQuickAction(app, 'open')"
            >
              打开
            </button>
            <button
              class="px-3 py-1.5 text-sm bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition"
              @click.stop="handleQuickAction(app, 'settings')"
            >
              设置
            </button>
          </div>
        </template>
      </Application>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Application, { type AppItem } from './index.vue'

const router = useRouter()

const appList = ref<AppItem[]>([
  {
    id: 'user-management',
    name: '用户管理',
    description: '管理系统用户、角色权限和组织架构',
    icon: '/images/icons/user.svg',
  },
  {
    id: 'payment-center',
    name: '支付中心',
    description: '处理支付、退款和交易流水管理',
    icon: '/images/icons/payment.svg',
  },
  {
    id: 'order-management',
    name: '订单管理',
    description: '订单查询、处理和售后服务',
    icon: '/images/icons/order.svg',
  },
  {
    id: 'analytics',
    name: '数据统计',
    description: '业务数据分析和可视化报表',
    icon: '/images/icons/analytics.svg',
  },
  {
    id: 'marketing',
    name: '营销中心',
    description: '营销活动、优惠券和促销管理',
    icon: '/images/icons/marketing.svg',
  },
  {
    id: 'customer-service',
    name: '客服中心',
    description: '客户咨询、投诉和建议处理',
    icon: '/images/icons/service.svg',
  },
  {
    id: 'inventory',
    name: '库存管理',
    description: '商品库存、出入库和盘点管理',
    icon: '/images/icons/inventory.svg',
  },
  {
    id: 'system-settings',
    name: '系统设置',
    description: '系统参数、字典和日志管理',
    icon: '/images/icons/settings.svg',
  },
])

const handleAppClick = (app: AppItem): void => {
  console.log('点击应用:', app.name)
  // 路由跳转
  router.push(`/app/${app.id}`)
}

const handleMouseEnter = (app: AppItem): void => {
  console.log('鼠标移入:', app.name)
}

const handleMouseLeave = (app: AppItem): void => {
  console.log('鼠标移出:', app.name)
}

const handleQuickAction = (app: AppItem, action: string): void => {
  console.log(`应用 [${app.name}] 快捷操作：${action}`)
  if (action === 'open') {
    router.push(`/app/${app.id}`)
  } else if (action === 'settings') {
    router.push(`/app/${app.id}/settings`)
  }
}
</script>
```

## 响应式效果说明

组件内置了完整的响应式布局：

- **大屏（>1024px）**：显示指定列数
- **中屏（768px-1024px）**：4 列自动变为 3 列
- **小屏（640px-768px）**：3-4 列自动变为 2 列
- **手机（<640px）**：所有布局变为 1 列

## 设计令牌说明

组件使用 CSS 变量实现设计令牌，方便主题定制：

```scss
// 间距令牌
--app-spacing-sm: 0.5rem; // 8px
--app-spacing-md: 1rem; // 16px
--app-spacing-lg: 1.5rem; // 24px
--app-spacing-xl: 2rem; // 32px

// 圆角令牌
--app-radius-sm: 0.5rem; // 8px
--app-radius-md: 0.75rem; // 12px
--app-radius-lg: 1rem; // 16px
--app-radius-xl: 1.25rem; // 20px

// 阴影令牌
--app-shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
--app-shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
--app-shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
--app-shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.1);

// 颜色令牌（亮色模式）
--app-bg-primary: #ffffff;
--app-text-primary: #111827;
--app-text-secondary: #6b7280;
--app-accent-color: #3b82f6;

// 颜色令牌（暗黑模式）
--app-bg-primary: #1f2937;
--app-text-primary: #f9fafb;
--app-text-secondary: #9ca3af;
--app-accent-color: #60a5fa;
```

## API 文档

### Props

| 属性 | 类型               | 默认值  | 说明             |
| ---- | ------------------ | ------- | ---------------- |
| apps | `AppItem[]`        | `[]`    | 应用列表数据     |
| cols | `1 \| 2 \| 3 \| 4` | `3`     | 网格列数         |
| dark | `boolean`          | `false` | 是否启用暗黑模式 |

### Events

| 事件名     | 参数                     | 说明               |
| ---------- | ------------------------ | ------------------ |
| click      | `(app: AppItem) => void` | 点击应用卡片时触发 |
| mouseenter | `(app: AppItem) => void` | 鼠标移入时触发     |
| mouseleave | `(app: AppItem) => void` | 鼠标移出时触发     |

### Slots

| 插槽名      | 作用域参数         | 说明           |
| ----------- | ------------------ | -------------- |
| icon        | `{ app: AppItem }` | 自定义应用图标 |
| title       | `{ app: AppItem }` | 自定义应用标题 |
| description | `{ app: AppItem }` | 自定义应用描述 |
| extra       | `{ app: AppItem }` | 自定义额外内容 |

### TypeScript 类型

```typescript
interface AppItem {
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
```
