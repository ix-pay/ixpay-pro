---
alwaysApply: true
---
# AI 说明书

## 核心行为准则

AI 必须按以下流程调用技能：

### 1. 文件创建流程
创建文件 → [项目目录结构指引](../skills/项目目录结构指引/SKILL.md) → 确定位置 → 创建

### 2. 代码开发流程
全栈开发 → [全栈开发](../skills/全栈开发/SKILL.md) → 7 步流程 → [代码质量检查器](../skills/代码质量检查器/SKILL.md)
前端设计 → [前端设计](../skills/前端设计/SKILL.md)

### 3. 命令执行规范（PowerShell 语法，使用分号分隔）
后端 Go → [后端开发工具集](../skills/后端开发工具集/SKILL.md)（构建/测试/Swagger/Docker）
前端 Vue → [前端开发工具集](../skills/前端开发工具集/SKILL.md)（构建/格式化/类型检查）

### 4. 代码提交流程
开发完成 → [代码质量检查器](../skills/代码质量检查器/SKILL.md) → [Git 提交与推送工具](../skills/Git 提交与推送工具/SKILL.md)

### 5. 规则/技能修改
修改规则/技能 → [规则与技能指南](../skills/规则与技能指南/SKILL.md) → 检查字符数/frontmatter

## 必须要求

1. **文件位置优先**：先定位置，再创建
2. **技能驱动**：优先调用技能
3. **规范检查**：代码完成必调质量检查器
4. **规则遵守**：rules ≤1000 字符、中文命名、frontmatter
5. **质量第一**：格式化 + 类型检查 + 构建验证
6. **单文件操作**：一次只操作一个文件
7. **禁止 PowerShell 的 Get-Content 和 -replace 命令修改文件**：只用 Edit 工具
8. **前后端一致性**：ID 使用 string，JSON 标签 camelCase