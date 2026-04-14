package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	auth "github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/captcha"
	"github.com/ix-pay/ixpay-pro/internal/utils/encryption"
	"gorm.io/gorm"
)

// service包实现用户领域的业务逻辑和服务接口
// 包含用户注册、登录、信息管理等核心业务功能
// 与数据访问层和API控制器层交互，实现完整的业务流程

// UserService 实现用户领域服务接口
// 封装用户相关的业务逻辑和操作
// 字段:
// - repo: 用户数据仓库，提供数据访问能力
// - settingRepo: 用户设置数据仓库，提供用户设置数据访问能力
// - roleService: 角色服务，用于获取用户角色和权限
// - jwtAuth: JWT 认证服务，用于生成和验证令牌
// - config: 应用配置，包含系统设置
// - log: 日志记录器，用于记录操作日志
// - cache: 缓存服务，用于缓存和会话管理
// - captcha: 验证码服务，用于验证码生成和验证
// - loginLogService: 登录日志服务，用于记录用户登录日志
type UserService struct {
	repo                  repo.UserRepository        // 用户数据仓库
	settingRepo           repo.UserSettingRepository // 用户设置数据仓库
	roleService           *RoleService               // 角色服务
	rolePermissionService *RolePermissionService     // 角色权限服务
	jwtAuth               *auth.JWTAuth              // JWT 认证服务
	config                *config.Config             // 应用配置
	log                   logger.Logger              // 日志记录器
	cache                 cache.Cache                // 缓存服务
	captcha               *captcha.Captcha           // 验证码服务
	loginLogService       *LoginLogService           // 登录日志服务
}

// NewUserService 创建用户服务实例
// 参数:
// - repo: 用户数据仓库，提供数据访问能力
// - settingRepo: 用户设置数据仓库，提供用户设置数据访问能力
// - roleService: 角色服务，用于获取用户角色和权限
// - rolePermissionService: 角色权限服务，用于管理角色权限缓存
// - jwtAuth: JWT 认证服务，用于生成和验证令牌
// - config: 应用配置，包含系统设置
// - log: 日志记录器，用于记录操作日志
// - cache: 缓存服务，用于缓存和会话管理
// - captcha: 验证码服务，用于验证码生成和验证
// - loginLogService: 登录日志服务，用于记录用户登录日志
// 返回:
// - *UserService: 用户服务实现
func NewUserService(repo repo.UserRepository, settingRepo repo.UserSettingRepository, roleService *RoleService, rolePermissionService *RolePermissionService, jwtAuth *auth.JWTAuth, config *config.Config, log logger.Logger, cache cache.Cache, captcha *captcha.Captcha, loginLogService *LoginLogService) *UserService {
	// 创建并返回用户服务实例，注入所有依赖
	return &UserService{
		repo:                  repo,
		settingRepo:           settingRepo,
		roleService:           roleService,
		rolePermissionService: rolePermissionService,
		jwtAuth:               jwtAuth,
		config:                config,
		log:                   log,
		cache:                 cache,
		captcha:               captcha,
		loginLogService:       loginLogService,
	}
}

// Captcha 生成验证码
// 调用验证码服务生成新的验证码图片
// 返回:
// - string: 验证码ID
// - string: base64编码的验证码图片
// - int: 验证码长度
// - bool: 验证码功能是否开启
// - error: 错误信息
func (s *UserService) Captcha() (string, string, int, bool, error) {
	// 调用验证码服务生成新的验证码
	id, base64Img, err := s.captcha.NewCaptcha()
	if err != nil {
		s.log.Error("生成验证码失败", "error", err)
		return id, "", s.captcha.GetCaptchaLen(), s.captcha.IsOpenCaptcha(), err
	}

	// 返回验证码ID、图片数据、长度、开关状态
	return id, base64Img, s.captcha.GetCaptchaLen(), s.captcha.IsOpenCaptcha(), nil
}

// Register 用户注册
// 创建新用户账户，包括密码加密和基本信息设置
// 参数:
// - username: 用户名
// - password: 密码（明文）
// - email: 电子邮箱
// 返回:
// - *entity.User: 新创建的用户对象
// - error: 错误信息
func (s *UserService) Register(username, password, email string) (*entity.User, error) {
	// 检查用户是否已存在
	_, err := s.repo.GetByUsername(username)
	if err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	_, err = s.repo.GetByEmail(email)
	if err == nil {
		return nil, errors.New("邮箱已存在")
	}

	// 生成密码哈希
	passwordHash, err := encryption.GeneratePasswordHash(password)
	if err != nil {
		s.log.Error("生成密码哈希失败", "error", err)
		return nil, err
	}

	// 创建新用户
	user := &entity.User{
		Username:     username,
		PasswordHash: passwordHash,
		Email:        email,
		Status:       1,
	}

	// 保存用户
	if err := s.repo.Create(user); err != nil {
		s.log.Error("创建用户失败", "error", err)
		return nil, err
	}

	// 为用户分配默认角色（user）
	if err := s.assignDefaultRole(user.ID); err != nil {
		s.log.Warn("为用户分配默认角色失败", "userID", user.ID, "error", err)
		// 不返回错误，因为用户创建已经成功
	}

	s.log.Info("用户注册成功", "username", username)
	return user, nil
}

