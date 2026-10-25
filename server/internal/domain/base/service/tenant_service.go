package service

import (
	"fmt"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/cryptoutil"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/support/task"
)

// TenantService 租户生命周期服务（管理平台一键建租户/初始化/重试/启停/授权/路由）
// 库名规则：ixpay_t_{tenant_code}（编码白名单校验防 DDL 注入，见设计文档 01 §4.2）
type TenantService struct {
	repo      repo.TenantRepository
	manager   *database.TenantManager
	tasks     *task.TaskManager
	encryptor *cryptoutil.AESGCMEncryptor
	log       logger.Logger
}

// NewTenantService 创建租户服务
func NewTenantService(
	tenantRepo repo.TenantRepository,
	manager *database.TenantManager,
	tasks *task.TaskManager,
	cfg *config.Config,
	log logger.Logger,
) *TenantService {
	encryptor, err := cryptoutil.NewAESGCMEncryptor(cfg.Security.ConfigEncryptKey)
	if err != nil {
		log.Error("租户服务加密器初始化失败（独立实例密码将无法加密保存）", "error", err)
	}
	return &TenantService{
		repo:      tenantRepo,
		manager:   manager,
		tasks:     tasks,
		encryptor: encryptor,
		log:       log,
	}
}

// CreateTenant 创建租户（status=init → 投递 InitTenantTask 异步建库+迁移）
// dbPassword 为明文输入，内部加密后落库
func (s *TenantService) CreateTenant(tenantCode, tenantName string, dbHost string, dbPort int, dbUser, dbPassword, description string, operatorID int64) (*entity.Tenant, error) {
	// 编码白名单校验（保留 default 供默认租户使用）
	if !TenantCodePattern.MatchString(tenantCode) {
		return nil, fmt.Errorf("租户编码不合法（需匹配 ^[a-z][a-z0-9_]{2,30}$，且以字母开头）")
	}
	if tenantCode == "default" {
		return nil, fmt.Errorf("default 为默认租户保留编码，不可使用")
	}
	if tenantName == "" {
		return nil, fmt.Errorf("租户名称不能为空")
	}
	// 重复校验
	if existing, err := s.repo.GetByCode(tenantCode); err == nil && existing != nil {
		return nil, fmt.Errorf("租户编码已存在: %s", tenantCode)
	}

	// 独立实例密码加密落库（明文仅存在于本函数内存）
	dbPassEnc := ""
	if dbPassword != "" {
		if s.encryptor == nil {
			return nil, fmt.Errorf("配置加密器不可用，无法保存独立实例密码")
		}
		enc, err := s.encryptor.Encrypt(dbPassword)
		if err != nil {
			return nil, fmt.Errorf("独立实例密码加密失败: %w", err)
		}
		dbPassEnc = enc
	}

	tenant := &entity.Tenant{
		TenantCode:  tenantCode,
		TenantName:  tenantName,
		DBName:      "ixpay_t_" + tenantCode,
		DBHost:      dbHost,
		DBPort:      dbPort,
		DBUser:      dbUser,
		DBPassword:  dbPassEnc,
		Status:      entity.TenantStatusInit,
		Description: description,
	}
	if err := s.repo.Create(tenant); err != nil {
		return nil, fmt.Errorf("创建租户记录失败: %w", err)
	}
	_ = s.repo.AddOpLog(&entity.TenantOpLog{
		TenantID:   tenant.ID,
		Action:     "create",
		Detail:     fmt.Sprintf("创建租户 %s（库名 %s）", tenantCode, tenant.DBName),
		OperatorID: operatorID,
	})

	// 投递初始化一次性任务（建库+迁移，失败自动置 failed 可重试）
	s.tasks.AddOneTimeTask(NewInitTenantTask(tenant.ID, s.repo, s.manager, s.log), 0)

	s.log.Info("租户已创建，初始化任务已投递", "tenant_id", tenant.ID, "code", tenantCode)
	return tenant, nil
}

