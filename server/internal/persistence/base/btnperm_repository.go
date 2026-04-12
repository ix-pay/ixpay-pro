package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// btnPermModel 按钮权限数据库模型
type btnPermModel struct {
	database.SnowflakeBaseModel
	MenuID      int64  `gorm:"index"`
	Code        string `gorm:"size:100;not null;unique"`
	Name        string `gorm:"size:50;not null"`
	Description string `gorm:"size:255"`
	Status      int    `gorm:"default:1"`

	// GORM 关联关系 - 多对一（所属菜单）
	Menu *menuModel `gorm:"foreignKey:MenuID;references:ID"`

	// GORM 关联关系 - 多对多（通过中间表 base_btn_perm_api_routes）
	APIRoutes []apiModel `gorm:"many2many:base_btn_perm_api_routes;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName 指定表名
func (btnPermModel) TableName() string {
	return "base_btn_perms"
}

// toDomain 将数据库模型转换为领域实体
func (m *btnPermModel) toDomain() *entity.BtnPerm {
	if m == nil {
		return nil
	}
	btnPerm := &entity.BtnPerm{
		ID:        common.ToString(m.ID),
		MenuID:    common.ToString(m.MenuID),
		Code:      m.Code,
		Name:      m.Name,
		Status:    m.Status,
		CreatedBy: common.ToString(m.CreatedBy),
		CreatedAt: m.CreatedAt,
		UpdatedBy: common.ToString(m.UpdatedBy),
		UpdatedAt: m.UpdatedAt,
	}

	// ⭐ 处理关联数据 - 菜单
	if m.Menu != nil {
		btnPerm.Menu = m.Menu.toDomain()
	}

	// ⭐ 处理关联数据 - API 路由（同时填充 APIRouteIds 和 APIRoutes）
	if len(m.APIRoutes) > 0 {
		apiRoutes := make([]*entity.API, len(m.APIRoutes))
		apiRouteIDs := make([]string, len(m.APIRoutes))
		for i, apiRoute := range m.APIRoutes {
			apiRoutes[i] = apiRoute.toDomain()
			apiRouteIDs[i] = common.ToString(apiRoute.ID)
		}
		btnPerm.APIRoutes = apiRoutes
		btnPerm.APIRouteIds = apiRouteIDs
	}

	return btnPerm
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainBtnPerm(btnPerm *entity.BtnPerm) (*btnPermModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(btnPerm.ID, btnPerm.CreatedBy, btnPerm.UpdatedBy)

	return &btnPermModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		MenuID: common.TryParseInt64(btnPerm.MenuID),
		Code:   btnPerm.Code,
		Name:   btnPerm.Name,
		Status: btnPerm.Status,
	}, nil
}

// btnPermRepository Repository 实现
type btnPermRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.BtnPermRepository = (*btnPermRepository)(nil)

// NewBtnPermRepository 创建按钮权限仓库实现
func NewBtnPermRepository(db *database.PostgresDB) repo.BtnPermRepository {
	return &btnPermRepository{db: db}
}

// GetByID 根据 ID 查询按钮权限并支持加载关联数据
func (r *btnPermRepository) GetByID(id string, relations ...repo.BtnPermRelation) (*entity.BtnPerm, error) {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}

	var dbModel btnPermModel
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

// GetByCode 根据编码查询按钮权限
func (r *btnPermRepository) GetByCode(code string) (*entity.BtnPerm, error) {
	var dbModel btnPermModel
	result := r.db.Where("code = ?", code).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetBtnPermsByMenu 根据菜单获取按钮权限
func (r *btnPermRepository) GetBtnPermsByMenu(menuID string) ([]*entity.BtnPerm, error) {
	intMenuID := common.TryParseInt64(menuID)
	var dbModels []btnPermModel
	result := r.db.Where("menu_id = ?", intMenuID).Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	btnPerms := make([]*entity.BtnPerm, len(dbModels))
	for i, model := range dbModels {
		btnPerms[i] = model.toDomain()
	}

	return btnPerms, nil
}

// Create 创建按钮权限
func (r *btnPermRepository) Create(button *entity.BtnPerm) error {
	dbModel, err := fromDomainBtnPerm(button)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
}

// Update 更新按钮权限
func (r *btnPermRepository) Update(button *entity.BtnPerm) error {
	dbModel, err := fromDomainBtnPerm(button)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除按钮权限
func (r *btnPermRepository) Delete(id string) error {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&btnPermModel{}, intID).Error
}

// List 分页查询按钮权限列表
func (r *btnPermRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.BtnPerm, int64, error) {
	var total int64
	var dbModels []btnPermModel

	query := r.db.Model(&btnPermModel{})

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

	btnPerms := make([]*entity.BtnPerm, len(dbModels))
	for i, model := range dbModels {
		btnPerms[i] = model.toDomain()
	}

	return btnPerms, total, nil
}

// AddAPIToBtnPerm 添加 API 路由到按钮权限
func (r *btnPermRepository) AddAPIToBtnPerm(buttonID, routeID string) error {
	// TODO: 实现按钮权限关联 API 路由表操作
	return nil
}

// RemoveAPIFromBtnPerm 从按钮权限移除 API 路由
func (r *btnPermRepository) RemoveAPIFromBtnPerm(buttonID, routeID string) error {
	// TODO: 实现按钮权限关联 API 路由表操作
	return nil
}

// GetAPIsByBtnPerm 获取按钮权限下的所有 API 路由
func (r *btnPermRepository) GetAPIsByBtnPerm(buttonID string) ([]*entity.API, error) {
	// TODO: 实现按钮权限关联 API 路由表操作
	return nil, nil
}

// GetBtnPermsByAPI 获取 API 路由的所有按钮权限
func (r *btnPermRepository) GetBtnPermsByAPI(routeID string) ([]*entity.BtnPerm, error) {
	// TODO: 实现按钮权限关联 API 路由表操作
	return nil, nil
}