// Login 用户登录
// 验证用户凭据并生成访问令牌
// 参数:
// - username: 用户名
// - password: 密码（明文）
// - captchaId: 验证码 ID
// - captchaVal: 用户输入的验证码值
// - ip: 用户登录 IP 地址
// - userAgent: 用户代理字符串
// 返回:
// - *entity.User: 登录成功的用户对象
// - string: 访问令牌
// - string: 刷新令牌
// - time.Time: 访问令牌过期时间
// - time.Time: 刷新令牌过期时间
// - error: 错误信息
func (s *UserService) Login(username, password, captchaId, captchaVal, ip, userAgent string) (*entity.User, string, string, time.Time, time.Time, error) {
	// 解析 User-Agent 获取浏览器和操作系统信息
	browser, os := parseUserAgent(userAgent)
	// 组合设备信息
	device := fmt.Sprintf("%s / %s", browser, os)
	// 获取登录地点（基于 IP 的简单定位）
	loginPlace := getLoginPlaceByIP(ip)

	// 检查是否开启验证码，如果开启则验证验证码
	if s.captcha.IsOpenCaptcha() {
		// 验证用户输入的验证码是否正确
		ok, err := s.captcha.VerifyCaptcha(captchaId, captchaVal)
		if err != nil {
			s.log.Error("获取验证码失败", "error", err)
			// 记录失败的登录日志（验证码错误）
			s.loginLogService.RecordLogin(0, username, ip, loginPlace, device, browser, os, userAgent, false, "验证码错误")
			return nil, "", "", time.Time{}, time.Time{}, errors.New("验证码已过期或无效")
		}

		if !ok {
			// 记录失败的登录日志（验证码错误）
			s.loginLogService.RecordLogin(0, username, ip, loginPlace, device, browser, os, userAgent, false, "验证码错误")
			return nil, "", "", time.Time{}, time.Time{}, errors.New("验证码错误")
		}
	}

	// 根据用户名获取用户
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		s.log.Error("查找用户失败", "error", err)
		// 记录失败的登录日志（用户不存在）
		s.loginLogService.RecordLogin(0, username, ip, loginPlace, device, browser, os, userAgent, false, "用户不存在")
		return nil, "", "", time.Time{}, time.Time{}, errors.New("用户名或密码错误")
	}

	// 验证密码
	if verifyErr := encryption.VerifyPassword(user.PasswordHash, password); verifyErr != nil {
		s.log.Error("密码验证失败", "username", username)
		// 记录失败的登录日志（密码错误）
		s.loginLogService.RecordLogin(user.ID, username, ip, loginPlace, device, browser, os, userAgent, false, "密码错误")
		return nil, "", "", time.Time{}, time.Time{}, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		// 记录失败的登录日志（用户未激活）
		s.loginLogService.RecordLogin(user.ID, username, ip, loginPlace, device, browser, os, userAgent, false, "用户账户未激活")
		return nil, "", "", time.Time{}, time.Time{}, errors.New("用户账户未激活")
	}

	// 加载用户的角色 ID 列表
	userRoles, err := s.roleService.GetRolesForUser(user.ID)
	if err != nil {
		s.log.Error("获取用户角色失败", "userID", user.ID, "error", err)
	} else {
		// 将角色 ID 填充到 user.RoleIds
		roleIds := make([]int64, len(userRoles))
		for i, role := range userRoles {
			roleIds[i] = role.ID
		}
		user.RoleIds = roleIds
	}

	// 生成令牌，获取用户角色
	role := "" // 默认角色
	var firstRoleID int64 = 0
	var firstRole *entity.Role
	if len(user.RoleIds) > 0 {
		// 获取用户角色详情
		userRoles, err := s.roleService.GetRolesForUser(user.ID)
		if err == nil && len(userRoles) > 0 {
			firstRole = userRoles[0]
			firstRoleID = firstRole.ID
			role = firstRole.Code // 使用第一个角色的 Code
		}
	}

	if role == "" {
		s.log.Error("用户未授权！", "userID", user.ID)
		return nil, "", "", time.Time{}, time.Time{}, nil
	}

	// 传入 nickname（如果有昵称则使用昵称，否则使用用户名）
	nickname := user.Nickname
	if nickname == "" {
		nickname = user.Username
	}
	// 【修改 1】生成 JWT 时 role 参数传空字符串
	accessToken, refreshToken, accessExpire, refreshExpire, err := s.jwtAuth.GenerateToken(fmt.Sprintf("%d", user.ID), user.Username, nickname, "", "password")
	if err != nil {
		s.log.Error("生成令牌失败", "error", err)
		// 记录失败的登录日志（令牌生成失败）
		s.loginLogService.RecordLogin(user.ID, username, ip, loginPlace, device, browser, os, userAgent, false, err.Error())
		return nil, "", "", time.Time{}, time.Time{}, err
	}

	s.log.Info("用户登录成功", "username", username)

	// 【修改 2】登录成功后，缓存默认角色（第一个角色）到 Redis
	if firstRoleID > 0 {
		// 使用 string 格式，与认证中间件保持一致
		currentRoleKey := fmt.Sprintf("user:current_role:%s", fmt.Sprintf("%d", user.ID))
		if cacheErr := s.cache.Set(currentRoleKey, fmt.Sprintf("%d", firstRoleID), 24*time.Hour); cacheErr != nil {
			s.log.Error("缓存用户默认角色失败", "error", cacheErr, "userID", user.ID, "roleID", firstRoleID)
			// 不阻塞主流程
		} else {
			s.log.Info("已缓存用户默认角色", "userID", user.ID, "roleID", firstRoleID)
		}

		// 【修改 3】缓存角色详情到 Redis
		// 使用 string 格式，与认证中间件保持一致
		roleCacheKey := fmt.Sprintf("role:info:%s", fmt.Sprintf("%d", firstRoleID))
		simpleRole := map[string]interface{}{
			"ID":     fmt.Sprintf("%d", firstRole.ID),
			"Code":   firstRole.Code,
			"Name":   firstRole.Name,
			"Status": firstRole.Status,
		}
		roleData, marshalErr := json.Marshal(simpleRole)
		if marshalErr == nil {
			if cacheErr := s.cache.Set(roleCacheKey, string(roleData), 24*time.Hour); cacheErr != nil {
				s.log.Error("缓存角色详情失败", "error", cacheErr, "roleID", firstRoleID)
				// 不阻塞主流程
			} else {
				s.log.Info("已缓存角色详情", "roleID", firstRoleID, "roleCode", firstRole.Code, "roleName", firstRole.Name)
			}
		} else {
			s.log.Error("序列化角色详情失败", "error", marshalErr, "roleID", firstRoleID)
		}
	}

	// 记录成功的登录日志
	s.loginLogService.RecordLogin(user.ID, username, ip, loginPlace, device, browser, os, userAgent, true, "")
	return user, accessToken, refreshToken, accessExpire, refreshExpire, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userID int64) (*entity.User, error) {
	user, err := s.repo.GetByID(userID, repo.DEPARTMENT, repo.POSITION, repo.ROLES)

	if err != nil {
		s.log.Error("获取用户信息失败", "error", err)
		return nil, err
	}
	return user, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(user *entity.User, updatedBy int64) error {
	// 验证 updatedBy 参数不能为空
	if updatedBy == 0 {
		return errors.New("updatedBy 不能为空")
	}
	user.UpdatedBy = updatedBy
	if err := s.repo.Update(user); err != nil {
		s.log.Error("更新用户信息失败", "error", err)
		return err
	}
	s.log.Info("用户信息更新成功", "userID", user.ID)
	return nil
}

// UpdateUserDepartment 更新用户部门
func (s *UserService) UpdateUserDepartment(userID int64, departmentID int64, updatedBy int64) error {
	// 获取用户
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("查找用户失败", "error", err)
		return errors.New("用户不存在")
	}

	// 更新部门信息
	user.DepartmentID = departmentID
	user.UpdatedBy = updatedBy

	if err := s.repo.Update(user); err != nil {
		s.log.Error("更新用户部门失败", "error", err)
		return err
	}

	s.log.Info("用户部门更新成功", "userID", userID, "departmentID", departmentID, "updatedBy", updatedBy)
	return nil
}

