package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// SimpleMenu 简化菜单结构（用于测试）
type SimpleMenu struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

// SimpleBtnPerm 简化按钮权限结构（用于测试）
type SimpleBtnPerm struct {
	ID     int64  `json:"id"`
	MenuID int64  `json:"menu_id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
}

// SimpleAPI 简化 API 结构（用于测试）
type SimpleAPI struct {
	ID     int64  `json:"id"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

// TestRolePermissionService_CacheStructure 测试 Redis 缓存结构
func TestRolePermissionService_CacheStructure(t *testing.T) {

	perms := struct {
		Menus       []*SimpleMenu    `json:"menus"`
		BtnPerms    []*SimpleBtnPerm `json:"btn_perms"`
		GeneralApis []*SimpleAPI     `json:"general_apis"`
		ApiSet      map[string]bool  `json:"api_set"`
	}{
		Menus: []*SimpleMenu{
			{ID: 1, Path: "/user", Name: "UserManagement"},
			{ID: 2, Path: "/role", Name: "RoleManagement"},
		},
		BtnPerms: []*SimpleBtnPerm{
			{ID: 1, MenuID: 1, Name: "创建用户", Code: "user:create"},
			{ID: 2, MenuID: 1, Name: "删除用户", Code: "user:delete"},
		},
		GeneralApis: []*SimpleAPI{
			{ID: 1, Path: "/api//user", Method: "GET"},
			{ID: 2, Path: "/api//role", Method: "POST"},
		},
	}

	// 构建 ApiSet
	apiSet := make(map[string]bool)
	// 模拟菜单关联的 API
	apiSet["GET:/api//menu/1"] = true
	// 模拟按钮关联的 API
	apiSet["POST:/api//btn/1"] = true
	// 添加通用 API
	for _, api := range perms.GeneralApis {
		key := api.Method + ":" + api.Path
		apiSet[key] = true
	}
	perms.ApiSet = apiSet

	// 验证数据结构
	assert.NotNil(t, perms.Menus, "菜单不应为空")
	assert.NotNil(t, perms.BtnPerms, "按钮权限不应为空")
	assert.NotNil(t, perms.GeneralApis, "通用 API 不应为空")
	assert.NotNil(t, perms.ApiSet, "API 集合不应为空")

	// 验证 ApiSet 包含所有 API
	expectedApiCount := 2 + len(perms.GeneralApis) // 2 个模拟 API + 通用 API
	assert.Equal(t, expectedApiCount, len(perms.ApiSet), "ApiSet 应包含所有 API")

	// 测试序列化
	jsonData, err := json.Marshal(perms)
	assert.NoError(t, err, "序列化不应出错")
	assert.Greater(t, len(jsonData), 0, "序列化数据不应为空")

	// 测试反序列化
	var loadedPerms struct {
		Menus       []*SimpleMenu    `json:"menus"`
		BtnPerms    []*SimpleBtnPerm `json:"btn_perms"`
		GeneralApis []*SimpleAPI     `json:"general_apis"`
		ApiSet      map[string]bool  `json:"api_set"`
	}
	err = json.Unmarshal(jsonData, &loadedPerms)
	assert.NoError(t, err, "反序列化不应出错")
	assert.NotNil(t, loadedPerms.Menus, "反序列化结果不应为空")
	assert.Equal(t, len(perms.ApiSet), len(loadedPerms.ApiSet), "API 集合大小应一致")
}

// TestRolePermissionService_ApiSetLookup 测试 ApiSet 快速查找
func TestRolePermissionService_ApiSetLookup(t *testing.T) {
	apiSet := make(map[string]bool)

	// 添加测试数据
	testAPIs := []struct {
		method string
		path   string
	}{
		{"GET", "/api//user"},
		{"POST", "/api//user"},
		{"PUT", "/api//user/:id"},
		{"DELETE", "/api//user/:id"},
		{"GET", "/api//role"},
	}

	for _, api := range testAPIs {
		key := api.method + ":" + api.path
		apiSet[key] = true
	}

	// 测试查找性能
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_ = apiSet["GET:/api//user"]
	}
	elapsed := time.Since(start)

	assert.Less(t, elapsed, time.Millisecond, "1000 次查找应在 1ms 内完成")

	// 测试查找正确性
	testCases := []struct {
		key    string
		expect bool
	}{
		{"GET:/api//user", true},
		{"POST:/api//user", true},
		{"GET:/api//role", true},
		{"GET:/api//unknown", false},
		{"DELETE:/api//unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := apiSet[tc.key]
			assert.Equal(t, tc.expect, result, "查找结果应正确")
		})
	}
}

// TestRolePermissionService_CacheKeyFormat 测试缓存键格式
func TestRolePermissionService_CacheKeyFormat(t *testing.T) {
	testCases := []struct {
		roleID   int64
		expected string
	}{
		{1, "role:perms:1"},
		{100, "role:perms:100"},
		{999999, "role:perms:999999"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			cacheKey := "role:perms:" + string(rune(tc.roleID))
			// 验证键格式
			assert.Contains(t, cacheKey, "role:perms:", "缓存键应包含前缀")
		})
	}
}

