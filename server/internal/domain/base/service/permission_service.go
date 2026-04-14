package service

import (
	"errors"
	"fmt"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// PermissionService 权限服务实现
// 支持RBAC+ABAC混合模型的权限管理
type PermissionService struct {
	roleService         *RoleService
	userService         *UserService
	roleRepo            repo.RoleRepository
	btnPermRepo         repo.BtnPermRepository
	apiRepo             repo.APIRepository
	permissionRuleRepo  repo.PermissionRuleRepository
	permissionGroupRepo repo.PermissionGroupRepository
	logger              logger.Logger
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(
	roleService *RoleService,
	userService *UserService,
	roleRepo repo.RoleRepository,
	btnPermRepo repo.BtnPermRepository,
	apiRepo repo.APIRepository,
	permissionRuleRepo repo.PermissionRuleRepository,
	permissionGroupRepo repo.PermissionGroupRepository,
	logger logger.Logger,
) *PermissionService {
	return &PermissionService{
		roleService:         roleService,
		userService:         userService,
		roleRepo:            roleRepo,
		btnPermRepo:         btnPermRepo,
		apiRepo:             apiRepo,
		permissionRuleRepo:  permissionRuleRepo,
		permissionGroupRepo: permissionGroupRepo,
		logger:              logger,
	}
}

// GetRolesByUserId 根据用户 ID 获取角色列表
func (p *PermissionService) GetRolesByUserId(userId int64) ([]entity.Role, error) {
	// 使用 RoleService 获取用户的角色列表
	roles, err := p.roleService.GetRolesForUser(userId)
	if err != nil {
		p.logger.Error("获取用户角色失败", "error", err, "userId", userId)
		return nil, err
	}

	// 将 []*entity.Role 转换为 []entity.Role
	result := make([]entity.Role, 0, len(roles))
	for _, role := range roles {
		if role != nil {
			result = append(result, *role)
		}
	}

	return result, nil
}

// GetBtnPermsByRole 根据角色 ID 获取按钮权限列表
func (p *PermissionService) GetBtnPermsByRole(roleId int64) ([]*entity.BtnPerm, error) {
	return p.roleRepo.GetBtnPermsByRole(roleId)
}

// AssignBtnPermToRole 为角色分配按钮权限
func (p *PermissionService) AssignBtnPermToRole(roleId int64, btnPermIds []int64) error {
	// 检查角色是否存在
	role, err := p.roleService.GetRoleByID(roleId)
	if err != nil {
		p.logger.Error("角色不存在", "error", err, "roleId", roleId)
		return err
	}

	for _, btnId := range btnPermIds {
		// 检查按钮权限是否存在
		button, err := p.btnPermRepo.GetByID(btnId)
		if err != nil {
			p.logger.Error("获取按钮权限失败", "error", err, "btnId", btnId)
			return err
		}

		// 检查按钮权限是否已在角色中
		buttons, err := p.roleRepo.GetBtnPermsByRole(roleId)
		if err != nil {
			p.logger.Error("获取角色按钮权限失败", "error", err, "roleId", roleId)
			return err
		}

		buttonExists := false
		for _, b := range buttons {
			if b.ID == btnId {
				buttonExists = true
				break
			}
		}

		if !buttonExists {
			if err := p.roleRepo.AddBtnPermToRole(roleId, btnId); err != nil {
				p.logger.Error("为角色分配按钮权限失败", "error", err, "roleId", roleId, "btnId", btnId)
				return err
			}
			p.logger.Info("按钮权限分配成功", "roleId", roleId, "roleName", role.Name, "btnId", btnId, "btnName", button.Name)
		}
	}
	return nil
}

// RevokeBtnPermFromRole 从角色撤销按钮权限
func (p *PermissionService) RevokeBtnPermFromRole(roleId int64, btnPermId int64) error {
	// 检查角色是否存在
	role, err := p.roleService.GetRoleByID(roleId)
	if err != nil {
		p.logger.Error("角色不存在", "error", err, "roleId", roleId)
		return err
	}

	// 检查按钮权限是否在角色中
	buttons, err := p.roleRepo.GetBtnPermsByRole(roleId)
	if err != nil {
		p.logger.Error("获取角色按钮权限失败", "error", err, "roleId", roleId)
		return err
	}

	buttonExists := false
	for _, b := range buttons {
		if b.ID == btnPermId {
			buttonExists = true
			break
		}
	}

	if !buttonExists {
		p.logger.Error("角色中不存在按钮权限", "roleId", roleId, "btnId", btnPermId)
		return errors.New("按钮权限不在角色中")
	}

	if err := p.roleRepo.RemoveBtnPermFromRole(roleId, btnPermId); err != nil {
		p.logger.Error("从角色撤销按钮权限失败", "error", err, "roleId", roleId, "btnId", btnPermId)
		return err
	}

	p.logger.Info("按钮权限撤销成功", "roleId", roleId, "roleName", role.Name, "btnId", btnPermId)
	return nil
}

// CheckAPIAccess 检查用户是否有 API 访问权限（支持 RBAC+ABAC）
func (p *PermissionService) CheckAPIAccess(userId int64, apiPath, method string) (bool, error) {
	user, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return false, err
	}

	// 2. 检查用户特殊 API 权限
	specialPerms, err := p.userService.GetUserSpecialPermissions(userId)
	if err != nil {
		p.logger.Error("获取用户特殊权限失败", "error", err, "userId", userId)
		return false, err
	}

	// 检查是否有直接匹配的特殊权限
	for _, api := range specialPerms {
		if api.Path == apiPath && api.Method == method {
			return true, nil
		}
	}

	// 3. 检查用户角色的 API 权限（包括继承和权限组）
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

		// 检查 API 权限
		for _, api := range apiPerms {
			if api.Path == apiPath && api.Method == method {
				return true, nil
			}
		}

		// 获取角色的权限组
		groups, err := p.permissionGroupRepo.GetGroupsByRole(role.ID)
		if err != nil {
			p.logger.Error("获取角色权限组失败", "error", err, "roleId", role.ID)
			continue
		}

		// 检查权限组的 API 权限
		for _, group := range groups {
			groupAPIs, err := p.permissionGroupRepo.GetAPIsByGroup(group.ID)
			if err != nil {
				p.logger.Error("获取权限组 API 失败", "error", err, "groupId", group.ID)
				continue
			}

			for _, api := range groupAPIs {
				if api.Path == apiPath && api.Method == method {
					return true, nil
				}
			}
		}
	}

	// 4. 检查 ABAC 权限规则
	// 评估规则
	allow, err := p.permissionRuleRepo.FindMatchingRules(apiPath, method, []entity.PermissionAttribute{
		{Key: "user_id", Value: fmt.Sprintf("%d", userId), Type: "user"},
		{Key: "department_id", Value: fmt.Sprintf("%d", user.DepartmentID), Type: "user"},
		{Key: "position_id", Value: fmt.Sprintf("%d", user.PositionID), Type: "user"},
	})
	if err != nil {
		p.logger.Error("评估权限规则失败", "error", err, "userId", userId, "apiPath", apiPath, "method", method)
		return false, err
	}

	// 如果有匹配的规则，检查是否有允许的规则
	for _, rule := range allow {
		if rule.Effect == "allow" && rule.Status == 1 {
			return true, nil
		}
	}

	return false, nil
}

