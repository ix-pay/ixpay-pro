package persistence

import (
	"encoding/json"
	"strconv"
	"strings"

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
	var conditionNode *entity.ConditionNode

	if m.Conditions != "" {
		// 尝试解析为新格式（结构化条件节点）
		var node entity.ConditionNode
		if err := json.Unmarshal([]byte(m.Conditions), &node); err == nil {
			// 成功解析为新格式
			conditionNode = &node
		} else {
			// 回退到旧格式（简单属性数组）
			json.Unmarshal([]byte(m.Conditions), &attributes)
		}
	}

	rule := &entity.PermissionRule{
		ID:            m.ID,
		Name:          m.Name,
		Description:   m.Description,
		Effect:        m.Effect,
		APIPath:       m.APIPath,
		Method:        m.Method,
		Conditions:    m.Conditions,
		Attributes:    attributes,
		ConditionNode: conditionNode,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
		UpdatedBy:     m.UpdatedBy,
		UpdatedAt:     m.UpdatedAt,
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
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, name ASC, id ASC").Find(&dbModels).Error; err != nil {
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
	result := r.db.Order("sort ASC, name ASC, id ASC").Find(&dbModels)
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
	result := r.db.Where("status = ?", status).Order("sort ASC, name ASC, id ASC").Find(&dbModels)
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
	return r.db.Table("base_permission_rule_roles").Create(map[string]interface{}{
		"rule_id": ruleID,
		"role_id": roleID,
	}).Error
}

// RemoveRoleFromRule 从权限规则移除角色
func (r *permissionRuleRepository) RemoveRoleFromRule(ruleID, roleID int64) error {
	return r.db.Table("base_permission_rule_roles").
		Where("rule_id = ? AND role_id = ?", ruleID, roleID).
		Delete(nil).Error
}

// GetRolesByRule 获取权限规则的所有角色
func (r *permissionRuleRepository) GetRolesByRule(ruleID int64) ([]*entity.Role, error) {
	var roles []*entity.Role
	err := r.db.Table("base_permission_rule_roles").
		Joins("JOIN base_roles ON base_roles.id = base_permission_rule_roles.role_id").
		Where("base_permission_rule_roles.rule_id = ?", ruleID).
		Select("base_roles.*").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRulesByRole 获取角色的所有权限规则
func (r *permissionRuleRepository) GetRulesByRole(roleID int64) ([]*entity.PermissionRule, error) {
	var dbModels []permissionRuleModel
	err := r.db.Table("base_permission_rules").
		Joins("JOIN base_permission_rule_roles ON base_permission_rule_roles.rule_id = base_permission_rules.id").
		Where("base_permission_rule_roles.role_id = ? AND base_permission_rules.status = 1", roleID).
		Select("base_permission_rules.*").
		Find(&dbModels).Error
	if err != nil {
		return nil, err
	}
	rules := make([]*entity.PermissionRule, len(dbModels))
	for i, model := range dbModels {
		rules[i] = model.toDomain()
	}
	return rules, nil
}

// AddUserToRule 添加用户到权限规则
func (r *permissionRuleRepository) AddUserToRule(ruleID, userID int64) error {
	return r.db.Table("base_permission_rule_users").Create(map[string]interface{}{
		"rule_id": ruleID,
		"user_id": userID,
	}).Error
}

// RemoveUserFromRule 从权限规则移除用户
func (r *permissionRuleRepository) RemoveUserFromRule(ruleID, userID int64) error {
	return r.db.Table("base_permission_rule_users").
		Where("rule_id = ? AND user_id = ?", ruleID, userID).
		Delete(nil).Error
}

// GetUsersByRule 获取权限规则的所有用户
func (r *permissionRuleRepository) GetUsersByRule(ruleID int64) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.Table("base_permission_rule_users").
		Joins("JOIN base_users ON base_users.id = base_permission_rule_users.user_id").
		Where("base_permission_rule_users.rule_id = ?", ruleID).
		Select("base_users.*").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetRulesByUser 获取用户的所有权限规则
func (r *permissionRuleRepository) GetRulesByUser(userID int64) ([]*entity.PermissionRule, error) {
	var dbModels []permissionRuleModel
	// 通过用户关联的角色获取规则
	err := r.db.Table("base_permission_rules").
		Joins("JOIN base_permission_rule_roles ON base_permission_rule_roles.rule_id = base_permission_rules.id").
		Joins("JOIN base_role_users ON base_role_users.role_id = base_permission_rule_roles.role_id").
		Where("base_role_users.user_id = ? AND base_permission_rules.status = 1", userID).
		Select("DISTINCT base_permission_rules.*").
		Find(&dbModels).Error
	if err != nil {
		return nil, err
	}

	// 加上直接分配给用户的规则
	var userDbModels []permissionRuleModel
	err = r.db.Table("base_permission_rules").
		Joins("JOIN base_permission_rule_users ON base_permission_rule_users.rule_id = base_permission_rules.id").
		Where("base_permission_rule_users.user_id = ? AND base_permission_rules.status = 1", userID).
		Select("base_permission_rules.*").
		Find(&userDbModels).Error
	if err != nil {
		return nil, err
	}

	// 合并去重
	seen := make(map[int64]bool)
	var rules []*entity.PermissionRule
	for _, model := range append(dbModels, userDbModels...) {
		if !seen[model.ID] {
			seen[model.ID] = true
			rules = append(rules, model.toDomain())
		}
	}
	return rules, nil
}

// FindMatchingRules 查找匹配的权限规则
// 匹配规则：API 路径和方法匹配，且用户属性满足规则条件
// 支持新旧两种条件格式：
//   旧格式：[{"key":"dept_id","value":"1","type":"user"}]
//   新格式：{"operator":"AND","conditions":[{"attribute":{"key":"dept_id","value":"1","type":"user","operator":"EQ"}}]}
func (r *permissionRuleRepository) FindMatchingRules(apiPath, method string, attributes []entity.PermissionAttribute) ([]*entity.PermissionRule, error) {
	// 1. 查询启用的、匹配 API 路径和方法的规则
	var dbModels []permissionRuleModel
	err := r.db.Where("status = 1 AND api_path = ? AND method = ?", apiPath, method).
		Order("sort ASC, id ASC").
		Find(&dbModels).Error
	if err != nil {
		return nil, err
	}

	// 2. 如果没有条件匹配的规则，返回空
	if len(dbModels) == 0 {
		return nil, nil
	}

	// 3. 构建属性 map 用于快速查找
	attrMap := buildAttributeMap(attributes)

	// 4. 解析条件并匹配属性
	var matchedRules []*entity.PermissionRule
	for _, model := range dbModels {
		rule := model.toDomain()

		// 如果规则没有条件，直接匹配（全局规则）
		if rule.Conditions == "" {
			matchedRules = append(matchedRules, rule)
			continue
		}

		// 根据规则类型选择匹配方式
		var matched bool
		if rule.ConditionNode != nil {
			// 新格式：结构化条件节点
			matched = matchConditionNode(rule.ConditionNode, attrMap)
		} else if len(rule.Attributes) > 0 {
			// 旧格式：简单属性数组（AND 逻辑）
			matched = matchSimpleAttributes(rule.Attributes, attrMap)
		} else {
			continue
		}

		if matched {
			matchedRules = append(matchedRules, rule)
		}
	}

	return matchedRules, nil
}

// buildAttributeMap 构建属性 map 用于快速查找
func buildAttributeMap(attributes []entity.PermissionAttribute) map[string]string {
	attrMap := make(map[string]string)
	for _, attr := range attributes {
		key := attr.Type + ":" + attr.Key
		attrMap[key] = attr.Value
	}
	return attrMap
}

// matchSimpleAttributes 检查简单属性数组是否匹配（旧格式，AND 逻辑）
func matchSimpleAttributes(conditions []entity.PermissionAttribute, attrMap map[string]string) bool {
	for _, cond := range conditions {
		key := cond.Type + ":" + cond.Key
		actualValue, exists := attrMap[key]
		if !exists {
			return false
		}
		if !matchOperator(cond.Operator, actualValue, cond.Value) {
			return false
		}
	}
	return true
}

// matchConditionNode 递归匹配结构化条件节点
func matchConditionNode(node *entity.ConditionNode, attrMap map[string]string) bool {
	if node == nil {
		return true
	}

	switch node.Operator {
	case entity.LogicAND:
		// AND：所有子条件必须满足
		for _, subCond := range node.Conditions {
			if !matchConditionNode(&subCond, attrMap) {
				return false
			}
		}
		return true

	case entity.LogicOR:
		// OR：任意一个子条件满足即可
		for _, subCond := range node.Conditions {
			if matchConditionNode(&subCond, attrMap) {
				return true
			}
		}
		return false

	case entity.LogicNOT:
		// NOT：取反子条件的结果
		if len(node.Conditions) > 0 {
			return !matchConditionNode(&node.Conditions[0], attrMap)
		}
		return false

	default:
		// 原子条件（叶子节点）
		if node.Attribute == nil {
			return true
		}
		attr := node.Attribute
		key := attr.Type + ":" + attr.Key
		actualValue, exists := attrMap[key]
		if !exists {
			return false
		}
		result := matchOperator(attr.Operator, actualValue, attr.Value)
		if node.Not {
			return !result
		}
		return result
	}
}

// matchOperator 根据运算符匹配属性值
func matchOperator(op entity.ConditionOperator, actualValue, expectedValue string) bool {
	switch op {
	case entity.OperatorEQ, "":
		// 默认等于
		return actualValue == expectedValue
	case entity.OperatorNEQ:
		return actualValue != expectedValue
	case entity.OperatorIN:
		// 检查 actualValue 是否在 expectedValue 的逗号分隔列表中
		values := strings.Split(expectedValue, ",")
		for _, v := range values {
			if actualValue == strings.TrimSpace(v) {
				return true
			}
		}
		return false
	case entity.OperatorNotIN:
		values := strings.Split(expectedValue, ",")
		for _, v := range values {
			if actualValue == strings.TrimSpace(v) {
				return false
			}
		}
		return true
	case entity.OperatorGT:
		return compareNumeric(actualValue, expectedValue) > 0
	case entity.OperatorGTE:
		return compareNumeric(actualValue, expectedValue) >= 0
	case entity.OperatorLT:
		return compareNumeric(actualValue, expectedValue) < 0
	case entity.OperatorLTE:
		return compareNumeric(actualValue, expectedValue) <= 0
	case entity.OperatorLIKE:
		return strings.Contains(actualValue, expectedValue)
	default:
		return actualValue == expectedValue
	}
}

// compareNumeric 比较两个数字字符串
func compareNumeric(a, b string) int {
	// 尝试解析为 float64
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	if errA == nil && errB == nil {
		if fa > fb {
			return 1
		} else if fa < fb {
			return -1
		}
		return 0
	}
	// 如果无法解析为数字，按字符串比较
	return strings.Compare(a, b)
}
