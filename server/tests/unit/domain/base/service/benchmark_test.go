package service

import (
	"testing"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
)

// BenchmarkOnlineUserService_AddOnlineUser 基准测试：添加在线用户
func BenchmarkOnlineUserService_AddOnlineUser(b *testing.B) {
	repo := NewMockOnlineUserRepositoryForTest()
	log := &MockLogger{}
	service := service.NewOnlineUserService(repo, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := string(rune(i + 1))
		err := service.AddOnlineUser(
			userID,
			"user"+string(rune(i+1)),
			"用户"+string(rune(i+1)),
			"session_"+string(rune(i+1)),
			"192.168.1."+string(rune(i%256)),
			"测试地点",
			"Chrome",
			"Chrome 120",
			"Windows 10",
			"Mozilla/5.0",
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkOnlineUserService_GetOnlineUser 基准测试：获取在线用户
func BenchmarkOnlineUserService_GetOnlineUser(b *testing.B) {
	repo := NewMockOnlineUserRepositoryForTest()
	log := &MockLogger{}
	service := service.NewOnlineUserService(repo, log)

	// 准备测试数据
	for i := 0; i < 100; i++ {
		userID := string(rune(i + 1))
		_ = service.AddOnlineUser(
			userID,
			"user"+string(rune(i+1)),
			"用户"+string(rune(i+1)),
			"session_"+string(rune(i+1)),
			"192.168.1.1",
			"测试地点",
			"Chrome",
			"Chrome 120",
			"Windows 10",
			"Mozilla/5.0",
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := string(rune(i%100 + 1))
		_, _ = service.GetOnlineUserByID(userID)
	}
}

// BenchmarkOnlineUserService_GetOnlineUserList 基准测试：获取在线用户列表
func BenchmarkOnlineUserService_GetOnlineUserList(b *testing.B) {
	repo := NewMockOnlineUserRepositoryForTest()
	log := &MockLogger{}
	service := service.NewOnlineUserService(repo, log)

	// 准备大量测试数据
	for i := 0; i < 1000; i++ {
		userID := string(rune(i + 1))
		_ = service.AddOnlineUser(
			userID,
			"user"+string(rune(i+1)),
			"用户"+string(rune(i+1)),
			"session_"+string(rune(i+1)),
			"192.168.1.1",
			"测试地点",
			"Chrome",
			"Chrome 120",
			"Windows 10",
			"Mozilla/5.0",
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetOnlineUserList()
	}
}

// BenchmarkConfigService_GetConfigByKey 基准测试：根据配置键获取配置
func BenchmarkConfigService_GetConfigByKey(b *testing.B) {
	repo := NewMockConfigRepositoryForTest()
	log := &MockLogger{}
	service := service.NewConfigService(repo, log)

	// 准备测试数据
	for i := 0; i < 100; i++ {
		key := "bench.config.key." + string(rune(i+1))
		config := &entity.Config{
			ConfigKey:   key,
			ConfigValue: "value_" + string(rune(i+1)),
			ConfigType:  "string",
			Description: "基准测试配置",
			Status:      1,
		}
		_ = repo.Create(config)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "bench.config.key." + string(rune(i%100+1))
		_, _ = service.GetConfigByKey(key)
	}
}

// BenchmarkConfigService_GetConfigList 基准测试：获取配置列表
func BenchmarkConfigService_GetConfigList(b *testing.B) {
	repo := NewMockConfigRepositoryForTest()
	log := &MockLogger{}
	service := service.NewConfigService(repo, log)

	// 准备大量测试数据
	for i := 0; i < 1000; i++ {
		key := "bench.config.key." + string(rune(i+1))
		config := &entity.Config{
			ConfigKey:   key,
			ConfigValue: "value_" + string(rune(i+1)),
			ConfigType:  "string",
			Description: "基准测试配置",
			Status:      1,
		}
		_ = repo.Create(config)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.GetConfigList(1, 20, map[string]interface{}{"status": 1})
	}
}