// UpdateUserPosition 更新用户岗位
func (s *UserService) UpdateUserPosition(userID int64, positionID int64, updatedBy int64) error {
	// 获取用户
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("查找用户失败", "error", err)
		return errors.New("用户不存在")
	}

	// 更新岗位信息
	user.PositionID = positionID
	user.UpdatedBy = updatedBy

	if err := s.repo.Update(user); err != nil {
		s.log.Error("更新用户岗位失败", "error", err)
		return err
	}

	s.log.Info("用户岗位更新成功", "userID", userID, "positionID", positionID, "updatedBy", updatedBy)
	return nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(userID int64, status int, updatedBy int64) error {
	// 获取用户
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("查找用户失败", "error", err)
		return errors.New("用户不存在")
	}

	// 更新状态信息
	user.Status = status
	user.UpdatedBy = updatedBy

	if err := s.repo.Update(user); err != nil {
		s.log.Error("更新用户状态失败", "error", err)
		return err
	}

	s.log.Info("用户状态更新成功", "userID", userID, "status", status, "updatedBy", updatedBy)
	return nil
}

// ChangePassword 更改密码
func (s *UserService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	// 获取用户
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("获取用户失败", "error", err)
		return err
	}

	// 验证旧密码
	if verifyErr := encryption.VerifyPassword(user.PasswordHash, oldPassword); verifyErr != nil {
		s.log.Error("旧密码验证失败", "userID", userID)
		return errors.New("旧密码错误")
	}

	// 生成新密码哈希
	passwordHash, err := encryption.GeneratePasswordHash(newPassword)
	if err != nil {
		s.log.Error("生成密码哈希失败", "error", err)
		return err
	}

	// 更新密码
	user.PasswordHash = passwordHash
	user.UpdatedBy = userID
	if err := s.repo.Update(user); err != nil {
		s.log.Error("更新密码失败", "error", err)
		return err
	}

	s.log.Info("密码修改成功", "userID", userID)
	return nil
}

// RefreshToken 刷新令牌
// 参数:
// - refreshToken: 刷新令牌
// 返回:
// - string: 新的访问令牌
// - string: 新的刷新令牌
// - time.Time: 访问令牌过期时间
// - time.Time: 刷新令牌过期时间
// - error: 错误信息
func (s *UserService) RefreshToken(refreshToken string) (string, string, time.Time, time.Time, error) {
	// 调用JWT认证服务刷新令牌
	return s.jwtAuth.RefreshToken(refreshToken)
}

// Logout 退出登录
// 将用户的令牌添加到黑名单中，实现强制登出
func (s *UserService) Logout(userID string) error {
	// 由于 JWT 是无状态的，我们可以使用缓存来维护一个黑名单
	// 这里简化处理，记录用户的退出时间，后续可以在认证中间件中检查
	blacklistKey := fmt.Sprintf("blacklist:user:%s", userID)

	// 将用户添加到黑名单，有效期为 1 小时（与 JWT 令牌的过期时间保持一致）
	// 这样即使令牌还没有过期，也会被拒绝
	err := s.cache.Set(blacklistKey, time.Now().Unix(), time.Hour)
	if err != nil {
		s.log.Error("添加用户到黑名单失败", "userID", userID, "error", err)
		return err
	}

	s.log.Info("用户退出登录成功", "userID", userID)
	return nil
}

// GenerateToken 生成访问令牌和刷新令牌
func (s *UserService) GenerateToken(userID string, username string, nickname string, role string, loginType string) (string, string, time.Time, time.Time, error) {
	return s.jwtAuth.GenerateToken(userID, username, nickname, role, loginType)
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList(page, pageSize int, filters map[string]interface{}) ([]*entity.User, int64, error) {
	users, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取用户列表失败", "error", err)
		return nil, 0, err
	}
	s.log.Info("用户列表获取成功", "count", len(users), "total", total)
	return users, total, nil
}

