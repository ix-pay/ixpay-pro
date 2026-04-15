---
name: 代码质量检查器
description: 在代码完成后自动执行格式化、类型检查、构建验证和文档更新；或在用户要求检查规范时执行全面的规范符合性检查
---

## 核心功能概览

本技能集成了代码质量检查的全流程工具：

1. **代码完成后置检查** - 代码完成后自动执行全面检查
2. **规范符合性检查** - 定期扫描项目，检查规范符合性

***

## 功能 1：代码完成后置检查

**触发条件**：当用户完成代码编写后（说"代码写完了"、"完成开发"等）

### 自动执行检查

#### 1.1 代码格式化

**后端 Go 代码**：

```bash
go fmt ./...
go imports -w .
```

**前端 Vue/TS 代码**：

```bash
npm run format
# 或
npm run lint -- --fix
```

#### 1.2 类型检查

**前端 TypeScript**：

```bash
npm run type-check
```

**检查项目**：

- TypeScript 类型错误
- 接口定义是否完整
- 类型导入是否正确

#### 1.3 文档更新检查

**检查项目**：

- [ ] API 文档是否已更新（Swagger）
- [ ] 新增的 API 是否有 @Summary 注释
- [ ] 新增的 API 是否有 @Router 注释
- [ ] 新增的 API 是否有 @Tags 注释

**自动执行**：

```bash
# 后端 Swagger 文档生成
swag init -g cmd/ixpay-pro/main.go --output ./docs
```

#### 1.4 项目完整性验证

**后端检查**：

- [ ] 是否有编译错误
- [ ] 是否有未使用的导入
- [ ] 是否有语法错误

**前端检查**：

- [ ] 是否有编译错误
- [ ] 是否有未使用的变量
- [ ] 是否有未导入的组件

#### 1.5 构建验证（重要）⭐

**后端构建验证**：

```bash
# 执行后端项目构建
go build -o build/ixpay-pro.exe ./cmd/ixpay-pro
```

**检查项目**：

- [ ] 构建是否成功（退出代码 0）
- [ ] 是否生成可执行文件（build/ixpay-pro.exe）
- [ ] 是否有编译错误
- [ ] 是否有类型错误

**后端测试验证**（可选但推荐）：

```bash
# 执行后端单元测试
go test ./tests/unit/... -v
```

**检查项目**：

- [ ] 测试是否全部通过
- [ ] 是否有失败的测试用例
- [ ] 测试覆盖率是否达标

**前端构建验证**：

```bash
# 执行前端生产构建
npm run build
```

**检查项目**：

- [ ] 构建是否成功（退出代码 0）
- [ ] 是否有 TypeScript 类型错误
- [ ] 是否有 ESLint 错误
- [ ] 是否生成 dist/ 目录
- [ ] 资源文件是否正确打包

### 检查报告

**生成报告格式**：

```
✅ 代码完成后检查完成

**检查项目**：
1. ✅ 代码格式化 - 通过
2. ✅ 类型检查 - 通过
3. ✅ API 文档更新 - 已生成
4. ✅ 项目编译 - 通过
5. ✅ 后端构建验证 - 通过
6. ✅ 前端构建验证 - 通过
7. ⚠️ 后端测试验证 - 部分测试失败（可选）

**修复的问题**：
- 自动格式化：3 个文件
- 自动导入修复：2 个文件
- 自动文档更新：5 个 API

**构建验证结果**：
【后端】
✅ 构建成功：生成了 build/ixpay-pro.exe 可执行文件
✅ 无编译错误
✅ 无类型错误
⚠️ 测试通过率：95% (19/20 个测试通过)

【前端】
✅ 构建成功：生成了 dist/ 目录
✅ 无 TypeScript 类型错误
✅ 无 ESLint 错误
✅ 资源文件已正确打包

**发现的问题**：
⚠️ 发现 2 个警告（不影响运行）
- src/views/user/index.vue:80 - 未使用的变量
- internal/app/base/api/user_handler.go:120 - 注释缺失

建议：
1. 删除未使用的变量
2. 添加缺失的注释
3. 修复失败的测试用例：TestUserDomainService_CreateUser

代码可以提交了！
```

***

## 功能 2：规范符合性检查

**触发条件**：用户要求"检查规范"、"代码审查"或定期执行

### 后端规范检查

