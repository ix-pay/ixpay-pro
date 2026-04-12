package baseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// AuthController 认证控制器
// 处理认证相关的 HTTP 请求
// 提供用户登录、验证码、刷新令牌等功能的 API 接口
type AuthController struct {
	service *service.UserService // 用户服务接口
	jwtAuth *auth.JWTAuth        // JWT 认证
	log     logger.Logger        // 日志记录器
}

// NewAuthController 创建认证控制器实例
// 参数:
// - service: 用户服务接口实现
// - jwtAuth: JWT 认证
// - log: 日志记录器
// - redisClient: Redis 客户端
// 返回:
// - *AuthController: 认证控制器实例
func NewAuthController(service *service.UserService, jwtAuth *auth.JWTAuth, log logger.Logger) *AuthController {
	return &AuthController{
		service: service,
		jwtAuth: jwtAuth,
		log:     log,
	}
}

// Login 用户登录
//
//	@Summary		用户登录
//	@Description	用户登录获取访问令牌
//	@Tags			认证服务
//	@Accept			json
//	@Produce		json
//	@Param			login	body		request.LoginRequest								true	"登录请求参数"
//	@Success		200		{object}	baseRes.Response{data=response.LoginResponse,msg=string}	"登录成功"
//	@Failure		400		{object}	map[string]string									"请求参数错误"
//	@Failure		401		{object}	map[string]string									"用户名或密码错误"
//	@Router			/api/admin//auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 获取客户端 IP
	ip := ctx.ClientIP()
	// 获取 User-Agent
	userAgent := ctx.Request.UserAgent()

	// 调用服务层进行登录
	user, accessToken, refreshToken, _, _, err := c.service.Login(req.Username, req.Password, req.CaptchaId, req.Captcha, ip, userAgent)
	if err != nil {
		c.log.Error("登录失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 通过用户角色 ID 列表获取完整的角色信息
	roleInfos := make([]*response.RoleInfo, 0, len(user.RoleIds))
	for _, roleID := range user.RoleIds {
		role, err := c.service.GetRoleByID(roleID)
		if err != nil {
			c.log.Warn("获取角色信息失败", "roleID", roleID, "error", err)
			continue
		}
		roleInfo := &response.RoleInfo{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
			Type:        role.Type,
			ParentId:    role.ParentID,
			Status:      role.Status,
			IsSystem:    role.IsSystem,
			Sort:        role.Sort,
		}
		roleInfos = append(roleInfos, roleInfo)
	}

	userInfo := response.UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Status:   user.Status,
		Roles:    roleInfos,
	}

	// 构建响应数据
	loginResponse := response.LoginResponse{
		User:         userInfo,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	baseRes.OkWithDetailed(loginResponse, "登录成功", ctx)
}

// Captcha 获取验证码
//
//	@Summary		获取验证码
//	@Description	获取登录验证码
//	@Tags			认证服务
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	baseRes.Response{data=response.CaptchaResponse,msg=string}	"验证码信息"
//	@Failure		500	{object}	map[string]string									"服务器内部错误"
//	@Router			/api/admin//auth/captcha [post]
func (c *AuthController) Captcha(ctx *gin.Context) {
	captchaId, captchaImage, captchaLen, openCaptcha, err := c.service.Captcha()
	if err != nil {
		c.log.Error("获取验证码失败", "error", err)
		baseRes.FailWithMessage("获取验证码失败", ctx)
		return
	}

	responseData := response.CaptchaResponse{
		CaptchaId:     captchaId,
		PicPath:       captchaImage,
		CaptchaLength: captchaLen,
		OpenCaptcha:   openCaptcha,
	}

	baseRes.OkWithDetailed(responseData, "获取验证码成功", ctx)
}

// RefreshToken 刷新令牌
//
//	@Summary		刷新令牌
//	@Description	使用刷新令牌获取新的访问令牌
//	@Tags			认证服务
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{object}	baseRes.Response{data=response.LoginResponse,msg=string}	"刷新成功"
//	@Failure		400				{object}	map[string]string							"请求参数错误"
//	@Failure		401				{object}	map[string]string							"刷新令牌无效"
//	@Router			/api/admin//auth/refresh-token [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req request.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 调用服务层刷新令牌
	accessToken, refreshToken, _, _, err := c.service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.log.Error("刷新令牌失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 构建响应数据
	loginResponse := response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	baseRes.OkWithDetailed(loginResponse, "刷新成功", ctx)
}

// Logout 用户登出
//
//	@Summary		用户登出
//	@Description	用户登出系统
//	@Tags			认证服务
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{msg=string}	"登出成功"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin//auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// 从上下文中获取用户 ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	if err := c.service.Logout(userID.(string)); err != nil {
		c.log.Error("登出失败", "error", err)
		baseRes.FailWithMessage("登出失败", ctx)
		return
	}

	baseRes.OkWithMessage("登出成功", ctx)
}
