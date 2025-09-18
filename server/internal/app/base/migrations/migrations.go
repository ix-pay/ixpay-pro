package migrations

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
)

// MigrateDatabase 使用原始SQL语句执行base应用的数据库迁移
func MigrateDatabase(db *database.PostgresDB, log logger.Logger) {
	log.Info("开始执行base应用数据库迁移")

	// 创建users表
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(100) NOT NULL,
		nickname VARCHAR(50),
		email VARCHAR(100) UNIQUE,
		phone VARCHAR(20) UNIQUE,
		avatar VARCHAR(255),
		role VARCHAR(20) DEFAULT 'user',
		wechat_open_id VARCHAR(100) UNIQUE,
		wechat_union_id VARCHAR(100) UNIQUE,
		status INTEGER DEFAULT 1,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	`

	if err := db.Exec(createUsersTableSQL).Error; err != nil {
		log.Error("创建users表失败", "error", err)
	} else {
		log.Info("Users表创建成功")
	}

	// 可以在这里添加其他base应用相关的表迁移
	log.Info("base应用数据库迁移完成")
}
