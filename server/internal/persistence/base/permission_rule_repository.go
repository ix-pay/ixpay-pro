package persistence

import (
	"encoding/json"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// permissionRuleModel 权限规则数据库模型
type permissionRuleModel struct {
	database.SnowflakeBaseModel
	Name        string `gorm:"size:100;not null;unique"`
	Description string `gorm:"size:500"`
	Effect      string `gorm:"size:10;not null"`
	APIPath     string `gorm:"size:255;not null"`
	Method      string `gorm:"size:20;not null"`
	Conditions  string `gorm:"type:text"`
	Status      *int   `gorm:"not null;default:1"`
	Sort        *int   `gorm:"not null;default:0"`
	IsSystem    *bool  `gorm:"not null;default:false"`
}

// TableName 指定表名
func (permissionRuleModel) TableName() string {
	return "base_permission_rules"
}

// toDomain 将数据库模型转换为领域实体
func (m *permissionRuleModel) toDomain() *entity.PermissionRule {
	if m == nil {
		return nil
	}

	var attributes []entity.PermissionAttribute
	if m.Conditions != "" {
		json.Unmarshal([]byte(m.Conditions), &attributes)
	}

	rule := &entity.PermissionRule{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Effect:      m.Effect,
		APIPath:     m.APIPath,
		Method:      m.Method,
		Conditions:  m.Conditions,
		Attributes:  attributes,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   m.UpdatedBy,
		UpdatedAt:   m.UpdatedAt,
	}

	// 安全解引用，提供默认值
	if m.Status != nil {
		rule.Status = *m.Status
	} else {
		rule.Status = 1
	}

	if m.Sort != nil {
		rule.Sort = *m.Sort
	} else {
		rule.Sort = 0
	}

	if m.IsSystem != nil {
		rule.IsSystem = *m.IsSystem
	} else {
		rule.IsSystem = false
	}

	return rule
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainPermissionRule(rule *entity.PermissionRule) (*permissionRuleModel, error) {
	conditionsJSON := ""
	if len(rule.Attributes) > 0 {
		jsonData, err := json.Marshal(rule.Attributes)
		if err != nil {
			return nil, err
		}
		conditionsJSON = string(jsonData)
	}

	return &permissionRuleModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        rule.ID,
			CreatedBy: rule.CreatedBy,
			UpdatedBy: rule.UpdatedBy,
		},
		Name:        rule.Name,
		Description: rule.Description,
		Effect:      rule.Effect,
		APIPath:     rule.APIPath,
		Method:      rule.Method,
		Conditions:  conditionsJSON,
		Status:      common.IntPtr(rule.Status),
		Sort:        common.IntPtr(rule.Sort),
		IsSystem:    common.BoolPtr(rule.IsSystem),
	}, nil
}

// permissionRuleRepository Repository 实现
type permissionRuleRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.PermissionRuleRepository = (*permissionRuleRepository)(nil)

// NewPermissionRuleRepository 创建权限规则仓库实现
func NewPermissionRuleRepository(db *database.PostgresDB) repo.PermissionRuleRepository {
	return &permissionRuleRepository{db: db}
}

// GetByID 根据 ID 查询权限规则
func (r *permissionRuleRepository) GetByID(id int64) (*entity.PermissionRule, error) {
	var dbModel permissionRuleModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByName 根据名称查询权限规则
func (r *permissionRuleRepository) GetByName(name string) (*entity.PermissionRule, error) {
	var dbModel permissionRuleModel
	result := r.db.Where("name = ?", name).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建权限规则
func (r *permissionRuleRepository) Create(rule *entity.PermissionRule) error {
	dbModel, err := fromDomainPermissionRule(rule)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	rule.ID = dbModel.ID
	return nil
}

// Update 更新权限规则
func (r *permissionRuleRepository) Update(rule *entity.PermissionRule) error {
	dbModel, err := fromDomainPermissionRule(rule)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除权限规则
func (r *permissionRuleRepository) Delete(id int64) error {
	return r.db.Delete(&permissionRuleModel{}, id).Error
}

// List 分页查询权限规则列表
func (r *permissionRuleRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.PermissionRule, int64, error) {
	var total int64
	var dbModels []permissionRuleModel

	query := r.db.Model(&permissionRuleModel{})

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

	rules := make([]*entity.PermissionRule, len(dbModels))
	for i, model := range dbModels {
		rules[i] = model.toDomain()
	}

	return rules, total, nil
}

// GetAllRules 获取所有权限规则
func (r *permissionRuleRepository) GetAllRules() ([]*entity.PermissionRule, error) {
	var dbModels []permissionRuleModel
	result := r.db.Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	rules := make([]*entity.PermissionRule, len(dbModels))
	for i, model := range dbModels {
		rules[i] = model.toDomain()
	}

	return rules, nil
}

// GetRulesByStatus 根据状态获取权限规则
func (r *permissionRuleRepository) GetRulesByStatus(status int) ([]*entity.PermissionRule, error) {
	var dbModels []permissionRuleModel
	result := r.db.Where("status = ?", status).Order("sort ASC").Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	rules := make([]*entity.PermissionRule, len(dbModels))
	for i, model := range dbModels {
		rules[i] = model.toDomain()
	}

	return rules, nil
}

// AddRoleToRule 添加角色到权限规则
func (r *permissionRuleRepository) AddRoleToRule(ruleID, roleID int64) error {
	// TODO: 实现权限规则关联角色表操作
	return nil
}

// RemoveRoleFromRule 从权限规则移除角色
func (r *permissionRuleRepository) RemoveRoleFromRule(ruleID, roleID int64) error {
	// TODO: 实现权限规则关联角色表操作
	return nil
}

// GetRolesByRule 获取权限规则的所有角色
func (r *permissionRuleRepository) GetRolesByRule(ruleID int64) ([]*entity.Role, error) {
	// TODO: 实现权限规则关联角色表操作
	return nil, nil
}

// GetRulesByRole 获取角色的所有权限规则
func (r *permissionRuleRepository) GetRulesByRole(roleID int64) ([]*entity.PermissionRule, error) {
	// TODO: 实现权限规则关联角色表操作
	return nil, nil
}

// AddUserToRule 添加用户到权限规则
func (r *permissionRuleRepository) AddUserToRule(ruleID, userID int64) error {
	// TODO: 实现权限规则关联用户表操作
	return nil
}

// RemoveUserFromRule 从权限规则移除用户
func (r *permissionRuleRepository) RemoveUserFromRule(ruleID, userID int64) error {
	// TODO: 实现权限规则关联用户表操作
	return nil
}

// GetUsersByRule 获取权限规则的所有用户
func (r *permissionRuleRepository) GetUsersByRule(ruleID int64) ([]*entity.User, error) {
	// TODO: 实现权限规则关联用户表操作
	return nil, nil
}

// GetRulesByUser 获取用户的所有权限规则
func (r *permissionRuleRepository) GetRulesByUser(userID int64) ([]*entity.PermissionRule, error) {
	// TODO: 实现权限规则关联用户表操作
	return nil, nil
}

// FindMatchingRules 查找匹配的权限规则
func (r *permissionRuleRepository) FindMatchingRules(apiPath, method string, attributes []entity.PermissionAttribute) ([]*entity.PermissionRule, error) {
	// TODO: 实现权限规则匹配逻辑
	return nil, nil
}
