package monitor

import (
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/redis/go-redis/v9"
)

// SetupSystemMonitor 创建系统监控服务实例
func SetupSystemMonitor() *SystemMonitor {
	return NewSystemMonitor()
}

// SetupCacheMonitor 创建缓存监控服务实例
// 参数:
// - redisClient: Redis 客户端实例
func SetupCacheMonitor(redisClient *redis.Client) *CacheMonitor {
	return NewCacheMonitor(redisClient)
}

// SetupDatabaseMonitor 创建数据库监控服务实例
// 参数:
// - db: PostgreSQL 数据库实例
// - config: 配置信息
// - log: 日志记录器
func SetupDatabaseMonitor(db *database.PostgresDB, config *config.Config, log logger.Logger) *DatabaseMonitor {
	// 从配置中获取慢查询阈值，默认 1000ms
	slowQueryThreshold := int64(1000)
	if config.Server.SlowQueryThreshold > 0 {
		slowQueryThreshold = config.Server.SlowQueryThreshold
	}

	return NewDatabaseMonitor(db, slowQueryThreshold)
}
