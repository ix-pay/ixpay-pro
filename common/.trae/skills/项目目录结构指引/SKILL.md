---
name: 项目目录结构指引
description: 当 AI 需要创建任何文件（代码文件、配置文件、文档等）或不确定文件位置时触发，用于确定正确的文件创建位置，确保项目结构规范性和一致性
---

# 项目目录结构指引

## 工作流程

```
用户要求创建文件
    ↓
判断文件类型（后端/前端/网关/文档）
    ↓
查询对应目录结构
    ↓
返回完整文件路径
    ↓
在正确位置创建文件
```

## 项目概览

| 项目 | 路径 | 说明 |
|------|------|------|
| 后端服务 | `server/` | Go DDD 架构后端 |
| 前端管理后台 | `web/` | Vue3 + TypeScript |
| API 网关 | `gxy/` | Go 标准库网关 |
| H5 应用 | `h5app/` | 待开发 |
| 企业微信 H5 | `weapp/` | 待开发 |
| 小程序 | `miniapp/` | 待开发 |
| 公共资源 | `common/` | 规则、技能、文档 |

## 后端目录结构（server/）

### 核心架构（DDD 分层）

```
server/
├── cmd/ixpay-pro/main.go          # 应用入口
├── configs/config.yaml            # 配置文件
├── docs/                          # Swagger 文档
├── internal/                      # 核心代码
│   ├── app/                       # 【应用层】
│   │   ├── base/                  #   基础管理模块
│   │   │   ├── api/               #   API Handler
│   │   │   ├── middleware/        #   中间件
│   │   │   ├── migrations/        #   数据库迁移
│   │   │   ├── application.go     #   应用服务入口
│   │   │   ├── routes.go          #   路由配置
│   │   │   └── wire.go            #   依赖注入
│   │   └── wx/                    #   微信支付模块（同上结构）
│   ├── domain/                    # 【领域层】
│   │   ├── base/                  #   基础领域
│   │   │   ├── entity/            #   领域实体
│   │   │   ├── repo/              #   仓库接口
│   │   │   └── service/           #   领域服务
│   │   ├── wx/                    #   微信领域（同上结构）
│   │   └── converter/             #   实体转换器
│   ├── persistence/               # 【基础设施层】
│   │   ├── base/                  #   base 模块仓库实现
│   │   ├── wx/                    #   wx 模块仓库实现
│   │   └── common/                #   通用持久化工具
│   ├── dto/                       # 【数据传输对象】
│   │   ├── base/request/          #   请求 DTO
│   │   ├── base/response/         #   响应 DTO
│   │   └── wx/                    #   wx 模块 DTO
│   ├── infrastructure/            # 【技术基础设施】
│   │   ├── observability/         #   可观测性（日志、监控）
│   │   ├── persistence/           #   数据库、缓存、Redis
│   │   ├── security/              #   JWT、权限、验证码
│   │   ├── support/               #   错误处理、雪花算法、任务
│   │   └── transport/             #   HTTP 响应、中间件
│   ├── config/                    # 配置管理
│   └── utils/                     # 工具函数
└── tests/unit/                    # 单元测试
```

### 后端文件位置速查表

| 文件类型 | 路径模板 | 示例 |
|---------|---------|------|
| API Handler | `internal/app/[模块]/api/` | `internal/app/base/api/user_handler.go` |
| 领域实体 | `internal/domain/[模块]/entity/` | `internal/domain/base/entity/user.go` |
| 仓库接口 | `internal/domain/[模块]/repo/` | `internal/domain/base/repo/user_repository.go` |
| 领域服务 | `internal/domain/[模块]/service/` | `internal/domain/base/service/user_service.go` |
| 仓库实现 | `internal/persistence/[模块]/` | `internal/persistence/base/user_repository.go` |
| DTO 请求 | `internal/dto/[模块]/request/` | `internal/dto/base/request/user.go` |
| DTO 响应 | `internal/dto/[模块]/response/` | `internal/dto/base/response/user.go` |
| 单元测试 | `tests/unit/[层]/[模块]/` | `tests/unit/domain/base/service/user_test.go` |
| 需求文档 | `server/.trae/documents/` | `server/.trae/documents/用户管理需求.md` |

## 前端目录结构（web/）

### 核心架构

