package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// OnlineUserService 在线用户服务实现
type OnlineUserService struct {
	repo repo.OnlineUserRepository
	log  logger.Logger
}

// NewOnlineUserService 创建在线用户服务实例
func NewOnlineUserService(repo repo.OnlineUserRepository, log logger.Logger) *OnlineUserService {
	return &OnlineUserService{
		repo: repo,
		log:  log,
	}
}

// AddOnlineUser 添加用户到在线列表
func (s *OnlineUserService) AddOnlineUser(
	userID string,
	username, nickname, sessionID, ip, place, device, browser, os, userAgent string,
) error {
	now := time.Now()

	user := &entity.OnlineUser{
		UserID:       userID,
		Username:     username,
		Nickname:     nickname,
		SessionID:    sessionID,
		LoginIP:      ip,
		LoginPlace:   place,
		LoginTime:    now,
		LastActiveAt: now,
		Device:       device,
		Browser:      browser,
		OS:           os,
		UserAgent:    userAgent,
	}

	if err := s.repo.Add(user); err != nil {
		s.log.Error("添加在线用户失败", "error", err, "user_id", userID, "username", username)
		return err
	}

	s.log.Info("添加在线用户成功", "user_id", userID, "username", username, "session_id", sessionID)
	return nil
}

// UpdateUserActive 更新用户活跃状态
func (s *OnlineUserService) UpdateUserActive(userID string) error {
	if err := s.repo.UpdateActiveTime(userID); err != nil {
		s.log.Error("更新用户活跃状态失败", "error", err, "user_id", userID)
		return err
	}
	return nil
}

// RemoveOnlineUser 移除在线用户（登出）
func (s *OnlineUserService) RemoveOnlineUser(userID string) error {
	if err := s.repo.Remove(userID); err != nil {
		s.log.Error("移除在线用户失败", "error", err, "user_id", userID)
		return err
	}

	s.log.Info("移除在线用户成功", "user_id", userID)
	return nil
}

// RemoveOnlineUserBySessionID 根据会话 ID 移除在线用户
func (s *OnlineUserService) RemoveOnlineUserBySessionID(sessionID string) error {
	if err := s.repo.RemoveBySessionID(sessionID); err != nil {
		s.log.Error("根据会话 ID 移除在线用户失败", "error", err, "session_id", sessionID)
		return err
	}

	s.log.Info("根据会话 ID 移除在线用户成功", "session_id", sessionID)
	return nil
}

// GetOnlineUserList 获取在线用户列表
func (s *OnlineUserService) GetOnlineUserList() ([]*entity.OnlineUser, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		s.log.Error("获取在线用户列表失败", "error", err)
		return nil, err
	}

	s.log.Info("获取在线用户列表成功", "count", len(users))
	return users, nil
}

// GetOnlineUserByID 获取在线用户详情
func (s *OnlineUserService) GetOnlineUserByID(userID string) (*entity.OnlineUser, error) {
	user, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.Error("获取在线用户详情失败", "error", err, "user_id", userID)
		return nil, err
	}

	if user == nil {
		return nil, errors.New("用户不在线")
	}

	return user, nil
}

// GetOnlineUserBySessionID 根据会话 ID 获取在线用户
func (s *OnlineUserService) GetOnlineUserBySessionID(sessionID string) (*entity.OnlineUser, error) {
	user, err := s.repo.GetBySessionID(sessionID)
	if err != nil {
		s.log.Error("根据会话 ID 获取在线用户失败", "error", err, "session_id", sessionID)
		return nil, err
	}

	if user == nil {
		return nil, errors.New("会话不存在或已过期")
	}

	return user, nil
}

// ForceOffline 强制用户下线
// 该方法会移除用户的在线状态，但不会使 JWT token 失效
// 如果需要完全禁止用户访问，需要配合 JWT 黑名单机制
func (s *OnlineUserService) ForceOffline(userID string, operatorID string) error {
	// 检查用户是否在线
	online, err := s.repo.Exists(userID)
	if err != nil {
		s.log.Error("检查用户在线状态失败", "error", err, "user_id", userID)
		return err
	}

	if !online {
		return errors.New("用户不在线")
	}

	// 获取用户信息用于日志
	user, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.Error("获取用户信息失败", "error", err, "user_id", userID)
		return err
	}

	// 移除在线用户
	if err := s.repo.Remove(userID); err != nil {
		s.log.Error("强制用户下线失败", "error", err, "user_id", userID)
		return err
	}

	s.log.Info("强制用户下线成功",
		"user_id", userID,
		"username", user.Username,
		"operator_id", operatorID,
	)
	return nil
}

// GetOnlineCount 获取在线用户数量
func (s *OnlineUserService) GetOnlineCount() (int, error) {
	count, err := s.repo.GetCount()
	if err != nil {
		s.log.Error("获取在线用户数量失败", "error", err)
		return 0, err
	}

	return count, nil
}

// IsOnline 检查用户是否在线
func (s *OnlineUserService) IsOnline(userID string) (bool, error) {
	return s.repo.Exists(userID)
}

// KickoutUser 踢出用户（用于管理员强制下线）
// 与 ForceOffline 类似，但会记录更详细的日志
func (s *OnlineUserService) KickoutUser(userID string, reason string, operatorID string) error {
	// 检查用户是否在线
	online, err := s.repo.Exists(userID)
	if err != nil {
		s.log.Error("检查用户在线状态失败", "error", err, "user_id", userID)
		return err
	}

	if !online {
		return errors.New("用户不在线")
	}

	// 获取用户信息用于日志
	user, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.log.Error("获取用户信息失败", "error", err, "user_id", userID)
		return err
	}

	// 移除在线用户
	if err := s.repo.Remove(userID); err != nil {
		s.log.Error("踢出用户失败", "error", err, "user_id", userID)
		return err
	}

	s.log.Info("踢出用户成功",
		"user_id", userID,
		"username", user.Username,
		"reason", reason,
		"operator_id", operatorID,
	)
	return nil
}

// GetOnlineUsersByCondition 根据条件获取在线用户（预留方法）
// 可以用于实现更复杂的查询，如按部门、角色等筛选
func (s *OnlineUserService) GetOnlineUsersByCondition(condition map[string]interface{}) ([]*entity.OnlineUser, error) {
	// 目前实现返回所有在线用户
	// 后续可以根据条件进行过滤
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	// TODO: 根据条件过滤用户
	// 例如：按部门、角色、登录 IP 等

	return users, nil
}

// RefreshUserExpiration 刷新用户会话过期时间
// 当用户活跃时，可以调用此方法延长会话时间
func (s *OnlineUserService) RefreshUserExpiration(userID string) error {
	// 先更新活跃时间
	if err := s.UpdateUserActive(userID); err != nil {
		return err
	}

	s.log.Debug("刷新用户会话过期时间", "user_id", userID)
	return nil
}

// BatchKickoutUsers 批量踢出用户
func (s *OnlineUserService) BatchKickoutUsers(userIDs []string, reason string, operatorID string) error {
	successCount := 0
	failCount := 0

	for _, userID := range userIDs {
		if err := s.KickoutUser(userID, reason, operatorID); err != nil {
			s.log.Error("批量踢出用户失败", "user_id", userID, "error", err)
			failCount++
		} else {
			successCount++
		}
	}

	s.log.Info("批量踢出用户完成",
		"total", len(userIDs),
		"success", successCount,
		"fail", failCount,
		"reason", reason,
		"operator_id", operatorID,
	)

	if failCount > 0 {
		return fmt.Errorf("批量踢出用户完成，成功 %d 人，失败 %d 人", successCount, failCount)
	}

	return nil
}
