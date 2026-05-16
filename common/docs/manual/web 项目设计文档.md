# ixpay-pro Web 前端项目设计文档

## 1. 项目概述

### 1.1 项目简介

ixpay-pro 是一个基于 Vue 3 + TypeScript + Vite 构建的企业级中后台管理系统前端项目。项目采用现代化的前端技术栈，提供了完整的用户认证、权限管理、动态路由、菜单配置等功能。

### 1.2 技术栈

- **核心框架**: Vue 3.5.18（Composition API）
- **构建工具**: Vite 7.0.6
- **开发语言**: TypeScript 5.8.0
- **状态管理**: Pinia 3.0.3
- **路由管理**: Vue Router 4.5.1
- **UI 组件库**: Element Plus 2.11.2
- **HTTP 客户端**: Axios 1.12.2
- **工具库**: @vueuse/core 9.13.0
- **图表库**: ECharts 6.0.0
- **样式预处理**: Sass 1.92.1

### 1.3 开发环境要求

- Node.js: ^20.19.0 || >=22.12.0
- 推荐 IDE: VSCode + Volar 插件

## 2. 项目架构

### 2.1 目录结构

```
web/
├── public/                    # 静态资源目录
├── src/
│   ├── api/                   # API 接口层
│   │   ├── modules/          # 按模块划分的 API 接口
│   │   └── index.ts          # API 统一导出
│   ├── app/                   # 应用核心配置
│   │   ├── config/           # 应用配置
│   │   ├── router/           # 路由配置
│   │   ├── App.vue           # 根组件
│   │   └── main.ts           # 入口文件
│   ├── assets/                # 静态资源
│   │   ├── images/           # 图片资源
│   │   └── styles/           # 全局样式
│   ├── components/            # 组件库
│   │   ├── business/         # 业务组件
│   │   └── layout/           # 布局组件
│   ├── core/                  # 核心功能
│   │   ├── global.ts         # 全局注册
│   │   └── ixpay-pro.ts      # 框架核心
│   ├── directive/             # 自定义指令
│   │   └── auth.ts           # 权限指令
│   ├── hooks/                 # 组合式函数
│   │   └── responsive.ts     # 响应式 hook
│   ├── stores/                # 状态管理
│   │   ├── modules/          # 模块 store
│   │   └── index.ts          # store 入口
│   ├── types/                 # TypeScript 类型定义
│   │   ├── date.d.ts         # 日期类型
│   │   ├── index.ts          # 通用类型
│   │   └── menu.ts           # 菜单类型
│   ├── utils/                 # 工具函数
│   │   ├── asyncRouter.ts    # 动态路由处理
│   │   ├── bus.ts            # 事件总线
│   │   ├── date.ts           # 日期工具
│   │   ├── dictionary.ts     # 字典工具
│   │   ├── format.ts         # 格式化工具
│   │   ├── permission.ts     # 权限工具
│   │   └── request.ts        # HTTP 请求封装
│   └── views/                  # 页面视图
│       ├── base/             # 基础页面
│       ├── log/              # 日志管理
│       ├── monitor/          # 监控管理
│       ├── system/           # 系统管理
│       └── task/             # 任务管理
├── docs/                      # 项目文档
│   └── manual/               # 手动文档
├── .env.development          # 开发环境变量
├── .env.production           # 生产环境变量
├── vite.config.mts           # Vite 配置
├── tsconfig.json             # TypeScript 配置
└── package.json              # 项目依赖
```

### 2.2 架构分层

项目采用经典的三层架构：

```
┌─────────────────────────────────────┐
│         View Layer (视图层)          │
│   - views/ (页面组件)                 │
│   - components/ (可复用组件)          │
└─────────────────────────────────────┘
                ↓
┌─────────────────────────────────────┐
│       Store Layer (状态层)           │
│   - stores/ (Pinia 状态管理)          │
│   - hooks/ (组合式函数)              │
└─────────────────────────────────────┘
                ↓
┌─────────────────────────────────────┐
│        API Layer (接口层)            │
│   - api/ (API 接口封装)               │
│   - utils/request.ts (HTTP 封装)      │
└─────────────────────────────────────┘
```

## 3. 核心模块设计

### 3.1 路由系统

#### 3.1.1 路由配置

