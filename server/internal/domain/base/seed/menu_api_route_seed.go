package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// MenuAPIRouteSeed 菜单-API 关联种子数据
// 建立菜单（type=2）和按钮（type=3）与 API 路由的关联关系
// 当给角色授权菜单或按钮时，自动拥有关联的 API 权限
type MenuAPIRouteSeed struct {
	menuRepo repo.MenuRepository
	apiRepo  repo.APIRepository
}

// NewMenuAPIRouteSeed 创建菜单-API 关联种子数据实例
func NewMenuAPIRouteSeed(menuRepo repo.MenuRepository, apiRepo repo.APIRepository) Seed {
	return &MenuAPIRouteSeed{
		menuRepo: menuRepo,
		apiRepo:  apiRepo,
	}
}

// Version 返回种子数据版本
func (ms *MenuAPIRouteSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (ms *MenuAPIRouteSeed) Name() string {
	return "menu_api_route_seed"
}

// Order 返回初始化顺序（第八个执行，在菜单和 API 种子之后）
func (ms *MenuAPIRouteSeed) Order() int {
	return 8
}

// Init 初始化菜单-API 关联种子数据
func (ms *MenuAPIRouteSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化菜单-API 关联种子数据")

	// 获取所有菜单
	allMenus, err := ms.menuRepo.GetAll()
	if err != nil {
		logger.Error("获取所有菜单失败", "error", err)
		return err
	}

	// 构建菜单名称到菜单的映射
	menuMap := make(map[string]*entity.Menu)
	for _, m := range allMenus {
		menuMap[m.Name] = m
	}

	// 定义菜单与 API 的关联关系
	// key: 菜单名称(Name), value: API 路径和方法列表
	menuAPIRoutes := map[string][]struct {
		Path   string
		Method string
	}{
		// ==================== 用户管理菜单 ====================
		// 页面加载时获取列表数据、角色/部门/岗位下拉数据
		"UserManagement": {
			{Path: "/api/admin/user", Method: "GET"},
			{Path: "/api/admin/role/all", Method: "GET"},
			{Path: "/api/admin/dept", Method: "GET"},
			{Path: "/api/admin/position", Method: "GET"},
		},
		"UserAdd": {
			{Path: "/api/admin/user", Method: "POST"},
		},
		"UserEdit": {
			{Path: "/api/admin/user/info", Method: "PUT"},
		},
		"UserDelete": {
			{Path: "/api/admin/user/:id", Method: "DELETE"},
		},
		"UserView": {
			{Path: "/api/admin/user/reset-password", Method: "PUT"},
		},

		// ==================== 角色管理菜单 ====================
		// 页面加载时获取角色列表数据
		"RoleManagement": {
			{Path: "/api/admin/role", Method: "GET"},
		},
		"RoleAdd": {
			{Path: "/api/admin/role", Method: "POST"},
		},
		"RoleEdit": {
			{Path: "/api/admin/role", Method: "PUT"},
		},
		"RoleDelete": {
			{Path: "/api/admin/role", Method: "DELETE"},
		},
		"RoleAssign": {
			{Path: "/api/admin/menu/tree", Method: "GET"},
			{Path: "/api/admin/role/:id/detail", Method: "GET"},
			{Path: "/api/admin/role/:id/available-apis", Method: "GET"},
			{Path: "/api/admin/role/:id/permissions", Method: "POST"},
		},

		// ==================== 菜单管理菜单 ====================
		// 页面加载时获取菜单树数据
		"MenuManagement": {
			{Path: "/api/admin/menu/tree", Method: "GET"},
		},
		"MenuAdd": {
			{Path: "/api/admin/menu", Method: "POST"},
			{Path: "/api/admin/apis", Method: "GET"},
		},
		"MenuEdit": {
			{Path: "/api/admin/menu", Method: "PUT"},
			{Path: "/api/admin/apis", Method: "GET"},
		},
		"MenuDelete": {
			{Path: "/api/admin/menu/:id", Method: "DELETE"},
		},

		// ==================== 部门管理菜单 ====================
		// 页面加载时获取部门列表数据
		"DepartmentManagement": {
			{Path: "/api/admin/dept", Method: "GET"},
		},
		"DeptAdd": {
			{Path: "/api/admin/dept", Method: "POST"},
		},
		"DeptEdit": {
			{Path: "/api/admin/dept", Method: "PUT"},
		},
		"DeptDelete": {
			{Path: "/api/admin/dept/:id", Method: "DELETE"},
		},

		// ==================== 岗位管理菜单 ====================
		// 页面加载时获取岗位列表数据
		"PositionManagement": {
			{Path: "/api/admin/position", Method: "GET"},
		},
		"PositionAdd": {
			{Path: "/api/admin/position", Method: "POST"},
		},
		"PositionEdit": {
			{Path: "/api/admin/position", Method: "PUT"},
		},
		"PositionDelete": {
			{Path: "/api/admin/position/:id", Method: "DELETE"},
		},

		// ==================== 字典管理菜单 ====================
		// 页面加载时获取字典列表数据
		"DictManagement": {
			{Path: "/api/admin/dict", Method: "GET"},
		},
		"DictAdd": {
			{Path: "/api/admin/dict", Method: "POST"},
		},
		"DictEdit": {
			{Path: "/api/admin/dict", Method: "PUT"},
		},
		"DictDelete": {
			{Path: "/api/admin/dict/:id", Method: "DELETE"},
		},
		"DictItemAdd": {
			{Path: "/api/admin/dict/item", Method: "POST"},
		},
		"DictItemEdit": {
			{Path: "/api/admin/dict/item", Method: "PUT"},
		},
		"DictItemDelete": {
			{Path: "/api/admin/dict/item/:id", Method: "DELETE"},
		},

		// ==================== 系统配置菜单 ====================
		// 页面加载时获取配置列表数据
		"SystemConfig": {
			{Path: "/api/admin/config", Method: "GET"},
		},
		"ConfigAdd": {
			{Path: "/api/admin/config", Method: "POST"},
		},
		"ConfigEdit": {
			{Path: "/api/admin/config", Method: "PUT"},
		},
		"ConfigDelete": {
			{Path: "/api/admin/config/:id", Method: "DELETE"},
		},

		// ==================== 公告管理菜单 ====================
		// 页面加载时获取公告列表数据
		"NoticeManagement": {
			{Path: "/api/admin/notices", Method: "GET"},
		},
		"NoticeAdd": {
			{Path: "/api/admin/notices", Method: "POST"},
		},
		"NoticeEdit": {
			{Path: "/api/admin/notices", Method: "PUT"},
		},
		"NoticeDelete": {
			{Path: "/api/admin/notices/:id", Method: "DELETE"},
		},

		// ==================== API 管理菜单 ====================
		// 页面加载时获取API路由列表数据
		"ApiRouteManagement": {
			{Path: "/api/admin/apis", Method: "GET"},
		},
		"ApiRouteAdd": {
			{Path: "/api/admin/apis", Method: "POST"},
		},
		"ApiRouteEdit": {
			{Path: "/api/admin/apis/:id", Method: "PUT"},
		},
		"ApiRouteDelete": {
			{Path: "/api/admin/apis/:id", Method: "DELETE"},
		},

		// ==================== 定时任务管理菜单 ====================
		// 页面加载时获取任务列表和仪表盘统计数据
		"ScheduledTask": {
			{Path: "/api/admin/task", Method: "GET"},
			{Path: "/api/admin/task/dashboard", Method: "GET"},
		},
		"TaskAdd": {
			{Path: "/api/admin/task", Method: "POST"},
		},
		"TaskEdit": {
			{Path: "/api/admin/task/:id", Method: "GET"},
			{Path: "/api/admin/task/:id", Method: "PUT"},
		},
		"TaskDelete": {
			{Path: "/api/admin/task/:id", Method: "DELETE"},
		},
		"TaskExecute": {
			{Path: "/api/admin/task/:id/start", Method: "POST"},
		},
		"TaskEnable": {
			{Path: "/api/admin/task/:id/enable", Method: "POST"},
		},
		"TaskDisable": {
			{Path: "/api/admin/task/:id/disable", Method: "POST"},
		},

		// ==================== 操作日志菜单 ====================
		// 页面加载时获取操作日志列表数据
		"OperationLogManagement": {
			{Path: "/api/admin/logs", Method: "GET"},
		},
		"OpLogView": {
			{Path: "/api/admin/logs/:id", Method: "GET"},
		},
		"OpLogDelete": {
			{Path: "/api/admin/logs/:id", Method: "DELETE"},
		},
		"OpLogBatchDelete": {
			{Path: "/api/admin/logs/batch-delete", Method: "POST"},
		},
		"OpLogClear": {
			{Path: "/api/admin/logs/clear", Method: "POST"},
		},

		// ==================== 登录日志菜单 ====================
		// 页面加载时获取登录日志列表数据
		"LoginLogManagement": {
			{Path: "/api/admin/login-log", Method: "GET"},
		},
		"LoginLogView": {
			{Path: "/api/admin/login-log/:id", Method: "GET"},
		},
		"LoginLogClear": {
			{Path: "/api/admin/login-log/clear", Method: "POST"},
		},

		// ==================== 系统监控菜单 ====================
		// 页面加载时获取系统监控、数据库监控、缓存监控数据
		"Monitor": {
			{Path: "/api/admin/monitor/system", Method: "GET"},
			{Path: "/api/admin/monitor/database", Method: "GET"},
			{Path: "/api/admin/monitor/cache", Method: "GET"},
		},

		// ==================== 在线用户菜单 ====================
		// 页面加载时获取在线用户列表数据
		"OnlineUserManagement": {
			{Path: "/api/admin/online-user", Method: "GET"},
		},

		// ==================== 网关服务菜单 ====================
		// 页面加载时获取网关服务列表数据
		"GatewayServices": {
			{Path: "/api/admin/gateway/services", Method: "GET"},
		},

		// ==================== 任务统计菜单 ====================
		// 页面加载时获取任务统计数据和任务列表
		"TaskStatistics": {
			{Path: "/api/admin/task/statistics", Method: "GET"},
			{Path: "/api/admin/task", Method: "GET"},
		},

		// ==================== 任务执行日志菜单 ====================
		// 页面加载时获取任务执行日志列表和任务列表
		"TaskExecutionLogs": {
			{Path: "/api/admin/task/execution-logs", Method: "GET"},
			{Path: "/api/admin/task", Method: "GET"},
		},

		// ==================== 节点管理菜单 ====================
		// 页面加载时获取节点列表数据
		"TaskNodes": {
			{Path: "/api/admin/nodes", Method: "GET"},
		},
	}

	// 遍历关联关系，建立菜单-API 关联
	for menuName, routes := range menuAPIRoutes {
		menu, ok := menuMap[menuName]
		if !ok {
			logger.Warn("菜单不存在，跳过关联", "menuName", menuName)
			continue
		}

		// 查找 API ID
		apiIDs := make([]int64, 0)
		for _, route := range routes {
			api, err := ms.apiRepo.GetByPathAndMethod(route.Path, route.Method)
			if err != nil {
				logger.Warn("API 路由不存在，跳过关联",
					"menuName", menuName,
					"path", route.Path,
					"method", route.Method,
					"error", err)
				continue
			}
			apiIDs = append(apiIDs, api.ID)
		}

		if len(apiIDs) == 0 {
			logger.Warn("菜单没有找到可关联的 API", "menuName", menuName)
			continue
		}

		// 保存菜单-API 关联
		if err := ms.menuRepo.SaveMenuAPIRoutes(menu.ID, apiIDs); err != nil {
			logger.Error("保存菜单-API 关联失败",
				"menuName", menuName,
				"menuID", menu.ID,
				"apiIDs", apiIDs,
				"error", err)
			return err
		}
		logger.Info("菜单-API 关联成功",
			"menuName", menuName,
			"menuID", menu.ID,
			"apiCount", len(apiIDs))
	}

	logger.Info("菜单-API 关联种子数据初始化完成")
	return nil
}