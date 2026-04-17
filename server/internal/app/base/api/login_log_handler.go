package baseapi

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/domain/converter"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// LoginLogController 登录日志控制器
// 处理登录日志相关的 HTTP 请求
type LoginLogController struct {
	service *service.LoginLogService
	log     logger.Logger
}

// NewLoginLogController 创建登录日志控制器实例
func NewLoginLogController(service *service.LoginLogService, log logger.Logger) *LoginLogController {
	return &LoginLogController{
		service: service,
		log:     log,
	}
}

// GetLoginLogList 获取登录日志列表
//
//	@Summary		获取登录日志列表
//	@Description	获取登录日志列表（支持分页和筛选）
//	@Tags			登录日志管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int									true	"页码"
//	@Param			page_size	query		int									true	"每页数量"
//	@Param			user_id		query		int64								false	"用户 ID"
//	@Param			userName	query		string								false	"用户名"
//	@Param			login_ip	query		string								false	"登录 IP"
//	@Param			result		query		int									false	"登录结果：0-失败，1-成功"
//	@Param			start_date	query		string								false	"开始日期（YYYY-MM-DD）"
//	@Param			end_date	query		string								false	"结束日期（YYYY-MM-DD）"
//	@Success		200			{object}	baseRes.Response{data=baseRes.PageResult{list=[]response.LoginLogListDTO},msg=string}	"登录日志列表"
//	@Failure		400			{object}	map[string]string					"请求参数错误"
//	@Failure		401			{object}	map[string]string					"未授权"
//	@Failure		500			{object}	map[string]string					"服务器内部错误"
//	@Router			/api/admin/login-log [get]
func (c *LoginLogController) GetLoginLogList(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetLoginLogListRequest
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
	if req.LoginIP != "" {
		filters["login_ip"] = req.LoginIP
	}
	if req.Result != nil {
		filters["result"] = *req.Result
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

	logs, total, err := c.service.GetLoginLogList(req.Page, req.PageSize, filters)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 使用转换器将 Entity 转换为 DTO
	dtoList := converter.ConvertSliceWithFunc(logs, converter.LoginLogToListDTO)

	pageResult := baseRes.PageResult{
		List:     dtoList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	baseRes.OkWithDetailed(pageResult, "获取登录日志列表成功", ctx)
}

// GetStatistics 获取登录统计
//
//	@Summary		获取登录统计
//	@Description	获取登录统计信息（按日期、用户、状态等）
//	@Tags			登录日志管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			start_date	query		string								true	"开始日期（YYYY-MM-DD）"
//	@Param			end_date	query		string								true	"结束日期（YYYY-MM-DD）"
//	@Success		200			{object}	baseRes.Response{data=entity.LoginStatistics,msg=string}	"统计信息"
//	@Failure		400			{object}	map[string]string					"请求参数错误"
//	@Failure		401			{object}	map[string]string					"未授权"
//	@Failure		500			{object}	map[string]string					"服务器内部错误"
//	@Router			/api/admin/login-log/statistics [get]
func (c *LoginLogController) GetStatistics(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetLoginStatisticsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 解析日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.log.Error("开始日期格式错误", "error", err)
		baseRes.FailWithMessage("开始日期格式错误，应为 YYYY-MM-DD", ctx)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.log.Error("结束日期格式错误", "error", err)
		baseRes.FailWithMessage("结束日期格式错误，应为 YYYY-MM-DD", ctx)
		return
	}

	// 结束日期加一天，包含结束日期当天
	endDate = endDate.Add(24 * time.Hour)

	stats, err := c.service.GetStatistics(startDate, endDate)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// TODO: 需要创建 LoginStatistics 转换函数
	// statsDTO := converter.ConvertLoginStatistics(stats)
	baseRes.OkWithDetailed(stats, "获取登录统计成功", ctx)
}

// GetAbnormalLogins 获取异常登录查询
//
//	@Summary		获取异常登录查询
//	@Description	获取异常登录记录（同一 IP 多次失败、异地登录等）
//	@Tags			登录日志管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int										true	"页码"
//	@Param			page_size	query		int										true	"每页数量"
//	@Success		200			{object}	baseRes.Response{data=baseRes.PageResult{list=[]response.AbnormalLoginInfoDTO},msg=string}	"异常登录列表"
//	@Failure		400			{object}	map[string]string					"请求参数错误"
//	@Failure		401			{object}	map[string]string					"未授权"
//	@Failure		500			{object}	map[string]string					"服务器内部错误"
//	@Router			/api/admin/login-log/abnormal [get]
func (c *LoginLogController) GetAbnormalLogins(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetAbnormalLoginsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	abnormalLogins, total, err := c.service.GetAbnormalLogins(req.Page, req.PageSize)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 使用转换器将 Entity 转换为 DTO
	dtoList := converter.ConvertSliceWithFunc(abnormalLogins, converter.AbnormalLoginInfoToDTO)

	pageResult := baseRes.PageResult{
		List:     dtoList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	baseRes.OkWithDetailed(pageResult, "获取异常登录记录成功", ctx)
}

// RecordLogin 记录登录日志（内部调用）
// 该接口仅供内部服务调用，不对外暴露
//
//	@Summary		记录登录日志（内部）
//	@Description	记录用户登录日志（仅供内部调用）
//	@Tags			登录日志管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.RecordLoginRequest	true	"登录日志信息"
//	@Success		200		{object}	baseRes.Response{msg=string}	"记录成功"
//	@Failure		400		{object}	map[string]string			"请求参数错误"
//	@Failure		401		{object}	map[string]string			"未授权"
//	@Failure		500		{object}	map[string]string			"服务器内部错误"
//	@Router			/api/admin/login-log [post]
func (c *LoginLogController) RecordLogin(ctx *gin.Context) {
	// 检查授权（内部调用需要特殊权限）
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 检查是否为内部服务调用（这里简化处理，实际应该检查特定的内部服务令牌）
	isInternal := ctx.GetHeader("X-Internal-Request") == "true"
	if !isInternal {
		c.log.Error("非内部服务调用，拒绝访问")
		baseRes.FailWithMessage("无权访问", ctx)
		return
	}

	var req struct {
		UserID    string `json:"user_id,string" binding:"required"`
		Username  string `json:"userName" binding:"required"`
		IP        string `json:"ip" binding:"required"`
		Place     string `json:"place"`
		Device    string `json:"device"`
		Browser   string `json:"browser"`
		OS        string `json:"os"`
		UserAgent string `json:"user_agent"`
		Success   bool   `json:"success"`
		ErrorMsg  string `json:"error_msg"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 将 UserID 从 string 转换为 int64
	userIDInt, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		c.log.Error("用户 ID 格式错误", "user_id", req.UserID, "error", err)
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}

	if err := c.service.RecordLogin(
		userIDInt,
		req.Username,
		req.IP,
		req.Place,
		req.Device,
		req.Browser,
		req.OS,
		req.UserAgent,
		req.Success,
		req.ErrorMsg,
	); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("记录登录日志成功", "user_id", userIDInt, "userName", req.Username, "operator_id", userID)
	baseRes.OkWithMessage("记录登录日志成功", ctx)
}

// GetLoginLogByID 获取登录日志详情
//
//	@Summary		获取登录日志详情
//	@Description	根据 ID 获取登录日志详细信息
//	@Tags			登录日志管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string									true	"登录日志 ID"
//	@Success		200	{object}	baseRes.Response{data=response.LoginLogDetailDTO,msg=string}	"登录日志详情"
//	@Failure		400	{object}	map[string]string						"请求参数错误"
//	@Failure		401	{object}	map[string]string						"未授权"
//	@Failure		500	{object}	map[string]string						"服务器内部错误"
//	@Router			/api/admin/login-log/:id [get]
func (c *LoginLogController) GetLoginLogByID(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	logIDStr := ctx.Param("id")
	if logIDStr == "" {
		c.log.Error("登录日志 ID 不能为空")
		baseRes.FailWithMessage("登录日志 ID 不能为空", ctx)
		return
	}

	logID, err := strconv.ParseInt(logIDStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "id", logIDStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	log, err := c.service.GetLoginLogByID(logID)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 使用转换器将 Entity 转换为详情 DTO
	detailDTO := converter.ConvertWithFunc(log, converter.LoginLogToDetailDTO)

	baseRes.OkWithDetailed(detailDTO, "获取登录日志详情成功", ctx)
}
