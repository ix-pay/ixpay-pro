-- =====================================================================
-- 多租户管理 菜单注册 SQL（管理平台端）
-- 执行方式二选一：
--   方式A：管理平台「系统管理 → 菜单管理」界面添加（推荐，见下方参数说明）
--   方式B：直接执行本 SQL（ID 为预留段，避免与雪花ID冲突）
-- 注意：菜单 component 填 base/tenant/index，
--       前端动态路由自动解析为 src/views/base/tenant/index.vue
-- =====================================================================

-- 一级目录：多租户管理（type=1 目录）
INSERT INTO base_menus (id, parent_id, path, name, component, title, icon, hidden, sort, status, type, keep_alive, default_menu, breadcrumb)
VALUES (900000000000000010, 0, '/tenant', 'TenantRoot', NULL, '多租户管理', 'OfficeBuilding', false, 45, 1, 1, false, false, true)
ON CONFLICT (name) DO NOTHING;

-- 二级菜单：租户管理（type=2 页面）
INSERT INTO base_menus (id, parent_id, path, name, component, title, icon, hidden, sort, status, type, keep_alive, default_menu, breadcrumb)
VALUES (900000000000000011, 900000000000000010, 'list', 'TenantManage', 'base/tenant/index', '租户管理', 'Menu', false, 1, 1, 2, false, false, true)
ON CONFLICT (name) DO NOTHING;

-- 菜单管理界面添加参数对照：
--   租户管理：路由名 TenantManage | 路由路径 list | 组件路径 base/tenant/index | 标题 租户管理
-- 同时在「API 管理」与「角色权限」中放行 /api/admin/tenant* 接口（GET/POST/PUT/DELETE）
