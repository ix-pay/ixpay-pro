package persistence

import (
	"strings"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/cryptoutil"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// configModel 配置数据库模型
type configModel struct {
	database.SnowflakeBaseModel
	ConfigKey   string `gorm:"size:100;not null"`
	ConfigValue string `gorm:"type:text"`
	ConfigType  *int   `gorm:"not null;default:1"`
	Description string `gorm:"size:255"`
	Encrypted   bool   `gorm:"not null;default:false"` // 敏感配置加密存储标记
	Status      *int   `gorm:"not null;default:1"`
}

// TableName 指定表名
func (configModel) TableName() string {
	return "base_configs"
}

// toDomain 将数据库模型转换为领域实体
func (m *configModel) toDomain() *entity.Config {
	if m == nil {
		return nil
	}
	config := &entity.Config{
		ID:          m.ID,
		ConfigKey:   m.ConfigKey,
		ConfigValue: m.ConfigValue,
		Description: m.Description,
		Encrypted:   m.Encrypted,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   m.UpdatedBy,
		UpdatedAt:   m.UpdatedAt,
	}

	// 安全解引用，提供默认值
	if m.ConfigType != nil {
		config.ConfigType = *m.ConfigType
	} else {
		config.ConfigType = 1
	}

	if m.Status != nil {
		config.Status = *m.Status
	} else {
		config.Status = 1
	}

	return config
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainConfig(config *entity.Config) (*configModel, error) {
	return &configModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        config.ID,
			CreatedBy: config.CreatedBy,
			UpdatedBy: config.UpdatedBy,
		},
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
		ConfigType:  common.IntPtr(config.ConfigType),
		Description: config.Description,
		Encrypted:   config.Encrypted,
		Status:      common.IntPtr(config.Status),
	}, nil
}

// sensitiveConfigKeys 敏感配置键集合（AES-GCM 加密存储，读取自动解密）
var sensitiveConfigKeys = map[string]bool{
	"wechat_app_secret":         true, // 微信公众号 AppSecret
	"wechat_miniapp_app_secret": true, // 微信小程序 AppSecret
	"wechat_apiv3_key":          true, // 微信支付 APIv3 密钥
	"wechat_private_key":        true, // 商户私钥 PEM 内容
	"wechat_med_partner_secret": true, // 医保授权查询密钥
}

// configRepository Repository 实现
type configRepository struct {
	db        *database.PostgresDB
	encryptor *cryptoutil.AESGCMEncryptor
}

// 确保实现接口
var _ repo.ConfigRepository = (*configRepository)(nil)

// NewConfigRepository 创建配置仓库实现
// 依据 config.Security.ConfigEncryptKey 构建加密器；密钥为空时不加密（明文兼容）
func NewConfigRepository(db *database.PostgresDB, cfg *config.Config) (repo.ConfigRepository, error) {
	encryptor, err := cryptoutil.NewAESGCMEncryptor(cfg.Security.ConfigEncryptKey)
	if err != nil {
		return nil, err
	}
	return &configRepository{db: db, encryptor: encryptor}, nil
}

// encryptIfNeeded 敏感配置加密（写入前调用）
func (r *configRepository) encryptIfNeeded(c *entity.Config) error {
	if r.encryptor == nil || !sensitiveConfigKeys[c.ConfigKey] || c.ConfigValue == "" {
		return nil
	}
	enc, err := r.encryptor.Encrypt(c.ConfigValue)
	if err != nil {
		return err
	}
	c.ConfigValue = enc
	c.Encrypted = true
	return nil
}

// decryptIfNeeded 敏感配置解密（读取后调用；解密失败保留原文并交由上层处理）
func (r *configRepository) decryptIfNeeded(c *entity.Config) {
	if r.encryptor == nil || c == nil || !c.Encrypted || c.ConfigValue == "" {
		return
	}
	plain, err := r.encryptor.Decrypt(c.ConfigValue)
	if err != nil {
		// 解密失败不阻断（密钥可能更换），保留密文原样返回
		return
	}
	c.ConfigValue = plain
}

// GetByID 根据 ID 查询配置
func (r *configRepository) GetByID(id int64) (*entity.Config, error) {
	var dbModel configModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	cfg := dbModel.toDomain()
	r.decryptIfNeeded(cfg)
	return cfg, nil
}

// GetByKey 根据键查询配置
func (r *configRepository) GetByKey(configKey string) (*entity.Config, error) {
	var dbModel configModel
	result := r.db.Where("config_key = ?", configKey).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	cfg := dbModel.toDomain()
	r.decryptIfNeeded(cfg)
	return cfg, nil
}

// Create 创建配置
func (r *configRepository) Create(config *entity.Config) error {
	if err := r.encryptIfNeeded(config); err != nil {
		return err
	}
	dbModel, err := fromDomainConfig(config)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	config.ID = dbModel.ID
	return nil
}

// Update 更新配置
func (r *configRepository) Update(config *entity.Config) error {
	if err := r.encryptIfNeeded(config); err != nil {
		return err
	}
	dbModel, err := fromDomainConfig(config)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除配置
func (r *configRepository) Delete(id int64) error {
	return r.db.Delete(&configModel{}, id).Error
}

// List 分页查询配置列表
func (r *configRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Config, int64, error) {
	var total int64
	var dbModels []configModel

	query := r.db.Model(&configModel{})

	// 应用过滤条件
	for key, value := range filters {
		if strings.Contains(key, "?") {
			// key 已包含操作符（如 "name LIKE ?"），直接使用
			if values, ok := value.([]string); ok {
				args := make([]interface{}, len(values))
				for i, v := range values {
					args[i] = v
				}
				query = query.Where(key, args...)
			} else {
				query = query.Where(key, value)
			}
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("config_key ASC, id ASC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	configs := make([]*entity.Config, len(dbModels))
	for i, model := range dbModels {
		cfg := model.toDomain()
		r.decryptIfNeeded(cfg)
		configs[i] = cfg
	}

	return configs, total, nil
}

// GetAllActive 获取所有启用的配置
func (r *configRepository) GetAllActive() ([]*entity.Config, error) {
	var dbModels []configModel
	result := r.db.Where("status = ?", 1).Order("config_key ASC, id ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	configs := make([]*entity.Config, len(dbModels))
	for i, model := range dbModels {
		cfg := model.toDomain()
		r.decryptIfNeeded(cfg)
		configs[i] = cfg
	}

	return configs, nil
}
