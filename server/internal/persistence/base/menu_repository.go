package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// menuModel 菜单数据库模型
type menuModel struct {
	database.SnowflakeBaseModel
	ParentID   *int64 `gorm:"not null;default:0"`
	Path       string `gorm:"size:255;not null"`
	Name       string `gorm:"size:100;not null;unique"`
	Component  string `gorm:"size:255"`
	Title      string `gorm:"size:50;not null"`
	Icon       string `gorm:"size:50"`
	Hidden     *bool  `gorm:"not null;default:false"`
	Sort       *int   `gorm:"not null;default:0"`
	Status     *int   `gorm:"not null;default:1"`
	IsExt      *bool  `gorm:"not null;default:false"`
	Redirect   string `gorm:"size:255"`
	Permission string `gorm:"size:100"`
	Type       *int   `gorm:"not null;default:2"`
	FrameSrc   string `gorm:"size:255"`

	// GORM 关联关系 - 一对多（子菜单）
	Children []menuModel `gorm:"foreignKey:parent_id;references:id"`

	// GORM 关联关系 - 多对一（父菜单）
	Parent *menuModel `gorm:"foreignKey:parent_id;references:id"`

	// GORM 关联关系 - 多对多（通过中间表 base_menu_api_routes）
	APIRoutes []apiModel `gorm:"many2many:base_menu_api_routes;joinForeignKey:menu_id;joinReferences:route_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// GORM 关联关系 - 一对多（按钮权限）
	BtnPerms []btnPermModel `gorm:"foreignKey:menu_id;references:id"`
}

// TableName 指定表名
func (menuModel) TableName() string {
	return "base_menus"
}

// toDomain 将数据库模型转换为领域实体
func (m *menuModel) toDomain() *entity.Menu {
	if m == nil {
		return nil
	}
	menu := &entity.Menu{
		ID:         m.ID,
		Path:       m.Path,
		Name:       m.Name,
		Component:  m.Component,
		Title:      m.Title,
		Icon:       m.Icon,
		Redirect:   m.Redirect,
		Permission: m.Permission,
		FrameSrc:   m.FrameSrc,
		CreatedBy:  m.CreatedBy,
		CreatedAt:  m.CreatedAt,
		UpdatedBy:  m.UpdatedBy,
		UpdatedAt:  m.UpdatedAt,
	}

	// 安全解引用，提供默认值
	if m.ParentID != nil {
		menu.ParentID = *m.ParentID
	} else {
		menu.ParentID = 0
	}

	if m.Hidden != nil {
		menu.Hidden = *m.Hidden
	} else {
		menu.Hidden = false
	}

	if m.Sort != nil {
		menu.Sort = *m.Sort
	} else {
		menu.Sort = 0
	}

	if m.Status != nil {
		menu.Status = *m.Status
	} else {
		menu.Status = 1
	}

	if m.IsExt != nil {
		menu.IsExt = *m.IsExt
	} else {
		menu.IsExt = false
	}

	if m.Type != nil {
		menu.Type = entity.MenuType(*m.Type)
	} else {
		menu.Type = entity.MenuType(2)
	}

	// ⭐ 处理关联数据 - 子菜单
	if len(m.Children) > 0 {
		children := make([]*entity.Menu, len(m.Children))
		for i, child := range m.Children {
			children[i] = child.toDomain()
		}
		menu.Children = children
	}

	// ⭐ 处理关联数据 - 父菜单
	if m.Parent != nil {
		menu.Parent = m.Parent.toDomain()
	}

	// ⭐ 处理关联数据 - API 路由（同时填充 APIRouteIds 和 APIRoutes）
	if len(m.APIRoutes) > 0 {
		apiRoutes := make([]*entity.API, len(m.APIRoutes))
		apiRouteIDs := make([]int64, len(m.APIRoutes))
		for i, apiRoute := range m.APIRoutes {
			apiRoutes[i] = apiRoute.toDomain()
			apiRouteIDs[i] = apiRoute.ID
		}
		menu.APIRoutes = apiRoutes
		menu.APIRouteIds = apiRouteIDs
	}

	// ⭐ 处理关联数据 - 按钮权限（同时填充 BtnPermIds 和 BtnPerms）
	if len(m.BtnPerms) > 0 {
		btnPerms := make([]*entity.BtnPerm, len(m.BtnPerms))
		btnPermIDs := make([]int64, len(m.BtnPerms))
		for i, btnPerm := range m.BtnPerms {
			btnPerms[i] = btnPerm.toDomain()
			btnPermIDs[i] = btnPerm.ID
		}
		menu.BtnPerms = btnPerms
		menu.BtnPermIds = btnPermIDs
	}

	return menu
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainMenu(menu *entity.Menu) (*menuModel, error) {
	return &menuModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        menu.ID,
			CreatedBy: menu.CreatedBy,
			UpdatedBy: menu.UpdatedBy,
		},
		ParentID:   common.Int64Ptr(menu.ParentID),
		Path:       menu.Path,
		Name:       menu.Name,
		Component:  menu.Component,
		Title:      menu.Title,
		Icon:       menu.Icon,
		Hidden:     common.BoolPtr(menu.Hidden),
		Sort:       common.IntPtr(menu.Sort),
		Status:     common.IntPtr(menu.Status),
		IsExt:      common.BoolPtr(menu.IsExt),
		Redirect:   menu.Redirect,
		Permission: menu.Permission,
		Type:       common.IntPtr(int(menu.Type)),
		FrameSrc:   menu.FrameSrc,
	}, nil
}

// menuRepository Repository 实现
type menuRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.MenuRepository = (*menuRepository)(nil)

// NewMenuRepository 创建菜单仓库实现
func NewMenuRepository(db *database.PostgresDB) repo.MenuRepository {
	return &menuRepository{db: db}
}

// GetByID 根据 ID 查询菜单并支持加载关联数据
func (r *menuRepository) GetByID(id int64, relations ...repo.MenuRelation) (*entity.Menu, error) {
	var dbModel menuModel
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

// GetByPath 根据路径查询菜单
func (r *menuRepository) GetByPath(path string) (*entity.Menu, error) {
	var dbModel menuModel
	result := r.db.Where("path = ?", path).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByCode 根据编码查询菜单
func (r *menuRepository) GetByCode(code string) (*entity.Menu, error) {
	var dbModel menuModel
	result := r.db.Where("name = ?", code).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetAll 获取所有菜单
func (r *menuRepository) GetAll() ([]*entity.Menu, error) {
	var dbModels []menuModel
	result := r.db.Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	menus := make([]*entity.Menu, len(dbModels))
	for i, model := range dbModels {
		menus[i] = model.toDomain()
	}

	return menus, nil
}

// GetMenusByRole 根据角色获取菜单
func (r *menuRepository) GetMenusByRole(roleID int64) ([]*entity.Menu, error) {
	var dbModels []menuModel
	// 通过角色 - 菜单关联表查询
	result := r.db.Joins("JOIN base_role_menus ON base_role_menus.menu_id = base_menus.id").
		Where("base_role_menus.role_id = ?", roleID).
		Where("base_menus.status = ?", 1).
		Order("base_menus.sort ASC").
		Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	menus := make([]*entity.Menu, len(dbModels))
	for i, model := range dbModels {
		menus[i] = model.toDomain()
	}

	return menus, nil
}

// GetMenusByUserID 根据用户 ID 获取菜单
func (r *menuRepository) GetMenusByUserID(userID int64) ([]*entity.Menu, error) {
	// TODO: 实现用户菜单查询
	return nil, nil
}

// GetMenusByType 根据类型获取菜单
func (r *menuRepository) GetMenusByType(menuType entity.MenuType) ([]*entity.Menu, error) {
	var dbModels []menuModel
	result := r.db.Where("type = ?", int(menuType)).Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	menus := make([]*entity.Menu, len(dbModels))
	for i, model := range dbModels {
		menus[i] = model.toDomain()
	}

	return menus, nil
}

// GetDefaultRouter 获取默认路由
func (r *menuRepository) GetDefaultRouter(roleID int64) (string, error) {
	// TODO: 实现默认路由查询
	return "", nil
}

// GetMenuTree 获取菜单树
func (r *menuRepository) GetMenuTree(parentID int64) ([]*entity.Menu, error) {
	var dbModels []menuModel
	result := r.db.Where("parent_id = ?", parentID).Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	menus := make([]*entity.Menu, len(dbModels))
	for i, model := range dbModels {
		menus[i] = model.toDomain()
	}

	return menus, nil
}

// GetAllMenuTree 获取所有菜单树
func (r *menuRepository) GetAllMenuTree() ([]*entity.Menu, error) {
	var dbModels []menuModel
	result := r.db.Order("parent_id ASC, sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	menus := make([]*entity.Menu, len(dbModels))
	for i, model := range dbModels {
		menus[i] = model.toDomain()
	}

	return menus, nil
}

// GetMenuPath 获取菜单路径
func (r *menuRepository) GetMenuPath(menuID int64) ([]*entity.Menu, error) {
	// TODO: 实现菜单路径查询
	return nil, nil
}

// Create 创建菜单
func (r *menuRepository) Create(menu *entity.Menu) error {
	dbModel, err := fromDomainMenu(menu)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	menu.ID = dbModel.ID
	return nil
}

// Update 更新菜单
func (r *menuRepository) Update(menu *entity.Menu) error {
	dbModel, err := fromDomainMenu(menu)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除菜单
func (r *menuRepository) Delete(id int64) error {
	return r.db.Delete(&menuModel{}, id).Error
}

// List 分页查询菜单列表
func (r *menuRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Menu, int64, error) {
	var total int64
	var dbModels []menuModel

	query := r.db.Model(&menuModel{})

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

	menus := make([]*entity.Menu, len(dbModels))
	for i, model := range dbModels {
		menus[i] = model.toDomain()
	}

	return menus, total, nil
}

// BatchDelete 批量删除菜单
func (r *menuRepository) BatchDelete(ids []int64) error {
	return r.db.Where("id IN ?", ids).Delete(&menuModel{}).Error
}

// CheckMenuChildren 检查菜单是否有子菜单
func (r *menuRepository) CheckMenuChildren(menuID int64) (bool, error) {
	var count int64
	result := r.db.Model(&menuModel{}).Where("parent_id = ?", menuID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}
