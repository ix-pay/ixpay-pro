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

// Order 返回初始化顺序（第五个执行）
func (ms *MenuSeed) Order() int {
	return 5
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

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
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

	// 日志管理下的菜单
	_, err = ms.createOrGetMenu(logger, &entity.Menu{
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

	_, err = ms.createOrGetMenu(logger, &entity.Menu{
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
		Permission: "system:user:edit",
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
		Permission: "system:role:edit",
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
		Permission: "system:menu:edit",
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
		Permission: "system:department:edit",
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
		Permission: "task:task:edit",
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

	logger.Info("菜单种子数据初始化完成")
	return nil
}

// createOrGetMenu 创建菜单，如果菜单已存在则获取已存在的菜单
func (ms *MenuSeed) createOrGetMenu(logger logger.Logger, menu *entity.Menu) (*entity.Menu, error) {
	// 先根据名称检查是否已存在
	existing, err := ms.menuRepo.GetByCode(menu.Name)
	if err == nil {
		logger.Info("菜单已存在，跳过创建", "name", menu.Name, "id", existing.ID)
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
