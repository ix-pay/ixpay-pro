# 混合架构开发规范

## Element Plus + Tailwind CSS + 设计系统

### 一、架构概述

本项目采用 **Element Plus + Tailwind CSS + 自定义设计系统** 的混合架构，结合了三者的优势：

- **Element Plus**: 提供成熟的 UI 组件（表格、表单、弹窗等）
- **Tailwind CSS**: 提供原子化的样式类，快速构建布局和装饰
- **设计系统**: 提供统一的设计令牌（配色、间距、圆角、阴影等）

### 二、使用原则

#### 2.1 使用 Tailwind CSS 的场景

✅ **布局**

```vue
<div class="flex items-center justify-between"></div>
```

✅ **排版**

```vue
<h1 class="text-2xl font-bold text-gray-900"></h1>
```

✅ **装饰**

```vue
<div class="bg-white rounded-xl shadow-lg hover:shadow-xl transition-shadow"></div>
```

✅ **响应式**

```vue
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4"></div>
```

#### 2.2 使用 Element Plus 的场景

✅ **复杂交互组件**

```vue
<el-table :data="users"></el-table>
```

✅ **需要无障碍支持的组件**

```vue
<el-button></el-button>
```

✅ **需要键盘导航的组件**

```vue
<el-menu></el-menu>
```

#### 2.3 使用自定义 SCSS 的场景

✅ **复杂的动画效果**

```scss
@keyframes customAnimation {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
}
```

✅ **需要深度定制 Element Plus 组件**

```scss
:deep(.el-table) {
  --el-table-header-bg-color: var(--bg-light);
  --el-table-row-hover-bg-color: var(--bg-hover);
}
```

✅ **性能敏感的场景**

```scss
// 减少 CSS 体积，避免生成大量 Tailwind 类
.optimized-component {
  @apply p-4 bg-white rounded-lg;
}
```

### 三、代码示例

#### 3.1 好的实践 ✅

```vue
<template>
  <!-- 使用 Tailwind 进行布局和装饰 -->
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- 页头 -->
    <header class="bg-white dark:bg-gray-800 shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">用户管理</h1>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加用户
          </el-button>
        </div>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <!-- 统计卡片 -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
        <IxStatCard title="总用户数" value="12,345" change="+12.5%" icon="User" color="blue" />
      </div>

      <!-- 表格卡片 -->
      <IxCard title="用户列表" hover-effect>
        <el-table :data="users" class="w-full">
          <el-table-column prop="userName" label="用户名" />
          <el-table-column prop="email" label="邮箱" />
        </el-table>
      </IxCard>
    </main>
  </div>
</template>

<style scoped>
/* 只需保留少量定制样式 */
.modern-table :deep(.el-table) {
  --el-table-header-bg-color: var(--bg-light);
  --el-table-row-hover-bg-color: var(--bg-hover);
  border-radius: var(--radius-lg);
  overflow: hidden;
}
</style>
```

**优点**:

- ✅ 布局清晰，使用 Tailwind 类名
- ✅ Element Plus 组件保持默认功能
- ✅ 使用设计系统的 CSS 变量
- ✅ 样式代码简洁，易于维护

#### 3.2 不好的实践 ❌

```vue
<template>
  <!-- 避免内联样式 -->
  <div :style="{ padding: '24px', backgroundColor: '#fff' }">
    <!-- 避免硬编码颜色值 -->
    <h1 style="color: #333; font-size: 20px;">用户管理</h1>

    <!-- 避免使用复杂的嵌套 class -->
    <div class="user-management-container-wrapper-inner">
      <div class="user-table-box-content">
        <table class="user-data-table-style">
          <!-- ... -->
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 避免写大量重复的 SCSS */
.user-management-container-wrapper-inner {
  background-color: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 20px;
  margin: 20px;
}

.user-table-box-content {
  background-color: #ffffff;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 16px;
}
</style>
```

**问题**:

- ❌ 使用内联样式，无法复用
- ❌ 硬编码颜色值，不支持暗黑模式
- ❌ class 命名冗长，难以维护
- ❌ 样式代码重复，没有使用设计令牌

### 四、设计令牌使用指南

#### 4.1 配色系统

```vue
<template>
  <!-- 使用 Tailwind 颜色 -->
  <div class="bg-blue-500 text-white">
  <div class="text-gray-900 dark:text-white">

  <!-- 使用 CSS 变量 -->
  <div :style="{ color: 'var(--primary-color)' }">
</template>

<style scoped>
.my-class {
  color: var(--primary-color);
  background-color: var(--bg-light);
}
</style>
```

#### 4.2 间距系统

```vue
<template>
  <!-- 使用 Tailwind 间距 -->
  <div class="p-4 m-2 space-x-4">
  <div class="gap-6">

  <!-- 使用 CSS 变量 -->
  <div :style="{ padding: 'var(--space-md)' }">
</template>
```

#### 4.3 圆角系统

```vue
<template>
  <!-- 使用 Tailwind 圆角 -->
  <div class="rounded-lg rounded-xl rounded-full">

  <!-- 使用 CSS 变量 -->
  <div :style="{ borderRadius: 'var(--radius-lg)' }">
</template>
```

