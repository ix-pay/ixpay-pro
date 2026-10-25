package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/cryptoutil"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/modules"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	_ "github.com/lib/pq"
)

// TenantManager 租户数据库连接管理器（gorm 实例池池化 + 生命周期）
// 租户信息读取自主库 base_tenants 表（db_name 空 = 使用主库，即默认租户）；
// 独立实例覆盖连接密码为 AES-GCM 密文（security.config_encrypt_key 解密）
type TenantManager struct {
	master    *sql.DB
	cfg       *config.PostgresConfig
	encryptor *cryptoutil.AESGCMEncryptor
	log       logger.Logger

	mu         sync.RWMutex
	pools      map[int64]*sql.DB
	gormPools  map[int64]*gorm.DB
	lastUsedAt map[int64]time.Time
	cooldown   map[int64]time.Time // 连接失败冷却（避免抖动期打爆）

	maxPools    int           // 池上限（LRU 淘汰）
	idleRecycle time.Duration // 空闲回收
	cooldownFor time.Duration
}

// NewTenantManager 创建租户连接管理器
func NewTenantManager(master *sql.DB, cfg *config.Config, log logger.Logger) *TenantManager {
	encryptor, err := cryptoutil.NewAESGCMEncryptor(cfg.Security.ConfigEncryptKey)
	if err != nil {
		log.Error("租户加密器初始化失败（独立实例密码将无法解密）", "error", err)
	}
	return &TenantManager{
		master:      master,
		cfg:         &cfg.Postgres,
		encryptor:   encryptor,
		log:         log,
		pools:       make(map[int64]*sql.DB),
		gormPools:   make(map[int64]*gorm.DB),
		lastUsedAt:  make(map[int64]time.Time),
		cooldown:    make(map[int64]time.Time),
		maxPools:    256,
		idleRecycle: 30 * time.Minute,
		cooldownFor: 30 * time.Second,
	}
}

// tenantConnInfo 租户连接信息（主库 base_tenants 行）
type tenantConnInfo struct {
	ID     int64
	Code   string
	DBName string
	DBHost string
	DBPort int
	DBUser string
	DBPass string // 已解密
	Status string
}

// Get 获取租户数据库连接（池化；dbname 空 = 默认租户返回主库连接）
func (m *TenantManager) Get(ctx context.Context, tenantID int64) (*sql.DB, error) {
	// 快路径：池命中
	m.mu.RLock()
	pool, ok := m.pools[tenantID]
	m.mu.RUnlock()
	if ok {
		m.mu.Lock()
		m.lastUsedAt[tenantID] = time.Now()
		m.mu.Unlock()
		return pool, nil
	}

	// 失败冷却期内快速失败
	m.mu.RLock()
	cd, inCooldown := m.cooldown[tenantID]
	m.mu.RUnlock()
	if inCooldown && time.Now().Before(cd) {
		return nil, fmt.Errorf("租户库暂不可用(tenant=%d)，请稍后重试", tenantID)
	}

	// 查租户信息（主库直连，不走路由）
	info, err := m.loadTenantInfo(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// 默认租户（未配置独立库名）→ 主库连接
	if info.DBName == "" {
		m.mu.Lock()
		m.pools[tenantID] = m.master
		m.lastUsedAt[tenantID] = time.Now()
		m.mu.Unlock()
		return m.master, nil
	}

	// 打开租户库连接
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		info.DBHost, info.DBPort, info.DBUser, info.DBPass, info.DBName, m.cfg.SSLMode,
	)
	newPool, err := sql.Open("postgres", connStr)
	if err != nil {
		m.markCooldown(tenantID, err)
		return nil, fmt.Errorf("打开租户库失败(tenant=%d db=%s): %w", tenantID, info.DBName, err)
	}
	newPool.SetMaxOpenConns(10)
	newPool.SetMaxIdleConns(5)
	newPool.SetConnMaxLifetime(30 * time.Minute)
	if err := newPool.PingContext(ctx); err != nil {
		_ = newPool.Close()
		m.markCooldown(tenantID, err)
		return nil, fmt.Errorf("租户库连接失败(tenant=%d db=%s): %w", tenantID, info.DBName, err)
	}

	m.mu.Lock()
	// 双检（并发下可能已建）
	if existing, ok := m.pools[tenantID]; ok {
		m.mu.Unlock()
		_ = newPool.Close()
		return existing, nil
	}
	m.evictIfNeededLocked()
	m.pools[tenantID] = newPool
	m.lastUsedAt[tenantID] = time.Now()
	delete(m.cooldown, tenantID)
	m.mu.Unlock()

	m.log.Info("租户库连接已建立", "tenant_id", tenantID, "db", info.DBName)
	return newPool, nil
}

