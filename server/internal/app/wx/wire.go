package wx

import (
	"github.com/google/wire"
	wxapi "github.com/ix-pay/ixpay-pro/internal/app/wx/api/v1"
	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/repository"
	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/service"
)

// 定义依赖注入的提供者
var ProviderSetWX = wire.NewSet(
	// 仓库层
	repository.NewPaymentRepository,
	repository.NewWXUserRepository,
	repository.NewWXAuthSessionRepository,

	// 服务层
	service.NewPaymentService,
	service.NewWXAuthService,

	// 控制器层
	wxapi.NewPaymentController,
	wxapi.NewAuthController,

	// 应用层
	NewAppWX,
)
