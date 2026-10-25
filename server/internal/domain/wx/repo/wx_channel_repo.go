package repo

import (
	"context"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
)

// WxChannelRepository 微信多渠道配置仓库接口（应用配置/商户配置/应用-商户关联/租户UI配置）
// 方法携带 ctx：ctx 含租户上下文时路由租户库（渠道配置归租户资产，见设计文档 01 D4）
type WxChannelRepository interface {
	// 应用配置
	GetAppByKey(ctx context.Context, appKey string) (*entity.WxAppConfig, error)
	GetAppByID(ctx context.Context, id int64) (*entity.WxAppConfig, error)
	ListApps(ctx context.Context, page, pageSize int) ([]*entity.WxAppConfig, int, error)
	CreateApp(ctx context.Context, app *entity.WxAppConfig) error
	UpdateApp(ctx context.Context, app *entity.WxAppConfig) error
	DeleteApp(ctx context.Context, id int64) error
	// 商户配置
	GetMerchantByID(ctx context.Context, id int64) (*entity.WxMerchantConfig, error)
	GetMerchantByMchID(ctx context.Context, mchID string) (*entity.WxMerchantConfig, error)
	ListMerchants(ctx context.Context, page, pageSize int) ([]*entity.WxMerchantConfig, int, error)
	ListEnabledMerchants(ctx context.Context) ([]*entity.WxMerchantConfig, error)
	CreateMerchant(ctx context.Context, m *entity.WxMerchantConfig) error
	UpdateMerchant(ctx context.Context, m *entity.WxMerchantConfig) error
	DeleteMerchant(ctx context.Context, id int64) error
	// 应用-商户关联
	ListAppMerchants(ctx context.Context, appConfigID int64) ([]*entity.WxAppMerchant, error)
	GetAppMerchant(ctx context.Context, appConfigID, merchantConfigID int64) (*entity.WxAppMerchant, error)
	// GetDefaultAppMerchant 获取应用标记为默认的商户关联（is_default=true，唯一数据源；无默认返回 gorm.ErrRecordNotFound）
	GetDefaultAppMerchant(ctx context.Context, appConfigID int64) (*entity.WxAppMerchant, error)
	// ListDefaultAppMerchants 批量获取多个应用的默认商户关联（管理端列表展示用，避免逐应用 N+1 查询）
	ListDefaultAppMerchants(ctx context.Context, appConfigIDs []int64) ([]*entity.WxAppMerchant, error)
	CreateAppMerchant(ctx context.Context, link *entity.WxAppMerchant) error
	UpdateAppMerchant(ctx context.Context, link *entity.WxAppMerchant) error
	DeleteAppMerchant(ctx context.Context, id int64) error
	// 租户前端个性化配置（wxapp/miniapp 菜单/图片/主题等，按应用实例一份）
	GetUIByAppConfigID(ctx context.Context, appConfigID int64) (*entity.TenantUIConfig, error)
	SaveUIConfig(ctx context.Context, cfg *entity.TenantUIConfig) error
}
