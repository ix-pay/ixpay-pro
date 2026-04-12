package persistence

import (
	"strconv"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// menuModel 菜单数据库模型
type menuModel struct {
	database.SnowflakeBaseModel
	ParentID   int64  `gorm:"default:0"`
	Path       string `gorm:"size:255;not null"`
	Name       string `gorm:"size:100;not null;unique"`
	Component  string `gorm:"size:255"`
	Title      string `gorm:"size:50;not null"`
	Icon       string `gorm:"size:50"`
	Hidden     bool   `gorm:"default:false"`
	Sort       int    `gorm:"default:0"`
	Status     int    `gorm:"default:1"`
	IsExt      bool   `gorm:"default:false"`
	Redirect   string `gorm:"size:255"`
	Permission string `gorm:"size:100"`
	Type       int    `gorm:"default:2"`
	FrameSrc   string `gorm:"size:255"`
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
	return &entity.Menu{
		ID:         strconv.FormatInt(m.ID, 10),
		ParentID:   strconv.FormatInt(m.ParentID, 10),
		Path:       m.Path,
		Name:       m.Name,
		Component:  m.Component,
		Title:      m.Title,
		Icon:       m.Icon,
		Hidden:     m.Hidden,
		Sort:       m.Sort,
		Status:     m.Status,
		IsExt:      m.IsExt,
		Redirect:   m.Redirect,
		Permission: m.Permission,
		Type:       entity.MenuType(m.Type),
		FrameSrc:   m.FrameSrc,
		CreatedBy:  strconv.FormatInt(m.CreatedBy, 10),
		CreatedAt:  m.CreatedAt,
		UpdatedBy:  strconv.FormatInt(m.UpdatedBy, 10),
		UpdatedAt:  m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainMenu(menu *entity.Menu) (*menuModel, error) {
	var id int64
	var err error

	if menu.ID != "" {
		id, err = strconv.ParseInt(menu.ID, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	var parentID int64
	if menu.ParentID != "" {
		parentID, _ = strconv.ParseInt(menu.ParentID, 10, 64)
	}

	var createdBy int64
	if menu.CreatedBy != "" {
		createdBy, _ = strconv.ParseInt(menu.CreatedBy, 10, 64)
	}

	var updatedBy int64
	if menu.UpdatedBy != "" {
		updatedBy, _ = strconv.ParseInt(menu.UpdatedBy, 10, 64)
	}

	return &menuModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		ParentID:   parentID,
		Path:       menu.Path,
		Name:       menu.Name,
		Component:  menu.Component,
		Title:      menu.Title,
		Icon:       menu.Icon,
		Hidden:     menu.Hidden,
		Sort:       menu.Sort,
		Status:     menu.Status,
		IsExt:      menu.IsExt,
		Redirect:   menu.Redirect,
		Permission: menu.Permission,
		Type:       int(menu.Type),
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

// GetByID 根据 ID 查询菜单
func (r *menuRepository) GetByID(id string) (*entity.Menu, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var dbModel menuModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
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
func (r *menuRepository) GetMenusByRole(role string) ([]*entity.Menu, error) {
	// TODO: 实现角色关联菜单查询
	return nil, nil
}

// GetMenusByUserID 根据用户 ID 获取菜单
func (r *menuRepository) GetMenusByUserID(userID string) ([]*entity.Menu, error) {
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
func (r *menuRepository) GetDefaultRouter(role string) (string, error) {
	// TODO: 实现默认路由查询
	return "", nil
}

// GetMenuTree 获取菜单树
func (r *menuRepository) GetMenuTree(parentID string) ([]*entity.Menu, error) {
	intParentID, _ := strconv.ParseInt(parentID, 10, 64)
	var dbModels []menuModel
	result := r.db.Where("parent_id = ?", intParentID).Order("sort ASC").Find(&dbModels)
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
func (r *menuRepository) GetMenuPath(menuID string) ([]*entity.Menu, error) {
	// TODO: 实现菜单路径查询
	return nil, nil
}

// Create 创建菜单
func (r *menuRepository) Create(menu *entity.Menu) error {
	dbModel, err := fromDomainMenu(menu)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
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
func (r *menuRepository) Delete(id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	return r.db.Delete(&menuModel{}, intID).Error
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
func (r *menuRepository) BatchDelete(ids []string) error {
	intIDs, err := common.StringToInt64s(ids)
	if err != nil {
		return err
	}

	return r.db.Where("id IN ?", intIDs).Delete(&menuModel{}).Error
}

// CheckMenuChildren 检查菜单是否有子菜单
func (r *menuRepository) CheckMenuChildren(menuID string) (bool, error) {
	intID, err := strconv.ParseInt(menuID, 10, 64)
	if err != nil {
		return false, err
	}

	var count int64
	result := r.db.Model(&menuModel{}).Where("parent_id = ?", intID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}
