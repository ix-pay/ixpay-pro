package repo

import (
	"context"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
)

// TenantRepository 租户仓库接口（主库）
type TenantRepository interface {
	// 租户
	GetByID(id int64) (*entity.Tenant, error)
	GetByCode(code string) (*entity.Tenant, error)
	List(page, pageSize int, status string) ([]*entity.Tenant, int, error)
	Create(tenant *entity.Tenant) error
	Update(tenant *entity.Tenant) error
	Delete(id int64) error
	// 应用路由
	GetAppRoute(appKey string) (*entity.TenantApp, error)
	CreateAppRoute(route *entity.TenantApp) error
	DeleteAppRoute(id int64) error
	ListAppRoutes(tenantID int64) ([]*entity.TenantApp, error)
	// 用户授权
	ListGrants(tenantID int64) ([]*entity.TenantGrant, error)
	GetGrant(userID, tenantID int64) (*entity.TenantGrant, error)
	ListUserGrantTenants(userID int64) ([]*entity.Tenant, error)
	CreateGrant(grant *entity.TenantGrant) error
	DeleteGrant(userID, tenantID int64) error
	// 运维审计
	AddOpLog(log *entity.TenantOpLog) error
	ListOpLogs(tenantID int64, limit int) ([]*entity.TenantOpLog, error)
	// 商户号→租户映射（支付回调租户定位，设计文档 01 §九）
	// UpsertMchRef 登记商户配置归属（渠道配置保存/更新时调用）；GetTenantByPublicKeyID 按回调头 Wechatpay-Serial 定位租户
	UpsertMchRef(mchID, publicKeyID string, tenantID int64) error
	DeleteMchRef(mchID string) error
	GetTenantByPublicKeyID(publicKeyID string) (int64, error)
	// 租户库连接（database/sql 层，供 TenantManager 使用）
	GetTenantConnInfo(ctx context.Context, tenantID int64) (*entity.Tenant, error)
}
