package migrations

import (
	"database/sql"
	"fmt"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// CreateDatabaseIfNotExistsParams 创建数据库参数
type CreateDatabaseIfNotExistsParams struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// CreateDatabaseIfNotExists 如果数据库不存在则创建
func CreateDatabaseIfNotExists(params CreateDatabaseIfNotExistsParams, log logger.Logger) error {
	// 构建连接到postgres数据库的连接字符串（而不是目标数据库）
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		params.Host,
		params.Port,
		params.User,
		params.Password,
		params.SSLMode,
	)

	// 连接到 postgres 数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("连接 PostgreSQL 数据库失败（用于创建数据库）", "error", err)
		return fmt.Errorf("PostgreSQL 连接失败：%w", err)
	}
	defer db.Close()

	// 检查数据库是否存在
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, params.DBName).Scan(&exists)
	if err != nil {
		log.Error("检查数据库是否存在失败", "error", err)
		return fmt.Errorf("检查数据库是否存在失败：%w", err)
	}

	// 如果数据库不存在，则创建
	if !exists {
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s", params.DBName)
		log.Info("正在创建数据库", "db_name", params.DBName)
		_, err = db.Exec(createDBQuery)
		if err != nil {
			log.Error("创建数据库失败", "error", err, "db_name", params.DBName)
			return fmt.Errorf("创建数据库失败：%w", err)
		}
		log.Info("数据库创建成功", "db_name", params.DBName)
	} else {
		log.Info("数据库已存在", "db_name", params.DBName)
	}

	return nil
}

