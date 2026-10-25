package app

import (
	"github.com/gin-gonic/gin"
)

// setupRoutes 设置路由
func (a *Application) setupRoutes() {

	// 根路由
	a.router.GET("/", func(c *gin.Context) {
		c.JSON(200, "ixpay-pro")
	})

	// 注意：/health 路由在 setupHealthCheck 中定义，以提供更详细的状态信息
}

// setupHealthCheck 设置健康检查路由（带详细信息）
func (a *Application) setupHealthCheck() {
	a.router.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"status":   "UP",
			"service":  a.cfg.Gateway.ServiceName,
			"nodeRole": a.cfg.Server.NodeRole,
			"nodeID":   a.cfg.Server.NodeID,
		}

		if a.cfg.Gateway.Enabled && a.gatewayClient != nil {
			instance := a.gatewayClient.GetInstance()
			if instance != nil {
				status["gatewayRegistered"] = true
				status["instanceID"] = instance.ID
			} else {
				status["gatewayRegistered"] = false
			}
		} else {
			status["gatewayRegistered"] = false
		}

		c.JSON(200, status)
	})
}
