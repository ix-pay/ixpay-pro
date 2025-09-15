package app

import (
	"context"
	"net/http"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/app/controller"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/middleware"

	_ "github.com/ix-pay/ixpay-pro/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Application 应用程序结构
type Application struct {
	router            *gin.Engine
	db                *database.PostgresDB
	auth              *auth.JWTAuth
	permissions       *auth.PermissionManager
	logger            logger.Logger
	server            *http.Server
	userController    *controller.UserController
	paymentController *controller.PaymentController
	taskController    *controller.TaskController
}

// NewApplication 创建应用程序实例
func NewApplication(
	cfg *config.Config,
	log logger.Logger,
	db *database.PostgresDB,
	auth *auth.JWTAuth,
	permissions *auth.PermissionManager,
	userController *controller.UserController,
	paymentController *controller.PaymentController,
	taskController *controller.TaskController,
) (*Application, error) {
	// 执行数据库迁移，创建所有需要的表
	if err := db.Migrate(log); err != nil {
		log.Error("Failed to migrate database", "error", err)
		return nil, err
	}

	// 创建路由引擎
	router := gin.New()

	// 配置Swagger
	if cfg.Swagger.Enabled {
		router.GET(cfg.Swagger.Path+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 创建应用实例
	app := &Application{
		router:            router,
		db:                db,
		auth:              auth,
		permissions:       permissions,
		logger:            log,
		userController:    userController,
		paymentController: paymentController,
		taskController:    taskController,
	}

	// 设置中间件
	app.setupMiddleware()

	// 设置路由
	app.setupRoutes()

	// 初始化权限
	app.initializePermissions()

	// 创建HTTP服务器
	app.server = &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	return app, nil
}

// setupMiddleware 设置中间件
func (a *Application) setupMiddleware() {
	// 使用Gin默认的日志和恢复中间件
	a.router.Use(gin.Logger())
	a.router.Use(gin.Recovery())

	// 使用自定义的日志中间件
	a.router.Use(middleware.LoggerMiddleware(a.logger))

	// 使用CORS中间件
	a.router.Use(middleware.CORSMiddleware())
}

// setupRoutes 方法在routes.go文件中定义

// initializePermissions 初始化权限
func (a *Application) initializePermissions() {
	// 定义系统权限
	permissions := []auth.Permission{
		// 用户路由权限
		{Path: "/api/v1/user/info", Method: "GET", Roles: []string{"user", "admin"}, WechatGrant: true},

		// 支付路由权限
		{Path: "/api/v1/payment", Method: "POST", Roles: []string{"user", "admin"}, WechatGrant: true},
		{Path: "/api/v1/payment/:id", Method: "GET", Roles: []string{"user", "admin"}, WechatGrant: true},
		{Path: "/api/v1/payment", Method: "GET", Roles: []string{"user", "admin"}, WechatGrant: true},
		{Path: "/api/v1/payment/:id/cancel", Method: "PUT", Roles: []string{"user", "admin"}, WechatGrant: true},

		// 任务路由权限（需要admin角色）
		{Path: "/api/v1/task", Method: "POST", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id", Method: "DELETE", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id/start", Method: "POST", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id/stop", Method: "POST", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id/retry", Method: "POST", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task", Method: "GET", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id", Method: "GET", Roles: []string{"admin"}, WechatGrant: false},
	}

	// 缓存权限数据
	if err := a.permissions.CachePermissions(permissions); err != nil {
		a.logger.Error("Failed to cache permissions", "error", err)
	}
}

// Start 启动HTTP服务器
func (a *Application) Start() error {
	a.logger.Info("Starting HTTP server", "address", a.server.Addr)
	return a.server.ListenAndServe()
}

// Shutdown 优雅关闭HTTP服务器
func (a *Application) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down HTTP server")
	return a.server.Shutdown(ctx)
}