// CheckBtnPermission 检查用户是否有按钮权限（支持 RBAC+ABAC）
func (p *PermissionService) CheckBtnPermission(userId int64, btnPermCode string) (bool, error) {
	// 1. 检查用户是否存在
	_, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return false, err
	}

	// 2. 检查用户特殊按钮权限
	specialBtnPerms, err := p.userService.GetUserSpecialBtnPermissions(userId)
	if err != nil {
		p.logger.Error("获取用户特殊按钮权限失败", "error", err, "userId", userId)
		return false, err
	}

	// 检查是否有直接匹配的特殊按钮权限
	for _, btn := range specialBtnPerms {
		if btn.Code == btnPermCode {
			return true, nil
		}
	}

	// 3. 检查用户角色的按钮权限（包括继承）
	roles, err := p.roleService.GetRolesForUser(userId)
	if err != nil {
		p.logger.Error("获取用户角色失败", "error", err, "userId", userId)
		return false, err
	}

	for _, role := range roles {
		// 获取角色的所有按钮权限（包括继承）
		btnPerms, err := p.roleService.GetAllInheritedBtnPerms(role.ID)
		if err != nil {
			p.logger.Error("获取角色继承按钮权限失败", "error", err, "roleId", role.ID)
			continue
		}

		// 检查按钮权限
		for _, btn := range btnPerms {
			if btn.Code == btnPermCode {
				return true, nil
			}
		}

		// 获取角色的权限组
		groups, err := p.permissionGroupRepo.GetGroupsByRole(role.ID)
		if err != nil {
			p.logger.Error("获取角色权限组失败", "error", err, "roleId", role.ID)
			continue
		}

		// 检查权限组的按钮权限
		for _, group := range groups {
			groupBtnPerms, err := p.permissionGroupRepo.GetBtnPermsByGroup(group.ID)
			if err != nil {
				p.logger.Error("获取权限组按钮权限失败", "error", err, "groupId", group.ID)
				continue
			}

			for _, btn := range groupBtnPerms {
				if btn.Code == btnPermCode {
					return true, nil
				}
			}
		}
	}

	// 4. TODO: 检查ABAC权限规则
	// 暂时返回false，因为ABAC规则的按钮权限需要进一步设计

	return false, nil
}

