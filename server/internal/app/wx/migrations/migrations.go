package migrations

import (
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/modules"
	"gorm.io/gorm"
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
			order_type VARCHAR(20) DEFAULT 'jsapi',
			refunded_fee BIGINT NOT NULL DEFAULT 0,
			openid VARCHAR(100) DEFAULT '',
			code_url VARCHAR(255) DEFAULT '',
			trade_state VARCHAR(30) DEFAULT '',
			mix_trade_no VARCHAR(100) DEFAULT '',
			med_info TEXT,
			wx_last_response TEXT,
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
		CREATE INDEX IF NOT EXISTS idx_wx_payments_order_type ON wx_payments(order_type);
		CREATE INDEX IF NOT EXISTS idx_wx_payments_openid ON wx_payments(openid);
		CREATE INDEX IF NOT EXISTS idx_wx_payments_mix_trade_no ON wx_payments(mix_trade_no);
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

	// 创建 wx_refunds 表（普通退款单）
	createRefundsTable := `
		CREATE TABLE IF NOT EXISTS wx_refunds (
			id BIGINT PRIMARY KEY,
			out_refund_no VARCHAR(64) NOT NULL,
			out_trade_no VARCHAR(100) NOT NULL,
			refund_id VARCHAR(64),
			reason VARCHAR(255),
			refund_fee BIGINT NOT NULL DEFAULT 0,
			total_fee BIGINT NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'PROCESSING',
			applied BOOLEAN NOT NULL DEFAULT false,
			user_received_account VARCHAR(128),
			success_time VARCHAR(32),
			wx_last_response TEXT,
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_wx_refunds_out_refund_no ON wx_refunds(out_refund_no);
		CREATE INDEX IF NOT EXISTS idx_wx_refunds_out_trade_no ON wx_refunds(out_trade_no);
		CREATE INDEX IF NOT EXISTS idx_wx_refunds_refund_id ON wx_refunds(refund_id);
		CREATE INDEX IF NOT EXISTS idx_wx_refunds_status ON wx_refunds(status);
	`

	// 创建 wx_med_refunds 表（医保退款单）
	createMedRefundsTable := `
		CREATE TABLE IF NOT EXISTS wx_med_refunds (
			id BIGINT PRIMARY KEY,
			hosp_out_refund_no VARCHAR(64) NOT NULL,
			out_trade_no VARCHAR(100) NOT NULL,
			part_refund_type VARCHAR(30),
			cash_refund_fee BIGINT NOT NULL DEFAULT 0,
			ins_refund_fee BIGINT NOT NULL DEFAULT 0,
			request_content TEXT,
			reason VARCHAR(255),
			med_refund_id VARCHAR(64),
			cash_refund_id VARCHAR(64),
			insurance_refund_id VARCHAR(64),
			med_refund_state VARCHAR(20),
			cash_refund_status VARCHAR(20),
			insurance_refund_status VARCHAR(20),
			refund_end_time VARCHAR(32),
			response_content TEXT,
			cash_applied BOOLEAN NOT NULL DEFAULT false,
			ins_applied BOOLEAN NOT NULL DEFAULT false,
			last_response TEXT,
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_wx_med_refunds_hosp_out_refund_no ON wx_med_refunds(hosp_out_refund_no);
		CREATE INDEX IF NOT EXISTS idx_wx_med_refunds_out_trade_no ON wx_med_refunds(out_trade_no);
		CREATE INDEX IF NOT EXISTS idx_wx_med_refunds_cash_refund_id ON wx_med_refunds(cash_refund_id);
	`

	// 创建 wx_app_configs / wx_merchant_configs / wx_app_merchants 多渠道配置表
	// 审计列（created_by/updated_by/deleted_at）对齐 SnowflakeBaseModel 持久化模型
	createWxChannelTables := `
		CREATE TABLE IF NOT EXISTS wx_app_configs (
			id BIGINT PRIMARY KEY,
			app_key VARCHAR(50) NOT NULL,
			app_type VARCHAR(20) NOT NULL DEFAULT 'mp',
			app_id VARCHAR(50) NOT NULL,
			app_secret TEXT,
			auth_scope VARCHAR(20) DEFAULT 'snsapi_base',
			oauth_callback_url VARCHAR(255) DEFAULT '',
			frontend_url VARCHAR(255) DEFAULT '',
			description VARCHAR(255) DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_wx_app_configs_app_key ON wx_app_configs(app_key);
		CREATE INDEX IF NOT EXISTS idx_wx_app_configs_status ON wx_app_configs(status);

		CREATE TABLE IF NOT EXISTS wx_merchant_configs (
			id BIGINT PRIMARY KEY,
			mch_id VARCHAR(32) NOT NULL,
			mch_name VARCHAR(100) DEFAULT '',
			mch_serial_no VARCHAR(64) DEFAULT '',
			apiv3_key TEXT,
			private_key TEXT,
			public_key_id TEXT DEFAULT '',
			public_key TEXT,
			pay_notify_url VARCHAR(255) DEFAULT '',
			refund_notify_url VARCHAR(255) DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_wx_merchant_configs_mch_id ON wx_merchant_configs(mch_id);
		CREATE INDEX IF NOT EXISTS idx_wx_merchant_configs_status ON wx_merchant_configs(status);

		CREATE TABLE IF NOT EXISTS wx_app_merchants (
			id BIGINT PRIMARY KEY,
			app_config_id BIGINT NOT NULL,
			merchant_config_id BIGINT NOT NULL,
			is_default BOOLEAN NOT NULL DEFAULT false,
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_wx_app_merchants_app ON wx_app_merchants(app_config_id);
		CREATE INDEX IF NOT EXISTS idx_wx_app_merchants_mch ON wx_app_merchants(merchant_config_id);

		CREATE TABLE IF NOT EXISTS tenant_ui_configs (
			id BIGINT PRIMARY KEY,
			app_config_id BIGINT NOT NULL,
			ui_json TEXT,
			status INTEGER NOT NULL DEFAULT 1,
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_tenant_ui_configs_app ON tenant_ui_configs(app_config_id);
	`

	// 执行SQL语句
	sqlStatements := []string{
		createPaymentsTable,
		createWechatPayInfosTable,
		createWxUsersTable,
		createWxAuthSessionsTable,
		createRefundsTable,
		createMedRefundsTable,
		createWxChannelTables,
	}

	for _, sql := range sqlStatements {
		execResult := db.Exec(sql)
		if execResult.Error != nil {
			log.Error("执行迁移 SQL 失败", "error", execResult.Error)
			return
		}
	}

	log.Info("微信应用数据库迁移完成")
}