```
web/
├── src/
│   ├── api/modules/               # API 接口（按模块）
│   ├── views/                     # 页面视图
│   │   ├── base/                  #   基础管理（首页、登录、资料、设置）
│   │   ├── system/                #   系统管理（用户、角色、菜单等）
│   │   ├── log/                   #   日志（登录日志、操作日志）
│   │   ├── monitor/               #   监控（系统监控、在线用户）
│   │   └── task/                  #   任务（定时任务）
│   ├── components/                # 组件
│   │   ├── business/              #   业务组件（Card、Chart、PageTemplate 等）
│   │   ├── layout/                #   布局组件（BaseLayout、Header、Sidebar 等）
│   │   └── IconSelector/          #   图标选择器
│   ├── stores/modules/            # Pinia 状态管理（app、router、user）
│   ├── types/                     # TypeScript 类型定义
│   ├── utils/                     # 工具函数（request、permission 等）
│   ├── app/                       # 应用配置（router、App.vue、main.ts）
│   ├── assets/styles/             # 样式资源（global、tailwind、design-tokens）
│   ├── directive/                 # 自定义指令（auth、auth-btn）
│   ├── hooks/                     # 组合式函数
│   └── core/                      # 核心功能
├── public/                        # 公共静态资源
└── .env.*                         # 环境变量
```

### 前端文件位置速查表

| 文件类型 | 路径模板 | 示例 |
|---------|---------|------|
| API 接口 | `src/api/modules/` | `src/api/modules/user.ts` |
| 页面视图 | `src/views/[模块]/[功能]/` | `src/views/system/user/index.vue` |
| 业务组件 | `src/components/business/` | `src/components/business/Card/index.vue` |
| 布局组件 | `src/components/layout/` | `src/components/layout/BaseLayout.vue` |
| Store | `src/stores/modules/` | `src/stores/modules/user.ts` |
| 类型定义 | `src/types/` | `src/types/user.ts` |
| 工具函数 | `src/utils/` | `src/utils/request.ts` |
| 需求文档 | `web/.trae/documents/` | `web/.trae/documents/用户界面需求.md` |

## 网关目录结构（gxy/）

```
gxy/
├── cmd/gateway/main.go            # 应用入口
├── internal/                      # 核心代码
│   ├── api/                       # API 处理和路由
│   ├── cluster/                   # 集群节点管理
│   ├── discovery/                 # 服务注册与健康检查
│   ├── loadbalance/               # 负载均衡（轮询）
│   └── proxy/                     # 请求代理转发
├── pkg/                           # 公共包
│   ├── config/                    # 配置管理
│   └── utils/                     # 工具函数（HTTP、日志）
└── config.json                    # 网关配置
```

### 网关文件位置速查表

| 文件类型 | 路径模板 | 示例 |
|---------|---------|------|
| API 处理 | `internal/api/` | `internal/api/handler.go` |
| 集群管理 | `internal/cluster/` | `internal/cluster/node.go` |
| 服务发现 | `internal/discovery/` | `internal/discovery/registry.go` |
| 负载均衡 | `internal/loadbalance/` | `internal/loadbalance/roundrobin.go` |
| 代理转发 | `internal/proxy/` | `internal/proxy/proxy.go` |
| 配置管理 | `pkg/config/` | `pkg/config/config.go` |
| 工具函数 | `pkg/utils/` | `pkg/utils/log.go` |

## 命名规范

### 后端（Go）
- 文件名：`snake_case`（如 `user_handler.go`）
- 包名：小写，与目录名一致
- 结构体：`PascalCase`（如 `UserDTO`）
- JSON 标签：`camelCase`（如 `json:"userId"`）
- ID 字段：应用层/DTO 用 `string`，数据库用 `int64`

### 前端（Vue3 + TypeScript）
- 文件名：`kebab-case`（如 `user-list.vue`）
- 组件名：`PascalCase`（如 `UserList`）
- 变量/函数：`camelCase`（如 `getUserInfo`）
- TypeScript 接口：`PascalCase`（如 `UserDTO`）

## 核心原则

1. **领域优先**：新业务功能优先创建 Domain 层代码
2. **依赖方向**：上层依赖下层，禁止反向依赖
3. **接口与实现分离**：Domain 层定义接口，Persistence 层实现
4. **DTO 独立**：DTO 与 Domain Entity 分离，避免耦合
5. **测试配套**：创建业务代码时同步创建单元测试
6. **文档同步**：重要功能创建对应的需求文档
7. **分层架构**：严格遵循 DDD 分层架构
8. **类型转换**：Repository 层负责 `string ↔ int64` 转换
9. **错误处理**：使用 `fmt.Errorf` 包装错误，保留错误链
10. **跨模块调用**：通过 Domain 层，避免调用 Application 或 API 层

## 项目说明

- **server/**：Go DDD 架构后端，包含 base（基础管理）和 wx（微信支付）模块
- **web/**：Vue3 + TypeScript 管理后台，使用 Pinia + Vue Router + Tailwind CSS
- **gxy/**：Go 标准库 API 网关，实现服务发现、负载均衡、集群同步
- **h5app/weapp/miniapp/**：移动端项目，当前为初始化状态
- **common/**：规则、技能、文档等共享资源
