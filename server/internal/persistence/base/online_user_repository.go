package persistence

import (
	"encoding/json"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
)

const (
	onlineUserKeyPrefix = "online:user:"
	onlineUserSetKey    = "online:users:set"
	onlineUserExpire    = 30 * time.Minute // 在线用户过期时间
)

// onlineUserRepository Repository 实现（基于缓存）
type onlineUserRepository struct {
	cache cache.Cache
}

// 确保实现接口
var _ repo.OnlineUserRepository = (*onlineUserRepository)(nil)

// NewOnlineUserRepository 创建在线用户仓库实现
func NewOnlineUserRepository(cache cache.Cache) repo.OnlineUserRepository {
	return &onlineUserRepository{cache: cache}
}

// Add 添加在线用户
func (r *onlineUserRepository) Add(user *entity.OnlineUser) error {
	// 序列化用户数据
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// 存储用户数据
	userKey := onlineUserKeyPrefix + user.UserID
	if err := r.cache.Set(userKey, string(data), onlineUserExpire); err != nil {
		return err
	}

	// 添加到在线用户集合（使用缓存的 Set 方法模拟）
	// 由于缓存接口不支持集合操作，我们使用一个列表来模拟
	onlineUsersListKey := onlineUserSetKey + ":list"
	existingUsers, _ := r.cache.Get(onlineUsersListKey)
	var userIDs []string
	if existingUsers != "" {
		json.Unmarshal([]byte(existingUsers), &userIDs)
	}

	// 检查是否已存在，不存在则添加
	exists := false
	for _, id := range userIDs {
		if id == user.UserID {
			exists = true
			break
		}
	}
	if !exists {
		userIDs = append(userIDs, user.UserID)
		listData, _ := json.Marshal(userIDs)
		r.cache.Set(onlineUsersListKey, string(listData), onlineUserExpire)
	}

	return nil
}

