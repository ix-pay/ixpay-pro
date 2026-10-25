// Package service 租户初始化一次性任务（设计文档 01 §4.1）
// 幂等步骤：白名单校验 → CREATE DATABASE → 遍历已启用模块 TenantMigrate → 状态置 active
// 失败：status=failed + init_error 落库，管理平台可见并支持一键重试
package service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/modules"
)

// TenantCodePattern 租户编码白名单（防 DDL 注入；库名由代码拼前缀 ixpay_t_{code}）
var TenantCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,30}$`)

// InitTenantTask 租户库初始化任务（task.Task 接口实现，由 TaskManager.AddOneTimeTask 投递）
type InitTenantTask struct {
	tenantID int64
	repo     repo.TenantRepository
	manager  *database.TenantManager
	log      logger.Logger
}

// NewInitTenantTask 创建租户初始化任务
func NewInitTenantTask(tenantID int64, tenantRepo repo.TenantRepository, manager *database.TenantManager, log logger.Logger) *InitTenantTask {
	return &InitTenantTask{
		tenantID: tenantID,
		repo:     tenantRepo,
		manager:  manager,
		log:      log,
	}
}

// GetName 任务名（含租户ID，便于任务日志定位）
func (t *InitTenantTask) GetName() string {
	return fmt.Sprintf("tenant_init_%d", t.tenantID)
}

// Run 执行初始化（幂等可重试）
func (t *InitTenantTask) Run(ctx context.Context) error {
	tenant, err := t.repo.GetByID(t.tenantID)
	if err != nil {
		return fmt.Errorf("读取租户失败(tenant=%d): %w", t.tenantID, err)
	}
	if tenant == nil {
		return fmt.Errorf("租户不存在(tenant=%d)", t.tenantID)
	}

	// 防御性白名单校验（创建入口已校验，任务内二次校验防绕过）
	if !TenantCodePattern.MatchString(tenant.TenantCode) {
		return t.markFailed(tenant, fmt.Sprintf("租户编码不符合白名单规则: %s", tenant.TenantCode))
	}

	// 1. 建库（幂等；库名 = ixpay_t_{code}，已由编码白名单保证标识符安全）
	dbName := tenant.DBName
	if dbName == "" {
		dbName = "ixpay_t_" + tenant.TenantCode
	}
	if err := t.manager.CreateDatabase(ctx, dbName); err != nil {
		return t.markFailed(tenant, fmt.Sprintf("建库失败: %v", err))
	}

	// 2. 打开租户库连接并执行全部已注册模块迁移（幂等）
	gdb, err := t.manager.GormDB(ctx, tenant.ID)
	if err != nil {
		return t.markFailed(tenant, fmt.Sprintf("连接租户库失败: %v", err))
	}
	if gdb == nil {
		// 默认租户（dbname 空）不需要初始化，理论上不会走到这里
		return t.markFailed(tenant, "默认租户无需初始化独立库")
	}
	if err := modules.MigrateAllTenantDB(gdb, t.log); err != nil {
		return t.markFailed(tenant, fmt.Sprintf("模块迁移失败: %v", err))
	}

	// 3. 状态置 active
	tenant.Status = entity.TenantStatusActive
	tenant.InitError = ""
	tenant.DBName = dbName
	if err := t.repo.Update(tenant); err != nil {
		return fmt.Errorf("更新租户状态为 active 失败(tenant=%d): %w", t.tenantID, err)
	}
	_ = t.repo.AddOpLog(&entity.TenantOpLog{
		TenantID: tenant.ID,
		Action:   "init",
		Detail:   "租户库初始化完成（建库+模块迁移）",
	})

	t.log.Info("租户库初始化完成", "tenant_id", tenant.ID, "code", tenant.TenantCode, "db", dbName)
	return nil
}

// markFailed 标记初始化失败（状态+错误落库，管理平台可见可重试）
func (t *InitTenantTask) markFailed(tenant *entity.Tenant, msg string) error {
	tenant.Status = entity.TenantStatusFailed
	tenant.InitError = msg
	if err := t.repo.Update(tenant); err != nil {
		t.log.Error("写入租户失败状态出错", "tenant_id", tenant.ID, "error", err)
	}
	_ = t.repo.AddOpLog(&entity.TenantOpLog{
		TenantID: tenant.ID,
		Action:   "init_failed",
		Detail:   msg,
	})
	t.log.Error("租户库初始化失败", "tenant_id", tenant.ID, "reason", msg)
	return fmt.Errorf("租户 %s 初始化失败: %s", tenant.TenantCode, msg)
}
