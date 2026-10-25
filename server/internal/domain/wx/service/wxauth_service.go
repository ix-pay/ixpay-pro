package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/dictconst"
	baseRepo "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	wxRepo "github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	wxpayv3 "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayv3"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
)

// 微信 API 地址与 CSRF 缓存
const (
	// snsCode2SessionAPI 小程序登录凭证校验接口（jscode2session）
	snsCode2SessionAPI = "https://api.weixin.qq.com/sns/jscode2session"

	csrfCachePrefix = "wx_oauth_csrf:"
	csrfCacheTTL    = 5 * time.Minute

	// wechatAPIHTTPTimeout 微信平台接口请求超时
	wechatAPIHTTPTimeout = 10 * time.Second
)

// WXAuthService 实现微信认证服务接口
type WXAuthService struct {
	log                  logger.Logger
	auth                 *auth.JWTAuth
	wxUserRepo           wxRepo.WXUserRepository
	wxAuthSessionRepo    wxRepo.WXAuthSessionRepository
	baseConfigRepository baseRepo.ConfigRepository
	tenantRepo           baseRepo.TenantRepository // base_tenant_apps 路由表（app_key → 租户，登录签发 JWT 时写入 TenantID）
	cache                cache.Cache
	wechatPayService     *WechatPayService        // 用于获取微信配置（含 YAML 降级）
	channelRepo          repo.WxChannelRepository // 多渠道配置仓库（多服务号/多小程序）
}

// NewWXAuthService 创建微信认证服务实例
func NewWXAuthService(auth *auth.JWTAuth, log logger.Logger, wxUserRepo wxRepo.WXUserRepository, wxAuthSessionRepo wxRepo.WXAuthSessionRepository, baseConfigRepository baseRepo.ConfigRepository, tenantRepo baseRepo.TenantRepository, cache cache.Cache, wechatPayService *WechatPayService, channelRepo repo.WxChannelRepository) *WXAuthService {
	return &WXAuthService{
		log:                  log,
		auth:                 auth,
		wxUserRepo:           wxUserRepo,
		wxAuthSessionRepo:    wxAuthSessionRepo,
		baseConfigRepository: baseConfigRepository,
		tenantRepo:           tenantRepo,
		cache:                cache,
		wechatPayService:     wechatPayService,
		channelRepo:          channelRepo,
	}
}

// resolveTenantID 按 app_key 查 base_tenant_apps 路由（设计文档 01 §3.1）
// 未绑定路由 = 默认租户（主库），返回 0；路由表在主库，登录前无租户上下文也可查询
func (s *WXAuthService) resolveTenantID(appKey string) int64 {
	if appKey == "" {
		return 0
	}
	route, err := s.tenantRepo.GetAppRoute(appKey)
	if err != nil || route == nil {
		return 0
	}
	return route.TenantID
}

// ctxWithTenant 登录链路专用：先查路由表定位租户，再返回携带租户上下文的 ctx
// （登录前无 JWT，渠道配置读取须先经主库路由表定位租户库；默认租户回主库）
func (s *WXAuthService) ctxWithTenant(appKey string) context.Context {
	if appKey == "" {
		return context.Background()
	}
	if tid := s.resolveTenantID(appKey); tid > 0 {
		return database.WithTenantID(context.Background(), tid)
	}
	return context.Background()
}

