package wx

import (
	"github.com/google/wire"
	wxapi "github.com/ix-pay/ixpay-pro/internal/app/wx/api"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
	repository "github.com/ix-pay/ixpay-pro/internal/persistence/wx"
)

// 定义依赖注入的提供者
var ProviderSetWXRepo = wire.NewSet(
	// 仓库层
	repository.NewPaymentRepository,
	repository.NewWechatPayInfoRepository,
	repository.NewWXUserRepository,
	repository.NewWXAuthSessionRepository,
)

var ProviderSetWXService = wire.NewSet(
	// 服务层
	service.NewPaymentService,
	service.NewWechatPayInfoService,
	service.NewWXAuthService,
)

var ProviderSetWXController = wire.NewSet(
	// 控制器层
	wxapi.NewPaymentController,
	wxapi.NewAuthController,
)

var ProviderSetWXApp = wire.NewSet(
	// 应用层
	NewAppWX,
)