// AddUser 增加用户（管理员功能）
func (s *UserService) AddUser(username, password, email, nickname, phone, avatar string, departmentID, positionID int64, createdBy string, status int) (*entity.User, error) {
	// 检查用户是否已存在
	_, err := s.repo.GetByUsername(username)
	if err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 不再检查邮箱和手机号的唯一性，因为已经移除了唯一约束

	// 生成密码哈希
	passwordHash, err := encryption.GeneratePasswordHash(password)
	if err != nil {
		s.log.Error("生成密码哈希失败", "error", err)
		return nil, err
	}

	// 创建新用户
	user := &entity.User{
		Username:     username,
		PasswordHash: passwordHash,
		Nickname:     nickname,
		Email:        email,
		Phone:        phone,
		Avatar:       avatar,
		DepartmentID: departmentID,
		PositionID:   positionID,
		Status:       status,
	}

	// 保存用户
	if err := s.repo.Create(user); err != nil {
		s.log.Error("创建用户失败", "error", err)
		return nil, err
	}

	// 为用户分配默认角色（user）
	if err := s.assignDefaultRole(user.ID); err != nil {
		s.log.Warn("为用户分配默认角色失败", "userID", user.ID, "error", err)
		// 不返回错误，因为用户创建已经成功
	}

	s.log.Info("管理员创建用户成功", "username", username, "createdBy", createdBy)
	return user, nil
}

// assignDefaultRole 为用户分配默认角色
func (s *UserService) assignDefaultRole(userID int64) error {
	// 获取所有角色
	roles, err := s.roleService.GetAllRoles()
	if err != nil {
		return err
	}

	// 查找 user 角色
	var userRole *entity.Role
	for _, role := range roles {
		if role.Code == "user" {
			userRole = role
			break
		}
	}

	// 如果存在 user 角色，为用户分配
	if userRole != nil {
		if err := s.roleService.AssignUserToRole(userRole.ID, userID); err != nil {
			return err
		}
		s.log.Info("为用户分配默认角色成功", "userID", userID, "roleID", userRole.ID, "roleName", userRole.Name)

		// 【新增】加载角色权限到 Redis 缓存
		if err := s.rolePermissionService.LoadRolePermissionsToRedis(fmt.Sprintf("%d", userRole.ID)); err != nil {
			s.log.Error("加载默认角色权限缓存失败", "error", err, "roleID", userRole.ID)
		}
	}

	return nil
}

// DeleteUser 删除用户（管理员功能）
func (s *UserService) DeleteUser(userID int64) error {
	// 获取用户
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("查找用户失败", "error", err)
		return errors.New("用户不存在")
	}

	// 删除用户
	if err := s.repo.Delete(user.ID); err != nil {
		s.log.Error("删除用户失败", "error", err)
		return err
	}

	s.log.Info("用户删除成功", "userID", userID)
	return nil
}

// ResetPassword 重置密码（管理员功能）
func (s *UserService) ResetPassword(userID int64, newPassword string, updatedBy int64) error {
	// 检查用户是否存在
	_, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("查找用户失败", "error", err)
		return errors.New("用户不存在")
	}

	// 生成新密码哈希
	passwordHash, err := encryption.GeneratePasswordHash(newPassword)
	if err != nil {
		s.log.Error("生成密码哈希失败", "error", err)
		return err
	}

	// 只更新密码和更新时间字段，避免更新 wechat_open_id 字段
	updates := map[string]interface{}{
		"password_hash": passwordHash,
		"updated_at":    time.Now(),
		"updated_by":    updatedBy,
	}

	if err := s.repo.UpdateFields(userID, updates); err != nil {
		s.log.Error("更新密码失败", "error", err)
		return err
	}

	s.log.Info("管理员重置密码成功", "userID", userID, "updatedBy", updatedBy)
	return nil
}

// GetSelfSetting 获取用户设置
func (s *UserService) GetSelfSetting(userID int64) (*entity.UserSetting, error) {

	// 尝试从数据库获取设置
	setting, err := s.settingRepo.GetByUserID(userID)
	if err == nil {
		// 找到设置，直接返回
		return setting, nil
	}

	// 如果是记录不存在的错误，创建默认设置
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建默认设置
		defaultSetting := &entity.UserSetting{
			UserID:           userID,
			ThemeColor:       "#1890ff",
			SidebarColor:     "#001529",
			NavbarColor:      "#fff",
			FontSize:         14,
			Language:         "zh-CN",
			AutoLogin:        false,
			RememberPassword: false,
		}

		// 保存默认设置
		if createErr := s.settingRepo.Create(defaultSetting); createErr != nil {
			s.log.Error("创建用户默认设置失败", "userID", userID, "error", createErr)
			return nil, createErr
		}

		s.log.Info("用户默认设置创建成功", "userID", userID)
		return defaultSetting, nil
	}

	// 其他错误
	s.log.Error("获取用户设置失败", "userID", userID, "error", err)
	return nil, err
}

// SetSelfSetting 保存用户设置
func (s *UserService) SetSelfSetting(userID int64, setting *entity.UserSetting) (*entity.UserSetting, error) {

	// 验证用户 ID 匹配
	if setting.UserID != 0 && setting.UserID != userID {
		return nil, errors.New("用户 ID 不匹配")
	}

	// 尝试从数据库获取现有设置
	existingSetting, err := s.settingRepo.GetByUserID(userID)
	if err != nil {
		// 如果是记录不存在的错误，创建新设置
		if errors.Is(err, gorm.ErrRecordNotFound) {
			setting.UserID = userID

			if createErr := s.settingRepo.Create(setting); createErr != nil {
				s.log.Error("创建用户设置失败", "userID", userID, "error", createErr)
				return nil, createErr
			}

			s.log.Info("用户设置创建成功", "userID", userID)
			return setting, nil
		}

		// 其他错误
		s.log.Error("获取用户设置失败", "userID", userID, "error", err)
		return nil, err
	}

	// 更新现有设置
	existingSetting.ThemeColor = setting.ThemeColor
	existingSetting.SidebarColor = setting.SidebarColor
	existingSetting.NavbarColor = setting.NavbarColor
	existingSetting.FontSize = setting.FontSize
	existingSetting.Language = setting.Language
	existingSetting.AutoLogin = setting.AutoLogin
	existingSetting.RememberPassword = setting.RememberPassword

	if err := s.settingRepo.Update(existingSetting); err != nil {
		s.log.Error("更新用户设置失败", "userID", userID, "error", err)
		return nil, err
	}

	s.log.Info("用户设置更新成功", "userID", userID)
	return existingSetting, nil
}

