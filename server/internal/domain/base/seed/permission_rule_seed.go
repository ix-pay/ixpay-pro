package seed

import (
	"encoding/json"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// PermissionRuleSeed ABAC 权限规则种子数据
// 用于初始化系统级 ABAC 规则，如部门经理可访问本部门数据等
type PermissionRuleSeed struct {
	ruleRepo repo.PermissionRuleRepository
	roleRepo repo.RoleRepository
}

// NewPermissionRuleSeed 创建 ABAC 权限规则种子数据实例
func NewPermissionRuleSeed(ruleRepo repo.PermissionRuleRepository, roleRepo repo.RoleRepository) Seed {
	return &PermissionRuleSeed{
		ruleRepo: ruleRepo,
		roleRepo: roleRepo,
	}
}

// Version 返回种子数据版本
func (ps *PermissionRuleSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (ps *PermissionRuleSeed) Name() string {
	return "permission_rule_seed"
}

// Order 返回初始化顺序（第九个执行，在菜单-API 关联之后）
func (ps *PermissionRuleSeed) Order() int {
	return 9
}

// Init 初始化 ABAC 权限规则种子数据
func (ps *PermissionRuleSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化 ABAC 权限规则种子数据")

	// 获取 admin 角色
	adminRole, err := ps.roleRepo.GetByCode("admin")
	if err != nil {
		logger.Warn("admin 角色不存在，跳过权限规则关联", "error", err)
	}

	// 定义系统级 ABAC 规则
	rules := []*entity.PermissionRule{
		{
			Name:        "部门经理数据访问权限",
			Description: "部门经理可以访问本部门所有数据",
			Effect:      "allow",
			APIPath:     "/api/admin/**",
			Method:      "GET",
			// 新格式：结构化条件 - 用户部门ID等于1（技术部）且职位为部门经理
			Conditions: `{"operator":"AND","conditions":[{"attribute":{"key":"department_id","value":"1","type":"user","operator":"EQ"}},{"attribute":{"key":"position_id","value":"1","type":"user","operator":"EQ"}}]}`,
			Status:     1,
			Sort:       1,
			IsSystem:   true,
			RoleIds:    []int64{adminRole.ID},
		},
		{
			Name:        "管理员全局API访问",
			Description: "管理员角色拥有所有API访问权限",
			Effect:      "allow",
			APIPath:     "/api/admin/**",
			Method:      "*",
			Conditions:  "", // 无条件，全局规则
			Status:      1,
			Sort:        0,
			IsSystem:    true,
			RoleIds:     []int64{adminRole.ID},
		},
		{
			Name:        "禁止非管理员删除用户",
			Description: "非管理员角色不能执行删除用户的操作",
			Effect:      "deny",
			APIPath:     "/api/admin/user/:id",
			Method:      "DELETE",
			// 旧格式（兼容）：简单属性数组
			Conditions: func() string {
				data, _ := json.Marshal([]entity.PermissionAttribute{
					{Key: "role_code", Value: "admin", Type: "role", Operator: entity.OperatorNEQ},
				})
				return string(data)
			}(),
			Status:   1,
			Sort:     100,
			IsSystem: true,
		},
		{
			Name:        "禁止普通用户访问系统配置",
			Description: "普通用户角色不能访问系统配置模块",
			Effect:      "deny",
			APIPath:     "/api/admin/config/**",
			Method:      "*",
			Conditions: func() string {
				data, _ := json.Marshal([]entity.PermissionAttribute{
					{Key: "role_code", Value: "user", Type: "role", Operator: entity.OperatorEQ},
				})
				return string(data)
			}(),
			Status:   1,
			Sort:     90,
			IsSystem: true,
		},
		{
			Name:        "用户仅可查看自己的数据",
			Description: "普通用户只能查看自己的个人信息",
			Effect:      "allow",
			APIPath:     "/api/admin/user/:id",
			Method:      "GET",
			// 新格式：OR 条件 - 当前用户ID等于请求中的用户ID，或者是管理员
			Conditions: func() string {
				node := entity.ConditionNode{
					Operator: entity.LogicOR,
					Conditions: []entity.ConditionNode{
						{
							Attribute: &entity.PermissionAttribute{
								Key:      "user_id",
								Value:    "{request.user_id}", // 请求路径中的用户ID
								Type:     "request",
								Operator: entity.OperatorEQ,
							},
						},
						{
							Attribute: &entity.PermissionAttribute{
								Key:      "role_code",
								Value:    "admin",
								Type:     "role",
								Operator: entity.OperatorEQ,
							},
						},
					},
				}
				data, _ := json.Marshal(node)
				return string(data)
			}(),
			Status:   1,
			Sort:     50,
			IsSystem: true,
		},
	}

	// 保存规则
	for _, rule := range rules {
		// 检查是否已存在
		existing, err := ps.ruleRepo.GetByName(rule.Name)
		if err == nil {
			logger.Info("ABAC 规则已存在，跳过创建", "name", rule.Name, "id", existing.ID)
			continue
		}

		// 创建规则
		if err := ps.ruleRepo.Create(rule); err != nil {
			logger.Error("创建 ABAC 规则失败", "name", rule.Name, "error", err)
			return err
		}
		logger.Info("创建 ABAC 规则成功", "name", rule.Name, "id", rule.ID)

		// 关联角色
		for _, roleID := range rule.RoleIds {
			if err := ps.ruleRepo.AddRoleToRule(rule.ID, roleID); err != nil {
				logger.Error("关联 ABAC 规则到角色失败", "ruleName", rule.Name, "roleID", roleID, "error", err)
				return err
			}
			logger.Info("ABAC 规则已关联到角色", "ruleName", rule.Name, "roleID", roleID)
		}
	}

	logger.Info("ABAC 权限规则种子数据初始化完成")
	return nil
}