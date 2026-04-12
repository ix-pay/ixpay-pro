package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// RoleService 角色服务实现
type RoleService struct {
	roleRepo            repo.RoleRepository
	userRepo            repo.UserRepository
	menuRepo            repo.MenuRepository
	apiRepo             repo.APIRepository
	btnPermRepo         repo.BtnPermRepository
	permissionGroupRepo repo.PermissionGroupRepository
	log                 logger.Logger
}

// NewRoleService 创建角色服务实例
func NewRoleService(roleRepo repo.RoleRepository, userRepo repo.UserRepository, menuRepo repo.MenuRepository, apiRepo repo.APIRepository, btnPermRepo repo.BtnPermRepository, permissionGroupRepo repo.PermissionGroupRepository, log logger.Logger) *RoleService {
	return &RoleService{
		roleRepo:            roleRepo,
		userRepo:            userRepo,
		menuRepo:            menuRepo,
		apiRepo:             apiRepo,
		btnPermRepo:         btnPermRepo,
		permissionGroupRepo: permissionGroupRepo,
		log:                 log,
	}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(name, code, description, parentID string, roleType int64, createdBy string, status, sort int, isSystem bool) (*entity.Role, error) {
	// 检查角色名称是否已存在
	existingRole, err := s.roleRepo.GetByName(name)
	if err == nil && existingRole != nil {
		s.log.Error("角色名称已存在", "name", name)
		return nil, errors.New("角色名称已存在")
	}

	// 检查角色编码是否已存在
	existingRole, err = s.roleRepo.GetByCode(code)
	if err == nil && existingRole != nil {
		s.log.Error("角色编码已存在", "code", code)
		return nil, errors.New("角色编码已存在")
	}

	role := &entity.Role{
		Name:        name,
		Code:        code,
		Description: description,
		Type:        int(roleType),
		ParentID:    parentID,
		Status:      status,
		Sort:        sort,
		IsSystem:    isSystem,
	}

	if err := s.roleRepo.Create(role); err != nil {
		s.log.Error("创建角色失败", "error", err, "name", name, "code", code)
		return nil, err
	}

	s.log.Info("创建角色成功", "id", role.ID, "name", name, "code", code)
	return role, nil
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(id, name, code, description, parentID string, roleType int64, updatedBy string, status, sort int, isSystem bool) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "id", id)
		return errors.New("角色不存在")
	}

	// 保护系统角色，不允许修改
	if role.IsSystem {
		s.log.Error("系统角色不允许修改", "id", id, "name", role.Name)
		return errors.New("系统角色不允许修改")
	}

	// 检查角色名称是否已存在（排除当前角色）
	existingRole, err := s.roleRepo.GetByName(name)
	if err == nil && existingRole != nil && existingRole.ID != id {
		s.log.Error("角色名称已存在", "name", name, "existing_id", existingRole.ID)
		return errors.New("角色名称已存在")
	}

	// 检查角色编码是否已存在（排除当前角色）
	existingRole, err = s.roleRepo.GetByCode(code)
	if err == nil && existingRole != nil && existingRole.ID != id {
		s.log.Error("角色编码已存在", "code", code, "existing_id", existingRole.ID)
		return errors.New("角色编码已存在")
	}

	role.Name = name
	role.Code = code
	role.Description = description
	role.Type = int(roleType)
	role.ParentID = parentID
	role.Status = status
	role.Sort = sort
	role.IsSystem = isSystem
	role.UpdatedBy = updatedBy

	if err := s.roleRepo.Update(role); err != nil {
		s.log.Error("更新角色失败", "error", err, "id", id)
		return err
	}

	s.log.Info("更新角色成功", "id", id, "name", name, "code", code)
	return nil
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(id string) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "id", id)
		return errors.New("角色不存在")
	}

	// 保护系统角色，不允许删除
	if role.IsSystem {
		s.log.Error("系统角色不允许删除", "id", id, "name", role.Name)
		return errors.New("系统角色不允许删除")
	}

	// 检查是否有用户关联到该角色
	users, err := s.roleRepo.GetUsersByRole(id)
	if err != nil {
		s.log.Error("获取角色关联用户失败", "error", err, "id", id)
		return err
	}
	if len(users) > 0 {
		s.log.Error("角色关联了用户，无法删除", "id", id, "user_count", len(users))
		return errors.New("角色关联了用户，无法删除")
	}

	if err := s.roleRepo.Delete(id); err != nil {
		s.log.Error("删除角色失败", "error", err, "id", id)
		return err
	}

	s.log.Info("删除角色成功", "id", id, "name", role.Name)
	return nil
}