// MigrateDatabase 使用原始SQL语句执行base应用的数据库迁移
func MigrateDatabase(db *database.PostgresDB, log logger.Logger) {
	log.Info("开始执行base应用数据库迁移")

	// 创建机构表
	createOrganizationsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_organizations (
		id BIGINT PRIMARY KEY,
		code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		contact_person VARCHAR(50),
		contact_phone VARCHAR(20),
		status INTEGER DEFAULT 1,
		database_type VARCHAR(20) DEFAULT 'postgres',
		database_host VARCHAR(100),
		database_port VARCHAR(10),
		database_name VARCHAR(100),
		database_user VARCHAR(50),
		database_password VARCHAR(100),
		database_ssl_mode VARCHAR(20) DEFAULT 'disable',
		max_connections INTEGER DEFAULT 20,
		idle_connections INTEGER DEFAULT 10,
		connection_timeout INTEGER DEFAULT 60,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	`

	if err := db.Exec(createOrganizationsTableSQL).Error; err != nil {
		log.Error("创建base_organizations表失败", "error", err)
	} else {
		log.Info("base_organizations表创建成功")
	}

	// 创建users表
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS base_users (
		id BIGINT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(100) NOT NULL,
		nickname VARCHAR(50),
		email VARCHAR(100),
		phone VARCHAR(20),
		avatar VARCHAR(255),
		wechat_open_id VARCHAR(100),
		status INTEGER DEFAULT 1,
		gender INTEGER DEFAULT 0,
		birthday VARCHAR(20),
		address VARCHAR(255),
		position_id BIGINT DEFAULT 0,
		department_id BIGINT DEFAULT 0,
		entry_date VARCHAR(20),
		last_login_ip VARCHAR(50),
		last_login_time VARCHAR(50),
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	
	-- 创建wechat_open_id的部分唯一索引，只对非空值应用唯一约束
	CREATE UNIQUE INDEX IF NOT EXISTS idx_base_users_wechat_open_id ON base_users(wechat_open_id) WHERE wechat_open_id IS NOT NULL;
	`

	if err := db.Exec(createUsersTableSQL).Error; err != nil {
		log.Error("创建base_users表失败", "error", err)
	} else {
		log.Info("base_users表创建成功")
	}

	// 创建roles表
	createRolesTableSQL := `
	CREATE TABLE IF NOT EXISTS base_roles (
		id BIGINT PRIMARY KEY,
		name VARCHAR(50) UNIQUE NOT NULL,
		code VARCHAR(50) UNIQUE NOT NULL,
		description VARCHAR(255),
		type INTEGER DEFAULT 1,
		parent_id BIGINT DEFAULT 0,
		status INTEGER DEFAULT 1,
		is_system BOOLEAN DEFAULT false,
		sort INTEGER DEFAULT 0,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	`

	if err := db.Exec(createRolesTableSQL).Error; err != nil {
		log.Error("创建 base_roles 表失败", "error", err)
	} else {
		log.Info("base_roles 表创建成功")
	}

	// 创建角色 - 用户关联表
	createRoleUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS base_role_users (
		role_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		PRIMARY KEY (role_id, user_id),
		FOREIGN KEY (role_id) REFERENCES base_roles(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES base_users(id) ON DELETE CASCADE
	);
	`

	if err := db.Exec(createRoleUsersTableSQL).Error; err != nil {
		log.Error("创建base_role_users表失败", "error", err)
	} else {
		log.Info("base_role_users表创建成功")
	}

	// 创建菜单表
	createMenusTableSQL := `
	CREATE TABLE IF NOT EXISTS base_menus (
		id BIGINT PRIMARY KEY,
		parent_id BIGINT DEFAULT 0,
		path VARCHAR(255) NOT NULL,
		name VARCHAR(100) NOT NULL UNIQUE,
		component VARCHAR(255),
		title VARCHAR(50) NOT NULL,
		icon VARCHAR(50),
		hidden BOOLEAN DEFAULT false,
		sort INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		is_ext BOOLEAN DEFAULT false,
		redirect VARCHAR(100),
		permission VARCHAR(100),
		keep_alive BOOLEAN DEFAULT false,
		default_menu BOOLEAN DEFAULT false,
		breadcrumb BOOLEAN DEFAULT true,
		active_menu VARCHAR(255),
		affix BOOLEAN DEFAULT false,
		type INTEGER DEFAULT 2,
		frame_src VARCHAR(255),
		frame_loading BOOLEAN DEFAULT true,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	`

	if err := db.Exec(createMenusTableSQL).Error; err != nil {
		log.Error("创建 base_menus 表失败", "error", err)
	} else {
		log.Info("base_menus 表创建成功")
	}

	// 创建 API 接口表
	createAPIsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_apis (
		id BIGINT PRIMARY KEY,
		path VARCHAR(255) NOT NULL,
		method VARCHAR(20) NOT NULL,
		"group" VARCHAR(50),
		auth_required BOOLEAN DEFAULT false,
		auth_type INTEGER DEFAULT 1,
		description VARCHAR(255),
		status INTEGER DEFAULT 1,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP,
		UNIQUE (path, method)
	);
	`

	if err := db.Exec(createAPIsTableSQL).Error; err != nil {
		log.Error("创建 base_apis 表失败", "error", err)
	} else {
		log.Info("base_apis 表创建成功")
	}

	// 创建按钮权限表
	createBtnPermsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_btn_perms (
		id BIGINT PRIMARY KEY,
		menu_id BIGINT,
		code VARCHAR(100) NOT NULL UNIQUE,
		name VARCHAR(50) NOT NULL,
		description VARCHAR(255),
		status INTEGER DEFAULT 1,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP,
		CONSTRAINT fk_button_menu FOREIGN KEY (menu_id) REFERENCES base_menus(id)
	);
	`

	if err := db.Exec(createBtnPermsTableSQL).Error; err != nil {
		log.Error("创建base_btn_perms表失败", "error", err)
	} else {
		log.Info("base_btn_perms表创建成功")
	}

	// 创建角色-菜单关联表
	createRoleMenusTableSQL := `
	CREATE TABLE IF NOT EXISTS base_role_menus (
		role_id BIGINT NOT NULL,
		menu_id BIGINT NOT NULL,
		PRIMARY KEY (role_id, menu_id),
		FOREIGN KEY (role_id) REFERENCES base_roles(id) ON DELETE CASCADE,
		FOREIGN KEY (menu_id) REFERENCES base_menus(id) ON DELETE CASCADE
	);
	`

	if err := db.Exec(createRoleMenusTableSQL).Error; err != nil {
		log.Error("创建base_role_menus表失败", "error", err)
	} else {
		log.Info("base_role_menus表创建成功")
	}

	// 创建角色-API 路由关联表
	createRoleAPIRoutesTableSQL := `
	CREATE TABLE IF NOT EXISTS base_role_api_routes (
		role_id BIGINT NOT NULL,
		route_id BIGINT NOT NULL,
		source INTEGER DEFAULT 1,
		note VARCHAR(255),
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (role_id, route_id),
		FOREIGN KEY (role_id) REFERENCES base_roles(id) ON DELETE CASCADE,
		FOREIGN KEY (route_id) REFERENCES base_apis(id) ON DELETE CASCADE
	);
	
	-- 添加唯一索引避免重复授权
	CREATE UNIQUE INDEX IF NOT EXISTS uk_role_route ON base_role_api_routes(role_id, route_id);
	CREATE INDEX IF NOT EXISTS idx_role_api_routes_route_id ON base_role_api_routes(route_id);
	CREATE INDEX IF NOT EXISTS idx_role_api_routes_source ON base_role_api_routes(source);
	
	-- 添加字段注释（PostgreSQL 语法）
	COMMENT ON COLUMN base_role_api_routes.source IS '授权来源：1-直接授权，2-菜单恢复，3-按钮恢复';
	COMMENT ON COLUMN base_role_api_routes.note IS '备注说明';
	`

	if err := db.Exec(createRoleAPIRoutesTableSQL).Error; err != nil {
		log.Error("创建base_role_api_routes表失败", "error", err)
	} else {
		log.Info("base_role_api_routes表创建成功")
	}

	// 创建按钮权限-API 路由关联表
	createBtnPermAPIRoutesTableSQL := `
	CREATE TABLE IF NOT EXISTS base_btn_perm_api_routes (
		btn_perm_id BIGINT NOT NULL,
		route_id BIGINT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (btn_perm_id, route_id),
		FOREIGN KEY (btn_perm_id) REFERENCES base_btn_perms(id) ON DELETE CASCADE,
		FOREIGN KEY (route_id) REFERENCES base_apis(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_btn_perm_api_routes_route_id ON base_btn_perm_api_routes(route_id);
	`

	if err := db.Exec(createBtnPermAPIRoutesTableSQL).Error; err != nil {
		log.Error("创建 base_btn_perm_api_routes 表失败", "error", err)
	} else {
		log.Info("base_btn_perm_api_routes 表创建成功")
	}

	// 创建菜单-API 路由关联表
	createMenuAPIRoutesTableSQL := `
	CREATE TABLE IF NOT EXISTS base_menu_api_routes (
		id BIGINT PRIMARY KEY,
		menu_id BIGINT NOT NULL,
		route_id BIGINT NOT NULL,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP,
		FOREIGN KEY (menu_id) REFERENCES base_menus(id) ON DELETE CASCADE,
		FOREIGN KEY (route_id) REFERENCES base_apis(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_menu_api_routes_menu_id ON base_menu_api_routes(menu_id);
	CREATE INDEX IF NOT EXISTS idx_menu_api_routes_route_id ON base_menu_api_routes(route_id);
	`

	if err := db.Exec(createMenuAPIRoutesTableSQL).Error; err != nil {
		log.Error("创建 base_menu_api_routes 表失败", "error", err)
	} else {
		log.Info("base_menu_api_routes 表创建成功")
	}

	// 创建角色-按钮权限关联表
	createRoleButtonPermissionsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_role_btn_perms (
		role_id BIGINT NOT NULL,
		button_id BIGINT NOT NULL,
		PRIMARY KEY (role_id, button_id),
		FOREIGN KEY (role_id) REFERENCES base_roles(id) ON DELETE CASCADE,
		FOREIGN KEY (button_id) REFERENCES base_btn_perms(id) ON DELETE CASCADE
	);
	`

	if err := db.Exec(createRoleButtonPermissionsTableSQL).Error; err != nil {
		log.Error("创建base_role_btn_perms表失败", "error", err)
	} else {
		log.Info("base_role_btn_perms表创建成功")
	}

	// 创建字典表
	createDictsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_dicts (
		id BIGINT PRIMARY KEY,
		code VARCHAR(100) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		description VARCHAR(255),
		status INTEGER DEFAULT 1,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	`

	if err := db.Exec(createDictsTableSQL).Error; err != nil {
		log.Error("创建base_dicts表失败", "error", err)
	} else {
		log.Info("base_dicts表创建成功")
	}

	// 创建字典项表
	createDictItemsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_dict_items (
		id BIGINT PRIMARY KEY,
		dict_id BIGINT NOT NULL,
		value VARCHAR(100) NOT NULL,
		label VARCHAR(100) NOT NULL,
		sort INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		description VARCHAR(255),
		extend_params JSONB,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP,
		CONSTRAINT fk_dict_item FOREIGN KEY (dict_id) REFERENCES base_dicts(id) ON DELETE CASCADE
	);
	`

	if err := db.Exec(createDictItemsTableSQL).Error; err != nil {
		log.Error("创建base_dict_items表失败", "error", err)
	} else {
		log.Info("base_dict_items表创建成功")
	}

	// 添加字典项索引
	addDictItemIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_base_dict_items_dict_id ON base_dict_items(dict_id);
	CREATE INDEX IF NOT EXISTS idx_base_dict_items_value ON base_dict_items(value);
	CREATE INDEX IF NOT EXISTS idx_base_dict_items_label ON base_dict_items(label);
	`

	if err := db.Exec(addDictItemIndexSQL).Error; err != nil {
		log.Error("添加base_dict_items表索引失败", "error", err)
	} else {
		log.Info("base_dict_items表索引添加成功")
	}

	// 创建操作日志表
	createOperationLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_operation_logs (
		id BIGINT PRIMARY KEY,
		user_id BIGINT DEFAULT 0,
		username VARCHAR(50),
		nickname VARCHAR(50),
		operation_type SMALLINT DEFAULT 0,
		module VARCHAR(50),
		description VARCHAR(255),
		method VARCHAR(20),
		path VARCHAR(255),
		params TEXT,
		client_ip VARCHAR(50),
		user_agent TEXT,
		status_code INT DEFAULT 0,
		result TEXT,
		duration BIGINT DEFAULT 0,
		error_message TEXT,
		is_success BOOLEAN DEFAULT true,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	`

	if err := db.Exec(createOperationLogsTableSQL).Error; err != nil {
		log.Error("创建 base_operation_logs 表失败", "error", err)
	} else {
		log.Info("base_operation_logs 表创建成功")
	}

	// 添加操作日志表索引
	addOperationLogIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_base_operation_logs_user_id ON base_operation_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_base_operation_logs_module ON base_operation_logs(module);
	CREATE INDEX IF NOT EXISTS idx_base_operation_logs_path ON base_operation_logs(path);
	CREATE INDEX IF NOT EXISTS idx_base_operation_logs_client_ip ON base_operation_logs(client_ip);
	CREATE INDEX IF NOT EXISTS idx_base_operation_logs_status_code ON base_operation_logs(status_code);
	CREATE INDEX IF NOT EXISTS idx_base_operation_logs_is_success ON base_operation_logs(is_success);
	CREATE INDEX IF NOT EXISTS idx_base_operation_logs_created_at ON base_operation_logs(created_at);
	`

	if err := db.Exec(addOperationLogIndexSQL).Error; err != nil {
		log.Error("添加 base_operation_logs 表索引失败", "error", err)
	} else {
		log.Info("base_operation_logs 表索引添加成功")
	}

	// 创建登录日志表
	createLoginLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_login_logs (
		id BIGINT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		username VARCHAR(50),
		login_ip VARCHAR(50),
		login_time TIMESTAMP NOT NULL,
		login_place VARCHAR(100),
		device VARCHAR(100),
		browser VARCHAR(50),
		os VARCHAR(50),
		result SMALLINT DEFAULT 1,
		error_msg TEXT,
		user_agent TEXT,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	`

	if err := db.Exec(createLoginLogsTableSQL).Error; err != nil {
		log.Error("创建 base_login_logs 表失败", "error", err)
	} else {
		log.Info("base_login_logs 表创建成功")
	}

	// 添加登录日志表索引
	addLoginLogIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_base_login_logs_user_id ON base_login_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_base_login_logs_username ON base_login_logs(username);
	CREATE INDEX IF NOT EXISTS idx_base_login_logs_login_ip ON base_login_logs(login_ip);
	CREATE INDEX IF NOT EXISTS idx_base_login_logs_login_time ON base_login_logs(login_time);
	CREATE INDEX IF NOT EXISTS idx_base_login_logs_result ON base_login_logs(result);
	`

	if err := db.Exec(addLoginLogIndexSQL).Error; err != nil {
		log.Error("添加 base_login_logs 表索引失败", "error", err)
	} else {
		log.Info("base_login_logs 表索引添加成功")
	}

	// 创建配置表
	createConfigsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_configs (
		id BIGINT PRIMARY KEY,
		config_key VARCHAR(100) UNIQUE NOT NULL,
		config_value TEXT,
		config_type VARCHAR(20),
		description VARCHAR(255),
		status INTEGER DEFAULT 1,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	`

	if err := db.Exec(createConfigsTableSQL).Error; err != nil {
		log.Error("创建 base_configs 表失败", "error", err)
	} else {
		log.Info("base_configs 表创建成功")
	}

	// 创建任务执行日志表
	createTaskExecutionLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS base_task_execution_logs (
		id BIGINT PRIMARY KEY,
		task_id VARCHAR(100) NOT NULL,
		task_name VARCHAR(200) NOT NULL,
		"group" VARCHAR(100),
		execute_at TIMESTAMP NOT NULL,
		duration BIGINT NOT NULL,
		result VARCHAR(20) NOT NULL,
		error_info TEXT,
		retry_count INTEGER DEFAULT 0,
		cron_expr VARCHAR(100),
		trigger_type VARCHAR(50) NOT NULL,
		operator_id BIGINT DEFAULT 0,
		created_by BIGINT DEFAULT 0,
		updated_by BIGINT DEFAULT 0,
		deleted_by BIGINT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
	`

	if err := db.Exec(createTaskExecutionLogsTableSQL).Error; err != nil {
		log.Error("创建 base_task_execution_logs 表失败", "error", err)
	} else {
		log.Info("base_task_execution_logs 表创建成功")
	}

	// 添加任务执行日志表索引
	addTaskExecutionLogIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_base_task_execution_logs_task_id ON base_task_execution_logs(task_id);
	CREATE INDEX IF NOT EXISTS idx_base_task_execution_logs_group ON base_task_execution_logs("group");
	CREATE INDEX IF NOT EXISTS idx_base_task_execution_logs_result ON base_task_execution_logs(result);
	CREATE INDEX IF NOT EXISTS idx_base_task_execution_logs_execute_at ON base_task_execution_logs(execute_at);
	`

	if err := db.Exec(addTaskExecutionLogIndexSQL).Error; err != nil {
		log.Error("添加 base_task_execution_logs 表索引失败", "error", err)
	} else {
		log.Info("base_task_execution_logs 表索引添加成功")
	}

	// 创建权限审计日志表
	CreatePermissionLogsTable(db, log)

	log.Info("base 应用数据库迁移完成")
}

// CreatePermissionLogsTable 创建权限审计日志表
func CreatePermissionLogsTable(db *database.PostgresDB, log logger.Logger) {
	log.Info("开始创建权限审计日志表")

	createPermissionLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS sys_permission_logs (
		id BIGINT PRIMARY KEY,
		operator_id BIGINT NOT NULL,
		operator_name VARCHAR(100),
		action_type VARCHAR(50) NOT NULL,
		target_type VARCHAR(50),
		target_id BIGINT,
		before_data JSONB,
		after_data JSONB,
		ip_address VARCHAR(50),
		user_agent VARCHAR(500),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_operator_id ON sys_permission_logs(operator_id);
	CREATE INDEX IF NOT EXISTS idx_action_type ON sys_permission_logs(action_type);
	CREATE INDEX IF NOT EXISTS idx_created_at ON sys_permission_logs(created_at);
	`

	if err := db.Exec(createPermissionLogsTableSQL).Error; err != nil {
		log.Error("创建 sys_permission_logs 表失败", "error", err)
	} else {
		log.Info("sys_permission_logs 表创建成功")
	}
}

// FullDatabaseMigration 执行完整的数据库迁移流程（创建数据库 + 表迁移）
func FullDatabaseMigration(params CreateDatabaseIfNotExistsParams, db *database.PostgresDB, log logger.Logger) error {
	// 1. 创建数据库（如果不存在）
	if err := CreateDatabaseIfNotExists(params, log); err != nil {
		return fmt.Errorf("数据库创建失败：%w", err)
	}

	// 2. 执行表迁移
	MigrateDatabase(db, log)
	return nil
}