// UpdateTenant 更新租户基础信息（名称/独立实例连接/描述）；dbPassword 非空时更新加密密码
func (s *TenantService) UpdateTenant(id int64, tenantName string, dbHost string, dbPort int, dbUser, dbPassword, description string, operatorID int64) (*entity.Tenant, error) {
	tenant, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("租户不存在(id=%d)", id)
	}
	if tenant.ID == 1 {
		return nil, fmt.Errorf("默认租户不允许修改")
	}
	if tenantName != "" {
		tenant.TenantName = tenantName
	}
	tenant.DBHost = dbHost
	tenant.DBPort = dbPort
	tenant.DBUser = dbUser
	if dbPassword != "" {
		if s.encryptor == nil {
			return nil, fmt.Errorf("配置加密器不可用，无法保存独立实例密码")
		}
		enc, err := s.encryptor.Encrypt(dbPassword)
		if err != nil {
			return nil, fmt.Errorf("独立实例密码加密失败: %w", err)
		}
		tenant.DBPassword = enc
	}
	tenant.Description = description
	if err := s.repo.Update(tenant); err != nil {
		return nil, fmt.Errorf("更新租户失败: %w", err)
	}
	_ = s.repo.AddOpLog(&entity.TenantOpLog{
		TenantID:   id,
		Action:     "update",
		Detail:     fmt.Sprintf("更新租户基本信息 %s", tenant.TenantCode),
		OperatorID: operatorID,
	})

	// 连接信息变更后驱逐旧连接池，下次访问按新配置重建
	s.manager.Evict(id)
	return tenant, nil
}

// RetryInit 重试/重跑租户库初始化（failed 重试；active 重跑即增量迁移升级，见设计文档 01 §5.1）
func (s *TenantService) RetryInit(id int64, operatorID int64) error {
	tenant, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("租户不存在(id=%d)", id)
	}
	if tenant.ID == 1 {
		return fmt.Errorf("默认租户无需初始化")
	}
	switch tenant.Status {
	case entity.TenantStatusInit:
		return fmt.Errorf("初始化进行中，请稍后")
	case entity.TenantStatusDisabled, entity.TenantStatusArchived:
		return fmt.Errorf("租户已禁用/归档，请先启用再重试")
	}

	tenant.Status = entity.TenantStatusInit
	tenant.InitError = ""
	if err := s.repo.Update(tenant); err != nil {
		return fmt.Errorf("重置租户状态失败: %w", err)
	}
	_ = s.repo.AddOpLog(&entity.TenantOpLog{
		TenantID:   id,
		Action:     "retry",
		Detail:     "重新投递租户库初始化任务",
		OperatorID: operatorID,
	})

	s.tasks.AddOneTimeTask(NewInitTenantTask(id, s.repo, s.manager, s.log), 0)
	return nil
}

// DisableTenant 禁用租户（驱逐连接池；登录/支付由状态校验拒绝）
func (s *TenantService) DisableTenant(id int64, operatorID int64) error {
	return s.changeStatus(id, entity.TenantStatusDisabled, "disable", operatorID)
}

// EnableTenant 启用租户（恢复 active）
func (s *TenantService) EnableTenant(id int64, operatorID int64) error {
	return s.changeStatus(id, entity.TenantStatusActive, "enable", operatorID)
}

// ArchiveTenant 归档租户（只读保留；物理删库仅运维 CLI，见设计文档 01 §4.1）
func (s *TenantService) ArchiveTenant(id int64, operatorID int64) error {
	return s.changeStatus(id, entity.TenantStatusArchived, "archive", operatorID)
}

// changeStatus 状态流转（init/failed 状态需走初始化/重试通道）
func (s *TenantService) changeStatus(id int64, status, action string, operatorID int64) error {
	tenant, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("租户不存在(id=%d)", id)
	}
	if tenant.ID == 1 && status != entity.TenantStatusActive {
		return fmt.Errorf("默认租户不允许禁用/归档")
	}
	if tenant.Status == entity.TenantStatusInit {
		return fmt.Errorf("初始化进行中，暂不能变更状态")
	}

	tenant.Status = status
	if err := s.repo.Update(tenant); err != nil {
		return fmt.Errorf("更新租户状态失败: %w", err)
	}
	_ = s.repo.AddOpLog(&entity.TenantOpLog{
		TenantID:   id,
		Action:     action,
		Detail:     fmt.Sprintf("租户状态变更为 %s", status),
		OperatorID: operatorID,
	})

	// 禁用/归档立即驱逐连接池（启用后按需重建）
	if status != entity.TenantStatusActive {
		s.manager.Evict(id)
	}
	return nil
}

// ListTenants 租户列表（分页+状态过滤）
func (s *TenantService) ListTenants(page, pageSize int, status string) ([]*entity.Tenant, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(page, pageSize, status)
}

// ListVisibleTenants 当前用户可见租户（管理端租户切换器数据源，见设计文档 01 §3.4）
// admin 全量；普通用户按 base_tenant_grants 授权列表
func (s *TenantService) ListVisibleTenants(userID int64, isAdmin bool) ([]*entity.Tenant, error) {
	if isAdmin {
		tenants, _, err := s.repo.List(1, 100, "")
		return tenants, err
	}
	return s.repo.ListUserGrantTenants(userID)
}

