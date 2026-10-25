package persistence

import (
	"context"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// tenantModel 租户数据库模型（base_tenants 表）
type tenantModel struct {
	database.SnowflakeBaseModel
	TenantCode  string     `gorm:"size:50;not null;uniqueIndex"`
	TenantName  string     `gorm:"size:100;not null"`
	DBName      string     `gorm:"size:63;not null;default:''"`
	DBHost      string     `gorm:"size:100;not null;default:''"`
	DBPort      int        `gorm:"not null;default:0"`
	DBUser      string     `gorm:"size:63;not null;default:''"`
	DBPassword  string     `gorm:"type:text;not null;default:''"` // AES-GCM 密文
	Status      string     `gorm:"size:20;not null;default:'init';index"`
	InitError   string     `gorm:"type:text;not null;default:''"`
	Description string     `gorm:"size:255;not null;default:''"`
	DeletedAt   *time.Time `gorm:"index"`
}

func (tenantModel) TableName() string { return "base_tenants" }

func (m *tenantModel) toDomain() *entity.Tenant {
	if m == nil {
		return nil
	}
	return &entity.Tenant{
		ID:          m.ID,
		TenantCode:  m.TenantCode,
		TenantName:  m.TenantName,
		DBName:      m.DBName,
		DBHost:      m.DBHost,
		DBPort:      m.DBPort,
		DBUser:      m.DBUser,
		DBPassword:  m.DBPassword,
		Status:      m.Status,
		InitError:   m.InitError,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func fromDomainTenant(t *entity.Tenant) *tenantModel {
	return &tenantModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID: t.ID,
			// 带上原值：Create 时为零值由 GORM autoCreateTime 自动填充；
			// Update(Save) 时回写数据库原值，避免审计时间被覆盖
			CreatedAt: t.CreatedAt,
		},
		TenantCode:  t.TenantCode,
		TenantName:  t.TenantName,
		DBName:      t.DBName,
		DBHost:      t.DBHost,
		DBPort:      t.DBPort,
		DBUser:      t.DBUser,
		DBPassword:  t.DBPassword,
		Status:      t.Status,
		InitError:   t.InitError,
		Description: t.Description,
	}
}

// tenantAppModel 租户应用路由模型（base_tenant_apps 表）
type tenantAppModel struct {
	database.SnowflakeBaseModel
	AppKey     string `gorm:"size:50;not null;uniqueIndex"`
	TenantID   int64  `gorm:"not null;index"`
	ModuleCode string `gorm:"size:20;not null;default:'wx'"`
}

func (tenantAppModel) TableName() string { return "base_tenant_apps" }

func (m *tenantAppModel) toDomain() *entity.TenantApp {
	if m == nil {
		return nil
	}
	return &entity.TenantApp{
		ID:         m.ID,
		AppKey:     m.AppKey,
		TenantID:   m.TenantID,
		ModuleCode: m.ModuleCode,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// tenantGrantModel 租户授权模型（base_tenant_grants 表）
type tenantGrantModel struct {
	database.SnowflakeBaseModel
	UserID   int64 `gorm:"not null;index"`
	TenantID int64 `gorm:"not null;index"`
}

func (tenantGrantModel) TableName() string { return "base_tenant_grants" }

func (m *tenantGrantModel) toDomain() *entity.TenantGrant {
	if m == nil {
		return nil
	}
	return &entity.TenantGrant{
		ID:        m.ID,
		UserID:    m.UserID,
		TenantID:  m.TenantID,
		CreatedAt: m.CreatedAt,
	}
}

// tenantOpLogModel 租户运维审计模型（base_tenant_ops_log 表）
type tenantOpLogModel struct {
	database.SnowflakeBaseModel
	TenantID   int64  `gorm:"not null;default:0;index"`
	Action     string `gorm:"size:50;not null"`
	Detail     string `gorm:"type:text;not null;default:''"`
	OperatorID int64  `gorm:"not null;default:0"`
}

func (tenantOpLogModel) TableName() string { return "base_tenant_ops_log" }

func (m *tenantOpLogModel) toDomain() *entity.TenantOpLog {
	if m == nil {
		return nil
	}
	return &entity.TenantOpLog{
		ID:         m.ID,
		TenantID:   m.TenantID,
		Action:     m.Action,
		Detail:     m.Detail,
		OperatorID: m.OperatorID,
		CreatedAt:  m.CreatedAt,
	}
}

// tenantRepository 租户仓储实现（主库）
type tenantRepository struct {
	db *database.PostgresDB
}

var _ repo.TenantRepository = (*tenantRepository)(nil)

// NewTenantRepository 创建租户仓储实现
func NewTenantRepository(db *database.PostgresDB) repo.TenantRepository {
	return &tenantRepository{db: db}
}

// ---------- 租户 ----------

func (r *tenantRepository) GetByID(id int64) (*entity.Tenant, error) {
	var m tenantModel
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *tenantRepository) GetByCode(code string) (*entity.Tenant, error) {
	var m tenantModel
	if err := r.db.Where("tenant_code = ? AND deleted_at IS NULL", code).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *tenantRepository) List(page, pageSize int, status string) ([]*entity.Tenant, int, error) {
	var total int64
	var models []tenantModel
	query := r.db.Model(&tenantModel{}).Where("deleted_at IS NULL")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id ASC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	tenants := make([]*entity.Tenant, len(models))
	for i := range models {
		tenants[i] = models[i].toDomain()
	}
	return tenants, int(total), nil
}

func (r *tenantRepository) Create(tenant *entity.Tenant) error {
	m := fromDomainTenant(tenant)
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	tenant.ID = m.ID
	return nil
}

func (r *tenantRepository) Update(tenant *entity.Tenant) error {
	// Omit 审计列：更新业务字段时不覆盖 created_at/created_by 等创建审计信息
	return r.db.Omit("created_at", "created_by", "updated_by", "deleted_by", "deleted_at").
		Save(fromDomainTenant(tenant)).Error
}

func (r *tenantRepository) Delete(id int64) error {
	return r.db.Delete(&tenantModel{}, id).Error
}

// ---------- 应用路由 ----------

func (r *tenantRepository) GetAppRoute(appKey string) (*entity.TenantApp, error) {
	var m tenantAppModel
	if err := r.db.Where("app_key = ?", appKey).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *tenantRepository) CreateAppRoute(route *entity.TenantApp) error {
	m := &tenantAppModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{ID: route.ID},
		AppKey:             route.AppKey,
		TenantID:           route.TenantID,
		ModuleCode:         route.ModuleCode,
	}
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	route.ID = m.ID
	return nil
}

// DeleteAppRoute 解绑应用路由（硬删除）
// 说明：app_key 有全局唯一约束，软删除的残留行会持续占用 app_key 导致
// "解绑后重绑"撞唯一索引（duplicate key base_tenant_apps_app_key_key）；
// 路由为纯映射关系（app_key→tenant_id），操作留痕由租户操作审计（base_tenant_ops_log）承担
func (r *tenantRepository) DeleteAppRoute(id int64) error {
	return r.db.Unscoped().Delete(&tenantAppModel{}, id).Error
}

func (r *tenantRepository) ListAppRoutes(tenantID int64) ([]*entity.TenantApp, error) {
	var models []tenantAppModel
	if err := r.db.Where("tenant_id = ?", tenantID).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	routes := make([]*entity.TenantApp, len(models))
	for i := range models {
		routes[i] = models[i].toDomain()
	}
	return routes, nil
}

// ---------- 用户授权 ----------

func (r *tenantRepository) ListGrants(tenantID int64) ([]*entity.TenantGrant, error) {
	var models []tenantGrantModel
	if err := r.db.Where("tenant_id = ?", tenantID).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	grants := make([]*entity.TenantGrant, len(models))
	for i := range models {
		grants[i] = models[i].toDomain()
	}
	return grants, nil
}

func (r *tenantRepository) GetGrant(userID, tenantID int64) (*entity.TenantGrant, error) {
	var m tenantGrantModel
	if err := r.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *tenantRepository) ListUserGrantTenants(userID int64) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	if err := r.db.Raw(`
		SELECT t.* FROM base_tenants t
		JOIN base_tenant_grants g ON g.tenant_id = t.id
		WHERE g.user_id = ? AND t.deleted_at IS NULL
		ORDER BY t.id ASC
	`, userID).Scan(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *tenantRepository) CreateGrant(grant *entity.TenantGrant) error {
	m := &tenantGrantModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{ID: grant.ID},
		UserID:             grant.UserID,
		TenantID:           grant.TenantID,
	}
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	grant.ID = m.ID
	return nil
}

func (r *tenantRepository) DeleteGrant(userID, tenantID int64) error {
	return r.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Delete(&tenantGrantModel{}).Error
}

// UpsertMchRef 登记商户号→租户映射（渠道配置保存/更新时调用；支付回调按 Wechatpay-Serial 定位租户）
func (r *tenantRepository) UpsertMchRef(mchID, publicKeyID string, tenantID int64) error {
	if mchID == "" {
		return nil
	}
	return r.db.Exec(`
		INSERT INTO base_tenant_mch_refs (mch_id, public_key_id, tenant_id, updated_at)
		VALUES (?, ?, ?, NOW())
		ON CONFLICT (mch_id) DO UPDATE SET public_key_id = EXCLUDED.public_key_id, tenant_id = EXCLUDED.tenant_id, updated_at = NOW()
	`, mchID, publicKeyID, tenantID).Error
}

// DeleteMchRef 删除商户映射（商户配置删除时调用）
func (r *tenantRepository) DeleteMchRef(mchID string) error {
	if mchID == "" {
		return nil
	}
	return r.db.Exec(`DELETE FROM base_tenant_mch_refs WHERE mch_id = ?`, mchID).Error
}

// GetTenantByPublicKeyID 按微信支付公钥ID（回调头 Wechatpay-Serial）定位商户归属租户
// 未登记返回 0（回退默认租户/主库全局客户端，兼容存量单商户模式）
func (r *tenantRepository) GetTenantByPublicKeyID(publicKeyID string) (int64, error) {
	if publicKeyID == "" {
		return 0, nil
	}
	var tenantID int64
	err := r.db.Raw(
		`SELECT tenant_id FROM base_tenant_mch_refs WHERE public_key_id = ? LIMIT 1`, publicKeyID,
	).Scan(&tenantID).Error
	if err != nil {
		return 0, err
	}
	return tenantID, nil
}

// ---------- 运维审计 ----------

func (r *tenantRepository) AddOpLog(log *entity.TenantOpLog) error {
	m := &tenantOpLogModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{ID: log.ID},
		TenantID:           log.TenantID,
		Action:             log.Action,
		Detail:             log.Detail,
		OperatorID:         log.OperatorID,
	}
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	log.ID = m.ID
	return nil
}

func (r *tenantRepository) ListOpLogs(tenantID int64, limit int) ([]*entity.TenantOpLog, error) {
	var models []tenantOpLogModel
	q := r.db.Model(&tenantOpLogModel{})
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if err := q.Order("id DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	logs := make([]*entity.TenantOpLog, len(models))
	for i := range models {
		logs[i] = models[i].toDomain()
	}
	return logs, nil
}

// GetTenantConnInfo 查询租户连接信息（database/sql 层使用；ctx 透传）
func (r *tenantRepository) GetTenantConnInfo(ctx context.Context, tenantID int64) (*entity.Tenant, error) {
	var m tenantModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", tenantID).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}
