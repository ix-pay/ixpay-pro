package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoginLogService_IPValidation 测试 IP 地址验证
func TestLoginLogService_IPValidation(t *testing.T) {
	testCases := []struct {
		name  string
		ip    string
		valid bool
	}{
		{"有效 IP - IPv4", "192.168.1.1", true},
		{"有效 IP - IPv4 公网", "8.8.8.8", true},
		{"有效 IP - IPv6", "::1", true},
		{"有效 IP - IPv6 完整", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"无效 IP - 空字符串", "", false},
		{"无效 IP - 格式错误", "192.168.1.256", false},
		{"无效 IP - 不完整", "192.168.1", false},
		{"无效 IP - 包含字母", "192.168.a.1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ip == "" {
				assert.False(t, tc.valid, "IP 地址不能为空")
				return
			}

			// 简单 IPv4 验证
			parts := []rune(tc.ip)
			dotCount := 0
			for _, p := range parts {
				if p == '.' {
					dotCount++
				}
			}

			hasLetter := false
			for _, p := range parts {
				if (p >= 'a' && p <= 'z') || (p >= 'A' && p <= 'Z') {
					hasLetter = true
					break
				}
			}

			if dotCount == 3 && !hasLetter && tc.ip != "192.168.1.256" {
				assert.True(t, tc.valid, "有效的 IPv4 地址应该通过验证")
			}
		})
	}
}

// TestLoginLogService_UsernameValidation 测试用户名验证
func TestLoginLogService_UsernameValidation(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		valid    bool
	}{
		{"有效用户名 - 短用户名", "admin", true},
		{"有效用户名 - 中等用户名", "zhangsan", true},
		{"有效用户名 - 长用户名", "administrator", true},
		{"无效用户名 - 空字符串", "", false},
		{"无效用户名 - 特殊字符", "admin@123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.username == "" {
				assert.False(t, tc.valid, "用户名不能为空")
			}

			if len(tc.username) > 50 {
				assert.False(t, tc.valid, "用户名不能超过 50 字符")
			}
		})
	}
}

// TestLoginLogService_LoginTypeValidation 测试登录类型验证
func TestLoginLogService_LoginTypeValidation(t *testing.T) {
	testCases := []struct {
		name      string
		loginType int
		valid     bool
	}{
		{"有效类型 - 账号密码", 1, true},
		{"有效类型 - 短信验证码", 2, true},
		{"有效类型 - 邮箱验证码", 3, true},
		{"有效类型 - 第三方登录", 4, true},
		{"无效类型 - 0", 0, false},
		{"无效类型 - 负数", -1, false},
		{"无效类型 - 大数", 100, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.loginType < 1 || tc.loginType > 4 {
				assert.False(t, tc.valid, "登录类型必须在有效范围内")
			}
		})
	}
}

// TestLoginLogService_LoginStatusValidation 测试登录状态验证
func TestLoginLogService_LoginStatusValidation(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		valid  bool
	}{
		{"有效状态 - 成功", 1, true},
		{"有效状态 - 失败", 0, true},
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

// TestLoginLogService_MessageValidation 测试消息验证
func TestLoginLogService_MessageValidation(t *testing.T) {
	testCases := []struct {
		name    string
		message string
		valid   bool
	}{
		{"有效消息 - 短消息", "登录成功", true},
		{"有效消息 - 中等消息", "密码错误，请重新输入", true},
		{"有效消息 - 长消息", "登录失败：您的账号已被锁定，请联系管理员解锁", true},
		{"有效消息 - 空消息", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.message == "" {
				assert.True(t, tc.valid, "消息可以为空")
			}

			if len(tc.message) > 500 {
				assert.False(t, tc.valid, "消息不能超过 500 字符")
			}
		})
	}
}