// GetRoleByID 根据 ID 获取角色
func (s *RoleService) GetRoleByID(id string) (*entity.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "id", id)
		return nil, errors.New("角色不存在")
	}

	// 预加载关联数据
	// TODO: 需要实现 Role 实体的关联字段
	// 预加载用户
	users, err := s.roleRepo.GetUsersByRole(id)
	if err != nil {
		s.log.Error("获取角色关联用户失败", "error", err, "id", id)
	}
	_ = users

	// 预加载菜单
	menus, err := s.roleRepo.GetMenusByRole(id)
	if err != nil {
		s.log.Error("获取角色关联菜单失败", "error", err, "id", id)
	}
	_ = menus

	// 预加载接口路由
	routes, err := s.roleRepo.GetsByRole(id)
	if err != nil {
		s.log.Error("获取角色关联接口路由失败", "error", err, "id", id)
	}
	_ = routes

	return role, nil
}

// GetRoleByName 根据名称获取角色
func (s *RoleService) GetRoleByName(name string) (*entity.Role, error) {
	role, err := s.roleRepo.GetByName(name)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "name", name)
		return nil, errors.New("角色不存在")
	}
	return role, nil
}

// GetRoleByCode 根据编码获取角色
func (s *RoleService) GetRoleByCode(code string) (*entity.Role, error) {
	role, err := s.roleRepo.GetByCode(code)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "code", code)
		return nil, errors.New("角色不存在")
	}
	return role, nil
}

// GetRoleList 获取角色列表
func (s *RoleService) GetRoleList(page, pageSize int, filters map[string]interface{}) ([]*entity.Role, int64, error) {
	roles, total, err := s.roleRepo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取角色列表失败", "error", err)
		return nil, 0, err
	}
	return roles, total, nil
}

// GetAllRoles 获取所有角色
func (s *RoleService) GetAllRoles() ([]*entity.Role, error) {
	roles, err := s.roleRepo.GetAllRoles()
	if err != nil {
		s.log.Error("获取所有角色失败", "error", err)
		return nil, err
	}
	return roles, nil
}

// AssignUserToRole 分配用户到角色
func (s *RoleService) AssignUserToRole(roleID, userID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查用户是否已在角色中
	users, err := s.roleRepo.GetUsersByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联用户失败", "error", err, "role_id", roleID)
		return err
	}

	for _, user := range users {
		if user.ID == userID {
			s.log.Info("用户已在角色中，跳过分配", "role_id", roleID, "role_name", role.Name, "user_id", userID)
			return nil // 用户已在角色中，直接返回成功
		}
	}

	if err := s.roleRepo.AddUserToRole(roleID, userID); err != nil {
		s.log.Error("分配用户到角色失败", "error", err, "role_id", roleID, "user_id", userID)
		return err
	}

	s.log.Info("分配用户到角色成功", "role_id", roleID, "role_name", role.Name, "user_id", userID, "is_system", role.IsSystem)
	return nil
}

// RevokeUserFromRole 从角色中撤销用户
func (s *RoleService) RevokeUserFromRole(roleID string, userID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查用户是否在角色中
	users, err := s.roleRepo.GetUsersByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联用户失败", "error", err, "role_id", roleID)
		return err
	}

	userExists := false
	for _, user := range users {
		if user.ID == userID {
			userExists = true
			break
		}
	}

	if !userExists {
		s.log.Info("用户不在角色中，无需撤销", "role_id", roleID, "role_name", role.Name, "user_id", userID)
		return nil // 用户不在角色中，直接返回成功
	}

	if err := s.roleRepo.RemoveUserFromRole(roleID, userID); err != nil {
		s.log.Error("从角色中撤销用户失败", "error", err, "role_id", roleID, "user_id", userID)
		return err
	}

	s.log.Info("从角色中撤销用户成功", "role_id", roleID, "role_name", role.Name, "user_id", userID, "is_system", role.IsSystem)
	return nil
}

// GetUsersForRole 获取角色的所有用户
func (s *RoleService) GetUsersForRole(roleID string) ([]*entity.User, error) {
	users, err := s.roleRepo.GetUsersByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联用户失败", "error", err, "role_id", roleID)
		return nil, err
	}
	return users, nil
}

// GetRolesForUser 获取用户的所有角色
func (s *RoleService) GetRolesForUser(userID string) ([]*entity.Role, error) {
	roles, err := s.roleRepo.GetRolesByUser(userID)
	if err != nil {
		s.log.Error("获取用户关联角色失败", "error", err, "user_id", userID)
		return nil, err
	}
	return roles, nil
}

