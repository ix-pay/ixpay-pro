---
name: 项目目录结构指引
description: 当 AI 需要创建任何文件（代码文件、配置文件、文档等）或不确定文件位置时触发，用于确定正确的文件创建位置，确保项目结构规范性和一致性
---

# 项目目录结构指引

## 功能描述

本技能提供 IXPay Pro 项目的完整目录结构，包括后端（server）、前端（web）、公共目录（common）、H5 应用（h5app）、企业微信 H5 应用（weapp）、小程序（miniapp）和网关（gxy）。当 AI 需要创建任何文件时，必须首先调用此技能来确定正确的文件位置。

**核心价值**：
- ✅ 确保所有文件都在正确的位置
- ✅ 保持项目结构的一致性和规范性
- ✅ 避免文件乱放导致的维护困难
- ✅ 提供快速查找和定位能力

## 核心功能

### 1. 完整目录结构展示

提供后端、前端、公共目录的完整树形结构，包括：
- 所有主要目录和子目录
- 关键文件的位置
- 目录职责说明

### 2. 文件创建位置快速查询

通过表格形式提供不同类型文件的标准位置：
- 后端文件位置表
- 前端文件位置表
- 文档文件位置表

### 3. 命名规范指导

提供各类型文件的命名规范：
- Go 文件命名规范（snake_case）
- Vue 文件命名规范（kebab-case）
- TypeScript 类型命名规范（PascalCase）

### 4. 使用示例

提供典型的文件创建场景示例，展示完整的决策过程。

## 工作流程

```
用户要求创建文件
    ↓
AI 调用本技能
    ↓
判断文件类型
    ├─ 后端代码 → 查询后端目录结构
    ├─ 前端代码 → 查询前端目录结构
    ├─ 配置文件 → 查询 configs/或对应模块目录
    └─ 文档文件 → 查询 .trae/documents/或 specifications/
    ↓
根据文件类型查找对应位置
    ↓
返回完整的文件路径
    ↓
AI 在正确位置创建文件
```

## 使用示例

### 示例 1：创建用户管理 API

**用户**：创建一个获取用户详情的 API

**技能响应**：
1. 判断：这是后端 API 文件
2. 查询后端目录结构 → `server/internal/app/base/api/`
3. 返回完整路径：`server/internal/app/base/api/user_handler.go`
4. 同时建议创建相关文件：
   - DTO: `server/internal/dto/base/response/user.go`
   - 如需要，领域服务：`server/internal/domain/base/service/user_domain_service.go`

### 示例 2：创建用户列表页面

**用户**：创建一个用户列表页面

**技能响应**：
1. 判断：这是前端页面文件
2. 查询前端目录结构 → `web/src/views/base/user/`
3. 返回完整路径：`web/src/views/base/user/index.vue`
4. 同时建议创建相关文件：
   - API 接口：`web/src/api/modules/user.ts`
   - TypeScript 类型：`web/src/types/user.ts`

### 示例 3：创建支付领域服务

**用户**：创建支付领域的订单查询功能

**技能响应**：
1. 判断：这是后端领域服务，需要创建完整的领域层代码
2. 依次返回以下文件路径：
   - 领域实体：`server/internal/domain/wx/entity/payment.go`
   - 领域服务：`server/internal/domain/wx/service/payment_service.go`
   - 仓库接口：`server/internal/domain/wx/repo/payment_repository.go`
   - 仓库实现：`server/internal/persistence/base/payment_repository.go`
   - API Handler：`server/internal/app/wx/api/payment_handler.go`

### 示例 4：创建后端需求文档

**用户**：写一个用户管理的需求文档

**技能响应**：
1. 判断：这是后端需求文档
2. 查询文档目录 → `server/.trae/documents/`
3. 返回完整路径：`server/.trae/documents/用户管理需求.md`

### 示例 5：创建网关健康检查功能

**用户**：创建网关的健康检查功能

