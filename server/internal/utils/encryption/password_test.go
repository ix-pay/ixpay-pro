package encryption

import (
	"crypto/subtle"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestGeneratePasswordHash(t *testing.T) {
	password := "testpassword123"

	// 生成密码哈希
	hash, err := GeneratePasswordHash(password)
	if err != nil {
		t.Fatalf("生成密码哈希失败: %v", err)
	}

	// 验证哈希不为空
	if hash == "" {
		t.Fatal("生成的密码哈希为空")
	}

	// 验证哈希格式正确（以$argon2id$v=19$开头）
	const expectedPrefix = "$argon2id$v=19$"
	if len(hash) < len(expectedPrefix) || hash[:len(expectedPrefix)] != expectedPrefix {
		t.Fatalf("生成的密码哈希格式不正确，应该以 %s 开头", expectedPrefix)
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"

	// 生成密码哈希
	hash, err := GeneratePasswordHash(password)
	if err != nil {
		t.Fatalf("生成密码哈希失败: %v", err)
	}

	// 使用正确的密码验证
	err = VerifyPassword(hash, password)
	if err != nil {
		t.Fatalf("验证正确密码失败: %v", err)
	}

	// 使用错误的密码验证
	err = VerifyPassword(hash, "wrongpassword")
	if err == nil {
		t.Fatal("验证错误密码应该失败，但成功了")
	}
}

func TestGenerateRandomPasswordHash(t *testing.T) {
	// 生成随机密码哈希
	hash1 := GenerateRandomPasswordHash()
	hash2 := GenerateRandomPasswordHash()

	// 验证哈希不为空
	if hash1 == "" || hash2 == "" {
		t.Fatal("生成的随机密码哈希为空")
	}

	// 验证两次生成的随机哈希不同（概率上应该如此）
	if hash1 == hash2 {
		t.Fatal("两次生成的随机密码哈希相同，这是极不可能的")
	}

	// 验证哈希格式正确
	const expectedPrefix = "$argon2id$v=19$"
	if len(hash1) < len(expectedPrefix) || hash1[:len(expectedPrefix)] != expectedPrefix {
		t.Fatalf("生成的随机密码哈希格式不正确，应该以 %s 开头", expectedPrefix)
	}
}

func TestDecodeHash(t *testing.T) {
	password := "testpassword123"

	// 生成密码哈希
	hash, err := GeneratePasswordHash(password)
	if err != nil {
		t.Fatalf("生成密码哈希失败: %v", err)
	}

	// 解析哈希
	params, salt, hashBytes, err := decodeHash(hash)
	if err != nil {
		t.Fatalf("解析哈希失败: %v", err)
	}

	// 验证参数
	if params.Memory != 64*1024 || params.Iterations != 1 || params.Parallelism != 2 {
		t.Fatalf("解析的参数不正确: %+v", params)
	}

	// 验证salt和hash不为空
	if len(salt) == 0 || len(hashBytes) == 0 {
		t.Fatal("解析的salt或hash为空")
	}

	// 验证使用解析的参数可以正确验证密码
	otherHash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)
	if subtle.ConstantTimeCompare(hashBytes, otherHash) != 1 {
		t.Fatal("使用解析的参数生成的哈希与原哈希不匹配")
	}
}

func TestEdgeCases(t *testing.T) {
	// 测试空密码
	hash, err := GeneratePasswordHash("")
	if err != nil {
		t.Fatalf("生成空密码哈希失败: %v", err)
	}

	err = VerifyPassword(hash, "")
	if err != nil {
		t.Fatalf("验证空密码失败: %v", err)
	}

	// 测试极长密码
	longPassword := make([]byte, 10000)
	for i := range longPassword {
		longPassword[i] = 'a'
	}

	hash, err = GeneratePasswordHash(string(longPassword))
	if err != nil {
		t.Fatalf("生成极长密码哈希失败: %v", err)
	}

	err = VerifyPassword(hash, string(longPassword))
	if err != nil {
		t.Fatalf("验证极长密码失败: %v", err)
	}

	// 测试无效的哈希格式
	err = VerifyPassword("invalidhashformat", "password")
	if err == nil {
		t.Fatal("验证无效格式的哈希应该失败，但成功了")
	}
}
