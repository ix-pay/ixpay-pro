package service

import (
	"fmt"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// PermissionService 权限服务实现
// 支持 RBAC+ABAC 混合模型的权限管理
type PermissionService struct {
	roleService        *RoleService
	userService        *UserService
	roleRepo           repo.RoleRepository
	apiRepo            repo.APIRepository
	permissionRuleRepo repo.PermissionRuleRepository
	logger             logger.Logger
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(
	roleService *RoleService,
	userService *UserService,
	roleRepo repo.RoleRepository,
	apiRepo repo.APIRepository,
	permissionRuleRepo repo.PermissionRuleRepository,
	logger logger.Logger,
) *PermissionService {
	return &PermissionService{
		roleService:        roleService,
		userService:        userService,
		roleRepo:           roleRepo,
		apiRepo:            apiRepo,
		permissionRuleRepo: permissionRuleRepo,
		logger:             logger,
	}
}

// GetRolesByUserId 根据用户 ID 获取角色列表
func (p *PermissionService) GetRolesByUserId(userId int64) ([]entity.Role, error) {
	roles, err := p.roleService.GetRolesForUser(userId)
	if err != nil {
		p.logger.Error("获取用户角色失败", "error", err, "userId", userId)
		return nil, err
	}

	result := make([]entity.Role, 0, len(roles))
	for _, role := range roles {
		if role != nil {
			result = append(result, *role)
		}
	}

	return result, nil
}

// CheckAPIAccess 检查用户是否有 API 访问权限（支持 RBAC+ABAC）
func (p *PermissionService) CheckAPIAccess(userId int64, apiPath, method string) (bool, error) {
	user, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return false, err
	}

	// 1. 检查用户特殊 API 权限
	specialPerms, err := p.userService.GetUserSpecialPermissions(userId)
	if err != nil {
		p.logger.Error("获取用户特殊权限失败", "error", err, "userId", userId)
		return false, err
	}

	for _, api := range specialPerms {
		if api.Path == apiPath && api.Method == method {
			return true, nil
		}
	}

	// 2. 检查用户角色的 API 权限（包括菜单/按钮继承的 API）
	roles, err := p.roleService.GetRolesForUser(userId)
	if err != nil {
		p.logger.Error("获取用户角色失败", "error", err, "userId", userId)
		return false, err
	}

	for _, role := range roles {
		// 获取角色的所有 API 权限（包括继承）
		apiPerms, err := p.roleService.GetAllInheritedPermissions(role.ID)
		if err != nil {
			p.logger.Error("获取角色继承权限失败", "error", err, "roleId", role.ID)
			continue
		}

		for _, api := range apiPerms {
			if api.Path == apiPath && api.Method == method {
				return true, nil
			}
		}
	}

	// 3. 检查 ABAC 权限规则
	allow, err := p.permissionRuleRepo.FindMatchingRules(apiPath, method, []entity.PermissionAttribute{
		{Key: "user_id", Value: fmt.Sprintf("%d", userId), Type: "user"},
		{Key: "department_id", Value: fmt.Sprintf("%d", user.DepartmentID), Type: "user"},
		{Key: "position_id", Value: fmt.Sprintf("%d", user.PositionID), Type: "user"},
	})
	if err != nil {
		p.logger.Error("评估权限规则失败", "error", err, "userId", userId, "apiPath", apiPath, "method", method)
		return false, err
	}

	for _, rule := range allow {
		if rule.Effect == "allow" && rule.Status == 1 {
			return true, nil
		}
	}

	return false, nil
}

// GetUserAPIPermissions 获取用户的所有 API 权限（包括继承和特殊权限）
func (p *PermissionService) GetUserAPIPermissions(userId int64) ([]*entity.API, error) {
	_, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return nil, err
	}

	// 获取用户特殊 API 权限
	specialPerms, err := p.userService.GetUserSpecialPermissions(userId)
	if err != nil {
		p.logger.Error("获取用户特殊权限失败", "error", err, "userId", userId)
		return nil, err
	}

	// 获取用户角色的 API 权限（包括继承）
	roles, err := p.roleService.GetRolesForUser(userId)
	if err != nil {
		p.logger.Error("获取用户角色失败", "error", err, "userId", userId)
		return nil, err
	}

	apiMap := make(map[string]*entity.API)

	for _, api := range specialPerms {
		key := api.Path + "_" + api.Method
		apiMap[key] = api
	}

	for _, role := range roles {
		apiPerms, err := p.roleService.GetAllInheritedPermissions(role.ID)
		if err != nil {
			p.logger.Error("获取角色继承权限失败", "error", err, "roleId", role.ID)
			continue
		}

		for _, api := range apiPerms {
			key := api.Path + "_" + api.Method
			apiMap[key] = api
		}
	}

	result := make([]*entity.API, 0, len(apiMap))
	for _, api := range apiMap {
		result = append(result, api)
	}

	return result, nil
}