**技能响应**：
1. 判断：这是网关核心功能
2. 查询网关目录结构 → `gxy/internal/discovery/`
3. 返回完整路径：`gxy/internal/discovery/health.go`
4. 相关文件：
   - 服务注册中心：`gxy/internal/discovery/registry.go`
   - 配置管理：`gxy/pkg/config/config.go`
   - 日志工具：`gxy/pkg/utils/log.go`

## 目录结构详情

### 后端目录结构（server/）

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
│   │   │   ├── api/             # API 处理器
│   │   │   ├── middleware/      # 模块中间件
│   │   │   ├── migrations/      # 数据库迁移文件
│   │   │   ├── seed/            # 数据库种子文件
│   │   │   ├── routes.go        # 路由配置
│   │   │   └── wire.go          # 依赖注入配置
│   │   ├── wx/                  # 微信支付模块
│   │   │   ├── api/             # API 处理器
│   │   │   ├── middleware/      # 模块中间件
│   │   │   ├── migrations/      # 数据库迁移文件
│   │   │   ├── routes.go        # 路由配置
│   │   │   └── wire.go          # 依赖注入配置
│   │   ├── application.go       # 应用总入口
│   │   ├── routes.go            # 全局路由配置
│   │   ├── wire.go              # 依赖注入配置
│   │   └── wire_gen.go          # 依赖注入配置（自动生成）
│   │
│   ├── domain/                  # 【领域层】核心业务逻辑，不依赖基础设施
│   │   ├── base/                # 基础管理领域
│   │   │   ├── entity/          # 领域实体
│   │   │   ├── repo/            # 仓库接口定义
│   │   │   └── service/         # 领域服务
│   │   └── wx/                  # 微信支付领域
│   │       ├── entity/          # 领域实体
│   │       ├── repo/            # 仓库接口定义
│   │       └── service/         # 领域服务
│   │
│   ├── persistence/             # 【基础设施层】实现 Repository 接口
│   │   ├── base/                # Base 模块仓库实现
│   │   ├── wx/                  # Wx 模块仓库实现
│   │   └── common/              # 通用持久化工具
│   │
│   ├── infrastructure/          # 【基础设施层】通用技术组件
│   │   ├── observability/       # 可观测性（日志、监控）
│   │   ├── persistence/         # 数据持久化（数据库、缓存、Redis）
│   │   ├── security/            # 安全相关（认证、验证码）
│   │   ├── support/             # 支撑工具（错误处理、雪花算法、任务管理）
│   │   └── transport/           # 传输层（HTTP、中间件）
│   │
│   ├── dto/                     # 【数据传输对象】请求和响应数据结构
│   │   ├── base/
│   │   │   ├── request/         # 请求 DTO
│   │   │   └── response/        # 响应 DTO
│   │   └── wx/
│   │       ├── request/         # 请求 DTO
│   │       └── response/        # 响应 DTO
│   │
│   ├── config/                  # 配置管理
│   └── utils/                   # 工具函数
│
├── tests/                       # 【测试目录】独立测试目录
│   └── unit/                    # 单元测试
│       └── domain/
│           └── base/
│               └── service/     # 领域服务测试
│
├── .trae/documents/            # 后端需求计划文档
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

### 前端目录结构（web/）