// GormDB 获取租户库 GORM 实例（池化；内含独立 sql.DB 连接池）
func (m *TenantManager) GormDB(ctx context.Context, tenantID int64) (*gorm.DB, error) {
	// 快路径
	m.mu.RLock()
	gdb, ok := m.gormPools[tenantID]
	m.mu.RUnlock()
	if ok {
		m.mu.Lock()
		m.lastUsedAt[tenantID] = time.Now()
		m.mu.Unlock()
		return gdb, nil
	}

	sqlDB, err := m.Get(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if sqlDB == m.master {
		// 默认租户：由调用方返回主 gorm 实例（manager 不持主 gorm 引用）
		return nil, nil
	}

	gdb, err = gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("租户库 GORM 初始化失败(tenant=%d): %w", tenantID, err)
	}

	// 全局审计字段保护：Save 全字段更新时零值 created_at 不覆盖库中原值
	RegisterAuditPreserveHooks(gdb)

	// 首次建立租户库 gorm 实例时执行一次全部已注册模块迁移对账（幂等）：
	// 自动补齐历史租户库缺失的业务表（如 wx 渠道配置表），部署新版本后访问即自愈；
	// 失败仅记日志不阻断连接建立，业务查询会以明确错误暴露真实原因
	if err := modules.MigrateAllTenantDB(gdb, m.log); err != nil {
		m.log.Error("租户库模块迁移对账失败", "tenant_id", tenantID, "error", err)
	}

	m.mu.Lock()
	m.gormPools[tenantID] = gdb
	m.lastUsedAt[tenantID] = time.Now()
	m.mu.Unlock()
	return gdb, nil
}

// loadTenantInfo 查询租户连接信息（主库直连）
func (m *TenantManager) loadTenantInfo(ctx context.Context, tenantID int64) (*tenantConnInfo, error) {
	var (
		code, status   string
		dbName, dbHost sql.NullString
		dbPort         sql.NullInt64
		dbUser         sql.NullString
		dbPassEnc      sql.NullString
	)
	err := m.master.QueryRowContext(ctx,
		`SELECT tenant_code, COALESCE(db_name,''), COALESCE(db_host,''), COALESCE(db_port,0),
		        COALESCE(db_user,''), COALESCE(db_password,''), status
		 FROM base_tenants WHERE id = $1 AND deleted_at IS NULL`, tenantID,
	).Scan(&code, &dbName, &dbHost, &dbPort, &dbUser, &dbPassEnc, &status)
	if err != nil {
		return nil, fmt.Errorf("租户不存在(tenant=%d): %w", tenantID, err)
	}
	if status != "active" && status != "init" {
		return nil, fmt.Errorf("租户 %s 当前状态不可访问: %s", code, status)
	}

	info := &tenantConnInfo{
		ID:     tenantID,
		Code:   code,
		DBName: dbName.String,
		DBHost: dbHost.String,
		DBPort: int(dbPort.Int64),
		DBUser: dbUser.String,
		Status: status,
	}
	// 独立实例覆盖（缺省沿用主库实例配置）；密码为密文时解密
	if info.DBHost == "" {
		info.DBHost = m.cfg.Host
	}
	if info.DBPort == 0 {
		info.DBPort, _ = atoi(m.cfg.Port)
	}
	if info.DBUser == "" {
		info.DBUser = m.cfg.User
	}
	if dbPassEnc.String != "" {
		plain, derr := m.encryptor.Decrypt(dbPassEnc.String)
		if derr != nil {
			return nil, fmt.Errorf("租户 %s 数据库密码解密失败: %w", code, derr)
		}
		info.DBPass = plain
	} else {
		info.DBPass = m.cfg.Password
	}
	return info, nil
}

// markCooldown 标记连接失败冷却
func (m *TenantManager) markCooldown(tenantID int64, err error) {
	m.mu.Lock()
	m.cooldown[tenantID] = time.Now().Add(m.cooldownFor)
	m.mu.Unlock()
	m.log.Error("租户库连接失败", "tenant_id", tenantID, "error", err)
}

