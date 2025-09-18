package baseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/model/request"
	"github.com/ix-pay/ixpay-pro/internal/app/base/domain/model/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// AuthController 用户控制器
// @Summary 用户相关API
// @Description 提供用户注册、登录、信息查询等功能
// @Tags 用户管理
// @Router /v1/auth [get]
type AuthController struct {
	service model.UserService
	log     logger.Logger
}

// NewAuthController 创建用户控制器
func NewAuthController(service model.UserService, log logger.Logger) *AuthController {
	return &AuthController{
		service: service,
		log:     log,
	}
}

// Captcha 生成验证码
// @Summary 生成验证码
// @Description 生成验证码
// @Tags 基础服务
// @Accept json
// @Produce json
// @Success   200  {object}  baseRes.Response{data=response.CaptchaResponse,msg=string}  "生成验证码,返回包括随机数id,base64,验证码长度,是否开启验证码"
// @Router /v1/auth/captcha [post]
func (c *AuthController) Captcha(ctx *gin.Context) {

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
func (c *AuthController) Login(ctx *gin.Context) {
	var req request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	auth, accessToken, refreshToken, err := c.service.Login(req.Username, req.Password)
	if err != nil {
		c.log.Error("用户登录失败", "error", err)
		baseRes.FailWithMessage("用户登录失败", ctx)
		return
	}

	u := response.UserInfoResponse{
		ID:       auth.ID,
		Username: auth.Username,
		Nickname: auth.Nickname,
		Email:    auth.Email,
		Phone:    auth.Phone,
		Avatar:   auth.Avatar,
		Role:     auth.Role,
		Status:   auth.Status,
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
func (c *AuthController) RefreshToken(ctx *gin.Context) {
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