```
web/
├── src/
│   ├── api/                  # API 接口
│   │   ├── index.ts          # API 总入口
│   │   ├── monitor.ts        # 监控 API
│   │   └── modules/          # 按模块组织的 API 服务
│   │       ├── user.ts       # 用户 API
│   │       ├── role.ts       # 角色 API
│   │       ├── menu.ts       # 菜单 API
│   │       └── ...
│   │
│   ├── views/                # 页面视图
│   │   ├── base/             # 基础管理模块
│   │   │   ├── index/        # 首页
│   │   │   ├── login/        # 登录页
│   │   │   ├── profile/      # 个人资料
│   │   │   ├── setting/      # 设置
│   │   │   └── user/         # 用户管理
│   │   ├── system/           # 系统管理模块
│   │   │   ├── api-route/    # API 路由管理
│   │   │   ├── config/       # 配置管理
│   │   │   ├── department/   # 部门管理
│   │   │   ├── dict/         # 字典管理
│   │   │   ├── menu/         # 菜单管理
│   │   │   ├── notice/       # 通知管理
│   │   │   ├── position/     # 岗位管理
│   │   │   ├── role/         # 角色管理
│   │   │   └── user/         # 用户管理
│   │   ├── log/              # 日志模块
│   │   │   ├── login-log/    # 登录日志
│   │   │   └── operation-log/# 操作日志
│   │   ├── monitor/          # 监控模块
│   │   │   ├── monitor/      # 系统监控
│   │   │   └── online-user/  # 在线用户
│   │   └── task/             # 任务模块
│   │       └── task/         # 定时任务
│   │
│   ├── components/           # 组件
│   │   ├── common/           # 通用组件
│   │   │   ├── IconSelector/ # 图标选择器
│   │   │   └── ...
│   │   ├── business/         # 业务组件
│   │   │   ├── Card/         # 卡片组件
│   │   │   ├── Chart/        # 图表组件
│   │   │   ├── PageTemplate/ # 页面模板
│   │   │   ├── StatCard/     # 统计卡片
│   │   │   ├── ThemePanel/   # 主题面板
│   │   │   └── error/        # 错误页面组件
│   │   └── layout/           # 布局组件
│   │       ├── BaseLayout.vue # 基础布局
│   │       ├── Header.vue    # 头部
│   │       ├── Sidebar.vue   # 侧边栏
│   │       ├── TabManager.vue # 标签管理
│   │       └── ContentWrapper.vue # 内容包装
│   │
│   ├── stores/               # 状态管理（Pinia）
│   │   ├── index.ts          # Store 总入口
│   │   └── modules/          # 模块
│   │       ├── app.ts        # 应用状态
│   │       ├── router.ts     # 路由状态
│   │       └── user.ts       # 用户状态
│   │
│   ├── types/                # TypeScript 类型定义
│   │   ├── index.ts          # 类型总入口
│   │   ├── user.ts           # 用户类型
│   │   ├── role.ts           # 角色类型
│   │   ├── menu.ts           # 菜单类型
│   │   ├── config.ts         # 配置类型
│   │   ├── department.ts     # 部门类型
│   │   ├── dict.ts           # 字典类型
│   │   ├── position.ts       # 岗位类型
│   │   └── date.d.ts         # 日期类型
│   │
│   ├── utils/                # 工具函数
│   │   ├── request.ts        # HTTP 请求封装
│   │   ├── asyncRouter.ts    # 异步路由
│   │   ├── bus.ts            # 事件总线
│   │   ├── date.ts           # 日期工具
│   │   ├── dictionary.ts     # 字典工具
│   │   ├── echarts.ts        # ECharts 工具
│   │   ├── format.ts         # 格式化工具
│   │   └── permission.ts     # 权限工具
│   │
│   ├── app/                  # 应用配置
│   │   ├── config/           # 配置
│   │   │   └── index.ts
│   │   ├── router/           # 路由配置
│   │   │   └── index.ts
│   │   ├── App.vue           # 根组件
│   │   └── main.ts           # 应用入口
│   │
│   ├── assets/               # 静态资源
│   │   ├── images/           # 图片资源
│   │   └── styles/           # 样式资源
│   │       ├── global.scss   # 全局样式
│   │       ├── tailwind.css  # Tailwind CSS
│   │       ├── transition.scss # 过渡动画
│   │       └── design-tokens.scss # 设计令牌
│   │
│   ├── directive/            # 自定义指令
│   │   └── auth.ts           # 权限指令
│   │
│   ├── hooks/                # 组合式函数
│   │   └── responsive.ts     # 响应式钩子
│   │
│   ├── core/                 # 核心功能
│   │   ├── global.ts         # 全局配置
│   │   └── ixpay-pro.ts      # IXPay Pro 核心
│   │
│   ├── auto-imports.d.ts     # 自动导入类型声明
│   ├── components.d.ts       # 组件类型声明
│   └── pathInfo.json         # 路径信息
│
├── public/                   # 公共静态资源
│   └── favicon.ico           # 网站图标
│
├── .vscode/                  # VSCode 配置
│   └── extensions.json       # 推荐插件
│
├── .env.development          # 开发环境变量
├── .env.production           # 生产环境变量
├── .editorconfig             # 编辑器配置
├── .eslintrc.cjs             # ESLint 配置
├── .gitignore
├── .prettierrc.json          # Prettier 配置
├── index.html                # HTML 入口
├── package.json              # 项目配置
├── package-lock.json         # 依赖锁定
├── postcss.config.js         # PostCSS 配置
├── tailwind.config.js        # Tailwind 配置
├── tsconfig.json             # TypeScript 配置
├── tsconfig.app.json         # 应用 TypeScript 配置
├── tsconfig.node.json        # Node TypeScript 配置
├── vite.config.mts           # Vite 配置
├── env.d.ts                  # 环境变量类型声明
└── README.md
```

