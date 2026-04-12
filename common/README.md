<p align="center">
  <img src="./docs/images/ixpay.png" width="300" height="300" alt="IXPay Pro Logo" />
</p>

[English](./README-en.md) | 简体中文

# IXPay Pro

IXPay Pro 是一个基于 Go 语言和 Gin 框架的高性能支付管理系统，专注于提供微信支付解决方案。系统采用前后端分离架构和 DDD（领域驱动设计）分层架构，集成完整的后台管理系统、支付处理和任务管理功能。

## 项目结构

```
ixpay-pro/
├── common/             # 公共模块（文档、规范、技能等）
├── gxy/                # 网关服务（Go 语言，纯标准库）
├── h5app/              # H5 应用前端（初始化状态）
├── miniapp/            # 小程序前端（初始化状态）
├── server/             # 后端服务（Go 语言，DDD 架构）
├── weapp/              # 企业微信 H5 应用（初始化状态）
└── web/                # Vue3 + TypeScript 前端管理后台
```

## 技术栈

### 后端 (server)

| 类别 | 技术/框架 | 版本 | 说明 |
|------|-----------|------|------|
| **开发语言** | Go | 1.24.6 | 核心开发语言，提供高性能和并发处理能力 |
| **Web 框架** | Gin | v1.10.1 | 轻量级 HTTP 服务框架，提供路由、中间件等功能 |
| **依赖注入** | Wire | v0.7.0 | 编译时依赖注入工具，提高代码可维护性 |
| **数据库** | PostgreSQL | 13+ | 强大的开源关系型数据库，支持复杂查询和事务 |
| | GORM | v1.30.3 | 功能丰富的 ORM 库，简化数据库操作 |
| **缓存** | Redis | 6+ | 高性能键值存储，用于缓存和会话管理 |
| **认证** | JWT | v5.3.0 | 无状态身份认证令牌，支持跨服务认证 |
| **配置管理** | Viper | v1.20.1 | 灵活的配置文件管理工具，支持多种配置格式 |
| **日志** | Zap | v1.27.0 | 高性能结构化日志库，支持多级别日志 |
| **任务调度** | Cron | v3.0.1 | 定时任务调度库，用于执行周期性任务 |
| **API 文档** | Swagger | - | 自动 API 文档生成工具，方便接口调试和对接 |
| **监控** | Prometheus | - | 开源监控系统，用于系统性能监控 |
| **限流** | golang.org/x/time/rate | - | API 速率限制库，防止系统过载 |
| **雪花算法** | Snowflake | - | 分布式 ID 生成算法，确保数据唯一性 |
| **验证码** | captcha | - | 验证码生成和验证库，提高系统安全性 |

### 前端 (web)

| 类别 | 技术/框架 | 版本 | 说明 |
|------|-----------|------|------|
| **开发框架** | Vue 3 | - | 现代化前端框架，提供响应式数据绑定和组件化开发 |
| **UI 组件库** | Element Plus | v2.11.2 | 基于 Vue 3 的 UI 组件库，提供丰富的界面元素 |
| **开发语言** | TypeScript | - | 静态类型检查，提高代码质量和可维护性 |
| **构建工具** | Vite | v7.0.6 | 现代前端构建工具，提供快速的开发体验 |
| **状态管理** | Pinia | v3.0.3 | Vue 3 官方推荐的状态管理库 |
| **路由** | Vue Router | v4.5.1 | Vue 官方路由库，实现单页应用导航 |
| **HTTP 客户端** | Axios | - | 基于 Promise 的 HTTP 客户端，用于 API 调用 |

### 网关 (gxy)

- **开发语言**: Go（纯标准库）
- **核心功能**: 服务注册发现、负载均衡（轮询算法）、健康检查、集群数据同步、请求代理转发
- **技术特点**: 高性能、轻量级、线程安全

## 核心功能

### 后台管理功能

- **🔐 用户认证**: 支持注册、登录、微信登录和令牌刷新，使用 JWT 进行身份验证
- **👮 权限管理**: 基于 RBAC+ABAC 混合模型的权限管理，支持菜单、API 路由和按钮级权限控制
- **👥 角色管理**: 角色的增删改查、权限分配、角色继承和权限组管理
- **📋 菜单管理**: 菜单的增删改查、树形结构管理，支持动态菜单生成
- **⚙️ 配置管理**: 系统配置的增删改查，支持多环境配置
- **📚 字典管理**: 字典表和字典项的管理，支持数据分类和标准化
- **📝 操作日志**: 记录用户操作日志，支持日志查询和分析
- **🌱 种子数据管理**: 系统初始化数据的管理，确保系统快速部署和配置

