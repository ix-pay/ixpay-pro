package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// dictItemModel 字典项数据库模型
type dictItemModel struct {
	database.SnowflakeBaseModel
	DictID      int64  `gorm:"index"`
	ItemKey     string `gorm:"size:50;not null"`
	ItemValue   string `gorm:"size:255"`
	Sort        int    `gorm:"default:0"`
	Description string `gorm:"size:255"`
	Status      int    `gorm:"default:1"`
}

// TableName 指定表名
func (dictItemModel) TableName() string {
	return "base_dict_items"
}

// toDomain 将数据库模型转换为领域实体
func (m *dictItemModel) toDomain() *entity.DictItem {
	if m == nil {
		return nil
	}
	return &entity.DictItem{
		ID:        m.ID,
		DictID:    m.DictID,
		ItemKey:   m.ItemKey,
		ItemValue: m.ItemValue,
		Sort:      m.Sort,
		Status:    m.Status,
		CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt,
		UpdatedBy: m.UpdatedBy,
		UpdatedAt: m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainDictItem(dictItem *entity.DictItem) (*dictItemModel, error) {
	return &dictItemModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        dictItem.ID,
			CreatedBy: dictItem.CreatedBy,
			UpdatedBy: dictItem.UpdatedBy,
		},
		DictID:    dictItem.DictID,
		ItemKey:   dictItem.ItemKey,
		ItemValue: dictItem.ItemValue,
		Sort:      dictItem.Sort,
		Status:    dictItem.Status,
	}, nil
}

// dictItemRepository Repository 实现
type dictItemRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.DictItemRepository = (*dictItemRepository)(nil)

// NewDictItemRepository 创建字典项仓库实现
func NewDictItemRepository(db *database.PostgresDB) repo.DictItemRepository {
	return &dictItemRepository{db: db}
}

// GetByID 根据 ID 查询字典项
func (r *dictItemRepository) GetByID(id int64) (*entity.DictItem, error) {
	var dbModel dictItemModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByDictID 根据字典 ID 查询字典项
func (r *dictItemRepository) GetByDictID(dictID int64) ([]*entity.DictItem, error) {
	var dbModels []dictItemModel
	result := r.db.Where("dict_id = ?", dictID).Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	dictItems := make([]*entity.DictItem, len(dbModels))
	for i, model := range dbModels {
		dictItems[i] = model.toDomain()
	}

	return dictItems, nil
}

// Create 创建字典项
func (r *dictItemRepository) Create(dictItem *entity.DictItem) error {
	dbModel, err := fromDomainDictItem(dictItem)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	dictItem.ID = dbModel.ID
	return nil
}

// Update 更新字典项
func (r *dictItemRepository) Update(dictItem *entity.DictItem) error {
	dbModel, err := fromDomainDictItem(dictItem)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除字典项
func (r *dictItemRepository) Delete(id int64) error {
	return r.db.Delete(&dictItemModel{}, id).Error
}

// List 分页查询字典项列表
func (r *dictItemRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.DictItem, int64, error) {
	var total int64
	var dbModels []dictItemModel

	query := r.db.Model(&dictItemModel{})

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

	dictItems := make([]*entity.DictItem, len(dbModels))
	for i, model := range dbModels {
		dictItems[i] = model.toDomain()
	}

	return dictItems, total, nil
}

// GetActiveByDictID 根据字典 ID 获取启用的字典项
func (r *dictItemRepository) GetActiveByDictID(dictID int64) ([]*entity.DictItem, error) {
	var dbModels []dictItemModel
	result := r.db.Where("dict_id = ? AND status = ?", dictID, 1).Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	dictItems := make([]*entity.DictItem, len(dbModels))
	for i, model := range dbModels {
		dictItems[i] = model.toDomain()
	}

	return dictItems, nil
}
