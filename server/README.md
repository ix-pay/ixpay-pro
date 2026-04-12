# IXPay Pro

## 项目简介

IXPay Pro 是一个基于 Go 语言和 Gin 框架的高性能支付管理系统，专注于提供微信支付解决方案，集成完整的后台管理系统。系统采用前后端分离架构，后端提供 RESTful API，前端使用 Vue3 + Element Plus 构建现代化界面。

## 项目目标

- 提供安全、稳定、高效的支付处理能力
- 实现完整的后台管理功能，包括用户、角色、权限、菜单等管理
- 支持微信支付等多种支付方式
- 提供丰富的 API 接口，方便第三方系统集成
- 采用现代化的技术栈和架构设计，确保系统的可扩展性和可维护性

## 系统架构概述

IXPay Pro 采用前后端分离的模块化架构设计，具有清晰的层次结构和职责划分。

### 架构层次

1. **前端层**：基于 Vue3 + Element Plus 构建的现代化用户界面
2. **API 层**：基于 Gin 框架的 RESTful API 接口
3. **服务层**：实现核心业务逻辑的服务组件
4. **数据访问层**：与数据库交互的数据仓库
5. **基础设施层**：提供认证、缓存、日志等基础服务

### 模块划分

- **基础管理模块**：用户、角色、权限、菜单等核心管理功能
- **微信支付模块**：微信支付相关的功能实现
- **基础设施模块**：认证、缓存、日志、数据库等基础服务

### 技术特性

- **模块化设计**：清晰的分层架构，便于扩展和维护
- **RESTful API**：遵循 RESTful 设计规范，提供标准化的接口
- **权限体系**：基于 RBAC+ABAC 混合模型的权限管理
- **缓存机制**：使用 Redis 缓存权限信息和热点数据
- **中间件**：实现了认证、权限验证、操作日志等中间件
- **依赖注入**：使用 Wire 实现编译时依赖注入，提高代码可维护性
- **统一错误处理**：实现了全局错误处理机制
- **完善的日志**：使用 Zap 实现高性能日志记录

## 功能特性

### 后台管理功能

- **🔐 用户认证**：支持注册、登录、微信登录和令牌刷新，使用JWT进行身份验证
- **👮 权限管理**：基于RBAC+ABAC混合模型的权限管理，支持菜单、API路由和按钮级权限控制
- **👥 角色管理**：角色的增删改查、权限分配、角色继承和权限组管理
- **📋 菜单管理**：菜单的增删改查、树形结构管理，支持动态菜单生成
- **⚙️ 配置管理**：系统配置的增删改查，支持多环境配置
- **📚 字典管理**：字典表和字典项的管理，支持数据分类和标准化
- **📝 操作日志**：记录用户操作日志，支持日志查询和分析
- **🌱 种子数据管理**：系统初始化数据的管理，确保系统快速部署和配置

### 支付功能

- **💳 支付处理**：支持创建支付、查询支付、取消支付和处理微信支付通知
- **📱 微信支付**：集成微信支付API，支持扫码支付、H5支付等多种支付方式
- **💰 交易管理**：支付交易的查询、统计和分析

### 系统功能

- **📄 文档化**：集成Swagger API文档，方便接口调试和对接
- **🛑 优雅关闭**：支持信号处理和优雅关闭，确保服务稳定退出
- **🆔 分布式ID**：集成Snowflake算法生成唯一ID，确保数据一致性
- **🔑 验证码服务**：支持生成和验证验证码，提高系统安全性
- **🌐 跨域支持**：内置CORS中间件，解决前后端分离架构下的跨域问题
- **📊 监控系统**：支持Prometheus监控和Zap日志记录，确保系统稳定运行
- **🔒 安全防护**：内置输入验证、防SQL注入、防XSS攻击等安全措施
- **⚡ 性能优化**：使用Redis缓存、数据库索引等技术优化系统性能
- **📦 容器化部署**：支持Docker容器化部署，简化部署和运维

## 技术栈

### 后端