// AssignMenuToRole 分配菜单到角色
func (s *RoleService) AssignMenuToRole(roleID string, menuID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 保护系统角色，不允许修改权限
	if role.IsSystem {
		s.log.Error("系统角色不允许修改权限", "role_id", roleID, "role_name", role.Name)
		return errors.New("系统角色不允许修改权限")
	}

	// 检查菜单是否已在角色中
	menus, err := s.roleRepo.GetMenusByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联菜单失败", "error", err, "role_id", roleID)
		return err
	}

	for _, menu := range menus {
		if menu.ID == menuID {
			s.log.Error("菜单已在角色中", "role_id", roleID, "menu_id", menuID)
			return errors.New("菜单已在角色中")
		}
	}

	// 获取菜单详情，用于匹配 API 路由
	menu, err := s.menuRepo.GetByID(menuID)
	if err != nil {
		s.log.Error("获取菜单详情失败", "error", err, "menu_id", menuID)
		return err
	}

	// 分配菜单到角色
	if err = s.roleRepo.AddMenuToRole(roleID, menuID); err != nil {
		s.log.Error("分配菜单到角色失败", "error", err, "role_id", roleID, "menu_id", menuID)
		return err
	}

	// 自动分配与菜单相关的 API 路由
	routes, err := s.apiRepo.GetAllRoutes()
	if err != nil {
		s.log.Error("获取所有 API 路由失败", "error", err)
		// 不阻止菜单分配，但记录错误
	} else {
		// 基于路径模式匹配相关的 API 路由
		menuPath := menu.Path
		s.log.Info("开始自动分配 API 路由", "menu_path", menuPath)

		assignedCount := 0
		for _, route := range routes {
			// 匹配规则：API 路由路径包含菜单路径，且不是公开 API（auth_type != 0）
			// 排除 Swagger 和认证相关的公开 API
			if route.AuthType != 0 && // 排除 auth_type = 0 的基础 API
				!strings.HasPrefix(route.Path, "/swagger") &&
				!strings.HasPrefix(route.Path, "/api//auth") &&
				!strings.HasPrefix(route.Path, "/api//health") &&
				!strings.HasPrefix(route.Path, "/api//pay/notify") &&
				strings.Contains(route.Path, menuPath) {
				// 检查路由是否已经分配给角色
				roleRoutes, err := s.roleRepo.GetsByRole(roleID)
				if err != nil {
					s.log.Error("获取角色关联 API 路由失败", "error", err, "role_id", roleID)
					continue
				}

				routeExists := false
				for _, r := range roleRoutes {
					if r.ID == route.ID {
						routeExists = true
						break
					}
				}

				if !routeExists {
					// 分配 API 路由给角色
					if err := s.roleRepo.AddToRole(roleID, route.ID); err != nil {
						s.log.Error("自动分配 API 路由失败", "error", err, "role_id", roleID, "route_id", route.ID)
					} else {
						assignedCount++
					}
				}
			}
		}
		s.log.Info("自动分配 API 路由完成", "role_id", roleID, "menu_id", menuID, "assigned_count", assignedCount)
	}

	s.log.Info("分配菜单到角色成功", "role_id", roleID, "role_name", role.Name, "menu_id", menuID)
	return nil
}

// RevokeMenuFromRole 从角色中撤销菜单
func (s *RoleService) RevokeMenuFromRole(roleID, menuID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查菜单是否在角色中
	menus, err := s.roleRepo.GetMenusByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联菜单失败", "error", err, "role_id", roleID)
		return err
	}

	menuExists := false
	for _, menu := range menus {
		if menu.ID == menuID {
			menuExists = true
			break
		}
	}

	if !menuExists {
		s.log.Error("菜单不在角色中", "role_id", roleID, "menu_id", menuID)
		return errors.New("菜单不在角色中")
	}

	if err := s.roleRepo.RemoveMenuFromRole(roleID, menuID); err != nil {
		s.log.Error("从角色中撤销菜单失败", "error", err, "role_id", roleID, "menu_id", menuID)
		return err
	}

	s.log.Info("从角色中撤销菜单成功", "role_id", roleID, "role_name", role.Name, "menu_id", menuID)
	return nil
}

// GetMenusForRole 获取角色的所有菜单
func (s *RoleService) GetMenusForRole(roleID string) ([]*entity.Menu, error) {
	menus, err := s.roleRepo.GetMenusByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联菜单失败", "error", err, "role_id", roleID)
		return nil, err
	}

	// 填充所有菜单的元数据
	for _, menu := range menus {
		fillMenuMeta(menu)
	}

	return menus, nil
}

// BatchAssignMenusToRole 批量分配菜单到角色
func (s *RoleService) BatchAssignMenusToRole(roleID string, menuIDs []string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 验证所有菜单是否存在
	for _, menuID := range menuIDs {
		_, err := s.menuRepo.GetByID(menuID)
		if err != nil {
			s.log.Error("菜单不存在", "menu_id", menuID)
			return fmt.Errorf("菜单 %s 不存在", menuID)
		}
	}

	// 批量添加菜单到角色（不清除现有菜单）
	for _, menuID := range menuIDs {
		if err := s.roleRepo.AddMenuToRole(roleID, menuID); err != nil {
			s.log.Error("分配菜单到角色失败", "error", err, "role_id", roleID, "menu_id", menuID)
			return err
		}
	}

	s.log.Info("批量分配菜单到角色成功", "role_id", roleID, "role_name", role.Name, "menu_count", len(menuIDs))
	return nil
}

