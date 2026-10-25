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
	repository.NewRefundRepository,
	repository.NewMedRefundRepository,
	repository.NewWxChannelRepository,
)

var ProviderSetWXService = wire.NewSet(
	// 服务层
	service.NewPaymentService,
	service.NewWechatPayInfoService,
	service.NewWXAuthService,
	service.NewWechatPayService,
	service.NewRefundService,
	service.NewBillService,
	service.NewMedicalService,
	service.NewMedRefundService,
	service.NewWxChannelService,
)

var ProviderSetWXController = wire.NewSet(
	// 控制器层
	wxapi.NewPaymentController,
	wxapi.NewAuthController,
	wxapi.NewMedicalController,
	wxapi.NewRefundController,
	wxapi.NewBillController,
	wxapi.NewWxPublicAppController,
)

var ProviderSetWXApp = wire.NewSet(
	// 应用层
	NewAppWX,
)
