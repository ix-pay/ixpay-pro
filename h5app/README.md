# ixpay H5 微信端应用

基于 Vue3 + TypeScript + Vite 构建的微信 H5 支付应用。

## 项目结构

```
h5app/
├── .env.development          # 开发环境变量
├── .env.production           # 生产环境变量
├── index.html                # HTML 入口
├── package.json              # 依赖管理
├── tsconfig.json             # TypeScript 配置
├── tsconfig.node.json        # Node 环境 TS 配置
├── vite.config.ts            # Vite 构建配置
├── src/
│   ├── main.ts               # 应用入口
│   ├── App.vue               # 根组件（含全局样式）
│   ├── env.d.ts              # 类型声明（jweixin、WeixinJSBridge）
│   ├── api/
│   │   └── index.ts          # API 接口模块（axios 封装）
│   ├── router/
│   │   └── index.ts          # 路由配置
│   └── pages/
│       ├── login/
│       │   └── index.vue     # 微信登录页面
│       └── payment/
│           └── index.vue     # 微信支付页面
└── README.md
```

## 快速开始

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览构建结果
npm run preview
```

## 环境变量

| 变量名 | 说明 | 默认值（开发） | 默认值（生产） |
|--------|------|---------------|---------------|
| `VITE_API_BASE_URL` | API 基础地址 | http://localhost:8080 | https://api.ixpay.com |
| `VITE_APP_TITLE` | 应用标题 | ixpay-开发环境 | ixpay-微信支付 |

## 页面路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/` | - | 自动重定向到 `/login` |
| `/login` | 微信登录 | 微信 OAuth 登录页 |
| `/payment` | 微信支付 | 订单支付页 |

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/wx/auth/login` | POST | 微信登录（code 换 token） |
| `/api/wx/pay/unified-order` | POST | 创建支付订单 |
| `/api/wx/pay/notify` | POST | 支付通知回调 |

## 技术栈

- **Vue 3** - 前端框架（Composition API）
- **TypeScript** - 类型安全
- **Vite 5** - 构建工具
- **Vue Router 4** - 路由管理
- **Axios** - HTTP 请求
- **jweixin** - 微信 JS-SDK
- **WeixinJSBridge** - 微信 H5 支付（兜底方案）