package middleware

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	httpresponse "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/http"
)

// PermissionMiddleware 权限中间件 - 支持多角色、基于菜单/API 的权限验证和按钮级权限
func PermissionMiddleware(permissionService *service.PermissionService, roleRepo repo.RoleRepository, log logger.Logger, cacheClient cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查认证
		userID, exists := c.Get("userID")
		if !exists || userID == "" {
			httpresponse.UnauthorizedResponse(c, "未授权")
			c.Abort()
			return
		}

		// 获取当前角色
		roleValue, roleExists := c.Get("role")
		if !roleExists {
			log.Error("✗ 角色不存在于上下文中", "path", path, "method", method, "userID", userID)
			httpresponse.UnauthorizedResponse(c, "未找到角色")
			c.Abort()
			return
		}
		// 确保角色是字符串类型
		role := ""
		switch v := roleValue.(type) {
		case string:
			role = v
		default:
			role = fmt.Sprintf("%v", v)
		}

		log.Info("✓ 权限检查开始", "userID", userID, "role", role, "roleType", fmt.Sprintf("%T", role), "path", path, "method", method)

		// 【新增】检查是否为管理员
		if role == "admin" {
			// 管理员角色拥有所有权限，直接放行
			log.Debug("✓ 管理员角色，跳过权限验证", "path", path, "method", method)
			c.Next()
			return
		}

		// 【新增】检查 API 的授权类型
		authType, err := getAPIAuthType(roleRepo, path, method, log, cacheClient)
		if err != nil {
			log.Error("获取 API 授权类型失败", "error", err, "path", path, "method", method)
			httpresponse.InternalServerErrorResponse(c, "获取 API 授权类型失败")
			c.Abort()
			return
		}

		// auth_type = 0 表示不需要授权（只要登录就能用）
		if authType == 0 {
			log.Debug("✓ API 不需要授权，跳过权限验证", "path", path, "method", method, "auth_type", authType)
			c.Next()
			return
		}

		// 从缓存获取角色权限（auth_type = 1 需要验证角色权限）
		hasPermission, err := checkPermissionFromCache(roleRepo, role, method, path, log, cacheClient)
		if err != nil {
			log.Error("从缓存检查权限失败", "error", err, "role", role, "path", path, "method", method)
			httpresponse.InternalServerErrorResponse(c, "检查权限失败")
			c.Abort()
			return
		}

		// 权限验证失败
		if !hasPermission {
			httpresponse.ForbiddenResponse(c, "禁止访问")
			c.Abort()
			return
		}

		// 获取用户的按钮权限（用于按钮级权限控制）
		var userIDInt int64
		switch v := userID.(type) {
		case string:
			userIDInt, _ = strconv.ParseInt(v, 10, 64)
		case int:
			userIDInt = int64(v)
		case int64:
			userIDInt = v
		default:
			userIDInt = 0
		}

		userButtons, err := getBtnPermsByUserId(userIDInt, permissionService, log)
		if err != nil {
			log.Error("获取用户按钮权限失败", "error", err, "userID", userIDInt)
		}

		// 将按钮权限存储在上下文中
		c.Set("userButtons", userButtons)

		c.Next()
	}
}

// getAPIAuthType 获取 API 的授权类型
// 返回值：
// - 0: 不需要授权（只要登录就能用）
// - 1: 需要授权（需要角色权限）
// - -1: 获取失败或 API 不存在
func getAPIAuthType(roleRepo repo.RoleRepository, path, method string, log logger.Logger, cacheClient cache.Cache) (int, error) {
	// 构建缓存 Key
	cacheKey := fmt.Sprintf("api:auth_type:%s:%s", method, path)

	// 从缓存获取缓存
	data, err := cacheClient.Get(cacheKey)
	if err == nil && data != "" {
		// 解析缓存数据
		var authType int
		if err := json.Unmarshal([]byte(data), &authType); err == nil {
			log.Debug("API 授权类型缓存命中", "path", path, "method", method, "auth_type", authType)
			return authType, nil
		}
	}

	// 缓存未命中，从数据库查询
	api, err := roleRepo.GetAPIByPathAndMethod(path, method)
	if err != nil {
		log.Error("查询 API 信息失败", "error", err, "path", path, "method", method)
		return -1, err
	}

	var authType int
	if api == nil {
		// API 不存在，默认需要授权
		log.Warn("API 不存在，默认需要授权", "path", path, "method", method)
		authType = 1
	} else {
		authType = api.AuthType
		log.Debug("API 授权类型查询成功", "path", path, "method", method, "auth_type", authType)
	}

	// 缓存（5 分钟）
	authTypeJSON, _ := json.Marshal(authType)
	cacheClient.Set(cacheKey, string(authTypeJSON), 5*time.Minute)

	return authType, nil
}

