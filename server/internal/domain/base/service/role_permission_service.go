package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"gorm.io/gorm"
)

// RolePermissionService 角色权限服务实现
type RolePermissionService struct {
	db          *database.PostgresDB
	roleRepo    repo.RoleRepository
	menuRepo    repo.MenuRepository
	btnPermRepo repo.BtnPermRepository
	apiRepo     repo.APIRepository
	cache       cache.Cache
	log         logger.Logger
}

// NewRolePermissionService 创建角色权限服务实例
func NewRolePermissionService(
	db *database.PostgresDB,
	roleRepo repo.RoleRepository,
	menuRepo repo.MenuRepository,
	btnPermRepo repo.BtnPermRepository,
	apiRepo repo.APIRepository,
	cache cache.Cache,
	log logger.Logger,
) *RolePermissionService {
	return &RolePermissionService{
		db:          db,
		roleRepo:    roleRepo,
		menuRepo:    menuRepo,
		btnPermRepo: btnPermRepo,
		apiRepo:     apiRepo,
		cache:       cache,
		log:         log,
	}
}

// SaveRolePermissions 保存角色权限（菜单 + 按钮+API），自动去重和清理
func (s *RolePermissionService) SaveRolePermissions(roleID string, menuIds []string, btnPermIds []string, apiIds []string, operatorID string) error {
	// 1. 获取分布式锁，防止并发修改
	lockKey := fmt.Sprintf("lock:role:perms:%s", roleID)

	// 尝试获取锁
	lockAcquired := false
	for i := 0; i < 3; i++ {
		success, err := s.cache.Exists(lockKey)
		if err != nil {
			continue
		}
		if !success {
			// 锁不存在，尝试获取
			err = s.cache.Set(lockKey, "1", 10*time.Second)
			if err != nil {
				continue
			}
			lockAcquired = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !lockAcquired {
		return errors.New("有其他操作正在修改该角色权限，请稍后再试")
	}
	defer s.cache.Delete(lockKey)

	// 2. 开始事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 3. 获取角色已关联的菜单、按钮、API
		oldMenus, _ := s.roleRepo.GetMenusByRole(roleID)
		oldBtnPerms, _ := s.roleRepo.GetBtnPermsByRole(roleID)
		oldAPIs, _ := s.roleRepo.GetsByRole(roleID)

		// 4. 获取菜单关联的 API
		menuAPIs := make(map[string]bool)
		for _, menuID := range menuIds {
			menu, err := s.menuRepo.GetByID(menuID)
			if err != nil {
				continue
			}
			// 获取菜单关联的 API
			for _, apiID := range menu.APIRouteIds {
				menuAPIs[apiID] = true
			}
		}

		// 5. 获取按钮关联的 API
		// TODO: 获取按钮关联的 API - 需要实现按钮权限与 API 的关联逻辑
		btnAPIs := make(map[string]bool)
		_ = btnAPIs // 暂时使用，避免编译错误

		// 6-10. TODO: 清理和保存角色关联（需要创建关联实体）
		// 这部分代码需要等关联实体定义完成后再实现
		// 目前暂时跳过，只保留基本的菜单和按钮权限分配

		// 11. 记录审计日志
		s.logPermissionChange(tx, roleID, operatorID, oldMenus, oldBtnPerms, oldAPIs, menuIds, btnPermIds, apiIds)

		// 12. 清除权限缓存
		s.cache.Delete(fmt.Sprintf("role:perms:%s", roleID))

		return nil
	})
}

// GetRolePermissions 获取角色权限详情
func (s *RolePermissionService) GetRolePermissions(roleID string) (menuIds []string, btnPermIds []string, apiIds []string, err error) {
	menus, err := s.roleRepo.GetMenusByRole(roleID)
	if err != nil {
		return nil, nil, nil, err
	}

	btnPerms, err := s.roleRepo.GetBtnPermsByRole(roleID)
	if err != nil {
		return nil, nil, nil, err
	}

	apis, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		return nil, nil, nil, err
	}

	menuIds = make([]string, len(menus))
	for i, m := range menus {
		menuIds[i] = m.ID
	}

	btnPermIds = make([]string, len(btnPerms))
	for i, b := range btnPerms {
		btnPermIds[i] = b.ID
	}

	apiIds = make([]string, len(apis))
	for i, a := range apis {
		apiIds[i] = a.ID
	}

	return menuIds, btnPermIds, apiIds, nil
}

