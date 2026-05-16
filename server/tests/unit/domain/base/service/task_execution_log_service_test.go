package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTaskExecutionLogService_TaskNameValidation 测试任务名称验证
func TestTaskExecutionLogService_TaskNameValidation(t *testing.T) {
	testCases := []struct {
		name     string
		taskName string
		valid    bool
	}{
		{"有效任务名 - 短名称", "数据同步", true},
		{"有效任务名 - 中等名称", "定时数据备份", true},
		{"有效任务名 - 长名称", "每日凌晨数据备份和清理任务", true},
		{"无效任务名 - 空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.taskName == "" {
				assert.False(t, tc.valid, "任务名称不能为空")
			}

			if len(tc.taskName) > 100 {
				assert.False(t, tc.valid, "任务名称不能超过 100 字符")
			}
		})
	}
}

// TestTaskExecutionLogService_StatusValidation 测试任务状态验证
func TestTaskExecutionLogService_StatusValidation(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		valid  bool
	}{
		{"有效状态 - 成功", 1, true},
		{"有效状态 - 失败", 0, true},
		{"有效状态 - 执行中", 2, true},
		{"无效状态 - 负数", -1, false},
		{"无效状态 - 大数", 100, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if status := tc.status; status < 0 || status > 2 {
				assert.False(t, tc.valid, "任务状态必须在有效范围内 (0-2)")
			}
		})
	}
}

// TestTaskExecutionLogService_ErrorMessageValidation 测试错误消息验证
func TestTaskExecutionLogService_ErrorMessageValidation(t *testing.T) {
	testCases := []struct {
		name     string
		errorMsg string
		valid    bool
	}{
		{"有效错误消息 - 短消息", "超时", true},
		{"有效错误消息 - 中等消息", "数据库连接超时，请检查网络", true},
		{"有效错误消息 - 长消息", "任务执行失败：数据库连接超时，请检查网络连接是否正常", true},
		{"有效错误消息 - 空消息", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.errorMsg) > 2000 {
				assert.False(t, tc.valid, "错误消息不能超过 2000 字符")
			}
		})
	}
}

// TestTaskExecutionLogService_ExecutionTimeValidation 测试执行时间验证
func TestTaskExecutionLogService_ExecutionTimeValidation(t *testing.T) {
	testCases := []struct {
		name          string
		executionTime int64
		valid         bool
	}{
		{"有效时间 - 毫秒级", 100, true},
		{"有效时间 - 秒级", 1000, true},
		{"有效时间 - 分钟级", 60000, true},
		{"有效时间 - 0", 0, true},
		{"无效时间 - 负数", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.executionTime < 0 {
				assert.False(t, tc.valid, "执行时间不能为负数")
			}
		})
	}
}

// TestTaskExecutionLogService_RetryCountValidation 测试重试次数验证
func TestTaskExecutionLogService_RetryCountValidation(t *testing.T) {
	testCases := []struct {
		name       string
		retryCount int
		valid      bool
	}{
		{"有效重试 - 0 次", 0, true},
		{"有效重试 - 1 次", 1, true},
		{"有效重试 - 3 次", 3, true},
		{"有效重试 - 5 次", 5, true},
		{"无效重试 - 负数", -1, false},
		{"无效重试 - 过多", 100, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.retryCount < 0 {
				assert.False(t, tc.valid, "重试次数不能为负数")
			}

			if tc.retryCount > 10 {
				assert.False(t, tc.valid, "重试次数不应该超过 10 次")
			}
		})
	}
}
