---
name: Git 提交与推送工具
description: 当用户明确要求执行 Git 提交或推送命令时触发（如"提交代码"、"推送代码"、"git commit"等），自动执行 git add、commit、push 操作，确保遵循 Conventional Commits 规范并推送到 main 分支
---

## 核心原则

1. **明确触发**：仅在用户明确指令时触发，不主动建议
2. **单一职责**：只执行 Git 提交推送，不做代码审查
3. **分支安全**：只能推送到 main 分支，禁止推送到其他分支
4. **提交规范**：严格遵循 Conventional Commits 规范
5. **命令兼容**：使用 PowerShell 语法（分号分隔命令）

## 核心功能概览

- **智能触发**：识别"提交代码"、"推送代码"、"git commit"、"git push"等指令
- **规范提交**：自动使用 Conventional Commits 格式（feat/fix/docs/style/refactor/test/chore）
- **安全推送**：强制推送到 main 分支，确保分支安全
- **错误处理**：执行失败时提供清晰的错误信息和解决方案

## 工作流程

```
用户指令 → 识别意图 → 确认提交信息 → 执行 git add → 执行 git commit → 执行 git push → 返回结果
```

### 详细步骤

1. **识别用户指令**
   - 监听关键词：提交、推送、git commit、git push
   - 提取提交信息（如用户提供）

2. **确认提交信息**
   - 用户已提供 → 直接使用
   - 用户未提供 → 询问提交类型和描述

3. **执行提交流程**
   ```powershell
   git add .; git commit -m "类型：描述"; git push origin main
   ```

4. **返回执行结果**
   - 成功：显示提交哈希和推送状态
   - 失败：显示错误原因和解决建议

## Conventional Commits 规范

### 提交类型

- **feat**：新功能
- **fix**：修复 bug
- **docs**：文档更新
- **style**：代码格式（不影响代码运行）
- **refactor**：重构（既不是新功能也不是 bug 修复）
- **test**：添加或修改测试
- **chore**：构建过程或辅助工具变动

### 提交格式

```
<类型>: <描述>
```

### 正确示例

✅ `feat: 新增用户管理模块`
✅ `fix: 修复登录验证逻辑`
✅ `docs: 更新 API 文档`
✅ `refactor: 优化数据库连接池`
✅ `test: 添加支付接口单元测试`
✅ `chore: 更新依赖版本`

### 错误示例

❌ `更新代码`（缺少类型）
❌ `feat 新增功能`（缺少冒号）
❌ `FIX: 修复 bug`（类型应小写）
❌ `feat: 修复某个问题`（类型和描述不匹配）

## 使用示例

### 示例 1：用户完整指令

**用户**：提交代码，feat: 新增用户管理

**执行**：
```powershell
git add .; git commit -m "feat: 新增用户管理"; git push origin main
```

**返回**：
✅ 提交成功（abc1234），已推送到 main 分支

### 示例 2：用户简化指令

**用户**：推送代码

**AI**：请提供提交信息，例如"feat: 新增功能"或描述本次提交的内容

**用户**：fix: 修复支付接口超时问题

**执行**：
```powershell
git add .; git commit -m "fix: 修复支付接口超时问题"; git push origin main
```

### 示例 3：分支保护

**用户**：推送到 dev 分支

**AI**：⚠️ 根据项目规范，代码只能推送到 main 分支。是否执行 `git push origin main`？

## 错误处理

### 常见错误及解决方案

1. **未提交更改**
   - 错误：`nothing to commit, working tree clean`
   - 处理：提示用户没有需要提交的更改

2. **远程同步失败**
   - 错误：`rejected ... fetch first`
   - 处理：建议先执行 `git pull` 同步远程更改

3. **权限错误**
   - 错误：`Permission denied`
   - 处理：提示用户检查 Git 凭证或 SSH 密钥配置

4. **网络错误**
   - 错误：`Could not resolve hostname`
   - 处理：提示用户检查网络连接

## 注意事项

1. **执行前确认**：推送前确保用户了解将推送到 main 分支
2. **本地测试**：建议用户先在本地运行测试确保代码正确
3. **冲突处理**：如遇合并冲突，指导用户手动解决后再提交
4. **大文件检查**：提交前检查是否有大文件（受.gitignore 约束）