| 类别 | 技术/框架 | 版本 | 说明 |
|------|-----------|------|------|
| **开发语言** | Go | 1.24.6 | 核心开发语言，提供高性能和并发处理能力 |
| **Web框架** | Gin | v1.10.1 | 轻量级HTTP服务框架，提供路由、中间件等功能 |
| **依赖注入** | Wire | v0.7.0 | 编译时依赖注入工具，提高代码可维护性 |
| **数据库** | PostgreSQL | 13+ | 强大的开源关系型数据库，支持复杂查询和事务 |
| | GORM | v1.30.3 | 功能丰富的ORM库，简化数据库操作 |
| **缓存** | Redis | 6+ | 高性能键值存储，用于缓存和会话管理 |
| **认证** | JWT | v5.3.0 | 无状态身份认证令牌，支持跨服务认证 |
| **配置管理** | Viper | v1.20.1 | 灵活的配置文件管理工具，支持多种配置格式 |
| **日志** | Zap | v1.27.0 | 高性能结构化日志库，支持多级别日志 |
| **任务调度** | Cron | v3.0.1 | 定时任务调度库，用于执行周期性任务 |
| **API文档** | Swagger | - | 自动API文档生成工具，方便接口调试和对接 |
| **监控** | Prometheus | - | 开源监控系统，用于系统性能监控 |
| **限流** | golang.org/x/time/rate | - | API速率限制库，防止系统过载 |
| **雪花算法** | Snowflake | - | 分布式ID生成算法，确保数据唯一性 |
| **验证码** | captcha | - | 验证码生成和验证库，提高系统安全性 |

### 前端

| 类别 | 技术/框架 | 版本 | 说明 |
|------|-----------|------|------|
| **开发框架** | Vue 3 | - | 现代化前端框架，提供响应式数据绑定和组件化开发 |
| **UI组件库** | Element Plus | - | 基于Vue 3的UI组件库，提供丰富的界面元素 |
| **开发语言** | TypeScript | - | 静态类型检查，提高代码质量和可维护性 |
| **构建工具** | Vite | - | 现代前端构建工具，提供快速的开发体验 |
| **状态管理** | Pinia | - | Vue 3官方推荐的状态管理库 |
| **路由** | Vue Router | - | Vue官方路由库，实现单页应用导航 |
| **HTTP客户端** | Axios | - | 基于Promise的HTTP客户端，用于API调用 |

## 系统设计

### 核心功能模块

IXPay Pro 系统由多个功能模块组成，每个模块负责特定的业务功能：

#### 基础管理模块 (`server/internal/app/base`)

| 功能模块         | 描述                                           | 实现状态 |
|------------------|------------------------------------------------|----------|
| 用户管理         | 注册、登录、信息管理、密码修改、用户设置等      | ✅ 已实现 |
| 角色管理         | 角色的增删改查、权限分配、角色继承等            | ✅ 已实现 |
| 权限管理         | 基于角色的访问控制，支持菜单、API路由和按钮级权限 | ✅ 已实现 |
| 菜单管理         | 菜单的增删改查、树形结构管理、动态菜单生成      | ✅ 已实现 |
| API路由管理      | API路由的定义、权限控制、路由分组              | ✅ 已实现 |
| 按钮权限管理     | 按钮级权限的定义、分配和管理                  | ✅ 已实现 |
| 配置管理         | 系统配置的增删改查、多环境配置支持            | ✅ 已实现 |
| 字典管理         | 字典表和字典项的管理、数据分类和标准化        | ✅ 已实现 |
| 种子数据管理     | 系统初始化数据的管理、快速部署和配置          | ✅ 已实现 |
| 操作日志         | 用户操作日志记录、查询和分析                  | ✅ 已实现 |
| 权限组管理       | 将相关权限分组管理、权限批量分配              | ✅ 已实现 |
| 权限规则管理     | 基于ABAC的权限规则管理、细粒度权限控制        | ✅ 已实现 |
| 组织架构管理     | 部门、岗位等组织架构的管理                    | ⏳ 待实现 |
| 双因素认证       | 短信验证码、TOTP等双因素认证                 | ⏳ 待实现 |

