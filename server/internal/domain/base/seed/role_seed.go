package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// RoleSeed 角色种子数据
type RoleSeed struct {
	roleRepo repo.RoleRepository
}

// NewRoleSeed 创建角色种子数据实例
func NewRoleSeed(roleRepo repo.RoleRepository) Seed {
	return &RoleSeed{
		roleRepo: roleRepo,
	}
}

// Version 返回种子数据版本
func (rs *RoleSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (rs *RoleSeed) Name() string {
	return "role_seed"
}

// Order 返回初始化顺序（第二个执行）
func (rs *RoleSeed) Order() int {
	return 2
}

// Init 初始化角色种子数据
func (rs *RoleSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化角色种子数据")

	// 初始化 admin 角色
	adminRole, err := rs.initAdminRole(logger)
	if err != nil {
		return err
	}
	logger.Info("admin 角色初始化完成", "id", adminRole.ID)

	// 初始化普通用户角色
	userRole, err := rs.initUserRole(logger)
	if err != nil {
		return err
	}
	logger.Info("普通用户角色初始化完成", "id", userRole.ID)

	// 注意：admin 角色和 user 角色都不需要在种子数据中关联 API 权限
	// - admin 角色：鉴权系统自动处理，管理员拥有所有权限
	// - user 角色：AuthType=0 的基础 API 所有登录用户都可访问，无需关联
	// 只有 AuthType=1 的 API 才需要在种子数据中关联到具体角色

	return nil
}

// initAdminRole 初始化 admin 角色
func (rs *RoleSeed) initAdminRole(logger logger.Logger) (*entity.Role, error) {
	// 1. 先检查 code 是否存在
	role, err := rs.roleRepo.GetByCode("admin")
	if err == nil {
		logger.Info("admin 角色已存在，跳过创建", "id", role.ID, "name", role.Name, "code", role.Code)
		return role, nil
	}

	// 2. 检查 name 是否存在
	role, err = rs.roleRepo.GetByName("管理员")
	if err == nil {
		logger.Info("角色名'管理员'已存在，跳过创建", "id", role.ID, "name", role.Name, "code", role.Code)
		return role, nil
	}

	// 3. 只有当 code 和 name 都不存在时才创建
	newRole := &entity.Role{
		Name:        "管理员",
		Code:        "admin",
		Description: "系统管理员，拥有所有权限",
		Status:      1,
		IsSystem:    true,
	}

	if err := rs.roleRepo.Create(newRole); err != nil {
		logger.Error("创建 admin 角色失败", "error", err)
		return nil, err
	}

	logger.Info("创建 admin 角色成功", "id", newRole.ID)
	return newRole, nil
}

// initUserRole 初始化普通用户角色
func (rs *RoleSeed) initUserRole(logger logger.Logger) (*entity.Role, error) {
	// 1. 先检查 code 是否存在
	role, err := rs.roleRepo.GetByCode("user")
	if err == nil {
		logger.Info("普通用户角色已存在，跳过创建", "id", role.ID, "name", role.Name, "code", role.Code)
		return role, nil
	}

	// 2. 检查 name 是否存在
	role, err = rs.roleRepo.GetByName("普通用户")
	if err == nil {
		logger.Info("角色名'普通用户'已存在，跳过创建", "id", role.ID, "name", role.Name, "code", role.Code)
		return role, nil
	}

	// 3. 只有当 code 和 name 都不存在时才创建
	newRole := &entity.Role{
		Name:        "普通用户",
		Code:        "user",
		Description: "普通用户角色，拥有基础权限",
		Status:      1,
		IsSystem:    false,
	}

	if err := rs.roleRepo.Create(newRole); err != nil {
		logger.Error("创建普通用户角色失败", "error", err)
		return nil, err
	}

	logger.Info("创建普通用户角色成功", "id", newRole.ID)
	return newRole, nil
}
