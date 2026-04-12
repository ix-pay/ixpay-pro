# IXPay Pro 后端项目设计文档

## 一、项目概述

### 1.1 项目简介

IXPay Pro 是一个基于 Go 语言和 Gin 框架的高性能支付管理系统，专注于提供微信支付解决方案，集成完整的后台管理系统。系统采用前后端分离架构，后端提供 RESTful API，前端使用 Vue3 + Element Plus 构建现代化界面。

### 1.2 项目目标

- 提供安全、稳定、高效的支付处理能力
- 实现完整的后台管理功能，包括用户、角色、权限、菜单等管理
- 支持微信支付等多种支付方式
- 提供丰富的 API 接口，方便第三方系统集成
- 采用现代化的技术栈和架构设计，确保系统的可扩展性和可维护性

### 1.3 系统架构特点

- **模块化设计**：清晰的分层架构（API 层、服务层、数据访问层）
- **RESTful API**：遵循 RESTful 设计规范
- **权限体系**：基于 RBAC+ABAC 混合模型的权限管理
- **缓存机制**：使用 Redis 缓存权限信息和热点数据
- **中间件**：实现了认证、权限验证、操作日志等中间件
- **依赖注入**：使用 Wire 实现编译时依赖注入
- **统一错误处理**：实现了全局错误处理机制
- **完善的日志**：使用 Zap 实现高性能日志记录

## 二、系统架构设计

### 2.1 分层架构

