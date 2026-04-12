package app

import (
	"github.com/gin-gonic/gin"
)

// setupRoutes 设置路由
func (a *Application) setupRoutes() {

	// 健康检查
	a.router.GET("/", func(c *gin.Context) {
		c.JSON(200, "ixpay-pro")
	})

	// 健康检查
	a.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, "OK")
	})

}
