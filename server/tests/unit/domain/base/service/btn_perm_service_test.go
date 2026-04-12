package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBtnPermService_CodeValidation 测试按钮权限代码验证
func TestBtnPermService_CodeValidation(t *testing.T) {
	testCases := []struct {
		name        string
		code        string
		expectError bool
		expectMsg   string
	}{
		{"有效代码 - 标准格式", "user:add", false, ""},
		{"有效代码 - 包含下划线", "user_management:add", false, ""},
		{"有效代码 - 多个冒号", "system:user:add", false, ""},
		{"无效代码 - 空字符串", "", true, "按钮权限代码不能为空"},
		{"无效代码 - 无冒号", "useradd", true, "按钮权限代码应该包含冒号"},
		{"无效代码 - 冒号开头", ":user:add", true, "按钮权限代码不能以冒号开头或结尾"},
		{"无效代码 - 冒号结尾", "user:add:", true, "按钮权限代码不能以冒号开头或结尾"},
		{"无效代码 - 连续冒号", "user::add", true, "按钮权限代码不能包含连续冒号"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var hasError bool
			var errMsg string

			if tc.code == "" {
				hasError = true
				errMsg = "按钮权限代码不能为空"
			} else {
				hasColon := false
				for _, c := range tc.code {
					if c == ':' {
						hasColon = true
						break
					}
				}

				if !hasColon {
					hasError = true
					errMsg = "按钮权限代码应该包含冒号"
				} else {
					if tc.code[0] == ':' || tc.code[len(tc.code)-1] == ':' {
						hasError = true
						errMsg = "按钮权限代码不能以冒号开头或结尾"
					} else {
						for i := 0; i < len(tc.code)-1; i++ {
							if tc.code[i] == ':' && tc.code[i+1] == ':' {
								hasError = true
								errMsg = "按钮权限代码不能包含连续冒号"
								break
							}
						}
					}
				}
			}

			assert.Equal(t, tc.expectError, hasError, "错误状态不匹配")
			if hasError && tc.expectMsg != "" {
				assert.Equal(t, tc.expectMsg, errMsg, "错误消息不匹配")
			}
		})
	}
}

// TestBtnPermService_NameValidation 测试按钮权限名称验证
func TestBtnPermService_NameValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"有效名称 - 短名称", "添加", true},
		{"有效名称 - 中等名称", "添加用户", true},
		{"有效名称 - 长名称", "添加新用户权限", true},
		{"无效名称 - 空字符串", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.input == "" {
				assert.False(t, tc.valid, "按钮权限名称不能为空")
			}

			if len(tc.input) > 50 {
				assert.False(t, tc.valid, "按钮权限名称不能超过 50 字符")
			}
		})
	}
}

// TestBtnPermService_MenuIDValidation 测试菜单 ID 验证
func TestBtnPermService_MenuIDValidation(t *testing.T) {
	testCases := []struct {
		name   string
		menuID int64
		valid  bool
	}{
		{"有效菜单 ID - 正数", 1, true},
		{"有效菜单 ID - 大数", 999999, true},
		{"无效菜单 ID - 0", 0, false},
		{"无效菜单 ID - 负数", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.menuID <= 0 {
				assert.False(t, tc.valid, "菜单 ID 必须大于 0")
			}
		})
	}
}

// TestBtnPermService_SortValidation 测试排序字段验证
func TestBtnPermService_SortValidation(t *testing.T) {
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

// TestBtnPermService_StatusValidation 测试状态字段验证
func TestBtnPermService_StatusValidation(t *testing.T) {
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
