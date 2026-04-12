package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// roleModel 角色数据库模型
type roleModel struct {
	database.SnowflakeBaseModel
	Name        string `gorm:"size:50;not null;unique"`
	Code        string `gorm:"size:50;not null;unique"`
	Description string `gorm:"size:255"`
	Type        int    `gorm:"default:1"`
	ParentID    int64  `gorm:"default:0"`
	Status      int    `gorm:"default:1"`
	IsSystem    bool   `gorm:"default:false"`
	Sort        int    `gorm:"default:0"`
}

// TableName 指定表名
func (roleModel) TableName() string {
	return "base_roles"
}

// toDomain 将数据库模型转换为领域实体
func (m *roleModel) toDomain() *entity.Role {
	if m == nil {
		return nil
	}
	return &entity.Role{
		ID:          common.ToString(m.ID),
		Name:        m.Name,
		Code:        m.Code,
		Description: m.Description,
		Type:        m.Type,
		ParentID:    common.ToString(m.ParentID),
		Status:      m.Status,
		IsSystem:    m.IsSystem,
		Sort:        m.Sort,
		CreatedBy:   common.ToString(m.CreatedBy),
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   common.ToString(m.UpdatedBy),
		UpdatedAt:   m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainRole(role *entity.Role) (*roleModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(role.ID, role.CreatedBy, role.UpdatedBy)

	return &roleModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Type:        role.Type,
		ParentID:    common.TryParseInt64(role.ParentID),
		Status:      role.Status,
		IsSystem:    role.IsSystem,
		Sort:        role.Sort,
	}, nil
}

// roleRepository Repository 实现
type roleRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.RoleRepository = (*roleRepository)(nil)

// NewRoleRepository 创建角色仓库实现
func NewRoleRepository(db *database.PostgresDB) repo.RoleRepository {
	return &roleRepository{db: db}
}

// GetByID 根据 ID 查询角色
func (r *roleRepository) GetByID(id string) (*entity.Role, error) {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}

	var dbModel roleModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByName 根据名称查询角色
func (r *roleRepository) GetByName(name string) (*entity.Role, error) {
	var dbModel roleModel
	result := r.db.Where("name = ?", name).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByCode 根据编码查询角色
func (r *roleRepository) GetByCode(code string) (*entity.Role, error) {
	var dbModel roleModel
	result := r.db.Where("code = ?", code).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建角色
func (r *roleRepository) Create(role *entity.Role) error {
	dbModel, err := fromDomainRole(role)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
}

// Update 更新角色
func (r *roleRepository) Update(role *entity.Role) error {
	dbModel, err := fromDomainRole(role)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除角色
func (r *roleRepository) Delete(id string) error {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&roleModel{}, intID).Error
}

// List 分页查询角色列表
func (r *roleRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Role, int64, error) {
	var total int64
	var dbModels []roleModel

	query := r.db.Model(&roleModel{})

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

	roles := make([]*entity.Role, len(dbModels))
	for i, model := range dbModels {
		roles[i] = model.toDomain()
	}

	return roles, total, nil
}

// GetAllRoles 获取所有角色
func (r *roleRepository) GetAllRoles() ([]*entity.Role, error) {
	var dbModels []roleModel
	result := r.db.Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	roles := make([]*entity.Role, len(dbModels))
	for i, model := range dbModels {
		roles[i] = model.toDomain()
	}

	return roles, nil
}

// AddUserToRole 添加用户到角色
func (r *roleRepository) AddUserToRole(roleID, userID string) error {
	// TODO: 实现角色关联用户表操作
	return nil
}

// RemoveUserFromRole 从角色移除用户
func (r *roleRepository) RemoveUserFromRole(roleID, userID string) error {
	// TODO: 实现角色关联用户表操作
	return nil
}

// ExistsUserInRole 检查用户是否在角色中
func (r *roleRepository) ExistsUserInRole(roleID, userID string) (bool, error) {
	// TODO: 实现角色关联用户表操作
	return false, nil
}

// GetUsersByRole 获取角色下的所有用户
func (r *roleRepository) GetUsersByRole(roleID string) ([]*entity.User, error) {
	// TODO: 实现角色关联用户表操作
	return nil, nil
}

// GetRolesByUser 获取用户的所有角色
func (r *roleRepository) GetRolesByUser(userID string) ([]*entity.Role, error) {
	// TODO: 实现角色关联用户表操作
	return nil, nil
}

// AddMenuToRole 添加菜单到角色
func (r *roleRepository) AddMenuToRole(roleID, menuID string) error {
	// TODO: 实现角色关联菜单表操作
	return nil
}

// RemoveMenuFromRole 从角色移除菜单
func (r *roleRepository) RemoveMenuFromRole(roleID, menuID string) error {
	// TODO: 实现角色关联菜单表操作
	return nil
}

// GetMenusByRole 获取角色下的所有菜单
func (r *roleRepository) GetMenusByRole(roleID string) ([]*entity.Menu, error) {
	// TODO: 实现角色关联菜单表操作
	return nil, nil
}

// GetRolesByMenu 获取菜单的所有角色
func (r *roleRepository) GetRolesByMenu(menuID string) ([]*entity.Role, error) {
	// TODO: 实现角色关联菜单表操作
	return nil, nil
}

// AddToRole 添加接口路由到角色
func (r *roleRepository) AddToRole(roleID, routeID string) error {
	// TODO: 实现角色关联接口路由表操作
	return nil
}

// RemoveFromRole 从角色移除接口路由
func (r *roleRepository) RemoveFromRole(roleID, routeID string) error {
	// TODO: 实现角色关联接口路由表操作
	return nil
}

// GetsByRole 获取角色下的所有接口路由
func (r *roleRepository) GetsByRole(roleID string) ([]*entity.API, error) {
	// TODO: 实现角色关联接口路由表操作
	return nil, nil
}

// GetRolesBy 获取接口路由的所有角色
func (r *roleRepository) GetRolesBy(routeID string) ([]*entity.Role, error) {
	// TODO: 实现角色关联接口路由表操作
	return nil, nil
}

// GetAPIByPathAndMethod 根据路径和方法获取接口路由
func (r *roleRepository) GetAPIByPathAndMethod(path, method string) (*entity.API, error) {
	// TODO: 实现接口路由查询
	return nil, nil
}

// AddBtnPermToRole 添加按钮权限到角色
func (r *roleRepository) AddBtnPermToRole(roleID, btnPermID string) error {
	// TODO: 实现角色关联按钮权限表操作
	return nil
}

// RemoveBtnPermFromRole 从角色移除按钮权限
func (r *roleRepository) RemoveBtnPermFromRole(roleID, btnPermID string) error {
	// TODO: 实现角色关联按钮权限表操作
	return nil
}

// GetBtnPermsByRole 获取角色下的所有按钮权限
func (r *roleRepository) GetBtnPermsByRole(roleID string) ([]*entity.BtnPerm, error) {
	// TODO: 实现角色关联按钮权限表操作
	return nil, nil
}

// GetRolesByBtnPerm 获取按钮权限的所有角色
func (r *roleRepository) GetRolesByBtnPerm(btnPermID string) ([]*entity.Role, error) {
	// TODO: 实现角色关联按钮权限表操作
	return nil, nil
}