#### 2.1 代码风格检查

**检查项目**：

- [ ] 缩进：4 空格
- [ ] 行宽：≤120 字符
- [ ] 括号：K&R 风格
- [ ] 运算符：两侧空格

**执行命令**：

```bash
go fmt ./...
go vet ./...
```

#### 2.2 命名规范检查

**检查项目**：

- [ ] 包名：小写
- [ ] 文件名：小写 + 下划线（snake_case）
- [ ] 函数/变量：驼峰（camelCase）
- [ ] 常量：大写 + 下划线（UPPER_CASE）
- [ ] 结构体：PascalCase
- [ ] **JSON 标签：camelCase** ⭐
- [ ] **ID 字段：应用层和 DTO 使用 string，数据库使用 int64** ⭐

**示例问题**：

```
⚠️ 命名规范问题

**位置**：internal/app/base/api/user_handler.go:50
**问题**：JSON 标签使用 snake_case
**当前**：CreatedAt string `json:"created_at"`
**建议**：CreatedAt string `json:"createdAt"`
```

#### 2.3 架构规范检查

**检查项目**：

- [ ] API Handler 是否直接操作数据库
- [ ] Application Service 是否协调多个 Domain Service
- [ ] Domain Service 是否注入 Repository 接口
- [ ] Repository 实现是否负责数据转换
- [ ] 是否使用 context.Context
- [ ] 是否有全局变量

**示例问题**：

```
⚠️ 分层架构问题

**位置**：internal/app/base/api/user_handler.go:80
**问题**：API Handler 直接操作数据库
**当前**：db.Where("id = ?", id).First(&user)
**建议**：userAppService.GetUserByID(id)
```

#### 2.4 响应格式检查

**检查项目**：

- [ ] 是否使用统一响应工具
- [ ] 响应格式是否为 `{code, data, msg}`
- [ ] msg 字段是否为中文
- [ ] **ID 字段是否使用 string** ⭐
- [ ] **JSON 标签是否使用 camelCase** ⭐

**示例问题**：

```
⚠️ 响应格式问题

**位置**：internal/app/base/api/user_handler.go:120
**问题**：ID 字段未使用 string 类型
**当前**：ID int64 `json:"id"`
**建议**：ID string `json:"id"`
```

#### 2.5 错误处理规范检查

**检查项目**：

- [ ] 是否使用 fmt.Errorf 包装错误
- [ ] 是否保留错误链（使用 %w）
- [ ] 错误消息是否清晰
- [ ] 是否在 Repository 层进行 ID 类型转换

**示例问题**：

```
⚠️ 错误处理问题

**位置**：internal/domain/base/service/user_domain_service.go:80
**问题**：直接返回原始错误，丢失上下文
**当前**：return nil, err
**建议**：return nil, fmt.Errorf("获取用户失败 (ID=%s): %w", id, err)
```

#### 2.6 跨模块调用检查

**检查项目**：

- [ ] 是否通过 Domain Service 跨模块调用
- [ ] 是否避免调用 Application 层或 API 层
- [ ] 依赖注入是否正确配置

**示例问题**：

```
⚠️ 跨模块调用问题

**位置**：internal/app/wx/application/payment_app_service.go:50
**问题**：直接调用 API 层
**当前**：http.Get("/api/base/user/info")
**建议**：userDomainService.GetUserByID(userID)
```

#### 2.7 日志规范检查

**检查项目**：

- [ ] 是否使用指定日志库
- [ ] **日志消息是否使用中文** ⭐
- [ ] 关键操作是否记录 user_id 和 IP
- [ ] 是否记录敏感信息

**示例问题**：

```
⚠️ 日志规范问题

**位置**：internal/app/base/api/auth_handler.go:85
**问题**：日志消息使用英文
**当前**：logger.Info("User login successful")
**建议**：logger.Info("用户登录成功")
```

### 前端规范检查

#### 2.8 代码风格检查

**检查项目**：

- [ ] 缩进：2 空格
- [ ] 引号：单引号
- [ ] 分号：不使用
- [ ] 空行：逻辑块之间保留

**执行命令**：

```bash
npm run lint
```

#### 2.9 命名规范检查

**检查项目**：

- [ ] 组件名：PascalCase
- [ ] 变量/方法名：camelCase
- [ ] 常量：UPPER_CASE
- [ ] **TypeScript 接口：全字段 camelCase** ⭐