### 支付功能

- **💳 支付处理**: 支持创建支付、查询支付、取消支付和处理微信支付通知
- **📱 微信支付**: 集成微信支付 API，支持扫码支付、H5 支付等多种支付方式
- **💰 交易管理**: 支付交易的查询、统计和分析

### 系统功能

- **📄 文档化**: 集成 Swagger API 文档，方便接口调试和对接
- **🛑 优雅关闭**: 支持信号处理和优雅关闭，确保服务稳定退出
- **🆔 分布式 ID**: 集成 Snowflake 算法生成唯一 ID，确保数据一致性
- **🔑 验证码服务**: 支持生成和验证验证码，提高系统安全性
- **🌐 跨域支持**: 内置 CORS 中间件，解决前后端分离架构下的跨域问题
- **📊 监控系统**: 支持 Prometheus 监控和 Zap 日志记录，确保系统稳定运行
- **🔒 安全防护**: 内置输入验证、防 SQL 注入、防 XSS 攻击等安全措施
- **⚡ 性能优化**: 使用 Redis 缓存、数据库索引等技术优化系统性能
- **📦 容器化部署**: 支持 Docker 容器化部署，简化部署和运维

### 网关功能

- **服务注册与发现**: 自动注册和管理后端服务实例
- **负载均衡**: 基于轮询算法的请求分发
- **健康检查**: 实时监测后端服务健康状态
- **集群同步**: 支持多网关节点数据同步
- **请求代理**: 高效的 HTTP 请求转发

## 快速开始

### 环境要求

| 组件 | 版本要求 | 用途 |
|------|---------|------|
| Go | 1.20+ | 后端开发语言，推荐使用 1.24.6 版本 |
| Node.js | 16+ | 前端开发环境，推荐使用 18.x 版本 |
| npm | 8+ | 前端依赖管理，推荐使用 9.x 版本 |
| PostgreSQL | 13+ | 关系型数据库，推荐使用 14.x 版本 |
| Redis | 6+ | 缓存、会话管理，推荐使用 7.x 版本 |
| Docker | 20.10+ | 容器化部署（可选） |
| Docker Compose | 1.29+ | 容器编排（可选） |

### 后端部署

#### Docker 部署（推荐）

```bash
cd server
# 创建.env 文件并配置环境变量
cp .env.example .env
# 启动服务
docker-compose up -d
```

这将启动以下服务：
- **ixpay-server**: 后端服务，端口 8586
- **postgres**: PostgreSQL 数据库，端口 5432
- **redis**: Redis 缓存，端口 6379

#### 本地运行

```bash
# 进入后端服务目录
cd server

# 安装依赖
go mod download
go mod tidy

# 配置数据库和 Redis
# 编辑 configs/config.yaml 文件

# 生成依赖注入代码
wire ./internal/app

# 生成 API 文档（在 server 目录下执行）
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseInternal --parseDependency

# 运行数据库迁移
go run cmd/ixpay-pro/main.go migrate

# 运行种子数据
go run cmd/ixpay-pro/main.go seed

# 运行应用
# 开发模式
go run cmd/ixpay-pro/main.go

# 生产模式
go build -o ixpay-server cmd/ixpay-pro/main.go
./ixpay-server

# 访问 API 文档
http://127.0.0.1:8586/swagger/index.html
```

### 前端运行

```bash
# 进入前端目录
cd web

# 安装依赖
npm install

# 开发环境运行
npm run serve

# 生产环境构建
npm run build
```

### 网关运行

```bash
# 进入网关目录
cd gxy

# 安装依赖
go mod download

# 运行网关
go run cmd/gateway/main.go

# 构建可执行文件
go build -o gateway cmd/gateway/main.go
```

## API 文档

### 生成 API 文档

使用 Swagger 生成 API 文档：

```bash
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseDependency --parseInternal --parseDepth 1
```

### 访问 API 文档

系统集成了 Swagger/OpenAPI 文档，启动服务后可访问：

- **Swagger UI**: http://localhost:8586/swagger/index.html
- **API 文档 JSON**: http://localhost:8586/swagger/doc.json
- **API 文档 YAML**: http://localhost:8586/swagger/doc.yaml

### API 接口分类

#### 认证与授权 API

