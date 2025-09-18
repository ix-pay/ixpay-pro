package encryption

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params Argon2参数
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// GeneratePasswordHash 生成密码哈希
func GeneratePasswordHash(password string) (string, error) {
	// 生成随机salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 配置参数
	params := &Argon2Params{
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	// 生成哈希
	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// 编码结果
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 格式化结果
	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.Memory, params.Iterations, params.Parallelism,
		b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyPassword 验证密码
func VerifyPassword(encodedHash, password string) error {
	// 解析编码后的哈希
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return err
	}

	// 生成输入密码的哈希
	otherHash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// 比较哈希
	if subtle.ConstantTimeCompare(hash, otherHash) != 1 {
		return errors.New("invalid password")
	}

	return nil
}

// GenerateRandomPasswordHash 生成随机密码哈希
func GenerateRandomPasswordHash() string {
	// 生成随机字节
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		// 如果随机数生成失败，使用默认值
		return "$argon2id$v=19$m=65536,t=1,p=2$0000000000000000$00000000000000000000000000000000"
	}

	// 编码为Base64
	randomStr := base64.RawStdEncoding.EncodeToString(randomBytes)

	// 生成哈希
	hash, err := GeneratePasswordHash(randomStr)
	if err != nil {
		// 如果哈希生成失败，使用默认值
		return "$argon2id$v=19$m=65536,t=1,p=2$0000000000000000$00000000000000000000000000000000"
	}

	return hash
}

// decodeHash 解析编码后的哈希
func decodeHash(encodedHash string) (p *Argon2Params, salt, hash []byte, err error) {
	p = &Argon2Params{
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	// 检查哈希格式
	const argon2Prefix = "$argon2id$v=19$"
	if !strings.HasPrefix(encodedHash, argon2Prefix) {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	// 解析参数部分
	rest := encodedHash[len(argon2Prefix):]
	parts := strings.Split(rest, "$")
	if len(parts) < 3 {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	// 解析内存、迭代次数和并行度参数
	paramsPart := parts[0]
	paramParts := strings.Split(paramsPart, ",")
	for _, param := range paramParts {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]
		switch key {
		case "m":
			fmt.Sscanf(value, "%d", &p.Memory)
		case "t":
			fmt.Sscanf(value, "%d", &p.Iterations)
		case "p":
			var parallelism uint32
			fmt.Sscanf(value, "%d", &parallelism)
			p.Parallelism = uint8(parallelism)
		}
	}

	// 解析salt
	saltStr := parts[1]
	salt, err = base64.RawStdEncoding.DecodeString(saltStr)
	if err != nil {
		return nil, nil, nil, errors.New("invalid salt format")
	}
	p.SaltLength = uint32(len(salt))

	// 解析hash
	hashStr := parts[2]
	hash, err = base64.RawStdEncoding.DecodeString(hashStr)
	if err != nil {
		return nil, nil, nil, errors.New("invalid hash format")
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