**示例问题**：

```
⚠️ 命名规范问题

**位置**：src/types/user.ts:10
**问题**：TypeScript 接口字段使用 snake_case
**当前**：user_name: string
**建议**：userName: string
```

#### 2.10 组件结构检查

**检查项目**：

- [ ] 顺序：template → script → style
- [ ] 是否使用 `<script setup>`
- [ ] props/emits 是否定义
- [ ] 生命周期钩子是否完整

#### 2.11 API 调用规范检查

**检查项目**：

- [ ] API 路径是否使用 `//` 前缀
- [ ] 响应格式是否为 `{code, data, msg}`
- [ ] 列表字段是否使用 `items`
- [ ] 分页参数是否使用 `page, page_size`
- [ ] **ID 字段是否使用 string** ⭐

**示例问题**：

```
⚠️ API 调用规范问题

**位置**：src/views/user/index.vue:50
**问题**：API 路径缺少 // 前缀
**当前**：url: '/api/user'
**建议**：url: '/user'
```

#### 2.12 权限标识检查

**检查项目**：

- [ ] 权限标识是否使用 `update` 而非 `edit`
- [ ] 权限标识是否使用 `delete` 而非 `remove`

**示例问题**：

```
⚠️ 权限标识问题

**位置**：src/views/user/index.vue:120
**问题**：权限标识使用 edit
**当前**：hasPermission('user:edit')
**建议**：hasPermission('user:update')
```

### 生成检查报告

**报告格式**：

```
✅ 规范符合性检查完成

**检查范围**：
- 后端 Go 代码：15 个文件
- 前端 Vue/TS 代码：28 个文件

**检查结果**：
✅ 后端代码风格 - 通过（95 分）
✅ 前端代码风格 - 通过（92 分）
⚠️ 命名规范 - 发现 3 个问题
✅ 架构规范 - 通过
⚠️ 响应格式 - 发现 2 个问题
⚠️ 错误处理 - 发现 1 个问题
✅ 跨模块调用 - 通过
✅ 日志规范 - 通过

**详细问题**：

【后端】
1. ⚠️ internal/app/base/api/user_handler.go:50
   JSON 标签使用 snake_case
   建议：created_at → createdAt

2. ⚠️ internal/app/base/api/user_handler.go:120
   ID 字段未使用 string 类型
   建议：使用 string 类型

3. ⚠️ internal/domain/base/service/user_domain_service.go:80
   错误处理未使用 fmt.Errorf 包装
   建议：使用 fmt.Errorf 包装错误并保留错误链

【前端】
4. ⚠️ src/types/user.ts:10
   字段命名使用 snake_case
   建议：user_name → userName

**修复建议**：
1. 立即修复：ID 字段类型问题（影响功能）
2. 尽快修复：JSON 标签命名问题（影响前后端一致性）
3. 尽快修复：错误处理问题（影响错误追踪）
4. 逐步优化：其他命名问题

需要我帮你自动修复这些问题吗？
```

***

## 自动修复功能

### 可自动修复的问题

#### 后端

- ✅ 代码格式化（go fmt）
- ✅ 导入整理（go imports）
- ✅ 简单命名问题
- ✅ Swagger 文档生成

#### 前端

- ✅ 代码格式化（eslint --fix）
- ✅ 简单命名问题
- ✅ 导入整理

### 自动修复流程

```
1. 识别可自动修复的问题
   ↓
2. 执行自动修复命令
   ├─ go fmt / go imports
   ├─ eslint --fix
   └─ swag init
   ↓
3. 验证修复结果
   ↓
4. 报告修复情况
```

**示例**：

```
🔧 正在自动修复问题...

**执行命令**：
1. ✅ go fmt ./... - 格式化 15 个文件
2. ✅ go imports -w . - 整理 8 个文件的导入
3. ✅ swag init - 生成 API 文档

**修复结果**：
✅ 已修复：12 个问题
⚠️ 待手动修复：3 个问题（需要人工确认）

**待手动修复的问题**：
1. internal/app/user.go:50 - JSON 标签命名
2. src/types/user.ts:10 - 字段命名
3. src/views/user/index.vue:80 - API 路径

需要我帮你查看这些问题的详细信息吗？
```

***

## 持续集成检查

### CI/CD 检查清单