#### 微信支付模块 (`server/internal/app/wx`)

| 功能模块         | 描述                                           | 实现状态 |
|------------------|------------------------------------------------|----------|
| 支付管理         | 创建支付、查询支付、取消支付                  | ✅ 已实现 |
| 微信认证         | 微信登录、微信用户信息管理                    | ✅ 已实现 |
| 支付通知         | 处理微信支付通知、订单状态更新                | ✅ 已实现 |
| 交易管理         | 交易记录查询、统计和分析                      | ⏳ 待实现 |

### 系统架构特点

IXPay Pro 采用现代化的系统架构设计，具有以下特点：

- **模块化设计**：清晰的分层架构，包括API层、服务层、数据访问层，便于代码维护和扩展
- **RESTful API**：遵循RESTful设计规范，提供标准化的接口，方便前端和第三方系统集成
- **权限体系**：基于RBAC+ABAC混合模型的权限管理，支持细粒度的权限控制
- **缓存机制**：使用Redis缓存权限信息和热点数据，提高系统性能
- **中间件**：实现了认证、权限验证、操作日志、错误处理等中间件，增强系统功能
- **依赖注入**：使用Wire实现编译时依赖注入，提高代码可维护性和可测试性
- **统一错误处理**：实现了全局错误处理机制，确保系统稳定运行
- **完善的日志**：使用Zap实现高性能日志记录，便于系统监控和问题排查
- **安全防护**：内置输入验证、防SQL注入、防XSS攻击等安全措施，保障系统安全
- **性能优化**：使用数据库索引、连接池、缓存等技术优化系统性能
- **容器化部署**：支持Docker容器化部署，简化部署和运维

### 项目结构

IXPay Pro 采用标准的 Go 项目结构，遵循清晰的目录组织和职责划分：

```
├── cmd/                 # 命令行入口
│   └── ixpay-pro/       # 主应用程序入口
│       └── main.go      # 主程序入口文件
├── configs/             # 配置文件目录
│   └── config.yaml      # 应用程序主配置文件
├── docker-compose.yml   # Docker Compose配置文件
├── Dockerfile           # Docker构建文件
├── docs/                # API文档目录
│   ├── docs.go          # Swagger文档定义
│   ├── swagger.json     # 自动生成的Swagger JSON
│   └── swagger.yaml     # 自动生成的Swagger YAML
├── go.mod               # Go模块定义文件
├── go.sum               # 依赖版本锁定文件
├── internal/            # 内部包（不对外暴露）
│   ├── app/             # 应用程序核心模块
│   │   ├── application.go # 应用程序结构定义
│   │   ├── routes.go    # 全局路由定义
│   │   ├── wire.go      # 依赖注入配置
│   │   ├── wire_gen.go  # 自动生成的依赖注入代码
│   │   ├── base/        # 基础管理模块
│   │   │   ├── api/     # API层，处理HTTP请求
│   │   │   │   └── v1/  # API版本
│   │   │   ├── domain/  # 领域层，实现核心业务逻辑
│   │   │   │   ├── model/       # 数据模型
│   │   │   │   │   ├── request/   # 请求模型
│   │   │   │   │   └── response/  # 响应模型
│   │   │   │   ├── repository/  # 数据访问层
│   │   │   │   └── service/     # 服务层
│   │   │   ├── middleware/      # 中间件
│   │   │   ├── migrations/      # 数据库迁移
│   │   │   └── seed/            # 种子数据
│   │   └── wx/          # 微信支付模块
│   │       ├── api/     # API层
│   │       ├── domain/  # 领域层
│   │       ├── middleware/      # 中间件
│   │       └── migrations/      # 数据库迁移
│   ├── config/          # 配置管理
│   ├── infrastructure/  # 基础设施层
│   │   ├── api/         # API相关工具
│   │   ├── auth/        # 认证相关
│   │   ├── cache/       # 缓存相关
│   │   ├── captcha/     # 验证码相关
│   │   ├── database/    # 数据库相关
│   │   ├── error/       # 错误处理
│   │   ├── logger/      # 日志相关
│   │   ├── middleware/  # 全局中间件
│   │   ├── redis/       # Redis相关
│   │   ├── snowflake/   # 分布式ID生成
│   │   └── task/        # 任务管理
│   └── utils/           # 工具函数
│       ├── common/      # 通用工具
│       └── encryption/  # 加密相关
├── API文档.md           # API文档
├── README.md            # 项目说明文档
├── 部署说明.md           # 部署说明
└── 项目设计文档.md        # 项目设计文档
```

