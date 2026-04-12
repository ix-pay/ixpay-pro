package baseapi

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// OperationLogController 操作日志控制器
//
//	@Summary		操作日志管理API
//	@Description	提供操作日志的查询、删除和统计功能
//	@Tags			系统管理
//	@Router			/api/admin/logs [get]
type OperationLogController struct {
	service *service.OperationLogService
}

// NewOperationLogController 创建操作日志控制器实例
func NewOperationLogController(service *service.OperationLogService) *OperationLogController {
	return &OperationLogController{
		service: service,
	}
}

// GetLogList 获取操作日志列表
//
//	@Summary		获取操作日志列表
//	@Description	获取系统操作日志列表，支持分页和多条件过滤
//	@Tags			系统管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page			query		int						true	"页码"
//	@Param			pageSize		query		int						true	"每页数量"
//	@Param			startTime		query		string					false	"开始时间 (格式:2006-01-02)"
//	@Param			endTime			query		string					false	"结束时间 (格式:2006-01-02)"
//	@Param			username		query		string					false	"用户名"
//	@Param			module			query		string					false	"操作模块"
//	@Param			operationType	query		int						false	"操作类型"
//	@Param			isSuccess		query		bool					false	"操作结果"
//	@Success		200				{object}	baseRes.Response{data=response.OperationLogListResponse}	"操作日志列表"
//	@Failure		400				{object}	map[string]string		"请求参数错误"
//	@Failure		500				{object}	map[string]string		"服务器内部错误"
//	@Router			/api/admin/logs [get]
func (c *OperationLogController) GetLogList(ctx *gin.Context) {
	var req request.GetOperationLogListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if req.StartTime != "" {
		filters["start_time"] = req.StartTime
	}
	if req.EndTime != "" {
		filters["end_time"] = req.EndTime
	}
	if req.Username != "" {
		filters["username"] = req.Username
	}
	if req.Module != "" {
		filters["module"] = req.Module
	}
	if req.OperationType != 0 {
		filters["operation_type"] = req.OperationType
	}
	if req.IsSuccess != nil {
		filters["is_success"] = *req.IsSuccess
	}

	// 获取日志列表
	logs, total, err := c.service.GetLogList(req.Page, req.PageSize, filters)
	if err != nil {
		baseRes.FailWithMessage("获取日志列表失败", ctx)
		return
	}

	// 转换为响应 DTO
	logResponses := make([]response.OperationLogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = response.OperationLogResponse{
			ID:            log.ID,
			UserID:        log.UserID,
			Username:      log.Username,
			Nickname:      log.Nickname,
			OperationType: int(log.OperationType),
			Module:        log.Module,
			Description:   log.Description,
			Method:        log.Method,
			Path:          log.Path,
			Params:        log.Params,
			ClientIP:      log.ClientIP,
			UserAgent:     log.UserAgent,
			StatusCode:    log.StatusCode,
			Result:        log.Result,
			Duration:      log.Duration,
			ErrorMessage:  log.ErrorMessage,
			IsSuccess:     log.IsSuccess,
			CreatedAt:     log.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     log.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	logListResponse := response.OperationLogListResponse{
		PageResult: baseRes.PageResult{
			List:     logResponses,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		List: logResponses,
	}

	baseRes.OkWithDetailed(logListResponse, "获取日志列表成功", ctx)
}

// GetLogByID 根据 ID 获取操作日志
//
//	@Summary		根据 ID 获取操作日志
//	@Description	根据日志 ID 获取详细的操作日志信息
//	@Tags			系统管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"日志 ID"
//	@Success		200	{object}	baseRes.Response{data=response.OperationLogResponse}	"操作日志详情"
//	@Failure		400	{object}	map[string]string	"请求参数错误"
//	@Failure		404	{object}	map[string]string	"日志不存在"
//	@Failure		500	{object}	map[string]string	"服务器内部错误"
//	@Router			/api/admin/logs/:id [get]
func (c *OperationLogController) GetLogByID(ctx *gin.Context) {
	// 解析 ID 参数
	id := ctx.Param("id")
	if id == "" {
		baseRes.FailWithMessage("无效的 ID 参数", ctx)
		return
	}

	// 获取日志详情
	log, err := c.service.GetLogByID(id)
	if err != nil {
		baseRes.FailWithMessage("获取日志详情失败", ctx)
		return
	}

	if log == nil {
		baseRes.FailWithMessage("日志不存在", ctx)
		return
	}

	// 转换为响应 DTO
	logResponse := response.OperationLogResponse{
		ID:            log.ID,
		UserID:        log.UserID,
		Username:      log.Username,
		Nickname:      log.Nickname,
		OperationType: int(log.OperationType),
		Module:        log.Module,
		Description:   log.Description,
		Method:        log.Method,
		Path:          log.Path,
		Params:        log.Params,
		ClientIP:      log.ClientIP,
		UserAgent:     log.UserAgent,
		StatusCode:    log.StatusCode,
		Result:        log.Result,
		Duration:      log.Duration,
		ErrorMessage:  log.ErrorMessage,
		IsSuccess:     log.IsSuccess,
		CreatedAt:     log.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     log.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	baseRes.OkWithDetailed(logResponse, "获取日志详情成功", ctx)
}

// DeleteLogByID 根据ID删除操作日志
//
//	@Summary		根据ID删除操作日志
//	@Description	根据日志ID删除指定的操作日志
//	@Tags			系统管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int					true	"日志ID"
//	@Success		200	{object}	map[string]string	"删除成功"
//	@Failure		400	{object}	map[string]string	"请求参数错误"
//	@Failure		500	{object}	map[string]string	"服务器内部错误"
//	@Router			/api/admin/logs/:id [delete]
func (c *OperationLogController) DeleteLogByID(ctx *gin.Context) {
	// 解析 ID 参数
	id := ctx.Param("id")
	if id == "" {
		baseRes.FailWithMessage("无效的 ID 参数", ctx)
		return
	}

	// 删除日志
	err := c.service.DeleteLogByID(id)
	if err != nil {
		baseRes.FailWithMessage("删除日志失败", ctx)
		return
	}

	baseRes.OkWithMessage("删除成功", ctx)
}

// BatchDeleteLog 批量删除操作日志
//
//	@Summary		批量删除操作日志
//	@Description	批量删除指定 ID 列表的操作日志
//	@Tags			系统管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		request.BatchDeleteOperationLogRequest	true	"批量删除请求参数"
//	@Success		200		{object}	map[string]string	"批量删除成功"
//	@Failure		400		{object}	map[string]string	"请求参数错误"
//	@Failure		500		{object}	map[string]string	"服务器内部错误"
//	@Router			/api/admin/logs/batch-delete [post]
func (c *OperationLogController) BatchDeleteLog(ctx *gin.Context) {
	// 解析请求体
	var req request.BatchDeleteOperationLogRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("无效的请求参数", ctx)
		return
	}

	// 批量删除日志
	err := c.service.BatchDeleteLog(req.IDs)
	if err != nil {
		baseRes.FailWithMessage("批量删除日志失败", ctx)
		return
	}

	baseRes.OkWithMessage("批量删除成功", ctx)
}

// GetLogStatistics 获取操作日志统计信息
//
//	@Summary		获取操作日志统计信息
//	@Description	获取操作日志的统计信息，包括操作类型分布等
//	@Tags			系统管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			startTime	query		string					false	"开始时间 (格式:2006-01-02)"
//	@Param			endTime		query		string					false	"结束时间 (格式:2006-01-02)"
//	@Success		200			{object}	baseRes.Response{data=response.OperationLogStatisticsResponse}	"操作日志统计信息"
//	@Failure		400			{object}	map[string]string		"请求参数错误"
//	@Failure		500			{object}	map[string]string		"服务器内部错误"
//	@Router			/api/admin/logs/statistics [get]
func (c *OperationLogController) GetLogStatistics(ctx *gin.Context) {
	var req request.GetOperationLogStatisticsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 解析时间参数
	var startTime, endTime time.Time
	var err error

	if req.StartTime != "" {
		startTime, err = time.Parse("2006-01-02", req.StartTime)
		if err != nil {
			baseRes.FailWithMessage("无效的开始时间格式", ctx)
			return
		}
	} else {
		// 默认查询最近 30 天的数据
		startTime = time.Now().AddDate(0, 0, -30)
	}

	if req.EndTime != "" {
		endTime, err = time.Parse("2006-01-02", req.EndTime)
		if err != nil {
			baseRes.FailWithMessage("无效的结束时间格式", ctx)
			return
		}
		endTime = endTime.Add(24*time.Hour - time.Second)
	} else {
		endTime = time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(-time.Second)
	}

	stats, err := c.service.GetLogStatistics(startTime, endTime)
	if err != nil {
		baseRes.FailWithMessage("获取统计信息失败", ctx)
		return
	}

	// 转换为响应 DTO
	statsResponse := response.OperationLogStatisticsResponse{
		TotalCount:   int64(stats["total_count"].(float64)),
		SuccessCount: int64(stats["success_count"].(float64)),
		FailedCount:  int64(stats["failed_count"].(float64)),
		SuccessRate:  int64(stats["success_rate"].(float64)),
		AvgDuration:  int64(stats["avg_duration"].(float64)),
	}

	baseRes.OkWithDetailed(statsResponse, "获取统计信息成功", ctx)
}

// ClearLogByTimeRange 根据时间范围清空操作日志
//
//	@Summary		根据时间范围清空操作日志
//	@Description	清空指定时间范围的操作日志
//	@Tags			系统管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		request.ClearOperationLogByTimeRangeRequest	true	"清空日志请求参数"
//	@Success		200		{object}	map[string]string	"清空成功"
//	@Failure		400		{object}	map[string]string	"请求参数错误"
//	@Failure		500		{object}	map[string]string	"服务器内部错误"
//	@Router			/api/admin/logs/clear [post]
func (c *OperationLogController) ClearLogByTimeRange(ctx *gin.Context) {
	var req request.ClearOperationLogByTimeRangeRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("无效的请求参数", ctx)
		return
	}

	// 解析时间范围
	startTime, err := time.Parse("2006-01-02", req.StartTime)
	if err != nil {
		baseRes.FailWithMessage("无效的开始时间格式", ctx)
		return
	}

	endTime, err := time.Parse("2006-01-02", req.EndTime)
	if err != nil {
		baseRes.FailWithMessage("无效的结束时间格式", ctx)
		return
	}
	endTime = endTime.Add(24*time.Hour - time.Second)

	// 清空日志
	err = c.service.ClearLogByTimeRange(startTime, endTime)
	if err != nil {
		baseRes.FailWithMessage("清空日志失败", ctx)
		return
	}

	baseRes.OkWithMessage("清空成功", ctx)
}