// resolveOAuthApp 解析网页授权应用配置：appKey 非空走多渠道应用配置，空则回退全局配置
// ctx 优先：调用方已携带租户上下文（admin 管理端/公开配置接口按租户库查询）时直接使用，
// 无租户上下文（登录链路）时按 app_key 查主库路由表定位租户库
func (s *WXAuthService) resolveOAuthApp(ctx context.Context, appKey string) (appID, appSecret, callbackURL, frontendURL, scope string, err error) {
	if appKey != "" {
		queryCtx := ctx
		if _, hasTenant := database.TenantIDFromContext(ctx); !hasTenant {
			queryCtx = s.ctxWithTenant(appKey)
		}
		app, aerr := s.channelRepo.GetAppByKey(queryCtx, appKey)
		if aerr != nil {
			return "", "", "", "", "", fmt.Errorf("微信应用配置不存在(app_key=%s)", appKey)
		}
		if !app.IsEnabled() {
			return "", "", "", "", "", fmt.Errorf("微信应用 %s 已禁用", appKey)
		}
		if app.AppType != entity.WxAppTypeMP {
			return "", "", "", "", "", fmt.Errorf("应用 %s 非服务号类型，不支持网页授权登录", appKey)
		}
		cfg := s.wechatPayService.GetConfig()
		callback := app.OauthCallbackURL
		if callback == "" {
			callback = cfg.OAuthCallbackURL
		}
		frontend := app.FrontendURL
		if frontend == "" {
			frontend = cfg.FrontendURL
		}
		scope := app.AuthScope
		if scope == "" {
			scope = "snsapi_base"
		}
		return app.AppID, app.AppSecret, callback, frontend, scope, nil
	}

	// 全局配置回退（历史单应用模式）
	appIDConfig, dbErr := s.baseConfigRepository.GetByKey("wechat_app_id")
	if dbErr == nil && appIDConfig.ConfigValue != "" {
		appID = appIDConfig.ConfigValue
	} else {
		appID = s.wechatPayService.GetConfig().AppID
	}
	secretConfig, dbErr := s.baseConfigRepository.GetByKey("wechat_app_secret")
	if dbErr == nil && secretConfig.ConfigValue != "" {
		appSecret = secretConfig.ConfigValue
	} else {
		appSecret = s.wechatPayService.GetConfig().AppSecret
	}
	callbackURL = s.wechatPayService.GetConfig().OAuthCallbackURL
	frontendURL = s.wechatPayService.GetConfig().FrontendURL
	scope = "snsapi_base"
	if appID == "" || appSecret == "" {
		return "", "", "", "", "", fmt.Errorf("微信应用未配置（数据库 wechat_app_id/wechat_app_secret 或应用配置表）")
	}
	return appID, appSecret, callbackURL, frontendURL, scope, nil
}

// GetOAuthURL 生成微信 OAuth 静默授权链接（scope=snsapi_base）
// 授权链接的 redirect_uri 固定指向后端回调接口，state 包含 CSRF token + 前端页面路径
// ctx 用于租户库路由：admin 端透传租户上下文直接在对应租户库查应用配置；
// 公开/登录链路 ctx 无租户上下文时按 app_key 查主库路由表定位
func (s *WXAuthService) GetOAuthURL(ctx context.Context, redirectURI string, appKey string) (string, error) {
	// 解析应用配置（多渠道应用 或 全局配置回退）
	appID, _, callbackURL, _, scope, err := s.resolveOAuthApp(ctx, appKey)
	if err != nil {
		s.log.Error("解析微信应用配置失败", "error", err)
		return "", err
	}

	if callbackURL == "" {
		s.log.Error("微信 OAuth 回调地址未配置，请在数据库或应用配置中配置 oauth_callback_url")
		return "", fmt.Errorf("微信 OAuth 回调地址未配置")
	}

	// 生成 CSRF token 并存入 Redis（5 分钟过期）
	csrfToken := generateNonceStr()
	csrfKey := csrfCachePrefix + csrfToken
	if err := s.cache.Set(csrfKey, redirectURI, csrfCacheTTL); err != nil {
		s.log.Error("CSRF token 缓存失败", "error", err)
		return "", fmt.Errorf("生成授权链接失败：%w", err)
	}

	// state = CSRF token|appKey|前端页面路径（appKey 可为空，兼容旧格式）
	stateValue := fmt.Sprintf("%s|%s|%s", csrfToken, appKey, redirectURI)

	oauthURL := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		appID,
		url.QueryEscape(callbackURL),
		url.QueryEscape(scope),
		url.QueryEscape(stateValue),
	)

	s.log.Info("生成微信 OAuth 授权链接", "callbackURL", callbackURL, "redirectURI", redirectURI, "appKey", appKey, "appID", appID)
	return oauthURL, nil
}

