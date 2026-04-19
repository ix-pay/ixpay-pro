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

**核心原则**：优先使用组件库能力，保持代码简洁。

1. **Element Plus 组件属性** - 优先使用组件内置功能和样式属性
2. **Tailwind CSS 原子类** - 布局和简单样式
3. **全局样式文件** - 组件样式自定义在 `global.scss` 中统一处理
4. **禁止组件内写 CSS** - 非必要不在页面或组件中编写 `<style>` 标签

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

#### 响应式规则
- **移动端（< 768px）**：侧边栏弹出式，表格卡片式，1-2 列布局
- **平板（768px - 1024px）**：双列布局，2-3 列卡片
- **桌面（> 1024px）**：4 列布局，完整表格

#### 暗黑模式
- **背景**：深蓝系（`--bg-primary: #0f172a`、`--bg-secondary: #1e293b`）
- **文字**：浅灰系（`--text-primary: #f1f5f9`、`--text-secondary: #cbd5e1`）
- **主色**：提高亮度（`--primary-color: #818cf8`）
- **阴影**：增强透明度（`--shadow-dark-*`）
- **对比度**：确保文字可读性至少 4.5:1

## 性能与可访问性

- 使用 CSS 变量实现主题切换
- 动画使用 CSS 和 transform
- 所有交互元素可键盘访问
- 文字对比度至少 4.5:1

## 相关文档

- [设计令牌定义](../../web/src/assets/styles/design-tokens.scss)
- [全局样式](../../web/src/assets/styles/global.scss)
