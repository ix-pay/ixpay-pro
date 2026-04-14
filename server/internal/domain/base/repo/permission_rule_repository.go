package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// PermissionRuleRepository 权限规则仓库接口
type PermissionRuleRepository interface {
	GetByID(id int64) (*entity.PermissionRule, error)
	GetByName(name string) (*entity.PermissionRule, error)
	Create(rule *entity.PermissionRule) error
	Update(rule *entity.PermissionRule) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.PermissionRule, int64, error)
	GetAllRules() ([]*entity.PermissionRule, error)
	GetRulesByStatus(status int) ([]*entity.PermissionRule, error)

	// 关联操作
	AddRoleToRule(ruleID, roleID int64) error
	RemoveRoleFromRule(ruleID, roleID int64) error
	GetRolesByRule(ruleID int64) ([]*entity.Role, error)
	GetRulesByRole(roleID int64) ([]*entity.PermissionRule, error)

	AddUserToRule(ruleID, userID int64) error
	RemoveUserFromRule(ruleID, userID int64) error
	GetUsersByRule(ruleID int64) ([]*entity.User, error)
	GetRulesByUser(userID int64) ([]*entity.PermissionRule, error)

	// 规则匹配
	FindMatchingRules(apiPath, method string, attributes []entity.PermissionAttribute) ([]*entity.PermissionRule, error)
}