// OAuthCallback 处理微信 OAuth 回调
// 微信授权后回调到后端接口，携带 code 和 state（state = CSRF token|appKey|前端页面路径）
// 处理完 code 后返回前端重定向 URL（使用 fragment 携带 token，避免泄露）
func (s *WXAuthService) OAuthCallback(code, state string) (string, error) {
	// 0. 解析并校验 CSRF token（state 携带 appKey 供多服务号路由）
	csrfToken, appKey, redirectURI, err := s.validateState(state)
	if err != nil {
		s.log.Error("state 校验失败", "error", err)
		return "", fmt.Errorf("授权参数校验失败：%w", err)
	}
	_ = csrfToken // 校验通过即视为有效，已删除缓存防止重放

	// 1. 使用 code 换取 openid，完成登录（按 appKey 路由到对应服务号）
	wxUser, accessToken, refreshToken, _, accessExpire, err := s.LoginByCode(code, appKey)
	if err != nil {
		s.log.Error("微信回调登录失败", "error", err)
		return "", fmt.Errorf("登录失败：%w", err)
	}

	// 2. 获取前端域名：优先从数据库读取，降级使用 YAML 配置
	frontendURL := ""
	frontendConfig, dbErr := s.baseConfigRepository.GetByKey("wechat_frontend_url")
	if dbErr == nil && frontendConfig.ConfigValue != "" {
		frontendURL = frontendConfig.ConfigValue
	} else {
		cfg := s.wechatPayService.getConfig()
		frontendURL = cfg.FrontendURL
		if dbErr != nil {
			s.log.Warn("数据库未配置 wechat_frontend_url，使用 YAML 配置")
		}
	}

	// 3. 构造前端重定向 URL，使用 fragment (#) 传递 token 避免泄露
	// fragment 不会出现在 HTTP 请求头中，不会被 Referer 泄露
	// app_key 一并回传：扫码/公众号菜单直达的新用户 localStorage 无实例标志，
	// 前端落地页凭它写入本地并加载对应租户的 UI 配置（登录态本身由 JWT 携带 appKey）
	redirectTarget := fmt.Sprintf(
		"%s/#%s#token=%s&refresh_token=%s&expire=%d&user=%s&uid=%d&app_key=%s",
		frontendURL, redirectURI, accessToken, refreshToken, accessExpire.Unix(),
		url.QueryEscape(wxUser.Nickname), wxUser.ID, url.QueryEscape(appKey),
	)
	s.log.Info("微信回调处理成功，重定向到前端", "openId", wxUser.OpenID, "target", redirectTarget)
	return redirectTarget, nil
}

// validateState 解析并校验 state 参数中的 CSRF token
// 格式（新）：CSRF token|appKey|前端页面路径；格式（旧）：CSRF token|前端页面路径
// 校验成功后删除缓存，防止重放攻击
func (s *WXAuthService) validateState(state string) (string, string, string, error) {
	parts := strings.SplitN(state, "|", 3)
	var csrfToken, appKey, redirectURI string
	switch len(parts) {
	case 3: // 新格式：csrf|appKey|redirectURI
		csrfToken, appKey, redirectURI = parts[0], parts[1], parts[2]
	case 2: // 旧格式：csrf|redirectURI
		csrfToken, redirectURI = parts[0], parts[1]
	default:
		return "", "", "", fmt.Errorf("无效的 state 参数格式")
	}

	if csrfToken == "" || redirectURI == "" {
		return "", "", "", fmt.Errorf("state 参数不完整")
	}

	// 从 Redis 校验 CSRF token
	csrfKey := csrfCachePrefix + csrfToken
	value, err := s.cache.Get(csrfKey)
	if err != nil {
		s.log.Error("CSRF token 校验失败（可能已过期）", "error", err)
		return "", "", "", fmt.Errorf("授权已过期，请重新授权")
	}

	// 校验 redirectURI 是否与缓存一致（防止 state 被篡改）
	if value != redirectURI {
		s.log.Error("state 中的 redirectURI 与缓存不一致", "cached", value, "state", redirectURI)
		// 清理非法缓存
		_ = s.cache.Delete(csrfKey)
		return "", "", "", fmt.Errorf("授权参数校验失败")
	}

	// 删除缓存，防止重放攻击
	if err := s.cache.Delete(csrfKey); err != nil {
		s.log.Warn("CSRF token 缓存删除失败", "error", err)
	}

	return csrfToken, appKey, redirectURI, nil
}

