package base

import (
	"github.com/ix-pay/ixpay-pro/internal/app/base/middleware"
)

// setupRoutes 设置路由
func (a *AppBase) setupRoutes() {

	// 管理后台路由组，添加/api/admin 前缀（保持向后兼容）
	admin := a.router.Group("/api/admin")
	{
		// 公共路由
		public := admin
		{
			// 认证路由
			auth := public.Group("/auth")
			{
				auth.POST("/register", a.userController.Register)
				auth.POST("/login", a.authController.Login)
				auth.POST("/captcha", a.authController.Captcha)
			}
		}

		// 需要认证的路由
		authenticated := admin
		authenticated.Use(middleware.AuthMiddleware(a.auth, a.cache, a.logger))
		authenticated.Use(middleware.PermissionMiddleware(a.permissionService, a.roleRepo, a.logger, a.cache))
		{
			// 认证相关路由（需要认证）
			auth := authenticated.Group("/auth")
			{
				auth.POST("/refresh-token", a.authController.RefreshToken)
				auth.POST("/logout", a.authController.Logout)
			}
			// 用户路由
			user := authenticated.Group("/user")
			{
				user.GET("/info", a.userController.GetUserInfo)
				user.PUT("/info", a.userController.UpdateUserInfo)
				user.GET("", a.userController.GetUserList)
				user.POST("", a.userController.AddUser)
				user.DELETE("/:id", a.userController.DeleteUser)
				user.PUT("/password", a.userController.ChangePassword)
				user.PUT("/reset-password", a.userController.ResetPassword)
				user.GET("/getSelfSetting", a.userController.GetSelfSetting)
				user.PUT("/setSelfSetting", a.userController.SetSelfSetting)
				user.POST("/switch-role", a.userController.SwitchRole)
			}

			// 角色路由
			role := authenticated.Group("/role")
			{
				role.POST("", a.roleController.CreateRole)
				role.GET("/detail", a.roleController.GetRoleByID)
				role.GET("/:id/detail", a.roleController.GetRoleDetail)
				role.GET("/:id/available-apis", a.roleController.GetAvailableAPIs)
				role.PUT("", a.roleController.UpdateRole)
				role.DELETE("", a.roleController.DeleteRole)
				role.GET("", a.roleController.GetRoleList)
				role.GET("/all", a.roleController.GetAllRoles)
				role.POST("/assign-users", a.roleController.AssignUserToRole)
				role.POST("/assign-menus", a.roleController.AssignMenuToRole)
				role.POST("/assign-api-routes", a.roleController.AssignAPIToRole)
			}

			// 任务路由（需要 admin 角色）
			task := authenticated.Group("/task")
			{
				task.POST("", a.taskController.AddTask)
				task.DELETE("/:id", a.taskController.RemoveTask)
				task.POST("/:id/start", a.taskController.StartTask)
				task.POST("/:id/stop", a.taskController.StopTask)
				task.POST("/:id/retry", a.taskController.RetryTask)
				task.GET("", a.taskController.GetTasks)
				task.GET("/:id", a.taskController.GetTask)
				// 新增任务执行日志和统计路由
				task.GET("/:id/execution-logs", a.taskController.GetExecutionLogs)
				task.GET("/statistics", a.taskController.GetStatistics)
				task.POST("/:id/group", a.taskController.SetTaskGroup)
			}

			// API路由管理（需要admin角色）
			apis := authenticated.Group("/apis")
			{
				// 分页获取API路由列表
				apis.GET("", a.apiController.GetAPIList)
				// 根据ID获取API路由
				apis.GET("/:id", a.apiController.GetRouteByID)
				// 创建API路由
				apis.POST("", a.apiController.CreateAPI)
				// 更新API路由
				apis.PUT("/:id", a.apiController.UpdateAPI)
				// 删除API路由
				apis.DELETE("/:id", a.apiController.DeleteAPI)
			}

			// 菜单管理路由
			menu := authenticated.Group("/menu")
			{
				menu.GET("", a.menuController.GetMenuList)
				menu.POST("", a.menuController.AddMenu)
				menu.PUT("", a.menuController.UpdateMenu)
				menu.DELETE("/:id", a.menuController.DeleteMenu)
				menu.GET("/page", a.menuController.GetMenuPage)
				menu.GET("/tree", a.menuController.GetMenuTree)
			}

			// 按钮权限管理路由
			btnPerm := authenticated.Group("/btn-perms")
			{
				btnPerm.POST("", a.btnPermController.CreateBtnPerm)
				btnPerm.GET("/detail", a.btnPermController.GetBtnPermByID)
				btnPerm.PUT("", a.btnPermController.UpdateBtnPerm)
				btnPerm.DELETE("", a.btnPermController.DeleteBtnPerm)
				btnPerm.GET("", a.btnPermController.GetBtnPermList)
				btnPerm.POST("/assign-api-routes", a.btnPermController.AssignToBtnPerm)
				btnPerm.POST("/revoke-api-route", a.btnPermController.RevokeFromBtnPerm)
				btnPerm.POST("/assign-to-role", a.btnPermController.AssignBtnPermToRole)
				btnPerm.POST("/revoke-from-role", a.btnPermController.RevokeBtnPermFromRole)
				btnPerm.GET("/api-routes", a.btnPermController.GetAPIRoutesByBtnPerm)
				btnPerm.GET("/for-route", a.btnPermController.GetBtnPermsByAPIRoute)
				btnPerm.GET("/by-role", a.btnPermController.GetBtnPermsByRole)
				btnPerm.GET("/by-menu", a.btnPermController.GetBtnPermsByMenu)
			}

			// 配置管理路由
			config := authenticated.Group("/config")
			{
				// 获取配置列表
				config.GET("", a.configController.GetConfigList)
				// 获取单个配置（通过键）
				config.GET("/key", a.configController.GetConfigByKey)
				// 获取单个配置（通过ID）
				config.GET("/:id", a.configController.GetConfigByID)
				// 创建配置
				config.POST("", a.configController.CreateConfig)
				// 更新配置
				config.PUT("", a.configController.UpdateConfig)
				// 删除配置
				config.DELETE("/:id", a.configController.DeleteConfig)
				// 获取所有启用的配置（需要认证）
				config.GET("/active", a.configController.GetAllActiveConfigs)
			}

			// 字典管理路由
			dict := authenticated.Group("/dict")
			{
				// 字典表相关接口
				dict.GET("", a.dictController.GetDictList)
				dict.GET("/code", a.dictController.GetDictByCode)
				dict.GET("/:id", a.dictController.GetDictByID)
				dict.POST("", a.dictController.CreateDict)
				dict.PUT("", a.dictController.UpdateDict)
				dict.DELETE("/:id", a.dictController.DeleteDict)

				// 字典项相关接口
				dict.GET("/item/:id", a.dictController.GetDictItemByID)
				dict.GET("/items", a.dictController.GetDictItemsByDictID)
				dict.POST("/item", a.dictController.CreateDictItem)
				dict.PUT("/item", a.dictController.UpdateDictItem)
				dict.DELETE("/item/:id", a.dictController.DeleteDictItem)
			}

			// 操作日志路由
			logs := authenticated.Group("/logs")
			{
				logs.GET("", a.operationLogController.GetLogList)
				logs.GET("/:id", a.operationLogController.GetLogByID)
				logs.DELETE("/:id", a.operationLogController.DeleteLogByID)
				logs.POST("/batch-delete", a.operationLogController.BatchDeleteLog)
				logs.GET("/statistics", a.operationLogController.GetLogStatistics)
				logs.POST("/clear", a.operationLogController.ClearLogByTimeRange)
			}

			// 部门管理路由
			dept := authenticated.Group("/dept")
			{
				// 获取部门列表
				dept.GET("", a.departmentController.GetDepartmentList)
				// 获取部门树形结构
				dept.GET("/tree", a.departmentController.GetDepartmentTree)
				// 获取部门详情
				dept.GET("/:id", a.departmentController.GetDepartmentByID)
				// 创建部门
				dept.POST("", a.departmentController.CreateDepartment)
				// 更新部门
				dept.PUT("", a.departmentController.UpdateDepartment)
				// 删除部门
				dept.DELETE("/:id", a.departmentController.DeleteDepartment)
				// 更新部门负责人
				dept.PUT("/:id/leader", a.departmentController.UpdateDepartmentLeader)
			}

			// 岗位管理路由
			position := authenticated.Group("/position")
			{
				// 获取岗位列表
				position.GET("", a.positionController.GetPositionList)
				// 获取所有岗位
				position.GET("/all", a.positionController.GetAllPositions)
				// 获取岗位详情
				position.GET("/:id", a.positionController.GetPositionByID)
				// 创建岗位
				position.POST("", a.positionController.CreatePosition)
				// 更新岗位
				position.PUT("", a.positionController.UpdatePosition)
				// 删除岗位
				position.DELETE("/:id", a.positionController.DeletePosition)
			}

			// 公告管理路由
			notice := authenticated.Group("/notices")
			{
				// 获取公告列表
				notice.GET("", a.noticeController.GetNoticeList)
				// 获取公告详情
				notice.GET("/:id", a.noticeController.GetNoticeByID)
				// 创建公告
				notice.POST("", a.noticeController.CreateNotice)
				// 更新公告
				notice.PUT("", a.noticeController.UpdateNotice)
				// 删除公告
				notice.DELETE("/:id", a.noticeController.DeleteNotice)
				// 发布公告
				notice.POST("/:id/publish", a.noticeController.PublishNotice)
				// 标记公告已读
				notice.POST("/:id/read", a.noticeController.MarkAsRead)
				// 检查是否已读
				notice.GET("/:id/is-read", a.noticeController.CheckIsRead)
				// 获取公告统计
				notice.GET("/statistics", a.noticeController.GetStatistics)
			}

			// 登录日志管理路由
			loginLog := authenticated.Group("/login-log")
			{
				// 获取登录日志列表
				loginLog.GET("", a.loginLogController.GetLoginLogList)
				// 获取登录日志详情
				loginLog.GET("/:id", a.loginLogController.GetLoginLogByID)
				// 获取登录统计
				loginLog.GET("/statistics", a.loginLogController.GetStatistics)
				// 获取异常登录查询
				loginLog.GET("/abnormal", a.loginLogController.GetAbnormalLogins)
				// 记录登录日志（内部调用）
				loginLog.POST("", a.loginLogController.RecordLogin)
			}

			// 在线用户管理路由
			onlineUser := authenticated.Group("/online-user")
			{
				// 获取在线用户列表
				onlineUser.GET("", a.onlineUserController.GetOnlineUserList)
				// 获取在线用户详情
				onlineUser.GET("/:user_id", a.onlineUserController.GetOnlineUserByID)
				// 获取在线用户数量
				onlineUser.GET("/count", a.onlineUserController.GetOnlineCount)
				// 检查用户是否在线
				onlineUser.GET("/online", a.onlineUserController.IsOnline)
				// 强制用户下线
				onlineUser.DELETE("/:user_id", a.onlineUserController.ForceOffline)
				// 批量强制用户下线
				onlineUser.POST("/batch", a.onlineUserController.BatchForceOffline)
			}

			// 系统监控路由
			monitor := authenticated.Group("/monitor")
			{
				// 获取系统监控信息
				monitor.GET("/system", a.monitorController.GetSystemMonitor)
				// 获取缓存监控信息
				monitor.GET("/cache", a.monitorController.GetCacheMonitor)
				// 获取数据库监控信息
				monitor.GET("/database", a.monitorController.GetDatabaseMonitor)
				// 查询 Redis 键
				monitor.GET("/redis-keys", a.monitorController.GetRedisKeys)
				// 查询慢查询日志
				monitor.GET("/slow-queries", a.monitorController.GetSlowQueries)
			}
		}
	}
}