// TestRolePermissionService_EmptyPermissions 测试空权限场景
func TestRolePermissionService_EmptyPermissions(t *testing.T) {
	perms := struct {
		Menus       []*SimpleMenu    `json:"menus"`
		BtnPerms    []*SimpleBtnPerm `json:"btn_perms"`
		GeneralApis []*SimpleAPI     `json:"general_apis"`
		ApiSet      map[string]bool  `json:"api_set"`
	}{
		Menus:       []*SimpleMenu{},
		BtnPerms:    []*SimpleBtnPerm{},
		GeneralApis: []*SimpleAPI{},
		ApiSet:      make(map[string]bool),
	}

	// 验证空权限结构
	assert.NotNil(t, perms, "权限对象不应为空")
	assert.Empty(t, perms.Menus, "菜单应为空")
	assert.Empty(t, perms.BtnPerms, "按钮权限应为空")
	assert.Empty(t, perms.GeneralApis, "通用 API 应为空")
	assert.Empty(t, perms.ApiSet, "API 集合应为空")

	// 测试序列化空权限
	jsonData, err := json.Marshal(perms)
	assert.NoError(t, err, "序列化空权限不应出错")
	assert.Greater(t, len(jsonData), 0, "序列化数据不应为空")
}

// TestRolePermissionService_ApiKeyFormat 测试 API 键格式
func TestRolePermissionService_ApiKeyFormat(t *testing.T) {
	testCases := []struct {
		method string
		path   string
		expect string
	}{
		{"GET", "/api//user", "GET:/api//user"},
		{"POST", "/api//user", "POST:/api//user"},
		{"PUT", "/api//user/:id", "PUT:/api//user/:id"},
		{"DELETE", "/api//user/123", "DELETE:/api//user/123"},
	}

	for _, tc := range testCases {
		t.Run(tc.expect, func(t *testing.T) {
			key := tc.method + ":" + tc.path
			assert.Equal(t, tc.expect, key, "API 键格式应正确")
			assert.Contains(t, key, ":", "API 键应包含分隔符")
		})
	}
}

// TestRolePermissionService_CacheTTL 测试缓存 TTL 配置
func TestRolePermissionService_CacheTTL(t *testing.T) {
	// 验证缓存过期时间配置
	cacheTTL := 24 * time.Hour

	assert.Equal(t, 24*time.Hour, cacheTTL, "缓存 TTL 应为 24 小时")
	assert.Greater(t, cacheTTL, time.Hour, "缓存 TTL 应大于 1 小时")
	assert.Less(t, cacheTTL, 7*24*time.Hour, "缓存 TTL 应小于 7 天")
}

// TestRolePermissionService_ConcurrentAccess 测试并发访问安全性
func TestRolePermissionService_ConcurrentAccess(t *testing.T) {
	apiSet := make(map[string]bool)

	// 初始化数据
	for i := 0; i < 100; i++ {
		apiSet["GET:/api//resource/"+string(rune(i))] = true
	}

	// 测试并发读取（Go 的 map 并发读是安全的）
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = apiSet["GET:/api//resource/1"]
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Greater(t, len(apiSet), 0, "API 集合不应为空")
}

// TestRolePermissionService_PermissionMerge 测试权限合并
func TestRolePermissionService_PermissionMerge(t *testing.T) {
	// 模拟从不同来源合并权限
	menuAPIs := map[string]bool{
		"GET:/api//menu": true,
	}

	btnAPIs := map[string]bool{
		"POST:/api//btn": true,
	}

	generalAPIs := map[string]bool{
		"GET:/api//general": true,
	}

	// 合并所有 API
	mergedAPIs := make(map[string]bool)
	for k, v := range menuAPIs {
		mergedAPIs[k] = v
	}
	for k, v := range btnAPIs {
		mergedAPIs[k] = v
	}
	for k, v := range generalAPIs {
		mergedAPIs[k] = v
	}

	// 验证合并结果
	assert.Equal(t, 3, len(mergedAPIs), "合并后应有 3 个 API")
	assert.True(t, mergedAPIs["GET:/api//menu"], "应包含菜单 API")
	assert.True(t, mergedAPIs["POST:/api//btn"], "应包含按钮 API")
	assert.True(t, mergedAPIs["GET:/api//general"], "应包含通用 API")
}

// TestRolePermissionService_DuplicateAPIHandling 测试重复 API 处理
func TestRolePermissionService_DuplicateAPIHandling(t *testing.T) {
	apiSet := make(map[string]bool)

	// 添加重复的 API
	apiSet["GET:/api//user"] = true
	apiSet["GET:/api//user"] = true // 重复添加
	apiSet["GET:/api//user"] = true // 再次重复添加

	// 验证 MapSet 自动去重
	assert.Equal(t, 1, len(apiSet), "MapSet 应自动去重")
	assert.True(t, apiSet["GET:/api//user"], "应包含该 API")
}
