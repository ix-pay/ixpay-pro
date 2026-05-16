---
name: frontend-design
description: 当用户需要创建 Web 组件、页面、界面或应用时触发（如网站、落地页、仪表板、React 组件、HTML/CSS 布局或美化 Web UI），创建具有独特设计美感的准生产级前端界面，生成富有创意、精致且避免通用 AI 美学的代码和 UI 设计。支持后台管理系统、H5 页面、微信小程序等多平台设计。
---

## 设计思维

编码前确定美学方向：
- **目的**：界面解决什么问题？谁使用？
- **基调**：选择风格（极简/极繁/复古/科技感等）
- **差异化**：令人难忘的视觉特点

## 样式使用优先级

**核心原则**：优先使用组件库能力，保持代码简洁。详细规范参考设计系统。

1. **Element Plus 组件属性** - 优先使用组件内置功能和样式属性
2. **Tailwind CSS 原子类** - 布局和简单样式（普通 HTML 标签）
3. **全局样式文件** - 组件样式自定义在 `global.scss` 中统一处理
4. **禁止组件内写 CSS** - 非必要不在页面或组件中编写 `<style>` 标签

---

## Element Plus + Tailwind CSS 协同规范

### 使用场景对比

| 场景 | 使用技术 | 示例 |
|------|---------|------|
| 布局（flex/grid） | Tailwind CSS | `flex items-center justify-between` |
| 间距（padding/margin） | Tailwind CSS | `p-4 m-2 gap-4` |
| 表格/表单/弹窗 | Element Plus | `<el-table>`, `<el-form>`, `<el-dialog>` |
| 按钮/输入框 | Element Plus | `<el-button>`, `<el-input>` |
| 复杂动画 | SCSS + CSS 变量 | `@keyframes` + `var(--duration-*)` |
| 主题定制 | CSS 变量 | `var(--primary-color)` |

### 禁止事项

- ❌ 禁止在 Element Plus 组件标签上使用 Tailwind CSS 类
- ❌ 禁止使用内联样式
- ❌ 禁止手写媒体查询（使用 Tailwind 响应式前缀）
- ❌ 禁止自定义无意义类名

### 示例代码

```vue
<!-- ✅ 正确：普通 HTML 使用 Tailwind，Element Plus 使用原生属性 -->
<template>
  <div class="flex items-center gap-4 p-4">
    <el-button type="primary" size="large" @click="handleClick">
      点击
    </el-button>
    <el-table :data="tableData" style="width: 100%">
      <el-table-column prop="name" label="姓名" />
    </el-table>
  </div>
</template>

<!-- ❌ 错误：在 Element Plus 组件上使用 Tailwind CSS 类 -->
<template>
  <el-button type="primary" class="mt-4 p-2">  <!-- ❌ -->
    点击
  </el-button>
  <el-table :data="tableData" class="w-full">  <!-- ❌ -->
    <el-table-column prop="name" label="姓名" class="p-4" />  <!-- ❌ -->
  </el-table>
</template>
```

---

## 统一页面布局规范

### 标准列表页面布局

所有列表页面采用以下标准结构：

```vue
<template>
  <div class="flex flex-col h-full">
    <!-- 顶部操作栏 - 上下两行 -->
    <div class="flex flex-col gap-3 p-4 border-b">
      <!-- 第一行：搜索条件 -->
      <div class="flex flex-wrap items-center gap-3">
        <el-input v-model="searchForm.keyword" placeholder="搜索" style="width: 192px;" />
        <el-button type="primary">搜索</el-button>
        <el-button>重置</el-button>
      </div>
      
      <!-- 第二行：功能按钮 -->
      <div class="flex flex-wrap items-center gap-2">
        <el-button type="primary">添加</el-button>
        <el-button>导出</el-button>
      </div>
    </div>

    <!-- 统计信息区域（可选） -->
    <div class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b">
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500"><User /></el-icon>
          总数：<span class="font-medium">100</span>
        </span>
      </div>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table :data="list" :height="'100%'" stripe>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button type="primary" size="small">编辑</el-button>
              <el-button type="danger" size="small">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页区域 -->
    <div class="flex items-center justify-between px-4 py-3 border-t">
      <span class="text-sm text-gray-600">共 100 条</span>
      <el-pagination layout="prev, pager, next" />
    </div>
  </div>
</template>
```

### 关键规范

1. **搜索框宽度**：使用内联样式 `style="width: 192px;"`（Element Plus 会覆盖 Tailwind 类）
2. **上下两行布局**：外层 `flex flex-col gap-3`，第一行搜索，第二行功能按钮
3. **操作列按钮**：统一使用 `size="small"`，容器 `flex flex-wrap gap-2` 支持换行
4. **表格高度**：使用 `:height="'100%'"` 占满容器
5. **自动换行**：每行都使用 `flex-wrap`，元素过多时自动换行

---

## 设计令牌使用示例

### 配色系统

```vue
<template>
  <!-- Tailwind 颜色 -->
  <div class="bg-blue-500 text-white" />
  <div class="text-gray-900 dark:text-white" />
  
  <!-- CSS 变量 -->
  <div :style="{ color: 'var(--primary-color)' }" />
  <div class="gradient-text-primary" />
</template>
```

### 间距系统

```vue
<template>
  <!-- Tailwind 间距 -->
  <div class="p-4 m-2 space-x-4 gap-6" />
  
  <!-- CSS 变量 -->
  <div :style="{ padding: 'var(--space-md)' }" />
</template>
```

### 圆角系统

```vue
<template>
  <!-- Tailwind 圆角 -->
  <div class="rounded-lg rounded-xl rounded-full" />
  
  <!-- CSS 变量 -->
  <div :style="{ borderRadius: 'var(--radius-lg)' }" />
</template>
```

