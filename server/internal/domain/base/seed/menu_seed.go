package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// MenuSeed 菜单种子数据
type MenuSeed struct {
	menuRepo repo.MenuRepository
	apiRepo  repo.APIRepository
}

// NewMenuSeed 创建菜单种子数据实例
func NewMenuSeed(menuRepo repo.MenuRepository, apiRepo repo.APIRepository) Seed {
	return &MenuSeed{
		menuRepo: menuRepo,
		apiRepo:  apiRepo,
	}
}

// Version 返回种子数据版本
func (ms *MenuSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (ms *MenuSeed) Name() string {
	return "menu_seed"
}

// Order 返回初始化顺序（第七个执行，菜单需要在API路由之后）
func (ms *MenuSeed) Order() int {
	return 7
}

// Init 初始化菜单种子数据
func (ms *MenuSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化菜单种子数据")

	// ==================== 第一层：目录 ====================
	// 首页（作为默认路由）
	_, err := ms.createOrGetMenu(logger, &entity.Menu{
		Path:        "index",
		Name:        "Index",
		Component:   "views/base/index/index",
		Title:       "首页",
		Icon:        "Dashboard",
		Hidden:      false,
		Sort:        0,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: true,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 创建一级目录
	systemDir, err := ms.createOrGetMenu(logger, &entity.Menu{
		Path:        "system",
		Name:        "SystemManagement",
		Component:   "views/system/index",
		Title:       "系统管理",
		Icon:        "Setting",
		Hidden:      false,
		Sort:        1,
		Status:      1,
		KeepAlive:   false,
		DefaultMenu: false,
		Type:        entity.MenuTypeDirectory,
	})
	if err != nil {
		return err
	}

	taskDir, err := ms.createOrGetMenu(logger, &entity.Menu{
		Path:        "task",
		Name:        "TaskManagement",
		Component:   "views/task/index",
		Title:       "任务管理",
		Icon:        "Clock",
		Hidden:      false,
		Sort:        2,
		Status:      1,
		KeepAlive:   false,
		DefaultMenu: false,
		Type:        entity.MenuTypeDirectory,
	})
	if err != nil {
		return err
	}

	monitorDir, err := ms.createOrGetMenu(logger, &entity.Menu{
		Path:        "monitor",
		Name:        "SystemMonitor",
		Component:   "views/monitor/index",
		Title:       "系统监控",
		Icon:        "Monitor",
		Hidden:      false,
		Sort:        3,
		Status:      1,
		KeepAlive:   false,
		DefaultMenu: false,
		Type:        entity.MenuTypeDirectory,
	})
	if err != nil {
		return err
	}

	logDir, err := ms.createOrGetMenu(logger, &entity.Menu{
		Path:        "log",
		Name:        "LogManagement",
		Component:   "views/log/index",
		Title:       "日志管理",
		Icon:        "Document",
		Hidden:      false,
		Sort:        4,
		Status:      1,
		KeepAlive:   false,
		DefaultMenu: false,
		Type:        entity.MenuTypeDirectory,
	})
	if err != nil {
		return err
	}

	// ==================== 第二层：菜单 ====================
	// 系统管理下的菜单
	userMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "user",
		Name:        "UserManagement",
		Component:   "views/system/user/index",
		Title:       "用户管理",
		Icon:        "User",
		Hidden:      false,
		Sort:        1,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	roleMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "role",
		Name:        "RoleManagement",
		Component:   "views/system/role/index",
		Title:       "角色管理",
		Icon:        "Avatar",
		Hidden:      false,
		Sort:        2,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	menuMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "menu",
		Name:        "MenuManagement",
		Component:   "views/system/menu/index",
		Title:       "菜单管理",
		Icon:        "Menu",
		Hidden:      false,
		Sort:        3,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	deptMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "department",
		Name:        "DepartmentManagement",
		Component:   "views/system/department/index",
		Title:       "部门管理",
		Icon:        "OfficeBuilding",
		Hidden:      false,
		Sort:        4,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	positionMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "position",
		Name:        "PositionManagement",
		Component:   "views/system/position/index",
		Title:       "岗位管理",
		Icon:        "Postcard",
		Hidden:      false,
		Sort:        5,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   positionMenu.ID,
		Name:       "PositionAdd",
		Title:      "添加职位",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:position:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   positionMenu.ID,
		Name:       "PositionEdit",
		Title:      "编辑职位",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:position:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   positionMenu.ID,
		Name:       "PositionDelete",
		Title:      "删除职位",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:position:delete",
	})
	if err != nil {
		return err
	}

	dictMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "dict",
		Name:        "DictManagement",
		Component:   "views/system/dict/index",
		Title:       "字典管理",
		Icon:        "Collection",
		Hidden:      false,
		Sort:        6,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   dictMenu.ID,
		Name:       "DictAdd",
		Title:      "添加字典",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:dict:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   dictMenu.ID,
		Name:       "DictEdit",
		Title:      "编辑字典",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:dict:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   dictMenu.ID,
		Name:       "DictDelete",
		Title:      "删除字典",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:dict:delete",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   dictMenu.ID,
		Name:       "DictItemAdd",
		Title:      "添加明细",
		Icon:       "Plus",
		Sort:       4,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:dict:item:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   dictMenu.ID,
		Name:       "DictItemEdit",
		Title:      "编辑明细",
		Icon:       "Edit",
		Sort:       5,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:dict:item:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   dictMenu.ID,
		Name:       "DictItemDelete",
		Title:      "删除明细",
		Icon:       "Delete",
		Sort:       6,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:dict:item:delete",
	})
	if err != nil {
		return err
	}

	configMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "config",
		Name:        "SystemConfig",
		Component:   "views/system/config/index",
		Title:       "系统配置",
		Icon:        "Tools",
		Hidden:      false,
		Sort:        7,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   configMenu.ID,
		Name:       "ConfigAdd",
		Title:      "添加配置",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:config:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   configMenu.ID,
		Name:       "ConfigEdit",
		Title:      "编辑配置",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:config:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   configMenu.ID,
		Name:       "ConfigDelete",
		Title:      "删除配置",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:config:delete",
	})
	if err != nil {
		return err
	}

	noticeMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "notice",
		Name:        "NoticeManagement",
		Component:   "views/system/notice/index",
		Title:       "公告管理",
		Icon:        "Bell",
		Hidden:      false,
		Sort:        8,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   noticeMenu.ID,
		Name:       "NoticeAdd",
		Title:      "添加公告",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:notice:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   noticeMenu.ID,
		Name:       "NoticeEdit",
		Title:      "编辑公告",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:notice:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   noticeMenu.ID,
		Name:       "NoticeDelete",
		Title:      "删除公告",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:notice:delete",
	})
	if err != nil {
		return err
	}

	apiRouteMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    systemDir.ID,
		Path:        "api-route",
		Name:        "ApiRouteManagement",
		Component:   "views/system/api-route/index",
		Title:       "API 管理",
		Icon:        "Link",
		Hidden:      false,
		Sort:        9,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   apiRouteMenu.ID,
		Name:       "ApiRouteAdd",
		Title:      "添加 API",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:api-route:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   apiRouteMenu.ID,
		Name:       "ApiRouteEdit",
		Title:      "编辑 API",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:api-route:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   apiRouteMenu.ID,
		Name:       "ApiRouteDelete",
		Title:      "删除 API",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:api-route:delete",
	})
	if err != nil {
		return err
	}

	// 任务管理下的菜单
	taskMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    taskDir.ID,
		Path:        "task",
		Name:        "ScheduledTask",
		Component:   "views/task/task/index",
		Title:       "定时任务管理",
		Icon:        "Clock",
		Hidden:      false,
		Sort:        1,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 系统监控下的菜单
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    monitorDir.ID,
		Path:        "monitor",
		Name:        "Monitor",
		Component:   "views/monitor/monitor/index",
		Title:       "系统监控",
		Icon:        "Monitor",
		Hidden:      false,
		Sort:        1,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    monitorDir.ID,
		Path:        "online-user",
		Name:        "OnlineUserManagement",
		Component:   "views/monitor/online-user/index",
		Title:       "在线用户",
		Icon:        "UserFilled",
		Hidden:      false,
		Sort:        2,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    monitorDir.ID,
		Path:        "gateway-services",
		Name:        "GatewayServices",
		Component:   "views/monitor/gateway-services/index",
		Title:       "网关服务",
		Icon:        "Connection",
		Hidden:      false,
		Sort:        3,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 日志管理下的菜单
	opLogMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    logDir.ID,
		Path:        "operation-log",
		Name:        "OperationLogManagement",
		Component:   "views/log/operation-log/index",
		Title:       "操作日志",
		Icon:        "Tickets",
		Hidden:      false,
		Sort:        1,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	loginLogMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    logDir.ID,
		Path:        "login-log",
		Name:        "LoginLogManagement",
		Component:   "views/log/login-log/index",
		Title:       "登录日志",
		Icon:        "DocumentCopy",
		Hidden:      false,
		Sort:        2,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// ==================== 第三层：按钮权限 ====================
	// 用户管理按钮权限
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   userMenu.ID,
		Name:       "UserAdd",
		Title:      "新增用户",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:user:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   userMenu.ID,
		Name:       "UserEdit",
		Title:      "编辑用户",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:user:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   userMenu.ID,
		Name:       "UserDelete",
		Title:      "删除用户",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:user:delete",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   userMenu.ID,
		Name:       "UserView",
		Title:      "查看用户",
		Icon:       "View",
		Sort:       4,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:user:view",
	})
	if err != nil {
		return err
	}

	// 角色管理按钮权限
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   roleMenu.ID,
		Name:       "RoleAdd",
		Title:      "新增角色",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:role:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   roleMenu.ID,
		Name:       "RoleEdit",
		Title:      "编辑角色",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:role:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   roleMenu.ID,
		Name:       "RoleDelete",
		Title:      "删除角色",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:role:delete",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   roleMenu.ID,
		Name:       "RoleAssign",
		Title:      "分配权限",
		Icon:       "Setting",
		Sort:       4,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:role:assign",
	})
	if err != nil {
		return err
	}

	// 菜单管理按钮权限
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   menuMenu.ID,
		Name:       "MenuAdd",
		Title:      "新增菜单",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:menu:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   menuMenu.ID,
		Name:       "MenuEdit",
		Title:      "编辑菜单",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:menu:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   menuMenu.ID,
		Name:       "MenuDelete",
		Title:      "删除菜单",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:menu:delete",
	})
	if err != nil {
		return err
	}

	// 部门管理按钮权限
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   deptMenu.ID,
		Name:       "DeptAdd",
		Title:      "新增部门",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:department:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   deptMenu.ID,
		Name:       "DeptEdit",
		Title:      "编辑部门",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:department:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   deptMenu.ID,
		Name:       "DeptDelete",
		Title:      "删除部门",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "system:department:delete",
	})
	if err != nil {
		return err
	}

	// 定时任务按钮权限
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   taskMenu.ID,
		Name:       "TaskAdd",
		Title:      "新增任务",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "task:task:add",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   taskMenu.ID,
		Name:       "TaskEdit",
		Title:      "编辑任务",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "task:task:update",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   taskMenu.ID,
		Name:       "TaskDelete",
		Title:      "删除任务",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "task:task:delete",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   taskMenu.ID,
		Name:       "TaskExecute",
		Title:      "执行任务",
		Icon:       "VideoPlay",
		Sort:       4,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "task:task:execute",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   taskMenu.ID,
		Name:       "TaskEnable",
		Title:      "启用任务",
		Icon:       "CircleCheck",
		Sort:       5,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "task:task:enable",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   taskMenu.ID,
		Name:       "TaskDisable",
		Title:      "禁用任务",
		Icon:       "CircleClose",
		Sort:       6,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "task:task:disable",
	})
	if err != nil {
		return err
	}

	// 任务统计子菜单
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    taskDir.ID,
		Path:        "statistics",
		Name:        "TaskStatistics",
		Component:   "views/task/statistics/index",
		Title:       "任务统计",
		Icon:        "DataAnalysis",
		Hidden:      false,
		Sort:        2,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 任务执行日志子菜单
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    taskDir.ID,
		Path:        "execution-logs",
		Name:        "TaskExecutionLogs",
		Component:   "views/task/execution-logs/index",
		Title:       "任务执行日志",
		Icon:        "Document",
		Hidden:      false,
		Sort:        3,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 节点管理子菜单
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:    taskDir.ID,
		Path:        "nodes",
		Name:        "TaskNodes",
		Component:   "views/task/nodes/index",
		Title:       "节点管理",
		Icon:        "Connection",
		Hidden:      false,
		Sort:        4,
		Status:      1,
		KeepAlive:   true,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 操作日志按钮权限
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   opLogMenu.ID,
		Name:       "OpLogView",
		Title:      "查看详情",
		Icon:       "View",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "log:operation:view",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   opLogMenu.ID,
		Name:       "OpLogDelete",
		Title:      "删除日志",
		Icon:       "Delete",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "log:operation:delete",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   opLogMenu.ID,
		Name:       "OpLogBatchDelete",
		Title:      "批量删除",
		Icon:       "Delete",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "log:operation:batchDelete",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   opLogMenu.ID,
		Name:       "OpLogClear",
		Title:      "清空日志",
		Icon:       "Delete",
		Sort:       4,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "log:operation:clear",
	})
	if err != nil {
		return err
	}

	// 登录日志按钮权限
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   loginLogMenu.ID,
		Name:       "LoginLogView",
		Title:      "查看详情",
		Icon:       "View",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "log:login:view",
	})
	if err != nil {
		return err
	}

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   loginLogMenu.ID,
		Name:       "LoginLogClear",
		Title:      "清空日志",
		Icon:       "Delete",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "log:login:clear",
	})
	if err != nil {
		return err
	}

	// ==================== 多租户管理 ====================
	// 页面菜单（一级）：租户生命周期/应用路由/用户授权/运维审计
	tenantMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		Path:        "tenant",
		Name:        "TenantManagement",
		Component:   "views/base/tenant/index",
		Title:       "多租户管理",
		Icon:        "OfficeBuilding",
		Hidden:      false,
		Sort:        4,
		Status:      1,
		KeepAlive:   false,
		DefaultMenu: false,
		Type:        entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 新增租户按钮（含应用路由绑定/用户授权）
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   tenantMenu.ID,
		Name:       "TenantAdd",
		Title:      "新增租户",
		Icon:       "Plus",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "tenant:tenant:add",
	})
	if err != nil {
		return err
	}

	// 编辑租户按钮（基础信息/独立实例连接）
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   tenantMenu.ID,
		Name:       "TenantEdit",
		Title:      "编辑租户",
		Icon:       "Edit",
		Sort:       2,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "tenant:tenant:update",
	})
	if err != nil {
		return err
	}

	// 生命周期操作按钮（重试初始化/禁用/启用/归档）
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   tenantMenu.ID,
		Name:       "TenantLifecycle",
		Title:      "生命周期操作",
		Icon:       "RefreshRight",
		Sort:       3,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "tenant:tenant:lifecycle",
	})
	if err != nil {
		return err
	}

	// 授权管理按钮（用户授权/取消/应用路由解绑）
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   tenantMenu.ID,
		Name:       "TenantGrantManage",
		Title:      "授权管理",
		Icon:       "Key",
		Sort:       4,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "tenant:tenant:grant",
	})
	if err != nil {
		return err
	}

	// ==================== 微信支付 ====================
	// 渠道配置升级为一级菜单（与多租户管理同级）：租户管理与渠道配置同属平台级管理业务，
	// 切换租户后需在渠道配置维护对应租户的应用/商户，挂同一层级减少跳转层级
	wxChannelMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		Path:      "wx/channel",
		Name:      "WxChannel",
		Component: "views/wx/channel/index",
		Title:     "渠道配置",
		Icon:      "Setting",
		Hidden:    false,
		Sort:      5,
		Status:    1,
		KeepAlive: false,
		Type:      entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 渠道配置按钮权限（type:3）：应用/商户/UI/绑定四类操作
	channelButtons := []struct {
		name       string
		title      string
		icon       string
		sort       int
		permission string
	}{
		{"WxChannelApp", "应用配置管理", "Setting", 1, "wx:channel:app"},
		{"WxChannelMerchant", "商户配置管理", "Shop", 2, "wx:channel:merchant"},
		{"WxChannelBind", "应用商户绑定", "Link", 3, "wx:channel:bind"},
		{"WxChannelUI", "租户UI配置", "Brush", 4, "wx:channel:ui"},
	}
	for _, btn := range channelButtons {
		if _, err := ms.createOrGetMenu(logger, &entity.Menu{
			ParentID:   wxChannelMenu.ID,
			Name:       btn.name,
			Title:      btn.title,
			Icon:       btn.icon,
			Sort:       btn.sort,
			Status:     1,
			Type:       entity.MenuTypeButton,
			Permission: btn.permission,
		}); err != nil {
			return err
		}
	}

	// 微信支付目录（对账中心的父级；渠道配置已升级为一级菜单）
	wxPayDir, err := ms.createOrGetMenu(logger, &entity.Menu{
		Path:        "wx",
		Name:        "WxPayRoot",
		Title:       "微信支付",
		Icon:        "Wallet",
		Hidden:      false,
		Sort:        6,
		Status:      1,
		KeepAlive:   false,
		DefaultMenu: false,
		Type:        entity.MenuTypeDirectory,
	})
	if err != nil {
		return err
	}

	// 对账中心菜单（对账汇总 + 支付单明细）
	wxReconMenu, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:  wxPayDir.ID,
		Path:      "recon",
		Name:      "WxRecon",
		Component: "views/wx/recon/index",
		Title:     "对账中心",
		Icon:      "DataAnalysis",
		Hidden:    false,
		Sort:      1,
		Status:    1,
		KeepAlive: false,
		Type:      entity.MenuTypeMenu,
	})
	if err != nil {
		return err
	}

	// 对账中心按钮权限（type:3）：对账查询
	if _, err := ms.createOrGetMenu(logger, &entity.Menu{
		ParentID:   wxReconMenu.ID,
		Name:       "WxReconView",
		Title:      "对账查询",
		Icon:       "View",
		Sort:       1,
		Status:     1,
		Type:       entity.MenuTypeButton,
		Permission: "wx:recon:view",
	}); err != nil {
		return err
	}

	logger.Info("菜单种子数据初始化完成")
	return nil
}

