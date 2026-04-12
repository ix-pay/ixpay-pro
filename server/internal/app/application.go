package app

import (
	"context"
	"net/http"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/app/base"
	"github.com/ix-pay/ixpay-pro/internal/app/wx"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/snowflake"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/middleware"

	_ "github.com/ix-pay/ixpay-pro/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Application 应用程序结构
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
) (*Application, error) {

	// 创建路由引擎
	router := gin.New()

	// 配置 Swagger
	if cfg.Swagger.Enabled {
		router.GET(cfg.Swagger.Path+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 创建中间件配置中心
	middlewareConfig := middleware.SetupMiddlewareConfig(auth, log, cache)

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
	}

	// 设置中间件
	app.setupMiddleware()

	// 设置路由
	app.setupRoutes()

	// 设置雪花算法实例到数据库模块
	database.SetSnowflakeInstance(app.snowflake)

	// 初始化模块应用
	app.appBase.Init(router)
	app.appWX.Init(router)

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

	// 通过中间件配置中心注册所有中间件
	a.middlewareConfig.RegisterAllMiddlewares(a.router)
	a.logger.Info("中间件注册成功")
}

// setupRoutes 方法在routes.go文件中定义

// Start 启动HTTP服务器
func (a *Application) Start() error {
	a.logger.Info("启动HTTP服务器", "address", a.server.Addr)
	return a.server.ListenAndServe()
}

// Shutdown 优雅关闭HTTP服务器
func (a *Application) Shutdown(ctx context.Context) error {
	a.logger.Info("正在关闭HTTP服务器")
	// 关闭缓存连接
	if a.cache != nil {
		a.cache.Close()
	}
	return a.server.Shutdown(ctx)
}
