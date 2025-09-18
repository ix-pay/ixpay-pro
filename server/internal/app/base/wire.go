package base

import (
	"github.com/google/wire"
	baseapi "github.com/ix-pay/ixpay-pro/internal/app/base/api/v1"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/repository"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/service"
)

// 导入必要的包

// 定义依赖注入的提供者
var ProviderSetBase = wire.NewSet(
	// 仓库层
	repository.NewUserRepository,

	// 服务层
	service.NewUserService,

	// 控制器层
	baseapi.NewAuthController,
	baseapi.NewUserController,
	baseapi.NewTaskController,

	// 应用层
	NewAppBase,
)
