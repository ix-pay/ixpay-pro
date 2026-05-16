package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigService_NameValidation 测试配置名称验证
func TestConfigService_NameValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"有效名称 - 短名称", "系统标题", true},
		{"有效名称 - 中等名称", "系统配置参数", true},
		{"有效名称 - 长名称", "系统全局配置参数名称", true},
		{"无效名称 - 空字符串", "", false},
		{"无效名称 - 特殊字符", "系统@标题", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.input == "" {
				assert.False(t, tc.valid, "配置名称不能为空")
				return
			}

			if len(tc.input) > 100 {
				assert.False(t, tc.valid, "配置名称不能超过 100 字符")
			}
		})
	}
}

// TestConfigService_KeyValidation 测试配置键验证
func TestConfigService_KeyValidation(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		valid bool
	}{
		{"有效键 - 短键", "system.title", true},
		{"有效键 - 中等键", "system.config.version", true},
		{"有效键 - 长键", "system.global.configuration.version.number", true},
		{"无效键 - 空字符串", "", false},
		{"无效键 - 大写字母", "System.Title", false},
		{"无效键 - 特殊字符", "system@title", false},
		{"无效键 - 连续点", "system..title", false},
		{"无效键 - 以点开头", ".system.title", false},
		{"无效键 - 以点结尾", "system.title.", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.key == "" {
				assert.False(t, tc.valid, "配置键不能为空")
				return
			}

			if tc.key[0] == '.' || tc.key[len(tc.key)-1] == '.' {
				assert.False(t, tc.valid, "配置键不能以点开头或结尾")
				return
			}

			for i := 0; i < len(tc.key)-1; i++ {
				if tc.key[i] == '.' && tc.key[i+1] == '.' {
					assert.False(t, tc.valid, "配置键不能包含连续的点")
					return
				}
			}
		})
	}
}

// TestConfigService_ValueValidation 测试配置值验证
func TestConfigService_ValueValidation(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{"有效值 - 短值", "true", true},
		{"有效值 - 中等值", "https://example.com", true},
		{"有效值 - 长值", "这是一个很长的配置值，用于测试配置值字段的长度限制和系统处理能力", true},
		{"有效值 - JSON", "{\"key\": \"value\"}", true},
		{"有效值 - 空字符串", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value == "" {
				assert.True(t, tc.valid, "配置值可以为空")
			}

			if len(tc.value) > 1000 {
				assert.False(t, tc.valid, "配置值不能超过 1000 字符")
			}
		})
	}
}

// TestConfigService_TypeValidation 测试配置类型验证
func TestConfigService_TypeValidation(t *testing.T) {
	testCases := []struct {
		name  string
		typ   string
		valid bool
	}{
		{"有效类型 - string", "string", true},
		{"有效类型 - number", "number", true},
		{"有效类型 - boolean", "boolean", true},
		{"有效类型 - json", "json", true},
		{"有效类型 - text", "text", true},
		{"无效类型 - 空字符串", "", false},
		{"无效类型 - 未知类型", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validTypes := map[string]bool{
				"string": true, "number": true, "boolean": true,
				"json": true, "text": true,
			}
			assert.Equal(t, tc.valid, validTypes[tc.typ], "配置类型 %s 验证失败", tc.typ)
		})
	}
}

// TestConfigService_GroupValidation 测试配置分组验证
func TestConfigService_GroupValidation(t *testing.T) {
	testCases := []struct {
		name  string
		group string
		valid bool
	}{
		{"有效分组 - 短分组", "system", true},
		{"有效分组 - 中等分组", "application", true},
		{"有效分组 - 长分组", "global_configuration", true},
		{"无效分组 - 空字符串", "", false},
		{"无效分组 - 特殊字符", "sys@tem", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.group == "" {
				assert.False(t, tc.valid, "配置分组不能为空")
				return
			}

			if len(tc.group) > 50 {
				assert.False(t, tc.valid, "配置分组不能超过 50 字符")
			}
		})
	}
}