// TenantMigrate 实现多租户模块契约（infrastructure/support/modules.TenantModule）
// 在租户库执行 wx 模块全部业务表迁移（幂等可重复执行）
func TenantMigrate(db *gorm.DB, log logger.Logger) error {
	pg := &database.PostgresDB{DB: db}
	MigrateDatabase(pg, log)
	return nil
}

// TenantSeed 实现多租户模块契约：租户库种子数据（wx 模块当前无需种子）
func TenantSeed(db *gorm.DB, log logger.Logger) error {
	log.Info("wx 模块租户库种子：无")
	return nil
}

// wxTenantModule wx 模块的租户契约实现（适配 infrastructure/support/modules.TenantModule）
type wxTenantModule struct{}

func (wxTenantModule) Name() string { return "wx" }

func (wxTenantModule) TenantMigrate(db *gorm.DB, log logger.Logger) error {
	return TenantMigrate(db, log)
}

func (wxTenantModule) TenantSeed(db *gorm.DB, log logger.Logger) error {
	return TenantSeed(db, log)
}

// RegisterTenantModule 注册 wx 模块到租户模块注册表（应用装配时调用一次；
// 租户库初始化任务遍历注册表执行各模块迁移）
func RegisterTenantModule() {
	modules.RegisterModule(wxTenantModule{})
}