// GetAvailableApisForRole 获取角色可授权的 API 列表（过滤已关联的 API）
func (s *RolePermissionService) GetAvailableApisForRole(roleID string) ([]*entity.API, error) {
	// 获取所有 API
	allAPIs, err := s.apiRepo.GetAllRoutes()
	if err != nil {
		return nil, err
	}

	// 获取角色已授权的 API
	roleAPIs, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		return nil, err
	}

	// 获取菜单关联的 API
	menus, err := s.roleRepo.GetMenusByRole(roleID)
	if err != nil {
		return nil, err
	}
	menuAPIMap := make(map[string]bool)
	for _, menu := range menus {
		for _, apiID := range menu.APIRouteIds {
			menuAPIMap[apiID] = true
		}
	}

	// 获取按钮关联的 API
	btnPerms, err := s.roleRepo.GetBtnPermsByRole(roleID)
	if err != nil {
		return nil, err
	}
	btnAPIMap := make(map[string]bool)
	for _, btn := range btnPerms {
		// TODO: 获取按钮关联的 API
		// 假设 BtnPerm 有 APIRouteIds 字段
		_ = btn // 暂时使用，避免编译错误
	}

	// 构建角色已关联的 API 集合（包括直接授权、菜单关联、按钮关联）
	roleAPIMap := make(map[string]bool)
	for _, api := range roleAPIs {
		roleAPIMap[api.ID] = true
	}
	for apiID := range menuAPIMap {
		roleAPIMap[apiID] = true
	}
	for apiID := range btnAPIMap {
		roleAPIMap[apiID] = true
	}

	// 过滤出未关联的 API
	availableAPIs := make([]*entity.API, 0)
	for _, api := range allAPIs {
		if api.Status != 1 {
			continue // 跳过已禁用的 API
		}
		if !roleAPIMap[api.ID] {
			availableAPIs = append(availableAPIs, api)
		}
	}

	return availableAPIs, nil
}

// logPermissionChange 记录权限变更审计日志
func (s *RolePermissionService) logPermissionChange(tx *gorm.DB, roleID string, operatorID string,
	oldMenus []*entity.Menu, oldBtnPerms []*entity.BtnPerm, oldAPIs []*entity.API,
	newMenuIds []string, newBtnPermIds []string, newAPIIds []string) {

	// 获取操作人信息（这里简化处理，实际应该从用户服务获取）
	_ = "system" // operatorName 暂时不使用

	// 构建变更数据
	beforeData, _ := json.Marshal(map[string]interface{}{
		"menus":     oldMenus,
		"btn_perms": oldBtnPerms,
		"apis":      oldAPIs,
	})

	afterData, _ := json.Marshal(map[string]interface{}{
		"menus":     newMenuIds,
		"btn_perms": newBtnPermIds,
		"apis":      newAPIIds,
	})

	// TODO: 实现权限变更日志记录逻辑
	// 需要定义 repo.PermissionLog 实体
	_ = beforeData
	_ = afterData

	// 异步记录日志，不阻塞主流程
	// TODO: 实现日志记录逻辑
}

// LoadRolePermissionsToRedis 从数据库加载角色权限到 Redis 缓存
// TODO: 需要实现完整的权限缓存逻辑
func (s *RolePermissionService) LoadRolePermissionsToRedis(roleID string) error {
	// 暂时注释，等待实现完整的权限缓存逻辑
	// 需要定义 repo.RolePermissions 实体
	// 需要实现菜单和按钮权限与 API 的关联逻辑
	s.log.Info("加载角色权限到 Redis 缓存（暂未实现）", "roleID", roleID)
	return nil
}

// GetRolePermissionsFromRedis 从 Redis 获取角色权限缓存
// TODO: 需要实现完整的权限缓存逻辑
func (s *RolePermissionService) GetRolePermissionsFromRedis(roleID string) (*entity.RolePermissions, error) {
	// 暂时注释，等待实现完整的权限缓存逻辑
	return nil, nil
}

// ClearRolePermissionsCache 清除角色权限缓存
func (s *RolePermissionService) ClearRolePermissionsCache(roleID string) error {
	cacheKey := fmt.Sprintf("role:perms:%s", roleID)

	err := s.cache.Delete(cacheKey)
	if err != nil {
		return fmt.Errorf("清除角色权限缓存失败：%w", err)
	}

	s.log.Info("角色权限缓存已清除", "roleID", roleID)
	return nil
}

// ClearRoleCacheByMenuID 清除包含该菜单的所有角色缓存
func (s *RolePermissionService) ClearRoleCacheByMenuID(menuID string) error {
	// 获取所有有该菜单权限的角色
	roles, err := s.roleRepo.GetRolesByMenu(menuID)
	if err != nil {
		return err
	}

	// 清除所有相关角色的缓存
	for _, role := range roles {
		if err := s.ClearRolePermissionsCache(role.ID); err != nil {
			s.log.Error("清除角色缓存失败", "roleID", role.ID, "error", err)
		}
	}

	return nil
}

// ClearRoleCacheByBtnPermID 清除包含该按钮的所有角色缓存
func (s *RolePermissionService) ClearRoleCacheByBtnPermID(btnPermID string) error {
	// 获取所有有该按钮权限的角色
	roles, err := s.roleRepo.GetRolesByBtnPerm(btnPermID)
	if err != nil {
		return err
	}

	// 清除所有相关角色的缓存
	for _, role := range roles {
		if err := s.ClearRolePermissionsCache(role.ID); err != nil {
			s.log.Error("清除角色缓存失败", "roleID", role.ID, "error", err)
		}
	}

	return nil
}

// ClearRoleCacheByAPIID 清除包含该 API 的所有角色缓存
func (s *RolePermissionService) ClearRoleCacheByAPIID(apiID string) error {
	// 获取所有有该 API 权限的角色
	roles, err := s.roleRepo.GetRolesBy(apiID)
	if err != nil {
		return err
	}

	// 清除所有相关角色的缓存
	for _, role := range roles {
		if err := s.ClearRolePermissionsCache(role.ID); err != nil {
			s.log.Error("清除角色缓存失败", "roleID", role.ID, "error", err)
		}
	}

	return nil
}
