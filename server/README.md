# IXPay Pro - 支付API服务

一个基于Go语言和Gin框架的支付API服务，专注于提供微信支付解决方案，集成用户认证、支付处理和任务管理功能。

## 功能特性

- **用户认证**：支持注册、登录、微信登录和令牌刷新
- **支付处理**：支持创建支付、查询支付、取消支付和处理微信支付通知
- **任务管理**：支持添加、移除、启动、停止和重试任务
- **权限管理**：基于角色的访问控制
- **文档化**：集成Swagger API文档
- **优雅关闭**：支持信号处理和优雅关闭
- **分布式ID**：集成Snowflake算法生成唯一ID
- **验证码服务**：支持生成和验证验证码
- **跨域支持**：内置CORS中间件

## 技术栈

- **Go**：1.24.6
- **Web框架**：Gin (v1.10.1)
- **依赖注入**：Wire (v0.7.0)
- **数据库**：PostgreSQL (GORM v1.30.3)
- **缓存**：Redis (v9.13.0)
- **认证**：JWT (v5.3.0)
- **配置管理**：Viper (v1.20.1)
- **日志**：Zap (v1.27.0)
- **任务调度**：Cron (v3.0.1)
- **API文档**：Swagger

## 项目结构

```
├── configs/             # 配置文件
│   └── config.yaml      # 应用程序配置
├── docker-compose.yml   # Docker Compose配置
├── Dockerfile           # Docker构建文件
├── docs/                # API文档
│   ├── docs.go          # Swagger文档定义
│   ├── swagger.json     # 自动生成的Swagger JSON
│   └── swagger.yaml     # 自动生成的Swagger YAML
├── go.mod               # Go模块定义
├── go.sum               # 依赖版本锁定
├── internal/            # 内部包
│   ├── app/             # 应用程序核心
│   │   ├── application.go # 应用程序结构
│   │   ├── base/        # 基础功能模块
│   │   │   ├── api/     # API控制器
│   │   │   ├── application.go # 模块应用程序结构
│   │   │   ├── domain/  # 领域模型和服务
│   │   │   ├── routes.go    # 模块路由定义
│   │   │   └── wire.go      # 模块依赖注入配置
│   │   ├── routes.go    # 全局路由定义
│   │   ├── wire.go      # 全局依赖注入配置
│   │   ├── wire_gen.go  # 自动生成的依赖注入代码
│   │   └── wx/          # 微信支付模块
│   │       ├── api/     # 微信API控制器
│   │       ├── application.go # 微信模块应用程序结构
│   │       ├── domain/  # 微信领域模型和服务
│   │       ├── routes.go    # 微信模块路由定义
│   │       └── wire.go      # 微信模块依赖注入配置
│   ├── config/          # 配置相关
│   │   └── config.go    # 配置结构体定义
│   ├── infrastructure/  # 基础设施层
│   │   ├── auth/        # 认证相关
│   │   │   ├── jwt.go   # JWT认证
│   │   │   └── permission.go # 权限管理
│   │   ├── database/    # 数据库连接
│   │   │   ├── migrations.go # 数据库迁移
│   │   │   └── postgresql.go # PostgreSQL连接
│   │   ├── logger/      # 日志配置
│   │   │   └── logger.go # 日志实现
│   │   ├── middleware/  # 中间件
│   │   │   ├── auth_middleware.go # 认证中间件
│   │   │   ├── cors_middleware.go # CORS中间件
│   │   │   ├── logger_middleware.go # 日志中间件
│   │   │   └── permission_middleware.go # 权限中间件
│   │   ├── redis/       # Redis连接
│   │   │   └── redis.go # Redis实现
│   │   ├── snowflake/   # 分布式ID生成
│   │   │   └── snowflake.go # Snowflake算法实现
│   │   └── task/        # 任务管理
│   │       └── task_manager.go # 任务管理器实现
│   ├── main.go          # 应用程序入口
│   └── utils/           # 工具函数
│       ├── captcha/     # 验证码功能
│       │   └── captcha.go # 验证码实现
│       └── common/      # 通用工具
│           ├── baseReq/ # 基础请求结构
│           └── baseRes/ # 基础响应结构
├── logs/                # 日志文件
│   └── app.log          # 应用程序日志
└── README.md            # 项目说明文档
```

