// Package cli 租户运维 CLI（设计文档 01 P5：物理删库/存量数据迁移/初始化重试）
// 用法：
//
//	ixpay-pro tenant list                                  查看租户列表
//	ixpay-pro tenant init --code T2                        同步执行租户库初始化（建库+迁移）
//	ixpay-pro tenant drop --code T2 --confirm T2           物理删除租户库（双重确认：confirm 值必须等于租户编码）
//	ixpay-pro tenant migrate-data --from default --to T2 --module wx [--app-key wxapp-main] [--batch 500]
//	                                                       存量业务数据迁移（事务分批+幂等续跑+计数校验）
package cli

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	basesvc "github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	base "github.com/ix-pay/ixpay-pro/internal/persistence/base"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	_ "github.com/lib/pq"
)

// RunTenantCLI 租户运维命令入口（返回进程退出码；不启动 HTTP 服务）
func RunTenantCLI(args []string, cfg *config.Config, log logger.Logger) int {
	if len(args) == 0 {
		fmt.Println(usage)
		return 1
	}

	// 主库连接（CLI 生命周期短，直接 database/sql + gorm 最小化装配）
	masterSQL, err := sql.Open("postgres", masterConnStr(cfg))
	if err != nil {
		fmt.Printf("连接主库失败: %v\n", err)
		return 1
	}
	defer masterSQL.Close()
	if err := masterSQL.Ping(); err != nil {
		fmt.Printf("主库连接失败: %v\n", err)
		return 1
	}
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: masterSQL}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		fmt.Printf("主库 GORM 初始化失败: %v\n", err)
		return 1
	}
	// 全局审计字段保护（与主服务 gorm 实例行为一致）
	database.RegisterAuditPreserveHooks(gdb)
	pdb := &database.PostgresDB{DB: gdb}
	tenantRepo := base.NewTenantRepository(pdb)
	manager := database.NewTenantManager(masterSQL, cfg, log)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	switch args[0] {
	case "list":
		return cmdList(tenantRepo)
	case "init":
		return cmdInit(ctx, args[1:], cfg, tenantRepo, manager, log)
	case "drop":
		return cmdDrop(ctx, args[1:], cfg, masterSQL, tenantRepo, log)
	case "migrate-data":
		return cmdMigrateData(ctx, args[1:], cfg, masterSQL, manager, log)
	default:
		fmt.Printf("未知子命令: %s\n%s\n", args[0], usage)
		return 1
	}
}

const usage = `租户运维命令:
  tenant list                                  查看全部租户
  tenant init --code <编码>                     同步执行租户库初始化（建库+模块迁移）
  tenant drop --code <编码> --confirm <编码>    物理删除租户库（confirm 必须等于租户编码，防误删）
  tenant migrate-data --from <编码> --to <编码> [--module wx] [--app-key <key>] [--batch 500]
                                               存量业务数据迁移（分批+幂等续跑+计数校验）
                                               from/to 为租户编码（default 表示主库默认租户）`

// masterConnStr 主库连接串
func masterConnStr(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.SSLMode)
}

// adminConnStr 管理实例连接串（连 postgres 库，用于 CREATE/DROP DATABASE）
func adminConnStr(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.SSLMode)
}

// ---------- tenant list ----------

func cmdList(tenantRepo repo.TenantRepository) int {
	tenants, _, err := tenantRepo.List(1, 100, "")
	if err != nil {
		fmt.Printf("查询租户列表失败（请先启动服务完成主库迁移）: %v\n", err)
		return 1
	}
	fmt.Printf("%-6s %-16s %-20s %-22s %-10s\n", "ID", "编码", "名称", "库名", "状态")
	for _, t := range tenants {
		dbName := t.DBName
		if dbName == "" {
			dbName = "(主库)"
		}
		fmt.Printf("%-6d %-16s %-20s %-22s %-10s\n", t.ID, t.TenantCode, t.TenantName, dbName, t.Status)
	}
	return 0
}

// ---------- tenant init ----------

func cmdInit(ctx context.Context, args []string, cfg *config.Config, tenantRepo repo.TenantRepository, manager *database.TenantManager, log logger.Logger) int {
	code := flagValue(args, "--code")
	if code == "" {
		fmt.Println("缺少 --code 参数\n" + usage)
		return 1
	}
	tenant, err := tenantRepo.GetByCode(code)
	if err != nil || tenant == nil {
		fmt.Printf("租户不存在: %s\n", code)
		return 1
	}
	if tenant.ID == 1 || tenant.DBName == "" {
		fmt.Println("默认租户使用主库，无需初始化")
		return 1
	}
	// 同步执行与 InitTenantTask 相同的幂等初始化流程
	tenant.Status = entity.TenantStatusInit
	_ = tenantRepo.Update(tenant)
	task := basesvc.NewInitTenantTask(tenant.ID, tenantRepo, manager, log)
	if err := task.Run(ctx); err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return 1
	}
	fmt.Printf("租户 %s 初始化完成（建库+模块迁移）\n", code)
	return 0
}

