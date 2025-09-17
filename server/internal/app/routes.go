package app

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/middleware"

	"github.com/gin-gonic/gin"
)

// setupRoutes 设置路由
func (a *Application) setupRoutes() {

	// 健康检查
	a.router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ixpay-pro"})
	})

	// 健康检查
	a.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	// 公共路由
	public := a.router.Group("/v1")
	{
		// 认证路由
		auth := public.Group("/auth")
		{
			auth.POST("/register", a.userController.Register)
			auth.POST("/login", a.userController.Login)
			auth.POST("/captcha", a.userController.Captcha)
			auth.POST("/wechat-login", a.userController.WechatLogin)
			auth.POST("/refresh-token", a.userController.RefreshToken)
		}

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
		// 用户路由
		user := authenticated.Group("/user")
		{
			user.GET("/info", a.userController.GetUserInfo)
		}

		// 支付路由
		payment := authenticated.Group("/payment")
		{
			payment.POST("", a.paymentController.CreatePayment)
			payment.GET("/:id", a.paymentController.GetPayment)
			payment.GET("", a.paymentController.GetUserPayments)
			payment.PUT("/:id/cancel", a.paymentController.CancelPayment)
		}

		// 任务路由（需要admin角色）
		task := authenticated.Group("/task")
		{
			task.POST("", a.taskController.AddTask)
			task.DELETE("/:id", a.taskController.RemoveTask)
			task.POST("/:id/start", a.taskController.StartTask)
			task.POST("/:id/stop", a.taskController.StopTask)
			task.POST("/:id/retry", a.taskController.RetryTask)
			task.GET("", a.taskController.GetTasks)
			task.GET("/:id", a.taskController.GetTask)
		}
	}

}
