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

	// GORM 关联关系 - 多对多（通过中间表）
	Users []userModel `gorm:"many2many:base_role_users;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Menus []menuModel `gorm:"many2many:base_role_menus;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// GORM 关联关系 - 多对多（通过中间表）
	APIRoutes []apiModel     `gorm:"many2many:base_role_api_routes;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	BtnPerms  []btnPermModel `gorm:"many2many:base_role_btn_perms;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// GORM 关联关系 - 一对多（子角色）
	Children []roleModel `gorm:"foreignKey:ParentID;references:ID"`

	// GORM 关联关系 - 多对一（父角色）
	Parent *roleModel `gorm:"foreignKey:ParentID;references:ID"`
}

// TableName 指定表名
func (roleModel) TableName() string {
	return "base_roles"
}

// roleUserModel 角色用户关联模型
type roleUserModel struct {
	RoleID int64 `gorm:"not null;index"`
	UserID int64 `gorm:"not null;index"`
}

// TableName 指定表名
func (roleUserModel) TableName() string {
	return "base_role_users"
}

// roleMenuModel 角色菜单关联模型
type roleMenuModel struct {
	RoleID int64 `gorm:"not null;index"`
	MenuID int64 `gorm:"not null;index"`
}

// TableName 指定表名
func (roleMenuModel) TableName() string {
	return "base_role_menus"
}

// roleAPIRouteModel 角色 API 路由关联模型
type roleAPIRouteModel struct {
	RoleID  int64  `gorm:"not null;index"`
	RouteID int64  `gorm:"not null;index"`
	Source  int    `gorm:"default:1"`
	Note    string `gorm:"size:255"`
}

// TableName 指定表名
func (roleAPIRouteModel) TableName() string {
	return "base_role_api_routes"
}

// roleBtnPermModel 角色按钮权限关联模型
type roleBtnPermModel struct {
	RoleID    int64 `gorm:"not null;index"`
	BtnPermID int64 `gorm:"not null;index"`
}

// TableName 指定表名
func (roleBtnPermModel) TableName() string {
	return "base_role_btn_perms"
}

// toDomain 将数据库模型转换为领域实体
func (m *roleModel) toDomain() *entity.Role {
	if m == nil {
		return nil
	}
	role := &entity.Role{
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

	// ⭐ 处理关联数据 - 用户（同时填充 UserIds 和 Users）
	if len(m.Users) > 0 {
		users := make([]*entity.User, len(m.Users))
		userIDs := make([]string, len(m.Users))
		for i, user := range m.Users {
			users[i] = user.toDomain()
			userIDs[i] = common.ToString(user.ID)
		}
		role.Users = users
		role.UserIds = userIDs
	}

	// ⭐ 处理关联数据 - 菜单（同时填充 MenuIds 和 Menus）
	if len(m.Menus) > 0 {
		menus := make([]*entity.Menu, len(m.Menus))
		menuIDs := make([]string, len(m.Menus))
		for i, menu := range m.Menus {
			menus[i] = menu.toDomain()
			menuIDs[i] = common.ToString(menu.ID)
		}
		role.Menus = menus
		role.MenuIds = menuIDs
	}

	// ⭐ 处理关联数据 - API 路由（同时填充 APIRouteIds 和 APIRoutes）
	if len(m.APIRoutes) > 0 {
		apiRoutes := make([]*entity.API, len(m.APIRoutes))
		apiRouteIDs := make([]string, len(m.APIRoutes))
		for i, apiRoute := range m.APIRoutes {
			apiRoutes[i] = apiRoute.toDomain()
			apiRouteIDs[i] = common.ToString(apiRoute.ID)
		}
		role.APIRoutes = apiRoutes
		role.APIRouteIds = apiRouteIDs
	}

	// ⭐ 处理关联数据 - 按钮权限（同时填充 BtnPermIds 和 BtnPerms）
	if len(m.BtnPerms) > 0 {
		btnPerms := make([]*entity.BtnPerm, len(m.BtnPerms))
		btnPermIDs := make([]string, len(m.BtnPerms))
		for i, btnPerm := range m.BtnPerms {
			btnPerms[i] = btnPerm.toDomain()
			btnPermIDs[i] = common.ToString(btnPerm.ID)
		}
		role.BtnPerms = btnPerms
		role.BtnPermIds = btnPermIDs
	}

	// ⭐ 处理关联数据 - 父角色
	if m.Parent != nil {
		role.Parent = m.Parent.toDomain()
	}

	// ⭐ 处理关联数据 - 子角色
	if len(m.Children) > 0 {
		children := make([]*entity.Role, len(m.Children))
		for i, child := range m.Children {
			children[i] = child.toDomain()
		}
		role.Children = children
	}

	return role
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

// GetByID 根据 ID 查询角色并支持加载关联数据
func (r *roleRepository) GetByID(id string, relations ...repo.RoleRelation) (*entity.Role, error) {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}

	var dbModel roleModel
	query := r.db.Where("id = ?", intID)

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
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return err
	}
	uID, err := common.ParseInt64(userID)
	if err != nil {
		return err
	}

	model := &roleUserModel{
		RoleID: rID,
		UserID: uID,
	}

	return r.db.Create(model).Error
}