// createOrGetMenu 创建菜单，如果菜单已存在则获取已存在的菜单
func (ms *MenuSeed) createOrGetMenu(logger logger.Logger, menu *entity.Menu) (*entity.Menu, error) {
	// 先根据名称检查是否已存在
	existing, err := ms.menuRepo.GetByCode(menu.Name)
	if err == nil {
		// 已存在：校正关键字段（挂载位置/排序/标题/图标/组件/权限标识），支持菜单结构调整后存量库自愈
		needsUpdate := existing.ParentID != menu.ParentID ||
			existing.Sort != menu.Sort ||
			existing.Title != menu.Title ||
			existing.Icon != menu.Icon ||
			existing.Component != menu.Component ||
			existing.Path != menu.Path ||
			existing.Hidden != menu.Hidden ||
			existing.Permission != menu.Permission
		if needsUpdate {
			existing.ParentID = menu.ParentID
			existing.Sort = menu.Sort
			existing.Title = menu.Title
			existing.Icon = menu.Icon
			existing.Component = menu.Component
			existing.Path = menu.Path
			existing.Hidden = menu.Hidden
			existing.Permission = menu.Permission
			if err := ms.menuRepo.Update(existing); err != nil {
				logger.Error("校正已有菜单失败", "name", menu.Name, "error", err)
				return nil, err
			}
			logger.Info("菜单已存在，校正挂载位置/排序等字段", "name", menu.Name, "id", existing.ID)
		} else {
			logger.Info("菜单已存在，跳过创建", "name", menu.Name, "id", existing.ID)
		}
		return existing, nil
	}

	// 不存在则创建
	if err := ms.menuRepo.Create(menu); err != nil {
		logger.Error("创建菜单失败", "name", menu.Name, "error", err)
		return nil, err
	}
	logger.Info("创建菜单成功", "name", menu.Name, "id", menu.ID)

	return menu, nil
}
