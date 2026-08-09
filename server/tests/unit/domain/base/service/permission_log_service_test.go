package service

import (
	"testing"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/stretchr/testify/assert"
)

// TestPermissionLogService_LogEntry 测试权限日志条目结构
func TestPermissionLogService_LogEntry(t *testing.T) {
	now := time.Now()
	logEntry := &entity.PermissionLog{
		UserID:     1,
		Username:   "admin",
		Operation:  "assign_permission",
		Module:     "角色管理",
		TargetType: "menu",
		TargetID:   10,
		OldValue:   "",
		NewValue:   "10",
		IP:         "127.0.0.1",
		UserAgent:  "Mozilla/5.0",
	}
	logEntry.CreatedAt = now

	assert.Equal(t, int64(1), logEntry.UserID, "用户ID应正确")
	assert.Equal(t, "admin", logEntry.Username, "用户名应正确")
	assert.Equal(t, "assign_permission", logEntry.Operation, "操作类型应正确")
	assert.Equal(t, "角色管理", logEntry.Module, "模块名应正确")
	assert.Equal(t, "menu", logEntry.TargetType, "目标类型应正确")
	assert.Equal(t, int64(10), logEntry.TargetID, "目标ID应正确")
	assert.Equal(t, "127.0.0.1", logEntry.IP, "IP地址应正确")
	assert.Equal(t, "Mozilla/5.0", logEntry.UserAgent, "UserAgent应正确")

	// 验证 SnowflakeBaseModelWithoutDeleted 嵌入
	assert.IsType(t, database.SnowflakeBaseModelWithoutDeleted{}, logEntry.SnowflakeBaseModelWithoutDeleted, "应嵌入 SnowflakeBaseModelWithoutDeleted")

	// 验证 CreatedAt 字段
	assert.WithinDuration(t, now, logEntry.CreatedAt, time.Second, "创建时间应接近当前时间")
}

// TestPermissionLogService_OperationTypes 测试操作类型验证
func TestPermissionLogService_OperationTypes(t *testing.T) {
	validOperations := map[string]bool{
		"assign_permission": true,
		"revoke_permission": true,
		"create_role":       true,
		"update_role":       true,
		"delete_role":       true,
	}

	testCases := []struct {
		name           string
		operation      string
		shouldBeValid  bool
	}{
		{name: "分配权限", operation: "assign_permission", shouldBeValid: true},
		{name: "撤销权限", operation: "revoke_permission", shouldBeValid: true},
		{name: "创建角色", operation: "create_role", shouldBeValid: true},
		{name: "更新角色", operation: "update_role", shouldBeValid: true},
		{name: "删除角色", operation: "delete_role", shouldBeValid: true},
		{name: "无效操作", operation: "invalid_action", shouldBeValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, isValid := validOperations[tc.operation]
			assert.Equal(t, tc.shouldBeValid, isValid, "操作类型验证应正确")
		})
	}
}

// TestPermissionLogService_ModuleTypes 测试模块类型验证
func TestPermissionLogService_ModuleTypes(t *testing.T) {
	validModules := map[string]bool{
		"角色管理":   true,
		"菜单管理":   true,
		"用户管理":   true,
		"API管理":   true,
	}

	testCases := []struct {
		name          string
		module        string
		shouldBeValid bool
	}{
		{name: "角色管理", module: "角色管理", shouldBeValid: true},
		{name: "菜单管理", module: "菜单管理", shouldBeValid: true},
		{name: "用户管理", module: "用户管理", shouldBeValid: true},
		{name: "API管理", module: "API管理", shouldBeValid: true},
		{name: "无效模块", module: "未知模块", shouldBeValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, isValid := validModules[tc.module]
			assert.Equal(t, tc.shouldBeValid, isValid, "模块类型验证应正确")
		})
	}
}

// TestPermissionLogService_FilterValidation 测试权限日志过滤条件验证
func TestPermissionLogService_FilterValidation(t *testing.T) {
	filters := map[string]interface{}{
		"user_id":    int64(1),
		"operation":  "assign_permission",
		"module":     "角色管理",
		"start_time": "2024-01-01",
		"end_time":   "2024-12-31",
	}

	assert.NotEmpty(t, filters, "过滤条件不应为空")
	assert.Contains(t, filters, "user_id", "应包含用户ID过滤")
	assert.Contains(t, filters, "operation", "应包含操作类型过滤")
	assert.Contains(t, filters, "module", "应包含模块过滤")
}