// BatchRevokeMenusFromRole 批量从角色中撤销菜单
func (s *RoleService) BatchRevokeMenusFromRole(roleID string, menuIDs []string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 批量从角色中撤销菜单
	for _, menuID := range menuIDs {
		if err := s.roleRepo.RemoveMenuFromRole(roleID, menuID); err != nil {
			s.log.Error("从角色中撤销菜单失败", "error", err, "role_id", roleID, "menu_id", menuID)
			return err
		}
	}

	s.log.Info("批量从角色中撤销菜单成功", "role_id", roleID, "role_name", role.Name, "menu_count", len(menuIDs))
	return nil
}

// AssignAPIToRole 分配API到角色
func (s *RoleService) AssignAPIToRole(roleID, apiID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查API是否存在
	api, err := s.apiRepo.GetByID(apiID)
	if err != nil {
		s.log.Error("API不存在", "api_id", apiID)
		return errors.New("API不存在")
	}

	// 检查API是否已分配给角色
	apis, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联API失败", "error", err, "role_id", roleID)
		return err
	}

	for _, a := range apis {
		if a.ID == apiID {
			s.log.Error("API已分配给角色", "role_id", roleID, "api_id", apiID)
			return errors.New("API已分配给角色")
		}
	}

	if err := s.roleRepo.AddToRole(roleID, apiID); err != nil {
		s.log.Error("分配API到角色失败", "error", err, "role_id", roleID, "api_id", apiID)
		return err
	}

	s.log.Info("分配API到角色成功", "role_id", roleID, "role_name", role.Name, "api_id", apiID, "api_path", api.Path)
	return nil
}

// RevokeAPIFromRole 从角色中撤销API
func (s *RoleService) RevokeAPIFromRole(roleID, apiID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查API是否存在
	api, err := s.apiRepo.GetByID(apiID)
	if err != nil {
		s.log.Error("API不存在", "api_id", apiID)
		return errors.New("API不存在")
	}

	// 检查API是否已分配给角色
	apis, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联API失败", "error", err, "role_id", roleID)
		return err
	}

	found := false
	for _, a := range apis {
		if a.ID == apiID {
			found = true
			break
		}
	}

	if !found {
		s.log.Error("API未分配给角色", "role_id", roleID, "api_id", apiID)
		return errors.New("API未分配给角色")
	}

	if err := s.roleRepo.RemoveFromRole(roleID, apiID); err != nil {
		s.log.Error("从角色中撤销API失败", "error", err, "role_id", roleID, "api_id", apiID)
		return err
	}

	s.log.Info("从角色中撤销API成功", "role_id", roleID, "role_name", role.Name, "api_id", apiID, "api_path", api.Path)
	return nil
}

// GetAPIsForRole 获取角色的所有 API
func (s *RoleService) GetAPIsForRole(roleID string) ([]*entity.API, error) {
	apis, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联 API 失败", "error", err, "role_id", roleID)
		return nil, err
	}
	return apis, nil
}

// GetRolesForAPI 获取 API 的所有角色
func (s *RoleService) GetRolesForAPI(apiID string) ([]*entity.Role, error) {
	roles, err := s.roleRepo.GetRolesBy(apiID)
	if err != nil {
		s.log.Error("获取API关联角色失败", "error", err, "api_id", apiID)
		return nil, err
	}
	return roles, nil
}

// BatchAssignAPIsToRole 批量分配 API 到角色
func (s *RoleService) BatchAssignAPIsToRole(roleID string, apiIDs []string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 保护系统角色，不允许修改权限
	if role.IsSystem {
		s.log.Error("系统角色不允许修改权限", "role_id", roleID, "role_name", role.Name)
		return errors.New("系统角色不允许修改权限")
	}

	// 验证所有 API 是否存在
	for _, apiID := range apiIDs {
		_, err := s.apiRepo.GetByID(apiID)
		if err != nil {
			s.log.Error("API 不存在", "api_id", apiID)
			return err
		}
	}

	// 批量添加 API 关联（不清除现有 API）
	for _, apiID := range apiIDs {
		if err := s.roleRepo.AddToRole(roleID, apiID); err != nil {
			s.log.Error("分配 API 到角色失败", "error", err, "role_id", roleID, "api_id", apiID)
			return err
		}
	}

	s.log.Info("批量分配 API 到角色成功", "role_id", roleID, "role_name", role.Name, "api_count", len(apiIDs))
	return nil
}

// BatchRevokeAPIsFromRole 批量从角色中撤销API
func (s *RoleService) BatchRevokeAPIsFromRole(roleID string, apiIDs []string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 批量移除API关联
	for _, apiID := range apiIDs {
		if err := s.roleRepo.RemoveFromRole(roleID, apiID); err != nil {
			s.log.Error("从角色中撤销API失败", "error", err, "role_id", roleID, "api_id", apiID)
			return err
		}
	}

	s.log.Info("批量从角色中撤销API成功", "role_id", roleID, "role_name", role.Name, "api_count", len(apiIDs))
	return nil
}

