package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
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
	Users []*userModel `gorm:"many2many:base_role_users;joinForeignKey:role_id;joinReferences:user_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Menus []*menuModel `gorm:"many2many:base_role_menus;joinForeignKey:role_id;joinReferences:menu_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// GORM 关联关系 - 多对多（通过中间表）
	APIRoutes []*apiModel     `gorm:"many2many:base_role_api_routes;joinForeignKey:role_id;joinReferences:route_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	BtnPerms  []*btnPermModel `gorm:"many2many:base_role_btn_perms;joinForeignKey:role_id;joinReferences:button_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// GORM 关联关系 - 一对多（子角色）
	Children []*roleModel `gorm:"foreignKey:parent_id;references:id"`

	// GORM 关联关系 - 多对一（父角色）
	Parent *roleModel `gorm:"foreignKey:parent_id;references:id"`
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
		ID:        m.ID,
		Name:      m.Name,
		Code:      m.Code,
		Type:      m.Type,
		ParentID:  m.ParentID,
		Status:    m.Status,
		IsSystem:  m.IsSystem,
		Sort:      m.Sort,
		CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt,
		UpdatedBy: m.UpdatedBy,
		UpdatedAt: m.UpdatedAt,
	}

	// ⭐ 处理关联数据 - 用户（同时填充 UserIds 和 Users）
	if len(m.Users) > 0 {
		users := make([]*entity.User, len(m.Users))
		userIDs := make([]int64, len(m.Users))
		for i, user := range m.Users {
			users[i] = user.toDomain()
			userIDs[i] = user.ID
		}
		role.Users = users
		role.UserIds = userIDs
	}

	// ⭐ 处理关联数据 - 菜单（同时填充 MenuIds 和 Menus）
	if len(m.Menus) > 0 {
		menus := make([]*entity.Menu, len(m.Menus))
		menuIDs := make([]int64, len(m.Menus))
		for i, menu := range m.Menus {
			menus[i] = menu.toDomain()
			menuIDs[i] = menu.ID
		}
		role.Menus = menus
		role.MenuIds = menuIDs
	}

	// ⭐ 处理关联数据 - API 路由（同时填充 APIRouteIds 和 APIRoutes）
	if len(m.APIRoutes) > 0 {
		apiRoutes := make([]*entity.API, len(m.APIRoutes))
		apiRouteIDs := make([]int64, len(m.APIRoutes))
		for i, apiRoute := range m.APIRoutes {
			apiRoutes[i] = apiRoute.toDomain()
			apiRouteIDs[i] = apiRoute.ID
		}
		role.APIRoutes = apiRoutes
		role.APIRouteIds = apiRouteIDs
	}

	// ⭐ 处理关联数据 - 按钮权限（同时填充 BtnPermIds 和 BtnPerms）
	if len(m.BtnPerms) > 0 {
		btnPerms := make([]*entity.BtnPerm, len(m.BtnPerms))
		btnPermIDs := make([]int64, len(m.BtnPerms))
		for i, btnPerm := range m.BtnPerms {
			btnPerms[i] = btnPerm.toDomain()
			btnPermIDs[i] = btnPerm.ID
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
	return &roleModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        role.ID,
			CreatedBy: role.CreatedBy,
			UpdatedBy: role.UpdatedBy,
		},
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Type:        role.Type,
		ParentID:    role.ParentID,
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
func (r *roleRepository) GetByID(id int64, relations ...repo.RoleRelation) (*entity.Role, error) {
	var dbModel roleModel
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

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	role.ID = dbModel.ID
	return nil
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
func (r *roleRepository) Delete(id int64) error {
	return r.db.Delete(&roleModel{}, id).Error
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
func (r *roleRepository) AddUserToRole(roleID, userID int64) error {
	// 先检查是否已存在关联，避免重复创建
	var count int64
	err := r.db.Model(&roleUserModel{}).Where("role_id = ? AND user_id = ?", roleID, userID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // 已存在关联则直接返回成功
	}

	model := &roleUserModel{
		RoleID: roleID,
		UserID: userID,
	}

	return r.db.Create(model).Error
}

// RemoveUserFromRole 从角色移除用户
func (r *roleRepository) RemoveUserFromRole(roleID, userID int64) error {
	return r.db.Where("role_id = ? AND user_id = ?", roleID, userID).Delete(&roleUserModel{}).Error
}

// ExistsUserInRole 检查用户是否在角色中
func (r *roleRepository) ExistsUserInRole(roleID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&roleUserModel{}).Where("role_id = ? AND user_id = ?", roleID, userID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUsersByRole 获取角色下的所有用户
func (r *roleRepository) GetUsersByRole(roleID int64) ([]*entity.User, error) {
	var userModels []userModel
	err := r.db.Table("base_users").
		Joins("JOIN base_role_users ON base_role_users.user_id = base_users.id").
		Where("base_role_users.role_id = ?", roleID).
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
func (r *roleRepository) GetRolesByUser(userID int64) ([]*entity.Role, error) {
	var roleModels []roleModel
	err := r.db.Table("base_roles").
		Joins("JOIN base_role_users ON base_role_users.role_id = base_roles.id").
		Where("base_role_users.user_id = ?", userID).
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
func (r *roleRepository) AddMenuToRole(roleID, menuID int64) error {
	// 先检查是否已存在关联，避免重复创建
	var count int64
	err := r.db.Model(&roleMenuModel{}).Where("role_id = ? AND menu_id = ?", roleID, menuID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // 已存在关联则直接返回成功
	}

	model := &roleMenuModel{
		RoleID: roleID,
		MenuID: menuID,
	}

	return r.db.Create(model).Error
}

// RemoveMenuFromRole 从角色移除菜单
func (r *roleRepository) RemoveMenuFromRole(roleID, menuID int64) error {
	return r.db.Where("role_id = ? AND menu_id = ?", roleID, menuID).Delete(&roleMenuModel{}).Error
}

// GetMenusByRole 获取角色下的所有菜单
func (r *roleRepository) GetMenusByRole(roleID int64) ([]*entity.Menu, error) {
	var menuModels []menuModel
	err := r.db.Table("base_menus").
		Joins("JOIN base_role_menus ON base_role_menus.menu_id = base_menus.id").
		Where("base_role_menus.role_id = ?", roleID).
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
func (r *roleRepository) GetRolesByMenu(menuID int64) ([]*entity.Role, error) {
	var roleModels []roleModel
	err := r.db.Table("base_roles").
		Joins("JOIN base_role_menus ON base_role_menus.role_id = base_roles.id").
		Where("base_role_menus.menu_id = ?", menuID).
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
func (r *roleRepository) AddToRole(roleID, routeID int64) error {
	// 先检查是否已存在关联，避免重复创建
	var count int64
	err := r.db.Model(&roleAPIRouteModel{}).Where("role_id = ? AND route_id = ?", roleID, routeID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // 已存在关联则直接返回成功
	}

	model := &roleAPIRouteModel{
		RoleID:  roleID,
		RouteID: routeID,
		Source:  1, // 1-直接授权
	}

	return r.db.Create(model).Error
}

// RemoveFromRole 从角色移除接口路由
func (r *roleRepository) RemoveFromRole(roleID, routeID int64) error {
	return r.db.Where("role_id = ? AND route_id = ?", roleID, routeID).Delete(&roleAPIRouteModel{}).Error
}

// GetsByRole 获取角色下的所有接口路由
func (r *roleRepository) GetsByRole(roleID int64) ([]*entity.API, error) {
	var apiModels []apiModel
	err := r.db.Table("base_apis").
		Joins("JOIN base_role_api_routes ON base_role_api_routes.route_id = base_apis.id").
		Where("base_role_api_routes.role_id = ?", roleID).
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
func (r *roleRepository) GetRolesBy(routeID int64) ([]*entity.Role, error) {
	var roleModels []roleModel
	err := r.db.Table("base_roles").
		Joins("JOIN base_role_api_routes ON base_role_api_routes.role_id = base_roles.id").
		Where("base_role_api_routes.route_id = ?", routeID).
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
func (r *roleRepository) AddBtnPermToRole(roleID, btnPermID int64) error {
	// 先检查是否已存在关联，避免重复创建
	var count int64
	err := r.db.Model(&roleBtnPermModel{}).Where("role_id = ? AND btn_perm_id = ?", roleID, btnPermID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // 已存在关联则直接返回成功
	}

	model := &roleBtnPermModel{
		RoleID:    roleID,
		BtnPermID: btnPermID,
	}

	return r.db.Create(model).Error
}

// RemoveBtnPermFromRole 从角色移除按钮权限
func (r *roleRepository) RemoveBtnPermFromRole(roleID, btnPermID int64) error {
	return r.db.Where("role_id = ? AND btn_perm_id = ?", roleID, btnPermID).Delete(&roleBtnPermModel{}).Error
}

// GetBtnPermsByRole 获取角色下的所有按钮权限
func (r *roleRepository) GetBtnPermsByRole(roleID int64) ([]*entity.BtnPerm, error) {
	var btnPermModels []btnPermModel
	err := r.db.Table("base_btn_perms").
		Joins("JOIN base_role_btn_perms ON base_role_btn_perms.btn_perm_id = base_btn_perms.id").
		Where("base_role_btn_perms.role_id = ?", roleID).
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
func (r *roleRepository) GetRolesByBtnPerm(btnPermID int64) ([]*entity.Role, error) {
	var roleModels []roleModel
	err := r.db.Table("base_roles").
		Joins("JOIN base_role_btn_perms ON base_role_btn_perms.role_id = base_roles.id").
		Where("base_role_btn_perms.btn_perm_id = ?", btnPermID).
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

// GetUserSpecialBtnPermissions 获取用户的特殊按钮权限
func (r *roleRepository) GetUserSpecialBtnPermissions(userID int64) ([]*entity.BtnPerm, error) {
	var btnPermModels []btnPermModel
	err := r.db.Table("base_btn_perms").
		Joins("JOIN base_user_btn_perms ON base_user_btn_perms.btn_perm_id = base_btn_perms.id").
		Where("base_user_btn_perms.user_id = ?", userID).
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
