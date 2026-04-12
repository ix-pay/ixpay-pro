package middleware

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	httpresponse "github.com/ix-pay/ixpay-pro/internal/infrastructure/transport/http"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtAuth *auth.JWTAuth, cacheClient cache.Cache) gin.HandlerFunc {
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
		c.Set("username", claims.Username)
		c.Set("nickname", claims.Nickname)
		c.Set("loginType", claims.LoginType)
		c.Set("claims", claims)

		// 【关键修改】从缓存获取用户的当前角色，而不是使用 JWT 令牌中的角色
		currentRoleKey := fmt.Sprintf("user:current_role:%s", claims.UserID)
		currentRoleIDStr, err := cacheClient.Get(currentRoleKey)
		if err != nil {
			currentRoleIDStr = ""
		}

		var currentRole string
		var currentRoleID int64
		roleFromRedis := false

		if currentRoleIDStr != "" {
			// 缓存中有存储，尝试使用当前角色
			var err error
			currentRoleID, err = strconv.ParseInt(currentRoleIDStr, 10, 64)
			if err == nil {
				// 获取角色信息
				role, err := getRoleByID(currentRoleID, cacheClient)
				if err == nil && role != nil && role.Code != "" {
					currentRole = role.Code
					roleFromRedis = true
					fmt.Printf("✓ 从缓存加载角色：roleID=%d, roleCode=%s, roleName=%s\n", currentRoleID, role.Code, role.Name)
				} else {
					// 获取角色失败，记录日志并回退到 JWT 角色
					fmt.Printf("✗ 获取角色信息失败：roleID=%d, role=%v, err=%v，将使用 JWT 角色\n", currentRoleID, role, err)
				}
			} else {
				// 解析角色 ID 失败
				fmt.Printf("✗ 解析角色 ID 失败：currentRoleIDStr=%s, err=%v，将使用 JWT 角色\n", currentRoleIDStr, err)
			}
		} else {
			fmt.Printf("→ 缓存中没有当前角色信息，将使用 JWT 角色：userID=%s\n", claims.UserID)
		}

		// 如果缓存中没有存储或获取失败，使用 JWT 令牌中的角色（兼容旧逻辑）
		if !roleFromRedis {
			currentRole = claims.Role
			fmt.Printf("→ 使用 JWT 令牌中的角色：role=%s, userID=%s\n", currentRole, claims.UserID)
		}

		// 设置到上下文中
		c.Set("role", currentRole)
		c.Set("nickname", claims.Nickname)
		if roleFromRedis {
			c.Set("currentRoleId", currentRoleID)
		}

		c.Next()
	}
}

// getRoleByID 根据角色 ID 获取角色信息
func getRoleByID(roleID int64, cacheClient cache.Cache) (*entity.Role, error) {
	// 尝试从缓存获取角色信息
	cacheKey := fmt.Sprintf("role:info:%d", roleID)
	roleData, err := cacheClient.Get(cacheKey)
	if err != nil || roleData == "" {
		// 缓存中没有，返回 nil
		return nil, nil
	}

	// 从缓存解析角色信息（简化版）
	var simpleRole map[string]interface{}
	if err := json.Unmarshal([]byte(roleData), &simpleRole); err == nil {
		role := &entity.Role{}

		// 安全地提取字段
		if id, ok := simpleRole["ID"].(float64); ok {
			role.ID = fmt.Sprintf("%.0f", id)
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
		if role.Code != "" {
			return role, nil
		}
	}

	return nil, nil
}
