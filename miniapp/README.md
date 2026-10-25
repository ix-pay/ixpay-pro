# 微信小程序端应用

ixpay 微信小程序端：点击授权登录 → 待支付列表 → 待支付项目明细 → 下单支付（微信支付 APIv3）。

## 页面结构

| 页面 | 路径 | 说明 |
|---|---|---|
| 登录 | `pages/login/index` | 点击「微信授权登录」→ `wx.login` 取 js_code → `POST /api/wx/auth/miniprogram-login` |
| 待支付列表 | `pages/pending-list/index` | `GET /api/wx/payment?status=pending`，支持下拉刷新、新增、取消 |
| 支付明细 | `pages/payment-detail/index` | `GET /api/wx/payment/:id` 查明细；「立即支付」→ `POST /api/wx/payment/unified-order` → `wx.requestPayment` 调起支付 → `?sync=true` 查单确认结果 |

## 配置

1. 修改 `utils/config.js` 中 `baseURL` 为后端地址（默认 `http://localhost:8081`）。
2. 开发工具中「详情 → 本地设置」勾选**不校验合法域名**（线上需 HTTPS 并在小程序后台配置 request 合法域名）。
3. AppID 在 `project.config.json`（当前 `wxd0f3a22e75e60564`）。

## 后端配套

- 后端 `server/config.yaml` 的 `wechat` 段需配置：`mini_app_id`/`mini_app_secret`（小程序登录与小程序下单）、`app_id`/`app_secret`（服务号 H5）、`mch_id`/`mch_serial_no`/`apiv3_key`/`private_key_path`/`public_key_id`/`public_key_path`（APIv3 支付）、`notify_url`（公网 HTTPS 回调）。
- 小程序支付要求：小程序 AppID 与商户号绑定，且下单 `openid` 为小程序 openid（后端按登录渠道自动选择 appid + openid）。
