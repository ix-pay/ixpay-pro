package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"gorm.io/gorm"
)

// RolePermissionService 角色权限服务实现
// RBAC + ABAC 混合权限模型
// 菜单表有三种类型：一级目录（type=1）、二级菜单（type=2）、三级按钮（type=3）
// 二级菜单和三级按钮都可以关联一个或多个 API（通过 base_menu_api_routes 表）
// 给角色授权菜单或按钮时，自动拥有关联的 API 权限
// 授权角色 API 时，无需重复写入菜单/按钮已关联的 API
type RolePermissionService struct {
	db       *database.PostgresDB
	roleRepo repo.RoleRepository
	menuRepo repo.MenuRepository
	apiRepo  repo.APIRepository
	cache    cache.Cache
	log      logger.Logger
}

// NewRolePermissionService 创建角色权限服务实例
func NewRolePermissionService(
	db *database.PostgresDB,
	roleRepo repo.RoleRepository,
	menuRepo repo.MenuRepository,
	apiRepo repo.APIRepository,
	cache cache.Cache,
	log logger.Logger,
) *RolePermissionService {
	return &RolePermissionService{
		db:       db,
		roleRepo: roleRepo,
		menuRepo: menuRepo,
		apiRepo:  apiRepo,
		cache:    cache,
		log:      log,
	}
}

// SaveRolePermissions 保存角色权限（菜单、按钮、API）
// 菜单 IDs 包含 type=2（菜单）和 type=3（按钮）的菜单项
// 按钮也作为菜单项存储在 base_menus 表中，type=3 表示按钮
func (s *RolePermissionService) SaveRolePermissions(roleID int64, menuIds, apiIds []int64, operatorID string) error {
	// 1. 尝试获取分布式锁
	lockKey := fmt.Sprintf("lock:role:%d", roleID)
	lockAcquired := false
	for i := 0; i < 3; i++ {
		err := s.cache.Set(lockKey, "1", 5*time.Second)
		if err == nil {
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
		// 3. 获取角色已关联的菜单、API（用于审计日志）
		oldMenus, _ := s.roleRepo.GetMenusByRole(roleID)
		oldAPIs, _ := s.roleRepo.GetsByRole(roleID)

		// 4. 获取菜单关联的 API（base_menu_api_routes 表）
		// 菜单（type=2）和按钮（type=3）都通过此表关联 API
		// 需要 Preload APIRoutes 才能获取到菜单关联的 API 列表
		menuAPIs := make(map[int64]bool)
		for _, menuID := range menuIds {
			menu, err := s.menuRepo.GetByID(menuID, repo.MenuRelationAPIRoutes)
			if err != nil {
				continue
			}
			for _, apiID := range menu.APIRouteIds {
				menuAPIs[apiID] = true
			}
		}

		// 5. 清理角色现有关联
		if err := tx.Table("base_role_menus").Where("role_id = ?", roleID).Delete(nil).Error; err != nil {
			return err
		}
		if err := tx.Table("base_role_api_routes").Where("role_id = ?", roleID).Delete(nil).Error; err != nil {
			return err
		}

		// 6. 插入新的菜单关联（包括菜单和按钮）
		for _, menuID := range menuIds {
			if err := tx.Table("base_role_menus").Create(map[string]interface{}{
				"role_id": roleID,
				"menu_id": menuID,
			}).Error; err != nil {
				return err
			}
		}

		// 7. 过滤掉已在菜单关联中的 API，避免重复授权
		// 用户在前端手动选中的 apiIds 中，如果某个 API 已通过菜单关联授权，则不再写入 base_role_api_routes
		directAPIIds := make([]int64, 0, len(apiIds))
		for _, apiID := range apiIds {
			if !menuAPIs[apiID] {
				directAPIIds = append(directAPIIds, apiID)
			}
		}

		// 8. 校验 API ID 是否存在，跳过不存在的 ID 避免外键约束错误
		validDirectAPIIds, err := s.filterValidAPIIDs(tx, directAPIIds)
		if err != nil {
			return err
		}
		validMenuAPIIds, err := s.filterValidAPIIDs(tx, func() []int64 {
			ids := make([]int64, 0, len(menuAPIs))
			for apiID := range menuAPIs {
				ids = append(ids, apiID)
			}
			return ids
		}())
		if err != nil {
			return err
		}

		// 9. 插入新的 API 直接授权（仅写入非菜单关联的有效 API）
		for _, apiID := range validDirectAPIIds {
			if err := tx.Table("base_role_api_routes").Create(map[string]interface{}{
				"role_id":  roleID,
				"route_id": apiID,
				"source":   1, // 1-直接授权
			}).Error; err != nil {
				return err
			}
		}

		// 10. 处理菜单/按钮关联的 API 自动授权
		// 将菜单和按钮关联的 API 自动授权给角色，使用 source=2
		for _, apiID := range validMenuAPIIds {
			if err := tx.Table("base_role_api_routes").Create(map[string]interface{}{
				"role_id":  roleID,
				"route_id": apiID,
				"source":   2, // 2-菜单/按钮自动授权
			}).Error; err != nil {
				return err
			}
		}

		// 11. 记录审计日志
		s.logPermissionChange(tx, roleID, operatorID, oldMenus, oldAPIs, menuIds, apiIds)

		// 12. 清除权限缓存
		s.cache.Delete(fmt.Sprintf("role:perms:%d", roleID))

		return nil
	})
}