路由系统位于 [`src/app/router/index.ts`](file:///d:/g/ixpay-pro/web/src/app/router/index.ts)，采用动态路由 + 静态路由结合的方式：

- **固定路由**：登录页、基础布局等不需要权限的路由
- **动态路由**：根据用户权限从后端获取的菜单路由

#### 3.1.2 路由守卫

```typescript
router.beforeEach(async (to, from, next) => {
  // 1. 检查 token，未登录重定向到登录页
  // 2. 已登录访问登录页，重定向到首页
  // 3. 检查动态路由是否已加载
  // 4. 未加载则获取用户信息和动态路由
  // 5. 将动态路由添加到 layout 下
  // 6. 重新导航以应用新路由
})
```

#### 3.1.3 路由特点

- 支持动态路由加载
- 支持路由级权限控制
- 支持页面缓存（keep-alive）
- 支持面包屑导航
- 支持多标签页管理

### 3.2 状态管理

#### 3.2.1 Store 模块

项目使用 Pinia 进行状态管理，主要包含三个核心 store：

1. **user store** ([`src/stores/modules/user.ts`](file:///d:/g/ixpay-pro/web/src/stores/modules/user.ts))
   - 用户信息管理
   - Token 管理
   - 登录/登出功能

2. **router store** ([`src/stores/modules/router.ts`](file:///d:/g/ixpay-pro/web/src/stores/modules/router.ts))
   - 动态路由管理
   - 菜单管理（顶部菜单、左侧菜单）
   - Keep-alive 路由管理

3. **app store** ([`src/stores/modules/app.ts`](file:///d:/g/ixpay-pro/web/src/stores/modules/app.ts))
   - 应用配置管理
   - 主题管理
   - 全局状态

#### 3.2.2 持久化策略

Pinia 插件实现状态持久化：

```typescript
// 持久化键名：pinia_${store.$id}
// 存储方式：localStorage
// 特殊处理：router store 只保存特定字段
```

### 3.3 HTTP 请求封装

#### 3.3.1 核心功能

[`src/utils/request.ts`](file:///d:/g/ixpay-pro/web/src/utils/request.ts) 提供了完整的 HTTP 请求封装：

- **请求拦截器**：
  - 自动添加 Token（Bearer 认证）
  - 自动显示 Loading
  - 主动检查并刷新快过期的 Token

- **响应拦截器**：
  - 统一响应格式处理
  - 错误处理和提示
  - 401 自动刷新 Token
  - 刷新失败自动跳转登录

#### 3.3.2 Token 管理机制

```typescript
class TokenManager {
  // 1. 解析 JWT Token
  parseJwt(token: string)
  
  // 2. 检查 Token 是否快过期（默认提前 10 分钟）
  isTokenExpiringSoon(token: string, thresholdMinutes: number = 10)
  
  // 3. 主动刷新 Token
  async proactiveRefresh()
  
  // 4. 刷新 Token
  async refresh(refreshTokenValue: string)
  
  // 5. 处理 401 错误
  async handleUnauthorizedError(config)
}
```

#### 3.3.3 Loading 管理

```typescript
class LoadingManager {
  // 延迟显示（400ms），避免闪烁
  // 强制关闭（30 秒超时），防止 Loading 卡死
  // 计数管理，多个请求同时进行时只保持一个 Loading
}
```

### 3.4 权限系统

#### 3.4.1 权限控制方式

1. **路由级权限**：通过动态路由实现
2. **菜单级权限**：通过 `hasMenuPermission()` 过滤菜单
3. **按钮级权限**：通过自定义指令 `v-auth` 实现

#### 3.4.2 权限工具函数

[`src/utils/permission.ts`](file:///d:/g/ixpay-pro/web/src/utils/permission.ts) 提供：

```typescript
// 检查单个权限
hasPermission(permission: string)

// 检查任意一个权限
hasAnyPermission(permissions: string[])

// 检查所有权限
hasAllPermissions(permissions: string[])

// 检查菜单权限
hasMenuPermission(menu: ApiMenuItem)

// 获取用户权限列表
getPermissions()

// 角色相关
hasRole(role: string)
hasAnyRole(roles: string[])
hasAllRoles(roles: string[])
getRoles()
```

### 3.5 动态路由处理

#### 3.5.1 路由格式化

[`src/utils/asyncRouter.ts`](file:///d:/g/ixpay-pro/web/src/utils/asyncRouter.ts) 负责动态路由处理：

```typescript
// 1. 使用 import.meta.glob 动态导入组件
const viewModules = import.meta.glob('../views/**/*.vue')
const pluginModules = import.meta.glob('../plugin/**/*.vue')

// 2. 递归处理路由，将字符串路径转换为组件函数
asyncRouterHandle(asyncRouter: ExtendedRouteRecordRaw[])

// 3. 支持三种匹配方式
//    - 精确匹配
//    - 不区分大小写匹配
//    - 前缀匹配
```

#### 3.5.2 路由加载流程

```
1. 用户登录成功
   ↓
2. 触发 router.beforeEach
   ↓
3. 检查 asyncRouterFlag
   ↓
4. 调用 routerStore.SetAsyncRouter()
   ↓
5. 从 API 获取菜单数据
   ↓
6. 根据权限过滤路由
   ↓
7. 格式化路由（component 字符串转函数）
   ↓
8. 添加到 router（router.addRoute）
   ↓
9. 重新导航
```

## 4. 组件设计

### 4.1 布局组件

#### 4.1.1 BaseLayout

[`src/components/layout/BaseLayout.vue`](file:///d:/g/ixpay-pro/web/src/components/layout/BaseLayout.vue) 是主布局组件：

```vue
<el-container>
  <el-aside>     <!-- 侧边栏 -->
    <gva-aside />
  </el-aside>
  
  <el-container> <!-- 右侧区域 -->
    <el-header>  <!-- 顶部 -->
      <gva-header />
    </el-header>
    
    <el-main>    <!-- 内容区 -->
      <tab-manager />
    </el-main>
    
    <el-footer>  <!-- 底部 -->
      <BottomInfo />
    </el-footer>
  </el-container>
</el-container>
```

#### 4.1.2 布局特点

- 左侧固定宽度（可收缩：240px ↔ 64px）
- 右侧上中下布局
- 支持响应式设计
- 支持暗黑模式
- 支持水印功能

### 4.2 业务组件

#### 4.2.1 组件分类

1. **布局组件**：`components/layout/`
   - BaseLayout.vue - 主布局
   - Header.vue - 顶部导航
   - Sidebar.vue - 侧边栏菜单
   - TabManager.vue - 多标签管理
   - ContentWrapper.vue - 内容包装器

2. **业务组件**：`components/business/`
   - application/index.vue - 应用相关组件
   - bottomInfo/bottomInfo.vue - 底部信息组件
   - error/index.vue - 错误页面组件
   - error/reload.vue - 错误刷新组件
   - errorPreviews/index.vue - 错误预览组件

### 4.3 指令系统

#### 4.3.1 权限指令

[`src/directive/auth.ts`](file:///d:/g/ixpay-pro/web/src/directive/auth.ts) 提供 `v-auth` 指令：

```typescript
// 使用方式
<button v-auth="'user:add'">添加用户</button>

// 实现原理
// 1. 获取指令参数（权限标识）
// 2. 检查用户是否有该权限
// 3. 无权限则移除元素
```

## 5. API 接口设计

### 5.1 接口规范

所有 API 接口位于 `src/api/modules/` 目录，按功能模块划分：

- `user.ts` - 用户管理
- `role.ts` - 角色管理
- `menu.ts` - 菜单管理
- `department.ts` - 部门管理
- `position.ts` - 职位管理
- `dict.ts` - 字典管理
- `config.ts` - 系统配置
- `notice.ts` - 公告管理
- `api-route.ts` - API 路由
- `task.ts` - 定时任务
- `operation-log.ts` - 操作日志
- `login-log.ts` - 登录日志
- `online-user.ts` - 在线用户
- `jwt.ts` - JWT Token 管理

### 5.2 接口格式

```typescript
// 统一返回格式
interface ApiResponse<T = unknown> {
  code: number      // 0: 成功，非 0: 失败
  data?: T          // 返回数据
  msg?: string      // 消息提示
}
```

### 5.3 接口示例

```typescript
// 用户登录
export const login = (data: LoginInfo): Promise<ApiResponse> => {
  return service({
    url: '/auth/login',
    method: 'post',
    data: data,
  })
}

// 获取用户信息
export const getUserInfo = (): Promise<ApiResponse<{ userInfo: UserInfo }>> => {
  return service({
    url: '/user/info',
    method: 'get',
  })
}
```

## 6. 类型系统

### 6.1 核心类型定义

[`src/types/index.ts`](file:///d:/g/ixpay-pro/web/src/types/index.ts) 定义了通用类型：

```typescript
// API 响应类型
interface ApiResponse<T = unknown> {
  code: number
  data?: T
  msg?: string
}

// 用户信息类型
interface UserInfo {
  uuid: string
  nickName: string
  headerImg: string
  authority: Record<string, unknown>
  // ... 其他字段
}

// 登录信息类型
interface LoginInfo {
  userName: string
  password: string
  captcha?: string
}
```

### 6.2 扩展类型声明

项目扩展了 Vue 的类型声明：

```typescript
declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $IXPAY_PRO: {
      appName: string
      appLogo: string
      showViteLogo: boolean
      logs: unknown[]
    }
  }
}
```

## 7. 构建配置

### 7.1 Vite 配置

[`vite.config.mts`](file:///d:/g/ixpay-pro/web/vite.config.mts) 关键配置：

#### 7.1.1 路径别名

```typescript
resolve: {
  alias: {
    '@': '/src',
    '@/app': '/src/app',
    '@/assets': '/src/assets',
    '@/components': '/src/components',
    '@/stores': '/src/stores',
    '@/utils': '/src/utils',
    '@/types': '/src/types',
    // ...
  }
}
```

#### 7.1.2 代码分割

```typescript
rollupOptions: {
  output: {
    manualChunks: {
      'vue-vendor': ['vue', 'vue-router', 'pinia'],
      'element-plus': ['element-plus', '@element-plus/icons-vue'],
      'axios': ['axios']
    }
  }
}
```

#### 7.1.3 生产优化

```typescript
build: {
  minify: 'terser',
  terserOptions: {
    compress: {
      drop_console: true,      // 移除 console
      drop_debugger: true      // 移除 debugger
    }
  }
}
```

### 7.2 环境变量

#### 7.2.1 开发环境

```env
VITE_CLI_PORT = 8585
VITE_SERVER_PORT = 8586
VITE_BASE_API = /api/admin
VITE_BASE_PATH = http://127.0.0.1
```

#### 7.2.2 生产环境

```env
VITE_BASE_API = /api/admin
VITE_BASE_PATH = /
```

### 7.3 代理配置

```typescript
server: {
  proxy: {
    '/api/admin': {
      target: 'http://127.0.0.1:8586/',
      changeOrigin: true,
      rewrite: (path) => path  // 不移除前缀
    }
  }
}
```

## 8. 开发规范

### 8.1 代码风格

- 使用 TypeScript 编写代码
- 使用 Composition API（`<script setup>`）
- 遵循 Vue 3 最佳实践
- 使用 ESLint + Prettier 统一代码格式

### 8.2 命名规范

- 文件命名：kebab-case（如 `user-list.vue`）
- 组件命名：PascalCase（如 `UserList`）
- 变量命名：camelCase（如 `userInfo`）
- 常量命名：UPPER_SNAKE_CASE（如 `API_BASE_URL`）

### 8.3 注释规范

- 关键逻辑必须添加中文注释
- 函数需要 JSDoc 注释
- 复杂算法需要说明思路

## 9. 性能优化

### 9.1 路由懒加载

所有路由组件都采用动态导入：

```typescript
component: () => import('@/views/system/user/index.vue')
```

### 9.2 组件缓存

通过 keep-alive 缓存组件：

```typescript
// router store 管理 keepAliveRouters
// TabManager 组件实现缓存逻辑
```

### 9.3 代码分割

通过 Vite 的 `manualChunks` 配置：

- 分离 Vue 核心库
- 分离 Element Plus
- 分离 Axios

### 9.4 请求优化

- 请求防抖（Loading 延迟 400ms 显示）
- Token 主动刷新机制
- 请求重试机制

## 10. 安全机制

### 10.1 认证安全

- JWT Token 认证
- Token 双令牌机制（access_token + refresh_token）
- Token 自动刷新
- Token 过期自动跳转登录

### 10.2 权限安全

- 路由级权限控制
- 菜单级权限过滤
- 按钮级权限控制
- 权限指令验证

### 10.3 数据安全

- 请求头自动添加 Authorization
- 敏感操作二次确认
- 水印功能（显示用户信息）

## 11. 开发指南

### 11.1 项目启动

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run serve

# 类型检查
npm run type-check

# 代码检查
npm run lint

# 格式化代码
npm run format
```

### 11.2 生产构建

```bash
# 构建生产版本
npm run build

# 预览构建结果
npm run preview
```

### 11.3 新增页面

1. 在 `src/views/` 下创建页面组件
2. 在后端菜单管理中添加菜单
3. 页面会自动加载（动态路由）

### 11.4 新增 API

1. 在 `src/api/modules/` 下创建模块文件
2. 引入 `service` 实例
3. 导出 API 函数

```typescript
import service from '@/utils/request'

export const getData = (params) => {
  return service({
    url: '/module/api',
    method: 'get',
    params: params,
  })
}
```

### 11.5 新增 Store

1. 在 `src/stores/modules/` 下创建 store 文件
2. 使用 `defineStore` 定义
3. 在 `src/stores/index.ts` 中导出

## 12. 常见问题

### 12.1 动态路由不生效

**原因**：
- asyncRouterFlag 未正确更新
- 菜单 API 返回数据格式错误
- 路由组件路径错误

**解决方案**：
- 检查 `routerStore.SetAsyncRouter()` 是否正常调用
- 检查后端返回的菜单数据结构
- 检查组件路径是否正确（views/ 或 plugin/ 开头）

### 12.2 Token 刷新失败

**原因**：
- refreshToken 过期
- 刷新接口调用失败
- 网络问题

**解决方案**：
- 清除本地存储，重新登录
- 检查后端刷新接口是否正常
- 检查网络连接

### 12.3 页面缓存不生效

**原因**：
- keep-alive 配置错误
- 组件 name 不匹配
- 路由配置问题

**解决方案**：
- 检查 router store 的 keepAliveRouters
- 确保组件有 name 属性
- 检查路由 meta.keepAlive 配置

## 13. 技术选型理由

### 13.1 为什么选择 Vue 3

- Composition API 提供更好的代码组织
- 更好的 TypeScript 支持
- 性能优于 Vue 2
- 活跃的社区和生态

### 13.2 为什么选择 Vite

- 极速的开发服务器启动
- 热模块替换（HMR）
- 开箱即用的 TypeScript 支持
- 优化的生产构建

### 13.3 为什么选择 Pinia

- 比 Vuex 更简洁的 API
- 完整的 TypeScript 支持
- 更好的 DevTools 集成
- 模块化设计

### 13.4 为什么选择 Element Plus

- 丰富的组件库
- 完善的文档
- 活跃的维护
- 良好的 TypeScript 支持

## 14. 未来规划

### 14.1 技术升级

- 跟进 Vue 3 最新版本
- 升级 Vite 到最新版本
- 优化构建性能
- 提升类型安全

### 14.2 功能增强

- 增加更多业务组件
- 完善权限控制粒度
- 优化移动端适配
- 增加国际化支持

### 14.3 性能优化

- 进一步优化首屏加载
- 优化大列表渲染
- 优化资源加载
- 增加 PWA 支持

## 15. 附录

### 15.1 相关文档

- [Vue 3 官方文档](https://vuejs.org/)
- [Vite 官方文档](https://vitejs.dev/)
- [Element Plus 文档](https://element-plus.org/)
- [Pinia 文档](https://pinia.vuejs.org/)

### 15.2 项目规范文档

- [Go API 设计规范](../../../common/.trae/rules/go API 设计规范.md)
- [Go 代码风格规范](../../../common/.trae/rules/go 代码风格规范.md)
- [Vue API 调用规范](../../../common/.trae/rules/vue API 调用规范.md)
- [Vue 代码风格规范](../../../common/.trae/rules/vue 代码风格规范.md)
- [Vue 组件开发规范](../../../common/.trae/rules/vue 组件开发规范.md)

---

**文档版本**: v1.0  
**最后更新**: 2026-04-01  
**维护者**: ixpay-pro 团队
