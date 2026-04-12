package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/utils/encryption"
)

// UserSeed 用户种子数据
type UserSeed struct {
	userRepo repo.UserRepository
	roleRepo repo.RoleRepository
}

// NewUserSeed 创建用户种子数据实例
func NewUserSeed(userRepo repo.UserRepository, roleRepo repo.RoleRepository) Seed {
	return &UserSeed{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Version 返回种子数据版本
func (us *UserSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (us *UserSeed) Name() string {
	return "user_seed"
}

// Order 返回初始化顺序（第三个执行）
func (us *UserSeed) Order() int {
	return 3
}

// Init 初始化用户种子数据
func (us *UserSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化用户种子数据")

	// 初始化 admin 用户
	adminUser, err := us.initAdminUser(logger)
	if err != nil {
		return err
	}
	logger.Info("admin 用户初始化完成", "id", adminUser.ID)

	// 初始化测试用户
	testUser, err := us.initTestUser(logger)
	if err != nil {
		return err
	}
	logger.Info("测试用户初始化完成", "id", testUser.ID)

	// 关联用户和角色
	return us.associateUsersWithRoles(logger, adminUser, testUser)
}

// initAdminUser 初始化 admin 用户
func (us *UserSeed) initAdminUser(logger logger.Logger) (*entity.User, error) {
	// 检查 admin 用户是否存在
	user, err := us.userRepo.GetByUsername("admin")
	if err == nil {
		// 用户已存在
		logger.Info("admin 用户已存在，跳过创建", "id", user.ID)
		return user, nil
	}

	// 生成密码哈希（使用默认密码'admin123'）
	passwordHash, err := encryption.GeneratePasswordHash("admin123")
	if err != nil {
		logger.Error("生成密码哈希失败，使用默认哈希", "error", err)
		// 使用备用密码哈希
		passwordHash = "$argon2id$v=19$m=65536,t=1,p=2$LbteWOhUGQ78dPlhlIHpCg$dxYA6EKHTEEux4KhoAxQaawnFiK6cnKnucpeGtEiirU" // admin123 的哈希
	}

	// 创建 admin 用户
	newUser := &entity.User{
		Username:     "admin",
		PasswordHash: passwordHash,
		Nickname:     "系统管理员",
		Email:        "admin@ixpay.com",
		Status:       1,
	}

	if err := us.userRepo.Create(newUser); err != nil {
		logger.Error("创建 admin 用户失败", "error", err)
		return nil, err
	}

	logger.Info("创建 admin 用户成功", "id", newUser.ID)
	return newUser, nil
}

// initTestUser 初始化测试用户
func (us *UserSeed) initTestUser(logger logger.Logger) (*entity.User, error) {
	// 检查测试用户是否存在
	user, err := us.userRepo.GetByUsername("test")
	if err == nil {
		// 用户已存在
		logger.Info("测试用户已存在，跳过创建", "id", user.ID)
		return user, nil
	}

	// 生成密码哈希（使用默认密码'test123'）
	passwordHash, err := encryption.GeneratePasswordHash("test123")
	if err != nil {
		logger.Error("生成密码哈希失败，使用默认哈希", "error", err)
		// 使用备用密码哈希
		passwordHash = "$argon2id$v=19$m=65536,t=1,p=2$LbteWOhUGQ78dPlhlIHpCg$dxYA6EKHTEEux4KhoAxQaawnFiK6cnKnucpeGtEiirU" // test123 的哈希（占位，实际会重新生成）
	}

	// 创建测试用户
	newUser := &entity.User{
		Username:     "test",
		PasswordHash: passwordHash,
		Nickname:     "测试用户",
		Email:        "test@ixpay.com",
		Status:       1,
	}

	if err := us.userRepo.Create(newUser); err != nil {
		logger.Error("创建测试用户失败", "error", err)
		return nil, err
	}

	logger.Info("创建测试用户成功", "id", newUser.ID)
	return newUser, nil
}

// associateUsersWithRoles 关联用户和角色
func (us *UserSeed) associateUsersWithRoles(logger logger.Logger, adminUser, testUser *entity.User) error {
	// 查找 admin 角色
	adminRole, err := us.roleRepo.GetByCode("admin")
	if err != nil {
		logger.Error("查找 admin 角色失败", "error", err)
		return err
	}

	// 查找普通用户角色
	userRole, err := us.roleRepo.GetByCode("user")
	if err != nil {
		logger.Error("查找普通用户角色失败", "error", err)
		return err
	}

	// 关联 admin 用户到 admin 角色
	if err := us.roleRepo.AddUserToRole(adminRole.ID, adminUser.ID); err != nil {
		logger.Error("关联 admin 用户到 admin 角色失败", "error", err)
		return err
	}
	logger.Info("admin 用户成功关联到 admin 角色")

	// 关联 admin 用户到普通用户角色
	if err := us.roleRepo.AddUserToRole(userRole.ID, adminUser.ID); err != nil {
		logger.Error("关联 admin 用户到普通用户角色失败", "error", err)
		return err
	}
	logger.Info("admin 用户成功关联到普通用户角色")

	// 关联测试用户到普通用户角色
	if err := us.roleRepo.AddUserToRole(userRole.ID, testUser.ID); err != nil {
		logger.Error("关联测试用户到普通用户角色失败", "error", err)
		return err
	}
	logger.Info("测试用户成功关联到普通用户角色")

	return nil
}