// RemoveUserFromRole 从角色移除用户
func (r *roleRepository) RemoveUserFromRole(roleID, userID string) error {
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return err
	}
	uID, err := common.ParseInt64(userID)
	if err != nil {
		return err
	}

	return r.db.Where("role_id = ? AND user_id = ?", rID, uID).Delete(&roleUserModel{}).Error
}

// ExistsUserInRole 检查用户是否在角色中
func (r *roleRepository) ExistsUserInRole(roleID, userID string) (bool, error) {
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return false, err
	}
	uID, err := common.ParseInt64(userID)
	if err != nil {
		return false, err
	}

	var count int64
	err = r.db.Model(&roleUserModel{}).Where("role_id = ? AND user_id = ?", rID, uID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUsersByRole 获取角色下的所有用户
func (r *roleRepository) GetUsersByRole(roleID string) ([]*entity.User, error) {
	intID, err := common.ParseInt64(roleID)
	if err != nil {
		return nil, err
	}

	var userModels []userModel
	err = r.db.Table("base_users").
		Joins("JOIN base_role_users ON base_role_users.user_id = base_users.id").
		Where("base_role_users.role_id = ?", intID).
		Find(&userModels).Error
	if err != nil {
		return nil, err
	}

	users := make([]*entity.User, len(userModels))
	for i, model := range userModels {
		users[i] = model.toDomain()
	}

	return users, nil
}

// GetRolesByUser 获取用户的所有角色
func (r *roleRepository) GetRolesByUser(userID string) ([]*entity.Role, error) {
	intID, err := common.ParseInt64(userID)
	if err != nil {
		return nil, err
	}

	var roleModels []roleModel
	err = r.db.Table("base_roles").
		Joins("JOIN base_role_users ON base_role_users.role_id = base_roles.id").
		Where("base_role_users.user_id = ?", intID).
		Find(&roleModels).Error
	if err != nil {
		return nil, err
	}

	roles := make([]*entity.Role, len(roleModels))
	for i, model := range roleModels {
		roles[i] = model.toDomain()
	}

	return roles, nil
}

// AddMenuToRole 添加菜单到角色
func (r *roleRepository) AddMenuToRole(roleID, menuID string) error {
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return err
	}
	mID, err := common.ParseInt64(menuID)
	if err != nil {
		return err
	}

	model := &roleMenuModel{
		RoleID: rID,
		MenuID: mID,
	}

	return r.db.Create(model).Error
}

// RemoveMenuFromRole 从角色移除菜单
func (r *roleRepository) RemoveMenuFromRole(roleID, menuID string) error {
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return err
	}
	mID, err := common.ParseInt64(menuID)
	if err != nil {
		return err
	}

	return r.db.Where("role_id = ? AND menu_id = ?", rID, mID).Delete(&roleMenuModel{}).Error
}

// GetMenusByRole 获取角色下的所有菜单
func (r *roleRepository) GetMenusByRole(roleID string) ([]*entity.Menu, error) {
	intID, err := common.ParseInt64(roleID)
	if err != nil {
		return nil, err
	}

	var menuModels []menuModel
	err = r.db.Table("base_menus").
		Joins("JOIN base_role_menus ON base_role_menus.menu_id = base_menus.id").
		Where("base_role_menus.role_id = ?", intID).
		Find(&menuModels).Error
	if err != nil {
		return nil, err
	}

	menus := make([]*entity.Menu, len(menuModels))
	for i, model := range menuModels {
		menus[i] = model.toDomain()
	}

	return menus, nil
}

// GetRolesByMenu 获取菜单的所有角色
func (r *roleRepository) GetRolesByMenu(menuID string) ([]*entity.Role, error) {
	intID, err := common.ParseInt64(menuID)
	if err != nil {
		return nil, err
	}

	var roleModels []roleModel
	err = r.db.Table("base_roles").
		Joins("JOIN base_role_menus ON base_role_menus.role_id = base_roles.id").
		Where("base_role_menus.menu_id = ?", intID).
		Find(&roleModels).Error
	if err != nil {
		return nil, err
	}

	roles := make([]*entity.Role, len(roleModels))
	for i, model := range roleModels {
		roles[i] = model.toDomain()
	}

	return roles, nil
}