```
server/
├── cmd/
│   └── ixpay-pro/
│       └── main.go              # 应用程序入口
│
├── configs/
│   └── config.yaml              # 配置文件
│
├── docs/                        # Swagger 文档（自动生成）
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── internal/                    # 核心业务代码
│   ├── app/                     # 【应用层】协调领域层，处理 HTTP 请求
│   │   ├── base/                # 基础管理模块
│   │   │   ├── api/             # API 处理器（接收请求、参数验证、返回响应）
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── user_handler.go
│   │   │   │   ├── role_handler.go
│   │   │   │   └── ...
│   │   │   ├── application/     # 应用服务（可选，复杂业务协调多个领域服务）
│   │   │   │   ├── user_app_service.go
│   │   │   │   └── ...
│   │   │   ├── middleware/      # 模块中间件
│   │   │   │   ├── auth_middleware.go
│   │   │   │   └── ...
│   │   │   ├── migrations/      # 数据库迁移文件
│   │   │   ├── seed/            # 数据库种子文件
│   │   │   ├── application.go   # 应用分入口
│   │   │   ├── wire.go          # 依赖注入配置
│   │   │   └── routes.go        # 路由配置
│   │   ├── wx/                  # 微信支付模块（结构同 base）
│   │   │   ├── api/
│   │   │   ├── application/
│   │   │   ├── middleware/
│   │   │   ├── migrations/
│   │   │   ├── application.go
│   │   │   ├── wire.go
│   │   │   └── routes.go
│   │   ├── application.go       # 应用总入口
│   │   ├── routes.go
│   │   ├── wire_gen.go          # 依赖注入配置（自动生成）
│   │   └── wire.go
│   │
│   ├── domain/                  # 【领域层】核心业务逻辑，不依赖基础设施
│   │   ├── base/                # 基础管理领域
│   │   │   ├── entity/          # 领域实体（纯业务模型，无 ORM 标签）
│   │   │   │   ├── user.go
│   │   │   │   ├── role.go
│   │   │   │   └── ...
│   │   │   ├── repo/            # 仓库接口定义（依赖抽象）
│   │   │   │   ├── user_repository.go
│   │   │   │   └── ...
│   │   │   │   └── mock/        # Mock 实现（用于单元测试）
│   │   │   │       └── user_repository_mock.go
│   │   │   └── service/         # 领域服务（核心业务逻辑）
│   │   │       ├── user_domain_service.go
│   │   │       └── ...
│   │   ├── wx/                  # 支付领域（结构同 base）
│   │   │   ├── entity/
│   │   │   ├── repo/
│   │   │   │   └── mock/
│   │   │   └── service/
│   │   └── shared/              # 共享领域对象
│   │       └── value_object/    # 值对象
│   │
│   ├── persistence/             # 【基础设施层】实现 Repository 接口
│   │   ├── base/                # Base 模块仓库实现
│   │   │   ├── user_repository.go
│   │   │   └── ...
│   │   ├── wx/                  # WX 模块仓库实现
│   │   │   ├── payment_repository.go
│   │   │   └── ...
│   │   └── common/              # 通用持久化工具
│   │       ├── base_repository.go
│   │       └── converter.go
│   │
│   ├── infrastructure/          # 【基础设施层】通用技术组件
│   │   ├── persistence/         # 数据持久化
│   │   │   ├── database/        # 数据库连接、迁移
│   │   │   ├── redis/           # Redis 客户端
│   │   │   └── cache/           # 缓存抽象
│   │   ├── transport/           # 传输层
│   │   │   ├── http/            # HTTP 响应工具
│   │   │   └── middleware/      # 全局中间件
│   │   ├── security/            # 安全相关
│   │   │   ├── auth/            # JWT、权限管理
│   │   │   └── captcha/         # 验证码
│   │   ├── observability/       # 可观测性
│   │   │   ├── logger/          # 日志
│   │   │   └── monitor/         # 监控
│   │   └── support/             # 支撑工具
│   │       ├── error/           # 错误处理
│   │       ├── snowflake/       # 分布式 ID
│   │       └── task/            # 任务调度
│   │
│   ├── dto/                     # 【数据传输对象】请求和响应数据结构
│   │   ├── base/
│   │   │   ├── request/         # 请求 DTO
│   │   │   │   ├── auth.go
│   │   │   │   └── user.go
│   │   │   └── response/        # 响应 DTO
│   │   │       ├── auth.go
│   │   │       └── user.go
│   │   └── wx/
│   │       ├── request/
│   │       └── response/
│   │
│   ├── utils/                   # 工具函数
│   │   ├── common/
│   │   └── encryption/
│   │
│   └── config/                  # 配置管理
│       └── config.go
│
├── tests/                       # 【测试目录】独立测试目录
│   ├── unit/                    # 单元测试
│   │   ├── domain/
│   │   │   ├── base/
│   │   │   └── wx/
│   │   └── app/
│   │       ├── base/
│   │       └── wx/
│   ├── integration/             # 集成测试
│   │   ├── api/
│   │   └── persistence/
│   └── e2e/                     # 端到端测试
│
├── scripts/                     # 【脚本目录】
│   ├── migrate.sh               # 数据库迁移脚本（Linux/Mac）
│   ├── migrate.bat              # 数据库迁移脚本（Windows）
│   ├── seed.sh                  # 种子数据脚本
│   ├── seed.bat                 # 种子数据脚本
│   └── generate_mocks.ps1       # Mock 生成脚本
│
├── .trae/
│   └── documents/               # 后端需求计划文档
├── .vscode/
├── .dockerignore
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

### 2.2 文件命名规则

#### 2.2.1 领域层（Domain）

| 文件类型 | 命名格式 | 示例 |
|---------|---------|------|
| 领域实体 | `{entity}_entity.go` 或 `{entity}.go` | `user.go`, `role.go` |
| 仓库接口 | `{entity}_repository.go` | `user_repository.go` |
| Mock 实现 | `{entity}_repository_mock.go` | `user_repository_mock.go` |
| 领域服务 | `{entity}_domain_service.go` | `user_domain_service.go` |

#### 2.2.2 应用层（Application）

| 文件类型 | 命名格式 | 示例 |
|---------|---------|------|
| API 处理器 | `{entity}_handler.go` | `user_handler.go` |
| 应用服务 | `{entity}_app_service.go` | `user_app_service.go` |
| 中间件 | `{feature}_middleware.go` | `auth_middleware.go` |

#### 2.2.3 持久化层（Persistence）

| 文件类型 | 命名格式 | 示例 |
|---------|---------|------|
| 仓库实现 | `{entity}_repository.go` | `user_repository.go` |
| 数据库模型 | `{entity}_model.go` | `user_model.go` |
| 转换器 | `converter.go` | `converter.go` |

#### 2.2.4 DTO 层

| 文件类型 | 命名格式 | 示例 |
|---------|---------|------|
| 请求 DTO | `{feature}.go` | `user.go`, `auth.go` |
| 响应 DTO | `{feature}.go` | `user.go`, `auth.go` |

#### 2.2.5 测试文件

| 测试类型 | 命名格式 | 示例 |
|---------|---------|------|
| 单元测试 | `{source_file}_test.go` | `user_domain_service_test.go` |
| 集成测试 | `{feature}_integration_test.go` | `user_api_integration_test.go` |
| E2E 测试 | `{feature}_e2e_test.go` | `user_registration_e2e_test.go` |

#### 2.2.6 通用规则

- ✅ 使用 **snake_case** 命名文件名
- ✅ 使用 **camelCase** 命名 JSON 字段
- ✅ ID 字段统一使用 **string** 类型（应用层和 DTO）
- ✅ 数据库 ID 使用 **int64** 类型
- ✅ 包名使用小写，不使用连字符

### 2.3 技术要点

#### 2.3.1 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Request                          │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  API Handler (应用层)                                    │
│  - 接收请求                                              │
│  - 参数验证（使用 DTO）                                  │
│  - 调用 Application Service                             │
│  - 返回响应（使用 DTO）                                  │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  Application Service (应用服务层 - 可选)                 │
│  - 协调多个 Domain Service                               │
│  - 事务管理                                              │
│  - 权限检查                                              │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  Domain Service (领域层 - 核心业务)                      │
│  - 纯业务逻辑                                            │
│  - 不依赖基础设施                                        │
│  - 可被其他模块调用 ⭐                                    │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  Repository Interface (领域层 - 接口定义)                │
│  - 定义数据访问接口                                      │
│  - 依赖抽象                                              │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  Repository Persistence (基础设施层 - 实现)           │
│  - 实现 Repository 接口                                  │
│  - 数据库操作（GORM）                                    │
│  - 数据转换（Database Model ↔ Domain Entity）           │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  Database (基础设施层)                                   │
│  - 数据库连接                                            │
│  - 基础 CRUD 操作                                        │
└─────────────────────────────────────────────────────────┘
```

