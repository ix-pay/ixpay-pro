package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// PermissionRuleRepository 权限规则仓库接口
type PermissionRuleRepository interface {
	GetByID(id string) (*entity.PermissionRule, error)
	GetByName(name string) (*entity.PermissionRule, error)
	Create(rule *entity.PermissionRule) error
	Update(rule *entity.PermissionRule) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.PermissionRule, int64, error)
	GetAllRules() ([]*entity.PermissionRule, error)
	GetRulesByStatus(status int) ([]*entity.PermissionRule, error)

	// 关联操作
	AddRoleToRule(ruleID, roleID string) error
	RemoveRoleFromRule(ruleID, roleID string) error
	GetRolesByRule(ruleID string) ([]*entity.Role, error)
	GetRulesByRole(roleID string) ([]*entity.PermissionRule, error)

	AddUserToRule(ruleID, userID string) error
	RemoveUserFromRule(ruleID, userID string) error
	GetUsersByRule(ruleID string) ([]*entity.User, error)
	GetRulesByUser(userID string) ([]*entity.PermissionRule, error)

	// 规则匹配
	FindMatchingRules(apiPath, method string, attributes []entity.PermissionAttribute) ([]*entity.PermissionRule, error)
}
