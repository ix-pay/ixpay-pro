package controller

import (
	"github.com/ix-pay/ixpay-pro/internal/app/common/baseRes"
	"github.com/ix-pay/ixpay-pro/internal/app/request"
	"github.com/ix-pay/ixpay-pro/internal/app/response"
	"github.com/ix-pay/ixpay-pro/internal/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
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

// Captcha 生成验证码
// @Summary 生成验证码
// @Description 生成验证码
// @Tags 基础服务
// @Accept json
// @Produce json
// @Success   200  {object}  baseRes.Response{data=response.CaptchaResponse,msg=string}  "生成验证码,返回包括随机数id,base64,验证码长度,是否开启验证码"
// @Router /v1/auth/captcha [post]
func (c *UserController) Captcha(ctx *gin.Context) {

	id, b64s, lent, oc, err := c.service.Captcha()
	if err != nil {
		c.log.Error("Us验证码获取失败", "error", err)
		baseRes.FailWithMessage("验证码获取失败", ctx)
		return
	}

	baseRes.OkWithDetailed(response.CaptchaResponse{
		CaptchaId:     id,
		PicPath:       b64s,
		CaptchaLength: lent,
		OpenCaptcha:   oc,
	}, "验证码获取成功", ctx)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户通过账号密码登录系统
// @Tags 基础服务
// @Accept json
// @Produce json
// @Param login body request.LoginRequest true "登录请求参数"
// @Success 200 {object} baseRes.Response{data=response.LoginResponse,msg=string} "登录成功，包含用户信息和令牌"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Router /v1/auth/login [post]
func (c *UserController) Login(ctx *gin.Context) {
	var req request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	user, accessToken, refreshToken, err := c.service.Login(req.Username, req.Password)
	if err != nil {
		c.log.Error("用户登录失败", "error", err)
		baseRes.FailWithMessage("用户登录失败", ctx)
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

	baseRes.OkWithDetailed(response.LoginResponse{
		User:         u,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, "登录成功", ctx)
}

// WechatLogin 微信登录
// @Summary 微信登录
// @Description 用户通过微信code进行登录
// @Tags 基础服务
// @Accept json
// @Produce json
// @Param wechat body request.WechatLoginRequest true "微信登录请求参数"
// @Success 200 {object} baseRes.Response{data=response.LoginResponse,msg=string} "登录成功，包含用户信息和令牌"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Router /v1/auth/wechat-login [post]
func (c *UserController) WechatLogin(ctx *gin.Context) {
	var req request.WechatLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	user, accessToken, refreshToken, err := c.service.WechatLogin(req.Code)
	if err != nil {
		c.log.Error("微信登录失败", "error", err)
		baseRes.FailWithMessage("微信登录失败", ctx)
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

	baseRes.OkWithDetailed(response.LoginResponse{
		User:         u,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, "登录成功", ctx)
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用refresh token获取新的access token和refresh token
// @Tags 基础服务
// @Accept json
// @Produce json
// @Param refresh body request.RefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} baseRes.Response{data=response.LoginResponse,msg=string} "刷新成功，包含新的令牌"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "认证失败"
// @Router /v1/auth/refresh-token [post]
func (c *UserController) RefreshToken(ctx *gin.Context) {
	var req request.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	accessToken, refreshToken, err := c.service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.log.Error("刷新令牌失败", "error", err)
		baseRes.FailWithMessage("刷新令牌失败", ctx)
		return
	}

	baseRes.OkWithDetailed(response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, "刷新令牌成功", ctx)
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
