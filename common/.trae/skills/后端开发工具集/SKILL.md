---
name: 后端开发工具集
description: 当用户进行 Go 后端开发时自动触发，执行 API 开发、代码构建、单元测试、依赖管理、Swagger 文档生成和 Docker 部署等任务，提升开发效率和代码质量
---

## 核心原则

1. **规范优先**：严格遵循项目 DDD 分层架构和代码规范
2. **自动化**：优先使用自动化工具完成重复性任务
3. **质量保障**：所有代码必须通过格式化、类型检查和测试验证
4. **统一响应**：所有 API 响应必须使用统一的响应格式
5. **日志完备**：关键操作必须记录日志，包含 user_id、IP 等关键信息

## 核心功能概览

1. **API 脚手架生成** - 自动生成符合 DDD 分层架构的完整代码框架
2. **代码格式化** - 自动执行 go fmt 和 go imports
3. **单元测试生成** - 自动生成 Go 单元测试，使用 testify 框架
4. **开发命令执行** - 构建、运行、依赖管理、Docker 部署
5. **API 文档生成** - 自动生成 Swagger 文档
6. **依赖注入配置** - 自动生成 Wire 依赖注入配置

## 工作流程

### 场景 1：新增 API 接口

```
用户请求 → 生成 Handler 骨架 → 生成 Service 接口 → 生成 Repository 接口
→ 生成 DTO → 生成 Wire 绑定 → 生成 Swagger 注释 → 生成单元测试
```

### 场景 2：修改现有 API

```
用户请求 → 修改 Handler 逻辑 → 修改 Service 实现 → 修改 Repository 实现
→ 更新 Swagger 文档 → 更新单元测试 → 执行测试验证
```

## 功能 1：API 脚手架生成与规范检查

### 自动检查项目

#### 1.1 API 路径检查

- 检测路由定义是否符合 `/api/[module]//` 格式
- 不符合时主动提醒并提供修正建议

#### 1.2 响应格式检查

- 检测是否使用统一响应格式 `github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes`
- 检测响应结构是否使用 `Result`、`Ok`、`Fail` 等标准方法

**正面示例**：
```go
baseRes.OkWithData(data, ctx)
baseRes.FailWithMessage("操作失败", ctx)
```

**反面示例**：
```go
ctx.JSON(200, gin.H{"code": 0, "data": data})  // ❌ 未使用统一响应格式
```

#### 1.3 分层架构检查

- 检测 API Handler 中是否有直接操作数据库的代码
- 发现 `db.Where()`、`db.First()` 等直接调用时提醒
- 检测领域层是否依赖基础设施层

**正面示例**：
```go
// Handler 层
func (c *UserController) GetUser(ctx *gin.Context) {
    user, err := c.service.GetByID(id)  // ✅ 调用 Service 层
    // ...
}

// Service 层
func (s *UserService) GetByID(id string) (*entity.User, error) {
    return s.repo.FindByID(id)  // ✅ 调用 Repository 层
}
```

**反面示例**：
```go
// Handler 层
func (c *UserController) GetUser(ctx *gin.Context) {
    var user entity.User
    c.db.First(&user, id)  // ❌ 直接操作数据库
}
```

#### 1.4 参数验证检查

- 检测字段是否有 `binding` 标签
- 检测是否有 `validate` 标签

**正面示例**：
```go
type RegisterRequest struct {
    Username string `json:"username" binding:"required" validate:"required,min=3,max=20"`
    Password string `json:"password" binding:"required" validate:"required,min=6"`
    Email    string `json:"email" binding:"required" validate:"required,email"`
}
```

#### 1.5 日志记录检查

- 检测登录、删除等操作是否有日志记录
- 检测日志是否包含 user_id、IP 等关键信息

**正面示例**：
```go
c.log.Info("用户登录成功", 
    "user_id", user.ID, 
    "username", user.Username,
    "ip", ctx.ClientIP(),
)
```

## 功能 2：代码格式化