// LoginByCode 通过微信授权码登录（多服务号版：appKey 定位应用配置，空则用全局配置）
func (s *WXAuthService) LoginByCode(code string, appKey string) (*entity.WXUser, string, string, time.Time, time.Time, error) {
	// 1. 解析应用配置并换取 openid（登录链路无租户上下文，resolveOAuthApp 内部按路由表定位）
	appID, appSecret, _, _, _, err := s.resolveOAuthApp(context.Background(), appKey)
	if err != nil {
		s.log.Error("解析微信应用配置失败", "error", err)
		return nil, "", "", time.Time{}, time.Time{}, err
	}
	tokenResult, err := s.getAccessToken(appID, appSecret, code)
	if err != nil {
		s.log.Error("获取访问令牌失败", "error", err)
		return nil, "", "", time.Time{}, time.Time{}, err
	}

	now := time.Now()

	// 2. 查询或创建微信用户
	wxUser, err := s.wxUserRepo.GetByOpenID(tokenResult.OpenID)
	if err != nil {
		// 用户不存在，创建新用户
		wxUser = &entity.WXUser{
			OpenID:      tokenResult.OpenID,
			UnionID:     tokenResult.UnionID,
			LastLoginAt: now,
		}
		if err := s.wxUserRepo.Create(wxUser); err != nil {
			s.log.Error("创建微信用户失败", "error", err)
			return nil, "", "", time.Time{}, time.Time{}, err
		}
	} else {
		// 更新最近登录时间和 UnionID（可能之前没有）
		wxUser.LastLoginAt = now
		if tokenResult.UnionID != "" && wxUser.UnionID == "" {
			wxUser.UnionID = tokenResult.UnionID
		}
		if err := s.wxUserRepo.Update(wxUser); err != nil {
			s.log.Error("更新微信用户失败", "error", err)
			return nil, "", "", time.Time{}, time.Time{}, err
		}
	}

	// 3. 生成会话 token 返回给客户端（JWT 携带 appKey 与 tenantID，业务请求据此路由到对应渠道与租户库）
	accessToken, refreshToken, accessExpire, refreshExpire, err := s.auth.GenerateTokenV2(fmt.Sprintf("%d", wxUser.ID), "", "", "user", dictconst.UserTypeWechat, wxUser.OpenID, appKey, s.resolveTenantID(appKey))
	if err != nil {
		s.log.Error("生成令牌失败", "error", err)
		return nil, "", "", time.Time{}, time.Time{}, err
	}

	// 4. 创建授权会话记录
	session := &entity.WXAuthSession{
		WXUserID:     wxUser.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Until(accessExpire).Seconds()),
		Scope:        appKey, // 记录登录所用应用标志
		IsActive:     true,
		ExpiresAt:    accessExpire,
	}
	if err := s.wxAuthSessionRepo.Create(session); err != nil {
		s.log.Error("创建授权会话失败", "error", err)
	}

	s.log.Info("微信登录成功", "wxUserId", wxUser.ID, "openId", wxUser.OpenID, "appKey", appKey)
	return wxUser, accessToken, refreshToken, accessExpire, refreshExpire, nil
}

