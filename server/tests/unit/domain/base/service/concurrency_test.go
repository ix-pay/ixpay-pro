package service

import (
	"sync"
	"testing"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/stretchr/testify/assert"
)

// MockOnlineUserRepository 用于并发测试的 Mock 实现
type MockOnlineUserRepository struct {
	users map[int64]*entity.OnlineUser
	mu    sync.RWMutex
}

func NewMockOnlineUserRepositoryForTest() *MockOnlineUserRepository {
	return &MockOnlineUserRepository{
		users: make(map[int64]*entity.OnlineUser),
	}
}

func (m *MockOnlineUserRepository) Add(user *entity.OnlineUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.UserID] = user
	return nil
}

func (m *MockOnlineUserRepository) GetByUserID(userID int64) (*entity.OnlineUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.users[userID], nil
}

func (m *MockOnlineUserRepository) GetBySessionID(sessionID string) (*entity.OnlineUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.SessionID == sessionID {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockOnlineUserRepository) GetAll() ([]*entity.OnlineUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	users := make([]*entity.OnlineUser, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockOnlineUserRepository) Remove(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, userID)
	return nil
}

func (m *MockOnlineUserRepository) RemoveBySessionID(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for userID, user := range m.users {
		if user.SessionID == sessionID {
			delete(m.users, userID)
			break
		}
	}
	return nil
}

func (m *MockOnlineUserRepository) UpdateActiveTime(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, exists := m.users[userID]; exists {
		user.LastActiveAt = time.Now()
	}
	return nil
}

func (m *MockOnlineUserRepository) Exists(userID int64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.users[userID]
	return exists, nil
}

func (m *MockOnlineUserRepository) GetCount() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.users), nil
}

func (m *MockOnlineUserRepository) ClearExpired() error {
	// 简单实现，实际应该清理过期用户
	return nil
}

// TestOnlineUserService_ConcurrentAddUsers 测试并发添加在线用户
func TestOnlineUserService_ConcurrentAddUsers(t *testing.T) {
	repo := NewMockOnlineUserRepositoryForTest()
	log := &MockLogger{}
	svc := service.NewOnlineUserService(repo, log)

	concurrency := 100
	var wg sync.WaitGroup
	errChan := make(chan error, concurrency)

	// 并发添加用户
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			err := svc.AddOnlineUser(
				int64(index),
				"user"+string(rune(index)),
				"用户"+string(rune(index)),
				"session_"+string(rune(index)),
				"192.168.1."+string(rune(index%256)),
				"测试地点",
				"Chrome",
				"Chrome 120",
				"Windows 10",
				"Mozilla/5.0",
			)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		assert.Fail(t, "并发添加用户失败", "error: %v", err)
	}

	// 验证所有用户都已添加
	count, err := repo.GetCount()
	assert.NoError(t, err)
	assert.Equal(t, concurrency, count, "应该成功添加所有用户")
}

// TestOnlineUserService_ConcurrentReadWrite 测试并发读写操作
func TestOnlineUserService_ConcurrentReadWrite(t *testing.T) {
	repo := NewMockOnlineUserRepositoryForTest()
	log := &MockLogger{}
	svc := service.NewOnlineUserService(repo, log)

	// 先添加一些用户
	for i := 0; i < 10; i++ {
		err := svc.AddOnlineUser(
			int64(i),
			"user"+string(rune(i)),
			"用户"+string(rune(i)),
			"session_"+string(rune(i)),
			"192.168.1.1",
			"测试地点",
			"Chrome",
			"Chrome 120",
			"Windows 10",
			"Mozilla/5.0",
		)
		assert.NoError(t, err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 200)

	// 并发读操作（50 个 goroutine）
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			userID := int64(index % 10)
			_, err := svc.GetOnlineUserByID(userID)
			if err != nil && err.Error() != "用户不在线" {
				errChan <- err
			}
		}(i)
	}

	// 并发写操作（50 个 goroutine）
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			userID := int64(index % 10)
			err := svc.UpdateUserActive(userID)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	// 并发移除操作（50 个 goroutine）
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			userID := int64(index % 10)
			_ = svc.RemoveOnlineUser(userID)
		}(i)
	}

	// 并发检查在线状态（50 个 goroutine）
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			userID := int64(index % 10)
			_, err := svc.IsOnline(userID)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		assert.Fail(t, "并发操作失败", "error: %v", err)
	}
}

// TestOnlineUserService_ConcurrentForceOffline 测试强制下线的并发安全性
func TestOnlineUserService_ConcurrentForceOffline(t *testing.T) {
	repo := NewMockOnlineUserRepositoryForTest()
	log := &MockLogger{}
	svc := service.NewOnlineUserService(repo, log)

	// 添加一个测试用户
	err := svc.AddOnlineUser(
		int64(1),
		"testuser",
		"测试用户",
		"session_1",
		"192.168.1.1",
		"测试地点",
		"Chrome",
		"Chrome 120",
		"Windows 10",
		"Mozilla/5.0",
	)
	assert.NoError(t, err)

	concurrency := 50
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// 并发尝试强制下线同一个用户
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			err := svc.ForceOffline(int64(1), string(rune(index)))
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successCount++
			}
		}(i)
	}

	wg.Wait()

	// 只有一个 goroutine 应该成功
	assert.Equal(t, 1, successCount, "应该只有一个 goroutine 成功强制下线")

	// 验证用户已被移除
	online, err := svc.IsOnline(int64(1))
	assert.NoError(t, err)
	assert.False(t, online, "用户应该已被强制下线")
}

