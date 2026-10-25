package persistence

import (
	"context"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/cryptoutil"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
	"gorm.io/gorm"
)

// wxAppConfigModel 微信应用配置模型（wx_app_configs 表）
type wxAppConfigModel struct {
	database.SnowflakeBaseModel
	AppKey            string `gorm:"size:50;not null;uniqueIndex"`
	AppType           string `gorm:"size:20;not null;default:mp"`
	AppID             string `gorm:"size:50;not null"`
	AppSecret         string `gorm:"type:text"` // 数据库加密存储
	AuthScope         string `gorm:"size:20;default:snsapi_base"`
	OauthCallbackURL  string `gorm:"size:255"`
	FrontendURL       string `gorm:"size:255"`
	Description       string `gorm:"size:255"`
	Status            *int   `gorm:"not null;default:1;index"` // 指针类型区分零值（禁用 0 与未设置）
}

func (wxAppConfigModel) TableName() string { return "wx_app_configs" }

func (m *wxAppConfigModel) toDomain() *entity.WxAppConfig {
	if m == nil {
		return nil
	}
	// 安全解引用，提供默认值
	status := 1
	if m.Status != nil {
		status = *m.Status
	}
	return &entity.WxAppConfig{
		ID:               m.ID,
		AppKey:           m.AppKey,
		AppType:          m.AppType,
		AppID:            m.AppID,
		AppSecret:        m.AppSecret,
		AuthScope:        m.AuthScope,
		OauthCallbackURL: m.OauthCallbackURL,
		FrontendURL:      m.FrontendURL,
		Description:      m.Description,
		Status:           status,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func fromDomainWxApp(a *entity.WxAppConfig) *wxAppConfigModel {
	return &wxAppConfigModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{ID: a.ID},
		AppKey:             a.AppKey,
		AppType:            a.AppType,
		AppID:              a.AppID,
		AppSecret:          a.AppSecret,
		AuthScope:          a.AuthScope,
		OauthCallbackURL:   a.OauthCallbackURL,
		FrontendURL:        a.FrontendURL,
		Description:        a.Description,
		Status:             common.IntPtr(a.Status),
	}
}

// wxMerchantConfigModel 微信商户配置模型（wx_merchant_configs 表）
type wxMerchantConfigModel struct {
	database.SnowflakeBaseModel
	MchID           string `gorm:"size:32;not null;uniqueIndex"`
	MchName         string `gorm:"size:100"`
	MchSerialNo     string `gorm:"size:64"`
	Apiv3Key        string `gorm:"type:text"` // 数据库加密存储
	PrivateKey      string `gorm:"type:text"` // 商户私钥 PEM，数据库加密存储
	PublicKeyID     string `gorm:"type:text"` // 公钥 ID，数据库加密存储（密文超 64 字符，需 text 列）
	PublicKey       string `gorm:"type:text"` // 公钥 PEM，数据库加密存储
	PayNotifyURL    string `gorm:"size:255"`
	RefundNotifyURL string `gorm:"size:255"`
	Status          *int   `gorm:"not null;default:1;index"` // 指针类型区分零值
}

func (wxMerchantConfigModel) TableName() string { return "wx_merchant_configs" }

func (m *wxMerchantConfigModel) toDomain() *entity.WxMerchantConfig {
	if m == nil {
		return nil
	}
	// 安全解引用，提供默认值
	status := 1
	if m.Status != nil {
		status = *m.Status
	}
	return &entity.WxMerchantConfig{
		ID:              m.ID,
		MchID:           m.MchID,
		MchName:         m.MchName,
		MchSerialNo:     m.MchSerialNo,
		Apiv3Key:        m.Apiv3Key,
		PrivateKey:      m.PrivateKey,
		PublicKeyID:     m.PublicKeyID,
		PublicKey:       m.PublicKey,
		PayNotifyURL:    m.PayNotifyURL,
		RefundNotifyURL: m.RefundNotifyURL,
		Status:          status,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func fromDomainWxMerchant(mch *entity.WxMerchantConfig) *wxMerchantConfigModel {
	return &wxMerchantConfigModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{ID: mch.ID},
		MchID:              mch.MchID,
		MchName:            mch.MchName,
		MchSerialNo:        mch.MchSerialNo,
		Apiv3Key:           mch.Apiv3Key,
		PrivateKey:         mch.PrivateKey,
		PublicKeyID:        mch.PublicKeyID,
		PublicKey:          mch.PublicKey,
		PayNotifyURL:       mch.PayNotifyURL,
		RefundNotifyURL:    mch.RefundNotifyURL,
		Status:             common.IntPtr(mch.Status),
	}
}

// wxAppMerchantModel 应用-商户关联模型（wx_app_merchants 表）
type wxAppMerchantModel struct {
	database.SnowflakeBaseModel
	AppConfigID      int64 `gorm:"not null;index"`
	MerchantConfigID int64 `gorm:"not null;index"`
	IsDefault        *bool `gorm:"not null;default:false"` // 指针类型区分零值
}

func (wxAppMerchantModel) TableName() string { return "wx_app_merchants" }

func (m *wxAppMerchantModel) toDomain() *entity.WxAppMerchant {
	if m == nil {
		return nil
	}
	// 安全解引用，提供默认值
	isDefault := false
	if m.IsDefault != nil {
		isDefault = *m.IsDefault
	}
	return &entity.WxAppMerchant{
		ID:               m.ID,
		AppConfigID:      m.AppConfigID,
		MerchantConfigID: m.MerchantConfigID,
		IsDefault:        isDefault,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func fromDomainWxAppMerchant(l *entity.WxAppMerchant) *wxAppMerchantModel {
	return &wxAppMerchantModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{ID: l.ID},
		AppConfigID:        l.AppConfigID,
		MerchantConfigID:   l.MerchantConfigID,
		IsDefault:          common.BoolPtr(l.IsDefault),
	}
}

// wxChannelRepository 微信多渠道配置仓储实现（方法按 ctx 租户路由：渠道配置归租户库，见设计文档 01 D4）
type wxChannelRepository struct {
	db        *database.PostgresDB
	encryptor *cryptoutil.AESGCMEncryptor
}

var _ repo.WxChannelRepository = (*wxChannelRepository)(nil)

// g 按上下文租户路由 GORM 实例（无租户上下文/默认租户 → 主库）
func (r *wxChannelRepository) g(ctx context.Context) *gorm.DB {
	pdb := r.db
	if ctx != nil {
		if gdb, err := pdb.TenantGorm(ctx); err == nil && gdb != nil {
			return gdb
		}
	}
	return pdb.DB
}

// NewWxChannelRepository 创建微信多渠道配置仓储实现
// app_secret/apiv3_key/private_key/public_key_id/public_key 使用与配置表一致的 AES-GCM 加密器加密存储
func NewWxChannelRepository(db *database.PostgresDB, cfg *config.Config) (repo.WxChannelRepository, error) {
	encryptor, err := cryptoutil.NewAESGCMEncryptor(cfg.Security.ConfigEncryptKey)
	if err != nil {
		return nil, err
	}
	return &wxChannelRepository{db: db, encryptor: encryptor}, nil
}

// ---------- 应用配置 ----------

func (r *wxChannelRepository) GetAppByKey(ctx context.Context, appKey string) (*entity.WxAppConfig, error) {
	var m wxAppConfigModel
	if err := r.g(ctx).Where("app_key = ?", appKey).First(&m).Error; err != nil {
		return nil, err
	}
	app := m.toDomain()
	if err := r.decryptApp(app); err != nil {
		return nil, err
	}
	return app, nil
}

func (r *wxChannelRepository) GetAppByID(ctx context.Context, id int64) (*entity.WxAppConfig, error) {
	var m wxAppConfigModel
	if err := r.g(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	app := m.toDomain()
	if err := r.decryptApp(app); err != nil {
		return nil, err
	}
	return app, nil
}

func (r *wxChannelRepository) ListApps(ctx context.Context, page, pageSize int) ([]*entity.WxAppConfig, int, error) {
	var total int64
	var models []wxAppConfigModel
	query := r.g(ctx).Model(&wxAppConfigModel{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	apps := make([]*entity.WxAppConfig, len(models))
	for i := range models {
		app := models[i].toDomain()
		if err := r.decryptApp(app); err != nil {
			return nil, 0, err
		}
		apps[i] = app
	}
	return apps, int(total), nil
}

func (r *wxChannelRepository) CreateApp(ctx context.Context, app *entity.WxAppConfig) error {
	if err := r.encryptApp(app); err != nil {
		return err
	}
	m := fromDomainWxApp(app)
	if err := r.g(ctx).Create(m).Error; err != nil {
		return err
	}
	app.ID = m.ID
	return nil
}

func (r *wxChannelRepository) UpdateApp(ctx context.Context, app *entity.WxAppConfig) error {
	if err := r.encryptApp(app); err != nil {
		return err
	}
	return r.g(ctx).Save(fromDomainWxApp(app)).Error
}

func (r *wxChannelRepository) DeleteApp(ctx context.Context, id int64) error {
	return r.g(ctx).Delete(&wxAppConfigModel{}, id).Error
}

// ---------- 商户配置 ----------

func (r *wxChannelRepository) GetMerchantByID(ctx context.Context, id int64) (*entity.WxMerchantConfig, error) {
	var m wxMerchantConfigModel
	if err := r.g(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	mch := m.toDomain()
	if err := r.decryptMerchant(mch); err != nil {
		return nil, err
	}
	return mch, nil
}

func (r *wxChannelRepository) GetMerchantByMchID(ctx context.Context, mchID string) (*entity.WxMerchantConfig, error) {
	var m wxMerchantConfigModel
	if err := r.g(ctx).Where("mch_id = ?", mchID).First(&m).Error; err != nil {
		return nil, err
	}
	mch := m.toDomain()
	if err := r.decryptMerchant(mch); err != nil {
		return nil, err
	}
	return mch, nil
}

func (r *wxChannelRepository) ListMerchants(ctx context.Context, page, pageSize int) ([]*entity.WxMerchantConfig, int, error) {
	var total int64
	var models []wxMerchantConfigModel
	query := r.g(ctx).Model(&wxMerchantConfigModel{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	merchants := make([]*entity.WxMerchantConfig, len(models))
	for i := range models {
		mch := models[i].toDomain()
		if err := r.decryptMerchant(mch); err != nil {
			return nil, 0, err
		}
		merchants[i] = mch
	}
	return merchants, int(total), nil
}

func (r *wxChannelRepository) ListEnabledMerchants(ctx context.Context) ([]*entity.WxMerchantConfig, error) {
	var models []wxMerchantConfigModel
	if err := r.g(ctx).Where("status = ?", 1).Order("id DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	merchants := make([]*entity.WxMerchantConfig, len(models))
	for i := range models {
		mch := models[i].toDomain()
		if err := r.decryptMerchant(mch); err != nil {
			return nil, err
		}
		merchants[i] = mch
	}
	return merchants, nil
}

func (r *wxChannelRepository) CreateMerchant(ctx context.Context, mch *entity.WxMerchantConfig) error {
	if err := r.encryptMerchant(mch); err != nil {
		return err
	}
	m := fromDomainWxMerchant(mch)
	if err := r.g(ctx).Create(m).Error; err != nil {
		return err
	}
	mch.ID = m.ID
	return nil
}

func (r *wxChannelRepository) UpdateMerchant(ctx context.Context, mch *entity.WxMerchantConfig) error {
	if err := r.encryptMerchant(mch); err != nil {
		return err
	}
	return r.g(ctx).Save(fromDomainWxMerchant(mch)).Error
}

func (r *wxChannelRepository) DeleteMerchant(ctx context.Context, id int64) error {
	return r.g(ctx).Delete(&wxMerchantConfigModel{}, id).Error
}

// ---------- 应用-商户关联 ----------

func (r *wxChannelRepository) ListAppMerchants(ctx context.Context, appConfigID int64) ([]*entity.WxAppMerchant, error) {
	var models []wxAppMerchantModel
	if err := r.g(ctx).Where("app_config_id = ?", appConfigID).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	links := make([]*entity.WxAppMerchant, len(models))
	for i := range models {
		links[i] = models[i].toDomain()
	}
	return links, nil
}

func (r *wxChannelRepository) GetAppMerchant(ctx context.Context, appConfigID, merchantConfigID int64) (*entity.WxAppMerchant, error) {
	var m wxAppMerchantModel
	if err := r.g(ctx).Where("app_config_id = ? AND merchant_config_id = ?", appConfigID, merchantConfigID).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

// GetDefaultAppMerchant 获取应用标记为默认的商户关联（is_default=true，默认商户唯一数据源）
func (r *wxChannelRepository) GetDefaultAppMerchant(ctx context.Context, appConfigID int64) (*entity.WxAppMerchant, error) {
	var m wxAppMerchantModel
	if err := r.g(ctx).Where("app_config_id = ? AND is_default = ?", appConfigID, true).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

// ListDefaultAppMerchants 批量获取多个应用的默认商户关联（管理端应用列表展示默认商户列用）
func (r *wxChannelRepository) ListDefaultAppMerchants(ctx context.Context, appConfigIDs []int64) ([]*entity.WxAppMerchant, error) {
	if len(appConfigIDs) == 0 {
		return nil, nil
	}
	var models []wxAppMerchantModel
	if err := r.g(ctx).Where("app_config_id IN ? AND is_default = ?", appConfigIDs, true).Find(&models).Error; err != nil {
		return nil, err
	}
	links := make([]*entity.WxAppMerchant, len(models))
	for i := range models {
		links[i] = models[i].toDomain()
	}
	return links, nil
}

func (r *wxChannelRepository) CreateAppMerchant(ctx context.Context, link *entity.WxAppMerchant) error {
	m := fromDomainWxAppMerchant(link)
	if err := r.g(ctx).Create(m).Error; err != nil {
		return err
	}
	link.ID = m.ID
	return nil
}

func (r *wxChannelRepository) UpdateAppMerchant(ctx context.Context, link *entity.WxAppMerchant) error {
	return r.g(ctx).Save(fromDomainWxAppMerchant(link)).Error
}

func (r *wxChannelRepository) DeleteAppMerchant(ctx context.Context, id int64) error {
	return r.g(ctx).Delete(&wxAppMerchantModel{}, id).Error
}

// ---------- 租户前端个性化配置 ----------

// tenantUIConfigModel 租户 UI 配置模型（tenant_ui_configs 表）
type tenantUIConfigModel struct {
	database.SnowflakeBaseModel
	AppConfigID int64  `gorm:"not null;uniqueIndex"`
	UIJSON      string `gorm:"type:text"`
	Status      *int   `gorm:"not null;default:1"` // 指针类型区分零值
}

func (tenantUIConfigModel) TableName() string { return "tenant_ui_configs" }

func (m *tenantUIConfigModel) toDomain() *entity.TenantUIConfig {
	if m == nil {
		return nil
	}
	// 安全解引用，提供默认值
	status := 1
	if m.Status != nil {
		status = *m.Status
	}
	return &entity.TenantUIConfig{
		ID:          m.ID,
		AppConfigID: m.AppConfigID,
		UIJSON:      m.UIJSON,
		Status:      status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (r *wxChannelRepository) GetUIByAppConfigID(ctx context.Context, appConfigID int64) (*entity.TenantUIConfig, error) {
	var m tenantUIConfigModel
	if err := r.g(ctx).Where("app_config_id = ?", appConfigID).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *wxChannelRepository) SaveUIConfig(ctx context.Context, cfg *entity.TenantUIConfig) error {
	m := &tenantUIConfigModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{ID: cfg.ID},
		AppConfigID:        cfg.AppConfigID,
		UIJSON:             cfg.UIJSON,
		Status:             common.IntPtr(cfg.Status),
	}
	// 按 app_config_id 幂等保存（存在则更新）
	var count int64
	if err := r.g(ctx).Model(&tenantUIConfigModel{}).Where("app_config_id = ?", cfg.AppConfigID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		// Omit 审计列：更新业务字段时不覆盖 created_at/created_by 等创建审计信息
		return r.g(ctx).Omit("created_at", "created_by", "updated_by", "deleted_by", "deleted_at").
			Where("app_config_id = ?", cfg.AppConfigID).Save(m).Error
	}
	if err := r.g(ctx).Create(m).Error; err != nil {
		return err
	}
	cfg.ID = m.ID
	return nil
}

// ---------- 加解密 ----------

func (r *wxChannelRepository) encryptApp(app *entity.WxAppConfig) error {
	if r.encryptor == nil || app.AppSecret == "" {
		return nil
	}
	enc, err := r.encryptor.Encrypt(app.AppSecret)
	if err != nil {
		return err
	}
	app.AppSecret = enc
	return nil
}

func (r *wxChannelRepository) decryptApp(app *entity.WxAppConfig) error {
	if r.encryptor == nil || app.AppSecret == "" {
		return nil
	}
	plain, err := r.encryptor.Decrypt(app.AppSecret)
	if err != nil {
		return err
	}
	app.AppSecret = plain
	return nil
}

func (r *wxChannelRepository) encryptMerchant(mch *entity.WxMerchantConfig) error {
	if r.encryptor == nil {
		return nil
	}
	fields := []*string{&mch.Apiv3Key, &mch.PrivateKey, &mch.PublicKeyID, &mch.PublicKey}
	for _, f := range fields {
		if *f == "" {
			continue
		}
		enc, err := r.encryptor.Encrypt(*f)
		if err != nil {
			return err
		}
		*f = enc
	}
	return nil
}

func (r *wxChannelRepository) decryptMerchant(mch *entity.WxMerchantConfig) error {
	if r.encryptor == nil {
		return nil
	}
	fields := []*string{&mch.Apiv3Key, &mch.PrivateKey, &mch.PublicKeyID, &mch.PublicKey}
	for _, f := range fields {
		if *f == "" {
			continue
		}
		plain, err := r.encryptor.Decrypt(*f)
		if err != nil {
			return err
		}
		*f = plain
	}
	return nil
}
