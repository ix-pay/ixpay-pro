---
name: 后端开发工具集
description: 当用户进行 Go 后端开发时自动触发，包括 API 开发、代码构建、单元测试、依赖管理、Swagger 文档生成和 Docker 部署等场景
---

## 核心功能概览

本技能集成了后端开发的全流程工具：

1. **API 脚手架生成** - 自动生成符合 DDD 分层架构的完整代码框架
2. **代码格式化** - 自动执行 go fmt 和 go imports
3. **单元测试生成** - 自动生成 Go 单元测试，支持 Mock 实现
4. **开发命令执行** - 构建、运行、依赖管理、Docker 部署
5. **API 文档生成** - 自动生成 Swagger 文档
6. **依赖注入配置** - 自动生成 Wire 依赖注入配置

***

## 功能 1：API 脚手架生成与规范检查

### 自动检查项目

#### 1.1 API 路径检查

- 检测路由定义是否符合 `/api/[module]//` 格式
- 不符合时主动提醒并提供修正建议

#### 1.2 响应格式检查

- 检测是否使用统一响应格式
- 检测响应结构是否符合规范

#### 1.3 分层架构检查

- 检测 API Handler 中是否有直接操作数据库的代码
- 发现 `db.Where()`、`db.First()` 等直接调用时提醒
- 检测领域层是否依赖基础设施层

#### 1.4 参数验证检查

- 检测字段是否有 `binding` 标签
- 检测是否有 `validate` 标签

#### 1.5 日志记录检查

- 检测登录、删除等操作是否有日志记录
- 检测日志是否包含 user\_id、IP 等关键信息

***

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

***

## 功能 3：单元测试生成

**触发条件**：用户说"生成单元测试"、"编写测试"

**自动生成**：

- 符合 Go 测试规范的单元测试文件
- 包含完整的测试用例
- 使用 testify 等测试框架（如果项目已集成）
- 自动生成 Mock 实现用于单元测试

***

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

### 4.4 wire 依赖注入代码生成

```bash
# 生成 Wire 依赖注入代码
wire gen ./internal/app
```

### 4.5 Docker 部署

```bash
# 构建 Docker 镜像
docker build -t ixpay-pro .

# 运行容器
docker run -p 8586:8586 ixpay-pro
```

***

## 功能 5：API swagger 文档自动生成

**触发条件**：新增或修改 API 接口后自动触发

**执行命令**：

```bash
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseDependency --parseInternal --parseDepth 1
```

**生成内容**：

- Swagger JSON 文件（docs/swagger.json）
- Swagger YAML 文件（docs/swagger.yaml）
- API 文档 HTML（docs/docs.go）

**文档访问**：<http://localhost:8586/swagger/index.html>