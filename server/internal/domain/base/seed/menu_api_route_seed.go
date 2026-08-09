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
		"UserManagement": {
			{Path: "/api/admin/user", Method: "GET"},
			{Path: "/api/admin/user", Method: "POST"},
			{Path: "/api/admin/user/:id", Method: "DELETE"},
			{Path: "/api/admin/user/reset-password", Method: "PUT"},
			{Path: "/api/admin/user/password", Method: "PUT"},
			{Path: "/api/admin/user/info", Method: "GET"},
			{Path: "/api/admin/user/info", Method: "PUT"},
			{Path: "/api/admin/user/switch-role", Method: "POST"},
			{Path: "/api/admin/user/get-user-settings", Method: "GET"},
			{Path: "/api/admin/user/update-user-settings", Method: "PUT"},
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
			{Path: "/api/admin/user", Method: "GET"},
			{Path: "/api/admin/user/info", Method: "GET"},
		},

		// ==================== 角色管理菜单 ====================
		"RoleManagement": {
			{Path: "/api/admin/role", Method: "GET"},
			{Path: "/api/admin/role", Method: "POST"},
			{Path: "/api/admin/role/:id", Method: "GET"},
			{Path: "/api/admin/role/:id", Method: "PUT"},
			{Path: "/api/admin/role/:id", Method: "DELETE"},
			{Path: "/api/admin/role/all", Method: "GET"},
			
			{Path: "/api/admin/role/assign-users", Method: "POST"},
			{Path: "/api/admin/role/assign-menus", Method: "POST"},
			{Path: "/api/admin/role/assign-api-routes", Method: "POST"},
			{Path: "/api/admin/role/:id/permissions", Method: "POST"},
			{Path: "/api/admin/role/:id/available-menus", Method: "GET"},
		},
		"RoleAdd": {
			{Path: "/api/admin/role", Method: "POST"},
		},
		"RoleEdit": {
			{Path: "/api/admin/role/:id", Method: "PUT"},
		},
		"RoleDelete": {
			{Path: "/api/admin/role/:id", Method: "DELETE"},
		},
		"RoleAssign": {
			{Path: "/api/admin/role/:id/permissions", Method: "POST"},
			{Path: "/api/admin/role/assign-menus", Method: "POST"},
			{Path: "/api/admin/role/assign-api-routes", Method: "POST"},
			{Path: "/api/admin/role/assign-users", Method: "POST"},
		},

		// ==================== 菜单管理按钮 ====================
		"MenuAdd": {
			{Path: "/api/admin/menu", Method: "POST"},
		},
		"MenuEdit": {
			{Path: "/api/admin/menu", Method: "PUT"},
		},
		"MenuDelete": {
			{Path: "/api/admin/menu/:id", Method: "DELETE"},
		},

		// ==================== 部门管理按钮 ====================
		"DeptAdd": {
			{Path: "/api/admin/dept", Method: "POST"},
		},
		"DeptEdit": {
			{Path: "/api/admin/dept/:id", Method: "PUT"},
		},
		"DeptDelete": {
			{Path: "/api/admin/dept/:id", Method: "DELETE"},
		},

		// ==================== 岗位管理按钮 ====================
		"PositionAdd": {
			{Path: "/api/admin/position", Method: "POST"},
		},
		"PositionEdit": {
			{Path: "/api/admin/position/:id", Method: "PUT"},
		},
		"PositionDelete": {
			{Path: "/api/admin/position/:id", Method: "DELETE"},
		},

		// ==================== 字典管理菜单 ====================
		"DictManagement": {
			{Path: "/api/admin/dict", Method: "GET"},
			{Path: "/api/admin/dict", Method: "POST"},
			{Path: "/api/admin/dict/:id", Method: "GET"},
			{Path: "/api/admin/dict/:id", Method: "PUT"},
			{Path: "/api/admin/dict/:id", Method: "DELETE"},
			{Path: "/api/admin/dict/item", Method: "POST"},
			{Path: "/api/admin/dict/item/:id", Method: "GET"},
			{Path: "/api/admin/dict/item/:id", Method: "PUT"},
			{Path: "/api/admin/dict/item/:id", Method: "DELETE"},
			{Path: "/api/admin/dict/items", Method: "GET"},
			{Path: "/api/admin/dict/code/:code/active-items", Method: "GET"},
		},
		"DictAdd": {
			{Path: "/api/admin/dict", Method: "POST"},
		},
		"DictEdit": {
			{Path: "/api/admin/dict/:id", Method: "PUT"},
		},
		"DictDelete": {
			{Path: "/api/admin/dict/:id", Method: "DELETE"},
		},
		"DictItemAdd": {
			{Path: "/api/admin/dict/item", Method: "POST"},
		},
		"DictItemEdit": {
			{Path: "/api/admin/dict/item/:id", Method: "PUT"},
		},
		"DictItemDelete": {
			{Path: "/api/admin/dict/item/:id", Method: "DELETE"},
		},

		// ==================== 系统配置按钮 ====================
		"ConfigAdd": {
			{Path: "/api/admin/config", Method: "POST"},
		},
		"ConfigEdit": {
			{Path: "/api/admin/config", Method: "PUT"},
		},
		"ConfigDelete": {
			{Path: "/api/admin/config/:id", Method: "DELETE"},
		},

		// ==================== 公告管理按钮 ====================
		"NoticeAdd": {
			{Path: "/api/admin/notices", Method: "POST"},
		},
		"NoticeEdit": {
			{Path: "/api/admin/notices/:id", Method: "PUT"},
		},
		"NoticeDelete": {
			{Path: "/api/admin/notices/:id", Method: "DELETE"},
		},

		// ==================== API 管理按钮 ====================
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
		"ScheduledTask": {
			{Path: "/api/admin/task", Method: "GET"},
			{Path: "/api/admin/task", Method: "POST"},
			{Path: "/api/admin/task/:id", Method: "GET"},
			{Path: "/api/admin/task/:id", Method: "PUT"},
			{Path: "/api/admin/task/:id", Method: "DELETE"},
			{Path: "/api/admin/task/:id/enable", Method: "POST"},
			{Path: "/api/admin/task/:id/disable", Method: "POST"},
			{Path: "/api/admin/task/:id/start", Method: "POST"},
			{Path: "/api/admin/task/:id/stop", Method: "POST"},
			{Path: "/api/admin/task/:id/retry", Method: "POST"},
			{Path: "/api/admin/task/execution-logs", Method: "GET"},
			{Path: "/api/admin/task/:id/execution-logs", Method: "GET"},
			{Path: "/api/admin/task/statistics", Method: "GET"},
			{Path: "/api/admin/task/dashboard", Method: "GET"},
			{Path: "/api/admin/task/:id/group", Method: "POST"},
		},
		"TaskAdd": {
			{Path: "/api/admin/task", Method: "POST"},
		},
		"TaskEdit": {
			{Path: "/api/admin/task/:id", Method: "PUT"},
		},
		"TaskDelete": {
			{Path: "/api/admin/task/:id", Method: "DELETE"},
		},
		"TaskExecute": {
			{Path: "/api/admin/task/:id/retry", Method: "POST"},
		},
		"TaskEnable": {
			{Path: "/api/admin/task/:id/enable", Method: "POST"},
			{Path: "/api/admin/task/:id/start", Method: "POST"},
		},
		"TaskDisable": {
			{Path: "/api/admin/task/:id/disable", Method: "POST"},
			{Path: "/api/admin/task/:id/stop", Method: "POST"},
		},

		// ==================== 操作日志按钮 ====================
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

		// ==================== 登录日志按钮 ====================
		"LoginLogView": {
			{Path: "/api/admin/login-log/:id", Method: "GET"},
		},
		"LoginLogClear": {
			{Path: "/api/admin/login-log/clear", Method: "POST"},
		},

		// ==================== 部门管理菜单 ====================
		"DepartmentManagement": {
			{Path: "/api/admin/dept", Method: "GET"},
			{Path: "/api/admin/dept", Method: "POST"},
			{Path: "/api/admin/dept/:id", Method: "GET"},
			{Path: "/api/admin/dept/:id", Method: "PUT"},
			{Path: "/api/admin/dept/:id", Method: "DELETE"},
			{Path: "/api/admin/dept/tree", Method: "GET"},
			{Path: "/api/admin/dept/:id/leader", Method: "PUT"},
		},

		// ==================== 菜单管理菜单 ====================
		"MenuManagement": {
			{Path: "/api/admin/menu", Method: "GET"},
			{Path: "/api/admin/menu", Method: "POST"},
			{Path: "/api/admin/menu", Method: "PUT"},
			{Path: "/api/admin/menu/:id", Method: "GET"},
			{Path: "/api/admin/menu/:id", Method: "DELETE"},
			{Path: "/api/admin/menu/tree", Method: "GET"},
			{Path: "/api/admin/menu/page", Method: "GET"},
			{Path: "/api/admin/menu/:id/delete-impact", Method: "GET"},
		},

		// ==================== 系统配置菜单 ====================
		"SystemConfig": {
			{Path: "/api/admin/config", Method: "GET"},
			{Path: "/api/admin/config", Method: "POST"},
			{Path: "/api/admin/config/:id", Method: "GET"},
			{Path: "/api/admin/config", Method: "PUT"},
			{Path: "/api/admin/config/:id", Method: "DELETE"},
			{Path: "/api/admin/config/key", Method: "GET"},
			{Path: "/api/admin/config/active", Method: "GET"},
		},

		// ==================== API 管理菜单 ====================
		"ApiRouteManagement": {
			{Path: "/api/admin/apis", Method: "GET"},
			{Path: "/api/admin/apis", Method: "POST"},
			{Path: "/api/admin/apis/:id", Method: "GET"},
			{Path: "/api/admin/apis/:id", Method: "PUT"},
			{Path: "/api/admin/apis/:id", Method: "DELETE"},
		},

		// ==================== 公告管理菜单 ====================
		"NoticeManagement": {
			{Path: "/api/admin/notices", Method: "GET"},
			{Path: "/api/admin/notices", Method: "POST"},
			{Path: "/api/admin/notices/:id", Method: "GET"},
			{Path: "/api/admin/notices/:id", Method: "PUT"},
			{Path: "/api/admin/notices/:id", Method: "DELETE"},
			{Path: "/api/admin/notices/:id/publish", Method: "POST"},
			{Path: "/api/admin/notices/:id/read", Method: "POST"},
			{Path: "/api/admin/notices/:id/is-read", Method: "GET"},
			{Path: "/api/admin/notices/statistics", Method: "GET"},
		},

		// ==================== 操作日志菜单 ====================
		"OperationLogManagement": {
			{Path: "/api/admin/logs", Method: "GET"},
			{Path: "/api/admin/logs/:id", Method: "GET"},
			{Path: "/api/admin/logs/:id", Method: "DELETE"},
			{Path: "/api/admin/logs/batch-delete", Method: "POST"},
			{Path: "/api/admin/logs/statistics", Method: "GET"},
			{Path: "/api/admin/logs/clear", Method: "POST"},
		},

		// ==================== 登录日志菜单 ====================
		"LoginLogManagement": {
			{Path: "/api/admin/login-log", Method: "GET"},
			{Path: "/api/admin/login-log/:id", Method: "GET"},
			{Path: "/api/admin/login-log/statistics", Method: "GET"},
			{Path: "/api/admin/login-log/abnormal", Method: "GET"},
			{Path: "/api/admin/login-log/batch-delete", Method: "POST"},
			{Path: "/api/admin/login-log/clear", Method: "POST"},
		},

		// ==================== 系统监控菜单 ====================
		"Monitor": {
			{Path: "/api/admin/monitor/system", Method: "GET"},
			{Path: "/api/admin/monitor/cache", Method: "GET"},
			{Path: "/api/admin/monitor/database", Method: "GET"},
			{Path: "/api/admin/monitor/redis-keys", Method: "GET"},
			{Path: "/api/admin/monitor/slow-queries", Method: "GET"},
		},

		// ==================== 在线用户菜单 ====================
		"OnlineUserManagement": {
			{Path: "/api/admin/online-user", Method: "GET"},
			{Path: "/api/admin/online-user/:user_id", Method: "GET"},
			{Path: "/api/admin/online-user/:user_id", Method: "DELETE"},
			{Path: "/api/admin/online-user/count", Method: "GET"},
			{Path: "/api/admin/online-user/online", Method: "GET"},
			{Path: "/api/admin/online-user/batch", Method: "POST"},
		},

		// ==================== 网关服务菜单 ====================
		"GatewayServices": {
			{Path: "/api/admin/gateway/services", Method: "GET"},
		},

		// ==================== 任务统计菜单 ====================
		"TaskStatistics": {
			{Path: "/api/admin/task/statistics", Method: "GET"},
			{Path: "/api/admin/task/dashboard", Method: "GET"},
		},

		// ==================== 任务执行日志菜单 ====================
		"TaskExecutionLogs": {
			{Path: "/api/admin/task/execution-logs", Method: "GET"},
			{Path: "/api/admin/task/:id/execution-logs", Method: "GET"},
		},

		// ==================== 节点管理菜单 ====================
		"TaskNodes": {
			{Path: "/api/admin/nodes", Method: "GET"},
			{Path: "/api/admin/nodes/:id", Method: "GET"},
			{Path: "/api/admin/nodes/:id/offline", Method: "POST"},
			{Path: "/api/admin/nodes/statistics", Method: "GET"},
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