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
				auth.POST("/miniprogram-login", a.authController.MiniappLogin)
				auth.POST("/refresh-token", a.authController.RefreshToken)
				auth.GET("/oauth-url", a.authController.GetOAuthURL)
				auth.GET("/callback", a.authController.OAuthCallback)
				auth.GET("/authorize", a.authController.Authorize)
			}
			// 回调通知路由（不需要认证，由微信服务端调用）
			pay := public.Group("/pay")
			{
				pay.POST("/notify/wechat", a.paymentController.HandleWechatPayNotify)
				pay.POST("/notify/refund", a.refundController.HandleRefundNotify)
				// JS-SDK 配置接口（不需要认证，用于前端初始化 wx.config()）
				pay.GET("/jsapi-config", a.paymentController.GetJSAPIConfig)
			}
			// 医保收款回调（不需要认证，由微信服务端调用）
			med := public.Group("/med")
			{
				med.POST("/notify", a.medicalController.HandleMedNotify)
			}
			// 公开应用配置（前端空壳启动加载，无需认证）
			configGroup := public.Group("/config")
			{
				configGroup.GET("/app", a.channelController.GetPublicApp)
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
				payment.POST("/combine-prepay", a.paymentController.CombinePrepay)
				payment.POST("/native-prepay", a.paymentController.NativePrepay)
				payment.GET("/native/qrcode", a.billController.NativeQRCode)
				payment.GET("/query", a.paymentController.WxQuery)
				payment.GET("/:id", a.paymentController.GetPayment)
				payment.GET("", a.paymentController.GetUserPayments)
				payment.PUT("/:id/cancel", a.paymentController.CancelPayment)
			}

			// 退款路由
			refund := authenticated.Group("/refund")
			{
				refund.POST("/create", a.refundController.Create)
				refund.GET("/query", a.refundController.Query)
				refund.GET("/list", a.refundController.List)
				refund.POST("/abnormal", a.refundController.ApplyAbnormal)
				// 医保退款
				refund.POST("/med", a.refundController.MedCreate)
				refund.GET("/med/query", a.refundController.MedQuery)
				refund.GET("/med/list", a.refundController.MedList)
			}

			// 医保支付路由
			med := authenticated.Group("/med")
			{
				med.GET("/auth/url", a.medicalController.GetMedAuthURL)
				med.POST("/auth/query", a.medicalController.MedAuthQuery)
				med.POST("/orders", a.medicalController.CreateMedOrder)
				med.POST("/prepay", a.medicalController.MedPrepay)
				med.GET("/query", a.medicalController.MedQuery)
			}

			// 账单路由
			bill := authenticated.Group("/bill")
			{
				bill.POST("/trade", a.billController.TradeBill)
				bill.POST("/fundflow", a.billController.FundFlowBill)
				bill.POST("/download", a.billController.DownloadBill)
			}

			// 说明：管理平台使用的渠道配置接口已迁移至 base 模块（/api/admin/wx/*），
			// 鉴权统一由 base 管理端负责；wx 模块仅保留提供给 wxapp/miniapp 前端的接口
		}
	}
}