// ---------- tenant drop（物理删库，双重确认） ----------

func cmdDrop(ctx context.Context, args []string, cfg *config.Config, masterSQL *sql.DB, tenantRepo repo.TenantRepository, log logger.Logger) int {
	code := flagValue(args, "--code")
	confirm := flagValue(args, "--confirm")
	if code == "" || confirm == "" {
		fmt.Println("缺少 --code 或 --confirm 参数\n" + usage)
		return 1
	}
	// 双重确认：--confirm 值必须与租户编码完全一致（防误删）
	if confirm != code {
		fmt.Printf("确认值不一致：--confirm 必须等于租户编码 %q（收到 %q）\n", code, confirm)
		return 1
	}
	if code == "default" {
		fmt.Println("默认租户使用主库，禁止删除")
		return 1
	}
	tenant, err := tenantRepo.GetByCode(code)
	if err != nil || tenant == nil {
		fmt.Printf("租户不存在: %s\n", code)
		return 1
	}
	dbName := tenant.DBName
	if dbName == "" {
		dbName = "ixpay_t_" + tenant.TenantCode
	}

	// 1. 物理删除数据库（断开既有连接强制删）
	adminDB, err := sql.Open("postgres", adminConnStr(cfg))
	if err != nil {
		fmt.Printf("连接管理实例失败: %v\n", err)
		return 1
	}
	defer adminDB.Close()
	if _, err := adminDB.ExecContext(ctx,
		fmt.Sprintf(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()`, dbName)); err != nil {
		fmt.Printf("断开租户库连接失败: %v\n", err)
		return 1
	}
	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS %s`, dbName)); err != nil {
		fmt.Printf("删除数据库失败(db=%s): %v\n", dbName, err)
		return 1
	}
	fmt.Printf("数据库 %s 已删除\n", dbName)

	// 2. 租户记录置 archived（保留审计；app_key 路由需人工处理）
	tenant.Status = entity.TenantStatusArchived
	_ = tenantRepo.Update(tenant)
	_ = tenantRepo.AddOpLog(&entity.TenantOpLog{
		TenantID: tenant.ID,
		Action:   "drop",
		Detail:   fmt.Sprintf("运维 CLI 物理删除数据库 %s", dbName),
	})
	fmt.Printf("租户 %s 已置为 archived；base_tenant_apps 中该租户的路由请按需解绑\n", code)
	log.Info("租户库已物理删除", "code", code, "db", dbName)
	return 0
}

// ---------- tenant migrate-data（存量数据迁移） ----------

// wxModuleTables wx 模块租户库业务表（迁移顺序考虑依赖：渠道配置先于 UI 配置）
var wxModuleTables = []string{
	"wx_app_configs",
	"wx_merchant_configs",
	"wx_app_merchants",
	"tenant_ui_configs",
	"wx_users",
	"wx_auth_sessions",
	"wx_payments",
	"wx_wechat_pay_infos",
	"wx_refunds",
	"wx_med_refunds",
}