### 阴影和光晕

```vue
<template>
  <!-- Tailwind 阴影 -->
  <div class="shadow-lg shadow-xl" />
  
  <!-- CSS 变量光晕 -->
  <div class="glow-primary" />
  <div :style="{ boxShadow: 'var(--shadow-primary-glow)' }" />
</template>
```

### 动效系统

```vue
<template>
  <!-- 过渡动画 -->
  <div class="transition-base hover:scale-105" />
  
  <!-- 动画关键帧 -->
  <div class="animate-fade-in animate-float" />
</template>
```

---

## 响应式规则详解

### 断点定义

使用 Tailwind CSS 标准断点：

- **sm**: 640px（小型设备）
- **md**: 768px（平板）
- **lg**: 1024px（桌面）
- **xl**: 1280px（大桌面）
- **2xl**: 1536px（超大桌面）

### 响应式布局策略

#### 移动端（< 768px）

- 侧边栏：弹出式抽屉
- 表格：转换为卡片布局
- 列数：1-2 列
- 按钮：全宽排列
- 搜索框：堆叠布局

```vue
<template>
  <div class="flex flex-col sm:flex-row gap-2">
    <el-input class="w-full sm:w-auto" />
    <el-button class="w-full sm:w-auto">搜索</el-button>
  </div>
</template>
```

#### 平板（768px - 1024px）

- 侧边栏：可收起
- 表格：完整表格
- 列数：2-3 列卡片
- 按钮：一行排列

#### 桌面（> 1024px）

- 侧边栏：展开 260px
- 表格：完整表格
- 列数：4 列布局
- 按钮：按功能分组

### 响应式工具类

```vue
<template>
  <!-- 显示/隐藏 -->
  <div class="hidden sm:block">桌面显示</div>
  <div class="block sm:hidden">移动显示</div>
  
  <!-- 响应式宽度 -->
  <div class="w-full md:w-1/2 lg:w-1/4">响应式宽度</div>
  
  <!-- 响应式间距 -->
  <div class="p-2 md:p-4 lg:p-6">响应式内边距</div>
  
  <!-- 响应式字体 -->
  <div class="text-sm md:text-base lg:text-lg">响应式字体</div>
</template>
```

---

## 暗黑模式适配指南

### 核心原则

1. **自动适配**：使用 `dark:` 前缀自动切换
2. **颜色变量**：使用 CSS 变量而非硬编码颜色
3. **对比度**：确保文字可读性至少 4.5:1
4. **一致性**：保持设计语言的连贯性

### 暗黑模式颜色方案

#### 背景色

- **主背景**：`#0f172a`（深蓝灰）
- **次级背景**：`#1e293b`（深蓝）
- **第三级背景**：`#334155`（中蓝灰）

#### 文字颜色

- **主文字**：`#f1f5f9`（浅灰白）
- **次级文字**：`#cbd5e1`（浅灰）
- **第三级文字**：`#94a3b8`（中灰）

#### 主色调

- **暗黑模式主色**：`#818cf8`（提高亮度的蓝紫）
- **光晕效果**：增强透明度

### 暗黑模式示例

```vue
<template>
  <!-- 背景 -->
  <div class="bg-white dark:bg-gray-900" />
  
  <!-- 文字 -->
  <div class="text-gray-900 dark:text-white" />
  
  <!-- 边框 -->
  <div class="border-gray-200 dark:border-gray-700" />
  
  <!-- 复杂背景 -->
  <div class="bg-gray-50 dark:bg-gray-800" />
</template>
```

### 暗黑模式最佳实践

1. **避免纯黑**：使用深蓝灰色系（`#0f172a` 而非 `#000000`）
2. **降低饱和度**：暗黑模式下颜色饱和度降低 20-30%
3. **增强对比**：阴影和光晕效果增强透明度
4. **测试验证**：在暗黑模式下验证所有组件的可读性

---

## 设计系统规范

使用 CSS 变量实现统一的设计令牌系统，支持主题切换和暗黑模式。

**变量定义**：`src/assets/styles/design-tokens.scss`  
**全局样式**：`src/assets/styles/global.scss`

### 核心变量

- **主色**：`--primary-color`、`--primary-gradient`、`--primary-glow`、`--primary-shadow`
- **功能色**：`--success-color`、`--warning-color`、`--danger-color`、`--info-color`（均支持渐变、光晕效果）
- **中性色**：`--bg-primary`、`--bg-secondary`、`--text-primary`、`--text-secondary`、`--border-primary`
- **布局色**：`--sidebar-bg`、`--header-bg`、`--sidebar-text`、`--sidebar-hover-bg`
- **间距**：`--space-1` ~ `--space-24`（4px 基准）
- **圆角**：`--radius-sm` ~ `--radius-full`
- **阴影**：`--shadow-sm` ~ `--shadow-2xl`、`--shadow-primary-glow` 等
- **动效**：`--duration-fast/normal/slow`、`--ease-in-out/ease-out/ease-bounce`
- **字体**：`--font-sans`、`--font-mono`、`--text-xs` ~ `--text-4xl`

### 后台管理系统规范

**设计基调**：现代极简主义 + 科技感

#### 布局组件
- **侧边栏**：渐变背景，展开 260px/收起 72px
- **头部导航**：玻璃态效果，高度 64px
- **主内容区**：浅灰色渐变背景，内边距 24px

---

## 性能与可访问性

- 使用 CSS 变量实现主题切换
- 动画使用 CSS 和 transform
- 所有交互元素可键盘访问
- 文字对比度至少 4.5:1

---

## 相关文档

- [设计令牌定义](../../web/src/assets/styles/design-tokens.scss)
- [全局样式](../../web/src/assets/styles/global.scss)