#### 2.3.2 依赖方向

**核心原则**：依赖倒置

- ✅ Domain 层定义接口
- ✅ Persistence 层实现接口
- ✅ Domain 层不依赖任何基础设施
- ✅ 高层模块不依赖低层模块

**依赖链**：
```
API → Application → Domain → Repository Interface → Repository Persistence → Database
```

#### 2.3.3 类型转换策略

| 层级                | ID 类型    | 原因                       |
| ----------------- | --------- | ------------------------ |
| **数据库存储**      | `int64`   | ✅ 索引性能好，存储空间小      |
| **Domain Entity** | `string`  | ✅ 避免 JSON 精度丢失，前后端统一 |
| **DTO**           | `string`  | ✅ API 传输安全，JavaScript 友好 |

**转换责任**：Repository 层负责 `string ↔ int64` 的转换

#### 2.3.4 错误处理规范

```go
// ✅ 正确：使用 fmt.Errorf 包装错误，保留错误链
func (s *UserDomainService) GetUserByID(id string) (*entity.User, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("获取用户失败 (ID=%s): %w", id, err)
    }
    return user, nil
}

// ❌ 错误：直接返回原始错误，丢失上下文
func (s *UserDomainService) GetUserByID(id string) (*entity.User, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        return nil, err  // ❌ 丢失上下文
    }
    return user, nil
}
```

#### 2.3.5 跨模块调用

**场景**：WX 模块需要验证用户信息

**实现方式**：
- ✅ WX 模块的 Application Service 注入 Base 模块的 Domain Service
- ✅ 直接调用 Domain 层（稳定依赖）
- ❌ 不调用 Application 层或 API 层（避免循环依赖）

**示例**：
```go
// WX 模块的 Application Service
type PaymentAppService struct {
    userDomainService    *base_domain.UserDomainService  // ⭐ 注入 Base 模块
    paymentDomainService *wx_domain.PaymentDomainService
}
```

#### 2.3.6 测试策略

**三层测试**：

1. **单元测试**（`tests/unit/`）
   - 测试 Domain Service 的业务逻辑
   - 使用 Mock Repository
   - 不依赖数据库

2. **集成测试**（`tests/integration/`）
   - 测试 API 层与数据库的集成
   - 使用测试数据库
   - 验证完整链路

3. **E2E 测试**（`tests/e2e/`）
   - 模拟真实用户场景
   - 完整业务流程测试

### 2.4 技术选型

| 技术组件         | 名称/版本                                    | 说明 |
|------------------|----------------------------------------------|------|
| **开发语言**     | Go 1.24.6                                    | 核心开发语言，提供高性能和并发处理能力 |
| **Web 框架**      | Gin v1.11.0                                  | 轻量级 HTTP 服务框架，提供路由、中间件等功能 |
| **依赖注入**     | Wire v0.7.0                                  | 编译时依赖注入工具，提高代码可维护性 |
| **数据库**       | PostgreSQL 13+                               | 强大的开源关系型数据库，支持复杂查询和事务 |
| **ORM**          | GORM v1.31.1                                 | 功能丰富的 ORM 库，简化数据库操作 |
| **缓存**         | Redis 6+                                     | 高性能键值存储，用于缓存和会话管理 |
| **认证**         | JWT v5.3.0                                   | 无状态身份认证令牌，支持跨服务认证 |
| **配置管理**     | Viper v1.21.0                                | 灵活的配置文件管理工具，支持多种配置格式 |
| **日志**         | Zap v1.27.1                                  | 高性能结构化日志库，支持多级别日志 |
| **任务调度**     | Cron v3.0.1                                  | 定时任务调度库，用于执行周期性任务 |
| **API 文档**      | Swagger v1.16.6                              | 自动 API 文档生成工具，方便接口调试和对接 |
| **监控**         | Prometheus v0.66.1                           | 开源监控系统，用于系统性能监控 |
| **限流**         | golang.org/x/time v0.14.0                    | API 速率限制库，防止系统过载 |
| **雪花算法**     | Snowflake                                    | 分布式 ID 生成算法，确保数据唯一性 |
| **验证码**       | base64Captcha v1.3.8                         | 验证码生成和验证库，提高系统安全性 |

## 三、核心功能模块

### 3.1 基础管理模块

#### 3.1.1 用户管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 用户注册           | 支持用户名、密码注册                           | ✅ 已实现 |
| 用户登录           | 支持用户名密码登录、验证码验证                 | ✅ 已实现 |
| 用户信息管理       | 获取和更新用户基本信息                         | ✅ 已实现 |
| 密码管理           | 修改密码、重置密码                             | ✅ 已实现 |
| 用户设置           | 获取和设置用户个性化配置                       | ✅ 已实现 |
| 用户列表           | 分页查询用户列表                               | ✅ 已实现 |
| 用户状态管理       | 启用/禁用用户账号                              | ✅ 已实现 |

**核心接口**：
- `POST /api/admin//auth/register` - 用户注册
- `POST /api/admin//auth/login` - 用户登录
- `GET /api/admin//user/info` - 获取当前用户信息
- `PUT /api/admin//user/info` - 更新用户信息
- `GET /api/admin//user` - 获取用户列表
- `POST /api/admin//user` - 添加用户
- `DELETE /api/admin//user/:id` - 删除用户
- `PUT /api/admin//user/password` - 修改密码
- `PUT /api/admin//user/reset-password` - 重置密码

