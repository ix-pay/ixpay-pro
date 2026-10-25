package wxapi

// 微信公开应用配置接口（wxapp/miniapp 前端空壳启动加载，无需认证）
// 说明：管理平台端的渠道配置接口已迁移至 base 模块（internal/app/base/api/wx_channel_handler.go），
// wx 模块仅保留公开配置入口，业务依赖 wx Domain Service（跨模块经 Domain 层）

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	channelService "github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

	"github.com/gin-gonic/gin"
)

// WxPublicAppController 微信公开应用配置控制器
type WxPublicAppController struct {
	service *channelService.WxChannelService
	log     logger.Logger
}

// NewWxPublicAppController 创建微信公开应用配置控制器
func NewWxPublicAppController(service *channelService.WxChannelService, log logger.Logger) *WxPublicAppController {
	return &WxPublicAppController{service: service, log: log}
}

// GetPublicApp 获取公开应用配置
// @Summary 获取公开应用配置
// @Description 前端空壳（wxapp/miniapp）启动时按 app_key 加载应用配置（脱敏：仅 app_type/app_id/授权入口链接/租户 UI 配置，无敏感信息）
// @Tags 微信渠道配置
// @Produce json
// @Param app_key query string true "应用实例标志"
// @Success 200 {object} baseRes.Response{data=response.WxPublicAppResponse} "应用配置"
// @Router /api/wx/config/app [get]
func (c *WxPublicAppController) GetPublicApp(ctx *gin.Context) {
	appKey := ctx.Query("app_key")
	if appKey == "" {
		baseRes.FailWithMessage("app_key 不能为空", ctx)
		return
	}
	app, err := c.service.GetPublicApp(ctx.Request.Context(), appKey)
	if err != nil {
		c.log.Error("获取公开应用配置失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	resp := response.WxPublicAppResponse{
		AppKey:      app.AppKey,
		AppType:     app.AppType,
		AppID:       app.AppID,
		Description: app.Description,
		Status:      app.Status,
	}
	// 服务号类型附带授权登录入口链接
	if app.AppType == entity.WxAppTypeMP {
		authURL, err := c.service.GetAuthURL(ctx.Request.Context(), app.AppKey, "/payment")
		if err == nil {
			resp.AuthURL = authURL
		}
	}
	// 租户个性化 UI 配置（菜单/图片/主题，前端空壳按此渲染）
	if ui, err := c.service.GetPublicAppUI(ctx.Request.Context(), app); err == nil && ui != nil {
		resp.UI = ui
	}
	baseRes.OkWithDetailed(resp, "获取应用配置成功", ctx)
}
