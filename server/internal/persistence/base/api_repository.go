package persistence

import (
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
	AuthRequired *bool  `gorm:"not null;default:false"`
	AuthType     *int   `gorm:"not null;default:0"`
	Description  string `gorm:"size:255"`
	Status       *int   `gorm:"not null;default:1"`
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

	api := &entity.API{
		ID:          m.ID,
		Path:        m.Path,
		Method:      m.Method,
		Group:       m.Group,
		Description: m.Description,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   m.UpdatedBy,
		UpdatedAt:   m.UpdatedAt,
	}

	// 安全解引用，提供默认值
	if m.AuthRequired != nil {
		api.AuthRequired = *m.AuthRequired
	} else {
		api.AuthRequired = false
	}

	if m.AuthType != nil {
		api.AuthType = *m.AuthType
	} else {
		api.AuthType = 0 // 默认值为 0（不需要授权）
	}

	if m.Status != nil {
		api.Status = *m.Status
	} else {
		api.Status = 1 // 默认值为 1（启用）
	}

	return api
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainAPI(api *entity.API) (*apiModel, error) {
	return &apiModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        api.ID,
			CreatedBy: api.CreatedBy,
			UpdatedBy: api.UpdatedBy,
		},
		Path:         api.Path,
		Method:       api.Method,
		Group:        api.Group,
		AuthRequired: common.BoolPtr(api.AuthRequired),
		AuthType:     common.IntPtr(api.AuthType),
		Description:  api.Description,
		Status:       common.IntPtr(api.Status),
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
func (r *apiRepository) GetByID(id int64) (*entity.API, error) {
	var dbModel apiModel
	result := r.db.Where("id = ?", id).First(&dbModel)
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

	// 使用 Create 直接插入，指针类型会自动包含零值
	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	route.ID = dbModel.ID
	return nil
}

// Update 更新 API 路由
func (r *apiRepository) Update(route *entity.API) error {
	dbModel, err := fromDomainAPI(route)
	if err != nil {
		return err
	}

	// 使用 Save 更新，指针类型会自动包含零值
	return r.db.Save(dbModel).Error
}

// Delete 删除 API 路由
func (r *apiRepository) Delete(id int64) error {
	return r.db.Delete(&apiModel{}, id).Error
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
func (r *apiRepository) GetAPIsByRole(roleID int64) ([]*entity.API, error) {
	// TODO: 实现角色关联 API 路由查询
	return nil, nil
}
