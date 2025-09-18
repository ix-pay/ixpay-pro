package baseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/model/request"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/model/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// UserController 用户控制器
// @Summary 用户相关API
// @Description 提供用户注册、登录、信息查询等功能
// @Tags 用户管理
// @Router /v1/user [get]
type UserController struct {
	service model.UserService
	log     logger.Logger
}

// NewUserController 创建用户控制器
func NewUserController(service model.UserService, log logger.Logger) *UserController {
	return &UserController{
		service: service,
		log:     log,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 基础服务
// @Accept json
// @Produce json
// @Param register body request.RegisterRequest true "注册请求参数"
// @Success 201 {object} baseRes.Response{data=response.UserInfoResponse,msg=string} "注册成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Router /v1/auth/register [post]
func (c *UserController) Register(ctx *gin.Context) {
	var req request.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	user, err := c.service.Register(req.Username, req.Password, req.Email)
	if err != nil {
		c.log.Error("用户注册失败", "error", err)
		baseRes.FailWithMessage("用户注册失败", ctx)
		return
	}

	response := response.UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
	}

	baseRes.OkWithDetailed(response, "验证码获取成功", ctx)
}

// GetUserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} baseRes.Response{data=response.UserInfoResponse,msg=string} "用户信息"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /v1/user/info [get]
func (c *UserController) GetUserInfo(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	user, err := c.service.GetUserInfo(userID.(uint))
	if err != nil {
		c.log.Error("获取用户信息失败", "error", err)
		baseRes.FailWithMessage("获取用户信息失败", ctx)
		return
	}

	u := response.UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
	}

	baseRes.OkWithDetailed(u, "获取用户信息成功", ctx)
}
