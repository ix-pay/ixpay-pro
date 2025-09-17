package database

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
)

// Migrate 执行数据库迁移，创建所有表
// 使用原始SQL语句而非AutoMigrate函数，以避免GORM参数绑定问题
func (db *PostgresDB) Migrate(log logger.Logger) error {
	log.Info("Starting database migration using raw SQL")

	// 1. 创建users表
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
		wechat_openid VARCHAR(100) UNIQUE,
		wechat_unionid VARCHAR(100) UNIQUE,
		status INTEGER DEFAULT 1,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	`

	if err := db.Exec(createUsersTableSQL).Error; err != nil {
		log.Error("Failed to create users table", "error", err)
		return err
	}
	log.Info("Users table created successfully")

	// 2. 创建payments表
	createPaymentsTableSQL := `
	CREATE TABLE IF NOT EXISTS payments (
		id SERIAL PRIMARY KEY,
		order_id VARCHAR(50) NOT NULL,
		user_id INTEGER NOT NULL,
		amount BIGINT NOT NULL,
		currency VARCHAR(10) DEFAULT 'CNY',
		method VARCHAR(20) NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		transaction_id VARCHAR(100),
		description VARCHAR(255),
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		paid_at TIMESTAMP,
		CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	if err := db.Exec(createPaymentsTableSQL).Error; err != nil {
		log.Error("Failed to create payments table", "error", err)
		return err
	}
	log.Info("Payments table created successfully")

	// 3. 创建wechat_pay_infos表
	createWechatPayInfosTableSQL := `
	CREATE TABLE IF NOT EXISTS wechat_pay_infos (
		id SERIAL PRIMARY KEY,
		payment_id INTEGER NOT NULL UNIQUE,
		app_id VARCHAR(50),
		mch_id VARCHAR(50),
		nonce_str VARCHAR(50),
		prepay_id VARCHAR(100),
		code_url VARCHAR(255),
		sign VARCHAR(255),
		timestamp VARCHAR(20),
		package VARCHAR(50),
		pay_sign VARCHAR(255),
		return_code VARCHAR(20),
		return_msg VARCHAR(255),
		result_code VARCHAR(20),
		err_code VARCHAR(20),
		err_code_des VARCHAR(255),
		notify_data TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		CONSTRAINT fk_payment FOREIGN KEY (payment_id) REFERENCES payments(id)
	);
	`

	if err := db.Exec(createWechatPayInfosTableSQL).Error; err != nil {
		log.Error("Failed to create wechat_pay_infos table", "error", err)
		return err
	}
	log.Info("Wechat_pay_infos table created successfully")

	log.Info("All database tables created successfully with raw SQL")
	return nil
}
