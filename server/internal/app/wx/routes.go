package wx

import (
	"github.com/ix-pay/ixpay-pro/internal/app/wx/middleware"
)

// setupRoutes 设置路由
func (a *AppWX) setupRoutes() {

	// 微信公众号路由组，添加/api/wx 前缀
	wx := a.router.Group("/api/wx")
	{
		// 公共路由
		public := wx
		{
			// 认证路由
			auth := public.Group("/auth")
			{
				auth.POST("/login", a.authController.LoginByCode)
				auth.POST("/refresh-token", a.authController.RefreshToken)
				auth.GET("/oauth-url", a.authController.GetOAuthURL)
				auth.GET("/callback", a.authController.OAuthCallback)
				auth.GET("/authorize", a.authController.Authorize)
			}
			// 支付通知路由（不需要认证）
			pay := public.Group("/pay")
			{
				pay.POST("/notify/wechat", a.paymentController.HandleWechatPayNotify)
				// JS-SDK 配置接口（不需要认证，用于前端初始化 wx.config()）
				pay.GET("/jsapi-config", a.paymentController.GetJSAPIConfig)
			}
		}

		// 需要认证的路由
		authenticated := wx
		authenticated.Use(middleware.AuthMiddleware(a.auth, a.logger))
		authenticated.Use(middleware.PermissionMiddleware(a.permissions))
		{
			// 支付路由
			payment := authenticated.Group("/payment")
			{
				payment.POST("", a.paymentController.CreatePayment)
				payment.POST("/unified-order", a.paymentController.CreateUnifiedOrder)
				payment.GET("/:id", a.paymentController.GetPayment)
				payment.GET("", a.paymentController.GetUserPayments)
				payment.PUT("/:id/cancel", a.paymentController.CancelPayment)
			}
		}
	}
}