### 公共目录结构（common/）

```
common/
├── .trae/
│   ├── rules/                # AI 规则文件
│   │   └── AI 说明书.md        # AI 核心行为准则
│   ├── skills/               # AI 技能文件
│   │   ├── 项目目录结构指引/
│   │   │   └── SKILL.md
│   │   ├── 规则与技能创建与编辑指南/
│   │   │   └── SKILL.md
│   │   ├── 代码质量检查器/
│   │   │   └── SKILL.md
│   │   ├── 全栈开发/
│   │   │   └── SKILL.md
│   │   ├── 前端开发工具集/
│   │   │   └── SKILL.md
│   │   ├── 后端开发工具集/
│   │   │   └── SKILL.md
│   │   └── git 提交推送/
│   │       └── SKILL.md
│   └── skill-config.json     # 技能配置
│
├── docs/                     # 公共文档
│   ├── images/               # 公共图片资源
│   │   ├── ixpay-master.png
│   │   ├── ixpay.png
│   │   └── ...
│   └── manual/               # 使用手册
│       ├── server 项目设计文档.md
│       ├── web 项目设计文档.md
│       └── 部署说明.md
│
├── LICENSE
└── README.md
```

### H5 应用目录结构（h5app/）

```
h5app/
└── README.md                 # 项目说明
```

**说明**：H5 移动端应用项目，当前为初始化状态，后续将根据实际开发需要添加代码。

### 企业微信 H5 应用目录结构（weapp/）

```
weapp/
└── README.md                 # 项目说明
```

**说明**：企业微信 H5 端应用项目，当前为初始化状态，后续将根据实际开发需要添加代码。

### 小程序目录结构（miniapp/）

```
miniapp/
└── README.md                 # 项目说明
```

**说明**：小程序项目，当前为初始化状态，后续将根据实际开发需要添加代码。

### 网关目录结构（gxy/）

```
gxy/
├── cmd/
│   └── gateway/
│       └── main.go           # 网关应用程序入口
│
├── internal/                 # 核心业务代码
│   ├── api/
│   │   ├── handler.go        # API 请求处理器
│   │   └── router.go         # 请求路由
│   ├── cluster/
│   │   ├── node.go           # 网关节点管理
│   │   └── sync.go           # 数据同步
│   ├── discovery/
│   │   ├── registry.go       # 服务注册中心
│   │   └── health.go         # 健康检查
│   ├── loadbalance/
│   │   └── roundrobin.go     # 轮询负载均衡
│   └── proxy/
│       └── proxy.go          # 请求代理转发
│
├── pkg/                      # 公共包
│   ├── config/
│   │   └── config.go         # 配置管理
│   └── utils/
│       ├── http.go           # HTTP 工具
│       └── log.go            # 日志工具
│
├── docs/                     # 网关文档
├── .gitignore
├── config.json               # 网关配置文件
├── Dockerfile                # Docker 镜像构建文件
├── gateway-dev-prompt.md     # 网关开发文档
├── go.mod                    # Go 模块依赖
├── performance_test_script.go # 性能测试脚本
└── README.md
```

