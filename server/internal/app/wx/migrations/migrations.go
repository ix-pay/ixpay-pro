package migrations

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
)

// MigrateDatabase 执行微信应用的数据库迁移
func MigrateDatabase(db *database.PostgresDB, log logger.Logger) {
	log.Info("开始执行微信应用数据库迁移")

	// 创建payments表
	createPaymentsTable := `
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
			paid_at TIMESTAMP
		);
	`

	// 创建wechat_pay_infos表
	createWechatPayInfosTable := `
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
			"package" VARCHAR(50),
			pay_sign VARCHAR(255),
			return_code VARCHAR(20),
			return_msg VARCHAR(255),
			result_code VARCHAR(20),
			err_code VARCHAR(20),
			err_code_des VARCHAR(255),
			notify_data TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT fk_payment FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE
		);
	`

	// 创建wx_users表
	createWxUsersTable := `
		CREATE TABLE IF NOT EXISTS wx_users (
			id SERIAL PRIMARY KEY,
			open_id VARCHAR(100) NOT NULL UNIQUE,
			union_id VARCHAR(100) UNIQUE,
			nickname VARCHAR(100),
			avatar VARCHAR(255),
			gender INTEGER DEFAULT 0,
			country VARCHAR(50),
			province VARCHAR(50),
			city VARCHAR(50),
			language VARCHAR(20),
			subscribe BOOLEAN DEFAULT false,
			subscribe_time TIMESTAMP,
			remark VARCHAR(255),
			group_id BIGINT DEFAULT 0,
			user_id INTEGER,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`

	// 创建wx_auth_sessions表
	createWxAuthSessionsTable := `
		CREATE TABLE IF NOT EXISTS wx_auth_sessions (
			id SERIAL PRIMARY KEY,
			wx_user_id INTEGER NOT NULL,
			access_token VARCHAR(255) NOT NULL,
			refresh_token VARCHAR(255) NOT NULL,
			expires_in BIGINT NOT NULL,
			scope VARCHAR(255),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '1 hour',
			CONSTRAINT fk_wx_user FOREIGN KEY (wx_user_id) REFERENCES wx_users(id) ON DELETE CASCADE
		);
	`

	// 执行SQL语句
	sqlStatements := []string{
		createPaymentsTable,
		createWechatPayInfosTable,
		createWxUsersTable,
		createWxAuthSessionsTable,
	}

	for _, sql := range sqlStatements {
		execResult := db.Exec(sql)
		if execResult.Error != nil {
			log.Error("Failed to execute migration SQL", "error", execResult.Error)
			return
		}
	}

	log.Info("微信应用数据库迁移完成")
}
