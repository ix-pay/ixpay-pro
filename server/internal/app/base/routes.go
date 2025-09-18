package base

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/middleware"
)

// setupRoutes 设置路由
func (a *AppBase) setupRoutes() {

	// 公共路由
	public := a.router.Group("/v1")
	{
		// 认证路由
		auth := public.Group("/auth")
		{
			auth.POST("/register", a.userController.Register)
			auth.POST("/login", a.authController.Login)
			auth.POST("/captcha", a.authController.Captcha)
			auth.POST("/refresh-token", a.authController.RefreshToken)
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