// GetUserAPIPermissions 获取用户的所有 API 权限（包括继承和特殊权限）
func (p *PermissionService) GetUserAPIPermissions(userId int64) ([]*entity.API, error) {
	// 1. 检查用户是否存在
	_, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return nil, err
	}

	// 2. 获取用户特殊 API 权限
	specialPerms, err := p.userService.GetUserSpecialPermissions(userId)
	if err != nil {
		p.logger.Error("获取用户特殊权限失败", "error", err, "userId", userId)
		return nil, err
	}

	// 3. 获取用户角色的 API 权限（包括继承）
	roles, err := p.roleService.GetRolesForUser(userId)
	if err != nil {
		p.logger.Error("获取用户角色失败", "error", err, "userId", userId)
		return nil, err
	}

	// 使用 map 去重
	apiMap := make(map[string]*entity.API)

	// 添加特殊权限到结果
	for _, api := range specialPerms {
		key := api.Path + "_" + api.Method
		apiMap[key] = api
	}

	// 添加角色权限到结果
	for _, role := range roles {
		// 获取角色的所有 API 权限（包括继承）
		apiPerms, err := p.roleService.GetAllInheritedPermissions(role.ID)
		if err != nil {
			p.logger.Error("获取角色继承权限失败", "error", err, "roleId", role.ID)
			continue
		}

		for _, api := range apiPerms {
			key := api.Path + "_" + api.Method
			apiMap[key] = api
		}

		// 获取角色的权限组
		groups, err := p.permissionGroupRepo.GetGroupsByRole(role.ID)
		if err != nil {
			p.logger.Error("获取角色权限组失败", "error", err, "roleId", role.ID)
			continue
		}

		// 添加权限组的 API 权限到结果
		for _, group := range groups {
			groupAPIs, err := p.permissionGroupRepo.GetAPIsByGroup(group.ID)
			if err != nil {
				p.logger.Error("获取权限组 API 失败", "error", err, "groupId", group.ID)
				continue
			}

			for _, api := range groupAPIs {
				key := api.Path + "_" + api.Method
				apiMap[key] = api
			}
		}
	}

	// 转换为切片
	result := make([]*entity.API, 0, len(apiMap))
	for _, api := range apiMap {
		result = append(result, api)
	}

	return result, nil
}

// GetUserBtnPermissions 获取用户的所有按钮权限（包括继承和特殊权限）
func (p *PermissionService) GetUserBtnPermissions(userId int64) ([]*entity.BtnPerm, error) {
	// 1. 检查用户是否存在
	_, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return nil, err
	}

	// 2. 获取用户特殊按钮权限
	specialBtnPerms, err := p.userService.GetUserSpecialBtnPermissions(userId)
	if err != nil {
		p.logger.Error("获取用户特殊按钮权限失败", "error", err, "userId", userId)
		return nil, err
	}

	// 3. 获取用户角色的按钮权限（包括继承）
	roles, err := p.roleService.GetRolesForUser(userId)
	if err != nil {
		p.logger.Error("获取用户角色失败", "error", err, "userId", userId)
		return nil, err
	}

	// 使用 map 去重
	btnMap := make(map[string]*entity.BtnPerm)

	// 添加特殊权限到结果
	for _, btn := range specialBtnPerms {
		btnMap[btn.Code] = btn
	}

	// 添加角色权限到结果
	for _, role := range roles {
		// 获取角色的所有按钮权限（包括继承）
		btnPerms, err := p.roleService.GetAllInheritedBtnPerms(role.ID)
		if err != nil {
			p.logger.Error("获取角色继承按钮权限失败", "error", err, "roleId", role.ID)
			continue
		}

		for _, btn := range btnPerms {
			btnMap[btn.Code] = btn
		}

		// 获取角色的权限组
		groups, err := p.permissionGroupRepo.GetGroupsByRole(role.ID)
		if err != nil {
			p.logger.Error("获取角色权限组失败", "error", err, "roleId", role.ID)
			continue
		}

		// 添加权限组的按钮权限到结果
		for _, group := range groups {
			groupBtnPerms, err := p.permissionGroupRepo.GetBtnPermsByGroup(group.ID)
			if err != nil {
				p.logger.Error("获取权限组按钮权限失败", "error", err, "groupId", group.ID)
				continue
			}

			for _, btn := range groupBtnPerms {
				btnMap[btn.Code] = btn
			}
		}
	}

	// 转换为切片
	result := make([]*entity.BtnPerm, 0, len(btnMap))
	for _, btn := range btnMap {
		result = append(result, btn)
	}

	return result, nil
}