| 接口名称 | 方法 | 路径 | 描述 |
|---------|------|------|------|
| 注册 | POST | /api/admin/auth/register | 用户注册 |
| 登录 | POST | /api/admin/auth/login | 用户登录 |
| 验证码 | POST | /api/admin/auth/captcha | 获取验证码 |
| 刷新令牌 | POST | /api/admin/auth/refresh-token | 刷新访问令牌 |
| 退出登录 | POST | /api/admin/auth/logout | 用户退出登录 |

#### 用户管理 API

| 接口名称 | 方法 | 路径 | 描述 |
|---------|------|------|------|
| 获取用户信息 | GET | /api/admin/user/info | 获取当前用户信息 |
| 更新用户信息 | PUT | /api/admin/user/info | 更新用户信息 |
| 获取用户列表 | GET | /api/admin/user | 获取用户列表 |
| 添加用户 | POST | /api/admin/user | 添加新用户 |
| 删除用户 | DELETE | /api/admin/user/:id | 删除用户 |
| 修改密码 | PUT | /api/admin/user/password | 修改用户密码 |
| 重置密码 | PUT | /api/admin/user/reset-password | 重置用户密码 |

#### 角色管理 API

| 接口名称 | 方法 | 路径 | 描述 |
|---------|------|------|------|
| 创建角色 | POST | /api/admin/roles | 创建新角色 |
| 获取角色详情 | GET | /api/admin/roles/detail | 获取角色详情 |
| 更新角色 | PUT | /api/admin/roles | 更新角色信息 |
| 删除角色 | DELETE | /api/admin/roles | 删除角色 |
| 获取角色列表 | GET | /api/admin/roles | 获取角色列表 |
| 分配用户到角色 | POST | /api/admin/roles/assign-users | 分配用户到角色 |
| 分配菜单到角色 | POST | /api/admin/roles/assign-menus | 分配菜单到角色 |

#### 支付管理 API

| 接口名称 | 方法 | 路径 | 描述 |
|---------|------|------|------|
| 创建支付 | POST | /api/payment | 创建支付订单 |
| 查询支付 | GET | /api/payment/{id} | 查询支付详情 |
| 获取用户支付列表 | GET | /api/payment | 获取用户支付列表 |
| 取消支付 | PUT | /api/payment/{id}/cancel | 取消支付订单 |

### API 文档说明

- **请求格式**: 所有 API 接口均支持 JSON 格式的请求体
- **响应格式**: 所有 API 接口均返回 JSON 格式的响应
- **认证方式**: 使用 JWT 令牌进行认证，在请求头中添加 `Authorization: Bearer <token>`
- **错误处理**: 统一的错误响应格式，包含错误码和错误信息
- **分页参数**: 列表接口支持 `page` 和 `page_size` 参数进行分页

## 系统架构

### 架构层次

1. **前端层**: 基于 Vue3 + Element Plus 构建的现代化用户界面
2. **API 层**: 基于 Gin 框架的 RESTful API 接口
3. **服务层**: 实现核心业务逻辑的服务组件
4. **数据访问层**: 与数据库交互的数据仓库
5. **基础设施层**: 提供认证、缓存、日志等基础服务

### 模块划分

- **基础管理模块** (`server/internal/app/base`): 用户、角色、权限、菜单等核心管理功能
- **微信支付模块** (`server/internal/app/wx`): 微信支付相关的功能实现
- **基础设施模块** (`server/internal/infrastructure`): 认证、缓存、日志、数据库等基础服务
- **网关模块** (`gxy`): 服务注册发现、负载均衡、健康检查

### 技术特性

- **模块化设计**: 清晰的分层架构，便于扩展和维护
- **RESTful API**: 遵循 RESTful 设计规范，提供标准化的接口
- **权限体系**: 基于 RBAC+ABAC 混合模型的权限管理
- **缓存机制**: 使用 Redis 缓存权限信息和热点数据
- **中间件**: 实现了认证、权限验证、操作日志等中间件
- **依赖注入**: 使用 Wire 实现编译时依赖注入，提高代码可维护性
- **统一错误处理**: 实现了全局错误处理机制
- **完善的日志**: 使用 Zap 实现高性能日志记录

## 配置说明

### 环境变量

IXPay Pro 支持通过环境变量来配置系统，环境变量会覆盖配置文件中的对应设置：

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
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

# Redis 配置
redis:
  host: "localhost"     # Redis 主机
  port: 6379            # Redis 端口
  password: ""          # Redis 密码
  db: 0                 # Redis 数据库编号