func cmdMigrateData(ctx context.Context, args []string, cfg *config.Config, masterSQL *sql.DB, manager *database.TenantManager, log logger.Logger) int {
	fromCode := flagValue(args, "--from")
	toCode := flagValue(args, "--to")
	module := flagValue(args, "--module")
	appKey := flagValue(args, "--app-key")
	batchSize := 500
	if b := flagValue(args, "--batch"); b != "" {
		if n, err := strconv.Atoi(b); err == nil && n > 0 && n <= 10000 {
			batchSize = n
		}
	}
	if fromCode == "" || toCode == "" {
		fmt.Println("缺少 --from 或 --to 参数\n" + usage)
		return 1
	}
	if module != "" && module != "wx" {
		fmt.Printf("暂不支持模块: %s（当前已实现 wx）\n", module)
		return 1
	}
	if fromCode == toCode {
		fmt.Println("--from 与 --to 不能相同")
		return 1
	}

	// 源连接：default = 主库；否则按租户配置连接
	var srcDB *sql.DB
	if fromCode == "default" {
		srcDB = masterSQL
	} else {
		fromTenant, err := lookupTenant(ctx, masterSQL, fromCode)
		if err != nil {
			fmt.Printf("源租户读取失败: %v\n", err)
			return 1
		}
		db, err := openTenantDB(ctx, cfg, fromTenant)
		if err != nil {
			fmt.Printf("源租户库连接失败: %v\n", err)
			return 1
		}
		defer db.Close()
		srcDB = db
	}

	// 目标租户（必须为独立库租户，default 无需迁移）
	toTenant, err := lookupTenant(ctx, masterSQL, toCode)
	if err != nil {
		fmt.Printf("目标租户读取失败: %v\n", err)
		return 1
	}
	if toTenant.DBName == "" {
		fmt.Println("目标租户为默认租户（主库），无需迁移；请先创建独立库租户")
		return 1
	}
	dstDB, err := openTenantDB(ctx, cfg, toTenant)
	if err != nil {
		fmt.Printf("目标租户库连接失败: %v\n", err)
		return 1
	}
	defer dstDB.Close()

	// 目标库需先完成建表迁移（幂等；未迁移则自动执行）
	if err := manager.CreateDatabase(ctx, toTenant.DBName); err != nil {
		fmt.Printf("目标库创建/校验失败: %v\n", err)
		return 1
	}
	if err := manager.MigrateTenantDBSync(ctx, toTenant.ID, log); err != nil {
		fmt.Printf("目标库模块迁移失败: %v\n", err)
		return 1
	}

	where := ""
	args2 := []interface{}{}
	if appKey != "" {
		where = " WHERE app_key = $1"
		args2 = append(args2, appKey)
	}

	fmt.Printf("开始迁移: %s(%s) → %s(%s) 模块=%s app_key=%s 批大小=%d\n",
		fromCode, mapDBName(fromCode), toCode, toTenant.DBName, orDefault(module, "wx"), orDefault(appKey, "全部"), batchSize)

	totalCopied := int64(0)
	for _, table := range wxModuleTables {
		copied, srcCount, err := copyTable(ctx, srcDB, dstDB, table, where, args2, batchSize)
		if err != nil {
			fmt.Printf("表 %s 迁移失败: %v（已复制 %d 行，修复后重跑命令可续传）\n", table, err, copied)
			return 1
		}
		fmt.Printf("  %-22s 源=%-8d 本次复制=%-8d\n", table, srcCount, copied)
		totalCopied += copied
	}
	fmt.Printf("迁移完成，共复制 %d 行。请业务校验后，在管理平台将对应 app_key 路由指向租户 %s\n", totalCopied, toCode)
	return 0
}