// evictIfNeededLocked 池满时淘汰最久未用（调用方需持写锁）
func (m *TenantManager) evictIfNeededLocked() {
	if len(m.pools) < m.maxPools {
		return
	}
	var oldestID int64
	var oldest time.Time
	first := true
	for id, t := range m.lastUsedAt {
		if first || t.Before(oldest) {
			oldestID, oldest, first = id, t, false
		}
	}
	if oldestID > 0 {
		if pool, ok := m.pools[oldestID]; ok && pool != m.master {
			_ = pool.Close()
		}
		delete(m.pools, oldestID)
		// 同步清除 gorm 实例缓存，避免残留的 gorm.DB 引用已关闭的 sql.DB（后续请求会报 database is closed）
		delete(m.gormPools, oldestID)
		delete(m.lastUsedAt, oldestID)
	}
}

// StartIdleRecycler 启动空闲连接回收（后台协程，应用启动时调用一次）
func (m *TenantManager) StartIdleRecycler() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			m.mu.Lock()
			now := time.Now()
			for id, pool := range m.pools {
				if pool == m.master {
					continue
				}
				if last, ok := m.lastUsedAt[id]; ok && now.Sub(last) > m.idleRecycle {
					_ = pool.Close()
					delete(m.pools, id)
					// 同步清除 gorm 实例缓存，避免残留的 gorm.DB 引用已关闭的 sql.DB（后续请求会报 database is closed）
					delete(m.gormPools, id)
					delete(m.lastUsedAt, id)
					m.log.Info("空闲租户库连接已回收", "tenant_id", id)
				}
			}
			m.mu.Unlock()
		}
	}()
}

// Evict 驱逐租户连接（禁用/归档租户时调用；默认租户的主库连接不关闭）
func (m *TenantManager) Evict(tenantID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if pool, ok := m.pools[tenantID]; ok && pool != m.master {
		_ = pool.Close()
	}
	delete(m.pools, tenantID)
	delete(m.gormPools, tenantID)
	delete(m.lastUsedAt, tenantID)
	delete(m.cooldown, tenantID)
}

// DecryptPassword 解密租户库密码（运维 CLI 读取独立实例连接信息用；密文为 AES-GCM enc:v1: 格式）
func (m *TenantManager) DecryptPassword(cipherText string) (string, error) {
	if m.encryptor == nil {
		return "", fmt.Errorf("加密器不可用（security.config_encrypt_key 配置缺失）")
	}
	return m.encryptor.Decrypt(cipherText)
}

// MigrateTenantDBSync 连接指定租户库并同步执行全部已注册模块迁移（幂等；运维 CLI 迁移前置步骤用）
func (m *TenantManager) MigrateTenantDBSync(ctx context.Context, tenantID int64, log logger.Logger) error {
	gdb, err := m.GormDB(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("连接租户库失败(tenant=%d): %w", tenantID, err)
	}
	if gdb == nil {
		return fmt.Errorf("租户 %d 为默认租户（主库），无需模块迁移", tenantID)
	}
	return modules.MigrateAllTenantDB(gdb, log)
}

// CreateDatabase 创建租户数据库（幂等；使用 admin 凭据连接主实例 postgres 库执行）
// dbName 已由调用方完成白名单校验
func (m *TenantManager) CreateDatabase(ctx context.Context, dbName string) error {
	admin := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		m.cfg.Host, m.cfg.Port, m.cfg.User, m.cfg.Password, m.cfg.SSLMode,
	)
	adminDB, err := sql.Open("postgres", admin)
	if err != nil {
		return fmt.Errorf("连接管理实例失败: %w", err)
	}
	defer adminDB.Close()

	// 库名已白名单校验，标识符安全
	var exists bool
	err = adminDB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("查询数据库是否存在失败: %w", err)
	}
	if exists {
		m.log.Info("租户库已存在，跳过创建", "db", dbName)
		return nil
	}
	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s`, dbName)); err != nil {
		return fmt.Errorf("创建租户库失败(db=%s): %w", dbName, err)
	}
	m.log.Info("租户库创建成功", "db", dbName)
	return nil
}

// atoi 端口字符串转整数
func atoi(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("非法端口: %s", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
