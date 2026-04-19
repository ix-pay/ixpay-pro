package baseapi

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// PermissionLogController 权限日志控制器
// 处理权限日志相关的 HTTP 请求
type PermissionLogController struct {
	service *service.PermissionLogService
	log     logger.Logger
}

// NewPermissionLogController 创建权限日志控制器实例
func NewPermissionLogController(service *service.PermissionLogService, log logger.Logger) *PermissionLogController {
	return &PermissionLogController{
		service: service,
		log:     log,
	}
}

// GetPermissionLogList 获取权限日志列表
//
//	@Summary		获取权限日志列表
//	@Description	获取系统权限变更日志（支持分页和筛选）
//	@Tags			权限日志管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int										true	"页码"
//	@Param			page_size	query		int										true	"每页数量"
//	@Param			user_id		query		int64									false	"用户 ID"
//	@Param			userName	query		string									false	"用户名"
//	@Param			operation	query		string									false	"操作类型"
//	@Param			module		query		string									false	"模块"
//	@Param			start_date	query		string									false	"开始日期（YYYY-MM-DD）"
//	@Param			end_date	query		string									false	"结束日期（YYYY-MM-DD）"
//	@Success		200			{object}	baseRes.Response{data=baseRes.PageResult{list=[]interface{}},msg=string}	"权限日志列表"
//	@Failure		400			{object}	map[string]string						"请求参数错误"
//	@Failure		401			{object}	map[string]string						"未授权"
//	@Failure		500			{object}	map[string]string						"服务器内部错误"
//	@Router			/api/admin/permission-logs [get]
func (c *PermissionLogController) GetPermissionLogList(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetPermissionLogListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 构建筛选条件
	filters := make(map[string]interface{})
	if req.UserID != nil {
		filters["user_id"] = *req.UserID
	}
	if req.Username != "" {
		filters["userName"] = req.Username
	}
	if req.Operation != "" {
		filters["operation"] = req.Operation
	}
	if req.Module != "" {
		filters["module"] = req.Module
	}
	if req.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			filters["start_time"] = startDate
		}
	}
	if req.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			filters["end_time"] = endDate.Add(24 * time.Hour)
		}
	}

	logs, total, err := c.service.GetPermissionLogList(req.Page, req.PageSize, filters)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// TODO: 需要将 logs 转换为 DTO
	pageResult := baseRes.PageResult{
		List:     logs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	baseRes.OkWithDetailed(pageResult, "获取权限日志列表成功", ctx)
}

// GetRolePermissionLogs 获取角色权限日志
//
//	@Summary		获取角色权限日志
//	@Description	获取指定角色的权限变更历史
//	@Tags			权限日志管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role_id		path		int										true	"角色 ID"
//	@Param			page		query		int										true	"页码"
//	@Param			page_size	query		int										true	"每页数量"
//	@Success		200			{object}	baseRes.Response{data=baseRes.PageResult{list=[]interface{}},msg=string}	"角色权限日志列表"
//	@Failure		400			{object}	map[string]string						"请求参数错误"
//	@Failure		401			{object}	map[string]string						"未授权"
//	@Failure		500			{object}	map[string]string						"服务器内部错误"
//	@Router			/api/admin/roles/:roleId/permission-logs [get]
func (c *PermissionLogController) GetRolePermissionLogs(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetRolePermissionLogsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	logs, total, err := c.service.GetRolePermissionLogs(req.RoleID, req.Page, req.PageSize)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// TODO: 需要将 logs 转换为 DTO
	pageResult := baseRes.PageResult{
		List:     logs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	baseRes.OkWithDetailed(pageResult, "获取角色权限日志成功", ctx)
}