// LoginByMiniappCode 小程序授权登录（多小程序版：appKey 定位应用配置，空则用全局配置）
// 流程：小程序端 wx.login 得到 code → 后端 jscode2session（对应小程序 appid/secret）换 openid → 查/建用户 → 签发 JWT（携带 appKey）
func (s *WXAuthService) LoginByMiniappCode(jsCode string, appKey string) (*entity.WXUser, string, string, time.Time, time.Time, error) {
	// 1. 小程序 code 换 openid（jscode2session，按 appKey 路由到对应小程序）
	openID, unionID, err := s.miniAppCode2Session(jsCode, appKey)
	if err != nil {
		s.log.Error("小程序 code 换取 openid 失败", "error", err)
		return nil, "", "", time.Time{}, time.Time{}, err
	}

	now := time.Now()

	// 2. 查询或创建微信用户（小程序 openid 与服务号 openid 不同，各自建用户，unionid 可关联）
	wxUser, err := s.wxUserRepo.GetByOpenID(openID)
	if err != nil {
		wxUser = &entity.WXUser{
			OpenID:      openID,
			UnionID:     unionID,
			LastLoginAt: now,
		}
		if err := s.wxUserRepo.Create(wxUser); err != nil {
			s.log.Error("创建微信小程序用户失败", "error", err)
			return nil, "", "", time.Time{}, time.Time{}, err
		}
	} else {
		wxUser.LastLoginAt = now
		if unionID != "" && wxUser.UnionID == "" {
			wxUser.UnionID = unionID
		}
		if err := s.wxUserRepo.Update(wxUser); err != nil {
			s.log.Error("更新微信小程序用户失败", "error", err)
			return nil, "", "", time.Time{}, time.Time{}, err
		}
	}

	// 3. 签发本系统登录凭据（JWT 携带 appKey 与 tenantID，业务请求据此路由到对应渠道/商户与租户库）
	accessToken, refreshToken, accessExpire, refreshExpire, err := s.auth.GenerateTokenV2(fmt.Sprintf("%d", wxUser.ID), "", "", "user", dictconst.UserTypeWxMini, wxUser.OpenID, appKey, s.resolveTenantID(appKey))
	if err != nil {
		s.log.Error("生成令牌失败", "error", err)
		return nil, "", "", time.Time{}, time.Time{}, err
	}

	// 4. 创建授权会话记录
	session := &entity.WXAuthSession{
		WXUserID:     wxUser.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Until(accessExpire).Seconds()),
		Scope:        appKey, // 记录登录所用应用标志
		IsActive:     true,
		ExpiresAt:    accessExpire,
	}
	if err := s.wxAuthSessionRepo.Create(session); err != nil {
		s.log.Error("创建授权会话失败", "error", err)
	}

	s.log.Info("小程序登录成功", "wxUserId", wxUser.ID, "openId", wxUser.OpenID, "appKey", appKey)
	return wxUser, accessToken, refreshToken, accessExpire, refreshExpire, nil
}

// getMiniAppCredentials 获取小程序 AppID/AppSecret（多小程序版：appKey 优先走应用配置，空则用全局配置）
func (s *WXAuthService) getMiniAppCredentials(appKey string) (appID, appSecret string, err error) {
	if appKey != "" {
		app, aerr := s.channelRepo.GetAppByKey(s.ctxWithTenant(appKey), appKey)
		if aerr != nil {
			return "", "", fmt.Errorf("微信应用配置不存在(app_key=%s)", appKey)
		}
		if !app.IsEnabled() {
			return "", "", fmt.Errorf("微信应用 %s 已禁用", appKey)
		}
		if app.AppType != entity.WxAppTypeMiniapp {
			return "", "", fmt.Errorf("应用 %s 非小程序类型，不支持小程序登录", appKey)
		}
		return app.AppID, app.AppSecret, nil
	}

	// 全局配置回退（历史单小程序模式）
	appIDConfig, dbErr := s.baseConfigRepository.GetByKey("wechat_miniapp_app_id")
	if dbErr == nil && appIDConfig.ConfigValue != "" {
		appID = appIDConfig.ConfigValue
	} else {
		appID = s.wechatPayService.GetConfig().MiniAppID
	}
	appSecretConfig, dbErr := s.baseConfigRepository.GetByKey("wechat_miniapp_app_secret")
	if dbErr == nil && appSecretConfig.ConfigValue != "" {
		appSecret = appSecretConfig.ConfigValue
	} else {
		appSecret = s.wechatPayService.GetConfig().MiniAppSecret
	}

	if appID == "" || appSecret == "" {
		return "", "", fmt.Errorf("小程序 AppID/AppSecret 未配置（wechat_miniapp_app_id / wechat_miniapp_app_secret）")
	}
	return appID, appSecret, nil
}

