---
name: 后端开发工具集
description: 当用户进行 Go 后端开发时自动触发，执行项目构建、运行、依赖管理、代码格式化、测试执行、Swagger 文档生成和 Docker 部署等命令，提升开发效率
---

## 核心原则

1. **命令驱动**：专注执行后端开发相关的命令
2. **自动化**：优先使用自动化工具完成重复性任务
3. **质量保障**：所有代码必须通过格式化、类型检查和测试验证
4. **规范检查**：代码规范检查由全栈开发技能负责

## 核心功能概览

1. **代码格式化** - 自动执行 go fmt 和 go imports
2. **项目构建** - 编译和构建可执行文件
3. **依赖管理** - 下载、整理和添加依赖
4. **测试执行** - 运行单元测试和覆盖率检查
5. **API 文档生成** - 自动生成 Swagger 文档
6. **Docker 部署** - 构建镜像和运行容器
7. **Wire 代码生成** - 生成依赖注入代码

## 工作流程

### 场景 1：代码格式化

```
用户请求 → 执行 go fmt → 执行 go imports → 完成
```

### 场景 2：项目构建与运行

```
用户请求 → 执行 go build → 输出构建结果 → 完成
用户请求 → 执行 go run → 启动服务 → 完成
```

### 场景 3：测试执行

```
用户请求 → 运行 go test → 检查覆盖率 → 输出测试结果 → 完成
```

### 场景 4：文档生成与部署

```
用户请求 → 执行 swag init → 生成 Swagger 文档 → 完成
用户请求 → 执行 docker build → 构建镜像 → 完成
```

## 功能 1：代码格式化

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

## 功能 2：项目构建

**构建项目**：
```bash
go build -o ./build/ixpay-pro.exe ./cmd/ixpay-pro/main.go
```

**运行项目**：
```bash
go run cmd/ixpay-pro/main.go
```

## 功能 3：依赖管理

**下载依赖**：
```bash
go mod download
```

**整理依赖**：
```bash
go mod tidy
```

**添加新依赖**：
```bash
go get github.com/gin-gonic/gin
```

## 功能 4：测试执行

**运行所有测试**：
```bash
go test ./...
```

**运行测试并显示覆盖率**：
```bash
go test ./... -cover
```

**运行特定包的测试**：
```bash
go test ./internal/domain/base/service/...
```

**运行测试并生成覆盖率报告**：
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 功能 5：Wire 依赖注入代码生成

**生成 Wire 代码**：
```bash
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

## 功能 6：API Swagger 文档生成

**触发条件**：新增或修改 API 接口后

**执行命令**：
```bash
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseDependency --parseInternal --parseDepth 1
```

**生成内容**：
- Swagger JSON 文件（docs/swagger.json）
- Swagger YAML 文件（docs/swagger.yaml）
- API 文档 HTML（docs/docs.go）

**文档访问**：http://localhost:8586/swagger/index.html

## 功能 7：Docker 部署

**构建 Docker 镜像**：
```bash
docker build -t ixpay-pro .
```

**运行容器**：
```bash
docker run -p 8586:8586 ixpay-pro
```

**Dockerfile 说明**：
- 使用多阶段构建减少镜像大小
- 构建阶段：`golang:1.24.6-alpine`
- 运行阶段：`alpine:3.19`
- 使用非 root 用户运行应用
- 包含健康检查配置