// BatchDeleteUsers 批量删除用户（管理员功能）
func (s *UserService) BatchDeleteUsers(userIDs []int64) error {
	// 批量获取用户信息
	for _, userID := range userIDs {
		// 检查用户是否存在
		_, err := s.repo.GetByID(userID)
		if err != nil {
			s.log.Error("查找用户失败", "error", err)
			return fmt.Errorf("用户 %d 不存在", userID)
		}
	}

	// 批量删除用户
	for _, userID := range userIDs {
		if err := s.repo.Delete(userID); err != nil {
			s.log.Error("删除用户失败", "userID", userID, "error", err)
			return err
		}
	}

	s.log.Info("批量删除用户成功", "userIDs", userIDs)
	return nil
}

// SetUserSpecialPermissions 设置用户特殊 API 权限
func (s *UserService) SetUserSpecialPermissions(userID int64, apiIDs []int64) error {
	if err := s.repo.SetUserSpecialPermissions(userID, apiIDs); err != nil {
		s.log.Error("设置用户特殊 API 权限失败", "userID", userID, "error", err)
		return err
	}
	s.log.Info("设置用户特殊 API 权限成功", "userID", userID, "apiCount", len(apiIDs))
	return nil
}

// SetUserSpecialBtnPermissions 设置用户特殊按钮权限
func (s *UserService) SetUserSpecialBtnPermissions(userID int64, btnPermIDs []int64) error {
	if err := s.repo.SetUserSpecialBtnPermissions(userID, btnPermIDs); err != nil {
		s.log.Error("设置用户特殊按钮权限失败", "userID", userID, "error", err)
		return err
	}
	s.log.Info("设置用户特殊按钮权限成功", "userID", userID, "btnPermCount", len(btnPermIDs))
	return nil
}

// GetUserSpecialPermissions 获取用户特殊 API 权限
func (s *UserService) GetUserSpecialPermissions(userID int64) ([]*entity.API, error) {
	apis, err := s.repo.GetUserSpecialPermissions(userID)
	if err != nil {
		s.log.Error("获取用户特殊API权限失败", "userID", userID, "error", err)
		return nil, err
	}
	return apis, nil
}

// ExportUsers 导出用户
func (s *UserService) ExportUsers(filters map[string]interface{}) ([]*entity.User, error) {
	// 不分页获取所有符合条件的用户
	users, _, err := s.repo.List(1, 10000, filters) // 假设最多10000个用户
	if err != nil {
		s.log.Error("导出用户失败", "error", err)
		return nil, err
	}
	return users, nil
}

// ImportUsers 导入用户
func (s *UserService) ImportUsers(users []*entity.User, createdBy string) (int, error) {
	count := 0
	for _, user := range users {
		// 检查用户是否已存在
		_, err := s.repo.GetByUsername(user.Username)
		if err == nil {
			continue // 用户已存在，跳过
		}

		// 检查邮箱是否已存在
		_, err = s.repo.GetByEmail(user.Email)
		if err == nil {
			continue // 邮箱已存在，跳过
		}

		// 检查手机号是否已存在
		_, err = s.repo.GetByPhone(user.Phone)
		if err == nil {
			continue // 手机号已存在，跳过
		}

		// 创建用户
		if err := s.repo.Create(user); err != nil {
			s.log.Error("导入创建用户失败", "username", user.Username, "error", err)
			continue
		}

		count++
	}

	s.log.Info("用户导入完成", "total", len(users), "success", count, "createdBy", createdBy)
	return count, nil
}

// GetUserPermissions 获取用户所有 API 权限
// 包括：
// 1. 用户关联的所有角色的 API 权限（包含权限继承）
// 2. 用户特殊的 API 权限
// 返回去重后的 API 权限列表
func (s *UserService) GetUserPermissions(userID int64) ([]*entity.API, error) {
	// 1. 获取用户信息
	_, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("获取用户信息失败", "userID", userID, "error", err)
		return nil, errors.New("用户不存在")
	}

	// 使用 map 存储 API 权限，键为 API 权限 ID，用于去重
	apiMap := make(map[int64]*entity.API)

	// 2. 获取用户关联的所有角色的 API 权限
	roles, err := s.roleService.GetRolesForUser(userID)
	if err != nil {
		s.log.Error("获取用户角色失败", "userID", userID, "error", err)
		return nil, err
	}

	// 遍历所有角色，获取 API 权限
	for _, role := range roles {
		// 检查角色状态，只获取启用的角色权限
		if role.Status != 1 {
			continue
		}

		// 获取角色及其所有父角色的 API 权限（实现权限继承）
		roleAPIs, err := s.roleService.GetAllInheritedPermissions(role.ID)
		if err != nil {
			s.log.Error("获取角色 API 权限失败", "roleID", role.ID, "error", err)
			continue
		}

		// 将角色 API 权限添加到 map 中去重
		for _, api := range roleAPIs {
			apiMap[api.ID] = api
		}
	}

	// 3. 获取用户特殊的 API 权限
	userSpecialAPIs, err := s.GetUserSpecialPermissions(userID)
	if err != nil {
		s.log.Error("获取用户特殊API权限失败", "userID", userID, "error", err)
		// 继续执行，不返回错误
	} else {
		// 将用户特殊API权限添加到map中去重
		for _, api := range userSpecialAPIs {
			apiMap[api.ID] = api
		}
	}

	// 4. 将map转换为切片
	apis := make([]*entity.API, 0, len(apiMap))
	for _, api := range apiMap {
		apis = append(apis, api)
	}

	s.log.Info("获取用户API权限成功", "userID", userID, "count", len(apis))
	return apis, nil
}