// AddToRole 添加接口路由到角色
func (r *roleRepository) AddToRole(roleID, routeID string) error {
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return err
	}
	rtID, err := common.ParseInt64(routeID)
	if err != nil {
		return err
	}

	model := &roleAPIRouteModel{
		RoleID:  rID,
		RouteID: rtID,
		Source:  1, // 1-直接授权
	}

	return r.db.Create(model).Error
}

// RemoveFromRole 从角色移除接口路由
func (r *roleRepository) RemoveFromRole(roleID, routeID string) error {
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return err
	}
	rtID, err := common.ParseInt64(routeID)
	if err != nil {
		return err
	}

	return r.db.Where("role_id = ? AND route_id = ?", rID, rtID).Delete(&roleAPIRouteModel{}).Error
}

// GetsByRole 获取角色下的所有接口路由
func (r *roleRepository) GetsByRole(roleID string) ([]*entity.API, error) {
	intID, err := common.ParseInt64(roleID)
	if err != nil {
		return nil, err
	}

	var apiModels []apiModel
	err = r.db.Table("base_apis").
		Joins("JOIN base_role_api_routes ON base_role_api_routes.route_id = base_apis.id").
		Where("base_role_api_routes.role_id = ?", intID).
		Find(&apiModels).Error
	if err != nil {
		return nil, err
	}

	apis := make([]*entity.API, len(apiModels))
	for i, model := range apiModels {
		apis[i] = model.toDomain()
	}

	return apis, nil
}

// GetRolesBy 获取接口路由的所有角色
func (r *roleRepository) GetRolesBy(routeID string) ([]*entity.Role, error) {
	intID, err := common.ParseInt64(routeID)
	if err != nil {
		return nil, err
	}

	var roleModels []roleModel
	err = r.db.Table("base_roles").
		Joins("JOIN base_role_api_routes ON base_role_api_routes.role_id = base_roles.id").
		Where("base_role_api_routes.route_id = ?", intID).
		Find(&roleModels).Error
	if err != nil {
		return nil, err
	}

	roles := make([]*entity.Role, len(roleModels))
	for i, model := range roleModels {
		roles[i] = model.toDomain()
	}

	return roles, nil
}

// GetAPIByPathAndMethod 根据路径和方法获取接口路由
func (r *roleRepository) GetAPIByPathAndMethod(path, method string) (*entity.API, error) {
	var apiModel apiModel
	result := r.db.Where("path = ? AND method = ?", path, method).First(&apiModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return apiModel.toDomain(), nil
}

// AddBtnPermToRole 添加按钮权限到角色
func (r *roleRepository) AddBtnPermToRole(roleID, btnPermID string) error {
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return err
	}
	bID, err := common.ParseInt64(btnPermID)
	if err != nil {
		return err
	}

	model := &roleBtnPermModel{
		RoleID:    rID,
		BtnPermID: bID,
	}

	return r.db.Create(model).Error
}

// RemoveBtnPermFromRole 从角色移除按钮权限
func (r *roleRepository) RemoveBtnPermFromRole(roleID, btnPermID string) error {
	rID, err := common.ParseInt64(roleID)
	if err != nil {
		return err
	}
	bID, err := common.ParseInt64(btnPermID)
	if err != nil {
		return err
	}

	return r.db.Where("role_id = ? AND btn_perm_id = ?", rID, bID).Delete(&roleBtnPermModel{}).Error
}

// GetBtnPermsByRole 获取角色下的所有按钮权限
func (r *roleRepository) GetBtnPermsByRole(roleID string) ([]*entity.BtnPerm, error) {
	intID, err := common.ParseInt64(roleID)
	if err != nil {
		return nil, err
	}

	var btnPermModels []btnPermModel
	err = r.db.Table("base_btn_perms").
		Joins("JOIN base_role_btn_perms ON base_role_btn_perms.btn_perm_id = base_btn_perms.id").
		Where("base_role_btn_perms.role_id = ?", intID).
		Find(&btnPermModels).Error
	if err != nil {
		return nil, err
	}

	btnPerms := make([]*entity.BtnPerm, len(btnPermModels))
	for i, model := range btnPermModels {
		btnPerms[i] = model.toDomain()
	}

	return btnPerms, nil
}

// GetRolesByBtnPerm 获取按钮权限的所有角色
func (r *roleRepository) GetRolesByBtnPerm(btnPermID string) ([]*entity.Role, error) {
	intID, err := common.ParseInt64(btnPermID)
	if err != nil {
		return nil, err
	}

	var roleModels []roleModel
	err = r.db.Table("base_roles").
		Joins("JOIN base_role_btn_perms ON base_role_btn_perms.role_id = base_roles.id").
		Where("base_role_btn_perms.btn_perm_id = ?", intID).
		Find(&roleModels).Error
	if err != nil {
		return nil, err
	}

	roles := make([]*entity.Role, len(roleModels))
	for i, model := range roleModels {
		roles[i] = model.toDomain()
	}

	return roles, nil
}
