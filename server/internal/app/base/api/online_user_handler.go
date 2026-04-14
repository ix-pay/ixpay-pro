package baseapi

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// OnlineUserController 在线用户控制器
// 处理在线用户相关的 HTTP 请求
type OnlineUserController struct {
	service *service.OnlineUserService
	log     logger.Logger
}

// NewOnlineUserController 创建在线用户控制器实例
func NewOnlineUserController(service *service.OnlineUserService, log logger.Logger) *OnlineUserController {
	return &OnlineUserController{
		service: service,
		log:     log,
	}
}

// GetOnlineUserList 获取在线用户列表
//
//	@Summary		获取在线用户列表
//	@Description	获取当前所有在线用户信息
//	@Tags			在线用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int																			true	"页码"
//	@Param			page_size	query		int																			true	"每页数量"
//	@Success		200			{object}	baseRes.Response{data=response.OnlineUserListResponse,msg=string}			"在线用户列表"
//	@Failure		400			{object}	map[string]string															"请求参数错误"
//	@Failure		401			{object}	map[string]string															"未授权"
//	@Failure		500			{object}	map[string]string															"服务器内部错误"
//	@Router			/api/admin/online-user [get]
func (c *OnlineUserController) GetOnlineUserList(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetOnlineUserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	users, err := c.service.GetOnlineUserList()
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 计算分页
	total := int64(len(users))
	page := req.Page
	pageSize := req.PageSize

	// 处理分页
	start := int((page - 1) * pageSize)
	end := start + int(pageSize)
	var pageUsers []*entity.OnlineUser
	if start >= len(users) {
		pageUsers = []*entity.OnlineUser{}
	} else {
		if end > len(users) {
			end = len(users)
		}
		pageUsers = users[start:end]
	}

	// 转换为响应 DTO
	userResponses := make([]response.OnlineUserResponse, len(pageUsers))
	for i, user := range pageUsers {
		userResponses[i] = response.OnlineUserResponse{
			UserID:       strconv.FormatInt(user.UserID, 10),
			Username:     user.Username,
			Nickname:     user.Nickname,
			SessionID:    user.SessionID,
			LoginIP:      user.LoginIP,
			LoginPlace:   user.LoginPlace,
			LoginTime:    user.LoginTime.Format("2006-01-02 15:04:05"),
			LastActiveAt: user.LastActiveAt.Format("2006-01-02 15:04:05"),
			Device:       user.Device,
			Browser:      user.Browser,
			OS:           user.OS,
			UserAgent:    user.UserAgent,
		}
	}

	userListResponse := response.OnlineUserListResponse{
		PageResult: baseRes.PageResult{
			List:     userResponses,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		List: userResponses,
	}

	baseRes.OkWithDetailed(userListResponse, "获取在线用户列表成功", ctx)
}

// GetOnlineUserByID 获取在线用户详情
//
//	@Summary		获取在线用户详情
//	@Description	根据用户 ID 获取在线用户详细信息
//	@Tags			在线用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		string																true	"用户 ID"
//	@Success		200		{object}	baseRes.Response{data=response.OnlineUserResponse,msg=string}		"在线用户详情"
//	@Failure		400		{object}	map[string]string													"请求参数错误"
//	@Failure		401		{object}	map[string]string													"未授权"
//	@Failure		404		{object}	map[string]string													"用户不在线"
//	@Failure		500		{object}	map[string]string													"服务器内部错误"
//	@Router			/api/admin/online-user/:user_id [get]
func (c *OnlineUserController) GetOnlineUserByID(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	userIDStr := ctx.Param("user_id")
	if userIDStr == "" {
		c.log.Error("用户 ID 不能为空")
		baseRes.FailWithMessage("用户 ID 不能为空", ctx)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "user_id", userIDStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	user, err := c.service.GetOnlineUserByID(userID)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为响应 DTO
	userResponse := response.OnlineUserResponse{
		UserID:       strconv.FormatInt(user.UserID, 10),
		Username:     user.Username,
		Nickname:     user.Nickname,
		SessionID:    user.SessionID,
		LoginIP:      user.LoginIP,
		LoginPlace:   user.LoginPlace,
		LoginTime:    user.LoginTime.Format("2006-01-02 15:04:05"),
		LastActiveAt: user.LastActiveAt.Format("2006-01-02 15:04:05"),
		Device:       user.Device,
		Browser:      user.Browser,
		OS:           user.OS,
		UserAgent:    user.UserAgent,
	}

	baseRes.OkWithDetailed(userResponse, "获取在线用户详情成功", ctx)
}

// ForceOffline 强制用户下线
//
//	@Summary		强制用户下线
//	@Description	强制指定用户下线（管理员权限）
//	@Tags			在线用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	body		int64							true	"用户 ID"
//	@Param			reason	body		string						false	"下线原因"
//	@Success		200		{object}	baseRes.Response{msg=string}	"强制下线成功"
//	@Failure		400		{object}	map[string]string			"请求参数错误"
//	@Failure		401		{object}	map[string]string			"未授权"
//	@Failure		404		{object}	map[string]string			"用户不在线"
//	@Failure		500		{object}	map[string]string			"服务器内部错误"
//	@Router			/api/admin/online-user/:user_id [delete]
func (c *OnlineUserController) ForceOffline(ctx *gin.Context) {
	// 检查用户是否已登录
	operatorID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 将字符串 ID 转换为 int64
	userIDStr := ctx.Param("user_id")
	if userIDStr == "" {
		c.log.Error("用户 ID 不能为空")
		baseRes.FailWithMessage("用户 ID 不能为空", ctx)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "user_id", userIDStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	// 将 operatorID 转换为 int64
	var operatorIDInt int64
	switch v := operatorID.(type) {
	case string:
		operatorIDInt, _ = strconv.ParseInt(v, 10, 64)
	case int64:
		operatorIDInt = v
	case int:
		operatorIDInt = int64(v)
	}

	// 强制用户下线
	if err := c.service.ForceOffline(userID, strconv.FormatInt(operatorIDInt, 10)); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("强制用户下线成功", "user_id", userID, "operator_id", operatorIDInt)
	baseRes.OkWithMessage("强制用户下线成功", ctx)
}

// GetOnlineCount 获取在线用户数量
//
//	@Summary		获取在线用户数量
//	@Description	获取当前在线用户总数
//	@Tags			在线用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{data=int64,msg=string}	"在线用户数量"
//	@Failure		401	{object}	map[string]string						"未授权"
//	@Failure		500	{object}	map[string]string						"服务器内部错误"
//	@Router			/api/admin/online-user/count [get]
func (c *OnlineUserController) GetOnlineCount(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	count, err := c.service.GetOnlineCount()
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(count, "获取在线用户数量成功", ctx)
}

// IsOnline 检查用户是否在线
//
//	@Summary		检查用户是否在线
//	@Description	根据用户 ID 检查用户是否在线
//	@Tags			在线用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	query		int64																true	"用户 ID"
//	@Success		200		{object}	baseRes.Response{data=bool,msg=string}							"是否在线"
//	@Failure		400		{object}	map[string]string													"请求参数错误"
//	@Failure		401		{object}	map[string]string													"未授权"
//	@Failure		500		{object}	map[string]string													"服务器内部错误"
//	@Router			/api/admin/online-user/online [get]
func (c *OnlineUserController) IsOnline(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetOnlineUserByIDRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	online, err := c.service.IsOnline(req.UserID)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(online, "检查用户在线状态成功", ctx)
}

// BatchForceOffline 批量强制用户下线
//
//	@Summary		批量强制用户下线
//	@Description	批量强制指定用户下线（管理员权限）
//	@Tags			在线用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.BatchForceOfflineRequest	true	"批量下线请求"
//	@Success		200		{object}	baseRes.Response{msg=string}		"批量强制下线成功"
//	@Failure		400		{object}	map[string]string					"请求参数错误"
//	@Failure		401		{object}	map[string]string					"未授权"
//	@Failure		500		{object}	map[string]string					"服务器内部错误"
//	@Router			/api/admin/online-user/batch [post]
func (c *OnlineUserController) BatchForceOffline(ctx *gin.Context) {
	// 检查用户是否已登录
	operatorID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
		Reason  string   `json:"reason"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	if len(req.UserIDs) == 0 {
		c.log.Error("批量下线用户 ID 不能为空")
		baseRes.FailWithMessage("批量下线用户 ID 不能为空", ctx)
		return
	}

	// 将字符串 ID 数组转换为 int64 数组
	userIDs, err := convertStringSliceToInt64Slice(req.UserIDs)
	if err != nil {
		c.log.Error("用户 ID 格式错误", "error", err)
		baseRes.FailWithMessage("用户 ID 格式错误", ctx)
		return
	}

	// 将 operatorID 转换为 int64
	var operatorIDInt int64
	switch v := operatorID.(type) {
	case string:
		operatorIDInt, _ = strconv.ParseInt(v, 10, 64)
	case int64:
		operatorIDInt = v
	case int:
		operatorIDInt = int64(v)
	}

	// 批量强制用户下线
	if err := c.service.BatchKickoutUsers(userIDs, req.Reason, strconv.FormatInt(operatorIDInt, 10)); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("批量强制用户下线成功", "count", len(userIDs), "operator_id", operatorIDInt)
	baseRes.OkWithMessage("批量强制下线成功", ctx)
}
