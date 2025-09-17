package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// PostgresDB 提供PostgreSQL数据库连接
type PostgresDB struct {
	*gorm.DB
}

// NewPostgresDB 创建新的PostgreSQL数据库连接
func NewPostgresDB(cfg *config.Config, log logger.Logger) (*PostgresDB, error) {
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

	log.Info("PostgreSQL database connected successfully")

	return &PostgresDB{DB: db},
		nil
}
