package wx

import (
	wxapi "github.com/ix-pay/ixpay-pro/internal/app/wx/api/v1"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Application 应用程序结构
type AppWX struct {
	router            *gin.Engine
	db                *database.PostgresDB
	auth              *auth.JWTAuth
	permissions       *auth.PermissionManager
	logger            logger.Logger
	authController    *wxapi.AuthController
	paymentController *wxapi.PaymentController
}

// NewApplication 创建应用程序实例
func NewAppWX(
	log logger.Logger,
	db *database.PostgresDB,
	auth *auth.JWTAuth,
	permissions *auth.PermissionManager,
	authController *wxapi.AuthController,
	paymentController *wxapi.PaymentController,
) (*AppWX, error) {
	// 执行数据库迁移，创建所有需要的表
	// if err := db.Migrate(log); err != nil {
	// 	log.Error("Failed to migrate database", "error", err)
	// 	return nil, err
	// }

	// 创建应用实例
	app := &AppWX{
		db:                db,
		auth:              auth,
		permissions:       permissions,
		logger:            log,
		authController:    authController,
		paymentController: paymentController,
	}

	return app, nil
}

// setupMiddleware 设置中间件
func (a *AppWX) setupMiddleware() {
}

// setupRoutes 方法在routes.go文件中定义

// initializePermissions 初始化权限
func (a *AppWX) initializePermissions() {
	// 定义系统权限
	permissions := []auth.Permission{
		// 支付路由权限
		{Path: "/api/v1/payment", Method: "POST", Roles: []string{"user", "admin"}, WechatGrant: true},
		{Path: "/api/v1/payment/:id", Method: "GET", Roles: []string{"user", "admin"}, WechatGrant: true},
		{Path: "/api/v1/payment", Method: "GET", Roles: []string{"user", "admin"}, WechatGrant: true},
		{Path: "/api/v1/payment/:id/cancel", Method: "PUT", Roles: []string{"user", "admin"}, WechatGrant: true},
	}

	// 缓存权限数据
	if err := a.permissions.CachePermissions(permissions); err != nil {
		a.logger.Error("Failed to cache permissions", "error", err)
	}
}

// 初始化微信应用
func (a *AppWX) Init(router *gin.Engine) {
	a.logger.Info("初始化微信应用")

	a.router = router
	// 设置中间件
	a.setupMiddleware()

	// 设置路由
	a.setupRoutes()

	// 初始化权限
	a.initializePermissions()

}