// GetUserBtnPermissions 获取用户所有按钮权限
// 包括：
// 1. 用户关联的所有角色的按钮权限（包含权限继承）
// 2. 用户特殊的按钮权限
// 返回去重后的按钮权限列表
func (s *UserService) GetUserBtnPermissions(userID int64) ([]*entity.BtnPerm, error) {
	// 1. 获取用户信息
	_, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("获取用户信息失败", "userID", userID, "error", err)
		return nil, errors.New("用户不存在")
	}

	// 使用 map 存储按钮权限，键为按钮权限 ID，用于去重
	btnPermMap := make(map[int64]*entity.BtnPerm)

	// 2. 获取用户关联的所有角色的按钮权限
	roles, err := s.roleService.GetRolesForUser(userID)
	if err != nil {
		s.log.Error("获取用户角色失败", "userID", userID, "error", err)
		return nil, err
	}

	// 遍历所有角色，获取按钮权限
	for _, role := range roles {
		// 检查角色状态，只获取启用的角色权限
		if role.Status != 1 {
			continue
		}

		// 获取角色及其所有父角色的按钮权限（实现权限继承）
		roleBtnPerms, err := s.roleService.GetAllInheritedBtnPerms(role.ID)
		if err != nil {
			s.log.Error("获取角色按钮权限失败", "roleID", role.ID, "error", err)
			continue
		}

		// 将角色按钮权限添加到 map 中去重
		for _, btnPerm := range roleBtnPerms {
			// 检查按钮权限状态，只获取启用的权限
			if btnPerm.Status == 1 {
				btnPermMap[btnPerm.ID] = btnPerm
			}
		}
	}

	// 3. 获取用户特殊的按钮权限
	userSpecialBtnPerms, err := s.GetUserSpecialBtnPermissions(userID)
	if err != nil {
		s.log.Error("获取用户特殊按钮权限失败", "userID", userID, "error", err)
		// 继续执行，不返回错误
	} else {
		// 将用户特殊按钮权限添加到map中去重
		for _, btnPerm := range userSpecialBtnPerms {
			// 检查按钮权限状态，只获取启用的权限
			if btnPerm.Status == 1 {
				btnPermMap[btnPerm.ID] = btnPerm
			}
		}
	}

	// 4. 将map转换为切片
	btnPerms := make([]*entity.BtnPerm, 0, len(btnPermMap))
	for _, btnPerm := range btnPermMap {
		btnPerms = append(btnPerms, btnPerm)
	}

	s.log.Info("获取用户按钮权限成功", "userID", userID, "count", len(btnPerms))
	return btnPerms, nil
}

// GetUserSpecialBtnPermissions 获取用户特殊按钮权限
func (s *UserService) GetUserSpecialBtnPermissions(userID int64) ([]*entity.BtnPerm, error) {
	btnPerms, err := s.repo.GetUserSpecialBtnPermissions(userID)
	if err != nil {
		s.log.Error("获取用户特殊按钮权限失败", "userID", userID, "error", err)
		return nil, err
	}
	return btnPerms, nil
}

// UpdateUserRoles 更新用户角色
func (s *UserService) UpdateUserRoles(userID int64, roleIDs []int64) error {
	// 检查用户是否存在
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("查找用户失败", "error", err, "userID", userID)
		return errors.New("用户不存在")
	}

	s.log.Info("========== 开始更新用户角色 ==========", "userID", userID, "username", user.Username, "roleIDs", roleIDs)

	// 获取用户当前的角色
	currentRoles, err := s.roleService.GetRolesForUser(userID)
	if err != nil {
		s.log.Error("获取用户当前角色失败", "error", err, "userID", userID)
		return err
	}

	s.log.Info("用户当前角色", "userID", userID, "currentRoles", len(currentRoles))

	// 打印当前所有角色的详细信息
	for _, role := range currentRoles {
		s.log.Info("当前角色详情", "userID", userID, "roleID", role.ID, "roleCode", role.Code, "roleName", role.Name)
	}

	// 检查用户当前是否有管理员角色
	hasAdminRole := false
	for _, role := range currentRoles {
		if role.Code == "admin" {
			hasAdminRole = true
			s.log.Info("用户当前拥有管理员角色", "userID", userID)
			break
		}
	}

	// 检查新角色列表中是否包含管理员角色
	willHaveAdminRole := false
	for _, roleID := range roleIDs {
		role, err := s.roleService.GetRoleByID(roleID)
		if err != nil {
			s.log.Error("角色不存在", "roleID", roleID, "userID", userID)
			return errors.New("角色不存在")
		}
		s.log.Info("新角色详情", "userID", userID, "roleID", roleID, "roleCode", role.Code, "roleName", role.Name)
		if role.Code == "admin" {
			willHaveAdminRole = true
			s.log.Info("新角色列表包含管理员角色", "userID", userID)
			break
		}
	}

	s.log.Info("角色检查完成", "userID", userID, "hasAdminRole", hasAdminRole, "willHaveAdminRole", willHaveAdminRole)

	// 保护管理员角色：如果用户当前有管理员角色，但新角色列表中没有，则拒绝
	if hasAdminRole && !willHaveAdminRole {
		s.log.Error("❌ 拒绝操作：不能删除用户的管理员角色", "userID", userID, "username", user.Username)
		return errors.New("不能删除用户的管理员角色")
	}

	s.log.Info("✅ 通过管理员角色保护检查", "userID", userID)

	// 清除用户现有的角色关联
	for _, role := range currentRoles {
		err := s.roleService.RevokeUserFromRole(role.ID, userID)
		if err != nil {
			s.log.Error("移除用户角色失败", "error", err, "userID", userID, "roleID", role.ID)
			return err
		}
	}

	s.log.Info("已清除用户现有角色", "userID", userID, "clearedCount", len(currentRoles))

	// 为用户分配新的角色
	for _, roleID := range roleIDs {
		// 检查角色是否存在
		role, err := s.roleService.GetRoleByID(roleID)
		if err != nil {
			s.log.Error("角色不存在", "roleID", roleID, "userID", userID)
			return errors.New("角色不存在")
		}

		err = s.roleService.AssignUserToRole(roleID, userID)
		if err != nil {
			s.log.Error("分配用户角色失败", "error", err, "userID", userID, "roleID", roleID, "roleName", role.Name)
			return err
		}
	}

	s.log.Info("用户角色分配完成", "userID", userID, "assignedCount", len(roleIDs))

	// 清除用户所有相关角色的权限缓存
	// 清除旧角色的缓存
	for _, role := range currentRoles {
		if err := s.rolePermissionService.ClearRolePermissionsCache(role.ID); err != nil {
			s.log.Error("清除旧角色权限缓存失败", "error", err, "roleID", role.ID)
		}
	}

	// 清除新角色的缓存
	for _, roleID := range roleIDs {
		if err := s.rolePermissionService.ClearRolePermissionsCache(roleID); err != nil {
			s.log.Error("清除新角色权限缓存失败", "error", err, "roleID", roleID)
		}
	}

	s.log.Info("========== 更新用户角色完成 ==========", "userID", userID, "roleCount", len(roleIDs))
	return nil
}