#### 4.4 阴影系统

```vue
<template>
  <!-- 使用 Tailwind 阴影 -->
  <div class="shadow-md shadow-lg hover:shadow-xl">

  <!-- 使用 CSS 变量 -->
  <div :style="{ boxShadow: 'var(--shadow-lg)' }">
</template>
```

### 五、响应式设计规范

#### 5.1 断点定义

```javascript
// tailwind.config.js
theme: {
  screens: {
    'sm': '640px',  // 小型设备
    'md': '768px',  // 中型设备
    'lg': '1024px', // 大型设备
    'xl': '1280px', // 超大型设备
    '2xl': '1536px',// 超大屏
  }
}
```

#### 5.2 响应式类名使用

```vue
<template>
  <!-- 移动优先 -->
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">

  <!-- 隐藏/显示 -->
  <div class="hidden md:block lg:hidden">

  <!-- 响应式间距 -->
  <div class="p-4 md:p-6 lg:p-8">

  <!-- 响应式字体 -->
  <h1 class="text-lg md:text-xl lg:text-2xl">
</template>
```

### 六、暗黑模式适配

#### 6.1 使用 dark: 前缀

```vue
<template>
  <div class="bg-white dark:bg-gray-800">
  <div class="text-gray-900 dark:text-white">
  <div class="border-gray-200 dark:border-gray-700">
</template>
```

#### 6.2 使用 CSS 变量

```scss
.my-component {
  background-color: var(--bg-light);
  color: var(--text-primary);
  border-color: var(--border-color);
}
```

### 七、业务组件使用

#### 7.1 IxCard - 业务卡片

```vue
<template>
  <IxCard title="卡片标题" hover-effect padding="md" shadow="md">
    <template #header-actions>
      <el-button size="small" type="primary">操作</el-button>
    </template>

    卡片内容

    <template #footer>
      <div class="text-sm text-gray-500">页脚信息</div>
    </template>
  </IxCard>
</template>
```

#### 7.2 IxStatCard - 统计卡片

```vue
<template>
  <IxStatCard
    title="总用户数"
    value="12,345"
    change="+12.5%"
    change-type="positive"
    change-label="较上月"
    icon="User"
    color="blue"
    :progress="75"
    progress-label="完成度"
  />
</template>
```

#### 7.3 IxPageTemplate - 页面模板

```vue
<template>
  <IxPageTemplate
    title="页面标题"
    description="页面描述"
    show-header
    show-breadcrumb
    :breadcrumb-items="[{ name: '首页', to: '/' }, { name: '用户管理' }]"
    show-refresh
    @refresh="handleRefresh"
  >
    <template #header-actions>
      <el-button type="primary">添加</el-button>
    </template>

    页面内容

    <template #footer>
      <p>自定义页脚</p>
    </template>
  </IxPageTemplate>
</template>
```

### 八、AI 辅助开发技巧

#### 8.1 高效提示词模板

```
请帮我用 Tailwind CSS 设计一个 [组件类型]，要求：
- 风格：[现代/简约/商务/科技感]
- 配色：[主色调]
- 功能：[具体功能描述]
- 响应式：[移动端适配需求]
- 暗黑模式：[是否需要]

示例参考：[提供截图或描述]
```

#### 8.2 实际使用案例

**提示词**:

```
请帮我用 Tailwind CSS 设计一个现代化的登录页面，要求：
- 风格：科技感、现代
- 配色：蓝紫渐变
- 功能：用户名密码输入、验证码、记住密码
- 响应式：完美适配手机和桌面
- 暗黑模式：支持
- 特效：背景动画、输入框聚焦动画
```

### 九、性能优化

#### 9.1 启用 JIT 模式

Tailwind CSS 默认启用 JIT（Just-In-Time）模式，按需编译样式，大幅减少包体积。

#### 9.2 配置 purge

```javascript
// tailwind.config.js
module.exports = {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  // ...
}
```

#### 9.3 使用 CSS 变量

使用 CSS 变量减少重复代码，提高可维护性：

```scss
.my-component {
  @apply p-4 rounded-lg;
  background-color: var(--bg-light);
  color: var(--text-primary);
}
```

### 十、常见问题

#### Q1: Tailwind CSS 和 Element Plus 样式冲突怎么办？

**A**: 在 `tailwind.config.js` 中禁用 preflight：

```javascript
corePlugins: {
  preflight: false,
}
```

#### Q2: 如何选择使用 Tailwind 还是 SCSS？

**A**:

- 布局、间距、排版 → Tailwind CSS
- 复杂动画、深度定制 → SCSS
- 组件样式 → 使用 CSS 变量 + Tailwind

#### Q3: 如何统一团队代码风格？

**A**:

1. 使用 ESLint + Prettier
2. 建立代码审查机制
3. 定期组织技术分享
4. 使用 AI 辅助生成代码

### 十一、持续改进

本规范将根据项目发展和团队反馈持续更新，欢迎提出宝贵意见。

**联系方式**:

- 技术负责人：[填写]
- 更新时间：2024-01-15
- 版本：v1.0
