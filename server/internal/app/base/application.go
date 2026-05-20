package base

import (
	"context"

	baseapi "github.com/ix-pay/ixpay-pro/internal/app/base/api"
	"github.com/ix-pay/ixpay-pro/internal/app/base/middleware"
	"github.com/ix-pay/ixpay-pro/internal/app/base/migrations"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/seed"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/eventbus"
	infraMiddleware "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/middleware"

	"github.com/gin-gonic/gin"
)

// AppBase 应用程序结构
type AppBase struct {
	router                   *gin.Engine
	db                       *database.PostgresDB
	auth                     *auth.JWTAuth
	permissions              *auth.PermissionManager
	logger                   logger.Logger
	config                   *config.Config
	authController           *baseapi.AuthController
	userController           *baseapi.UserController
	taskController           *baseapi.TaskController
	apiController            *baseapi.APIController
	menuController           *baseapi.MenuController
	roleController           *baseapi.RoleController
	btnPermController        *baseapi.BtnPermController
	configController         *baseapi.ConfigController
	dictController           *baseapi.DictController
	operationLogController   *baseapi.OperationLogController
	departmentController     *baseapi.DepartmentController
	positionController       *baseapi.PositionController
	noticeController         *baseapi.NoticeController
	loginLogController       *baseapi.LoginLogController
	onlineUserController     *baseapi.OnlineUserController
	monitorController        *baseapi.MonitorController
	permissionLogController  *baseapi.PermissionLogController
	nodeController           *baseapi.NodeController
	eventController          *baseapi.EventController
	gatewayServiceController *baseapi.GatewayServiceController
	userRepo                 repo.UserRepository
	apiRepo                  repo.APIRepository
	roleRepo                 repo.RoleRepository
	menuRepo                 repo.MenuRepository
	configRepo               repo.ConfigRepository
	dictRepo                 repo.DictRepository
	dictItemRepo             repo.DictItemRepository
	departmentRepo           repo.DepartmentRepository
	positionRepo             repo.PositionRepository
	permissionService        *service.PermissionService
	operationLogService      *service.OperationLogService
	onlineUserService        *service.OnlineUserService
	taskExecutionLogRepo     repo.TaskExecutionLogRepository // 任务执行日志仓库
	taskRepo                 repo.TaskRepository             // 任务仓库
	cache                    cache.Cache
	eventBus                 *eventbus.EventBus
}

// NewAppBase 创建应用程序实例
func NewAppBase(
	log logger.Logger,
	config *config.Config,
	db *database.PostgresDB,
	auth *auth.JWTAuth,
	permissions *auth.PermissionManager,
	authController *baseapi.AuthController,
	userController *baseapi.UserController,
	taskController *baseapi.TaskController,
	apiController *baseapi.APIController,
	menuController *baseapi.MenuController,
	roleController *baseapi.RoleController,
	btnPermController *baseapi.BtnPermController,
	configController *baseapi.ConfigController,
	dictController *baseapi.DictController,
	operationLogController *baseapi.OperationLogController,
	departmentController *baseapi.DepartmentController,
	positionController *baseapi.PositionController,
	noticeController *baseapi.NoticeController,
	loginLogController *baseapi.LoginLogController,
	onlineUserController *baseapi.OnlineUserController,
	monitorController *baseapi.MonitorController,
	permissionLogController *baseapi.PermissionLogController,
	nodeController *baseapi.NodeController,
	eventController *baseapi.EventController,
	gatewayServiceController *baseapi.GatewayServiceController,
	userRepo repo.UserRepository,
	apiRepo repo.APIRepository,
	roleRepo repo.RoleRepository,
	menuRepo repo.MenuRepository,
	configRepo repo.ConfigRepository,
	dictRepo repo.DictRepository,
	dictItemRepo repo.DictItemRepository,
	departmentRepo repo.DepartmentRepository,
	positionRepo repo.PositionRepository,
	operationLogService *service.OperationLogService,
	onlineUserService *service.OnlineUserService,
	taskExecutionLogRepo repo.TaskExecutionLogRepository,
	taskRepo repo.TaskRepository,
	cache cache.Cache,
	eventBus *eventbus.EventBus,
) (*AppBase, error) {
	// 创建应用实例
	app := &AppBase{
		router:                   nil,
		db:                       db,
		auth:                     auth,
		permissions:              permissions,
		logger:                   log,
		config:                   config,
		authController:           authController,
		userController:           userController,
		taskController:           taskController,
		apiController:            apiController,
		menuController:           menuController,
		roleController:           roleController,
		btnPermController:        btnPermController,
		configController:         configController,
		dictController:           dictController,
		operationLogController:   operationLogController,
		departmentController:     departmentController,
		positionController:       positionController,
		noticeController:         noticeController,
		loginLogController:       loginLogController,
		onlineUserController:     onlineUserController,
		monitorController:        monitorController,
		permissionLogController:  permissionLogController,
		nodeController:           nodeController,
		eventController:          eventController,
		gatewayServiceController: gatewayServiceController,
		userRepo:                 userRepo,
		apiRepo:                  apiRepo,
		roleRepo:                 roleRepo,
		menuRepo:                 menuRepo,
		configRepo:               configRepo,
		dictRepo:                 dictRepo,
		dictItemRepo:             dictItemRepo,
		departmentRepo:           departmentRepo,
		positionRepo:             positionRepo,
		operationLogService:      operationLogService,
		onlineUserService:        onlineUserService,
		taskExecutionLogRepo:     taskExecutionLogRepo,
		taskRepo:                 taskRepo,
		cache:                    cache,
		eventBus:                 eventBus,
	}
	return app, nil
}