## 快速开始

### 前提条件

- Go 1.24.6 或更高版本 或 Docker
- PostgreSQL 数据库
- Redis 服务器
- 微信支付商户号和API密钥

### 安装依赖

```bash
go mod download
```

### 生成依赖注入代码

```bash
go run github.com/google/wire/cmd/wire ./internal/app
```

## API文档

应用启动后，可以通过以下URL访问Swagger API文档：
```bash
swag init -g internal/main.go --output docs --parseDependency --parseInternal
```
```bash
go run internal/main.go
```
```
http://127.0.0.1:8586/swagger/index.html
```

### 使用Docker部署

#### 1. 创建环境变量文件

在项目根目录创建 `.env` 文件，并添加以下配置：

```env
# 数据库配置
DB_USER=postgres
DB_PASSWORD=your-secure-password
DB_NAME=wire_test

# JWT配置
JWT_SECRET=your-secret-key

# 微信支付配置
WECHAT_APPID=your-wechat-appid
WECHAT_MCH_ID=your-wechat-mch-id
WECHAT_API_KEY=your-wechat-api-key
```

#### 2. 使用Docker Compose启动服务

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看应用日志
docker-compose logs -f app
```

#### 3. 服务访问

- API服务: http://127.0.0.1:8586
- Swagger文档: http://127.0.0.1:8586/swagger/index.html

#### 4. Docker相关命令

```bash
# 停止所有服务
docker-compose down

# 停止并删除所有服务和卷
docker-compose down -v

# 仅重建应用服务
docker-compose up -d --build app
```

### 本地开发运行

如果你不使用Docker，可以按照以下步骤在本地运行：

#### 1. 配置环境

直接修改配置文件：

```bash
vim configs/config.yaml  # 或使用您喜欢的编辑器
```

根据您的环境修改`configs/config.yaml`文件中的数据库连接、Redis连接、JWT密钥和微信支付配置。

#### 2. 运行应用

```bash
# 开发模式
ENV=development go run internal/main.go

# 生产模式
ENV=production go run internal/main.go
```

#### 3. 构建应用

```bash
go build -o ixpay-pro ./internal/main.go
```

## 主要API端点

### 公共路由

- **健康检查**: `GET /health`

### 认证路由

- **用户注册**: `POST /api/v1/auth/register`
- **用户登录**: `POST /api/v1/auth/login`
- **微信登录**: `POST /api/v1/auth/wechat-login`
- **刷新令牌**: `POST /api/v1/auth/refresh-token`

### 支付路由（不需要认证）

- **微信支付通知**: `POST /api/v1/pay/notify/wechat`

### 需要认证的路由

#### 用户路由
- **获取用户信息**: `GET /api/v1/user/info`

#### 支付路由
- **创建支付**: `POST /api/v1/payment`
- **查询支付**: `GET /api/v1/payment/:id`
- **获取用户支付列表**: `GET /api/v1/payment`
- **取消支付**: `PUT /api/v1/payment/:id/cancel`

#### 任务路由
- **添加任务**: `POST /api/v1/task`
- **移除任务**: `DELETE /api/v1/task/:id`
- **启动任务**: `POST /api/v1/task/:id/start`
- **停止任务**: `POST /api/v1/task/:id/stop`
- **重试任务**: `POST /api/v1/task/:id/retry`
- **获取任务列表**: `GET /api/v1/task`
- **获取任务详情**: `GET /api/v1/task/:id`

## 环境变量

- `ENV`: 设置运行环境，可选值为 `development` 和 `production`，默认为 `development`

## 日志

应用程序日志存储在 `logs/app.log` 文件中，使用Zap日志库进行日志管理。

## 贡献指南

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 Apache 2.0 许可证。如需完整的许可证文本，请访问 [Apache 2.0 许可证官方网站](https://www.apache.org/licenses/LICENSE-2.0)。

您可以在项目根目录创建一个 `LICENSE` 文件，复制 [Apache 2.0 许可证文本](https://www.apache.org/licenses/LICENSE-2.0) 来完整声明许可证。