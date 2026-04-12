//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/ix-pay/ixpay-pro/internal/app/base"
	"github.com/ix-pay/ixpay-pro/internal/app/wx"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	redisClient "github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/redis"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/captcha"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/snowflake"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"

	"github.com/redis/go-redis/v9"
)

// 定义全局服务提供者集合
var GlobalServiceSet = wire.NewSet(
	// 基础设施层
	config.LoadConfig,
	logger.SetupMultiLogger,
	logger.SetupLogger,
	// 数据库
	database.SetupPostgresDB,
	// 缓存
	redisClient.SetupRedisClient,
	cache.SetupCache,
	// 雪花 ID
	snowflake.SetupSnowflake,
	// 验证码
	captcha.SetupCaptcha,
	// 认证
	auth.SetupJWTAuth,
	// 权限管理
	auth.SetupPermissionManager,
	// 任务管理
	task.SetupTaskManager,

	// 提供 Redis 客户端
	ProvideRedisClient,

	// 创建应用程序实例
	SetupApplication,
)

// ProvideRedisClient 提供 Redis 客户端
func ProvideRedisClient(redisClient *redisClient.RedisClient) *redis.Client {
	return redisClient.Client
}

var ProviderSet = wire.NewSet(
	// Monitor
	base.ProviderSetBaseMonitor,
	// Repo
	base.ProviderSetBaseRepo,
	wx.ProviderSetWXRepo,
	// Service
	base.ProviderSetBaseService,
	wx.ProviderSetWXService,
	// Converter
	base.ProviderSetBaseConverter,
	// Controller
	base.ProviderSetBaseController,
	wx.ProviderSetWXController,
	// App
	base.ProviderSetBaseApp,
	wx.ProviderSetWXApp,
)

// 初始化完整应用
func InitializeApp() (*Application, error) {
	panic(wire.Build(
		GlobalServiceSet,
		// 应用
		ProviderSet,
	))
}
