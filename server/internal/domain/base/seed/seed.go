package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// Seed 种子数据接口
type Seed interface {
	// Version 返回种子数据版本
	Version() string
	// Name 返回种子数据名称
	Name() string
	// Init 初始化种子数据
	Init(db *database.PostgresDB, logger logger.Logger) error
	// Order 返回初始化顺序
	Order() int
}