// GetRolesForMenu 获取菜单的所有角色
func (s *RoleService) GetRolesForMenu(menuID string) ([]*entity.Role, error) {
	roles, err := s.roleRepo.GetRolesByMenu(menuID)
	if err != nil {
		s.log.Error("获取菜单关联角色失败", "error", err, "menu_id", menuID)
		return nil, err
	}
	return roles, nil
}

// GetRolesForBtnPerm 获取按钮权限的所有角色
func (s *RoleService) GetRolesForBtnPerm(btnPermID string) ([]*entity.Role, error) {
	roles, err := s.roleRepo.GetRolesByBtnPerm(btnPermID)
	if err != nil {
		s.log.Error("获取按钮权限关联角色失败", "error", err, "btn_perm_id", btnPermID)
		return nil, err
	}
	return roles, nil
}

// AssignToRole 分配接口路由到角色
func (s *RoleService) AssignToRole(roleID, routeID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查接口路由是否已在角色中
	routes, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联接口路由失败", "error", err, "role_id", roleID)
		return err
	}

	for _, route := range routes {
		if route.ID == routeID {
			s.log.Error("接口路由已在角色中", "role_id", roleID, "route_id", routeID)
			return errors.New("接口路由已在角色中")
		}
	}

	if err := s.roleRepo.AddToRole(roleID, routeID); err != nil {
		s.log.Error("分配接口路由到角色失败", "error", err, "role_id", roleID, "route_id", routeID)
		return err
	}

	s.log.Info("分配接口路由到角色成功", "role_id", roleID, "role_name", role.Name, "route_id", routeID)
	return nil
}

// RevokeFromRole 从角色中撤销接口路由
func (s *RoleService) RevokeFromRole(roleID, routeID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查接口路由是否在角色中
	routes, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联接口路由失败", "error", err, "role_id", roleID)
		return err
	}

	routeExists := false
	for _, route := range routes {
		if route.ID == routeID {
			routeExists = true
			break
		}
	}

	if !routeExists {
		s.log.Error("接口路由不在角色中", "role_id", roleID, "route_id", routeID)
		return errors.New("接口路由不在角色中")
	}

	if err := s.roleRepo.RemoveFromRole(roleID, routeID); err != nil {
		s.log.Error("从角色中撤销接口路由失败", "error", err, "role_id", roleID, "route_id", routeID)
		return err
	}

	s.log.Info("从角色中撤销接口路由成功", "role_id", roleID, "role_name", role.Name, "route_id", routeID)
	return nil
}

// BatchAssignBtnPermsToRole 批量分配按钮权限给角色
func (s *RoleService) BatchAssignBtnPermsToRole(roleID string, btnPermIDs []string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 获取角色已有的按钮权限
	existingBtnPerms, err := s.roleRepo.GetBtnPermsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联按钮权限失败", "error", err, "role_id", roleID)
		return err
	}

	// 创建已存在按钮权限ID的映射，用于快速检查
	existingBtnPermMap := make(map[string]bool)
	for _, btnPerm := range existingBtnPerms {
		existingBtnPermMap[btnPerm.ID] = true
	}

	// 批量添加按钮权限关联（不清除现有权限）
	for _, btnPermID := range btnPermIDs {
		// 跳过已存在的按钮权限
		if existingBtnPermMap[btnPermID] {
			continue
		}

		// 检查按钮权限是否存在
		_, err := s.btnPermRepo.GetByID(btnPermID)
		if err != nil {
			s.log.Error("按钮权限不存在", "btn_perm_id", btnPermID)
			return fmt.Errorf("按钮权限 %s 不存在", btnPermID)
		}

		if err := s.roleRepo.AddBtnPermToRole(roleID, btnPermID); err != nil {
			s.log.Error("分配按钮权限给角色失败", "error", err, "role_id", roleID, "btn_perm_id", btnPermID)
			return err
		}
	}

	s.log.Info("批量分配按钮权限给角色成功", "role_id", roleID, "role_name", role.Name, "btn_perm_count", len(btnPermIDs))
	return nil
}

// BatchRevokeBtnPermsFromRole 批量从角色中撤销按钮权限
func (s *RoleService) BatchRevokeBtnPermsFromRole(roleID string, btnPermIDs []string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 批量移除按钮权限关联
	for _, btnPermID := range btnPermIDs {
		if err := s.roleRepo.RemoveBtnPermFromRole(roleID, btnPermID); err != nil {
			s.log.Error("从角色撤销按钮权限失败", "error", err, "role_id", roleID, "btn_perm_id", btnPermID)
			return err
		}
	}

	s.log.Info("批量从角色撤销按钮权限成功", "role_id", roleID, "role_name", role.Name, "btn_perm_count", len(btnPermIDs))
	return nil
}

