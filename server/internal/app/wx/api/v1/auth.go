package wxapi

import (
	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/model/request"
	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/model/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// AuthController 用户控制器
// @Summary 用户相关API
// @Description 提供用户注册、登录、信息查询等功能
// @Tags 用户管理
// @Router /v1/auth [get]
type AuthController struct {
	service model.WXAuthService
	log     logger.Logger
}

// NewAuthController 创建用户控制器
func NewAuthController(service model.WXAuthService, log logger.Logger) *AuthController {
	return &AuthController{
		service: service,
		log:     log,
	}
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用refresh token获取新的access token和refresh token
// @Tags 基础服务
// @Accept json
// @Produce json
// @Param refresh body request.RefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} baseRes.Response{data=response.WXLoginResponse,msg=string} "刷新成功，包含新的令牌"
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

	baseRes.OkWithDetailed(response.WXLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, "刷新令牌成功", ctx)
}
