package persistence

import (
	"strconv"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// apiModel API 路由数据库模型
type apiModel struct {
	database.SnowflakeBaseModel
	Path         string `gorm:"size:255;not null"`
	Method       string `gorm:"size:20;not null"`
	Group        string `gorm:"size:50"`
	AuthRequired bool   `gorm:"default:false"`
	AuthType     int    `gorm:"default:1"`
	Description  string `gorm:"size:255"`
	Status       int    `gorm:"default:1"`
}

// TableName 指定表名
func (apiModel) TableName() string {
	return "base_apis"
}

// toDomain 将数据库模型转换为领域实体
func (m *apiModel) toDomain() *entity.API {
	if m == nil {
		return nil
	}
	return &entity.API{
		ID:           common.ToString(m.ID),
		Path:         m.Path,
		Method:       m.Method,
		Group:        m.Group,
		AuthRequired: m.AuthRequired,
		AuthType:     m.AuthType,
		Description:  m.Description,
		Status:       m.Status,
		CreatedBy:    common.ToString(m.CreatedBy),
		CreatedAt:    m.CreatedAt,
		UpdatedBy:    common.ToString(m.UpdatedBy),
		UpdatedAt:    m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainAPI(api *entity.API) (*apiModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(api.ID, api.CreatedBy, api.UpdatedBy)

	return &apiModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		Path:         api.Path,
		Method:       api.Method,
		Group:        api.Group,
		AuthRequired: api.AuthRequired,
		AuthType:     api.AuthType,
		Description:  api.Description,
		Status:       api.Status,
	}, nil
}

// apiRepository Repository 实现
type apiRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.APIRepository = (*apiRepository)(nil)

// NewAPIRepository 创建 API 路由仓库实现
func NewAPIRepository(db *database.PostgresDB) repo.APIRepository {
	return &apiRepository{db: db}
}

// BatchSave 批量保存 API 路由
func (r *apiRepository) BatchSave(routes []*entity.API) error {
	// TODO: 实现批量保存
	return nil
}

// GetByID 根据 ID 查询 API 路由
func (r *apiRepository) GetByID(id string) (*entity.API, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var dbModel apiModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetAllRoutes 获取所有 API 路由
func (r *apiRepository) GetAllRoutes() ([]*entity.API, error) {
	var dbModels []apiModel
	result := r.db.Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	apis := make([]*entity.API, len(dbModels))
	for i, model := range dbModels {
		apis[i] = model.toDomain()
	}

	return apis, nil
}

// GetByPathAndMethod 根据路径和方法查询 API 路由
func (r *apiRepository) GetByPathAndMethod(path, method string) (*entity.API, error) {
	var dbModel apiModel
	result := r.db.Where("path = ? AND method = ?", path, method).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建 API 路由
func (r *apiRepository) Create(route *entity.API) error {
	dbModel, err := fromDomainAPI(route)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
}

// Update 更新 API 路由
func (r *apiRepository) Update(route *entity.API) error {
	dbModel, err := fromDomainAPI(route)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除 API 路由
func (r *apiRepository) Delete(id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	return r.db.Delete(&apiModel{}, intID).Error
}

// List 分页查询 API 路由列表
func (r *apiRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.API, int64, error) {
	var total int64
	var dbModels []apiModel

	query := r.db.Model(&apiModel{})

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

	apis := make([]*entity.API, len(dbModels))
	for i, model := range dbModels {
		apis[i] = model.toDomain()
	}

	return apis, total, nil
}

// GetAPIsByRole 根据角色获取 API 路由
func (r *apiRepository) GetAPIsByRole(roleID string) ([]*entity.API, error) {
	// TODO: 实现角色关联 API 路由查询
	return nil, nil
}
