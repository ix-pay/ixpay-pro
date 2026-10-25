package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// PostgresDB 提供PostgreSQL数据库连接（支持多租户库路由）
// WithTenant(ctx) 按 ctx 中的租户ID返回对应租户库 GORM 实例（默认租户返回主库实例）
type PostgresDB struct {
	*gorm.DB
	TenantManager *TenantManager
}

// NewPostgresDB 创建新的PostgreSQL数据库连接
func SetupPostgresDB(cfg *config.Config, log logger.Logger) (*PostgresDB, error) {
	// 构建连接字符串
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	// 打开SQL连接

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("Failed to open database connection", "error", err)
		return nil, err
	}

	// 配置连接池

	sqlDB.SetMaxOpenConns(cfg.Postgres.MaxOpenConns)

	sqlDB.SetMaxIdleConns(cfg.Postgres.MaxIdleConns)

	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Postgres.ConnMaxLifetime) * time.Minute)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		log.Error("Failed to ping database", "error", err)
		return nil, err
	}

	// 创建GORM数据库实例，设置详细日志模式

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		log.Error("Failed to initialize GORM", "error", err)
		return nil, err
	}

	// 全局审计字段保护：Save 全字段更新时零值 created_at 不覆盖库中原值
	RegisterAuditPreserveHooks(db)

	// 租户连接管理器（多租户库路由）
	manager := NewTenantManager(sqlDB, cfg, log)
	manager.StartIdleRecycler()

	log.Info("PostgreSQL database connected successfully")

	return &PostgresDB{DB: db, TenantManager: manager},
		nil
}

// TenantGorm 返回租户库 GORM 实例（多租户业务数据路由入口）
// ctx 无租户上下文或默认租户（dbname 空）时返回主库实例
func (p *PostgresDB) TenantGorm(ctx context.Context) (*gorm.DB, error) {
	tenantID, ok := TenantIDFromContext(ctx)
	if !ok {
		return p.DB, nil
	}
	gdb, err := p.TenantManager.GormDB(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if gdb == nil {
		// 默认租户（dbname 空）→ 主库
		return p.DB, nil
	}
	return gdb, nil
}