// CheckResourceAccess 检查用户对资源的访问权限（ABAC）
func (p *PermissionService) CheckResourceAccess(userId int64, resourceType string, resourceID string, action string) (bool, error) {
	// 1. 检查用户是否存在
	_, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return false, err
	}

	// 2. 构建 API 路径和方法（资源访问可以映射为 API）
	apiPath := "/api//" + resourceType + "/" + resourceID
	method := action

	// 3. 检查 RBAC 权限
	// 调用 CheckAPIAccess 检查基本访问权限
	rbacAllowed, err := p.CheckAPIAccess(userId, apiPath, method)
	if err != nil {
		p.logger.Error("检查 RBAC 访问失败", "error", err, "userId", userId, "resourceType", resourceType, "resourceID", resourceID, "action", action)
		return false, err
	}

	if !rbacAllowed {
		return false, nil
	}

	// 4. 检查 ABAC 规则
	// 构建用户属性 - 暂时注释，等待实现完整的 ABAC 规则匹配逻辑
	// attributes := []repo.PermissionAttribute{
	// 	{Key: "user_id", Value: strconv.FormatInt(userId, 10), Type: "user"},
	// 	{Key: "department_id", Value: strconv.FormatInt(user.DepartmentID, 10), Type: "user"},
	// 	{Key: "position_id", Value: strconv.FormatInt(user.PositionID, 10), Type: "user"},
	// 	{Key: "resource_type", Value: resourceType, Type: "resource"},
	// 	{Key: "resource_id", Value: strconv.FormatInt(resourceID, 10), Type: "resource"},
	// 	{Key: "action", Value: action, Type: "environment"},
	// }

	// 评估 ABAC 规则
	rules, err := p.permissionRuleRepo.GetRulesByRole(userId)
	if err != nil {
		p.logger.Error("获取用户权限规则失败", "error", err, "userId", userId)
		return false, err
	}

	// 检查是否有明确的拒绝规则
	for _, rule := range rules {
		if rule.Effect == "deny" && rule.Status == 1 {
			// TODO: 实现规则条件匹配逻辑
			// 这里简化处理，实际应该解析 rule.Conditions 并匹配 attributes
			p.logger.Info("ABAC 规则拒绝访问", "ruleId", rule.ID, "ruleName", rule.Name, "userId", userId)
			return false, nil
		}
	}

	// 检查是否有明确的允许规则
	for _, rule := range rules {
		if rule.Effect == "allow" && rule.Status == 1 {
			// TODO: 实现规则条件匹配逻辑
			// 这里简化处理，实际应该解析 rule.Conditions 并匹配 attributes
			p.logger.Info("ABAC 规则允许访问", "ruleId", rule.ID, "ruleName", rule.Name, "userId", userId)
			return true, nil
		}
	}

	// 如果没有 ABAC 规则，则默认允许（基于 RBAC 结果）
	return true, nil
}

// RefreshPermissionCache 刷新用户权限缓存
func (p *PermissionService) RefreshPermissionCache(userId int64) error {
	// TODO: 实现权限缓存刷新逻辑
	// 例如：清除 Redis 中的用户权限缓存
	p.logger.Info("用户权限缓存已刷新", "userId", userId)
	return nil
}

// GetPermissionRules 获取用户的权限规则（ABAC）
func (p *PermissionService) GetPermissionRules(userId int64) ([]*entity.PermissionRule, error) {
	// 检查用户是否存在
	_, err := p.userService.GetUserInfo(userId)
	if err != nil {
		p.logger.Error("用户不存在", "error", err, "userId", userId)
		return nil, err
	}

	// 获取用户的权限规则
	rules, err := p.permissionRuleRepo.GetRulesByRole(userId)
	if err != nil {
		p.logger.Error("获取用户权限规则失败", "error", err, "userId", userId)
		return nil, err
	}

	return rules, nil
}