// GetRolePermissions 获取角色权限详情
func (s *RolePermissionService) GetRolePermissions(roleID int64) (menuIds []int64, apiIds []int64, err error) {
	menus, err := s.roleRepo.GetMenusByRole(roleID)
	if err != nil {
		return nil, nil, err
	}

	apis, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		return nil, nil, err
	}

	menuIds = make([]int64, len(menus))
	for i, m := range menus {
		menuIds[i] = m.ID
	}

	apiIds = make([]int64, len(apis))
	for i, a := range apis {
		apiIds[i] = a.ID
	}

	return menuIds, apiIds, nil
}

// GetAvailableApisForRole 获取角色可授权的 API 列表（过滤已关联的 API）
func (s *RolePermissionService) GetAvailableApisForRole(roleID int64) ([]*entity.API, error) {
	// 获取所有 API
	allAPIs, err := s.apiRepo.GetAllRoutes()
	if err != nil {
		return nil, err
	}

	// 获取角色已直接授权的 API
	roleAPIs, err := s.roleRepo.GetsByRole(roleID)
	if err != nil {
		return nil, err
	}

	// 获取角色关联的菜单（包括 type=2 的菜单和 type=3 的按钮）
	menus, err := s.roleRepo.GetMenusByRole(roleID)
	if err != nil {
		return nil, err
	}
	menuAPIMap := make(map[int64]bool)
	for _, menu := range menus {
		for _, apiID := range menu.APIRouteIds {
			menuAPIMap[apiID] = true
		}
	}

	// 构建角色已关联的 API 集合（包括直接授权、菜单关联、按钮关联）
	roleAPIMap := make(map[int64]bool)
	for _, api := range roleAPIs {
		roleAPIMap[api.ID] = true
	}
	for apiID := range menuAPIMap {
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
func (s *RolePermissionService) logPermissionChange(tx *gorm.DB, roleID int64, operatorID string,
	oldMenus []*entity.Menu, oldAPIs []*entity.API,
	newMenuIds []int64, newAPIIds []int64) {

	// 构建变更数据
	beforeData, _ := json.Marshal(map[string]interface{}{
		"menus": oldMenus,
		"apis":  oldAPIs,
	})

	afterData, _ := json.Marshal(map[string]interface{}{
		"menus": newMenuIds,
		"apis":  newAPIIds,
	})

	// 使用实体创建权限变更日志记录，字段通过 gorm 标签映射到表列
	// Omit 排除 SnowflakeBaseModelWithoutDeleted 中表中不存在的列（created_by, updated_by, updated_at）
	log := &entity.PermissionLog{
		OperatorID:   0, // 简化处理，实际应从用户服务获取
		OperatorName: operatorID,
		ActionType:   "SAVE_ROLE_PERMISSIONS",
		TargetType:   "role",
		TargetID:     roleID,
		BeforeData:   string(beforeData),
		AfterData:    string(afterData),
	}

	// 记录日志不阻塞主流程，失败只记录警告
	if err := tx.Omit("created_by", "updated_by", "updated_at").Create(log).Error; err != nil {
		s.log.Warn("记录权限变更日志失败", "error", err, "roleID", roleID)
	}
}

// LoadRolePermissionsToRedis 从数据库加载角色权限到 Redis 缓存
func (s *RolePermissionService) LoadRolePermissionsToRedis(roleID string) error {
	id, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的角色 ID：%s", roleID)
	}

	s.log.Info("加载角色权限到 Redis 缓存", "roleID", id)

	// 从数据库加载角色权限
	menus, err := s.roleRepo.GetMenusByRole(id)
	if err != nil {
		return fmt.Errorf("加载角色菜单权限失败：%w", err)
	}

	apiRoutes, err := s.roleRepo.GetsByRole(id)
	if err != nil {
		return fmt.Errorf("加载角色 API 权限失败：%w", err)
	}

	// 构建 apiSet（用于快速查找）
	apiSet := make(map[string]bool)
	for _, api := range apiRoutes {
		key := api.Method + ":" + api.Path
		apiSet[key] = true
	}

	// 构建缓存数据
	perms := entity.RolePermissions{
		Menus:  menus,
		APIs:   apiRoutes,
		ApiSet: apiSet,
	}

	// 序列化并缓存
	jsonData, err := json.Marshal(perms)
	if err != nil {
		return fmt.Errorf("序列化角色权限失败：%w", err)
	}

	cacheKey := fmt.Sprintf("role:perms:%d", id)
	if err := s.cache.Set(cacheKey, string(jsonData), 24*time.Hour); err != nil {
		return fmt.Errorf("缓存角色权限失败：%w", err)
	}

	s.log.Info("角色权限缓存加载成功", "roleID", id, "menuCount", len(menus), "apiCount", len(apiRoutes))
	return nil
}

// GetRolePermissionsFromRedis 从 Redis 获取角色权限缓存
func (s *RolePermissionService) GetRolePermissionsFromRedis(roleID string) (*entity.RolePermissions, error) {
	cacheKey := fmt.Sprintf("role:perms:%s", roleID)

	data, err := s.cache.Get(cacheKey)
	if err != nil {
		return nil, nil // 缓存未命中不返回错误
	}
	if data == "" {
		return nil, nil
	}

	var perms entity.RolePermissions
	if err := json.Unmarshal([]byte(data), &perms); err != nil {
		return nil, fmt.Errorf("解析角色权限缓存失败：%w", err)
	}

	return &perms, nil
}

// ClearRolePermissionsCache 清除角色权限缓存
func (s *RolePermissionService) ClearRolePermissionsCache(roleID int64) error {
	cacheKey := fmt.Sprintf("role:perms:%d", roleID)

	err := s.cache.Delete(cacheKey)
	if err != nil {
		return fmt.Errorf("清除角色权限缓存失败：%w", err)
	}

	s.log.Info("角色权限缓存已清除", "roleID", roleID)
	return nil
}

// ClearRoleCacheByMenuID 清除包含该菜单的所有角色缓存
func (s *RolePermissionService) ClearRoleCacheByMenuID(menuID int64) error {
	roles, err := s.roleRepo.GetRolesByMenu(menuID)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if err := s.ClearRolePermissionsCache(role.ID); err != nil {
			s.log.Error("清除角色缓存失败", "roleID", role.ID, "error", err)
		}
	}

	return nil
}

