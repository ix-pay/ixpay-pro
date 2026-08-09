package middleware

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/dictconst"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	httpresponse "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/http"
)

// PermissionMiddleware 权限中间件 - 支持 RBAC+ABAC 混合权限验证
// 权限检查流程：
// 1. 管理员直接放行
// 2. 检查 API 授权类型（auth_type=0 跳过）
// 3. RBAC 检查：角色是否有该 API 权限（从缓存读取）
// 4. ABAC 拒绝规则检查：如果 RBAC 通过，检查是否有 ABAC 拒绝规则匹配
// 5. ABAC 允许规则检查：如果 RBAC 未通过，检查 ABAC 允许规则是否匹配
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

		// 1. 检查是否为管理员
		if role == dictconst.UserTypeAdmin {
			log.Debug("✓ 管理员角色，跳过权限验证", "path", path, "method", method)
			c.Next()
			return
		}

		// 2. 检查 API 的授权类型
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

		// 将 userID 转换为 int64
		var uid int64
		switch v := userID.(type) {
		case string:
			uid, _ = strconv.ParseInt(v, 10, 64)
		case int:
			uid = int64(v)
		case int64:
			uid = v
		}

		// 3. RBAC 检查：从缓存获取角色权限
		hasPermission, err := checkPermissionFromCache(roleRepo, role, method, path, log, cacheClient)
		if err != nil {
			log.Error("从缓存检查权限失败", "error", err, "role", role, "path", path, "method", method)
			httpresponse.InternalServerErrorResponse(c, "检查权限失败")
			c.Abort()
			return
		}

		// 4. ABAC 拒绝规则检查：如果 RBAC 通过，检查是否有 ABAC 拒绝规则匹配
		// 拒绝规则优先于允许规则，即使 RBAC 允许，如果 ABAC 拒绝规则匹配，也拒绝访问
		if hasPermission {
			denied, err := checkABACDenyRules(permissionService, uid, path, method, log)
			if err != nil {
				log.Error("ABAC 拒绝规则检查失败", "error", err, "userID", uid, "path", path, "method", method)
				httpresponse.InternalServerErrorResponse(c, "ABAC 拒绝规则检查失败")
				c.Abort()
				return
			}
			if denied {
				log.Debug("✗ ABAC 拒绝规则阻止访问", "path", path, "method", method, "userID", userID)
				httpresponse.ForbiddenResponse(c, "禁止访问")
				c.Abort()
				return
			}
			log.Debug("✓ RBAC 权限验证通过，ABAC 无拒绝规则", "path", path, "method", method)
			c.Next()
			return
		}

		// 5. ABAC 允许规则检查：如果 RBAC 未通过，检查 ABAC 允许规则
		allowed, err := checkABACAllowRules(permissionService, uid, path, method, log)
		if err != nil {
			log.Error("ABAC 允许规则检查失败", "error", err, "userID", uid, "path", path, "method", method)
			httpresponse.InternalServerErrorResponse(c, "ABAC 允许规则检查失败")
			c.Abort()
			return
		}

		if !allowed {
			log.Debug("✗ 权限验证失败（RBAC 和 ABAC 均未通过）", "path", path, "method", method, "userID", userID)
			httpresponse.ForbiddenResponse(c, "禁止访问")
			c.Abort()
			return
		}

		log.Debug("✓ ABAC 允许规则通过", "path", path, "method", method, "userID", userID)
		c.Next()
	}
}

// checkABACDenyRules 检查 ABAC 拒绝规则
// 遍历用户的 ABAC 规则，检查是否有拒绝规则匹配当前请求
func checkABACDenyRules(permissionService *service.PermissionService, uid int64, path, method string, log logger.Logger) (bool, error) {
	rules, err := permissionService.GetPermissionRules(uid)
	if err != nil {
		log.Error("获取 ABAC 拒绝规则失败", "error", err, "userID", uid)
		return false, err
	}

	for _, rule := range rules {
		if rule.Effect == "deny" && rule.IsActive() {
			// 使用路径匹配检查规则是否适用于当前请求
			if matchAPIPath(rule.APIPath, path) && (rule.Method == "*" || rule.Method == method) {
				log.Debug("ABAC 拒绝规则匹配", "ruleId", rule.ID, "ruleName", rule.Name,
					"rulePath", rule.APIPath, "ruleMethod", rule.Method)
				// 如果有条件表达式，需要进一步检查条件是否满足
				// 当前简化处理：只要路径和方法匹配就拒绝
				return true, nil
			}
		}
	}
	return false, nil
}

// checkABACAllowRules 检查 ABAC 允许规则
// 遍历用户的 ABAC 规则，检查是否有允许规则匹配当前请求
func checkABACAllowRules(permissionService *service.PermissionService, uid int64, path, method string, log logger.Logger) (bool, error) {
	rules, err := permissionService.GetPermissionRules(uid)
	if err != nil {
		log.Error("获取 ABAC 允许规则失败", "error", err, "userID", uid)
		return false, err
	}

	for _, rule := range rules {
		if rule.Effect == "allow" && rule.IsActive() {
			// 使用路径匹配检查规则是否适用于当前请求
			if matchAPIPath(rule.APIPath, path) && (rule.Method == "*" || rule.Method == method) {
				log.Debug("ABAC 允许规则匹配", "ruleId", rule.ID, "ruleName", rule.Name,
					"rulePath", rule.APIPath, "ruleMethod", rule.Method)
				// 如果有条件表达式，需要进一步检查条件是否满足
				// 当前简化处理：只要路径和方法匹配就允许
				return true, nil
			}
		}
	}
	return false, nil
}

// matchAPIPath 检查请求路径是否匹配规则路径模式
// 支持：
//   - 精确匹配：/api/admin/user == /api/admin/user
//   - 通配符：/api/admin/** 匹配 /api/admin/user/123
//   - 参数匹配：/api/admin/user/:id 匹配 /api/admin/user/123
func matchAPIPath(pattern, requestPath string) bool {
	// 通配符匹配：pattern 以 ** 结尾
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(requestPath, prefix)
	}

	// 精确匹配
	if pattern == requestPath {
		return true
	}

	// 参数匹配：将 :param 替换为通配符再匹配
	patternParts := strings.Split(pattern, "/")
	requestParts := strings.Split(requestPath, "/")

	if len(patternParts) != len(requestParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			// :id, :key 等参数部分，匹配任意值
			continue
		}
		if part != requestParts[i] {
			return false
		}
	}
	return true
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
	Menus     []*entity.Menu  `json:"menus"`
	ApiRoutes []*entity.API   `json:"apiRoutes"`
	ApiSet    map[string]bool `json:"apiSet"` // 快速查找的 API 权限集合
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
