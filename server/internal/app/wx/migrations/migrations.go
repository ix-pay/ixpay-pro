package migrations

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// MigrateDatabase 执行微信应用的数据库迁移
func MigrateDatabase(db *database.PostgresDB, log logger.Logger) {
	log.Info("开始执行微信应用数据库迁移")

	// 创建 payments 表
	createPaymentsTable := `
		CREATE TABLE IF NOT EXISTS wx_payments (
			id BIGINT PRIMARY KEY,
			order_id VARCHAR(50) NOT NULL,
			user_id BIGINT NOT NULL DEFAULT 0,
			amount BIGINT NOT NULL DEFAULT 0,
			currency VARCHAR(10) DEFAULT 'CNY',
			method VARCHAR(20) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			transaction_id VARCHAR(100),
			description VARCHAR(255),
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			deleted_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			paid_at TIMESTAMP,
			deleted_at TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_wx_payments_order_id ON wx_payments(order_id);
		CREATE INDEX IF NOT EXISTS idx_wx_payments_user_id ON wx_payments(user_id);
		CREATE INDEX IF NOT EXISTS idx_wx_payments_transaction_id ON wx_payments(transaction_id);
		CREATE INDEX IF NOT EXISTS idx_wx_payments_paid_at ON wx_payments(paid_at);
	`

	// 创建 wechat_pay_infos 表
	createWechatPayInfosTable := `
		CREATE TABLE IF NOT EXISTS wx_wechat_pay_infos (
			id BIGINT PRIMARY KEY,
			payment_id BIGINT NOT NULL DEFAULT 0,
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
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			deleted_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP,
			CONSTRAINT fk_payment FOREIGN KEY (payment_id) REFERENCES wx_payments(id) ON DELETE CASCADE
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS idx_wx_wechat_pay_infos_payment_id ON wx_wechat_pay_infos(payment_id);
	`

	// 创建 wx_users 表
	createWxUsersTable := `
		CREATE TABLE IF NOT EXISTS wx_users (
			id BIGINT PRIMARY KEY,
			open_id VARCHAR(100) NOT NULL UNIQUE,
			union_id VARCHAR(100) UNIQUE,
			nickname VARCHAR(100),
			avatar VARCHAR(255),
			gender INTEGER NOT NULL DEFAULT 0,
			country VARCHAR(50),
			province VARCHAR(50),
			city VARCHAR(50),
			language VARCHAR(20),
			subscribe BOOLEAN NOT NULL DEFAULT false,
			subscribe_time TIMESTAMP,
			remark VARCHAR(255),
			group_id BIGINT NOT NULL DEFAULT 0,
			user_id BIGINT NOT NULL DEFAULT 0,
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			deleted_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS idx_wx_users_open_id ON wx_users(open_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_wx_users_union_id ON wx_users(union_id);
		CREATE INDEX IF NOT EXISTS idx_wx_users_subscribe_time ON wx_users(subscribe_time);
		CREATE INDEX IF NOT EXISTS idx_wx_users_user_id ON wx_users(user_id);
	`

	// 创建 wx_auth_sessions 表
	createWxAuthSessionsTable := `
		CREATE TABLE IF NOT EXISTS wx_auth_sessions (
			id BIGINT PRIMARY KEY,
			wx_user_id BIGINT NOT NULL DEFAULT 0,
			access_token VARCHAR(255) NOT NULL,
			refresh_token VARCHAR(255) NOT NULL,
			expires_in BIGINT NOT NULL DEFAULT 0,
			scope VARCHAR(255),
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			deleted_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '1 hour',
			deleted_at TIMESTAMP,
			CONSTRAINT fk_wx_user FOREIGN KEY (wx_user_id) REFERENCES wx_users(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_wx_auth_sessions_wx_user_id ON wx_auth_sessions(wx_user_id);
		CREATE INDEX IF NOT EXISTS idx_wx_auth_sessions_expires_at ON wx_auth_sessions(expires_at);
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