// TestOnlineUserService_ConcurrentGetOnlineCount 测试获取在线用户数量的并发安全性
func TestOnlineUserService_ConcurrentGetOnlineCount(t *testing.T) {
	repo := NewMockOnlineUserRepositoryForTest()
	log := &MockLogger{}
	svc := service.NewOnlineUserService(repo, log)

	// 添加一些用户
	for i := 0; i < 10; i++ {
		err := svc.AddOnlineUser(
			int64(i),
			"user"+string(rune(i)),
			"用户"+string(rune(i)),
			"session_"+string(rune(i)),
			"192.168.1.1",
			"测试地点",
			"Chrome",
			"Chrome 120",
			"Windows 10",
			"Mozilla/5.0",
		)
		assert.NoError(t, err)
	}

	var wg sync.WaitGroup
	counts := make([]int, 100)

	// 并发获取在线用户数量
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			count, err := svc.GetOnlineCount()
			if err == nil {
				counts[index] = count
			}
		}(i)
	}

	wg.Wait()

	// 验证所有读取的数量都一致
	expectedCount := 10
	for _, count := range counts {
		if count > 0 {
			assert.Equal(t, expectedCount, count, "并发读取的在线用户数量应该一致")
		}
	}
}

// TestConfigService_ConcurrentAccess 测试配置服务的并发访问
func TestConfigService_ConcurrentAccess(t *testing.T) {
	repo := NewMockConfigRepositoryForTest()
	log := &MockLogger{}
	svc := service.NewConfigService(repo, log)

	// 先创建一些配置
	for i := 0; i < 5; i++ {
		key := "test.key." + string(rune(i))
		config := &entity.Config{
			ConfigKey:   key,
			ConfigValue: "value" + string(rune(i)),
			ConfigType:  1, // 改为 int 类型
			Description: "测试配置",
			Status:      1,
		}
		err := repo.Create(config)
		if err != nil {
			t.Logf("创建配置失败：%v", err)
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 300)

	// 并发读操作（100 个 goroutine）
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := "test.key." + string(rune(index%5))
			_, err := svc.GetConfigByKey(key)
			if err != nil && err.Error() != "配置不存在" {
				errChan <- err
			}
		}(i)
	}

	// 并发写操作（100 个 goroutine）
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := "test.key." + string(rune(index%5))
			id := int64(index%5 + 1)
			_ = svc.UpdateConfig(id, key, "newvalue"+string(rune(index)), 1, "更新配置", 1, int64(index%10+1))
		}(i)
	}

	// 并发获取列表（100 个 goroutine）
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, _, err := svc.GetConfigList(1, 10, nil)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误（忽略预期的错误）
	for err := range errChan {
		// 只报告非预期的错误
		assert.Nil(t, err, "并发操作失败：%v", err)
	}
}

// MockConfigRepository 用于并发测试的 Mock 实现
type MockConfigRepositoryForTest struct {
	configs map[int64]*entity.Config
	keyMap  map[string]int64
	mu      sync.RWMutex
	nextID  int64
}

func NewMockConfigRepositoryForTest() *MockConfigRepositoryForTest {
	return &MockConfigRepositoryForTest{
		configs: make(map[int64]*entity.Config),
		keyMap:  make(map[string]int64),
		nextID:  1,
	}
}

func (m *MockConfigRepositoryForTest) GetByKey(configKey string) (*entity.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if id, exists := m.keyMap[configKey]; exists {
		return m.configs[id], nil
	}
	return nil, nil
}

func (m *MockConfigRepositoryForTest) GetByID(id int64) (*entity.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configs[id], nil
}

func (m *MockConfigRepositoryForTest) Create(config *entity.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	config.ID = m.nextID
	m.nextID++
	m.configs[config.ID] = config
	m.keyMap[config.ConfigKey] = config.ID
	return nil
}

func (m *MockConfigRepositoryForTest) Update(config *entity.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[config.ID] = config
	m.keyMap[config.ConfigKey] = config.ID
	return nil
}

func (m *MockConfigRepositoryForTest) Delete(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.configs, id)
	for key, configID := range m.keyMap {
		if configID == id {
			delete(m.keyMap, key)
			break
		}
	}
	return nil
}

func (m *MockConfigRepositoryForTest) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Config, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	configs := make([]*entity.Config, 0, len(m.configs))
	for _, config := range m.configs {
		configs = append(configs, config)
	}
	return configs, int64(len(configs)), nil
}

func (m *MockConfigRepositoryForTest) GetAllActive() ([]*entity.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	configs := make([]*entity.Config, 0, len(m.configs))
	for _, config := range m.configs {
		if config.Status == 1 {
			configs = append(configs, config)
		}
	}
	return configs, nil
}

// MockLogger 简单的 Mock 日志实现
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...interface{})  {}
func (m *MockLogger) Info(msg string, fields ...interface{})   {}
func (m *MockLogger) Warn(msg string, fields ...interface{})   {}
func (m *MockLogger) Error(msg string, fields ...interface{})  {}
func (m *MockLogger) Fatal(msg string, fields ...interface{})  {}
func (m *MockLogger) With(fields ...interface{}) logger.Logger { return &MockLogger{} }
func (m *MockLogger) Sync() error                              { return nil }
