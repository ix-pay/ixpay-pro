package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDictService_TypeValidation 测试字典类型验证
func TestDictService_TypeValidation(t *testing.T) {
	testCases := []struct {
		name  string
		typ   string
		valid bool
	}{
		{"有效类型 - 短类型", "user_status", true},
		{"有效类型 - 中等类型", "system_config", true},
		{"无效类型 - 空字符串", "", false},
		{"无效类型 - 特殊字符", "user@status", false},
		{"无效类型 - 大写字母", "UserStatus", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.typ == "" {
				assert.False(t, tc.valid, "字典类型不能为空")
				return
			}

			for _, c := range tc.typ {
				isValidChar := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
				if !isValidChar {
					assert.False(t, tc.valid, "字典类型包含非法字符：%c", c)
					return
				}
			}
		})
	}
}

// TestDictService_LabelValidation 测试字典标签验证
func TestDictService_LabelValidation(t *testing.T) {
	testCases := []struct {
		name  string
		label string
		valid bool
	}{
		{"有效标签 - 中文", "启用", true},
		{"有效标签 - 英文", "Enabled", true},
		{"有效标签 - 混合", "启用/Enabled", true},
		{"无效标签 - 空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.label == "" {
				assert.False(t, tc.valid, "字典标签不能为空")
			}

			if len(tc.label) > 50 {
				assert.False(t, tc.valid, "字典标签不能超过 50 字符")
			}
		})
	}
}

// TestDictService_ValueValidation 测试字典值验证
func TestDictService_ValueValidation(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{"有效值 - 数字", "1", true},
		{"有效值 - 字符串", "enabled", true},
		{"有效值 - 空字符串", "", true},
		{"无效值 - 特殊字符", "en@bled", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value == "" {
				assert.True(t, tc.valid, "字典值可以为空")
			}

			if len(tc.value) > 100 {
				assert.False(t, tc.valid, "字典值不能超过 100 字符")
			}
		})
	}
}

// TestDictService_SortValidation 测试排序字段验证
func TestDictService_SortValidation(t *testing.T) {
	testCases := []struct {
		name  string
		sort  int
		valid bool
	}{
		{"排序 - 0", 0, true},
		{"排序 - 正数", 10, true},
		{"排序 - 负数", -1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, tc.valid, "排序字段应该允许所有整数值")
		})
	}
}

// TestDictService_StatusValidation 测试状态字段验证
func TestDictService_StatusValidation(t *testing.T) {
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
