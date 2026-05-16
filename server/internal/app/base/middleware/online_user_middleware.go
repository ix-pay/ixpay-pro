package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// OnlineUserActiveMiddleware 更新在线用户活跃时间的中间件
func OnlineUserActiveMiddleware(onlineUserService *service.OnlineUserService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if exists {
			userIDStr, ok := userIDVal.(string)
			if ok {
				userID, err := strconv.ParseInt(userIDStr, 10, 64)
				if err == nil {
					if err := onlineUserService.UpdateUserActive(userID); err != nil {
						log.Debug("更新用户活跃时间失败", "userID", userID, "error", err)
					}
				}
			}
		}
		c.Next()
	}
}