// miniAppCode2Session 小程序 js_code 换取 openid 与 unionid（按 appKey 路由）
func (s *WXAuthService) miniAppCode2Session(jsCode string, appKey string) (string, string, error) {
	appID, appSecret, err := s.getMiniAppCredentials(appKey)
	if err != nil {
		return "", "", err
	}

	apiURL := fmt.Sprintf(
		"%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		snsCode2SessionAPI, url.QueryEscape(appID), url.QueryEscape(appSecret), url.QueryEscape(jsCode),
	)

	client := &http.Client{Timeout: wechatAPIHTTPTimeout}
	var lastErr error
	// 系统性错误（-1）重试 3 次，指数退避；业务错误（如 code 无效）不重试
	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(time.Duration(500*(1<<uint(i-1))) * time.Millisecond)
		}

		var result response.WechatAuthResult
		resp, reqErr := client.Get(apiURL)
		if reqErr != nil {
			lastErr = fmt.Errorf("发送请求失败：%w", reqErr)
			continue
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if decodeErr != nil {
			lastErr = fmt.Errorf("解析响应失败：%w", decodeErr)
			continue
		}

		if result.ErrCode == -1 {
			lastErr = fmt.Errorf("微信服务器错误：%d - %s", result.ErrCode, result.ErrMsg)
			continue
		}
		if result.ErrCode != 0 || result.OpenID == "" {
			return "", "", fmt.Errorf("小程序登录失败：errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
		}
		return result.OpenID, result.UnionID, nil
	}
	return "", "", fmt.Errorf("jscode2session 失败：%w", lastErr)
}

// RefreshToken 刷新访问令牌
func (s *WXAuthService) RefreshToken(refreshToken string) (string, string, time.Time, time.Time, error) {
	newAccessToken, newRefreshToken, accessExpire, refreshExpire, err := s.auth.RefreshToken(refreshToken)
	if err != nil {
		s.log.Error("刷新令牌失败", "error", err)
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("刷新令牌失败: %w", err)
	}

	claims, err := s.auth.ParseToken(newAccessToken)
	if err != nil {
		s.log.Error("解析新访问令牌失败", "error", err)
		return "", "", time.Time{}, time.Time{}, err
	}

	wxUserIDInt, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		s.log.Error("解析用户 ID 失败", "error", err)
		return "", "", time.Time{}, time.Time{}, err
	}

	session, err := s.wxAuthSessionRepo.GetActiveSessionByWXUserID(wxUserIDInt)
	if err != nil {
		s.log.Error("获取活跃会话失败", "error", err)
		return newAccessToken, newRefreshToken, accessExpire, refreshExpire, nil
	}

	session.AccessToken = newAccessToken
	session.RefreshToken = newRefreshToken
	session.ExpiresIn = int64(time.Until(accessExpire).Seconds())
	session.ExpiresAt = accessExpire
	if err := s.wxAuthSessionRepo.Update(session); err != nil {
		s.log.Error("更新会话失败", "error", err)
	}

	s.log.Info("令牌刷新成功", "wxUserId", wxUserIDInt)
	return newAccessToken, newRefreshToken, accessExpire, refreshExpire, nil
}