// RolePermissions Redis 缓存的角色权限结构
type RolePermissions struct {
	Menus     []*entity.Menu    `json:"menus"`
	BtnPerms  []*entity.BtnPerm `json:"btnPerms"`
	ApiRoutes []*entity.API     `json:"apiRoutes"`
	ApiSet    map[string]bool   `json:"apiSet"` // 快速查找的 API 权限集合
}

// checkPermissionFromCache 从缓存检查角色权限
func checkPermissionFromCache(roleRepo repo.RoleRepository, role, method, path string, log logger.Logger, cacheClient cache.Cache) (bool, error) {
	// 获取角色 ID（通过角色编码）
	roleObj, err := roleRepo.GetByCode(role)
	if err != nil {
		return false, fmt.Errorf("获取角色失败：%w", err)
	}

	if roleObj == nil {
		// 角色不存在，返回 false
		log.Warn("角色不存在", "role", role)
		return false, nil
	}

	// 构建缓存 Key
	cacheKey := fmt.Sprintf("role:perms:%d", roleObj.ID)

	// 从缓存获取缓存数据
	data, err := cacheClient.Get(cacheKey)
	if err != nil || data == "" {
		// 缓存不存在，从数据库加载并缓存
		log.Info("角色权限缓存未命中，从数据库加载", "roleID", roleObj.ID, "role", role)
		return loadAndCacheRolePermissions(roleObj.ID, roleRepo, log, cacheClient, method, path)
	}

	// 解析缓存数据
	var perms RolePermissions
	if err := json.Unmarshal([]byte(data), &perms); err != nil {
		log.Error("解析角色权限缓存失败", "error", err, "roleID", roleObj.ID)
		// 缓存数据损坏，重新加载
		return loadAndCacheRolePermissions(roleObj.ID, roleRepo, log, cacheClient, method, path)
	}

	// 使用 apiSet 快速验证（O(1) 时间复杂度）
	apiKey := method + ":" + path
	hasPermission := perms.ApiSet[apiKey]

	if hasPermission {
		log.Debug("权限验证通过", "roleID", roleObj.ID, "role", role, "path", path, "method", method)
	} else {
		log.Debug("权限验证失败", "roleID", roleObj.ID, "role", role, "path", path, "method", method)
	}

	return hasPermission, nil
}

// loadAndCacheRolePermissions 从数据库加载角色权限并缓存
func loadAndCacheRolePermissions(roleID int64, roleRepo repo.RoleRepository, log logger.Logger, cacheClient cache.Cache, method, path string) (bool, error) {
	// 从数据库加载角色权限
	menus, err := roleRepo.GetMenusByRole(roleID)
	if err != nil {
		return false, fmt.Errorf("加载角色菜单权限失败：%w", err)
	}

	btnPerms, err := roleRepo.GetBtnPermsByRole(roleID)
	if err != nil {
		return false, fmt.Errorf("加载角色按钮权限失败：%w", err)
	}

	apiRoutes, err := roleRepo.GetsByRole(roleID)
	if err != nil {
		return false, fmt.Errorf("加载角色 API 权限失败：%w", err)
	}

	// 构建 apiSet（用于快速查找）
	apiSet := make(map[string]bool)

	// 添加直接授权的 API
	for _, api := range apiRoutes {
		key := api.Method + ":" + api.Path
		apiSet[key] = true
	}

	// 构建缓存数据
	perms := RolePermissions{
		Menus:     menus,
		BtnPerms:  btnPerms,
		ApiRoutes: apiRoutes,
		ApiSet:    apiSet,
	}

	// 序列化并缓存
	jsonData, err := json.Marshal(perms)
	if err != nil {
		log.Error("序列化角色权限失败", "error", err, "roleID", roleID)
		return false, fmt.Errorf("序列化角色权限失败：%w", err)
	}

	cacheKey := fmt.Sprintf("role:perms:%d", roleID)
	// 缓存 24 小时
	if err := cacheClient.Set(cacheKey, string(jsonData), 24*time.Hour); err != nil {
		log.Error("缓存角色权限失败", "error", err, "roleID", roleID)
	} else {
		log.Info("角色权限已缓存", "roleID", roleID, "expire", "24h")
	}

	// 验证权限
	apiKey := method + ":" + path
	return apiSet[apiKey], nil
}

