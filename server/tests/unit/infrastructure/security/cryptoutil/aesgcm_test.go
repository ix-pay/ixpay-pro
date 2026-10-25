package cryptoutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutil "github.com/ix-pay/ixpay-pro/internal/infrastructure/security/cryptoutil"
)

// TestNewAESGCMEncryptor_EmptyKey 空密钥返回 nil 加密器（未启用加密，按明文处理）
func TestNewAESGCMEncryptor_EmptyKey(t *testing.T) {
	enc, err := cryptoutil.NewAESGCMEncryptor("")
	require.NoError(t, err)
	assert.Nil(t, enc)

	enc, err = cryptoutil.NewAESGCMEncryptor("   ")
	require.NoError(t, err)
	assert.Nil(t, enc)
}

// TestNewAESGCMEncryptor_ValidKey 合法密钥创建成功
func TestNewAESGCMEncryptor_ValidKey(t *testing.T) {
	enc, err := cryptoutil.NewAESGCMEncryptor("test-master-key")
	require.NoError(t, err)
	assert.NotNil(t, enc)
}

// TestAESGCMEncryptor_EncryptDecryptRoundTrip 加解密往返一致
func TestAESGCMEncryptor_EncryptDecryptRoundTrip(t *testing.T) {
	enc, err := cryptoutil.NewAESGCMEncryptor("round-trip-key")
	require.NoError(t, err)

	cases := []string{
		"hello-world",
		"微信AppSecret中文内容",
		"-----BEGIN PRIVATE KEY-----\nMIIEvQ...\n-----END PRIVATE KEY-----",
		strings.Repeat("a", 1024),
	}
	for _, plain := range cases {
		cipherText, err := enc.Encrypt(plain)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(cipherText, "enc:v1:"), "密文应带 enc:v1: 前缀")
		assert.NotEqual(t, plain, cipherText)

		got, err := enc.Decrypt(cipherText)
		require.NoError(t, err)
		assert.Equal(t, plain, got)
	}
}

// TestAESGCMEncryptor_RandomNonce 同一明文两次加密产生不同密文（随机 nonce）
func TestAESGCMEncryptor_RandomNonce(t *testing.T) {
	enc, err := cryptoutil.NewAESGCMEncryptor("nonce-key")
	require.NoError(t, err)

	c1, err := enc.Encrypt("same-plain")
	require.NoError(t, err)
	c2, err := enc.Encrypt("same-plain")
	require.NoError(t, err)
	assert.NotEqual(t, c1, c2)
}

// TestAESGCMEncryptor_PlainTextPassthrough 无前缀的历史明文原样返回（兼容存量数据）
func TestAESGCMEncryptor_PlainTextPassthrough(t *testing.T) {
	enc, err := cryptoutil.NewAESGCMEncryptor("passthrough-key")
	require.NoError(t, err)

	got, err := enc.Decrypt("plain-secret-not-encrypted")
	require.NoError(t, err)
	assert.Equal(t, "plain-secret-not-encrypted", got)
}

// TestAESGCMEncryptor_WrongKey 不同密钥解密应失败（GCM 认证不通过）
func TestAESGCMEncryptor_WrongKey(t *testing.T) {
	enc1, err := cryptoutil.NewAESGCMEncryptor("key-one")
	require.NoError(t, err)
	enc2, err := cryptoutil.NewAESGCMEncryptor("key-two")
	require.NoError(t, err)

	cipherText, err := enc1.Encrypt("sensitive-data")
	require.NoError(t, err)

	_, err = enc2.Decrypt(cipherText)
	assert.Error(t, err, "错误密钥解密应返回错误")
}

// TestAESGCMEncryptor_TamperedCipher 密文被篡改应解密失败
func TestAESGCMEncryptor_TamperedCipher(t *testing.T) {
	enc, err := cryptoutil.NewAESGCMEncryptor("tamper-key")
	require.NoError(t, err)

	cipherText, err := enc.Encrypt("tamper-test")
	require.NoError(t, err)

	// 篡改密文尾部字符
	tampered := cipherText[:len(cipherText)-2] + "AA"
	_, err = enc.Decrypt(tampered)
	assert.Error(t, err, "篡改后的密文应解密失败")
}