// GetUserInfo 获取微信用户信息
func (s *WXAuthService) GetUserInfo(accessToken string) (*entity.WXUser, error) {
	claims, err := s.auth.ParseToken(accessToken)
	if err != nil {
		s.log.Error("解析访问令牌失败", "error", err)
		return nil, err
	}

	wxUserID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		s.log.Error("解析用户 ID 失败", "error", err)
		return nil, err
	}

	user, err := s.wxUserRepo.GetByID(wxUserID)
	if err != nil {
		s.log.Error("获取用户信息失败", "error", err)
		return nil, err
	}

	session, err := s.wxAuthSessionRepo.GetActiveSessionByWXUserID(wxUserID)
	if err != nil || session.AccessToken != accessToken {
		s.log.Error("会话无效或已过期")
		return nil, fmt.Errorf("会话无效或已过期")
	}

	s.log.Info("用户信息获取成功", "wxUserId", wxUserID)
	return user, nil
}

// GetUserInfoByCode 根据授权码获取用户信息
func (s *WXAuthService) GetUserInfoByCode(code string) (*response.WXUserInfoResponse, error) {
	appID, appSecret, _, _, _, err := s.resolveOAuthApp(context.Background(), "")
	if err != nil {
		s.log.Error("解析微信应用配置失败", "error", err)
		return nil, err
	}
	tokenResult, err := s.getAccessToken(appID, appSecret, code)
	if err != nil {
		s.log.Error("获取访问令牌失败", "error", err)
		return nil, err
	}

	userInfo, err := s.getUserInfo(tokenResult.AccessToken, tokenResult.OpenID)
	if err != nil {
		s.log.Error("获取用户信息失败", "error", err)
		return nil, err
	}

	s.log.Info("微信用户信息获取成功", "openId", tokenResult.OpenID)
	return userInfo, nil
}

// getAccessToken 获取微信访问令牌（code 换用户级凭证，按 appID/appSecret 调用，含重试）
func (s *WXAuthService) getAccessToken(appID string, appSecret string, code string) (*response.WechatAuthResult, error) {
	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("微信应用 AppID/AppSecret 未配置")
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		appID, appSecret, code,
	)

	// 重试 3 次（微信服务器偶发超时或错误，重试可自愈）
	var lastErr error
	httpClient := &http.Client{Timeout: wechatAPIHTTPTimeout}
	for i := 0; i < 3; i++ {
		if i > 0 {
			// 指数退避：500ms -> 1s -> 2s
			backoff := time.Duration(500*(1<<uint(i-1))) * time.Millisecond
			time.Sleep(backoff)
			s.log.Warn("重试获取微信 access_token", "attempt", i+1)
		}

		var result response.WechatAuthResult
		resp, reqErr := httpClient.Get(url)
		if reqErr != nil {
			lastErr = fmt.Errorf("发送请求失败：%w", reqErr)
			continue
		}

		if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("解析响应失败：%w", decodeErr)
			continue
		}
		resp.Body.Close()

		// 微信服务器错误（-1）可重试，参数错误（其他）不重试
		if result.ErrCode == -1 {
			lastErr = fmt.Errorf("微信服务器错误：%d - %s", result.ErrCode, result.ErrMsg)
			continue
		}
		if result.ErrCode != 0 {
			return nil, fmt.Errorf("微信错误：%d - %s", result.ErrCode, result.ErrMsg)
		}

		return &result, nil
	}

	s.log.Error("获取 access_token 失败（已重试 3 次）", "error", lastErr)
	return nil, fmt.Errorf("获取 access_token 失败：%w", lastErr)
}

// getUserInfo 获取微信用户详细信息
func (s *WXAuthService) getUserInfo(accessToken, openID string) (*response.WXUserInfoResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		accessToken,
		openID,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%w", err)
	}
	defer resp.Body.Close()

	var result response.WXUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败：%w", err)
	}

	return &result, nil
}