### 目录说明

- **cmd/**: 命令行入口，包含应用程序的主入口文件
- **configs/**: 配置文件目录，存放应用程序的配置文件
- **docs/**: API文档目录，存放Swagger生成的API文档
- **internal/**: 内部包，不对外暴露的代码
  - **app/**: 应用程序核心模块，包含基础管理和微信支付模块
  - **config/**: 配置管理，处理应用程序配置
  - **infrastructure/**: 基础设施层，提供各种基础服务
  - **utils/**: 工具函数，提供通用的工具方法
- **API文档.md**: 详细的API接口文档
- **README.md**: 项目说明文档
- **部署说明.md**: 部署指南
- **项目设计文档.md**: 项目设计文档

## 环境要求

### 基础环境

| 组件 | 版本要求 | 用途 |
| --- | --- | --- |
| Go | 1.20+ | 后端开发语言，推荐使用 1.24.6 版本 |
| Node.js | 16+ | 前端开发环境，推荐使用 18.x 版本 |
| npm | 8+ | 前端依赖管理，推荐使用 9.x 版本 |
| PostgreSQL | 13+ | 关系型数据库，推荐使用 14.x 版本 |
| Redis | 6+ | 缓存、会话管理，推荐使用 7.x 版本 |
| Docker | 20.10+ | 容器化部署（可选） |
| Docker Compose | 1.29+ | 容器编排（可选） |

### 系统资源

| 环境 | CPU | 内存 | 磁盘 | 网络 |
| --- | --- | --- | --- | --- |
| 开发环境 | 至少 2 核 | 至少 4GB | 至少 50GB 可用空间 | 宽带网络 |
| 测试环境 | 至少 4 核 | 至少 8GB | 至少 100GB 可用空间 | 稳定网络 |
| 生产环境 | 至少 8 核 | 至少 16GB | 至少 200GB 可用空间 | 高速网络 |

### 操作系统

- **开发环境**: Windows 10/11, macOS, Linux
- **测试环境**: Linux (Ubuntu 20.04+, CentOS 7+)
- **生产环境**: Linux (Ubuntu 20.04+, CentOS 7+)

## 快速开始

### 安装步骤

#### 后端部署

1. **克隆代码仓库**

```bash
git clone https://github.com/ix-pay/ixpay-pro.git
cd ixpay-pro/server
```

2. **安装依赖**

```bash
go mod download
go mod tidy
```

3. **配置数据库和Redis**

编辑 `configs/config.yaml` 文件，设置数据库连接等信息：

```yaml
# 服务器配置
server:
  port: 8586
  mode: "debug"  # debug, release, test

# 数据库配置
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  user: "ixpay"
  password: "ixpay123"
  dbname: "ixpay_pro"
  sslmode: "disable"

# Redis配置
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# JWT配置
jwt:
  secret: "your-secret-key"
  expire: 3600

# 日志配置
logging:
  level: "info"
  file: "logs/"
```

4. **运行数据库迁移**

```bash
go run cmd/ixpay-pro/main.go migrate
```

5. **运行种子数据**

```bash
go run cmd/ixpay-pro/main.go seed
```

6. **启动服务**

```bash
# 开发模式
go run cmd/ixpay-pro/main.go

# 生产模式
go build -o ixpay-server cmd/ixpay-pro/main.go
./ixpay-server
```

#### 前端部署

```bash
# 进入前端目录
cd ../web

# 安装依赖
npm install

# 开发环境启动
npm run dev

# 生产环境构建
npm run build

# 部署到服务器
# 将 dist 目录部署到 Nginx 或其他 Web 服务器
```

### 验证服务

服务启动后，可以通过以下方式验证：

1. **API文档**：访问 http://localhost:8586/swagger/index.html
2. **健康检查**：访问 http://localhost:8586/health
3. **前端应用**：访问 http://localhost:3000（开发模式）

## API文档

### 生成API文档

使用 Swagger 生成 API 文档：

```bash
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseDependency --parseInternal --parseDepth 1
```

### 访问API文档

系统集成了 Swagger/OpenAPI 文档，启动服务后可访问：

- **Swagger UI**: http://localhost:8586/swagger/index.html
- **API文档JSON**: http://localhost:8586/swagger/doc.json
- **API文档YAML**: http://localhost:8586/swagger/doc.yaml

### API接口分类

#### 认证与授权 API

| 接口名称 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 注册 | POST | /api/admin/auth/register | 用户注册 |
| 登录 | POST | /api/admin/auth/login | 用户登录 |
| 验证码 | POST | /api/admin/auth/captcha | 获取验证码 |
| 刷新令牌 | POST | /api/admin/auth/refresh-token | 刷新访问令牌 |
| 退出登录 | POST | /api/admin/auth/logout | 用户退出登录 |

#### 用户管理 API

| 接口名称 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 获取用户信息 | GET | /api/admin/user/info | 获取当前用户信息 |
| 更新用户信息 | PUT | /api/admin/user/info | 更新用户信息 |
| 获取用户列表 | GET | /api/admin/user | 获取用户列表 |
| 添加用户 | POST | /api/admin/user | 添加新用户 |
| 删除用户 | DELETE | /api/admin/user/:id | 删除用户 |
| 修改密码 | PUT | /api/admin/user/password | 修改用户密码 |
| 重置密码 | PUT | /api/admin/user/reset-password | 重置用户密码 |
| 获取用户设置 | GET | /api/admin/user/getSelfSetting | 获取用户设置 |
| 设置用户设置 | PUT | /api/admin/user/setSelfSetting | 设置用户设置 |

#### 角色管理 API

| 接口名称 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 创建角色 | POST | /api/admin/roles | 创建新角色 |
| 获取角色详情 | GET | /api/admin/roles/detail | 获取角色详情 |
| 更新角色 | PUT | /api/admin/roles | 更新角色信息 |
| 删除角色 | DELETE | /api/admin/roles | 删除角色 |
| 获取角色列表 | GET | /api/admin/roles | 获取角色列表 |
| 分配用户到角色 | POST | /api/admin/roles/assign-users | 分配用户到角色 |
| 分配菜单到角色 | POST | /api/admin/roles/assign-menus | 分配菜单到角色 |
| 分配API路由到角色 | POST | /api/admin/roles/assign-api-routes | 分配API路由到角色 |

#### 按钮权限管理 API

| 接口名称 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 创建按钮权限 | POST | /api/admin/btn-perms | 创建按钮权限 |
| 获取按钮权限详情 | GET | /api/admin/btn-perms/detail | 获取按钮权限详情 |
| 更新按钮权限 | PUT | /api/admin/btn-perms | 更新按钮权限 |
| 删除按钮权限 | DELETE | /api/admin/btn-perms | 删除按钮权限 |
| 获取按钮权限列表 | GET | /api/admin/btn-perms | 获取按钮权限列表 |
| 分配API路由到按钮权限 | POST | /api/admin/btn-perms/assign-api-routes | 分配API路由到按钮权限 |
| 从按钮权限中撤销API路由 | POST | /api/admin/btn-perms/revoke-api-route | 从按钮权限中撤销API路由 |
| 分配按钮权限到角色 | POST | /api/admin/btn-perms/assign-to-role | 分配按钮权限到角色 |
| 从角色中撤销按钮权限 | POST | /api/admin/btn-perms/revoke-from-role | 从角色中撤销按钮权限 |
| 获取按钮权限下的API路由 | GET | /api/admin/btn-perms/api-routes | 获取按钮权限下的API路由 |
| 获取API路由下的按钮权限 | GET | /api/admin/btn-perms/for-route | 获取API路由下的按钮权限 |
| 获取角色下的按钮权限 | GET | /api/admin/btn-perms/by-role | 获取角色下的按钮权限 |
| 获取菜单下的按钮权限 | GET | /api/admin/btn-perms/by-menu | 获取菜单下的按钮权限 |

#### 支付管理 API

| 接口名称 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 创建支付 | POST | /api//payment | 创建支付订单 |
| 查询支付 | GET | /api//payment/{id} | 查询支付详情 |
| 获取用户支付列表 | GET | /api//payment | 获取用户支付列表 |
| 取消支付 | PUT | /api//payment/{id}/cancel | 取消支付订单 |

### API文档说明

- **请求格式**: 所有API接口均支持 JSON 格式的请求体
- **响应格式**: 所有API接口均返回 JSON 格式的响应
- **认证方式**: 使用 JWT 令牌进行认证，在请求头中添加 `Authorization: Bearer <token>`
- **错误处理**: 统一的错误响应格式，包含错误码和错误信息
- **分页参数**: 列表接口支持 `page` 和 `page_size` 参数进行分页

### 详细API文档

更多详细的 API 接口文档，请参考项目根目录下的 `API文档.md` 文件。

## 配置说明

### 环境变量

IXPay Pro 支持通过环境变量来配置系统，环境变量会覆盖配置文件中的对应设置：

| 变量名 | 描述 | 默认值 |
| --- | --- | --- |
| LOG_LEVEL | 日志级别（debug/info/warn/error） | info |
| SERVER_PORT | 服务端口 | 8586 |
| SERVER_MODE | 服务器运行模式（debug/release/test） | debug |
| JWT_SECRET | JWT 密钥 | 随机生成 |
| JWT_EXPIRE | JWT 过期时间（秒） | 3600 |
| REDIS_HOST | Redis 主机 | localhost |
| REDIS_PORT | Redis 端口 | 6379 |
| REDIS_PASSWORD | Redis 密码 | "" |
| REDIS_DB | Redis 数据库编号 | 0 |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户 | ixpay |
| DB_PASSWORD | 数据库密码 | ixpay123 |
| DB_NAME | 数据库名称 | ixpay_pro |
| DB_SSLMODE | 数据库 SSL 模式 | disable |

### 配置文件

主要配置文件位于 `server/configs/config.yaml`，包含以下主要部分：

```yaml
# 服务器配置
server:
  port: 8586            # 服务端口
  mode: "debug"         # 运行模式：debug, release, test

# 数据库配置
database:
  type: "postgres"      # 数据库类型
  host: "localhost"     # 数据库主机
  port: 5432            # 数据库端口
  user: "ixpay"         # 数据库用户
  password: "ixpay123"   # 数据库密码
  dbname: "ixpay_pro"   # 数据库名称
  sslmode: "disable"    # SSL 模式

# Redis配置
redis:
  host: "localhost"     # Redis 主机
  port: 6379            # Redis 端口
  password: ""          # Redis 密码
  db: 0                 # Redis 数据库编号

# JWT配置
jwt:
  secret: "your-secret-key"  # JWT 密钥
  expire: 3600            # 过期时间（秒）

# 日志配置
logging:
  level: "info"          # 日志级别
  file: "logs/"          # 日志文件目录

# 上传配置
upload:
  path: "uploads/"       # 上传文件目录
  max_size: 10485760     # 最大文件大小（10MB）
```

### 多环境配置

系统支持多环境配置，通过不同的配置文件来适应不同的环境：

- **开发环境**: `configs/config.yaml`
- **测试环境**: `configs/config.test.yaml`
- **生产环境**: `configs/config.prod.yaml`

启动服务时，可以通过 `--config` 参数指定配置文件：

```bash
go run cmd/ixpay-pro/main.go --config=configs/config.prod.yaml
```

## Docker部署

### 使用Docker Compose

IXPay Pro 提供了完整的 Docker Compose 配置，可以一键启动所有服务：

```bash
cd server
docker-compose up -d
```

这将启动以下服务：
- **ixpay-server**: 后端服务，端口 8586
- **postgres**: PostgreSQL 数据库，端口 5432
- **redis**: Redis 缓存，端口 6379

### 构建Docker镜像

```bash
cd server
# 构建镜像
docker build -t ixpay-server .

# 运行容器
docker run -d --name ixpay-server \
  -p 8586:8586 \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=ixpay \
  -e DB_PASSWORD=ixpay123 \
  -e DB_NAME=ixpay_pro \
  -e REDIS_HOST=redis \
  -e REDIS_PORT=6379 \
  ixpay-server
```

### Docker配置说明

Docker Compose 配置文件 `docker-compose.yml` 包含以下服务：

- **ixpay-server**: 后端服务，使用本地代码构建
- **postgres**: PostgreSQL 数据库，使用官方镜像
- **redis**: Redis 缓存，使用官方镜像

### 环境变量配置

在 Docker 环境中，可以通过环境变量来配置系统，具体参考上面的环境变量说明。

## 启动和停止服务

### 后端服务

#### 开发模式

```bash
cd server
go run cmd/ixpay-pro/main.go
```

#### 生产模式

```bash
cd server
# 构建可执行文件
go build -o ixpay-server cmd/ixpay-pro/main.go

# 运行服务
./ixpay-server

# 后台运行
nohup ./ixpay-server > server.log 2>&1 &
```

#### 停止服务

```bash
# 查找进程ID
ps aux | grep ixpay-server

# 停止服务
kill -9 <进程ID>
```

### 前端服务

#### 开发模式

```bash
cd web
npm run dev
```

#### 生产模式

```bash
cd web
# 构建生产版本
npm run build

# 部署到 Nginx
# 将 dist 目录复制到 Nginx 的网站根目录
# 配置 Nginx 服务器
```

### 服务管理

#### 健康检查

```bash
# 检查服务是否正常运行
curl http://localhost:8586/health
```

#### 查看日志

```bash
# 查看后端日志
cd server
cat logs/info.log

# 查看前端日志
cd web
npm run dev
```

## 故障排查

### 常见问题及解决方案

#### 数据库连接失败
- **问题现象**: 服务启动时出现数据库连接错误
- **解决方案**:
  - 检查 PostgreSQL 服务是否已启动
  - 检查数据库连接参数是否正确（主机、端口、用户名、密码）
  - 检查数据库用户是否有正确的权限
  - 检查数据库是否已创建

#### Redis连接失败
- **问题现象**: 服务启动时出现Redis连接错误
- **解决方案**:
  - 检查 Redis 服务是否已启动
  - 检查Redis连接参数是否正确（主机、端口、密码）
  - 检查Redis是否设置了密码
  - 检查网络连接是否正常

#### JWT认证失败
- **问题现象**: API请求返回401未授权错误
- **解决方案**:
  - 检查JWT密钥是否一致
  - 检查Token是否过期
  - 检查Token格式是否正确
  - 检查请求头中的Authorization字段是否正确

#### 跨域问题
- **问题现象**: 前端请求后端API时出现跨域错误
- **解决方案**:
  - 检查前端配置的API地址是否正确
  - 检查后端是否已配置CORS中间件
  - 检查浏览器控制台的错误信息

#### 服务启动失败
- **问题现象**: 服务无法启动或启动后立即退出
- **解决方案**:
  - 检查端口是否被占用
  - 检查配置文件是否正确
  - 查看日志文件中的错误信息

### 日志查看

后端日志默认保存在 `server/logs/` 目录下，按级别分类：

- `error.log`: 错误日志，记录系统错误和异常
- `warn.log`: 警告日志，记录系统警告信息
- `info.log`: 信息日志，记录系统运行状态
- `debug.log`: 调试日志，记录详细的调试信息

### 调试方法

1. **开启调试模式**:
   - 修改配置文件中的 `server.mode` 为 `debug`
   - 设置环境变量 `SERVER_MODE=debug`

2. **查看详细日志**:
   - 修改配置文件中的 `logging.level` 为 `debug`
   - 设置环境变量 `LOG_LEVEL=debug`

3. **使用curl测试API**:
   ```bash
   # 测试健康检查接口
   curl http://localhost:8586/health
   
   # 测试登录接口
   curl -X POST http://localhost:8586/api/admin/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "password123"}'
   ```

4. **检查数据库状态**:
   ```bash
   # 连接PostgreSQL
   psql -h localhost -U ixpay -d ixpay_pro
   
   # 查看表结构
   \dt
   
   # 查看数据
   SELECT * FROM users;
   ```

## 安全建议

### 生产环境配置

1. **基础安全配置**
   - 修改默认密码，使用强密码策略
   - 启用HTTPS，配置SSL证书
   - 配置防火墙规则，限制访问端口
   - 限制API访问IP，使用白名单

2. **服务器安全**
   - 定期更新操作系统和软件
   - 关闭不必要的服务和端口
   - 配置安全的SSH访问
   - 使用密钥认证，禁用密码登录

### 数据库安全

1. **数据库配置**
   - 使用强密码，定期更换
   - 限制数据库访问IP
   - 最小权限原则，为数据库用户分配适当的权限
   - 启用数据库审计日志

2. **数据保护**
   - 定期备份数据库，制定恢复计划
   - 对敏感数据进行加密存储
   - 定期清理过期数据
   - 实施数据访问控制

### 应用安全

1. **代码安全**
   - 定期更新依赖包，修复安全漏洞
   - 启用输入验证，防止恶意输入
   - 防止SQL注入，使用参数化查询
   - 防止XSS攻击，对用户输入进行过滤

2. **认证与授权**
   - 使用安全的密码哈希算法
   - 实施多因素认证
   - 定期轮换JWT密钥
   - 监控异常登录行为

3. **API安全**
   - 实施API速率限制，防止暴力攻击
   - 使用HTTPS保护API通信
   - 验证所有API请求的身份和权限
   - 记录API访问日志

### 监控与审计

1. **安全监控**
   - 部署入侵检测系统
   - 监控异常访问行为
   - 定期扫描系统漏洞
   - 建立安全事件响应机制

2. **日志审计**
   - 记录所有关键操作日志
   - 定期分析日志，发现异常行为
   - 保存日志备份，便于追溯
   - 实施日志访问控制

### 安全最佳实践

- **定期安全评估**：定期进行安全测试和评估
- **安全培训**：对开发和运维人员进行安全培训
- **安全编码规范**：制定并遵循安全编码规范
- **漏洞管理**：建立漏洞发现和修复流程
- **安全文档**：维护安全配置和应急响应文档

## 更新说明

### 后端更新

1. **获取最新代码**
   ```bash
   cd server
   git pull
   ```

2. **更新依赖**
   ```bash
   go mod tidy
   ```

3. **数据库迁移（如果有）**
   ```bash
   go run cmd/ixpay-pro/main.go migrate
   ```

4. **重启服务**
   ```bash
   # 停止旧服务
   kill -9 <进程ID>
   
   # 启动新服务
   go run cmd/ixpay-pro/main.go
   # 或使用生产模式
   go build -o ixpay-server cmd/ixpay-pro/main.go
   ./ixpay-server
   ```

### 前端更新

1. **获取最新代码**
   ```bash
   cd web
   git pull
   ```

2. **更新依赖**
   ```bash
   npm install
   ```

3. **构建生产版本**
   ```bash
   npm run build
   ```

4. **重新部署**
   - 将 `dist` 目录复制到 Web 服务器
   - 重启 Web 服务器

### 更新注意事项

1. **备份数据**：在更新前，建议备份数据库和配置文件
2. **检查兼容性**：查看更新日志，了解是否有破坏性变更
3. **测试环境**：在测试环境中先进行更新测试
4. **回滚计划**：制定更新失败的回滚计划
5. **监控**：更新后密切监控系统运行状态

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！
