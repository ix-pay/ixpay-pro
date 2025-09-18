package service

import (
	"errors"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/redis"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/snowflake"
	"github.com/ix-pay/ixpay-pro/internal/utils/captcha"
	"github.com/ix-pay/ixpay-pro/internal/utils/encryption"
)

// UserService 实现用户领域服务接口
type UserService struct {
	repo      model.UserRepository
	jwtAuth   *auth.JWTAuth
	config    *config.Config
	log       logger.Logger
	redis     *redis.RedisClient
	snowflake *snowflake.Snowflake
}

// NewUserService 创建用户服务实例
func NewUserService(repo model.UserRepository, jwtAuth *auth.JWTAuth, config *config.Config, log logger.Logger, redis *redis.RedisClient, snowflake *snowflake.Snowflake) model.UserService {
	return &UserService{
		repo:      repo,
		jwtAuth:   jwtAuth,
		config:    config,
		log:       log,
		redis:     redis,
		snowflake: snowflake,
	}
}

// Captcha 生成验证码
func (s *UserService) Captcha() (string, string, int, bool, error) {
	// 验证码长度
	len := 4

	// 设置验证码有效期（秒）
	expiry := s.config.Server.CaptchaTimeOut // 60秒

	// 开启验证码
	openCaptcha := s.config.Server.OpenCaptcha

	id, base64Img, err := captcha.NewCaptcha(len, expiry, s.redis)
	if err != nil {
		s.log.Error("Failed to generate captcha", "error", err)
		return id, "", len, openCaptcha, err
	}

	return id, base64Img, len, openCaptcha, nil
}

// Register 用户注册
func (s *UserService) Register(username, password, email string) (*model.User, error) {
	// 检查用户是否已存在
	_, err := s.repo.GetByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	_, err = s.repo.GetByEmail(email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// 生成密码哈希
	passwordHash, err := encryption.GeneratePasswordHash(password)
	if err != nil {
		s.log.Error("Failed to generate password hash", "error", err)
		return nil, err
	}

	// 创建新用户
	user := &model.User{
		Username:     username,
		PasswordHash: passwordHash,
		Email:        email,
		Role:         "user",
		Status:       1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存用户
	if err := s.repo.Create(user); err != nil {
		s.log.Error("Failed to create user", "error", err)
		return nil, err
	}

	s.log.Info("User registered successfully", "username", username)
	return user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (*model.User, string, string, error) {
	// 根据用户名获取用户
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		s.log.Error("Failed to find user", "error", err)
		return nil, "", "", errors.New("invalid username or password")
	}

	// 验证密码
	if err := encryption.VerifyPassword(user.PasswordHash, password); err != nil {
		s.log.Error("Password verification failed", "username", username)
		return nil, "", "", errors.New("invalid username or password")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, "", "", errors.New("user account is inactive")
	}

	// 生成令牌
	accessToken, refreshToken, err := s.jwtAuth.GenerateToken(user.ID, user.Username, user.Role, "password")
	if err != nil {
		s.log.Error("Failed to generate tokens", "error", err)
		return nil, "", "", err
	}

	s.log.Info("User logged in successfully", "username", username)
	return user, accessToken, refreshToken, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userID uint) (*model.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("Failed to get user info", "error", err)
		return nil, err
	}
	return user, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(user *model.User) error {
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(user); err != nil {
		s.log.Error("Failed to update user info", "error", err)
		return err
	}
	s.log.Info("User info updated successfully", "userID", user.ID)
	return nil
}

// ChangePassword 更改密码
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 获取用户
	user, err := s.repo.GetByID(userID)
	if err != nil {
		s.log.Error("Failed to get user", "error", err)
		return err
	}

	// 验证旧密码
	if err := encryption.VerifyPassword(user.PasswordHash, oldPassword); err != nil {
		s.log.Error("Old password verification failed", "userID", userID)
		return errors.New("invalid old password")
	}

	// 生成新密码哈希
	passwordHash, err := encryption.GeneratePasswordHash(newPassword)
	if err != nil {
		s.log.Error("Failed to generate password hash", "error", err)
		return err
	}

	// 更新密码
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(user); err != nil {
		s.log.Error("Failed to update password", "error", err)
		return err
	}

	s.log.Info("Password changed successfully", "userID", userID)
	return nil
}

// RefreshToken 刷新令牌
func (s *UserService) RefreshToken(refreshToken string) (string, string, error) {
	return s.jwtAuth.RefreshToken(refreshToken)
}
