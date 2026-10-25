package baseapi

// 微信多渠道配置管理（base 管理平台端：应用/商户配置、授权入口、对账）
// 说明：控制器属于 base 模块（管理平台路由统一在 base 鉴权），业务依赖 wx 模块 Domain Service（跨模块经 Domain 层）

import (
	"strconv"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	channelService "github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// WxChannelController 微信多渠道配置控制器（管理平台端：应用/商户配置、授权入口、对账）
// @Summary 微信渠道配置 API
// @Description 多服务号/多小程序应用配置、多商户配置、应用-商户关联、对账查询
// @Tags 微信渠道配置
type WxChannelController struct {
	service *channelService.WxChannelService
	log     logger.Logger
}

// NewWxChannelController 创建微信多渠道配置控制器
func NewWxChannelController(service *channelService.WxChannelService, log logger.Logger) *WxChannelController {
	return &WxChannelController{service: service, log: log}
}

// GetAppUI 读取应用 UI 配置
// @Summary 读取租户 UI 配置
// @Description 读取前端实例（wxapp/miniapp）的个性化配置 JSON（菜单/图片/主题）
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Success 200 {object} baseRes.Response{data=map[string]string} "ui 配置原文"
// @Router /api/admin/wx/apps/{id}/ui [get]
func (c *WxChannelController) GetAppUI(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	raw, err := c.service.GetAppUIRaw(ctx.Request.Context(), id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(gin.H{"uiJson": raw}, "查询成功", ctx)
}

// SaveAppUI 保存应用 UI 配置
// @Summary 保存租户 UI 配置
// @Description 保存前端实例个性化配置 JSON（服务端校验可解析；前端启动时动态加载生效，无需发版）
// @Tags 微信渠道配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Param body body request.SaveAppUIRequest true "UI 配置"
// @Success 200 {object} baseRes.Response "保存成功"
// @Router /api/admin/wx/apps/{id}/ui [put]
func (c *WxChannelController) SaveAppUI(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	var req request.SaveAppUIRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	if err := c.service.SaveAppUI(ctx.Request.Context(), id, req.UIJSON); err != nil {
		c.log.Error("保存 UI 配置失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithMessage("保存成功", ctx)
}

// ---------- 应用配置管理 ----------

// ListApps 应用配置列表
// @Summary 微信应用配置列表
// @Description 分页查询微信应用（服务号/小程序）配置，敏感字段脱敏
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码(默认1)"
// @Param pageSize query int false "每页条数(默认10)"
// @Success 200 {object} baseRes.Response{data=baseRes.PageResult} "应用配置列表"
// @Router /api/admin/wx/apps [get]
func (c *WxChannelController) ListApps(ctx *gin.Context) {
	page, pageSize := parsePage(ctx)
	apps, total, err := c.service.ListApps(ctx.Request.Context(), page, pageSize)
	if err != nil {
		c.log.Error("查询应用配置列表失败", "error", err)
		// 管理端透出真实原因便于排障（如租户库缺表：relation "wx_app_configs" does not exist）
		baseRes.FailWithMessage("查询应用配置列表失败: "+err.Error(), ctx)
		return
	}
	list := make([]response.WxAppConfigResponse, 0, len(apps))
	if len(apps) > 0 {
		// 批量查询各应用默认商户（唯一数据源：商户关联表 is_default），避免逐应用 N+1
		appIDs := make([]int64, 0, len(apps))
		for _, app := range apps {
			appIDs = append(appIDs, app.ID)
		}
		defaultMchMap := c.service.MapDefaultMerchants(ctx.Request.Context(), appIDs)
		for _, app := range apps {
			list = append(list, appToResp(app, defaultMchMap[app.ID]))
		}
	}
	baseRes.OkWithDetailed(baseRes.PageResult{List: list, Total: int64(total), Page: page, PageSize: pageSize}, "查询成功", ctx)
}

// GetApp 应用配置详情
// @Summary 微信应用配置详情
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Success 200 {object} baseRes.Response{data=response.WxAppConfigResponse} "应用配置"
// @Router /api/admin/wx/apps/{id} [get]
func (c *WxChannelController) GetApp(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	app, err := c.service.GetApp(ctx.Request.Context(), id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	defaultMchMap := c.service.MapDefaultMerchants(ctx.Request.Context(), []int64{app.ID})
	baseRes.OkWithDetailed(appToResp(app, defaultMchMap[app.ID]), "查询成功", ctx)
}

// CreateApp 创建应用配置
// @Summary 创建微信应用配置
// @Description 创建服务号(mp)或小程序(miniapp)应用配置，app_secret 加密存储
// @Tags 微信渠道配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.WxAppConfigRequest true "应用配置参数"
// @Success 200 {object} baseRes.Response{data=response.WxAppConfigResponse} "创建成功"
// @Router /api/admin/wx/apps [post]
func (c *WxChannelController) CreateApp(ctx *gin.Context) {
	var req request.WxAppConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误: "+err.Error(), ctx)
		return
	}
	app := &entity.WxAppConfig{
		AppKey:           req.AppKey,
		AppType:          req.AppType,
		AppID:            req.AppID,
		AppSecret:        req.AppSecret,
		AuthScope:        req.AuthScope,
		OauthCallbackURL: req.OauthCallbackURL,
		FrontendURL:      req.FrontendURL,
		Description:      req.Description,
	}
	if err := c.service.CreateApp(ctx.Request.Context(), app); err != nil {
		c.log.Error("创建应用配置失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(appToResp(app, nil), "创建成功", ctx)
}

// UpdateApp 更新应用配置
// @Summary 更新微信应用配置
// @Description 更新应用配置；app_secret 传空或 ****** 表示保留原值
// @Tags 微信渠道配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Param body body request.WxAppConfigRequest true "应用配置参数"
// @Success 200 {object} baseRes.Response{data=response.WxAppConfigResponse} "更新成功"
// @Router /api/admin/wx/apps/{id} [put]
func (c *WxChannelController) UpdateApp(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	var req request.WxAppConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误: "+err.Error(), ctx)
		return
	}
	app := &entity.WxAppConfig{
		AppKey:           req.AppKey,
		AppType:          req.AppType,
		AppID:            req.AppID,
		AppSecret:        req.AppSecret,
		AuthScope:        req.AuthScope,
		OauthCallbackURL: req.OauthCallbackURL,
		FrontendURL:      req.FrontendURL,
		Description:      req.Description,
	}
	if err := c.service.UpdateApp(ctx.Request.Context(), id, app); err != nil {
		c.log.Error("更新应用配置失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	updated, _ := c.service.GetApp(ctx.Request.Context(), id)
	defaultMchMap := c.service.MapDefaultMerchants(ctx.Request.Context(), []int64{id})
	baseRes.OkWithDetailed(appToResp(updated, defaultMchMap[id]), "更新成功", ctx)
}

// DeleteApp 删除应用配置
// @Summary 删除微信应用配置
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Success 200 {object} baseRes.Response "删除成功"
// @Router /api/admin/wx/apps/{id} [delete]
func (c *WxChannelController) DeleteApp(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	if err := c.service.DeleteApp(ctx.Request.Context(), id); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithMessage("删除成功", ctx)
}

// GetAppAuthURL 生成应用授权登录入口链接
// @Summary 生成授权登录入口链接
// @Description 为服务号应用生成网页授权登录入口链接（授权后登录并回跳 redirect_uri）
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Param redirect_uri query string false "授权后回跳路径(默认 /payment)"
// @Success 200 {object} baseRes.Response{data=map[string]string} "授权入口链接"
// @Router /api/admin/wx/apps/{id}/auth-url [get]
func (c *WxChannelController) GetAppAuthURL(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	app, err := c.service.GetApp(ctx.Request.Context(), id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	redirectURI := ctx.DefaultQuery("redirect_uri", "/payment")
	authURL, err := c.service.GetAuthURL(ctx.Request.Context(), app.AppKey, redirectURI)
	if err != nil {
		c.log.Error("生成授权入口链接失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(gin.H{"authUrl": authURL, "appKey": app.AppKey}, "生成成功", ctx)
}

// GetAppAuthQrcode 服务号授权入口二维码 PNG
// @Summary 获取授权入口二维码
// @Description 为服务号应用生成授权登录入口链接并渲染为二维码 PNG（微信扫码进入授权登录流程）；仅服务号类型支持
// @Tags 微信渠道配置
// @Produce png
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Success 200 {file} file "二维码PNG"
// @Failure 400 {object} map[string]string "应用类型不支持或链接生成失败"
// @Router /api/admin/wx/apps/{id}/auth-qrcode [get]
func (c *WxChannelController) GetAppAuthQrcode(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	app, err := c.service.GetApp(ctx.Request.Context(), id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	authURL, err := c.service.GetAuthURL(ctx.Request.Context(), app.AppKey, "/payment")
	if err != nil {
		c.log.Error("生成授权入口链接失败", "app_id", id, "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	png, err := qrcode.Encode(authURL, qrcode.Medium, 360)
	if err != nil {
		c.log.Error("生成授权入口二维码失败", "app_id", id, "error", err)
		baseRes.FailWithMessage("生成二维码失败", ctx)
		return
	}
	c.log.Info("生成授权入口二维码成功", "app_id", id, "app_key", app.AppKey)
	ctx.Data(200, "image/png", png)
}

// GetAppMiniappQrcode 小程序入口码 PNG
// @Summary 获取小程序入口码
// @Description 为小程序类型应用生成微信小程序码 PNG（scene 携带 app_key，扫码进入默认支付页）；仅小程序类型支持
// @Tags 微信渠道配置
// @Produce png
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Success 200 {file} file "小程序码PNG"
// @Failure 400 {object} map[string]string "应用类型不支持或配置缺失"
// @Router /api/admin/wx/apps/{id}/miniapp-qrcode [get]
func (c *WxChannelController) GetAppMiniappQrcode(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	png, appKey, err := c.service.GetMiniappEntryQrcode(ctx.Request.Context(), id)
	if err != nil {
		c.log.Error("生成小程序入口码失败", "app_id", id, "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	c.log.Info("生成小程序入口码成功", "app_id", id, "app_key", appKey)
	ctx.Data(200, "image/png", png)
}

// ---------- 商户配置管理 ----------

// ListMerchants 商户配置列表
// @Summary 微信商户配置列表
// @Description 分页查询商户配置，敏感字段脱敏
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码(默认1)"
// @Param pageSize query int false "每页条数(默认10)"
// @Success 200 {object} baseRes.Response{data=baseRes.PageResult} "商户配置列表"
// @Router /api/admin/wx/merchants [get]
func (c *WxChannelController) ListMerchants(ctx *gin.Context) {
	page, pageSize := parsePage(ctx)
	merchants, total, err := c.service.ListMerchants(ctx.Request.Context(), page, pageSize)
	if err != nil {
		c.log.Error("查询商户配置列表失败", "error", err)
		// 管理端透出真实原因便于排障（如租户库缺表：relation "wx_merchant_configs" does not exist）
		baseRes.FailWithMessage("查询商户配置列表失败: "+err.Error(), ctx)
		return
	}
	list := make([]response.WxMerchantConfigResponse, 0, len(merchants))
	for _, m := range merchants {
		list = append(list, merchantToResp(m))
	}
	baseRes.OkWithDetailed(baseRes.PageResult{List: list, Total: int64(total), Page: page, PageSize: pageSize}, "查询成功", ctx)
}

// GetMerchant 商户配置详情
// @Summary 微信商户配置详情
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "商户配置ID"
// @Success 200 {object} baseRes.Response{data=response.WxMerchantConfigResponse} "商户配置"
// @Router /api/admin/wx/merchants/{id} [get]
func (c *WxChannelController) GetMerchant(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	mch, err := c.service.GetMerchant(ctx.Request.Context(), id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(merchantToResp(mch), "查询成功", ctx)
}

// CreateMerchant 创建商户配置
// @Summary 创建微信商户配置
// @Description 创建商户配置，apiv3_key/私钥 PEM 加密存储
// @Tags 微信渠道配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.WxMerchantConfigRequest true "商户配置参数"
// @Success 200 {object} baseRes.Response{data=response.WxMerchantConfigResponse} "创建成功"
// @Router /api/admin/wx/merchants [post]
func (c *WxChannelController) CreateMerchant(ctx *gin.Context) {
	var req request.WxMerchantConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误: "+err.Error(), ctx)
		return
	}
	mch := &entity.WxMerchantConfig{
		MchID:           req.MchID,
		MchName:         req.MchName,
		MchSerialNo:     req.MchSerialNo,
		Apiv3Key:        req.Apiv3Key,
		PrivateKey:      req.PrivateKey,
		PublicKeyID:     req.PublicKeyID,
		PublicKey:       req.PublicKey,
		PayNotifyURL:    req.PayNotifyURL,
		RefundNotifyURL: req.RefundNotifyURL,
	}
	if err := c.service.CreateMerchant(ctx.Request.Context(), mch); err != nil {
		c.log.Error("创建商户配置失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(merchantToResp(mch), "创建成功", ctx)
}

// UpdateMerchant 更新商户配置
// @Summary 更新微信商户配置
// @Description 更新商户配置；apiv3_key/私钥/公钥ID/公钥PEM 传空或 ****** 表示保留原值
// @Tags 微信渠道配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "商户配置ID"
// @Param body body request.WxMerchantConfigRequest true "商户配置参数"
// @Success 200 {object} baseRes.Response{data=response.WxMerchantConfigResponse} "更新成功"
// @Router /api/admin/wx/merchants/{id} [put]
func (c *WxChannelController) UpdateMerchant(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	var req request.WxMerchantConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误: "+err.Error(), ctx)
		return
	}
	mch := &entity.WxMerchantConfig{
		MchID:           req.MchID,
		MchName:         req.MchName,
		MchSerialNo:     req.MchSerialNo,
		Apiv3Key:        req.Apiv3Key,
		PrivateKey:      req.PrivateKey,
		PublicKeyID:     req.PublicKeyID,
		PublicKey:       req.PublicKey,
		PayNotifyURL:    req.PayNotifyURL,
		RefundNotifyURL: req.RefundNotifyURL,
	}
	if err := c.service.UpdateMerchant(ctx.Request.Context(), id, mch); err != nil {
		c.log.Error("更新商户配置失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	updated, _ := c.service.GetMerchant(ctx.Request.Context(), id)
	baseRes.OkWithDetailed(merchantToResp(updated), "更新成功", ctx)
}

// DeleteMerchant 删除商户配置
// @Summary 删除微信商户配置
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "商户配置ID"
// @Success 200 {object} baseRes.Response "删除成功"
// @Router /api/admin/wx/merchants/{id} [delete]
func (c *WxChannelController) DeleteMerchant(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	if err := c.service.DeleteMerchant(ctx.Request.Context(), id); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithMessage("删除成功", ctx)
}

// ---------- 应用-商户关联 ----------

// ListAppMerchants 应用已关联商户
// @Summary 查询应用已关联商户
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Success 200 {object} baseRes.Response{data=baseRes.PageResult} "关联列表"
// @Router /api/admin/wx/apps/{id}/merchants [get]
func (c *WxChannelController) ListAppMerchants(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	links, merchants, err := c.service.ListAppMerchants(ctx.Request.Context(), id)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	mchMap := make(map[int64]*entity.WxMerchantConfig, len(merchants))
	for _, m := range merchants {
		mchMap[m.ID] = m
	}
	list := make([]response.WxAppMerchantResponse, 0, len(links))
	for _, l := range links {
		item := response.WxAppMerchantResponse{
			ID:               l.ID,
			AppConfigID:      l.AppConfigID,
			MerchantConfigID: l.MerchantConfigID,
			IsDefault:        l.IsDefault,
		}
		if m, ok := mchMap[l.MerchantConfigID]; ok {
			item.MchID = m.MchID
			item.MchName = m.MchName
		}
		list = append(list, item)
	}
	baseRes.OkWithDetailed(baseRes.PageResult{List: list, Total: int64(len(list)), Page: 1, PageSize: len(list)}, "查询成功", ctx)
}

// BindAppMerchant 应用关联商户
// @Summary 应用关联商户
// @Description 将商户关联到应用（is_default=true 时设为该应用默认商户）
// @Tags 微信渠道配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Param body body request.WxAppMerchantRequest true "关联参数"
// @Success 200 {object} baseRes.Response "关联成功"
// @Router /api/admin/wx/apps/{id}/merchants [post]
func (c *WxChannelController) BindAppMerchant(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	var req request.WxAppMerchantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}
	if err := c.service.BindAppMerchant(ctx.Request.Context(), id, req.MerchantID, req.IsDefault); err != nil {
		c.log.Error("应用关联商户失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithMessage("关联成功", ctx)
}

// UnbindAppMerchant 解除应用-商户关联
// @Summary 解除应用-商户关联
// @Tags 微信渠道配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "应用配置ID"
// @Param merchant_id path int true "商户配置ID"
// @Success 200 {object} baseRes.Response "解除成功"
// @Router /api/admin/wx/apps/{id}/merchants/{merchant_id} [delete]
func (c *WxChannelController) UnbindAppMerchant(ctx *gin.Context) {
	appID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("ID 格式错误", ctx)
		return
	}
	mchID, err := strconv.ParseInt(ctx.Param("merchant_id"), 10, 64)
	if err != nil {
		baseRes.FailWithMessage("商户ID 格式错误", ctx)
		return
	}
	if err := c.service.UnbindAppMerchantByIDs(ctx.Request.Context(), appID, mchID); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	baseRes.OkWithMessage("解除成功", ctx)
}

// ---------- 对账 ----------

// ReconSummary 对账汇总
// @Summary 对账汇总统计
// @Description 按时间范围统计支付单：总数/成功笔数/成功金额/退款笔数/退款金额/待支付笔数
// @Tags 微信对账
// @Produce json
// @Security BearerAuth
// @Param startDate query string false "开始日期 yyyy-MM-dd(默认近7天)"
// @Param endDate query string false "结束日期 yyyy-MM-dd"
// @Success 200 {object} baseRes.Response{data=response.WxReconSummaryResponse} "对账汇总"
// @Router /api/admin/wx/recon/summary [get]
func (c *WxChannelController) ReconSummary(ctx *gin.Context) {
	start, end, err := channelService.ParseDateRange(ctx.Query("startDate"), ctx.Query("endDate"))
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	summary, err := c.service.GetReconSummary(ctx.Request.Context(), start, end)
	if err != nil {
		c.log.Error("对账汇总查询失败", "error", err)
		baseRes.FailWithMessage("对账汇总查询失败: "+err.Error(), ctx)
		return
	}
	baseRes.OkWithDetailed(response.WxReconSummaryResponse{
		TotalCount:    summary.TotalCount,
		SuccessCount:  summary.SuccessCount,
		SuccessAmount: summary.SuccessAmount,
		RefundCount:   summary.RefundCount,
		RefundAmount:  summary.RefundAmount,
		PendingCount:  summary.PendingCount,
	}, "查询成功", ctx)
}

// ReconPayments 对账支付单列表
// @Summary 对账支付单列表
// @Description 按时间范围与状态分页查询支付单明细（对账核对用）
// @Tags 微信对账
// @Produce json
// @Security BearerAuth
// @Param startDate query string false "开始日期 yyyy-MM-dd(默认近7天)"
// @Param endDate query string false "结束日期 yyyy-MM-dd"
// @Param status query string false "状态过滤"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} baseRes.Response{data=baseRes.PageResult} "支付单列表"
// @Router /api/admin/wx/recon/payments [get]
func (c *WxChannelController) ReconPayments(ctx *gin.Context) {
	start, end, err := channelService.ParseDateRange(ctx.Query("startDate"), ctx.Query("endDate"))
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}
	page, pageSize := parsePage(ctx)
	status := ctx.Query("status")
	filters := map[string]interface{}{
		"created_at >= ?": start,
		"created_at <= ?": end,
	}
	if status != "" {
		filters["status"] = status
	}
	payments, total, err := c.service.ListPayments(ctx.Request.Context(), page, pageSize, filters)
	if err != nil {
		c.log.Error("对账支付单查询失败", "error", err)
		baseRes.FailWithMessage("对账支付单查询失败: "+err.Error(), ctx)
		return
	}
	list := make([]response.PaymentResponse, 0, len(payments))
	for _, p := range payments {
		list = append(list, paymentToResponse(p))
	}
	baseRes.OkWithDetailed(baseRes.PageResult{List: list, Total: int64(total), Page: page, PageSize: pageSize}, "查询成功", ctx)
}

// ---------- 内部转换 ----------

// maskSecret 敏感字段脱敏展示（非空返回 ******）
func maskSecret(v string) string {
	if v == "" {
		return ""
	}
	return "******"
}

// appToResp 应用实体转响应 DTO；defaultMch 为该应用当前默认商户（唯一数据源为商户关联表
// is_default 标记，由 handler 查询后传入），nil 表示未设置默认商户
func appToResp(app *entity.WxAppConfig, defaultMch *entity.WxMerchantConfig) response.WxAppConfigResponse {
	resp := response.WxAppConfigResponse{
		ID:               app.ID,
		AppKey:           app.AppKey,
		AppType:          app.AppType,
		AppID:            app.AppID,
		AppSecret:        maskSecret(app.AppSecret),
		AuthScope:        app.AuthScope,
		OauthCallbackURL: app.OauthCallbackURL,
		FrontendURL:      app.FrontendURL,
		Description:      app.Description,
		Status:           app.Status,
		CreatedAt:        app.CreatedAt.Format(time.DateTime),
	}
	if defaultMch != nil {
		resp.DefaultMerchantID = defaultMch.ID
		resp.DefaultMerchantMchID = defaultMch.MchID
		resp.DefaultMerchantName = defaultMch.MchName
	}
	return resp
}

func merchantToResp(mch *entity.WxMerchantConfig) response.WxMerchantConfigResponse {
	return response.WxMerchantConfigResponse{
		ID:              mch.ID,
		MchID:           mch.MchID,
		MchName:         mch.MchName,
		MchSerialNo:     mch.MchSerialNo,
		Apiv3Key:        maskSecret(mch.Apiv3Key),
		PrivateKey:      maskSecret(mch.PrivateKey),
		PublicKeyID:     maskSecret(mch.PublicKeyID),
		PublicKey:       maskSecret(mch.PublicKey),
		PayNotifyURL:    mch.PayNotifyURL,
		RefundNotifyURL: mch.RefundNotifyURL,
		Status:          mch.Status,
		CreatedAt:       mch.CreatedAt.Format(time.DateTime),
	}
}

// paymentToResponse 支付实体转响应 DTO（对账明细用；与 wx 模块 handler 同逻辑的本地副本）
func paymentToResponse(p *entity.Payment) response.PaymentResponse {
	resp := response.PaymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		UserID:        p.UserID,
		Amount:        float64(p.Amount) / 100.0,
		Currency:      p.Currency,
		PaymentMethod: p.Method,
		Status:        string(p.Status),
		TransactionID: p.TransactionID,
		Description:   p.Description,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if !p.CreatedAt.IsZero() {
		resp.CreatedAt = p.CreatedAt.Format("2006-01-02 15:04:05")
	}
	return resp
}

// parsePage 解析分页参数（base 管理端私有副本）
func parsePage(ctx *gin.Context) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}
