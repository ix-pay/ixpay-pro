<p align="center">
  <img src="./docs/images/ixpay.png" width="300" height="300" alt="IXPay Pro Logo" />
</p>

[English](./README-en.md) | 简体中文

# IXPay Pro

一个基于Go语言和Gin框架的高性能支付API服务，专注于提供微信支付解决方案，集成用户认证、支付处理和任务管理功能。

## 项目结构

```
ixpay-pro/
├── common/             # 公共模块（文档、规范等）
├── gxy/                # 网关服务（Go 语言）
├── h5app/              # H5 应用前端（开发中）
├── server/             # 后端服务（Go 语言）
├── weapp/              # 微信小程序前端（开发中）
└── web/                # Vue3 + TypeScript 前端
```

## 技术栈

### 后端 (server)

- **开发语言**: Go 1.24.6
- **Web框架**: Gin v1.10.1
- **依赖注入**: Wire v0.7.0
- **数据库**: PostgreSQL + GORM v1.30.3
- **缓存**: Redis v9.13.0
- **认证**: JWT v5.3.0
- **配置管理**: Viper v1.20.1
- **日志**: Zap v1.27.0

### 前端 (web)

- **框架**: Vue 3
- **语言**: TypeScript
- **UI组件库**: Element Plus v2.11.2
- **状态管理**: Pinia v3.0.3
- **路由**: Vue Router v4.5.1
- **构建工具**: Vite v7.0.6

## 核心功能

### 用户认证

- 注册、登录、微信登录
- 令牌刷新和权限管理

### 支付处理

- 创建支付、查询支付
- 取消支付和处理微信支付通知

### 任务管理

- 添加、移除、启动、停止和重试任务

## 快速开始

### 后端部署

#### Docker部署（推荐）

```bash
cd server
# 创建.env文件并配置环境变量
cp .env.example .env
# 启动服务
docker-compose up -d
```

#### 本地运行

```bash
# 进入后端服务目录
cd server

# 安装依赖
go mod download

# 生成依赖注入代码
wire ./internal/app

# 生成 API 文档（在 server 目录下执行）
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseInternal --parseDependency

# 运行应用
GO_ENV=development go run cmd/ixpay-pro/main.go

# 访问 API 文档
http://127.0.0.1:8586/swagger/index.html

# 构建可执行文件
go build -o ./build/ixpay-pro.exe cmd/ixpay-pro/main.go
```

### 前端运行

```bash
cd web
# 安装依赖
npm install
# 开发环境运行
npm run serve
# 生产环境构建
npm run build
```

## API文档

应用启动后，可以通过以下URL访问Swagger API文档：

```
http://127.0.0.1:8586/swagger/index.html
```

## 贡献指南

1. Fork 项目仓库
2. 创建功能分支: `git checkout -b feature/AmazingFeature`
3. 提交更改: `git commit -m 'Add some AmazingFeature'`
4. 推送到分支: `git push origin feature/AmazingFeature`
5. 开启 Pull Request

## 许可证

IXPay Pro 项目在 Apache License 2.0 下发布。
