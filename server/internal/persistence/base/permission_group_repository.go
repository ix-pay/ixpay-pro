package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// permissionGroupModel 权限组数据库模型
type permissionGroupModel struct {
	database.SnowflakeBaseModel
	Name        string `gorm:"size:100;not null;unique"`
	Description string `gorm:"size:500"`
	Status      int    `gorm:"default:1"`
	Sort        int    `gorm:"default:0"`
}

// TableName 指定表名
func (permissionGroupModel) TableName() string {
	return "base_permission_groups"
}

// toDomain 将数据库模型转换为领域实体
func (m *permissionGroupModel) toDomain() *entity.PermissionGroup {
	if m == nil {
		return nil
	}
	return &entity.PermissionGroup{
		ID:          common.ToString(m.ID),
		Name:        m.Name,
		Description: m.Description,
		Status:      m.Status,
		Sort:        m.Sort,
		CreatedBy:   common.ToString(m.CreatedBy),
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   common.ToString(m.UpdatedBy),
		UpdatedAt:   m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainPermissionGroup(group *entity.PermissionGroup) (*permissionGroupModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(group.ID, group.CreatedBy, group.UpdatedBy)

	return &permissionGroupModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		Name:        group.Name,
		Description: group.Description,
		Status:      group.Status,
		Sort:        group.Sort,
	}, nil
}

// permissionGroupRepository Repository 实现
type permissionGroupRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.PermissionGroupRepository = (*permissionGroupRepository)(nil)

// NewPermissionGroupRepository 创建权限组仓库实现
func NewPermissionGroupRepository(db *database.PostgresDB) repo.PermissionGroupRepository {
	return &permissionGroupRepository{db: db}
}

// GetByID 根据 ID 查询权限组
func (r *permissionGroupRepository) GetByID(id string) (*entity.PermissionGroup, error) {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}

	var dbModel permissionGroupModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByName 根据名称查询权限组
func (r *permissionGroupRepository) GetByName(name string) (*entity.PermissionGroup, error) {
	var dbModel permissionGroupModel
	result := r.db.Where("name = ?", name).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建权限组
func (r *permissionGroupRepository) Create(group *entity.PermissionGroup) error {
	dbModel, err := fromDomainPermissionGroup(group)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	group.ID = common.ToString(dbModel.ID)
	return nil
}

// Update 更新权限组
func (r *permissionGroupRepository) Update(group *entity.PermissionGroup) error {
	dbModel, err := fromDomainPermissionGroup(group)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除权限组
func (r *permissionGroupRepository) Delete(id string) error {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&permissionGroupModel{}, intID).Error
}

// List 分页查询权限组列表
func (r *permissionGroupRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.PermissionGroup, int64, error) {
	var total int64
	var dbModels []permissionGroupModel

	query := r.db.Model(&permissionGroupModel{})

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	groups := make([]*entity.PermissionGroup, len(dbModels))
	for i, model := range dbModels {
		groups[i] = model.toDomain()
	}

	return groups, total, nil
}

// GetAllGroups 获取所有权限组
func (r *permissionGroupRepository) GetAllGroups() ([]*entity.PermissionGroup, error) {
	var dbModels []permissionGroupModel
	result := r.db.Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	groups := make([]*entity.PermissionGroup, len(dbModels))
	for i, model := range dbModels {
		groups[i] = model.toDomain()
	}

	return groups, nil
}

// AddAPIToGroup 添加 API 路由到权限组
func (r *permissionGroupRepository) AddAPIToGroup(groupID, apiID string) error {
	// TODO: 实现权限组关联 API 路由表操作
	return nil
}

// RemoveAPIFromGroup 从权限组移除 API 路由
func (r *permissionGroupRepository) RemoveAPIFromGroup(groupID, apiID string) error {
	// TODO: 实现权限组关联 API 路由表操作
	return nil
}

// GetAPIsByGroup 获取权限组下的所有 API 路由
func (r *permissionGroupRepository) GetAPIsByGroup(groupID string) ([]*entity.API, error) {
	// TODO: 实现权限组关联 API 路由表操作
	return nil, nil
}

// GetGroupsByAPI 获取 API 路由的所有权限组
func (r *permissionGroupRepository) GetGroupsByAPI(apiID string) ([]*entity.PermissionGroup, error) {
	// TODO: 实现权限组关联 API 路由表操作
	return nil, nil
}

// AddBtnPermToGroup 添加按钮权限到权限组
func (r *permissionGroupRepository) AddBtnPermToGroup(groupID, btnPermID string) error {
	// TODO: 实现权限组关联按钮权限表操作
	return nil
}

// RemoveBtnPermFromGroup 从权限组移除按钮权限
func (r *permissionGroupRepository) RemoveBtnPermFromGroup(groupID, btnPermID string) error {
	// TODO: 实现权限组关联按钮权限表操作
	return nil
}

// GetBtnPermsByGroup 获取权限组下的所有按钮权限
func (r *permissionGroupRepository) GetBtnPermsByGroup(groupID string) ([]*entity.BtnPerm, error) {
	// TODO: 实现权限组关联按钮权限表操作
	return nil, nil
}

// GetGroupsByBtnPerm 获取按钮权限的所有权限组
func (r *permissionGroupRepository) GetGroupsByBtnPerm(btnPermID string) ([]*entity.PermissionGroup, error) {
	// TODO: 实现权限组关联按钮权限表操作
	return nil, nil
}

// GetRolesByGroup 获取权限组的所有角色
func (r *permissionGroupRepository) GetRolesByGroup(groupID string) ([]*entity.Role, error) {
	// TODO: 实现权限组关联角色表操作
	return nil, nil
}

// GetGroupsByRole 获取角色的所有权限组
func (r *permissionGroupRepository) GetGroupsByRole(roleID string) ([]*entity.PermissionGroup, error) {
	// TODO: 实现权限组关联角色表操作
	return nil, nil
}