#### 3.1.2 角色管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 角色创建           | 创建新角色，设置角色名称、编码、描述等         | ✅ 已实现 |
| 角色查询           | 获取角色详情、角色列表                         | ✅ 已实现 |
| 角色更新           | 更新角色信息                                   | ✅ 已实现 |
| 角色删除           | 删除角色                                       | ✅ 已实现 |
| 角色权限分配       | 为角色分配菜单、API 路由、按钮权限             | ✅ 已实现 |
| 角色用户分配       | 为角色分配用户                                 | ✅ 已实现 |
| 角色类型管理       | 系统角色、业务角色、数据角色分类               | ✅ 已实现 |
| 角色状态管理       | 启用/禁用角色                                  | ✅ 已实现 |

**核心接口**：
- `POST /api/admin//roles` - 创建角色
- `GET /api/admin//roles` - 获取角色列表
- `GET /api/admin//roles/detail` - 获取角色详情
- `PUT /api/admin//roles` - 更新角色
- `DELETE /api/admin//roles` - 删除角色
- `POST /api/admin//roles/assign-users` - 分配用户到角色
- `POST /api/admin//roles/assign-menus` - 分配菜单到角色
- `POST /api/admin//roles/assign-api-routes` - 分配 API 路由到角色

#### 3.1.3 权限管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| RBAC 基础模型       | 基于角色的访问控制                             | ✅ 已实现 |
| ABAC 扩展模型       | 基于属性的访问控制，支持复杂权限表达式         | ✅ 已实现 |
| 权限组管理         | 将相关权限分组管理                             | ✅ 已实现 |
| 权限规则管理       | 基于 ABAC 的权限规则管理                       | ✅ 已实现 |
| API 权限验证       | 检查用户是否有 API 访问权限                    | ✅ 已实现 |
| 按钮权限验证       | 检查用户是否有按钮权限                         | ✅ 已实现 |
| 权限缓存           | 使用 Redis 缓存用户权限信息                    | ✅ 已实现 |
| 权限刷新           | 手动刷新用户权限缓存                           | ✅ 已实现 |

**权限验证流程**：
1. 检查用户特殊权限（优先级最高）
2. 检查用户角色的直接权限
3. 检查用户角色所属权限组的权限
4. 检查 ABAC 权限规则

**权限模型**：
```
用户 -> 角色 -> 权限
     -> 用户特殊权限

角色 -> 权限组 -> 权限
```

#### 3.1.4 菜单管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 菜单创建           | 创建新菜单，支持目录、菜单、按钮、链接类型     | ✅ 已实现 |
| 菜单查询           | 获取菜单列表、菜单树形结构                     | ✅ 已实现 |
| 菜单更新           | 更新菜单信息                                   | ✅ 已实现 |
| 菜单删除           | 删除菜单                                       | ✅ 已实现 |
| 菜单分页           | 分页查询菜单                                   | ✅ 已实现 |
| 动态菜单生成       | 根据用户权限动态生成可访问菜单                 | ✅ 已实现 |
| 菜单类型管理       | 目录、菜单、按钮、链接四种类型                 | ✅ 已实现 |
| 菜单状态管理       | 启用/禁用菜单                                  | ✅ 已实现 |

**核心接口**：
- `POST /api/admin//menu` - 添加菜单
- `GET /api/admin//menu` - 获取菜单列表
- `GET /api/admin//menu/page` - 分页获取菜单
- `PUT /api/admin//menu` - 更新菜单
- `DELETE /api/admin//menu/:id` - 删除菜单

#### 3.1.5 API 路由管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| API 定义           | 定义 API 路由路径、方法、描述等                | ✅ 已实现 |
| API 查询           | 获取 API 路由列表、API 路由详情                | ✅ 已实现 |
| API 创建           | 创建新的 API 路由                               | ✅ 已实现 |
| API 更新           | 更新 API 路由信息                               | ✅ 已实现 |
| API 删除           | 删除 API 路由                                   | ✅ 已实现 |
| 路由分组           | API 路由按功能分组管理                         | ✅ 已实现 |
| 路由认证控制       | 设置 API 是否需要认证                          | ✅ 已实现 |
| 自动收集路由       | 启动时自动收集并保存路由到数据库               | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//apis` - 获取 API 路由列表
- `GET /api/admin//apis/:id` - 获取 API 路由详情
- `POST /api/admin//apis` - 创建 API 路由
- `PUT /api/admin//apis/:id` - 更新 API 路由
- `DELETE /api/admin//apis/:id` - 删除 API 路由

#### 3.1.6 按钮权限管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 按钮权限创建       | 创建按钮权限，设置编码、名称等                 | ✅ 已实现 |
| 按钮权限查询       | 获取按钮权限列表、详情                         | ✅ 已实现 |
| 按钮权限更新       | 更新按钮权限信息                               | ✅ 已实现 |
| 按钮权限删除       | 删除按钮权限                                   | ✅ 已实现 |
| 按钮权限分配       | 为角色分配按钮权限                             | ✅ 已实现 |
| API 路由关联       | 为按钮权限关联 API 路由                        | ✅ 已实现 |
| 按钮权限撤销       | 从角色撤销按钮权限                             | ✅ 已实现 |
| 按钮权限查询       | 根据角色、菜单查询按钮权限                     | ✅ 已实现 |