// BatchAssignUsersToRole 批量分配用户到角色
func (s *RoleService) BatchAssignUsersToRole(roleID string, userIDs []string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 验证所有用户是否存在
	for _, userID := range userIDs {
		_, err := s.userRepo.GetByID(userID)
		if err != nil {
			s.log.Error("用户不存在", "user_id", userID)
			return fmt.Errorf("用户 %s 不存在", userID)
		}
	}

	// 批量添加用户角色关联
	for _, userID := range userIDs {
		// 检查用户是否已在角色中
		users, err := s.roleRepo.GetUsersByRole(roleID)
		if err != nil {
			s.log.Error("获取角色关联用户失败", "error", err, "role_id", roleID)
			continue
		}

		userExists := false
		for _, user := range users {
			if user.ID == userID {
				userExists = true
				break
			}
		}

		if !userExists {
			if err := s.roleRepo.AddUserToRole(roleID, userID); err != nil {
				s.log.Error("分配用户到角色失败", "error", err, "role_id", roleID, "user_id", userID)
				continue
			}
		}
	}

	s.log.Info("批量分配用户到角色成功", "role_id", roleID, "role_name", role.Name, "user_count", len(userIDs))
	return nil
}

// GetsForRole 获取角色的所有接口路由
func (s *RoleService) GetsForRole(roleID string) ([]*entity.API, error) {
	routes, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联接口路由失败", "error", err, "role_id", roleID)
		return nil, err
	}
	return routes, nil
}

// AssignBtnPermToRole 分配按钮权限给角色
func (s *RoleService) AssignBtnPermToRole(roleID, buttonID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查按钮权限是否存在
	button, err := s.btnPermRepo.GetByID(buttonID)
	if err != nil {
		s.log.Error("获取按钮权限失败", "error", err, "button_id", buttonID)
		return err
	}

	// 检查按钮权限是否已在角色中
	buttons, err := s.roleRepo.GetBtnPermsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联按钮权限失败", "error", err, "role_id", roleID)
		return err
	}

	for _, b := range buttons {
		if b.ID == buttonID {
			s.log.Error("按钮权限已在角色中", "role_id", roleID, "button_id", buttonID)
			return errors.New("按钮权限已在角色中")
		}
	}

	if err := s.roleRepo.AddBtnPermToRole(roleID, buttonID); err != nil {
		s.log.Error("分配按钮权限给角色失败", "error", err, "role_id", roleID, "button_id", buttonID)
		return err
	}

	s.log.Info("分配按钮权限给角色成功", "role_id", roleID, "role_name", role.Name, "button_id", buttonID, "button_name", button.Name)
	return nil
}

// RevokeBtnPermFromRole 从角色中撤销按钮权限
func (s *RoleService) RevokeBtnPermFromRole(roleID, buttonID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查按钮权限是否在角色中
	buttons, err := s.roleRepo.GetBtnPermsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联按钮权限失败", "error", err, "role_id", roleID)
		return err
	}

	buttonExists := false
	for _, b := range buttons {
		if b.ID == buttonID {
			buttonExists = true
			break
		}
	}

	if !buttonExists {
		s.log.Error("按钮权限不在角色中", "role_id", roleID, "button_id", buttonID)
		return errors.New("按钮权限不在角色中")
	}

	if err := s.roleRepo.RemoveBtnPermFromRole(roleID, buttonID); err != nil {
		s.log.Error("从角色中撤销按钮权限失败", "error", err, "role_id", roleID, "button_id", buttonID)
		return err
	}

	s.log.Info("从角色中撤销按钮权限成功", "role_id", roleID, "role_name", role.Name, "button_id", buttonID)
	return nil
}

// GetBtnPermsForRole 获取角色的所有按钮权限
func (s *RoleService) GetBtnPermsForRole(roleID string) ([]*entity.BtnPerm, error) {
	buttons, err := s.roleRepo.GetBtnPermsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色关联按钮权限失败", "error", err, "role_id", roleID)
		return nil, err
	}
	return buttons, nil
}

// GetRolesFor 获取接口路由的所有角色
func (s *RoleService) GetRolesFor(routeID string) ([]*entity.Role, error) {
	roles, err := s.roleRepo.GetRolesBy(routeID)
	if err != nil {
		s.log.Error("获取接口路由关联角色失败", "error", err, "route_id", routeID)
		return nil, err
	}
	return roles, nil
}

