package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// AdminTenantAuth 管理端租户切换中间件（渠道配置/对账等跨租户查询用）
// 必须挂载在 AuthMiddleware 之后，依赖其注入的 userID/role 上下文（角色 Code，如 admin/user）。
// 说明：管理平台登录签发的 JWT 中 Role 为空串，角色以 Redis 缓存解析的 Code 为准，
// 因此此处不解析 JWT，统一从上下文取角色——与 AdminOnly、租户管理 isAdmin 语义一致。
// 规则（设计文档 01 §3.4）：
//   - 请求携带 ?tenant_id= 时：admin 角色直接放行任意租户；其他角色校验 base_tenant_grants 授权
//   - 未携带 tenant_id：默认租户（主库），不注入租户上下文
//   - 一次只查一个租户，前端通过切换器更换 tenant_id
func AdminTenantAuth(grantRepo repo.TenantRepository, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取当前角色（AuthMiddleware 注入的角色 Code）与用户ID
		roleValue, _ := c.Get("role")
		roleCode, _ := roleValue.(string)
		userIDValue, _ := c.Get("userID")
		userIDStr, _ := userIDValue.(string)

		// 解析目标租户（未传 = 默认租户主库）
		tenantIDStr := c.Query("tenant_id")
		if tenantIDStr == "" {
			// 未指定租户：默认主库（不注入租户上下文）
			c.Next()
			return
		}
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
		if err != nil || tenantID <= 0 {
			baseRes.FailWithMessage("tenant_id 格式错误", c)
			c.Abort()
			return
		}

		// 授权校验：admin 全量；其他角色按 base_tenant_grants
		if roleCode != "admin" {
			if _, err := grantRepo.GetGrant(parseUserID(userIDStr), tenantID); err != nil {
				log.Warn("用户访问未授权租户", "user_id", userIDStr, "tenant_id", tenantID, "role", roleCode)
				baseRes.NoAuth("无权访问该租户数据", c)
				c.Abort()
				return
			}
		}

		// 注入租户上下文（后续 GORM 查询路由到该租户库）
		c.Set("queryTenantID", tenantID)
		c.Request = c.Request.WithContext(database.WithTenantID(c.Request.Context(), tenantID))
		c.Next()
	}
}

// parseUserID 用户ID字符串转 int64（解析失败返回 0，由授权查询自然拒绝）
func parseUserID(userID string) int64 {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