**核心接口**：
- `POST /api/admin//btn-perms` - 创建按钮权限
- `GET /api/admin//btn-perms` - 获取按钮权限列表
- `GET /api/admin//btn-perms/detail` - 获取按钮权限详情
- `PUT /api/admin//btn-perms` - 更新按钮权限
- `DELETE /api/admin//btn-perms` - 删除按钮权限
- `POST /api/admin//btn-perms/assign-to-role` - 分配按钮权限到角色
- `POST /api/admin//btn-perms/assign-api-routes` - 分配 API 路由到按钮权限

#### 3.1.7 配置管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 配置创建           | 创建系统配置项                                 | ✅ 已实现 |
| 配置查询           | 获取配置列表、根据键查询配置                   | ✅ 已实现 |
| 配置更新           | 更新配置项                                     | ✅ 已实现 |
| 配置删除           | 删除配置项                                     | ✅ 已实现 |
| 激活配置           | 获取所有启用的配置                             | ✅ 已实现 |
| 配置分类           | 按分类管理配置项                               | ✅ 已实现 |
| 配置状态管理       | 启用/禁用配置项                                | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//config` - 获取配置列表
- `GET /api/admin//config/key` - 根据键获取配置
- `GET /api/admin//config/:id` - 根据 ID 获取配置
- `POST /api/admin//config` - 创建配置
- `PUT /api/admin//config` - 更新配置
- `DELETE /api/admin//config/:id` - 删除配置
- `GET /api/admin//config/active` - 获取所有启用的配置

#### 3.1.8 字典管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 字典表管理         | 字典表的增删改查                               | ✅ 已实现 |
| 字典项管理         | 字典项的增删改查                               | ✅ 已实现 |
| 字典编码           | 使用唯一编码标识字典表                         | ✅ 已实现 |
| 字典查询           | 根据编码查询字典表及其项                       | ✅ 已实现 |
| 字典排序           | 支持字典项排序                                 | ✅ 已实现 |
| 字典状态管理       | 启用/禁用字典表和字典项                        | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//dict` - 获取字典列表
- `GET /api/admin//dict/code` - 根据编码获取字典
- `GET /api/admin//dict/:id` - 根据 ID 获取字典
- `POST /api/admin//dict` - 创建字典
- `PUT /api/admin//dict` - 更新字典
- `DELETE /api/admin//dict/:id` - 删除字典
- `GET /api/admin//dict/items` - 获取字典项列表
- `POST /api/admin//dict/item` - 创建字典项

#### 3.1.9 操作日志

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 日志记录           | 自动记录用户操作日志                           | ✅ 已实现 |
| 日志查询           | 分页查询操作日志                               | ✅ 已实现 |
| 日志详情           | 获取操作日志详情                               | ✅ 已实现 |
| 日志统计           | 获取操作日志统计信息                           | ✅ 已实现 |
| 日志删除           | 删除单条或批量删除操作日志                     | ✅ 已实现 |
| 日志清理           | 按时间范围清理操作日志                         | ✅ 已实现 |
| 自动捕获           | 通过中间件自动捕获请求和响应                   | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//logs` - 获取操作日志列表
- `GET /api/admin//logs/:id` - 获取操作日志详情
- `DELETE /api/admin//logs/:id` - 删除操作日志
- `POST /api/admin//logs/batch-delete` - 批量删除操作日志
- `GET /api/admin//logs/statistics` - 获取操作日志统计
- `POST /api/admin//logs/clear` - 按时间范围清理日志

#### 3.1.10 组织架构管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 部门管理           | 部门的增删改查                                 | ✅ 已实现 |
| 部门树形结构       | 获取部门树形结构                               | ✅ 已实现 |
| 部门负责人         | 设置和更新部门负责人                           | ✅ 已实现 |
| 岗位管理           | 岗位的增删改查                                 | ✅ 已实现 |
| 岗位列表           | 获取所有岗位                                   | ✅ 已实现 |
| 部门状态管理       | 启用/禁用部门                                  | ✅ 已实现 |
| 岗位状态管理       | 启用/禁用岗位                                  | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//dept` - 获取部门列表
- `GET /api/admin//dept/tree` - 获取部门树形结构
- `GET /api/admin//dept/:id` - 获取部门详情
- `POST /api/admin//dept` - 创建部门
- `PUT /api/admin//dept` - 更新部门
- `DELETE /api/admin//dept/:id` - 删除部门
- `PUT /api/admin//dept/:id/leader` - 更新部门负责人
- `GET /api/admin//position` - 获取岗位列表
- `POST /api/admin//position` - 创建岗位
- `PUT /api/admin//position` - 更新岗位
- `DELETE /api/admin//position/:id` - 删除岗位

