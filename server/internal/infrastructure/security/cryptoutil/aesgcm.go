// Package cryptoutil 通用加解密工具
// 提供 AES-256-GCM 加解密能力，用于数据库敏感配置（如微信 AppSecret、APIv3 密钥、商户私钥 PEM）的加密存储
package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// 加密前缀标识（区分密文与明文，兼容历史明文数据）
const cipherPrefix = "enc:v1:"

// AESGCMEncryptor AES-256-GCM 加解密器
// 密钥由原始字符串经 SHA-256 派生为 32 字节，任意长度配置密钥均可使用
type AESGCMEncryptor struct {
	gcm cipher.AEAD
}

// NewAESGCMEncryptor 创建加解密器；key 为空时返回 nil（表示未启用加密，读写按明文处理）
func NewAESGCMEncryptor(key string) (*AESGCMEncryptor, error) {
	if strings.TrimSpace(key) == "" {
		return nil, nil
	}
	// SHA-256 派生 32 字节密钥（AES-256）
	keyBytes := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return nil, fmt.Errorf("初始化AES加密失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("初始化GCM失败: %w", err)
	}
	return &AESGCMEncryptor{gcm: gcm}, nil
}

// Encrypt 加密明文，返回 "enc:v1:" 前缀 + base64(nonce + ciphertext)
func (e *AESGCMEncryptor) Encrypt(plain string) (string, error) {
	if e == nil {
		return plain, nil
	}
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("生成随机nonce失败: %w", err)
	}
	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plain), nil)
	return cipherPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密密文（自动识别前缀；无前缀的历史明文原样返回）
func (e *AESGCMEncryptor) Decrypt(value string) (string, error) {
	if e == nil || !strings.HasPrefix(value, cipherPrefix) {
		return value, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, cipherPrefix))
	if err != nil {
		return "", fmt.Errorf("密文base64解码失败: %w", err)
	}
	if len(raw) < e.gcm.NonceSize() {
		return "", errors.New("密文长度非法")
	}
	nonce, ciphertext := raw[:e.gcm.NonceSize()], raw[e.gcm.NonceSize():]
	plain, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败（加密密钥可能不匹配）: %w", err)
	}
	return string(plain), nil
}
