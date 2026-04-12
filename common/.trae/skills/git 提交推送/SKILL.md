---
name: Git 提交与推送工具
description: 在用户明确要求执行 git 提交或推送命令时触发，如"提交代码"、"推送代码"、"git commit"、"git push"等精确指令
---

## 核心规则

1. **触发条件**：只在用户明确说 git 相关命令时触发
2. **职责**：只执行 git add、git commit、git push 命令，不做代码审查
3. **分支推送**：只能推送到 main 分支，禁止推送到 dev 或其他分支
4. **提交规范**：遵循 Conventional Commits 规范（feat、fix、docs、style、refactor、test、chore）
5. **命令语法**：使用 PowerShell 兼容语法（`;` 分隔命令）

## 工作流程

1. 用户明确指令 → 2. 确认提交信息 → 3. 执行 git add/commit/push → 4. 返回结果

## 示例

**用户**：提交代码，feat: 新增用户管理

**执行**：
```bash
git add .; git commit -m "feat: 新增用户管理"; git push origin main
```