// SwitchRole 切换用户当前角色
// 该方法仅改变用户的当前活动角色，不修改用户的角色关联关系
func (s *UserService) SwitchRole(userID int64, roleID int64) error {
	s.log.Info("========== 开始切换用户角色 ==========", "userID", userID, "targetRoleID", roleID)

	// 检查用户是否存在
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("查找用户失败", "error", err, "userID", userID)
		return errors.New("用户不存在")
	}

	// 检查角色是否存在
	role, err := s.roleService.GetRoleByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "roleID", roleID, "userID", userID)
		return errors.New("角色不存在")
	}

	// 检查用户是否拥有该角色
	userRoles, err := s.roleService.GetRolesForUser(userID)
	if err != nil {
		s.log.Error("获取用户角色失败", "error", err, "userID", userID)
		return errors.New("获取用户角色失败")
	}

	hasRole := false
	for _, userRole := range userRoles {
		if userRole.ID == roleID {
			hasRole = true
			break
		}
	}

	if !hasRole {
		s.log.Error("用户不拥有该角色", "userID", userID, "roleID", roleID)
		return errors.New("用户不拥有该角色")
	}

	// 检查角色状态
	if role.Status != 1 {
		s.log.Error("角色已禁用", "roleID", roleID, "userID", userID)
		return errors.New("角色已禁用")
	}

	// 【步骤 1】缓存用户当前角色选择到 Redis
	// key: "user:current_role:{userID}", value: roleID (字符串格式)
	// 使用 string 格式，与认证中间件保持一致
	currentRoleKey := fmt.Sprintf("user:current_role:%s", fmt.Sprintf("%d", userID))
	currentRoleValue := fmt.Sprintf("%d", roleID)
	if err := s.cache.Set(currentRoleKey, currentRoleValue, 24*time.Hour); err != nil {
		s.log.Error("❌ 缓存用户当前角色失败", "error", err, "userID", userID, "roleID", roleID, "cacheKey", currentRoleKey)
		// 不返回错误，因为角色切换逻辑已经成功
	} else {
		s.log.Info("✅ 已缓存用户当前角色", "userID", userID, "roleID", roleID, "cacheKey", currentRoleKey, "expire", "24h")
	}

	// 【步骤 2】缓存角色详情到 Redis，供认证中间件使用
	// key: "role:info:{roleID}", value: 简化的角色 JSON（只包含必要字段）
	// 使用 string 格式，与认证中间件保持一致
	roleCacheKey := fmt.Sprintf("role:info:%s", fmt.Sprintf("%d", roleID))
	simpleRole := map[string]interface{}{
		"ID":     fmt.Sprintf("%d", role.ID),
		"Code":   role.Code,
		"Name":   role.Name,
		"Status": role.Status,
	}

	roleData, marshalErr := json.Marshal(simpleRole)
	if marshalErr != nil {
		s.log.Error("❌ 序列化角色信息失败", "error", marshalErr, "roleID", roleID)
	} else {
		if cacheErr := s.cache.Set(roleCacheKey, string(roleData), 24*time.Hour); cacheErr != nil {
			s.log.Error("❌ 缓存角色详情失败", "error", cacheErr, "roleID", roleID, "cacheKey", roleCacheKey)
		} else {
			s.log.Info("✅ 已缓存角色详情", "roleID", roleID, "roleCode", role.Code, "roleName", role.Name, "cacheKey", roleCacheKey, "expire", "24h")
		}
	}

	s.log.Info("========== 用户角色切换完成 ==========",
		"userID", userID,
		"username", user.Username,
		"newRoleID", roleID,
		"newRoleCode", role.Code,
		"newRoleName", role.Name)
	return nil
}

// SetCurrentRoleID 设置用户的当前角色 ID
// 将用户的当前角色选择缓存到 Redis
// 参数:
// - userID: 用户 ID
// - roleID: 角色 ID
// 返回:
// - error: 错误信息
func (s *UserService) SetCurrentRoleID(userID int64, roleID int64) error {
	currentRoleKey := fmt.Sprintf("user:current_role:%d", userID)

	// 将角色 ID 缓存到 Redis，设置 24 小时过期时间
	return s.cache.Set(currentRoleKey, fmt.Sprintf("%d", roleID), 24*time.Hour)
}