# JWT 配置
jwt:
  secret: "your-secret-key"  # JWT 密钥
  expire: 3600            # 过期时间（秒）

# 日志配置
logging:
  level: "info"          # 日志级别
  file: "logs/"          # 日志文件目录
```

## Docker 部署

### 使用 Docker Compose

IXPay Pro 提供了完整的 Docker Compose 配置，可以一键启动所有服务：

```bash
cd server
docker-compose up -d
```

这将启动以下服务：
- **ixpay-server**: 后端服务，端口 8586
- **postgres**: PostgreSQL 数据库，端口 5432
- **redis**: Redis 缓存，端口 6379

### 构建 Docker 镜像

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

## 故障排查

### 常见问题及解决方案

#### 数据库连接失败

- **问题现象**: 服务启动时出现数据库连接错误
- **解决方案**:
  - 检查 PostgreSQL 服务是否已启动
  - 检查数据库连接参数是否正确（主机、端口、用户名、密码）
  - 检查数据库用户是否有正确的权限
  - 检查数据库是否已创建

#### Redis 连接失败

- **问题现象**: 服务启动时出现 Redis 连接错误
- **解决方案**:
  - 检查 Redis 服务是否已启动
  - 检查 Redis 连接参数是否正确（主机、端口、密码）
  - 检查 Redis 是否设置了密码
  - 检查网络连接是否正常

#### JWT 认证失败

- **问题现象**: API 请求返回 401 未授权错误
- **解决方案**:
  - 检查 JWT 密钥是否一致
  - 检查 Token 是否过期
  - 检查 Token 格式是否正确
  - 检查请求头中的 Authorization 字段是否正确

#### 跨域问题

- **问题现象**: 前端请求后端 API 时出现跨域错误
- **解决方案**:
  - 检查前端配置的 API 地址是否正确
  - 检查后端是否已配置 CORS 中间件
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

3. **使用 curl 测试 API**:
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
   # 连接 PostgreSQL
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
   - 启用 HTTPS，配置 SSL 证书
   - 配置防火墙规则，限制访问端口
   - 限制 API 访问 IP，使用白名单

2. **服务器安全**
   - 定期更新操作系统和软件
   - 关闭不必要的服务和端口
   - 配置安全的 SSH 访问
   - 使用密钥认证，禁用密码登录

### 数据库安全

1. **数据库配置**
   - 使用强密码，定期更换
   - 限制数据库访问 IP
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
   - 防止 SQL 注入，使用参数化查询
   - 防止 XSS 攻击，对用户输入进行过滤

2. **认证与授权**
   - 使用安全的密码哈希算法
   - 实施多因素认证
   - 定期轮换 JWT 密钥
   - 监控异常登录行为

3. **API 安全**
   - 实施 API 速率限制，防止暴力攻击
   - 使用 HTTPS 保护 API 通信
   - 验证所有 API 请求的身份和权限
   - 记录 API 访问日志

## 贡献指南

我们欢迎各种形式的贡献！

1. **Fork 项目仓库**
2. **创建功能分支**: `git checkout -b feature/AmazingFeature`
3. **提交更改**: `git commit -m 'feat: add some AmazingFeature'`
4. **推送到分支**: `git push origin feature/AmazingFeature`
5. **开启 Pull Request**

### 代码风格

- **后端**: 遵循 [Go 代码风格与开发规范](.trae/rules/Go 代码风格与开发规范.md)
- **前端**: 遵循 [Vue 代码风格规范](.trae/rules/Vue 代码风格规范.md)
- **提交信息**: 遵循 Conventional Commits 规范

### 开发流程

1. **克隆项目**: `git clone https://github.com/ix-pay/ixpay-pro.git`
2. **安装依赖**: 根据各子项目的 README 安装依赖
3. **配置环境**: 配置数据库、Redis 等环境
4. **开发功能**: 在对应模块下开发新功能
5. **编写测试**: 为新增功能编写单元测试
6. **提交代码**: 确保所有测试通过后提交代码

## 许可证

IXPay Pro 项目在 Apache License 2.0 下发布。

## 联系方式

- **项目主页**: https://github.com/ix-pay/ixpay-pro
- **问题反馈**: https://github.com/ix-pay/ixpay-pro/issues
- **邮箱**: support@ixpay.pro

---

<p align="center">Made with ❤️ by IXPay Pro Team</p>