// ClearRoleCacheByAPIID 清除包含该 API 的所有角色缓存
func (s *RolePermissionService) ClearRoleCacheByAPIID(apiID int64) error {
	roles, err := s.roleRepo.GetRolesBy(apiID)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if err := s.ClearRolePermissionsCache(role.ID); err != nil {
			s.log.Error("清除角色缓存失败", "roleID", role.ID, "error", err)
		}
	}

	return nil
}

// filterValidAPIIDs 过滤出在 base_apis 表中存在的 API ID，跳过不存在的 ID 避免外键约束错误
func (s *RolePermissionService) filterValidAPIIDs(tx *gorm.DB, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return ids, nil
	}

	var validIDs []int64
	if err := tx.Table("base_apis").Where("id IN ? AND status = ?", ids, 1).Pluck("id", &validIDs).Error; err != nil {
		return nil, err
	}

	// 如果有无效的 ID，记录警告日志
	if len(validIDs) < len(ids) {
		invalidIDs := make([]int64, 0, len(ids)-len(validIDs))
		validMap := make(map[int64]bool, len(validIDs))
		for _, id := range validIDs {
			validMap[id] = true
		}
		for _, id := range ids {
			if !validMap[id] {
				invalidIDs = append(invalidIDs, id)
			}
		}
		s.log.Warn("跳过不存在的 API ID", "valid_count", len(validIDs), "invalid_count", len(invalidIDs), "invalid_ids", invalidIDs)
	}

	return validIDs, nil
}