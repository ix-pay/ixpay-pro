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
// 说明：wx 模块仅提供 wxapp/miniapp 前端使用的接口（认证/支付/退款/医保/账单）；
// 管理平台使用的渠道配置接口已迁移至 base 模块（/api/admin/wx/*），鉴权统一由 base 负责
type AppWX struct {
	router            *gin.Engine
	db                *database.PostgresDB
	auth              *auth.JWTAuth
	permissions       *auth.PermissionManager
	logger            logger.Logger
	authController    *wxapi.AuthController
	paymentController *wxapi.PaymentController
	medicalController *wxapi.MedicalController
	refundController  *wxapi.RefundController
	billController    *wxapi.BillController
	// 渠道公开控制器：wx 模块仅使用其公开配置接口 GetPublicApp（wxapp/miniapp 前端空壳启动加载）；
	// 管理端渠道接口已迁移至 base 模块（internal/app/base/api/wx_channel_handler.go）
	channelController *wxapi.WxPublicAppController
}

// NewApplication 创建应用程序实例
func NewAppWX(
	log logger.Logger,
	db *database.PostgresDB,
	auth *auth.JWTAuth,
	permissions *auth.PermissionManager,
	authController *wxapi.AuthController,
	paymentController *wxapi.PaymentController,
	medicalController *wxapi.MedicalController,
	refundController *wxapi.RefundController,
	billController *wxapi.BillController,
	channelController *wxapi.WxPublicAppController,
) (*AppWX, error) {
	// 注册 wx 模块到租户模块注册表（租户库初始化任务按注册表执行迁移）
	migrations.RegisterTenantModule()

	// 创建应用实例
	app := &AppWX{
		db:                db,
		auth:              auth,
		permissions:       permissions,
		logger:            log,
		authController:    authController,
		paymentController: paymentController,
		medicalController: medicalController,
		refundController:  refundController,
		billController:    billController,
		channelController: channelController,
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
