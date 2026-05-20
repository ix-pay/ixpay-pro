package app

import (
	"context"
	"net/http"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/app/base"
	"github.com/ix-pay/ixpay-pro/internal/app/wx"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/gateway"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/snowflake"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/middleware"

	_ "github.com/ix-pay/ixpay-pro/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Application 应用程序实例
type Application struct {
	router           *gin.Engine
	db               *database.PostgresDB
	snowflake        *snowflake.Snowflake
	auth             *auth.JWTAuth
	permissions      *auth.PermissionManager
	logger           logger.Logger
	loggerManager    *logger.MultiLogger
	cache            cache.Cache
	server           *http.Server
	middlewareConfig *middleware.MiddlewareConfig
	appBase          *base.AppBase
	appWX            *wx.AppWX
	gatewayClient    *gateway.GatewayClient
	cfg              *config.Config
	nodeRegistry     *task.NodeRegistry
	taskManager      *task.TaskManager
}

// NewApplication 创建应用程序实例
func SetupApplication(
	cfg *config.Config,
	logManager *logger.MultiLogger,
	log logger.Logger,
	db *database.PostgresDB,
	snowflake *snowflake.Snowflake,
	auth *auth.JWTAuth,
	permissions *auth.PermissionManager,
	cache cache.Cache,
	appBase *base.AppBase,
	appWX *wx.AppWX,
	nodeRegistry *task.NodeRegistry,
	taskManager *task.TaskManager,
) (*Application, error) {

	// 创建路由引擎
	router := gin.New()

	// 配置 Swagger
	if cfg.Swagger.Enabled {
		router.GET(cfg.Swagger.Path+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 创建中间件配置中心
	middlewareConfig := middleware.SetupMiddlewareConfig(auth, log, cache)

	// 设置节点注册表到任务管理器
	taskManager.SetNodeRegistry(nodeRegistry)

	// 设置任务工厂和任务仓库到任务管理器（仅 Task 和 all 角色需要）
	if cfg.Server.NodeRole == config.NodeRoleTask || cfg.Server.NodeRole == config.NodeRoleAll {
		taskManager.SetTaskFactory(task.NewTaskFactory())
		taskManager.SetDefaultTimeout(5 * time.Minute)

		// 从数据库加载启用的任务
		if err := taskManager.LoadTasksFromDB(); err != nil {
			log.Warn("从数据库加载任务失败", "error", err)
		} else {
			log.Info("从数据库加载任务成功", "node_role", cfg.Server.NodeRole)
		}
	}

	// 注册当前节点到 Redis
	if err := nodeRegistry.Register(context.Background()); err != nil {
		log.Warn("注册节点到 Redis 失败", "error", err)
	} else {
		log.Info("成功注册节点到 Redis", "node_id", nodeRegistry.GetNodeID())
	}

	// 启动节点心跳
	nodeRegistry.StartHeartbeat(context.Background(), 10*time.Second)
	log.Info("节点心跳已启动", "interval", "10s")

	// 创建应用实例
	app := &Application{
		router:           router,
		db:               db,
		snowflake:        snowflake,
		auth:             auth,
		permissions:      permissions,
		logger:           log,
		loggerManager:    logManager,
		cache:            cache,
		middlewareConfig: middlewareConfig,
		appBase:          appBase,
		appWX:            appWX,
		cfg:              cfg,
		nodeRegistry:     nodeRegistry,
		taskManager:      taskManager,
	}

	// 设置中间件（仅 API 和 all 角色需要）
	if cfg.Server.NodeRole == config.NodeRoleAPI || cfg.Server.NodeRole == config.NodeRoleAll {
		app.setupMiddleware()

		// 设置路由（仅 API 和 all 角色需要）
		app.setupRoutes()

		// 设置雪花算法实例到数据库模块
		database.SetSnowflakeInstance(app.snowflake)

		// 初始化模块应用（仅 API 和 all 角色需要）
		app.appBase.Init(router)
		app.appWX.Init(router)

		// 设置健康检查路由
		app.setupHealthCheck()

		// 创建 HTTP 服务器（仅 API 和 all 角色需要）
		app.server = &http.Server{
			Addr:         ":" + cfg.Server.Port,
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
		}

		// 初始化网关客户端（仅 API 和 all 角色需要注册到网关）
		if cfg.Gateway.Enabled {
			port := gateway.ParsePort(cfg.Server.Port, 8586)

			metadata := map[string]string{
				"node_role": cfg.Server.NodeRole,
				"version":   "1.0.0",
			}

			instance := gateway.BuildServiceInstance(
				cfg.Gateway.ServiceName,
				cfg.Server.Host,
				port,
				cfg.Server.NodeRole,
				cfg.Server.NodeID,
				metadata,
			)

			app.gatewayClient = gateway.NewGatewayClient(cfg.Gateway.GatewayURL, cfg.Gateway.AuthKey)

			if err := app.gatewayClient.Register(context.Background(), instance); err != nil {
				log.Warn("注册到网关失败", "error", err)
			} else {
				log.Info("成功注册到网关", "service", cfg.Gateway.ServiceName, "instance_id", instance.ID)
			}
		}
	} else {
		log.Info("Task 节点模式：不启动 API 服务，不注册到网关", "node_role", cfg.Server.NodeRole)
	}

	return app, nil
}

// setupMiddleware 设置中间件
func (a *Application) setupMiddleware() {
	// 使用Gin默认的日志和恢复中间件
	a.router.Use(gin.Logger())
	a.router.Use(gin.Recovery())

	// 通过中间件配置中心注册所有中间件
	a.middlewareConfig.RegisterAllMiddlewares(a.router)
	a.logger.Info("中间件注册成功")
}

// setupRoutes 方法在routes.go文件中定义

// Start 启动应用程序
func (a *Application) Start() error {
	// 根据节点角色启动不同的服务
	if a.cfg.Server.NodeRole == config.NodeRoleAPI || a.cfg.Server.NodeRole == config.NodeRoleAll {
		// API 节点：启动 HTTP 服务器和网关心跳
		a.logger.Info("启动 HTTP 服务器", "address", a.server.Addr)

		// 启动网关心跳
		if a.cfg.Gateway.Enabled && a.gatewayClient != nil {
			heartbeatInterval := time.Duration(a.cfg.Gateway.HeartbeatInterval) * time.Second
			if heartbeatInterval == 0 {
				heartbeatInterval = 10 * time.Second
			}
			a.gatewayClient.StartHeartbeat(context.Background(), heartbeatInterval)
			a.logger.Info("网关心跳已启动", "interval", heartbeatInterval)
		}

		return a.server.ListenAndServe()
	} else if a.cfg.Server.NodeRole == config.NodeRoleTask || a.cfg.Server.NodeRole == config.NodeRoleAll {
		// Task 节点：启动任务调度器
		a.logger.Info("启动任务调度器", "node_role", a.cfg.Server.NodeRole)
		a.taskManager.Start()

		// Task 节点不启动 HTTP 服务器，进入等待状态
		// 使用 channel 阻塞，等待关闭信号
		select {}
	}

	return nil
}

// Shutdown 优雅关闭HTTP服务器
func (a *Application) Shutdown(ctx context.Context) error {
	a.logger.Info("正在关闭HTTP服务器")

	// 注销节点
	if a.nodeRegistry != nil {
		a.logger.Info("正在注销节点")
		a.nodeRegistry.Unregister(ctx)
		a.nodeRegistry.Stop()
	}

	// 注销网关服务
	if a.cfg.Gateway.Enabled && a.gatewayClient != nil {
		a.logger.Info("正在从网关注销服务")
		if err := a.gatewayClient.Deregister(ctx); err != nil {
			a.logger.Warn("从网关注销失败", "error", err)
		} else {
			a.logger.Info("成功从网关注销")
		}
		a.gatewayClient.Stop()
	}

	// 关闭缓存连接
	if a.cache != nil {
		a.cache.Close()
	}
	return a.server.Shutdown(ctx)
}
