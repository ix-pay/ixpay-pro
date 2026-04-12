package wx

import (
	wxapi "github.com/ix-pay/ixpay-pro/internal/app/wx/api"
	"github.com/ix-pay/ixpay-pro/internal/app/wx/migrations"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	auth "github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"

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

// 初始化微信应用
func (a *AppWX) Init(router *gin.Engine) {
	a.logger.Info("初始化微信应用")

	// 执行数据库迁移
	migrations.MigrateDatabase(a.db, a.logger)

	a.router = router
	// 设置中间件
	a.setupMiddleware()

	// 设置路由
	a.setupRoutes()

}