**核心功能**：
- 服务注册与发现
- 负载均衡（轮询算法）
- 健康检查
- 网关集群数据同步
- 请求代理转发

**技术栈**：纯 Go 语言（标准库）

## 文件创建位置快速参考

### 后端文件

| 文件类型 | 创建位置 | 示例 |
|---------|---------|------|
| 应用程序入口 | `server/cmd/ixpay-pro/` | `server/cmd/ixpay-pro/main.go` |
| 配置文件 | `server/configs/` | `server/configs/config.yaml` |
| API Handler | `server/internal/app/[模块]/api/` | `server/internal/app/base/api/user_handler.go` |
| 应用服务 | `server/internal/app/[模块]/application/` | `server/internal/app/base/application/user_app_service.go` |
| 领域实体 | `server/internal/domain/[模块]/entity/` | `server/internal/domain/base/entity/user.go` |
| 仓库接口 | `server/internal/domain/[模块]/repo/` | `server/internal/domain/base/repo/user_repository.go` |
| 领域服务 | `server/internal/domain/[模块]/service/` | `server/internal/domain/base/service/user_domain_service.go` |
| 仓库实现 | `server/internal/persistence/[模块]/` | `server/internal/persistence/base/user_repository.go` |
| DTO 请求 | `server/internal/dto/[模块]/request/` | `server/internal/dto/base/request/user.go` |
| DTO 响应 | `server/internal/dto/[模块]/response/` | `server/internal/dto/base/response/user.go` |
| 单元测试 | `server/tests/unit/[层]/[模块]/` | `server/tests/unit/domain/base/service/user_service_test.go` |
| 集成测试 | `server/tests/integration/[类型]/` | `server/tests/integration/api/user_test.go` |
| 脚本文件 | `server/scripts/` | `server/scripts/migrate.sh` |
| 后端需求文档 | `server/.trae/documents/` | `server/.trae/documents/用户管理需求.md` |
| Swagger 文档 | `server/docs/` | `server/docs/swagger.json` |

### 前端文件

| 文件类型 | 创建位置 | 示例 |
|---------|---------|------|
| API 接口 | `web/src/api/modules/` | `web/src/api/modules/user.ts` |
| 页面视图 | `web/src/views/[模块]/[功能]/` | `web/src/views/base/user/index.vue` |
| 组件 | `web/src/components/[类型]/` | `web/src/components/common/modal.vue` |
| Store | `web/src/stores/` | `web/src/stores/user.ts` |
| TypeScript 类型 | `web/src/types/` | `web/src/types/user.ts` |
| 前端文档 | `web/.trae/documents/` | `web/.trae/documents/用户界面需求.md` |
| 技术文档 | `web/.trae/documents/` | `web/.trae/documents/组件设计.md` |

### 网关文件

| 文件类型 | 创建位置 | 示例 |
|---------|---------|------|
| 应用程序入口 | `gxy/cmd/gateway/` | `gxy/cmd/gateway/main.go` |
| API 处理器 | `gxy/internal/api/` | `gxy/internal/api/handler.go` |
| 路由配置 | `gxy/internal/api/` | `gxy/internal/api/router.go` |
| 集群节点管理 | `gxy/internal/cluster/` | `gxy/internal/cluster/node.go` |
| 服务注册中心 | `gxy/internal/discovery/` | `gxy/internal/discovery/registry.go` |
| 健康检查 | `gxy/internal/discovery/` | `gxy/internal/discovery/health.go` |
| 负载均衡 | `gxy/internal/loadbalance/` | `gxy/internal/loadbalance/roundrobin.go` |
| 请求代理 | `gxy/internal/proxy/` | `gxy/internal/proxy/proxy.go` |
| 配置管理 | `gxy/pkg/config/` | `gxy/pkg/config/config.go` |
| 工具函数 | `gxy/pkg/utils/` | `gxy/pkg/utils/http.go` |
| 网关配置 | `gxy/` | `gxy/config.json` |
| 网关文档 | `gxy/docs/` | `gxy/docs/网关设计文档.md` |