// getBtnPermsByUserId 根据用户 ID 获取按钮权限
func getBtnPermsByUserId(userId int64, permissionService *service.PermissionService, log logger.Logger) ([]string, error) {
	// 通过用户 ID 获取角色
	roles, err := permissionService.GetRolesByUserId(userId)
	if err != nil {
		log.Error("通过用户 ID 获取角色失败", "error", err, "userId", userId)
		return nil, err
	}

	// 如果角色列表为空，返回空的按钮权限列表
	if len(roles) == 0 {
		log.Info("用户没有角色", "userId", userId)
		return []string{}, nil
	}

	// 存储用户所有的按钮权限编码
	btnPerms := make(map[string]bool)

	// 获取每个角色的按钮权限
	for _, role := range roles {
		buttons, err := permissionService.GetBtnPermsByRole(role.ID)
		if err != nil {
			log.Error("通过角色获取按钮权限失败", "error", err, "roleID", role.ID)
			continue
		}

		// 添加到结果集
		for _, button := range buttons {
			if button != nil && button.Status == 1 { // 只添加启用状态的按钮权限
				btnPerms[button.Code] = true
			}
		}
	}

	// 转换为切片返回
	result := make([]string, 0, len(btnPerms))
	for code := range btnPerms {
		result = append(result, code)
	}

	return result, nil
}

// RolePermissionMiddleware 基于角色的权限中间件
// 用于快速验证特定角色是否有权限访问
func RolePermissionMiddleware(requiredRoles []string, roleRepo repo.RoleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前请求的路径
		path := c.Request.URL.Path

		// 检查路径是否需要权限控制
		if strings.HasPrefix(path, "/swagger") ||
			strings.HasPrefix(path, "/api//auth") ||
			strings.HasPrefix(path, "/api//health") ||
			strings.HasPrefix(path, "/api//pay/notify") {
			c.Next()
			return
		}

		// 从 gin.Context 中获取用户 ID
		userID, exists := c.Get("userID")
		if !exists {
			httpresponse.UnauthorizedResponse(c, "用户未认证")
			c.Abort()
			return
		}

		// 尝试将 userID 转换为 int64 类型
		var userIDInt int64
		switch v := userID.(type) {
		case string:
			userIDInt, _ = strconv.ParseInt(v, 10, 64)
		case int:
			userIDInt = int64(v)
		case int64:
			userIDInt = v
		default:
			httpresponse.BadRequestResponse(c, "用户 ID 格式无效")
			c.Abort()
			return
		}

		// 获取用户所有角色
		roles, err := roleRepo.GetRolesByUser(userIDInt)
		if err != nil {
			httpresponse.InternalServerErrorResponse(c, "获取用户角色失败")
			c.Abort()
			return
		}

		// 检查用户是否有任何一个所需角色
		hasRequiredRole := false
		for _, userRole := range roles {
			for _, requiredRole := range requiredRoles {
				if userRole.Code == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		// 如果没有所需角色，拒绝访问
		if !hasRequiredRole {
			httpresponse.ForbiddenResponse(c, fmt.Sprintf("需要角色：%s", strings.Join(requiredRoles, ", ")))
			c.Abort()
			return
		}

		c.Next()
	}
}