// GetByUserID 根据用户 ID 获取在线用户
func (r *onlineUserRepository) GetByUserID(userID string) (*entity.OnlineUser, error) {
	userKey := onlineUserKeyPrefix + userID
	data, err := r.cache.Get(userKey)
	if err != nil {
		return nil, err
	}

	var user entity.OnlineUser
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetBySessionID 根据会话 ID 获取在线用户
func (r *onlineUserRepository) GetBySessionID(sessionID string) (*entity.OnlineUser, error) {
	// 需要通过扫描所有在线用户来查找匹配 sessionID 的用户
	// 这种方式效率较低，建议只在必要时使用
	onlineUsersListKey := onlineUserSetKey + ":list"
	existingUsers, err := r.cache.Get(onlineUsersListKey)
	if err != nil {
		return nil, err
	}

	var userIDs []string
	if existingUsers != "" {
		json.Unmarshal([]byte(existingUsers), &userIDs)
	}

	for _, userID := range userIDs {
		user, err := r.GetByUserID(userID)
		if err != nil {
			continue
		}
		if user != nil && user.SessionID == sessionID {
			return user, nil
		}
	}

	return nil, nil
}

// UpdateActiveTime 更新用户活跃时间
func (r *onlineUserRepository) UpdateActiveTime(userID string) error {
	// 刷新过期时间（通过重新设置缓存实现）
	user, err := r.GetByUserID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	user.LastActiveAt = time.Now()
	return r.Add(user)
}

// Remove 移除在线用户
func (r *onlineUserRepository) Remove(userID string) error {
	userKey := onlineUserKeyPrefix + userID

	// 删除用户数据
	if err := r.cache.Delete(userKey); err != nil {
		return err
	}

	// 从在线用户列表中移除
	onlineUsersListKey := onlineUserSetKey + ":list"
	existingUsers, err := r.cache.Get(onlineUsersListKey)
	if err != nil {
		return err
	}

	var userIDs []string
	if existingUsers != "" {
		json.Unmarshal([]byte(existingUsers), &userIDs)
	}

	// 移除该用户
	newUserIDs := make([]string, 0)
	for _, id := range userIDs {
		if id != userID {
			newUserIDs = append(newUserIDs, id)
		}
	}

	if len(newUserIDs) > 0 {
		listData, _ := json.Marshal(newUserIDs)
		r.cache.Set(onlineUsersListKey, string(listData), onlineUserExpire)
	} else {
		r.cache.Delete(onlineUsersListKey)
	}

	return nil
}

// RemoveBySessionID 根据会话 ID 移除在线用户
func (r *onlineUserRepository) RemoveBySessionID(sessionID string) error {
	user, err := r.GetBySessionID(sessionID)
	if err != nil {
		return err
	}
	if user != nil {
		return r.Remove(user.UserID)
	}
	return nil
}

// GetAll 获取所有在线用户
func (r *onlineUserRepository) GetAll() ([]*entity.OnlineUser, error) {
	onlineUsersListKey := onlineUserSetKey + ":list"
	existingUsers, err := r.cache.Get(onlineUsersListKey)
	if err != nil {
		return nil, err
	}

	var userIDs []string
	if existingUsers != "" {
		json.Unmarshal([]byte(existingUsers), &userIDs)
	}

	users := make([]*entity.OnlineUser, 0, len(userIDs))
	for _, userID := range userIDs {
		user, err := r.GetByUserID(userID)
		if err != nil {
			continue
		}
		if user != nil {
			users = append(users, user)
		}
	}

	return users, nil
}

// GetCount 获取在线用户数量
func (r *onlineUserRepository) GetCount() (int, error) {
	onlineUsersListKey := onlineUserSetKey + ":list"
	existingUsers, err := r.cache.Get(onlineUsersListKey)
	if err != nil {
		return 0, err
	}

	if existingUsers == "" {
		return 0, nil
	}

	var userIDs []string
	json.Unmarshal([]byte(existingUsers), &userIDs)
	return len(userIDs), nil
}

// Exists 检查用户是否在线
func (r *onlineUserRepository) Exists(userID string) (bool, error) {
	userKey := onlineUserKeyPrefix + userID
	_, err := r.cache.Get(userKey)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ClearExpired 清理过期的在线用户
func (r *onlineUserRepository) ClearExpired() error {
	// 缓存会自动处理过期键，此方法可以作为手动清理的补充
	// 清理在线用户列表中已经不存在的用户
	onlineUsersListKey := onlineUserSetKey + ":list"
	existingUsers, err := r.cache.Get(onlineUsersListKey)
	if err != nil {
		return err
	}

	if existingUsers == "" {
		return nil
	}

	var userIDs []string
	json.Unmarshal([]byte(existingUsers), &userIDs)

	validUserIDs := make([]string, 0)
	for _, userID := range userIDs {
		userKey := onlineUserKeyPrefix + userID
		_, err := r.cache.Get(userKey)
		if err == nil {
			// 用户数据还存在
			validUserIDs = append(validUserIDs, userID)
		}
	}

	if len(validUserIDs) > 0 {
		listData, _ := json.Marshal(validUserIDs)
		r.cache.Set(onlineUsersListKey, string(listData), onlineUserExpire)
	} else {
		r.cache.Delete(onlineUsersListKey)
	}

	return nil
}

// serialize 序列化在线用户为 JSON
func serialize(user *entity.OnlineUser) (string, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// deserialize 反序列化 JSON 为在线用户
func deserialize(data string) (*entity.OnlineUser, error) {
	var user entity.OnlineUser
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// generateUserKey 生成用户缓存键
func generateUserKey(userID string) string {
	return onlineUserKeyPrefix + userID
}

// generateSessionKey 生成会话缓存键（可选，用于 sessionID 到 userID 的映射）
func generateSessionKey(sessionID string) string {
	return onlineUserKeyPrefix + "session:" + sessionID
}
