package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
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
		ID:        common.ToString(m.ID),
		DictID:    common.ToString(m.DictID),
		ItemKey:   m.ItemKey,
		ItemValue: m.ItemValue,
		Sort:      m.Sort,
		Status:    m.Status,
		CreatedBy: common.ToString(m.CreatedBy),
		CreatedAt: m.CreatedAt,
		UpdatedBy: common.ToString(m.UpdatedBy),
		UpdatedAt: m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainDictItem(dictItem *entity.DictItem) (*dictItemModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(dictItem.ID, dictItem.CreatedBy, dictItem.UpdatedBy)

	return &dictItemModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		DictID:    common.TryParseInt64(dictItem.DictID),
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
func (r *dictItemRepository) GetByID(id string) (*entity.DictItem, error) {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}

	var dbModel dictItemModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByDictID 根据字典 ID 查询字典项
func (r *dictItemRepository) GetByDictID(dictID string) ([]*entity.DictItem, error) {
	intDictID := common.TryParseInt64(dictID)
	var dbModels []dictItemModel
	result := r.db.Where("dict_id = ?", intDictID).Order("sort ASC").Find(&dbModels)
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
	dictItem.ID = common.ToString(dbModel.ID)
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
func (r *dictItemRepository) Delete(id string) error {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&dictItemModel{}, intID).Error
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
func (r *dictItemRepository) GetActiveByDictID(dictID string) ([]*entity.DictItem, error) {
	intDictID := common.TryParseInt64(dictID)
	var dbModels []dictItemModel
	result := r.db.Where("dict_id = ? AND status = ?", intDictID, 1).Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	dictItems := make([]*entity.DictItem, len(dbModels))
	for i, model := range dbModels {
		dictItems[i] = model.toDomain()
	}

	return dictItems, nil
}
