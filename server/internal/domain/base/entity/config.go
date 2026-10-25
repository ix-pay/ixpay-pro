package entity

import "time"

// Config 系统配置领域实体
// 纯业务模型，无 GORM 标签
type Config struct {
	ID          int64     // 配置 ID
	ConfigKey   string    // 配置键
	ConfigValue string    // 配置值（敏感配置加密存储，读写自动加解密，实体层为明文）
	ConfigType  int       // 配置类型：1-系统配置，2-业务配置
	Description string    // 配置描述
	Encrypted   bool      // 是否加密存储（敏感配置为 true，数据库中为密文）
	Status      int       // 状态：1-启用，0-禁用
	CreatedBy   int64     // 创建人 ID
	CreatedAt   time.Time // 创建时间
	UpdatedBy   int64     // 更新人 ID
	UpdatedAt   time.Time // 更新时间
}

// IsActive 检查配置是否启用
func (c *Config) IsActive() bool {
	return c.Status == 1
}
