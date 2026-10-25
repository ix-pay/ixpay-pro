-- =====================================================================
-- 微信渠道配置 + 对账模块 菜单注册 SQL（管理平台端）
-- 执行方式二选一：
--   方式A：管理平台「系统管理 → 菜单管理」界面添加（推荐，见下方参数说明）
--   方式B：直接执行本 SQL（ID 为预留段，避免与雪花ID冲突）
-- 注意：菜单 component 填 wx/channel/index、wx/recon/index，
--       前端动态路由自动解析为 src/views/wx/channel/index.vue 与 src/views/wx/recon/index.vue
-- =====================================================================

-- 一级目录：微信支付（type=1 目录）
INSERT INTO base_menus (id, parent_id, path, name, component, title, icon, hidden, sort, status, type, keep_alive, default_menu, breadcrumb)
VALUES (900000000000000001, 0, '/wx', 'WxPayRoot', NULL, '微信支付', 'Wallet', false, 50, 1, 1, false, false, true)
ON CONFLICT (name) DO NOTHING;

-- 二级菜单：渠道配置（type=2 页面）
INSERT INTO base_menus (id, parent_id, path, name, component, title, icon, hidden, sort, status, type, keep_alive, default_menu, breadcrumb)
VALUES (900000000000000002, 900000000000000001, 'channel', 'WxChannel', 'wx/channel/index', '渠道配置', 'Setting', false, 1, 1, 2, false, false, true)
ON CONFLICT (name) DO NOTHING;

-- 二级菜单：对账中心（type=2 页面）
INSERT INTO base_menus (id, parent_id, path, name, component, title, icon, hidden, sort, status, type, keep_alive, default_menu, breadcrumb)
VALUES (900000000000000003, 900000000000000001, 'recon', 'WxRecon', 'wx/recon/index', '对账中心', 'DataAnalysis', false, 2, 1, 2, false, false, true)
ON CONFLICT (name) DO NOTHING;

-- 菜单管理界面添加参数对照：
--   渠道配置：路由名 WxChannel | 路由路径 channel | 组件路径 wx/channel/index | 标题 渠道配置
--   对账中心：路由名 WxRecon   | 路由路径 recon   | 组件路径 wx/recon/index   | 标题 对账中心
-- 同时在「API 管理」与「角色权限」中放行 /api/wx/admin/* 接口（GET/POST/PUT/DELETE）