func mapDBName(code string) string {
	if code == "default" {
		return "主库"
	}
	return "ixpay_t_" + code
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// lookupTenant 按编码查租户（含连接信息解密）
func lookupTenant(ctx context.Context, master *sql.DB, code string) (*entity.Tenant, error) {
	if !basesvc.TenantCodePattern.MatchString(code) && code != "default" {
		return nil, fmt.Errorf("租户编码不合法: %s", code)
	}
	var (
		id        int64
		dbName    sql.NullString
		dbHost    sql.NullString
		dbPort    sql.NullInt64
		dbUser    sql.NullString
		dbPassEnc sql.NullString
	)
	err := master.QueryRowContext(ctx,
		`SELECT id, COALESCE(db_name,''), COALESCE(db_host,''), COALESCE(db_port,0), COALESCE(db_user,''), COALESCE(db_password,'')
		 FROM base_tenants WHERE tenant_code = $1 AND deleted_at IS NULL`, code,
	).Scan(&id, &dbName, &dbHost, &dbPort, &dbUser, &dbPassEnc)
	if err != nil {
		return nil, fmt.Errorf("租户 %s 不存在: %w", code, err)
	}
	return &entity.Tenant{
		ID:         id,
		TenantCode: code,
		DBName:     dbName.String,
		DBHost:     dbHost.String,
		DBPort:     int(dbPort.Int64),
		DBUser:     dbUser.String,
		DBPassword: dbPassEnc.String,
	}, nil
}

// openTenantDB 打开租户库连接（独立实例覆盖；密码密文解密）
func openTenantDB(ctx context.Context, cfg *config.Config, t *entity.Tenant) (*sql.DB, error) {
	host, port, user, pass := cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password
	if t.DBHost != "" {
		host = t.DBHost
	}
	if t.DBPort > 0 {
		port = strconv.Itoa(t.DBPort)
	}
	if t.DBUser != "" {
		user = t.DBUser
	}
	if t.DBPassword != "" {
		// 独立实例密码为 AES-GCM 密文，复用 TenantManager 的解密逻辑
		plain, err := managerDecrypt(cfg, t.DBPassword)
		if err != nil {
			return nil, fmt.Errorf("租户密码解密失败: %w", err)
		}
		pass = plain
	}
	dbName := t.DBName
	if dbName == "" {
		dbName = "ixpay_t_" + t.TenantCode
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, dbName, cfg.Postgres.SSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// managerDecrypt 复用 TenantManager 的解密器（独立构建避免依赖运行期连接池）
func managerDecrypt(cfg *config.Config, cipherText string) (string, error) {
	m := database.NewTenantManager(nil, cfg, nopLogger{})
	return m.DecryptPassword(cipherText)
}

// nopLogger CLI 场景静默日志
type nopLogger struct{}

func (nopLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (nopLogger) Info(msg string, keysAndValues ...interface{})  {}
func (nopLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (nopLogger) Error(msg string, keysAndValues ...interface{}) {}
func (nopLogger) Fatal(msg string, keysAndValues ...interface{}) {}
func (nopLogger) Sync() error                                    { return nil }
func (l nopLogger) With(fields ...interface{}) logger.Logger     { return l }
func (l nopLogger) WithContext(ctx context.Context) logger.Logger {
	return l
}

// copyTable 通用表复制：源游标分批读取 → 目标幂等写入（ON CONFLICT (id) DO NOTHING）
// 返回本次复制行数与源表总行数；断点续跑依赖幂等写入（重跑自动跳过已存在 id）
func copyTable(ctx context.Context, src, dst *sql.DB, table, where string, whereArgs []interface{}, batchSize int) (int64, int64, error) {
	// 源总行数
	var srcCount int64
	if err := src.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+where, whereArgs...).Scan(&srcCount); err != nil {
		return 0, 0, fmt.Errorf("源表计数失败: %w", err)
	}
	if srcCount == 0 {
		return 0, 0, nil
	}

	// 两库列交集（按源表列序），目标库缺列则跳过该列
	srcCols, err := tableColumns(ctx, src, table)
	if err != nil {
		return 0, srcCount, err
	}
	dstCols, err := tableColumns(ctx, dst, table)
	if err != nil {
		return 0, srcCount, err
	}
	dstColSet := make(map[string]bool, len(dstCols))
	for _, c := range dstCols {
		dstColSet[c] = true
	}
	cols := make([]string, 0, len(srcCols))
	for _, c := range srcCols {
		if dstColSet[c] {
			cols = append(cols, c)
		}
	}
	if len(cols) == 0 {
		return 0, srcCount, fmt.Errorf("目标库无同名列")
	}

	colList := strings.Join(cols, ", ")
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) DO NOTHING",
		table, colList, strings.Join(placeholders, ", "))

	var lastID int64
	var copied int64
	for {
		// 游标分批：WHERE id > lastID ORDER BY id LIMIT batch（稳定不重不漏）
		query := fmt.Sprintf("SELECT %s FROM %s%s", colList, table, where)
		args := append([]interface{}{}, whereArgs...)
		if where != "" {
			query += fmt.Sprintf(" AND id > $%d", len(args)+1)
		} else {
			query += " WHERE id > $1"
		}
		args = append(args, lastID)
		query += fmt.Sprintf(" ORDER BY id LIMIT %d", batchSize)

		rows, err := src.QueryContext(ctx, query, args...)
		if err != nil {
			return copied, srcCount, fmt.Errorf("源表读取失败: %w", err)
		}

		batchRows := 0
		for rows.Next() {
			vals := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				rows.Close()
				return copied, srcCount, fmt.Errorf("行扫描失败: %w", err)
			}
			// 取 id 游标（列序与 cols 一致，id 必在首列或指定列中）
			for i, c := range cols {
				if c == "id" {
					if v, ok := vals[i].(int64); ok {
						lastID = v
					}
				}
			}
			if _, err := dst.ExecContext(ctx, insertSQL, vals...); err != nil {
				rows.Close()
				return copied, srcCount, fmt.Errorf("目标写入失败(id游标=%d): %w", lastID, err)
			}
			copied++
			batchRows++
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return copied, srcCount, err
		}
		if batchRows < batchSize {
			break
		}
	}
	return copied, srcCount, nil
}

// tableColumns 查询表列名（按 ordinal_position）
func tableColumns(ctx context.Context, db *sql.DB, table string) ([]string, error) {
	// 表名白名单校验（防注入；仅允许字母数字下划线）
	if !regexp.MustCompile(`^[a-z_][a-z0-9_]*$`).MatchString(table) {
		return nil, fmt.Errorf("非法表名: %s", table)
	}
	rows, err := db.QueryContext(ctx,
		`SELECT column_name FROM information_schema.columns
		 WHERE table_name = $1 AND table_schema = 'public' ORDER BY ordinal_position`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

// flagValue 从参数列表取 --key value 值
func flagValue(args []string, key string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == key {
			return strings.TrimSpace(args[i+1])
		}
	}
	return ""
}