// BindWXUserToSystemUser 绑定微信用户到系统用户
func (s *WXAuthService) BindWXUserToSystemUser(openID string, userID int64) error {
	wxUser, err := s.wxUserRepo.GetByOpenID(openID)
	if err != nil {
		s.log.Error("获取微信用户失败", "error", err)
		return fmt.Errorf("微信用户不存在：%w", err)
	}

	wxUser.UserID = userID
	if err := s.wxUserRepo.Update(wxUser); err != nil {
		s.log.Error("更新微信用户失败", "error", err)
		return fmt.Errorf("绑定用户失败：%w", err)
	}

	s.log.Info("微信用户绑定成功", "wxUserId", wxUser.ID, "systemUserId", userID)
	return nil
}

// UnbindWXUserFromSystemUser 解绑微信用户与系统用户
func (s *WXAuthService) UnbindWXUserFromSystemUser(openID string) error {
	wxUser, err := s.wxUserRepo.GetByOpenID(openID)
	if err != nil {
		s.log.Error("获取微信用户失败", "error", err)
		return fmt.Errorf("微信用户不存在：%w", err)
	}

	wxUser.UserID = 0
	if err := s.wxUserRepo.Update(wxUser); err != nil {
		s.log.Error("更新微信用户失败", "error", err)
		return fmt.Errorf("解绑用户失败：%w", err)
	}

	if err := s.wxAuthSessionRepo.InvalidateAllSessionsByWXUserID(wxUser.ID); err != nil {
		s.log.Error("使会话失效失败", "error", err)
	}

	s.log.Info("微信用户解绑成功", "wxUserId", wxUser.ID)
	return nil
}

// UpdateWXUserInfo 更新微信用户信息
func (s *WXAuthService) UpdateWXUserInfo(user *entity.WXUser) error {
	existingUser, err := s.wxUserRepo.GetByID(user.ID)
	if err != nil {
		s.log.Error("用户不存在", "error", err)
		return fmt.Errorf("用户不存在：%w", err)
	}

	existingUser.Nickname = user.Nickname
	existingUser.Avatar = user.Avatar
	existingUser.Gender = user.Gender
	existingUser.Country = user.Country
	existingUser.Province = user.Province
	existingUser.City = user.City
	existingUser.Language = user.Language
	existingUser.Subscribe = user.Subscribe
	existingUser.SubscribeTime = user.SubscribeTime
	existingUser.Remark = user.Remark
	existingUser.GroupID = user.GroupID

	if err := s.wxUserRepo.Update(existingUser); err != nil {
		s.log.Error("更新用户信息失败", "error", err)
		return fmt.Errorf("更新用户信息失败：%w", err)
	}

	s.log.Info("微信用户信息更新成功", "wxUserId", user.ID)
	return nil
}

// Logout 用户登出
func (s *WXAuthService) Logout(accessToken string) error {
	claims, err := s.auth.ParseToken(accessToken)
	if err != nil {
		s.log.Error("解析访问令牌失败", "error", err)
		return err
	}

	wxUserID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		s.log.Error("解析用户 ID 失败", "error", err)
		return err
	}

	session, err := s.wxAuthSessionRepo.GetActiveSessionByWXUserID(wxUserID)
	if err != nil {
		s.log.Error("获取会话失败", "error", err)
		return err
	}

	if err := s.wxAuthSessionRepo.InvalidateSession(session.ID); err != nil {
		s.log.Error("使会话失效失败", "error", err)
		return err
	}

	s.log.Info("用户登出成功", "wxUserId", wxUserID)
	return nil
}

// generateNonceStr 生成随机字符串
// 统一委托微信支付官方 WXPayUtility 工具类的 GenerateNonce（此前为项目内自定义 crypto/rand 实现，已收敛替换）
func generateNonceStr() string {
	return wxpayv3.GenerateNonce()
}
