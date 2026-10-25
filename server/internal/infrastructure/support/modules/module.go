// Package modules 业务模块注册契约（多租户可插拔：wx / ali / wecom ...）
// 每个业务模块实现 TenantModule 并注册到 registry，租户库初始化时遍历执行迁移
package modules

import (
	"fmt"
	"sync"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"gorm.io/gorm"
)

// TenantModule 业务模块的租户库契约
type TenantModule interface {
	// Name 模块名（wx / ali / wecom），同时是租户库表前缀
	Name() string
	// TenantMigrate 在租户库执行本模块建表迁移（幂等，可重复执行）
	TenantMigrate(db *gorm.DB, log logger.Logger) error
	// TenantSeed 在租户库执行本模块种子数据（可选实现）
	TenantSeed(db *gorm.DB, log logger.Logger) error
}

var (
	mu       sync.RWMutex
	registry = map[string]TenantModule{}
)

// RegisterModule 注册业务模块（main/wire 装配时调用；重名 panic 提示装配错误）
func RegisterModule(m TenantModule) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := registry[m.Name()]; exists {
		panic(fmt.Sprintf("业务模块重复注册: %s", m.Name()))
	}
	registry[m.Name()] = m
}

// GetModule 按名称取模块
func GetModule(name string) (TenantModule, bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := registry[name]
	return m, ok
}

// AllModules 全部已注册模块
func AllModules() []TenantModule {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]TenantModule, 0, len(registry))
	for _, m := range registry {
		out = append(out, m)
	}
	return out
}

// MigrateTenantDB 在租户库执行指定模块迁移（模块未注册返回错误）
func MigrateTenantDB(moduleName string, db *gorm.DB, log logger.Logger) error {
	m, ok := GetModule(moduleName)
	if !ok {
		return fmt.Errorf("业务模块未注册: %s", moduleName)
	}
	return m.TenantMigrate(db, log)
}

// MigrateAllTenantDB 在租户库执行全部已注册模块迁移
// 每模块迁移成功后写入 tenant_schema_migrations 版本记录（幂等 upsert，见设计文档 01 §4.1 步骤4）
func MigrateAllTenantDB(db *gorm.DB, log logger.Logger) error {
	for _, m := range AllModules() {
		if err := m.TenantMigrate(db, log); err != nil {
			return fmt.Errorf("模块 %s 迁移失败: %w", m.Name(), err)
		}
		if err := recordMigrationVersion(db, m.Name()); err != nil {
			return fmt.Errorf("模块 %s 版本记录失败: %w", m.Name(), err)
		}
	}
	return nil
}

// recordMigrationVersion 记录模块迁移版本到租户库（增量迁移判断依据；当前迁移全幂等，重跑安全）
func recordMigrationVersion(db *gorm.DB, moduleName string) error {
	createSQL := `
	CREATE TABLE IF NOT EXISTS tenant_schema_migrations (
		module VARCHAR(20) PRIMARY KEY,
		version VARCHAR(32) NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`
	if err := db.Exec(createSQL).Error; err != nil {
		return err
	}
	upsertSQL := `
	INSERT INTO tenant_schema_migrations (module, version) VALUES (?, 'v1')
	ON CONFLICT (module) DO UPDATE SET version = EXCLUDED.version, applied_at = NOW();`
	return db.Exec(upsertSQL, moduleName).Error
}