#### 3.1.11 登录日志

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 登录记录           | 记录用户登录信息                               | ✅ 已实现 |
| 日志查询           | 分页查询登录日志                               | ✅ 已实现 |
| 登录统计           | 获取登录统计信息                               | ✅ 已实现 |
| 异常登录查询       | 查询异常登录记录                               | ✅ 已实现 |
| 自动记录           | 登录成功后自动记录                             | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//login-log` - 获取登录日志列表
- `GET /api/admin//login-log/:id` - 获取登录日志详情
- `GET /api/admin//login-log/statistics` - 获取登录统计
- `GET /api/admin//login-log/abnormal` - 获取异常登录查询

#### 3.1.12 在线用户

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 在线用户列表       | 获取当前在线用户列表                           | ✅ 已实现 |
| 在线用户详情       | 获取单个在线用户详情                           | ✅ 已实现 |
| 在线用户数量       | 获取在线用户数量                               | ✅ 已实现 |
| 在线状态检查       | 检查用户是否在线                               | ✅ 已实现 |
| 强制下线           | 强制用户下线                                   | ✅ 已实现 |
| 批量强制下线       | 批量强制多个用户下线                           | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//online-user` - 获取在线用户列表
- `GET /api/admin//online-user/:user_id` - 获取在线用户详情
- `GET /api/admin//online-user/count` - 获取在线用户数量
- `GET /api/admin//online-user/online` - 检查用户是否在线
- `DELETE /api/admin//online-user/:user_id` - 强制用户下线

#### 3.1.13 公告管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 公告创建           | 创建新公告                                     | ✅ 已实现 |
| 公告查询           | 获取公告列表、公告详情                         | ✅ 已实现 |
| 公告更新           | 更新公告信息                                   | ✅ 已实现 |
| 公告删除           | 删除公告                                       | ✅ 已实现 |
| 公告发布           | 发布公告                                       | ✅ 已实现 |
| 已读标记           | 标记公告已读                                   | ✅ 已实现 |
| 已读检查           | 检查公告是否已读                               | ✅ 已实现 |
| 公告统计           | 获取公告统计信息                               | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//notice` - 获取公告列表
- `GET /api/admin//notice/:id` - 获取公告详情
- `POST /api/admin//notice` - 创建公告
- `PUT /api/admin//notice` - 更新公告
- `DELETE /api/admin//notice/:id` - 删除公告
- `POST /api/admin//notice/:id/publish` - 发布公告
- `POST /api/admin//notice/:id/read` - 标记公告已读
- `GET /api/admin//notice/:id/is-read` - 检查公告是否已读
- `GET /api/admin//notice/statistics` - 获取公告统计

#### 3.1.14 系统监控

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 系统资源监控       | 监控 CPU、内存、磁盘等系统资源                 | ✅ 已实现 |
| 缓存监控           | 监控 Redis 缓存状态、键数量等                  | ✅ 已实现 |
| 数据库监控         | 监控数据库连接数、慢查询等                     | ✅ 已实现 |
| Redis 键查询       | 查询 Redis 中的键                              | ✅ 已实现 |
| 慢查询日志         | 查询数据库慢查询日志                           | ✅ 已实现 |
| Prometheus 集成     | 集成 Prometheus 监控系统                       | ✅ 已实现 |

**核心接口**：
- `GET /api/admin//monitor/system` - 获取系统资源监控
- `GET /api/admin//monitor/cache` - 获取缓存监控
- `GET /api/admin//monitor/database` - 获取数据库监控
- `GET /api/admin//monitor/redis-keys` - 查询 Redis 键
- `GET /api/admin//monitor/slow-queries` - 查询慢查询日志

#### 3.1.15 任务管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 任务创建           | 创建定时任务                                   | ✅ 已实现 |
| 任务查询           | 获取任务列表、任务详情                         | ✅ 已实现 |
| 任务启动           | 启动任务                                       | ✅ 已实现 |
| 任务停止           | 停止任务                                       | ✅ 已实现 |
| 任务重试           | 重试失败的任务                                 | ✅ 已实现 |
| 任务删除           | 删除任务                                       | ✅ 已实现 |
| 任务分组           | 设置任务分组                                   | ✅ 已实现 |
| 执行日志           | 获取任务执行日志                               | ✅ 已实现 |
| 任务统计           | 获取任务统计信息                               | ✅ 已实现 |

**核心接口**：
- `POST /api/admin//task` - 添加任务
- `GET /api/admin//task` - 获取任务列表
- `GET /api/admin//task/:id` - 获取任务详情
- `POST /api/admin//task/:id/start` - 启动任务
- `POST /api/admin//task/:id/stop` - 停止任务
- `POST /api/admin//task/:id/retry` - 重试任务
- `DELETE /api/admin//task/:id` - 删除任务
- `GET /api/admin//task/:id/execution-logs` - 获取任务执行日志
- `GET /api/admin//task/statistics` - 获取任务统计
- `POST /api/admin//task/:id/group` - 设置任务分组

### 3.2 微信支付模块

#### 3.2.1 微信认证

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 微信登录           | 支持微信授权登录                               | ✅ 已实现 |
| 会话管理           | 管理微信认证会话                               | ✅ 已实现 |
| 用户信息同步       | 同步微信用户信息                               | ✅ 已实现 |

**核心接口**：
- `POST /api//wx/auth/login` - 微信登录
- `GET /api//wx/auth/session` - 获取认证会话

