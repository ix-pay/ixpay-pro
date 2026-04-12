package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// dictModel 字典数据库模型
type dictModel struct {
	database.SnowflakeBaseModel
	DictName    string `gorm:"size:100;not null"`
	DictCode    string `gorm:"size:50;not null"`
	Description string `gorm:"size:255"`
	Status      int    `gorm:"default:1"`
}

// TableName 指定表名
func (dictModel) TableName() string {
	return "base_dicts"
}

// toDomain 将数据库模型转换为领域实体
func (m *dictModel) toDomain() *entity.Dict {
	if m == nil {
		return nil
	}
	return &entity.Dict{
		ID:          common.ToString(m.ID),
		DictName:    m.DictName,
		DictCode:    m.DictCode,
		Description: m.Description,
		Status:      m.Status,
		CreatedBy:   common.ToString(m.CreatedBy),
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   common.ToString(m.UpdatedBy),
		UpdatedAt:   m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainDict(dict *entity.Dict) (*dictModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(dict.ID, dict.CreatedBy, dict.UpdatedBy)

	return &dictModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		DictName:    dict.DictName,
		DictCode:    dict.DictCode,
		Description: dict.Description,
		Status:      dict.Status,
	}, nil
}

// dictRepository Repository 实现
type dictRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.DictRepository = (*dictRepository)(nil)

// NewDictRepository 创建字典仓库实现
func NewDictRepository(db *database.PostgresDB) repo.DictRepository {
	return &dictRepository{db: db}
}

// GetByID 根据 ID 查询字典
func (r *dictRepository) GetByID(id string) (*entity.Dict, error) {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}

	var dbModel dictModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByCode 根据编码查询字典
func (r *dictRepository) GetByCode(dictCode string) (*entity.Dict, error) {
	var dbModel dictModel
	result := r.db.Where("dict_code = ?", dictCode).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建字典
func (r *dictRepository) Create(dict *entity.Dict) error {
	dbModel, err := fromDomainDict(dict)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
}

// Update 更新字典
func (r *dictRepository) Update(dict *entity.Dict) error {
	dbModel, err := fromDomainDict(dict)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除字典
func (r *dictRepository) Delete(id string) error {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&dictModel{}, intID).Error
}

// List 分页查询字典列表
func (r *dictRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Dict, int64, error) {
	var total int64
	var dbModels []dictModel

	query := r.db.Model(&dictModel{})

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

	dicts := make([]*entity.Dict, len(dbModels))
	for i, model := range dbModels {
		dicts[i] = model.toDomain()
	}

	return dicts, total, nil
}

// GetAllActive 获取所有启用的字典
func (r *dictRepository) GetAllActive() ([]*entity.Dict, error) {
	var dbModels []dictModel
	result := r.db.Where("status = ?", 1).Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	dicts := make([]*entity.Dict, len(dbModels))
	for i, model := range dbModels {
		dicts[i] = model.toDomain()
	}

	return dicts, nil
}
