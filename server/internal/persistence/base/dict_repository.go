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
	Status      *int   `gorm:"not null;default:1"`

	// GORM 关联关系 - 一对多（字典项）
	DictItems []dictItemModel `gorm:"foreignKey:dict_id;references:id"`
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
	dict := &entity.Dict{
		ID:          m.ID,
		DictName:    m.DictName,
		DictCode:    m.DictCode,
		Description: m.Description,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   m.UpdatedBy,
		UpdatedAt:   m.UpdatedAt,
	}

	// 安全解引用，提供默认值
	if m.Status != nil {
		dict.Status = *m.Status
	} else {
		dict.Status = 1
	}

	// ⭐ 处理关联数据 - 字典项
	if len(m.DictItems) > 0 {
		dictItems := make([]entity.DictItem, len(m.DictItems))
		for i, item := range m.DictItems {
			dictItems[i] = *item.toDomain()
		}
		dict.DictItems = dictItems
	}

	return dict
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainDict(dict *entity.Dict) (*dictModel, error) {
	return &dictModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        dict.ID,
			CreatedBy: dict.CreatedBy,
			UpdatedBy: dict.UpdatedBy,
		},
		DictName:    dict.DictName,
		DictCode:    dict.DictCode,
		Description: dict.Description,
		Status:      common.IntPtr(dict.Status),
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

// GetByID 根据 ID 查询字典并支持加载关联数据
func (r *dictRepository) GetByID(id int64, relations ...repo.DictRelation) (*entity.Dict, error) {
	var dbModel dictModel
	query := r.db.Where("id = ?", id)

	// 根据指定的关联关系进行 Preload
	for _, relation := range relations {
		query = query.Preload(string(relation))
	}

	result := query.First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByCode 根据编码查询字典并支持加载关联数据
func (r *dictRepository) GetByCode(dictCode string, relations ...repo.DictRelation) (*entity.Dict, error) {
	var dbModel dictModel
	query := r.db.Where("dict_code = ?", dictCode)

	// 根据指定的关联关系进行 Preload
	for _, relation := range relations {
		query = query.Preload(string(relation))
	}

	result := query.First(&dbModel)
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

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	dict.ID = dbModel.ID
	return nil
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
func (r *dictRepository) Delete(id int64) error {
	return r.db.Delete(&dictModel{}, id).Error
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
