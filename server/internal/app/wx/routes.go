package wx

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/middleware"
)

// setupRoutes 设置路由
func (a *AppWX) setupRoutes() {

	// 公共路由
	public := a.router.Group("/v1")
	{

		// 支付通知路由（不需要认证）
		pay := public.Group("/pay")
		{
			pay.POST("/notify/wechat", a.paymentController.HandleWechatPayNotify)
		}
	}

	// 需要认证的路由
	authenticated := a.router.Group("/v1")
	authenticated.Use(middleware.AuthMiddleware(a.auth))
	authenticated.Use(middleware.PermissionMiddleware(a.permissions))
	{
		// 支付路由
		payment := authenticated.Group("/payment")
		{
			payment.POST("", a.paymentController.CreatePayment)
			payment.GET("/:id", a.paymentController.GetPayment)
			payment.GET("", a.paymentController.GetUserPayments)
			payment.PUT("/:id/cancel", a.paymentController.CancelPayment)
		}

	}

}
