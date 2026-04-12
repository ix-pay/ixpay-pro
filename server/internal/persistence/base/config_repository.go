package persistence

import (
	"strconv"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// configModel 配置数据库模型
type configModel struct {
	database.SnowflakeBaseModel
	ConfigKey   string `gorm:"size:100;not null"`
	ConfigValue string `gorm:"type:text"`
	ConfigType  string `gorm:"size:20"`
	Description string `gorm:"size:255"`
	Status      int    `gorm:"default:1"`
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
	return &entity.Config{
		ID:          strconv.FormatInt(m.ID, 10),
		ConfigKey:   m.ConfigKey,
		ConfigValue: m.ConfigValue,
		ConfigType:  m.ConfigType,
		Description: m.Description,
		Status:      m.Status,
		CreatedBy:   strconv.FormatInt(m.CreatedBy, 10),
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   strconv.FormatInt(m.UpdatedBy, 10),
		UpdatedAt:   m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainConfig(config *entity.Config) (*configModel, error) {
	var id int64
	var err error

	if config.ID != "" {
		id, err = strconv.ParseInt(config.ID, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	var createdBy int64
	if config.CreatedBy != "" {
		createdBy, _ = strconv.ParseInt(config.CreatedBy, 10, 64)
	}

	var updatedBy int64
	if config.UpdatedBy != "" {
		updatedBy, _ = strconv.ParseInt(config.UpdatedBy, 10, 64)
	}

	return &configModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
		ConfigType:  config.ConfigType,
		Description: config.Description,
		Status:      config.Status,
	}, nil
}

// configRepository Repository 实现
type configRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.ConfigRepository = (*configRepository)(nil)

// NewConfigRepository 创建配置仓库实现
func NewConfigRepository(db *database.PostgresDB) repo.ConfigRepository {
	return &configRepository{db: db}
}

// GetByID 根据 ID 查询配置
func (r *configRepository) GetByID(id string) (*entity.Config, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var dbModel configModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByKey 根据键查询配置
func (r *configRepository) GetByKey(configKey string) (*entity.Config, error) {
	var dbModel configModel
	result := r.db.Where("config_key = ?", configKey).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建配置
func (r *configRepository) Create(config *entity.Config) error {
	dbModel, err := fromDomainConfig(config)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
}

// Update 更新配置
func (r *configRepository) Update(config *entity.Config) error {
	dbModel, err := fromDomainConfig(config)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除配置
func (r *configRepository) Delete(id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	return r.db.Delete(&configModel{}, intID).Error
}

// List 分页查询配置列表
func (r *configRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Config, int64, error) {
	var total int64
	var dbModels []configModel

	query := r.db.Model(&configModel{})

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	configs := make([]*entity.Config, len(dbModels))
	for i, model := range dbModels {
		configs[i] = model.toDomain()
	}

	return configs, total, nil
}

// GetAllActive 获取所有启用的配置
func (r *configRepository) GetAllActive() ([]*entity.Config, error) {
	var dbModels []configModel
	result := r.db.Where("status = ?", 1).Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	configs := make([]*entity.Config, len(dbModels))
	for i, model := range dbModels {
		configs[i] = model.toDomain()
	}

	return configs, nil
}