## 命名规范

### 后端（Go）
- 文件名：**snake_case**，如 `user_handler.go`
- 包名：小写，与目录名一致
- 结构体：PascalCase，如 `UserDTO`
- JSON 标签：camelCase，如 `json:"userId"`
- ID 字段：统一使用 `string` 类型（应用层和 DTO）
- 数据库 ID：使用 `int64` 类型

### 前端（Vue3 + TypeScript）
- 文件名：kebab-case，如 `user-list.vue`
- 组件名：PascalCase，如 `UserList`
- 变量/函数：camelCase，如 `getUserInfo`
- TypeScript 接口：PascalCase，如 `UserDTO`

### 网关（Go）
- 文件名：**snake_case**，如 `roundrobin.go`
- 包名：小写，与目录名一致
- 结构体：PascalCase，如 `ServiceInstance`
- JSON 标签：camelCase，如 `json:"instance_id"`

## 核心原则

### 通用原则

1. **领域优先**：新业务功能优先创建 Domain 层代码，确保业务逻辑的核心地位
2. **依赖方向**：上层依赖下层，禁止反向依赖，遵循依赖倒置原则
3. **接口与实现分离**：Domain 层定义接口，Infrastructure 层实现，保持业务逻辑的独立性
4. **DTO 独立**：DTO 与 Domain Entity 分离，避免耦合，确保数据传输的灵活性
5. **测试配套**：创建业务代码时同步创建单元测试，确保代码质量
6. **文档同步**：重要功能需要创建对应的需求文档和技术文档
7. **分层架构**：严格遵循 DDD 分层架构，确保代码结构清晰
8. **类型转换策略**：Repository 层负责 `string ↔ int64` 的转换，保持 ID 类型的一致性
9. **错误处理规范**：使用 `fmt.Errorf` 包装错误，保留错误链，提供清晰的错误信息
10. **跨模块调用**：跨模块调用应通过 Domain 层，避免调用 Application 层或 API 层

### 网关原则

1. **高性能**：使用 Go 语言标准库，避免不必要的依赖
2. **高可用**：集群部署、故障自动转移、健康检查
3. **线程安全**：使用 sync 包确保并发安全
4. **轻量级**：保持代码简洁，避免过度设计
5. **可观测性**：完善的日志记录、监控指标

## 相关文件

- [AI 说明书](../../../rules/AI 说明书.md) - 指导 AI 如何使用本技能
- [Go 代码风格规范](../../../rules/Go 代码风格与开发规范.md)
- [Vue 代码风格规范](../../../rules/vue 代码风格规范.md)
- [前端技术栈规范](../../../rules/前端技术栈规范.md)
- [网关开发文档](../../../../gxy/gateway-dev-prompt.md) - 网关开发详细指南

## 项目说明

### 后端（server/）
基于 Go 语言的 DDD 架构后端系统，采用分层架构设计，包含应用层、领域层、基础设施层。

### 前端（web/）
基于 Vue3 + TypeScript 的 Web 管理后台，使用 Pinia 状态管理、Vue Router 路由管理。

### H5 应用（h5app/）
H5 移动端应用项目（初始化状态），后续将基于 Vue3 + TypeScript 技术栈开发。

### 企业微信 H5 应用（weapp/）
企业微信 H5 端应用项目（初始化状态），后续将根据实际开发需要添加代码。

### 小程序（miniapp/）
小程序项目（初始化状态），后续将根据实际开发需要添加代码。

### 网关（gxy/）
基于纯 Go 语言标准库开发的轻量级 API 网关，实现服务注册发现、负载均衡、健康检查和集群同步功能。

### 公共目录（common/）
存放项目公共文档、技能、规则等共享资源。