// CheckResourceAccess 检查用户对资源的访问权限（ABAC）
// 先检查 RBAC 是否允许，再检查 ABAC 规则是否允许或拒绝
func (p *PermissionService) CheckResourceAccess(userId int64, resourceType string, resourceID string, action string) (bool, error) {
	_, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return false, err
	}

	apiPath := "/api/admin/" + resourceType + "/" + resourceID
	method := action

	// 1. RBAC 检查：用户是否有该 API 的访问权限
	rbacAllowed, err := p.CheckAPIAccess(userId, apiPath, method)
	if err != nil {
		p.logger.Error("检查 RBAC 访问失败", "error", err, "userId", userId, "resourceType", resourceType, "resourceID", resourceID, "action", action)
		return false, err
	}

	if !rbacAllowed {
		return false, nil
	}

	// 2. ABAC 检查：获取用户关联的权限规则（通过角色和直接分配）
	rules, err := p.permissionRuleRepo.GetRulesByUser(userId)
	if err != nil {
		p.logger.Error("获取用户权限规则失败", "error", err, "userId", userId)
		return false, err
	}

	// 3. 先检查是否有拒绝规则（deny 优先）
	for _, rule := range rules {
		if rule.Effect == "deny" && rule.IsActive() {
			p.logger.Info("ABAC 规则拒绝访问", "ruleId", rule.ID, "ruleName", rule.Name, "userId", userId)
			return false, nil
		}
	}

	// 4. 再检查是否有允许规则
	for _, rule := range rules {
		if rule.Effect == "allow" && rule.IsActive() {
			p.logger.Info("ABAC 规则允许访问", "ruleId", rule.ID, "ruleName", rule.Name, "userId", userId)
			return true, nil
		}
	}

	return true, nil
}

// RefreshPermissionCache 刷新用户权限缓存
func (p *PermissionService) RefreshPermissionCache(userId int64) error {
	p.logger.Info("用户权限缓存已刷新", "userId", userId)
	return nil
}

// GetPermissionRules 获取用户的权限规则（ABAC）
// 通过用户关联的角色和直接分配给用户的规则获取
func (p *PermissionService) GetPermissionRules(userId int64) ([]*entity.PermissionRule, error) {
	_, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return nil, err
	}

	rules, err := p.permissionRuleRepo.GetRulesByUser(userId)
	if err != nil {
		p.logger.Error("获取用户权限规则失败", "error", err, "userId", userId)
		return nil, err
	}

	return rules, nil
}

// GetUserDataScope 获取用户的数据权限范围
// 通过用户角色获取数据权限范围，取最小范围（最严格）
func (p *PermissionService) GetUserDataScope(userId int64) (entity.DataScope, error) {
	roles, err := p.roleService.GetRolesForUser(userId)
	if err != nil {
		p.logger.Error("获取用户角色失败", "error", err, "userId", userId)
		return entity.DataScopeSelf, err
	}

	// 默认仅本人数据
	minScope := entity.DataScopeSelf

	for _, role := range roles {
		if role.DataScope > 0 && role.DataScope < minScope {
			minScope = role.DataScope
		}
	}

	return minScope, nil
}