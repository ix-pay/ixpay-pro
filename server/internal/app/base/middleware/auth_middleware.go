package middleware

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	httpresponse "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/http"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtAuth *auth.JWTAuth, cacheClient cache.Cache, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization 头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httpresponse.UnauthorizedResponse(c, "Authorization header is required")
			c.Abort()
			return
		}

		// 检查令牌格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			httpresponse.UnauthorizedResponse(c, "Authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		// 解析和验证令牌
		claims, err := jwtAuth.ParseToken(parts[1])
		if err != nil {
			httpresponse.UnauthorizedResponse(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 检查用户是否已退出登录
		blacklistKey := "blacklist:user:" + claims.UserID
		exists, err := cacheClient.Exists(blacklistKey)
		if err != nil {
			httpresponse.InternalServerErrorResponse(c, "Failed to check user status")
			c.Abort()
			return
		}

		if exists {
			httpresponse.UnauthorizedResponse(c, "User has logged out")
			c.Abort()
			return
		}

		// 将用户信息添加到上下文
		c.Set("userID", claims.UserID)
		c.Set("userName", claims.Username)
		c.Set("nickname", claims.Nickname)
		c.Set("loginType", claims.LoginType)
		c.Set("claims", claims)

		// 【关键修改】从缓存获取用户的当前角色，实现故障降级策略
		currentRoleID, roleSource := getCurrentRoleFromCache(claims, cacheClient, c, log)

		// 根据角色来源设置上下文
		if roleSource == "cache" {
			// 成功从缓存读取角色
			c.Set("currentRoleId", currentRoleID)
			log.Info("角色来源：Redis 缓存", "userID", claims.UserID, "roleID", currentRoleID)
		} else if roleSource == "jwt" {
			// 降级使用 JWT 中的角色
			c.Set("currentRoleId", "")
			log.Warn("角色来源：JWT 降级", "userID", claims.UserID, "role", claims.Role)
		} else if roleSource == "fallback" {
			// 完全降级，使用默认角色
			c.Set("currentRoleId", "")
			log.Warn("角色来源：默认降级", "userID", claims.UserID)
		}

		c.Next()
	}
}

// getCurrentRoleFromCache 从缓存获取用户的当前角色，实现故障降级策略
// 返回：当前角色 ID, 角色来源 (cache/jwt/fallback)
func getCurrentRoleFromCache(claims *auth.Claims, cacheClient cache.Cache, c *gin.Context, log logger.Logger) (string, string) {
	// 安全检查：确保 claims 不为 nil
	if claims == nil {
		log.Error("claims 为 nil，无法获取角色信息")
		c.Set("role", "user") // 默认角色
		return "", "fallback"
	}

	// 步骤 1: 从 Redis 读取 user:current_role:{userID}
	currentRoleKey := fmt.Sprintf("user:current_role:%s", claims.UserID)

	currentRoleID, err := cacheClient.Get(currentRoleKey)
	if err != nil {
		// Redis 读取失败，记录错误日志
		log.Error("Redis 读取当前角色失败", "userID", claims.UserID, "key", currentRoleKey, "err", err.Error())
		// 降级策略 1: 尝试使用 JWT 中的角色
		if claims.Role != "" {
			c.Set("role", claims.Role)
			return "", "jwt"
		}
		// 降级策略 2: 使用默认角色或返回错误
		c.Set("role", "user") // 默认角色
		return "", "fallback"
	}

	if currentRoleID == "" {
		// 缓存中没有当前角色信息，尝试使用 JWT 角色
		log.Debug("缓存中无当前角色信息，使用 JWT 角色", "userID", claims.UserID)
		if claims.Role != "" {
			c.Set("role", claims.Role)
			return "", "jwt"
		}
		c.Set("role", "user") // 默认角色
		return "", "fallback"
	}

	// 步骤 2: 从 Redis 读取 role:info:{roleID} 获取角色详情
	role, err := getRoleByID(currentRoleID, cacheClient, log)
	if err != nil {
		// 获取角色详情失败，记录错误日志
		log.Error("获取角色详情失败", "userID", claims.UserID, "roleID", currentRoleID, "err", err.Error())
		// 降级策略 1: 尝试使用 JWT 中的角色
		if claims.Role != "" {
			c.Set("role", claims.Role)
			return "", "jwt"
		}
		// 降级策略 2: 使用默认角色
		c.Set("role", "user")
		return "", "fallback"
	}

	if role == nil || role.Code == "" {
		// 角色信息无效，记录警告日志
		log.Warn("角色信息无效", "userID", claims.UserID, "roleID", currentRoleID, "role", role)
		// 降级策略 1: 尝试使用 JWT 中的角色
		if claims.Role != "" {
			c.Set("role", claims.Role)
			return "", "jwt"
		}
		// 降级策略 2: 使用默认角色
		c.Set("role", "user")
		return "", "fallback"
	}

	// 步骤 3: 成功从缓存读取角色，设置到上下文
	c.Set("role", role.Code)
	log.Info("从缓存加载角色成功", "userID", claims.UserID, "roleID", currentRoleID, "roleCode", role.Code, "roleName", role.Name)

	return currentRoleID, "cache"
}

// getRoleByID 根据角色 ID 获取角色信息
// 从 Redis 缓存读取 role:info:{roleID}
func getRoleByID(roleID string, cacheClient cache.Cache, log logger.Logger) (*entity.Role, error) {
	// 从缓存获取角色信息
	cacheKey := fmt.Sprintf("role:info:%s", roleID)
	roleData, err := cacheClient.Get(cacheKey)

	if err != nil {
		// Redis 读取失败
		log.Error("Redis 读取角色信息失败", "roleID", roleID, "key", cacheKey, "err", err)
		return nil, err
	}

	if roleData == "" {
		// 缓存中没有该角色信息
		log.Debug("缓存中无角色信息", "roleID", roleID)
		return nil, nil
	}

	// 从缓存解析角色信息
	var simpleRole map[string]interface{}
	if err := json.Unmarshal([]byte(roleData), &simpleRole); err != nil {
		log.Error("解析角色数据失败", "roleID", roleID, "data", roleData, "err", err)
		return nil, err
	}

	role := &entity.Role{}

	// 安全地提取字段
	if id, ok := simpleRole["ID"].(string); ok {
		parsedID, err := strconv.ParseInt(id, 10, 64)
		if err == nil {
			role.ID = parsedID
		}
	}
	if code, ok := simpleRole["Code"].(string); ok {
		role.Code = code
	}
	if name, ok := simpleRole["Name"].(string); ok {
		role.Name = name
	}
	if status, ok := simpleRole["Status"].(float64); ok {
		role.Status = int(status)
	}

	// 只返回有 Code 的角色
	if role.Code == "" {
		log.Warn("角色缺少 Code 字段", "roleID", roleID)
		return nil, nil
	}

	log.Debug("成功获取角色信息", "roleID", roleID, "code", role.Code, "name", role.Name)
	return role, nil
}
