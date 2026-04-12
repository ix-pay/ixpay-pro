package baseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/monitor"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// MonitorController 监控控制器
type MonitorController struct {
	systemMonitor   *monitor.SystemMonitor
	cacheMonitor    *monitor.CacheMonitor
	databaseMonitor *monitor.DatabaseMonitor
}

// NewMonitorController 创建监控控制器实例
func NewMonitorController(
	systemMonitor *monitor.SystemMonitor,
	cacheMonitor *monitor.CacheMonitor,
	databaseMonitor *monitor.DatabaseMonitor,
) *MonitorController {
	return &MonitorController{
		systemMonitor:   systemMonitor,
		cacheMonitor:    cacheMonitor,
		databaseMonitor: databaseMonitor,
	}
}

// GetSystemMonitor 获取系统监控信息
// @Summary 获取系统监控信息
// @Description 获取系统资源监控信息，包括 CPU、内存、磁盘等
// @Tags 系统监控
// @Accept json
// @Produce json
// @Success 200 {object} baseRes.Response{data=monitor.SystemStatus} "成功返回系统监控信息"
// @Failure 500 {object} baseRes.Response "服务器错误"
// @Router /api/admin//monitor/system [get]
func (c *MonitorController) GetSystemMonitor(ctx *gin.Context) {
	status, err := c.systemMonitor.GetSystemStatus()
	if err != nil {
		baseRes.FailWithMessage("获取系统监控信息失败："+err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(status, "获取系统监控信息成功", ctx)
}

// GetCacheMonitor 获取缓存监控信息
// @Summary 获取缓存监控信息
// @Description 获取缓存监控信息
// @Tags 系统监控
// @Accept json
// @Produce json
// @Success 200 {object} baseRes.Response{data=object} "成功返回缓存监控信息"
// @Failure 500 {object} baseRes.Response "服务器错误"
// @Router /api/admin//monitor/cache [get]
func (c *MonitorController) GetCacheMonitor(ctx *gin.Context) {
	status, err := c.cacheMonitor.GetCacheStatus()
	if err != nil {
		baseRes.FailWithMessage("获取缓存监控信息失败："+err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(status, "获取缓存监控信息成功", ctx)
}

// GetDatabaseMonitor 获取数据库监控信息
// @Summary 获取数据库监控信息
// @Description 获取数据库监控信息
// @Tags 系统监控
// @Accept json
// @Produce json
// @Success 200 {object} baseRes.Response{data=object} "成功返回数据库监控信息"
// @Failure 500 {object} baseRes.Response "服务器错误"
// @Router /api/admin//monitor/database [get]
func (c *MonitorController) GetDatabaseMonitor(ctx *gin.Context) {
	status, err := c.databaseMonitor.GetDatabaseStatus()
	if err != nil {
		baseRes.FailWithMessage("获取数据库监控信息失败："+err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(status, "获取数据库监控信息成功", ctx)
}

// GetRedisKeys 查询 Redis 键
// @Summary 查询 Redis 键
// @Description 查询 Redis 键列表
// @Tags 系统监控
// @Accept json
// @Produce json
// @Param pattern query string false "键模式 (默认：*)"
// @Param limit query int false "限制数量 (默认：100)"
// @Success 200 {object} baseRes.Response{data=object} "成功返回 Redis 键列表"
// @Failure 500 {object} baseRes.Response "服务器错误"
// @Router /api/admin//monitor/redis-keys [get]
func (c *MonitorController) GetRedisKeys(ctx *gin.Context) {
	pattern := ctx.DefaultQuery("pattern", "*")
	limit := int64(100)

	keys, total, err := c.cacheMonitor.GetRedisKeys(pattern, limit)
	if err != nil {
		baseRes.FailWithMessage("查询 Redis 键失败："+err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(gin.H{
		"keys":  keys,
		"total": total,
	}, "查询 Redis 键成功", ctx)
}

// GetSlowQueries 查询慢查询日志
// @Summary 查询慢查询日志
// @Description 查询数据库慢查询日志
// @Tags 系统监控
// @Accept json
// @Produce json
// @Success 200 {object} baseRes.Response{data=object} "成功返回慢查询日志"
// @Failure 500 {object} baseRes.Response "服务器错误"
// @Router /api/admin//monitor/slow-queries [get]
func (c *MonitorController) GetSlowQueries(ctx *gin.Context) {
	// TODO: 实现慢查询日志查询
	baseRes.OkWithDetailed([]interface{}{}, "查询慢查询日志成功", ctx)
}