#### 3.2.2 支付管理

| 功能点             | 描述                                           | 实现状态 |
|--------------------|------------------------------------------------|----------|
| 创建支付           | 创建支付订单                                   | ✅ 已实现 |
| 查询支付           | 查询支付详情                                   | ✅ 已实现 |
| 取消支付           | 取消支付订单                                   | ✅ 已实现 |
| 支付通知           | 处理微信支付通知                               | ✅ 已实现 |
| 支付列表           | 获取用户支付列表                               | ✅ 已实现 |

**核心接口**：
- `POST /api//payment` - 创建支付
- `GET /api//payment/:id` - 查询支付详情
- `PUT /api//payment/:id/cancel` - 取消支付
- `POST /api//payment/notify` - 微信支付通知

## 四、系统架构实现

### 4.1 路由架构

系统采用 `/api/admin/` 作为基础管理模块的路由前缀，`/api/` 作为微信支付模块的路由前缀，保持向后兼容性：

```go
// 路由结构示例
// 基础管理模块路由
admin := router.Group("/api/admin")
{
    // 公共路由（无需认证）
    public := admin.Group("/")
    {
        auth.POST("/register", userController.Register)
        auth.POST("/login", authController.Login)
        auth.POST("/captcha", authController.Captcha)
    }

    // 需要认证的路由
    authenticated := admin.Group("/")
    authenticated.Use(AuthMiddleware())
    authenticated.Use(PermissionMiddleware())
    {
        // 用户管理路由
        user := authenticated.Group("/user")
        {
            user.GET("/info", userController.GetUserInfo)
            user.PUT("/info", userController.UpdateUserInfo)
            // ...
        }
        
        // 角色管理路由
        role := authenticated.Group("/roles")
        {
            role.POST("", roleController.CreateRole)
            role.GET("", roleController.GetRoleList)
            // ...
        }
    }
}

// 微信支付模块路由
wx := router.Group("/api/")
{
    wx.Use(WXAuthMiddleware())
    // 微信认证路由
    wxAuth := wx.Group("/wx/auth")
    {
        wxAuth.POST("/login", wxAuthController.Login)
        // ...
    }
    
    // 支付管理路由
    payment := wx.Group("/payment")
    {
        payment.POST("", paymentController.CreatePayment)
        // ...
    }
}
```

### 4.2 中间件架构

系统实现了完善的中间件体系，包括：

- **认证中间件**：验证 JWT 令牌，解析用户信息
- **权限中间件**：验证用户是否有权限访问资源
- **操作日志中间件**：记录用户操作日志
- **错误处理中间件**：统一处理错误和异常
- **日志中间件**：记录请求日志
- **CORS 中间件**：处理跨域请求
- **速率限制中间件**：限制 API 请求频率
- **熔断中间件**：实现服务熔断保护
- **缓存中间件**：实现响应缓存
- **Prometheus 中间件**：收集监控指标

### 4.3 数据库设计

#### 4.3.1 基础模型

所有数据模型都继承自基础模型，包含通用字段：

```go
type BaseModel struct {
    ID        int64     `gorm:"primarykey"`  // 雪花算法生成的 ID
    CreatedAt time.Time // 创建时间
    UpdatedAt time.Time // 更新时间
}
```

#### 4.3.2 核心数据模型

- **User**：用户模型，包含用户名、密码、邮箱、手机等字段
- **Role**：角色模型，包含角色名称、编码、类型等字段
- **Menu**：菜单模型，包含菜单名称、路径、类型等字段
- **API**：API 路由模型，包含路径、方法、描述等字段
- **BtnPerm**：按钮权限模型，包含编码、名称等字段
- **PermissionGroup**：权限组模型
- **PermissionRule**：权限规则模型（ABAC）
- **Department**：部门模型
- **Position**：岗位模型
- **Config**：配置模型
- **Dict**：字典模型
- **OperationLog**：操作日志模型
- **LoginLog**：登录日志模型
- **Notice**：公告模型
- **Task**：任务模型
- **TaskExecutionLog**：任务执行日志模型

### 4.4 缓存架构

系统使用 Redis 实现多级缓存：

- **权限缓存**：缓存用户权限信息，减少数据库查询
- **会话缓存**：缓存用户会话信息
- **配置缓存**：缓存系统配置
- **字典缓存**：缓存字典数据
- **在线用户缓存**：缓存在线用户信息

### 4.5 安全机制

- **JWT 认证**：无状态身份认证
- **密码加密**：使用 bcrypt 加密存储密码
- **验证码**：支持图形验证码
- **输入验证**：验证所有用户输入
- **SQL 注入防护**：使用 GORM 参数化查询
- **XSS 防护**：过滤用户输入
- **CSRF 防护**：CSRF Token 验证
- **速率限制**：防止暴力攻击
- **权限控制**：细粒度的权限验证

### 4.6 日志系统

使用 Zap 实现高性能日志记录：

- **日志分级**：DEBUG、INFO、WARN、ERROR
- **日志格式**：JSON 格式，便于日志收集
- **日志上下文**：包含请求 ID、用户 ID、IP 等信息
- **操作日志**：记录所有用户操作
- **请求日志**：记录所有 HTTP 请求

