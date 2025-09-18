package base

import (
	"time"

	baseapi "github.com/ix-pay/ixpay-pro/internal/app/base/api/v1"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/app/base/migrations"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// AppBase 应用程序结构
type AppBase struct {
	router         *gin.Engine
	db             *database.PostgresDB
	auth           *auth.JWTAuth
	permissions    *auth.PermissionManager
	logger         logger.Logger
	authController *baseapi.AuthController
	userController *baseapi.UserController
	taskController *baseapi.TaskController
	userRepo       model.UserRepository
}

// NewAppBase 创建应用程序实例
func NewAppBase(
	log logger.Logger,
	db *database.PostgresDB,
	auth *auth.JWTAuth,
	permissions *auth.PermissionManager,
	authController *baseapi.AuthController,
	userController *baseapi.UserController,
	taskController *baseapi.TaskController,
	userRepo model.UserRepository,
) (*AppBase, error) {
	// 执行数据库迁移，创建所有需要的表
	// if err := db.Migrate(log); err != nil {
	// 	log.Error("Failed to migrate database", "error", err)
	// 	return nil, err
	// }

	// 创建应用实例
	app := &AppBase{
		db:             db,
		auth:           auth,
		permissions:    permissions,
		logger:         log,
		authController: authController,
		userController: userController,
		taskController: taskController,
		userRepo:       userRepo,
	}
	return app, nil
}

// setupMiddleware 设置中间件
func (a *AppBase) setupMiddleware() {
}

// initializePermissions 初始化权限
func (a *AppBase) initializePermissions() {
	// 定义系统权限
	permissions := []auth.Permission{
		// 用户路由权限
		{Path: "/api/v1/user/info", Method: "GET", Roles: []string{"user", "admin"}, WechatGrant: true},

		// 任务路由权限（需要admin角色）
		{Path: "/api/v1/task", Method: "POST", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id", Method: "DELETE", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id/start", Method: "POST", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id/stop", Method: "POST", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id/retry", Method: "POST", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task", Method: "GET", Roles: []string{"admin"}, WechatGrant: false},
		{Path: "/api/v1/task/:id", Method: "GET", Roles: []string{"admin"}, WechatGrant: false},
	}

	// 缓存权限数据
	if err := a.permissions.CachePermissions(permissions); err != nil {
		a.logger.Error("Failed to cache permissions", "error", err)
	}
}

// initializeSeedData 初始化种子数据
func (a *AppBase) initializeSeedData() {
	a.logger.Info("初始化种子数据")

	// 检查admin用户是否存在
	_, err := a.userRepo.GetByUsername("admin")
	if err == nil {
		a.logger.Info("Admin用户已存在，跳过创建")
		return
	}

	// admin用户不存在，创建admin用户
	// 创建密码哈希（使用默认密码'admin123'）
	passwordHash := "$argon2id$v=19$m=65536,t=1,p=2$K0VYVVRtWXhkVXk3$vW5IqE4+ZgXc0V0t9tGtPq2sLzXjK5Yg" // admin123的哈希

	adminUser := &model.User{
		Username:     "admin",
		PasswordHash: passwordHash,
		Nickname:     "系统管理员",
		Email:        "admin@ixpay.com",
		Role:         "admin",
		Status:       1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存admin用户
	if err := a.userRepo.Create(adminUser); err != nil {
		a.logger.Error("创建admin用户失败", "error", err)
	} else {
		a.logger.Info("创建admin用户成功")
	}
}

// 初始化基础应用
func (a *AppBase) Init(router *gin.Engine) {
	a.logger.Info("初始化基础应用")
	a.router = router

	// 执行数据库迁移
	migrations.MigrateDatabase(a.db, a.logger)

	// 设置中间件
	a.setupMiddleware()

	// 设置路由
	a.setupRoutes()

	// 初始化权限
	a.initializePermissions()

	// 初始化种子数据
	a.initializeSeedData()

}
