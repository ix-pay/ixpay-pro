//go:build wireinject
// +build wireinject

package app

import (
	"github.com/ix-pay/ixpay-pro/internal/app/controller"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/domain/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/redis"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/repository"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/snowflake"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/task"

	"github.com/google/wire"
)

// 导入必要的包

// 定义依赖注入的提供者
var ProviderSet = wire.NewSet(
	// 配置
	config.LoadConfig,

	// 基础设施层
	logger.NewLogger,
	database.NewPostgresDB,
	redis.NewRedisClient,
	auth.NewJWTAuth,
	snowflake.SetupSnowflake,
	auth.NewPermissionManager,
	task.NewTaskManager,

	// 仓库层
	repository.NewUserRepository,
	repository.NewPaymentRepository,

	// 服务层
	service.NewWechatService,
	service.NewPaymentService,
	service.NewUserService,

	// 控制器层
	controller.NewUserController,
	controller.NewPaymentController,
	controller.NewTaskController,

	// 应用层
	NewApplication,
)

// InitializeApp 初始化应用程序
// wire会根据ProviderSet自动生成依赖注入代码
func InitializeApp() (*Application, error) {
	wire.Build(ProviderSet)
	return nil, nil
}