// GetCurrentRoleID 获取用户的当前角色 ID
// 从缓存中读取用户的当前角色选择
func (s *UserService) GetCurrentRoleID(userID int64) (string, error) {
	currentRoleKey := fmt.Sprintf("user:current_role:%d", userID)

	// 从缓存获取当前角色
	roleID, err := s.cache.Get(currentRoleKey)
	if err != nil {
		return "", nil
	}

	// 如果缓存中没有存储，返回空字符串
	if roleID == "" {
		return "", nil
	}

	return roleID, nil
}

// GetRoleByID 根据 ID 获取角色信息
// 用于 API Handler 获取用户角色详情
func (s *UserService) GetRoleByID(id int64) (*entity.Role, error) {
	return s.roleService.GetRoleByID(id)
}

// parseUserAgent 解析 User-Agent 字符串，获取浏览器和操作系统信息
func parseUserAgent(userAgent string) (browser string, os string) {
	// 默认值
	browser = "Unknown"
	os = "Unknown"

	// 解析操作系统
	if strings.Contains(userAgent, "Windows NT 10.0") {
		os = "Windows 10"
	} else if strings.Contains(userAgent, "Windows NT 6.3") {
		os = "Windows 8.1"
	} else if strings.Contains(userAgent, "Windows NT 6.2") {
		os = "Windows 8"
	} else if strings.Contains(userAgent, "Windows NT 6.1") {
		os = "Windows 7"
	} else if strings.Contains(userAgent, "Mac OS X") {
		// 提取 macOS 版本
		re := regexp.MustCompile(`Mac OS X ([0-9_\.]+)`)
		if matches := re.FindStringSubmatch(userAgent); len(matches) > 1 {
			os = "macOS " + strings.ReplaceAll(matches[1], "_", ".")
		} else {
			os = "macOS"
		}
	} else if strings.Contains(userAgent, "X11") {
		os = "Linux"
	} else if strings.Contains(userAgent, "Linux") {
		os = "Linux"
	} else if strings.Contains(userAgent, "Android") {
		os = "Android"
	} else if strings.Contains(userAgent, "iOS") {
		os = "iOS"
	} else if strings.Contains(userAgent, "iPhone") {
		os = "iOS (iPhone)"
	} else if strings.Contains(userAgent, "iPad") {
		os = "iOS (iPad)"
	}

	// 解析浏览器
	if strings.Contains(userAgent, "Edg/") {
		// Microsoft Edge (必须在 Chrome 之前检查)
		re := regexp.MustCompile(`Edg/([0-9.]+)`)
		if matches := re.FindStringSubmatch(userAgent); len(matches) > 1 {
			browser = "Microsoft Edge " + matches[1]
		} else {
			browser = "Microsoft Edge"
		}
	} else if strings.Contains(userAgent, "Chrome/") {
		// Google Chrome
		re := regexp.MustCompile(`Chrome/([0-9.]+)`)
		if matches := re.FindStringSubmatch(userAgent); len(matches) > 1 {
			browser = "Chrome " + matches[1]
		} else {
			browser = "Chrome"
		}
	} else if strings.Contains(userAgent, "Firefox/") {
		// Mozilla Firefox
		re := regexp.MustCompile(`Firefox/([0-9.]+)`)
		if matches := re.FindStringSubmatch(userAgent); len(matches) > 1 {
			browser = "Firefox " + matches[1]
		} else {
			browser = "Firefox"
		}
	} else if strings.Contains(userAgent, "Safari/") && !strings.Contains(userAgent, "Chrome") {
		// Safari (必须在 Chrome 之后检查，因为 Chrome 也包含 Safari)
		re := regexp.MustCompile(`Version/([0-9.]+)`)
		if matches := re.FindStringSubmatch(userAgent); len(matches) > 1 {
			browser = "Safari " + matches[1]
		} else {
			browser = "Safari"
		}
	} else if strings.Contains(userAgent, "MSIE") || strings.Contains(userAgent, "Trident/") {
		// Internet Explorer
		re := regexp.MustCompile(`(MSIE |rv:)([0-9.]+)`)
		if matches := re.FindStringSubmatch(userAgent); len(matches) > 2 {
			browser = "Internet Explorer " + matches[2]
		} else {
			browser = "Internet Explorer"
		}
	} else if strings.Contains(userAgent, "Mozilla/") {
		browser = "Mozilla"
	}

	return browser, os
}

// getLoginPlaceByIP 根据 IP 地址获取登录地点（简化版本）
// 实际生产环境可接入 IP 定位 API（如高德、百度、IP2Region 等）
func getLoginPlaceByIP(ip string) string {
	// 本地回环地址
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return "本地连接"
	}

	// 内网 IP 范围
	if isPrivateIP(ip) {
		return "内网"
	}

	// TODO: 可以接入 IP 定位服务
	// 示例：使用 IP2Region、高德地图 API、百度地图 API 等
	// 这里返回一个默认值
	return "中国"
}

// isPrivateIP 判断是否为内网 IP
func isPrivateIP(ip string) bool {
	// 简单的内网 IP 判断
	// 10.0.0.0/8
	if strings.HasPrefix(ip, "10.") {
		return true
	}
	// 172.16.0.0/12
	if strings.HasPrefix(ip, "172.") {
		parts := strings.Split(ip, ".")
		if len(parts) >= 2 {
			if second, err := strconv.Atoi(parts[1]); err == nil {
				if second >= 16 && second <= 31 {
					return true
				}
			}
		}
	}
	// 192.168.0.0/16
	if strings.HasPrefix(ip, "192.168.") {
		return true
	}
	return false
}