**后端检查**：

```yaml
- name: Go Format Check
  run: go fmt ./... && git diff --exit-code

- name: Go Vet Check
  run: go vet ./...

- name: Build Check
  run: go build -o build/ixpay-pro.exe ./cmd/ixpay-pro

- name: Unit Test Check
  run: go test ./tests/unit/... -v

- name: Swagger Doc Check
  run: swag init -g cmd/ixpay-pro/main.go --output ./docs
```

**前端检查**：

```yaml
- name: ESLint Check
  run: npm run lint

- name: TypeScript Check
  run: npm run type-check

- name: Build Check
  run: npm run build

- name: Build Output Check
  run: test -d dist/
```

***

## 工作流程

```
1. 识别触发条件
   ├─ 用户完成代码编写
   ├─ 用户要求检查规范
   └─ 定期执行检查
   ↓
2. 执行相应检查
   ├─ 代码格式化
   ├─ 类型检查
   ├─ 规范符合性检查
   └─ 文档更新检查
   ↓
3. 发现问题时主动提醒
   - 提供具体问题位置
   - 提供修正建议
   - 提供代码示例
   ↓
4. 自动修复简单问题
   - 代码格式化
   - 导入整理
   - 文档生成
   ↓
5. 生成检查报告
   - ✅ 符合规范的项目
   - ⚠️ 需要整改的问题
   - 💡 优化建议
```

***

## 使用示例

### 示例 1：代码完成后检查

**用户**：代码写完了

**技能响应**：

```
好的，我将执行代码完成后检查。

**正在执行检查**：
1. ✅ 正在格式化代码...
2. ✅ 正在执行类型检查...
3. ✅ 正在生成 API 文档...
4. ✅ 正在验证项目完整性...
5. ✅ 正在执行后端构建验证...
6. ✅ 正在执行前端构建验证...

✅ 检查完成！

**检查结果**：
- 代码格式化：通过（修复 3 个问题）
- 类型检查：通过
- API 文档：已更新（5 个 API）
- 项目编译：通过
- 后端构建：通过（生成了 build/ixpay-pro.exe）
- 前端构建：通过（生成了 dist/ 目录）

**修复的问题**：
- internal/app/base/api/user_handler.go - 格式化
- src/views/user/index.vue - 格式化
- docs/swagger.json - 已更新

**构建验证结果**：
【后端】
✅ 构建成功：生成了 build/ixpay-pro.exe 可执行文件
✅ 无编译错误
✅ 无类型错误

【前端】
✅ 构建成功：生成了 dist/ 目录
✅ 无 TypeScript 类型错误
✅ 无 ESLint 错误
✅ 资源文件已正确打包

代码可以提交了！

建议提交信息：feat: 完成用户管理功能
```

### 示例 2：规范符合性检查

**用户**：检查代码规范

**技能响应**：

```
好的，我将执行全面的规范符合性检查。

**检查范围**：
- 后端 Go 代码：15 个文件
- 前端 Vue/TS 代码：28 个文件

**检查结果**：
✅ 代码风格 - 通过（95 分）
⚠️ 命名规范 - 发现 3 个问题
✅ 架构规范 - 通过
⚠️ 响应格式 - 发现 2 个问题
⚠️ 错误处理 - 发现 1 个问题

**详细问题**：

【后端】
1. ⚠️ internal/app/base/api/user_handler.go:50
   JSON 标签使用 snake_case
   当前：`json:"created_at"`
   建议：`json:"createdAt"`

2. ⚠️ internal/app/base/api/user_handler.go:120
   ID 字段未使用 string 类型
   当前：`ID int64`
   建议：`ID string`

3. ⚠️ internal/domain/base/service/user_domain_service.go:80
   错误处理未使用 fmt.Errorf 包装
   当前：`return nil, err`
   建议：`return nil, fmt.Errorf("获取用户失败 (ID=%s): %w", id, err)`

【前端】
4. ⚠️ src/types/user.ts:10
   字段命名使用 snake_case
   当前：`user_name: string`
   建议：`userName: string`

**修复建议**：
1. 立即修复：ID 字段类型问题（影响功能）
2. 尽快修复：JSON 标签命名问题（影响前后端一致性）
3. 尽快修复：错误处理问题（影响错误追踪）

需要我帮你自动修复这些问题吗？
```

***