// BatchRevokeUsersFromRole 批量从角色中撤销用户
func (s *RoleService) BatchRevokeUsersFromRole(roleID string, userIDs []string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 批量移除用户角色关联
	for _, userID := range userIDs {
		// 检查用户是否在角色中
		users, err := s.roleRepo.GetUsersByRole(roleID)
		if err != nil {
			s.log.Error("获取角色关联用户失败", "error", err, "role_id", roleID)
			continue
		}

		userExists := false
		for _, user := range users {
			if user.ID == userID {
				userExists = true
				break
			}
		}

		if userExists {
			if err := s.roleRepo.RemoveUserFromRole(roleID, userID); err != nil {
				s.log.Error("从角色中撤销用户失败", "error", err, "role_id", roleID, "user_id", userID)
				continue
			}
		}
	}

	s.log.Info("批量从角色中撤销用户成功", "role_id", roleID, "role_name", role.Name, "user_count", len(userIDs))
	return nil
}

// AssignPermissionGroupToRole 分配权限组到角色
func (s *RoleService) AssignPermissionGroupToRole(roleID, groupID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查权限组是否存在
	_, err = s.permissionGroupRepo.GetByID(groupID)
	if err != nil {
		s.log.Error("权限组不存在", "group_id", groupID)
		return errors.New("权限组不存在")
	}

	// 检查权限组是否已分配给角色
	groups, err := s.permissionGroupRepo.GetGroupsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色权限组失败", "error", err, "role_id", roleID)
		return err
	}

	for _, g := range groups {
		if g.ID == groupID {
			s.log.Error("权限组已分配给角色", "role_id", roleID, "group_id", groupID)
			return errors.New("权限组已分配给角色")
		}
	}

	// 添加权限组 ID 到角色
	role.PermissionGroupIds = append(role.PermissionGroupIds, groupID)

	// 更新角色
	if err := s.roleRepo.Update(role); err != nil {
		s.log.Error("分配权限组失败", "error", err, "role_id", roleID, "group_id", groupID)
		return err
	}

	s.log.Info("分配权限组成功", "role_id", roleID, "group_id", groupID)
	return nil
}

// RevokePermissionGroupFromRole 从角色中撤销权限组
func (s *RoleService) RevokePermissionGroupFromRole(roleID, groupID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return errors.New("角色不存在")
	}

	// 检查权限组是否存在
	_, err = s.permissionGroupRepo.GetByID(groupID)
	if err != nil {
		s.log.Error("权限组不存在", "group_id", groupID)
		return errors.New("权限组不存在")
	}

	// 移除权限组 ID
	for i, id := range role.PermissionGroupIds {
		if id == groupID {
			role.PermissionGroupIds = append(role.PermissionGroupIds[:i], role.PermissionGroupIds[i+1:]...)
			break
		}
	}

	// 更新角色
	if err := s.roleRepo.Update(role); err != nil {
		s.log.Error("撤销权限组失败", "error", err, "role_id", roleID, "group_id", groupID)
		return err
	}

	s.log.Info("撤销权限组成功", "role_id", roleID, "group_id", groupID)
	return nil
}

// GetPermissionGroupsForRole 获取角色的所有权限组
func (s *RoleService) GetPermissionGroupsForRole(roleID string) ([]*entity.PermissionGroup, error) {
	// 检查角色是否存在
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("角色不存在", "role_id", roleID)
		return nil, errors.New("角色不存在")
	}

	groups, err := s.permissionGroupRepo.GetGroupsByRole(roleID)
	if err != nil {
		s.log.Error("获取角色权限组失败", "error", err, "role_id", roleID)
		return nil, err
	}

	return groups, nil
}

// GetRoleTree 获取角色树
func (s *RoleService) GetRoleTree() ([]*entity.Role, error) {
	// 获取所有角色
	roles, err := s.roleRepo.GetAllRoles()
	if err != nil {
		s.log.Error("获取所有角色失败", "error", err)
		return nil, err
	}

	// 构建角色树结构
	roleMap := make(map[string]*entity.Role)
	rootRoles := []*entity.Role{}

	// 首先将所有角色放入map中
	for i := range roles {
		roleMap[roles[i].ID] = roles[i]
	}

	// 然后构建树形结构
	for i := range roles {
		role := roles[i]
		if role.ParentID == "" {
			// 根角色
			rootRoles = append(rootRoles, role)
		} else {
			// 子角色，添加到父角色的Children中
			if parent, exists := roleMap[role.ParentID]; exists {
				parent.Children = append(parent.Children, role)
			}
		}
	}

	return rootRoles, nil
}

// GetRolePath 获取角色路径
func (s *RoleService) GetRolePath(id string) ([]*entity.Role, error) {
	// 获取当前角色
	currentRole, err := s.roleRepo.GetByID(id)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "id", id)
		return nil, errors.New("角色不存在")
	}

	// 构建角色路径（从当前角色到根角色）
	path := []*entity.Role{currentRole}

	// 向上遍历直到根角色
	currentID := currentRole.ParentID
	for currentID != "" {
		parentRole, err := s.roleRepo.GetByID(currentID)
		if err != nil {
			s.log.Error("获取父角色失败", "error", err, "id", currentID)
			return nil, err
		}

		// 将父角色添加到路径的开头
		path = append([]*entity.Role{parentRole}, path...)
		currentID = parentRole.ParentID
	}

	return path, nil
}

