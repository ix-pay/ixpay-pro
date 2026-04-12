package wxapi

import (
	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// AuthController 用户控制器
// @Summary 用户相关 API
// @Description 提供用户注册、登录、信息查询等功能
// @Tags 用户管理
// @Router /api/auth [get]
type AuthController struct {
	service *service.WXAuthService
	log     logger.Logger
}

// NewAuthController 创建用户控制器
func NewAuthController(service *service.WXAuthService, log logger.Logger) *AuthController {
	return &AuthController{
		service: service,
		log:     log,
	}
}

// LoginByCode 微信授权码登录
// @Summary 微信授权码登录
// @Description 使用微信授权码登录获取用户信息和令牌
// @Tags 微信认证
// @Accept json
// @Produce json
// @Param code body request.WechatLoginRequest true "微信授权码登录请求参数"
// @Success 200 {object} baseRes.Response{data=response.WXLoginResponse,msg=string} "登录成功，包含用户信息和令牌"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "认证失败"
// @Router /api/wx/auth/login [post]
func (c *AuthController) LoginByCode(ctx *gin.Context) {
	var req request.WechatLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 调用服务层登录方法，返回用户信息、访问令牌、刷新令牌和错误
	wxUser, accessToken, refreshToken, accessExpire, _, err := c.service.LoginByCode(req.Code)
	if err != nil {
		c.log.Error("微信登录失败", "error", err)
		baseRes.FailWithMessage("微信登录失败", ctx)
		return
	}

	// 转换为响应格式
	userInfoResponse := response.WXUserInfoResponse{
		ID:       wxUser.ID,
		Username: wxUser.Nickname,
		Nickname: wxUser.Nickname,
		Avatar:   wxUser.Avatar,
		Role:     "user", // 默认角色
		Status:   1,      // 默认为启用状态
	}

	// 创建登录响应
	loginResponse := response.WXLoginResponse{
		User:         userInfoResponse,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessExpire: accessExpire.Unix(),
	}

	baseRes.OkWithDetailed(loginResponse, "微信登录成功", ctx)
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用 refresh token 获取新的 access token 和 refresh token
// @Tags 微信认证
// @Accept json
// @Produce json
// @Param refresh body request.WechatRefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} baseRes.Response{data=response.WXLoginResponse,msg=string} "刷新成功，包含新的令牌"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "认证失败"
// @Router /api/wx/auth/refresh-token [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req request.WechatRefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	accessToken, refreshToken, accessExpire, _, err := c.service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.log.Error("刷新令牌失败", "error", err)
		baseRes.FailWithMessage("刷新令牌失败", ctx)
		return
	}

	baseRes.OkWithDetailed(response.WXLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessExpire: accessExpire.Unix(),
	}, "刷新令牌成功", ctx)
}
