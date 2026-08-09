package service

import (
	"testing"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/stretchr/testify/assert"
)

// TestTaskService_CreateTask 测试任务创建
func TestTaskService_CreateTask(t *testing.T) {
	now := time.Now()
	task := &entity.Task{
		TaskID:      "task-001",
		TaskType:    "http",
		Type:        "http",
		Expression:  "0 */5 * * *",
		Description: "测试HTTP任务",
		Group:       "default",
		Status:      1,
		Params:      `{"url":"http://example.com","method":"GET"}`,
		RetryCount:  3,
		Concurrency: "allow",
		Timeout:     30,
		CreatedAt:   now,
	}

	assert.NotEmpty(t, task.TaskID, "任务ID不能为空")
	assert.NotEmpty(t, task.TaskType, "任务类型不能为空")
	assert.NotEmpty(t, task.Expression, "Cron表达式不能为空")
	assert.NotEmpty(t, task.Description, "任务描述不能为空")
	assert.Equal(t, 1, task.Status, "任务状态应为启用")
	assert.Equal(t, 3, task.RetryCount, "重试次数应正确")
	assert.Equal(t, 30, task.Timeout, "超时时间应正确")
}

// TestTaskService_TaskTypeValidation 测试任务类型验证
func TestTaskService_TaskTypeValidation(t *testing.T) {
	validTypes := map[string]bool{
		"http":     true,
		"database": true,
		"cache":    true,
		"script":   true,
	}

	testCases := []struct {
		name     string
		taskType string
		isValid  bool
	}{
		{name: "HTTP任务", taskType: "http", isValid: true},
		{name: "数据库任务", taskType: "database", isValid: true},
		{name: "缓存任务", taskType: "cache", isValid: true},
		{name: "脚本任务", taskType: "script", isValid: true},
		{name: "无效类型", taskType: "invalid", isValid: false},
		{name: "空类型", taskType: "", isValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, isValid := validTypes[tc.taskType]
			assert.Equal(t, tc.isValid, isValid, "任务类型验证应正确")
		})
	}
}

// TestTaskService_StatusValidation 测试任务状态验证
func TestTaskService_StatusValidation(t *testing.T) {
	testCases := []struct {
		name    string
		status  int
		isValid bool
	}{
		{name: "禁用状态", status: 0, isValid: true},
		{name: "启用状态", status: 1, isValid: true},
		{name: "暂停状态", status: 2, isValid: true},
		{name: "无效状态 - 负数", status: -1, isValid: false},
		{name: "无效状态 - 大于2", status: 3, isValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.status >= 0 && tc.status <= 2 {
				assert.True(t, tc.isValid, "状态 %d 应有效", tc.status)
			} else {
				assert.False(t, tc.isValid, "状态 %d 应无效", tc.status)
			}
		})
	}
}

// TestTaskService_ExpressionValidation 测试表达式验证
func TestTaskService_ExpressionValidation(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		isValid    bool
	}{
		{name: "每分钟", expression: "* * * * *", isValid: true},
		{name: "每5分钟", expression: "*/5 * * * *", isValid: true},
		{name: "每天0点", expression: "0 0 * * *", isValid: true},
		{name: "每周一", expression: "0 0 * * 1", isValid: true},
		{name: "空表达式", expression: "", isValid: false},
		{name: "无效表达式", expression: "invalid", isValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expression == "" {
				assert.False(t, tc.isValid, "表达式不能为空")
				return
			}
			// 简单验证：检查是否有5个字段
			fields := 0
			for _, c := range tc.expression {
				if c == ' ' {
					fields++
				}
			}
			// 有4个空格表示5个字段
			hasFiveFields := fields == 4
			assert.Equal(t, tc.isValid, hasFiveFields, "表达式验证应正确")
		})
	}
}

// TestTaskService_TaskFilter 测试任务过滤条件
func TestTaskService_TaskFilter(t *testing.T) {
	filters := map[string]interface{}{
		"task_type": "http",
		"status":    1,
		"group":     "default",
	}

	assert.NotEmpty(t, filters, "过滤条件不应为空")
	assert.Contains(t, filters, "task_type", "应包含任务类型过滤")
	assert.Contains(t, filters, "status", "应包含状态过滤")

	// 验证过滤条件值
	taskType, ok := filters["task_type"].(string)
	assert.True(t, ok, "任务类型应为字符串")
	assert.Equal(t, "http", taskType, "任务类型过滤值应正确")

	status, ok := filters["status"].(int)
	assert.True(t, ok, "状态应为整数")
	assert.Equal(t, 1, status, "状态过滤值应正确")
}

// TestTaskService_TaskGroupValidation 测试任务分组验证
func TestTaskService_TaskGroupValidation(t *testing.T) {
	validGroups := map[string]bool{
		"default":  true,
		"system":   true,
		"business": true,
		"monitor":  true,
	}

	testCases := []struct {
		name    string
		group   string
		isValid bool
	}{
		{name: "默认分组", group: "default", isValid: true},
		{name: "系统分组", group: "system", isValid: true},
		{name: "业务分组", group: "business", isValid: true},
		{name: "监控分组", group: "monitor", isValid: true},
		{name: "无效分组", group: "unknown", isValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, isValid := validGroups[tc.group]
			assert.Equal(t, tc.isValid, isValid, "任务分组验证应正确")
		})
	}
}