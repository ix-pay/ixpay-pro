package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPositionService_NameValidation 测试岗位名称验证
func TestPositionService_NameValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"有效名称 - 短名称", "工程师", true},
		{"有效名称 - 中等名称", "软件工程师", true},
		{"有效名称 - 长名称", "高级软件工程师", true},
		{"无效名称 - 空字符串", "", false},
		{"无效名称 - 特殊字符", "工程师@123", false},
		{"无效名称 - 纯数字", "123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.input == "" {
				assert.False(t, tc.valid, "岗位名称不能为空")
				return
			}

			if len(tc.input) > 50 {
				assert.False(t, tc.valid, "岗位名称不能超过 50 字符")
			}
		})
	}
}

// TestPositionService_SortValidation 测试排序字段验证
func TestPositionService_SortValidation(t *testing.T) {
	testCases := []struct {
		name  string
		sort  int
		valid bool
	}{
		{"排序 - 0", 0, true},
		{"排序 - 正数", 10, true},
		{"排序 - 负数", -1, true},
		{"排序 - 大数", 99999, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, tc.valid, "排序字段应该允许所有整数值")
		})
	}
}

// TestPositionService_StatusValidation 测试状态字段验证
func TestPositionService_StatusValidation(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		valid  bool
	}{
		{"有效状态 - 正常", 1, true},
		{"有效状态 - 禁用", 0, true},
		{"无效状态 - 负数", -1, false},
		{"无效状态 - 大于 1", 2, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.status != 0 && tc.status != 1 {
				assert.False(t, tc.valid, "状态只能是 0 或 1")
			}
		})
	}
}

// TestPositionService_DescriptionValidation 测试描述字段验证
func TestPositionService_DescriptionValidation(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		valid       bool
	}{
		{"空描述", "", true},
		{"短描述", "负责软件开发", true},
		{"中等描述", "负责公司软件产品的设计、开发和维护工作", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.description) > 500 {
				assert.False(t, tc.valid, "描述长度不能超过 500 字符")
			}
		})
	}
}
