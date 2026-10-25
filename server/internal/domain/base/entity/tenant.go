package entity

import "time"

// 租户状态
const (
	TenantStatusInit     = "init"     // 初始化中（建库/迁移进行时）
	TenantStatusActive   = "active"   // 正常可用
	TenantStatusDisabled = "disabled" // 已禁用（连接池驱逐+访问拒绝）
	TenantStatusArchived = "archived" // 已归档（只读保留）
	TenantStatusFailed   = "failed"   // 初始化失败（可重试）
)

// Tenant 租户实体（主库注册表；一个租户 = 一个独立业务库）
type Tenant struct {
	ID          int64     // 租户 ID
	TenantCode  string    // 租户编码（唯一，库名=ixpay_t_{code}）
	TenantName  string    // 租户名称
	DBName      string    // 业务库名（空 = 默认租户，使用主库）
	DBHost      string    // 独立实例 Host（空 = 沿用主实例）
	DBPort      int       // 独立实例 Port（0 = 沿用主实例）
	DBUser      string    // 独立实例用户（空 = 沿用主实例）
	DBPassword  string    // 独立实例密码（AES-GCM 密文存储）
	Status      string    // 状态：init/active/disabled/archived/failed
	InitError   string    // 初始化失败原因
	Description string    // 描述
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}

// TenantApp 租户应用路由（app_key → 租户；登录前在主库完成路由）
type TenantApp struct {
	ID         int64     // 路由 ID
	AppKey     string    // 前端实例标志（唯一，对应 wx_app_configs.app_key）
	TenantID   int64     // 租户 ID
	ModuleCode string    // 所属模块：wx / ali / wecom
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
}

// TenantGrant 租户授权（管理平台用户 ↔ 租户，多对多；admin 天然全量）
type TenantGrant struct {
	ID        int64     // 授权 ID
	UserID    int64     // 管理平台用户 ID
	TenantID  int64     // 租户 ID
	CreatedAt time.Time // 创建时间
}

// TenantOpLog 租户运维操作审计
type TenantOpLog struct {
	ID         int64     // 日志 ID
	TenantID   int64     // 租户 ID（0 = 平台级操作）
	Action     string    // 动作：create/init/retry/disable/enable/grant/revoke/init_app
	Detail     string    // 详情
	OperatorID int64     // 操作人 ID
	CreatedAt  time.Time // 时间
}