// GetRolesForPermissionGroup 获取权限组的所有角色
func (s *RoleService) GetRolesForPermissionGroup(groupID string) ([]*entity.Role, error) {
	// 检查权限组是否存在
	_, err := s.permissionGroupRepo.GetByID(groupID)
	if err != nil {
		s.log.Error("权限组不存在", "group_id", groupID)
		return nil, errors.New("权限组不存在")
	}

	roles, err := s.permissionGroupRepo.GetRolesByGroup(groupID)
	if err != nil {
		s.log.Error("获取权限组角色失败", "error", err, "group_id", groupID)
		return nil, err
	}

	return roles, nil
}

// GetAllInheritedBtnPerms 获取角色及其所有父角色的按钮权限
func (s *RoleService) GetAllInheritedBtnPerms(roleID string) ([]*entity.BtnPerm, error) {
	// 获取所有父角色（包括当前角色）
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "role_id", roleID)
		return nil, errors.New("角色不存在")
	}

	// 获取所有父角色ID（包括当前角色）
	var roleIDs []string
	currentRole := role
	for currentRole != nil {
		roleIDs = append(roleIDs, currentRole.ID)
		if currentRole.ParentID == "" {
			break
		}
		currentRole, err = s.roleRepo.GetByID(currentRole.ParentID)
		if err != nil {
			s.log.Error("获取父角色失败", "error", err, "parent_id", currentRole.ParentID)
			break
		}
	}

	// 收集所有角色的按钮权限
	allBtnPerms := make(map[string]*entity.BtnPerm)
	for _, id := range roleIDs {
		btnPerms, err := s.roleRepo.GetBtnPermsByRole(id)
		if err != nil {
			s.log.Error("获取角色按钮权限失败", "error", err, "role_id", id)
			continue
		}
		for _, btnPerm := range btnPerms {
			allBtnPerms[btnPerm.ID] = btnPerm
		}
	}

	// 转换为切片返回
	result := make([]*entity.BtnPerm, 0, len(allBtnPerms))
	for _, btnPerm := range allBtnPerms {
		result = append(result, btnPerm)
	}

	return result, nil
}

// GetAllInheritedPermissions 获取角色及其所有父角色的API权限
func (s *RoleService) GetAllInheritedPermissions(roleID string) ([]*entity.API, error) {
	// 获取所有父角色（包括当前角色）
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "role_id", roleID)
		return nil, errors.New("角色不存在")
	}

	// 获取所有父角色ID（包括当前角色）
	var roleIDs []string
	currentRole := role
	for currentRole != nil {
		roleIDs = append(roleIDs, currentRole.ID)
		if currentRole.ParentID == "" {
			break
		}
		currentRole, err = s.roleRepo.GetByID(currentRole.ParentID)
		if err != nil {
			s.log.Error("获取父角色失败", "error", err, "parent_id", currentRole.ParentID)
			break
		}
	}

	// 收集所有角色的API权限
	allAPIs := make(map[string]*entity.API)
	for _, id := range roleIDs {
		apis, err := s.roleRepo.GetsByRole(id)
		if err != nil {
			s.log.Error("获取角色API权限失败", "error", err, "role_id", id)
			continue
		}
		for _, api := range apis {
			allAPIs[api.ID] = api
		}
	}

	// 转换为切片返回
	result := make([]*entity.API, 0, len(allAPIs))
	for _, api := range allAPIs {
		result = append(result, api)
	}

	return result, nil
}

// GetAllInheritedMenus 获取角色及其所有父角色的菜单
func (s *RoleService) GetAllInheritedMenus(roleID string) ([]*entity.Menu, error) {
	// 获取所有父角色（包括当前角色）
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("获取角色失败", "error", err, "role_id", roleID)
		return nil, errors.New("角色不存在")
	}

	// 获取所有父角色ID（包括当前角色）
	var roleIDs []string
	currentRole := role
	for currentRole != nil {
		roleIDs = append(roleIDs, currentRole.ID)
		if currentRole.ParentID == "" {
			break
		}
		currentRole, err = s.roleRepo.GetByID(currentRole.ParentID)
		if err != nil {
			s.log.Error("获取父角色失败", "error", err, "parent_id", currentRole.ParentID)
			break
		}
	}

	// 收集所有角色的菜单
	allMenus := make(map[string]*entity.Menu)
	for _, id := range roleIDs {
		menus, err := s.roleRepo.GetMenusByRole(id)
		if err != nil {
			s.log.Error("获取角色菜单失败", "error", err, "role_id", id)
			continue
		}
		for _, menu := range menus {
			allMenus[menu.ID] = menu
		}
	}

	// 转换为切片返回
	result := make([]*entity.Menu, 0, len(allMenus))
	for _, menu := range allMenus {
		result = append(result, menu)
	}

	return result, nil
}