// GetTenant 租户详情
func (s *TenantService) GetTenant(id int64) (*entity.Tenant, error) {
	return s.repo.GetByID(id)
}

// ---- 应用路由（app_key → 租户）----

// ListAppRoutes 租户下应用路由列表
func (s *TenantService) ListAppRoutes(tenantID int64) ([]*entity.TenantApp, error) {
	return s.repo.ListAppRoutes(tenantID)
}

// CreateAppRoute 绑定 app_key 到租户（app_key 唯一，绑定后该前端实例的登录请求路由到该租户）
func (s *TenantService) CreateAppRoute(appKey string, tenantID int64, moduleCode string, operatorID int64) (*entity.TenantApp, error) {
	if appKey == "" {
		return nil, fmt.Errorf("app_key 不能为空")
	}
	if moduleCode == "" {
		moduleCode = "wx"
	}
	if _, err := s.repo.GetByID(tenantID); err != nil {
		return nil, fmt.Errorf("租户不存在(id=%d)", tenantID)
	}
	if existing, err := s.repo.GetAppRoute(appKey); err == nil && existing != nil {
		return nil, fmt.Errorf("app_key 已绑定租户(tenant=%d): %s", existing.TenantID, appKey)
	}
	route := &entity.TenantApp{
		AppKey:     appKey,
		TenantID:   tenantID,
		ModuleCode: moduleCode,
	}
	if err := s.repo.CreateAppRoute(route); err != nil {
		return nil, fmt.Errorf("创建应用路由失败: %w", err)
	}
	_ = s.repo.AddOpLog(&entity.TenantOpLog{
		TenantID:   tenantID,
		Action:     "init_app",
		Detail:     fmt.Sprintf("绑定应用实例 app_key=%s（模块 %s）", appKey, moduleCode),
		OperatorID: operatorID,
	})
	return route, nil
}

// DeleteAppRoute 解绑应用路由
func (s *TenantService) DeleteAppRoute(routeID int64, operatorID int64) error {
	if err := s.repo.DeleteAppRoute(routeID); err != nil {
		return fmt.Errorf("删除应用路由失败: %w", err)
	}
	_ = s.repo.AddOpLog(&entity.TenantOpLog{
		TenantID:   0,
		Action:     "remove_app",
		Detail:     fmt.Sprintf("解绑应用路由 id=%d", routeID),
		OperatorID: operatorID,
	})
	return nil
}

// ---- 用户授权（非 admin 用户按授权查租户）----

// ListGrants 租户授权用户列表
func (s *TenantService) ListGrants(tenantID int64) ([]*entity.TenantGrant, error) {
	return s.repo.ListGrants(tenantID)
}

// CreateGrant 授权用户访问租户数据（admin 天然全量，无需授权）
func (s *TenantService) CreateGrant(userID, tenantID int64, operatorID int64) (*entity.TenantGrant, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("用户ID不合法")
	}
	if _, err := s.repo.GetByID(tenantID); err != nil {
		return nil, fmt.Errorf("租户不存在(id=%d)", tenantID)
	}
	if existing, err := s.repo.GetGrant(userID, tenantID); err == nil && existing != nil {
		return nil, fmt.Errorf("该用户已被授权此租户")
	}
	grant := &entity.TenantGrant{UserID: userID, TenantID: tenantID}
	if err := s.repo.CreateGrant(grant); err != nil {
		return nil, fmt.Errorf("创建授权失败: %w", err)
	}
	_ = s.repo.AddOpLog(&entity.TenantOpLog{
		TenantID:   tenantID,
		Action:     "grant",
		Detail:     fmt.Sprintf("授权用户 %d 查看租户数据", userID),
		OperatorID: operatorID,
	})
	return grant, nil
}

// DeleteGrant 取消授权
func (s *TenantService) DeleteGrant(userID, tenantID int64, operatorID int64) error {
	if err := s.repo.DeleteGrant(userID, tenantID); err != nil {
		return fmt.Errorf("取消授权失败: %w", err)
	}
	_ = s.repo.AddOpLog(&entity.TenantOpLog{
		TenantID:   tenantID,
		Action:     "revoke",
		Detail:     fmt.Sprintf("取消用户 %d 的租户授权", userID),
		OperatorID: operatorID,
	})
	return nil
}

// ListOpLogs 租户运维操作审计（最近的 limit 条）
func (s *TenantService) ListOpLogs(tenantID int64, limit int) ([]*entity.TenantOpLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListOpLogs(tenantID, limit)
}