**触发条件**：保存 Go 文件、提交代码前或用户明确要求

**自动执行**：
```bash
go fmt ./...
go imports -w .
```

**功能**：
- 自动格式化 Go 代码
- 自动导入缺失的包
- 移除未使用的导入

## 功能 3：单元测试生成

**触发条件**：用户说"生成单元测试"、"编写测试"

**自动生成**：
- 符合 Go 测试规范的单元测试文件
- 使用 testify 框架：`assert`、`require`
- 包含完整的测试用例和表驱动测试
- 自动生成 Mock 实现用于单元测试

**测试文件命名**：`*_test.go`

**测试函数命名**：`Test[ServiceName]_[Functionality]`

**正面示例**：
```go
func TestUserService_PasswordHandling(t *testing.T) {
    password := "testpassword123"
    
    hash, err := encryption.GeneratePasswordHash(password)
    require.NoError(t, err, "生成密码哈希失败")
    require.NotEmpty(t, hash, "生成的密码哈希为空")
    
    err = encryption.VerifyPassword(hash, password)
    assert.NoError(t, err, "验证正确密码失败")
}
```

## 功能 4：开发命令执行

### 4.1 项目构建

```bash
go build -o ./build/ixpay-pro.exe ./cmd/ixpay-pro/main.go
```

### 4.2 运行项目

```bash
go run cmd/ixpay-pro/main.go
```

### 4.3 依赖管理

```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy

# 添加新依赖
go get github.com/gin-gonic/gin
```

### 4.4 Wire 依赖注入代码生成

```bash
# 生成 Wire 依赖注入代码
wire gen ./internal/app
```

**Wire 配置文件位置**：
- 主配置：`internal/app/wire.go`
- Base 模块：`internal/app/base/wire.go`
- Wx 模块：`internal/app/wx/wire.go`

**ProviderSet 命名规范**：
```go
var ProviderSetBaseRepo = wire.NewSet(...)
var ProviderSetBaseService = wire.NewSet(...)
var ProviderSetBaseController = wire.NewSet(...)
```

### 4.5 Docker 部署

```bash
# 构建 Docker 镜像
docker build -t ixpay-pro .

# 运行容器
docker run -p 8586:8586 ixpay-pro
```

**Dockerfile 说明**：
- 使用多阶段构建减少镜像大小
- 构建阶段：`golang:1.24.6-alpine`
- 运行阶段：`alpine:3.19`
- 使用非 root 用户运行应用
- 包含健康检查配置

## 功能 5：API Swagger 文档自动生成

**触发条件**：新增或修改 API 接口后自动触发

**执行命令**：
```bash
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseDependency --parseInternal --parseDepth 1
```

**生成内容**：
- Swagger JSON 文件（docs/swagger.json）
- Swagger YAML 文件（docs/swagger.yaml）
- API 文档 HTML（docs/docs.go）

**文档访问**：http://localhost:8586/swagger/index.html

**Swagger 注释规范**：
```go
// @Summary      用户注册
// @Description  创建新用户账户
// @Tags         基础服务
// @Accept       json
// @Produce      json
// @Param        register  body  request.RegisterRequest  true  "注册请求参数"
// @Success      201  {object}  baseRes.Response{data=entity.User,msg=string}  "注册成功"
// @Failure      400  {object}  map[string]string  "请求参数错误"
// @Router       /api/admin//auth/register [post]
```

## 功能 6：错误处理规范

**错误包位置**：`github.com/ix-pay/ixpay-pro/internal/infrastructure/support/error`

**错误码分类**：
- 系统错误：10001-10009
- 认证授权错误：20001-20009
- 参数错误：30001-30009
- 业务逻辑错误：40001-40009
- 权限错误：50001-50009

**错误处理示例**：
```go
// 使用预定义错误
error.BadRequest("参数错误", nil)
error.Unauthorized("未授权", nil)
error.NotFound("资源不存在", nil)

// 自定义错误
error.NewAppError(400, error.ErrorCodeValidation, "参数验证失败", details)
```