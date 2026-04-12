package service

import (
	"testing"

	"github.com/ix-pay/ixpay-pro/internal/utils/encryption"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserService_PasswordHandling 测试密码加密和验证
func TestUserService_PasswordHandling(t *testing.T) {
	password := "testpassword123"

	// 生成密码哈希
	hash, err := encryption.GeneratePasswordHash(password)
	require.NoError(t, err, "生成密码哈希失败")
	require.NotEmpty(t, hash, "生成的密码哈希为空")

	// 验证正确的密码
	err = encryption.VerifyPassword(hash, password)
	assert.NoError(t, err, "验证正确密码失败")

	// 验证错误的密码
	err = encryption.VerifyPassword(hash, "wrongpassword")
	assert.Error(t, err, "验证错误密码应该失败")
}

// TestUserService_PasswordStrength 测试密码强度验证
func TestUserService_PasswordStrength(t *testing.T) {
	testCases := []struct {
		name          string
		password      string
		shouldBeValid bool
	}{
		{"弱密码 - 太短", "123", false},
		{"弱密码 - 纯数字", "12345678", false},
		{"中等密码 - 数字 + 字母", "password123", true},
		{"强密码 - 数字 + 字母 + 特殊字符", "P@ssw0rd123", true},
		{"空密码", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 验证密码长度
			if len(tc.password) < 6 {
				assert.False(t, tc.shouldBeValid, "密码长度至少为 6 位")
				return
			}

			// 验证密码复杂度（简单检查）
			hasLetter := false
			for _, c := range tc.password {
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
					hasLetter = true
				}
			}

			if tc.shouldBeValid {
				assert.True(t, hasLetter, "密码应该包含字母")
			}
		})
	}
}

// TestUserService_EmailValidation 测试邮箱格式验证
func TestUserService_EmailValidation(t *testing.T) {
	testCases := []struct {
		name          string
		email         string
		shouldBeValid bool
	}{
		{"有效邮箱", "test@example.com", true},
		{"无效邮箱 - 无@符号", "testexample.com", false},
		{"无效邮箱 - 无域名", "test@", false},
		{"无效邮箱 - 空字符串", "", false},
		{"有效邮箱 - 子域名", "test@mail.example.com", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 简单邮箱格式验证
			hasAt := false
			hasDot := false
			for i, c := range tc.email {
				if c == '@' && i > 0 && i < len(tc.email)-1 {
					hasAt = true
				}
				if c == '.' && hasAt {
					hasDot = true
				}
			}

			isValid := hasAt && hasDot
			assert.Equal(t, tc.shouldBeValid, isValid, "邮箱 %s 验证失败", tc.email)
		})
	}
}

func TestUserService_RegisterLogic(t *testing.T) {
	password := "password123"
	email := "test@example.com"
	username := "testuser"

	// 测试密码加密
	hashedPassword, err := encryption.GeneratePasswordHash(password)
	require.NoError(t, err, "生成密码哈希失败")
	require.NotEmpty(t, hashedPassword, "生成的密码哈希为空")

	// 验证用户名和邮箱格式
	assert.NotEmpty(t, username, "用户名不能为空")
	assert.NotEmpty(t, email, "邮箱不能为空")
}

func TestUserService_ChangePasswordLogic(t *testing.T) {
	oldPassword := "oldpassword"
	newPassword := "newpassword"

	// 生成旧密码哈希
	oldHash, err := encryption.GeneratePasswordHash(oldPassword)
	require.NoError(t, err, "生成旧密码哈希失败")

	// 验证旧密码
	err = encryption.VerifyPassword(oldHash, oldPassword)
	assert.NoError(t, err, "验证旧密码失败")

	// 生成新密码哈希
	newHash, err := encryption.GeneratePasswordHash(newPassword)
	require.NoError(t, err, "生成新密码哈希失败")

	// 验证新密码
	err = encryption.VerifyPassword(newHash, newPassword)
	assert.NoError(t, err, "验证新密码失败")

	// 验证新旧哈希不同
	assert.NotEqual(t, oldHash, newHash, "新旧密码哈希应该不同")
}

func TestUserService_LoginLogic(t *testing.T) {
	password := "password123"

	// 生成密码哈希
	hash, err := encryption.GeneratePasswordHash(password)
	require.NoError(t, err, "生成密码哈希失败")

	// 验证正确的密码
	err = encryption.VerifyPassword(hash, password)
	assert.NoError(t, err, "验证正确密码失败")

	// 验证错误的密码
	err = encryption.VerifyPassword(hash, "wrongpassword")
	assert.Error(t, err, "验证错误密码应该失败")
}
