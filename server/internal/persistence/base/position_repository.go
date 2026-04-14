package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// positionModel 岗位数据库模型
type positionModel struct {
	database.SnowflakeBaseModel
	Name        string `gorm:"size:50;not null"`
	Sort        int    `gorm:"default:0"`
	Status      int    `gorm:"default:1"`
	Description string `gorm:"size:255"`
}

// TableName 指定表名
func (positionModel) TableName() string {
	return "base_positions"
}

// toDomain 将数据库模型转换为领域实体
func (m *positionModel) toDomain() *entity.Position {
	if m == nil {
		return nil
	}
	return &entity.Position{
		ID:          m.ID,
		Name:        m.Name,
		Sort:        m.Sort,
		Status:      m.Status,
		Description: m.Description,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   m.UpdatedBy,
		UpdatedAt:   m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainPosition(pos *entity.Position) (*positionModel, error) {
	return &positionModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        pos.ID,
			CreatedBy: pos.CreatedBy,
			UpdatedBy: pos.UpdatedBy,
		},
		Name:        pos.Name,
		Sort:        pos.Sort,
		Status:      pos.Status,
		Description: pos.Description,
	}, nil
}

// positionRepository Repository 实现
type positionRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.PositionRepository = (*positionRepository)(nil)

// NewPositionRepository 创建岗位仓库实现
func NewPositionRepository(db *database.PostgresDB) repo.PositionRepository {
	return &positionRepository{db: db}
}

// GetByID 根据 ID 查询岗位
func (r *positionRepository) GetByID(id int64) (*entity.Position, error) {
	var dbModel positionModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建岗位
func (r *positionRepository) Create(position *entity.Position) error {
	dbModel, err := fromDomainPosition(position)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	position.ID = dbModel.ID
	return nil
}

// Update 更新岗位
func (r *positionRepository) Update(position *entity.Position) error {
	dbModel, err := fromDomainPosition(position)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除岗位
func (r *positionRepository) Delete(id int64) error {
	return r.db.Delete(&positionModel{}, id).Error
}

// List 分页查询岗位列表
func (r *positionRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Position, int64, error) {
	var total int64
	var dbModels []positionModel

	query := r.db.Model(&positionModel{})

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

	positions := make([]*entity.Position, len(dbModels))
	for i, model := range dbModels {
		positions[i] = model.toDomain()
	}

	return positions, total, nil
}

// GetAll 获取所有岗位
func (r *positionRepository) GetAll() ([]*entity.Position, error) {
	var dbModels []positionModel
	result := r.db.Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	positions := make([]*entity.Position, len(dbModels))
	for i, model := range dbModels {
		positions[i] = model.toDomain()
	}

	return positions, nil
}

// GetByName 根据名称查询岗位
func (r *positionRepository) GetByName(name string) (*entity.Position, error) {
	var dbModel positionModel
	result := r.db.Where("name = ?", name).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}