### 4.7 监控系统

集成 Prometheus 监控系统：

- **系统指标**：CPU、内存、磁盘等
- **HTTP 指标**：请求数、响应时间、错误率等
- **数据库指标**：连接数、查询时间等
- **缓存指标**：命中率、键数量等
- **业务指标**：自定义业务指标

## 五、种子数据管理

### 5.1 种子数据概念

种子数据是指系统初始化时需要加载的基础数据，包括角色、用户、菜单、API 路由等。种子数据管理确保系统在首次启动或重置时能够自动加载必要的基础数据，保证系统正常运行。

### 5.2 设计原则

- **模块化设计**：按功能模块划分种子数据，便于维护和扩展
- **增量更新**：支持只初始化新增或修改的种子数据
- **版本管理**：为每个种子数据模块添加版本标识
- **事务支持**：确保种子数据初始化的一致性
- **配置驱动**：支持通过配置控制种子数据的初始化

### 5.3 种子数据模块

- **角色种子数据**：初始化系统管理员、普通用户等默认角色
- **用户种子数据**：初始化系统管理员用户
- **菜单种子数据**：初始化系统菜单
- **API 路由种子数据**：初始化系统 API 路由

### 5.4 初始化流程

1. 系统启动时，加载配置
2. 检查是否需要初始化种子数据（通过 `config.Server.InitSeedData` 配置）
3. 注册所有种子数据模块
4. 按顺序初始化种子数据
5. 记录初始化结果

## 六、配置管理

### 6.1 配置文件

主要配置文件位于 `configs/config.yaml`：

```yaml
# 服务器配置
server:
  port: 8586
  mode: "debug"  # debug, release, test
  init_seed_data: true  # 是否初始化种子数据
  update_routes_on_start: true  # 启动时是否更新路由

# 数据库配置
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  user: "ixpay"
  password: "ixpay123"
  dbname: "ixpay_pro"
  sslmode: "disable"

# Redis 配置
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# JWT 配置
jwt:
  secret: "your-secret-key"
  expire: 3600

# 日志配置
logging:
  level: "info"
  file: "logs/"
```

### 6.2 多环境配置

系统支持多环境配置：

- **开发环境**：`configs/config.yaml`
- **测试环境**：`configs/config.test.yaml`
- **生产环境**：`configs/config.prod.yaml`

启动服务时，可以通过 `--config` 参数指定配置文件。

## 七、部署架构

### 7.1 技术架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Nginx      │────▶│  API 网关    │────▶│  应用服务    │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                       ┌─────────────┐     ┌───┴───┐     ┌─────────────┐
                       │  Redis      │◀────▶│  负载均衡  │◀────▶│  应用服务    │
                       └─────────────┘     └──────┬──────┘     └─────────────┘
                                               │
                       ┌─────────────┐     ┌───┴───┐     ┌─────────────┐
                       │  PostgreSQL  │◀────▶│  应用服务    │◀────▶│  应用服务    │
                       └─────────────┘     └─────────────┘     └─────────────┘
```

### 7.2 Docker 部署

系统支持 Docker 容器化部署，提供完整的 Docker Compose 配置：

```yaml
version: '3.8'
services:
  ixpay-server:
    build: .
    ports:
      - "8586:8586"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=ixpay
      - DB_PASSWORD=ixpay123
      - DB_NAME=ixpay_pro
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:14
    environment:
      - POSTGRES_USER=ixpay
      - POSTGRES_PASSWORD=ixpay123
      - POSTGRES_DB=ixpay_pro

  redis:
    image: redis:7
```

## 八、总结

IXPay Pro 后端项目采用现代化的 DDD 分层架构设计，实现了完整的后台管理功能和微信支付功能。系统具有以下核心特点：

1. **清晰的分层架构**：
   - 领域层（Domain）：核心业务逻辑，可被其他模块直接调用
   - 应用层（Application）：协调领域层，处理 HTTP 请求
   - 基础设施层（Infrastructure）：实现技术细节，依赖领域层接口
   - 持久化层（Persistence）：实现 Repository 接口，负责数据转换

2. **技术优势**：
   - ✅ 业务逻辑清晰，易于维护
   - ✅ 模块解耦，支持独立开发
   - ✅ 符合 Go 语言最佳实践
   - ✅ 支持未来扩展
   - ✅ 依赖倒置原则，提高代码可测试性

3. **关键技术特性**：
   - **统一 ID 管理**：数据库使用 int64，应用层和 DTO 使用 string，避免 JSON 精度丢失
   - **跨模块调用**：通过注入 Domain Service 实现模块间通信
   - **完善的测试策略**：单元测试、集成测试、E2E 测试三层保障
   - **规范化命名**：统一的文件命名和编码规范
   - **错误处理规范**：使用 fmt.Errorf 包装错误，保留错误链

4. **部署与运维**：
   - 支持 Docker 容器化部署
   - 集成 Prometheus 监控系统
   - 完善的日志系统
   - 支持多环境配置

通过本项目的实施，可以构建一个功能完善、性能优良、安全可靠的支付管理系统，满足企业级应用的需求。新的架构设计为系统的持续迭代和扩展奠定了坚实的基础。
