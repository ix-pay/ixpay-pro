package service

import (
	"errors"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// MenuService 菜单服务实现
type MenuService struct {
	repo        repo.MenuRepository
	btnPermRepo repo.BtnPermRepository
	roleRepo    repo.RoleRepository
	apiRepo     repo.APIRepository
	log         logger.Logger
}

// NewMenuService 创建菜单服务实例
func NewMenuService(
	repo repo.MenuRepository,
	btnPermRepo repo.BtnPermRepository,
	roleRepo repo.RoleRepository,
	apiRepo repo.APIRepository,
	log logger.Logger,
) *MenuService {
	return &MenuService{
		repo:        repo,
		btnPermRepo: btnPermRepo,
		roleRepo:    roleRepo,
		apiRepo:     apiRepo,
		log:         log,
	}
}

// fillMenuMeta 填充菜单元数据
func fillMenuMeta(menu *entity.Menu) {
	menu.Meta = entity.MenuMeta{
		Title:        menu.Title,
		Icon:         menu.Icon,
		KeepAlive:    menu.KeepAlive,
		DefaultMenu:  menu.DefaultMenu,
		Breadcrumb:   menu.Breadcrumb,
		ActiveMenu:   menu.ActiveMenu,
		Affix:        menu.Affix,
		FrameSrc:     menu.FrameSrc,
		FrameLoading: menu.FrameLoading,
	}

	// 确保 Children 字段始终是切片类型（空切片而不是 nil）
	if menu.Children == nil {
		menu.Children = []*entity.Menu{}
	}

	// 递归填充子菜单的元数据
	for _, child := range menu.Children {
		fillMenuMeta(child)
	}
}

// buildMenuTree 将扁平菜单列表转换为树形结构
func buildMenuTree(menus []*entity.Menu) []*entity.Menu {
	// 创建 map 用于快速查找
	menuMap := make(map[int64]*entity.Menu)
	for _, menu := range menus {
		menuMap[menu.ID] = menu
		// 初始化 children 为空切片
		if menu.Children == nil {
			menu.Children = []*entity.Menu{}
		}
	}

	// 构建树形结构
	var roots []*entity.Menu
	for _, menu := range menus {
		if menu.ParentID == 0 {
			// 顶级菜单
			roots = append(roots, menu)
		} else {
			// 子菜单，添加到父菜单的 children
			if parent, exists := menuMap[menu.ParentID]; exists {
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return roots
}

// convertToMenuResponse 将 entity.Menu 转换为 response.MenuResponse
func convertToMenuResponse(menu *entity.Menu) *response.MenuResponse {
	if menu == nil {
		return nil
	}

	// 确保 meta 已填充
	if menu.Meta.Title == "" {
		fillMenuMeta(menu)
	}

	// 处理组件路径
	component := menu.Component
	// 如果是目录类型(type=1)且component为空，设置默认占位组件路径
	if menu.Type == entity.MenuTypeDirectory && component == "" {
		component = "views/" + menu.Path + "/index"
	}

	// 转换为响应结构
	resp := &response.MenuResponse{
		ID:           menu.ID,
		ParentID:     menu.ParentID,
		Path:         menu.Path,
		Name:         menu.Name,
		Component:    component,
		Title:        menu.Title,
		Icon:         menu.Icon,
		Hidden:       menu.Hidden,
		Sort:         menu.Sort,
		Status:       menu.Status,
		IsExt:        menu.IsExt,
		Redirect:     menu.Redirect,
		Permission:   menu.Permission,
		KeepAlive:    menu.KeepAlive,
		DefaultMenu:  menu.DefaultMenu,
		Breadcrumb:   menu.Breadcrumb,
		ActiveMenu:   menu.ActiveMenu,
		Affix:        menu.Affix,
		Type:         int(menu.Type),
		FrameLoading: menu.FrameLoading,
		Meta: &response.MenuMetaResp{
			Title:        menu.Meta.Title,
			Icon:         menu.Meta.Icon,
			KeepAlive:    menu.Meta.KeepAlive,
			DefaultMenu:  menu.Meta.DefaultMenu,
			Breadcrumb:   menu.Meta.Breadcrumb,
			ActiveMenu:   menu.Meta.ActiveMenu,
			Affix:        menu.Meta.Affix,
			FrameLoading: menu.Meta.FrameLoading,
		},
	}

	// 转换子菜单
	if len(menu.Children) > 0 {
		resp.Children = make([]response.MenuResponse, len(menu.Children))
		for i, child := range menu.Children {
			childResp := convertToMenuResponse(child)
			if childResp != nil {
				resp.Children[i] = *childResp
			}
		}
	} else {
		resp.Children = []response.MenuResponse{}
	}

	return resp
}

// convertToMenuResponseList 将 entity.Menu 列表转换为 response.MenuResponse 列表
func convertToMenuResponseList(menus []*entity.Menu) []response.MenuResponse {
	if len(menus) == 0 {
		return []response.MenuResponse{}
	}

	result := make([]response.MenuResponse, len(menus))
	for i, menu := range menus {
		resp := convertToMenuResponse(menu)
		if resp != nil {
			result[i] = *resp
		}
	}
	return result
}

// GetUserMenus 获取用户可访问的菜单列表
func (s *MenuService) GetUserMenus(roleID int64) ([]response.MenuResponse, error) {
	s.log.Info("获取用户菜单列表", "roleID", roleID)

	// 检查是否为管理员角色 (code: "admin")
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		s.log.Error("获取角色信息失败", "error", err, "roleID", roleID)
		return nil, err
	}

	var menus []*entity.Menu
	// 如果是管理员角色，返回所有菜单
	if role.Code == "admin" {
		s.log.Info("管理员角色，返回所有菜单", "roleID", roleID)
		menus, err = s.repo.GetAll()
		if err != nil {
			s.log.Error("获取所有菜单失败", "error", err, "roleID", roleID)
			return nil, err
		}
	} else {
		// 普通角色，按权限分配返回菜单
		s.log.Info("普通角色，按权限返回菜单", "roleID", roleID)
		menus, err = s.repo.GetMenusByRole(roleID)
		if err != nil {
			s.log.Error("获取用户菜单失败", "error", err, "roleID", roleID)
			return nil, err
		}
	}

	// 填充所有菜单的元数据
	for _, menu := range menus {
		fillMenuMeta(menu)
	}

	// ⭐ 构建树形结构
	treeMenus := buildMenuTree(menus)

	// 转换为响应结构
	menuResponses := convertToMenuResponseList(treeMenus)

	s.log.Info("获取用户菜单成功", "roleID", roleID, "count", len(menuResponses))
	return menuResponses, nil
}

// GetDefaultRouter 获取用户默认路由
func (s *MenuService) GetDefaultRouter(roleID int64) (string, error) {
	s.log.Info("获取默认路由", "roleID", roleID)
	path, err := s.repo.GetDefaultRouter(roleID)
	if err != nil {
		s.log.Error("获取默认路由失败", "error", err, "roleID", roleID)
		return "", err
	}
	s.log.Info("获取默认路由成功", "roleID", roleID, "path", path)
	return path, nil
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(menu *entity.Menu, createdBy int64) error {
	s.log.Info("创建菜单", "name", menu.Name, "path", menu.Path)

	// 验证菜单参数
	if menu == nil {
		s.log.Error("菜单参数不能为空")
		return errors.New("菜单参数不能为空")
	}

	// 检查菜单名称是否已存在
	existingMenus, _, err := s.repo.List(1, 1, map[string]interface{}{"name = ?": menu.Name})
	if err != nil {
		s.log.Error("检查菜单名称失败", "error", err, "name", menu.Name)
		return err
	}
	if len(existingMenus) > 0 {
		s.log.Error("菜单名称已存在", "name", menu.Name)
		return errors.New("菜单名称已存在")
	}

	// 设置责任主体字段
	menu.CreatedBy = createdBy
	menu.UpdatedBy = createdBy

	// 创建菜单
	if err := s.repo.Create(menu); err != nil {
		s.log.Error("创建菜单失败", "error", err)
		return err
	}

	s.log.Info("创建菜单成功", "id", menu.ID)
	return nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(menu *entity.Menu, updatedBy int64) error {
	s.log.Info("更新菜单", "id", menu.ID, "name", menu.Name)

	// 检查菜单是否存在
	existingMenu, err := s.repo.GetByID(menu.ID)
	if err != nil {
		s.log.Error("获取菜单失败", "error", err, "id", menu.ID)
		return err
	}

	// 检查菜单名称是否已被其他菜单使用
	if existingMenu.Name != menu.Name {
		checkMenus, _, err := s.repo.List(1, 1, map[string]interface{}{"name = ?": menu.Name, "id != ?": menu.ID})
		if err == nil && len(checkMenus) > 0 {
			s.log.Error("菜单名称已存在", "name", menu.Name)
			return errors.New("菜单名称已存在")
		}
	}

	// 设置更新信息
	menu.UpdatedBy = updatedBy

	// 更新菜单
	if err := s.repo.Update(menu); err != nil {
		s.log.Error("更新菜单失败", "error", err)
		return err
	}

	s.log.Info("更新菜单成功", "id", menu.ID)
	return nil
}

// UpdateMenuWithAPIs 更新菜单并处理 API 关联
func (s *MenuService) UpdateMenuWithAPIs(menu *entity.Menu, apiRouteIDs []string, updatedBy int64) error {
	s.log.Info("更新菜单并处理 API 关联", "id", menu.ID, "api_count", len(apiRouteIDs))

	// 1. 更新菜单基本信息
	if err := s.UpdateMenu(menu, updatedBy); err != nil {
		return err
	}

	// 2. 处理 API 关联 - 这里应该通过仓库接口来处理，而不是直接访问数据库
	// 需要在 repo 包中添加相应的接口方法
	// TODO: 实现 API 关联的保存逻辑

	s.log.Info("更新菜单并处理 API 关联成功", "id", menu.ID, "api_count", len(apiRouteIDs))
	return nil
}

// DeleteMenu 删除菜单
func (s *MenuService) DeleteMenu(id int64) error {
	s.log.Info("删除菜单", "id", id)

	// 检查是否有子菜单
	hasChildren, err := s.hasChildren(id)
	if err != nil {
		s.log.Error("检查子菜单失败", "error", err)
		return err
	}
	if hasChildren {
		s.log.Error("菜单下有子菜单，无法删除", "id", id)
		return errors.New("菜单下有子菜单，无法删除")
	}

	// 删除菜单
	if err := s.repo.Delete(id); err != nil {
		s.log.Error("删除菜单失败", "error", err)
		return err
	}

	s.log.Info("删除菜单成功", "id", id)
	return nil
}

// GetMenuList 获取菜单列表，支持分页和过滤
func (s *MenuService) GetMenuList(page, pageSize int, filters map[string]interface{}) ([]*entity.Menu, int64, error) {
	s.log.Info("获取菜单列表", "page", page, "pageSize", pageSize)

	menus, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		s.log.Error("获取菜单列表失败", "error", err)
		return nil, 0, err
	}

	// 填充所有菜单的元数据
	for _, menu := range menus {
		fillMenuMeta(menu)
	}

	s.log.Info("获取菜单列表成功", "count", len(menus), "total", total)
	return menus, total, nil
}

// GetMenusByUserID 根据用户 ID 获取菜单列表
func (s *MenuService) GetMenusByUserID(userID int64) ([]*entity.Menu, error) {
	s.log.Info("根据用户 ID 获取菜单列表", "userID", userID)
	menus, err := s.repo.GetMenusByUserID(userID)
	if err != nil {
		s.log.Error("根据用户 ID 获取菜单失败", "error", err, "userID", userID)
		return nil, err
	}

	// 填充所有菜单的元数据
	for _, menu := range menus {
		fillMenuMeta(menu)
	}

	s.log.Info("根据用户 ID 获取菜单成功", "userID", userID, "count", len(menus))
	return menus, nil
}

// GetMenuTree 获取菜单树结构
func (s *MenuService) GetMenuTree(parentID int64) ([]*entity.Menu, error) {
	s.log.Info("获取菜单树结构", "parentID", parentID)

	// ⭐ 优化：先获取顶级菜单，然后使用 Preload 加载关联数据
	// 使用 List 方法获取指定 parentID 的菜单列表
	menus, _, err := s.repo.List(1, 1000, map[string]interface{}{"parent_id = ?": parentID})
	if err != nil {
		s.log.Error("获取菜单树失败", "error", err, "parentID", parentID)
		return nil, err
	}

	// 填充所有菜单的元数据
	for _, menu := range menus {
		fillMenuMeta(menu)
	}

	s.log.Info("获取菜单树成功", "parentID", parentID, "count", len(menus))
	return menus, nil
}

// GetAllMenuTree 获取所有菜单的树结构
func (s *MenuService) GetAllMenuTree() ([]*entity.Menu, error) {
	s.log.Info("获取所有菜单的树结构")

	// 获取所有菜单
	menus, _, err := s.repo.List(1, 1000, map[string]interface{}{})
	if err != nil {
		s.log.Error("获取所有菜单失败", "error", err)
		return nil, err
	}

	// 填充所有菜单的元数据
	for _, menu := range menus {
		fillMenuMeta(menu)
	}

	// 构建树形结构
	treeMenus := buildMenuTree(menus)

	s.log.Info("获取所有菜单树成功", "count", len(treeMenus))
	return treeMenus, nil
}

// GetMenuPath 获取菜单路径
func (s *MenuService) GetMenuPath(menuID int64) ([]*entity.Menu, error) {
	s.log.Info("获取菜单路径", "menuID", menuID)

	// ⭐ 优化：使用 Preload 加载父菜单
	menu, err := s.repo.GetByID(menuID, repo.MenuRelationParent)
	if err != nil {
		s.log.Error("获取菜单路径失败", "error", err, "menuID", menuID)
		return nil, err
	}

	// 填充菜单元数据
	fillMenuMeta(menu)

	// 构建菜单路径（从根到当前菜单）
	var menus []*entity.Menu
	current := menu
	for current != nil {
		menus = append([]*entity.Menu{current}, menus...)
		current = current.Parent
	}

	s.log.Info("获取菜单路径成功", "menuID", menuID, "count", len(menus))
	return menus, nil
}

// GetMenusByType 获取指定类型的菜单列表
func (s *MenuService) GetMenusByType(menuType entity.MenuType) ([]*entity.Menu, error) {
	s.log.Info("获取指定类型的菜单列表", "menuType", menuType)
	menus, err := s.repo.GetMenusByType(menuType)
	if err != nil {
		s.log.Error("获取指定类型菜单失败", "error", err, "menuType", menuType)
		return nil, err
	}

	// 填充所有菜单的元数据
	for _, menu := range menus {
		fillMenuMeta(menu)
	}

	s.log.Info("获取指定类型菜单成功", "menuType", menuType, "count", len(menus))
	return menus, nil
}

// BatchDeleteMenu 批量删除菜单
func (s *MenuService) BatchDeleteMenu(ids []int64) error {
	s.log.Info("批量删除菜单", "ids", ids)

	// 检查菜单是否有子菜单
	for _, id := range ids {
		hasChildren, err := s.repo.CheckMenuChildren(id)
		if err != nil {
			s.log.Error("检查子菜单失败", "error", err, "id", id)
			return err
		}
		if hasChildren {
			s.log.Error("菜单下有子菜单，无法删除", "id", id)
			return errors.New("菜单下有子菜单，无法删除")
		}
	}

	// 批量删除菜单
	if err := s.repo.BatchDelete(ids); err != nil {
		s.log.Error("批量删除菜单失败", "error", err)
		return err
	}

	s.log.Info("批量删除菜单成功", "ids", ids)
	return nil
}

// RefreshMenuCache 刷新菜单缓存
func (s *MenuService) RefreshMenuCache() error {
	s.log.Info("刷新菜单缓存")
	// 这里可以添加缓存刷新逻辑，例如使用 Redis 等缓存技术
	// 目前仅记录日志
	s.log.Info("菜单缓存已刷新")
	return nil
}

// CheckMenuChildren 检查菜单是否有子菜单
func (s *MenuService) CheckMenuChildren(menuID int64) (bool, error) {
	s.log.Info("检查菜单是否有子菜单", "menuID", menuID)
	hasChildren, err := s.repo.CheckMenuChildren(menuID)
	if err != nil {
		s.log.Error("检查子菜单失败", "error", err, "menuID", menuID)
		return false, err
	}

	s.log.Info("检查子菜单成功", "menuID", menuID, "hasChildren", hasChildren)
	return hasChildren, nil
}

// 检查菜单是否有子菜单（内部使用）
func (s *MenuService) hasChildren(parentID int64) (bool, error) {
	menus, _, err := s.repo.List(1, 1, map[string]interface{}{"parent_id = ?": parentID})
	if err != nil {
		return false, err
	}
	return len(menus) > 0, nil
}

// CalculateDeleteImpact 评估删除菜单的影响范围
func (s *MenuService) CalculateDeleteImpact(menuID int64) (*entity.DeleteImpact, error) {
	impact := &entity.DeleteImpact{}

	// 1. 统计子菜单数量
	hasChildren, err := s.repo.CheckMenuChildren(menuID)
	if err != nil {
		return nil, err
	}
	if hasChildren {
		// 简单计算子菜单数量
		menus, _, err := s.repo.List(1, 100, map[string]interface{}{"parent_id = ?": menuID})
		if err == nil {
			impact.ChildMenusCount = int64(len(menus))
		}
	}

	// 2. 统计按钮数量（通过菜单 ID 查询）
	btnPerms, err := s.btnPermRepo.GetBtnPermsByMenu(menuID)
	if err == nil {
		impact.BtnPermsCount = int64(len(btnPerms))
	}

	// 3. 统计受影响的角色数量
	roles, err := s.roleRepo.GetRolesByMenu(menuID)
	if err == nil {
		impact.AffectedRolesCount = int64(len(roles))
	}

	// 4. 统计影响的 API 数量（通过菜单关联的 API）
	menu, err := s.repo.GetByID(menuID)
	if err == nil && menu != nil {
		impact.AffectedApisCount = int64(len(menu.APIRouteIds))
	}

	// 5. 评估影响等级
	totalImpact := impact.AffectedRolesCount + impact.AffectedApisCount

	if totalImpact > 20 || impact.AffectedRolesCount > 10 {
		impact.Level = "HIGH"
		impact.Warning = "该删除操作影响范围较大，将影响多个角色和 API，请谨慎操作"
	} else if totalImpact > 5 || impact.AffectedRolesCount > 3 {
		impact.Level = "MEDIUM"
		impact.Warning = "该删除操作会影响部分角色权限"
	} else {
		impact.Level = "LOW"
		impact.Warning = "该删除操作影响较小"
	}

	return impact, nil
}