// setupMiddleware 设置中间件
// 该方法负责配置和注册base模块所需的所有中间件
// 目前该方法为基础框架预留，用于在模块初始化时设置特定的中间件
// 中间件配置采用中间件配置中心(MiddlewareConfig)模式，集中管理所有中间件
// 遵循项目中间件注册顺序规范：先注册全局中间件，再注册路由级中间件，最后注册控制器级中间件
// 与应用级中间件(application.go中的setupMiddleware)配合使用，提供模块化的中间件管理
func (a *AppBase) setupMiddleware() {
	// 添加Prometheus中间件
	a.router.Use(infraMiddleware.PrometheusMiddleware())

	// 添加速率限制中间件 (100个请求/秒，允许10个突发请求)
	a.router.Use(infraMiddleware.RateLimitMiddleware(100, 10))

	// 添加熔断中间件
	circuitBreakerSettings := infraMiddleware.DefaultCircuitBreakerSettings()
	circuitBreakerSettings.Name = "BaseServiceCircuitBreaker"
	a.router.Use(infraMiddleware.CircuitBreakerMiddleware("BaseService", circuitBreakerSettings))

	// 添加操作日志中间件
	a.router.Use(middleware.OperationLogMiddleware(a.operationLogService, a.logger))

	// 添加在线用户活跃时间更新中间件
	a.router.Use(middleware.OnlineUserActiveMiddleware(a.onlineUserService, a.logger))

	// 添加Prometheus指标导出路由
	a.router.GET("/metrics", infraMiddleware.PrometheusHandler())
}

// initializeSeedData 初始化种子数据
func (a *AppBase) initializeSeedData() {
	a.logger.Info("开始初始化种子数据")

	// 导入种子数据包
	seedManager := seed.NewSeedManager(a.logger)

	// 注册所有种子数据
	seedManager.RegisterAll([]seed.Seed{
		seed.NewConfigSeed(a.configRepo),
		seed.NewRoleSeed(a.roleRepo),
		seed.NewUserSeed(a.userRepo, a.roleRepo),
		seed.NewAPISeed(a.apiRepo),
		seed.NewMenuSeed(a.menuRepo, a.apiRepo),
		seed.NewDictSeed(a.dictRepo, a.dictItemRepo),
		seed.NewDepartmentSeed(a.departmentRepo),
		seed.NewPositionSeed(a.positionRepo),
		seed.NewTaskSeed(a.taskRepo),
	})

	// 初始化所有种子数据
	if err := seedManager.Init(a.db); err != nil {
		a.logger.Error("初始化种子数据失败", "error", err)
		return
	}

	a.logger.Info("种子数据初始化完成")
}

// 初始化基础应用
func (a *AppBase) Init(router *gin.Engine) {
	a.logger.Info("初始化基础应用")
	a.router = router

	// 执行数据库迁移
	migrations.MigrateDatabase(a.db, a.logger)

	// 设置中间件
	a.setupMiddleware()

	// 设置路由
	a.setupRoutes()

	// 执行种子数据初始化
	if a.config.Server.InitSeedData {
		a.logger.Info("配置允许初始化种子数据，执行种子数据初始化")
		a.initializeSeedData()
	} else {
		a.logger.Info("配置禁用初始化种子数据")
	}

	// 设置任务执行日志仓库到任务管理器
	if a.taskExecutionLogRepo != nil {
		a.taskController.GetManager().SetExecutionLogRepository(a.taskExecutionLogRepo)
		a.logger.Info("任务执行日志仓库已设置到任务管理器")
	}

	// 设置任务配置仓库和工厂到任务管理器
	if a.taskRepo != nil {
		a.taskController.GetManager().SetTaskRepository(a.taskRepo)
		a.logger.Info("任务配置仓库已设置到任务管理器")
	}

	if a.taskController.GetFactory() != nil {
		a.taskController.GetManager().SetTaskFactory(a.taskController.GetFactory())
		a.logger.Info("任务工厂已设置到任务管理器")
	}

	// 从数据库加载已启用的任务
	if err := a.taskController.GetManager().LoadTasksFromDB(); err != nil {
		a.logger.Error("从数据库加载任务失败", "error", err)
	} else {
		a.logger.Info("从数据库加载任务完成")
	}

	// 启动事件总线
	if a.eventBus != nil {
		if err := a.eventBus.Start(); err != nil {
			a.logger.Error("启动事件总线失败", "error", err)
		} else {
			a.logger.Info("事件总线启动成功")
		}
	}
}

// Shutdown 关闭基础应用
func (a *AppBase) Shutdown(ctx context.Context) error {
	// 停止事件总线
	if a.eventBus != nil {
		if err := a.eventBus.Stop(); err != nil {
			a.logger.Error("停止事件总线失败", "error", err)
			return err
		}
		a.logger.Info("事件总线已停止")
	}
	return nil
}
