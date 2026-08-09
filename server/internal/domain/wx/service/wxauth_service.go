package service

import (
	"crypto/rand"
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
	wxRepo "github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/cache"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
)

// CSRF 缓存前缀和过期时间
const (
	csrfCachePrefix = "wx_oauth_csrf:"
	csrfCacheTTL    = 5 * time.Minute
)

// WXAuthService 实现微信认证服务接口
type WXAuthService struct {
	log                  logger.Logger
	auth                 *auth.JWTAuth
	wxUserRepo           wxRepo.WXUserRepository
	wxAuthSessionRepo    wxRepo.WXAuthSessionRepository
	baseConfigRepository baseRepo.ConfigRepository
	cache                cache.Cache
	wechatPayService     *WechatPayService // 用于获取微信配置（含 YAML 降级）
}

// NewWXAuthService 创建微信认证服务实例
func NewWXAuthService(auth *auth.JWTAuth, log logger.Logger, wxUserRepo wxRepo.WXUserRepository, wxAuthSessionRepo wxRepo.WXAuthSessionRepository, baseConfigRepository baseRepo.ConfigRepository, cache cache.Cache, wechatPayService *WechatPayService) *WXAuthService {
	return &WXAuthService{
		log:                  log,
		auth:                 auth,
		wxUserRepo:           wxUserRepo,
		wxAuthSessionRepo:    wxAuthSessionRepo,
		baseConfigRepository: baseConfigRepository,
		cache:                cache,
		wechatPayService:     wechatPayService,
	}
}

// GetOAuthURL 生成微信 OAuth 静默授权链接（scope=snsapi_base）
// 授权链接的 redirect_uri 固定指向后端回调接口，state 包含 CSRF token + 前端页面路径
func (s *WXAuthService) GetOAuthURL(redirectURI string) (string, error) {
	appIDConfig, err := s.baseConfigRepository.GetByKey("wechat_app_id")
	if err != nil {
		s.log.Error("获取微信 AppID 配置失败", "error", err)
		return "", fmt.Errorf("获取微信配置失败：%w", err)
	}
	appID := appIDConfig.ConfigValue

	if appID == "" {
		return "", fmt.Errorf("微信 AppID 未配置")
	}

	// 获取后端回调地址：优先从数据库读取，降级使用 YAML 配置
	callbackURL := ""
	callbackConfig, dbErr := s.baseConfigRepository.GetByKey("wechat_oauth_callback_url")
	if dbErr == nil && callbackConfig.ConfigValue != "" {
		callbackURL = callbackConfig.ConfigValue
	} else {
		cfg := s.wechatPayService.getConfig()
		callbackURL = cfg.OAuthCallbackURL
		if dbErr != nil {
			s.log.Warn("数据库未配置 wechat_oauth_callback_url，使用 YAML 配置")
		}
	}

	if callbackURL == "" {
		s.log.Error("微信 OAuth 回调地址未配置，请在数据库或 config.yaml 中配置 oauth_callback_url")
		return "", fmt.Errorf("微信 OAuth 回调地址未配置")
	}

	// 生成 CSRF token 并存入 Redis（5 分钟过期）
	csrfToken := generateNonceStr()
	csrfKey := csrfCachePrefix + csrfToken
	if err := s.cache.Set(csrfKey, redirectURI, csrfCacheTTL); err != nil {
		s.log.Error("CSRF token 缓存失败", "error", err)
		return "", fmt.Errorf("生成授权链接失败：%w", err)
	}

	// state = CSRF token|前端页面路径
	stateValue := fmt.Sprintf("%s|%s", csrfToken, redirectURI)

	oauthURL := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=%s#wechat_redirect",
		appID,
		url.QueryEscape(callbackURL),
		url.QueryEscape(stateValue),
	)

	s.log.Info("生成微信 OAuth 授权链接", "callbackURL", callbackURL, "redirectURI", redirectURI)
	return oauthURL, nil
}

// OAuthCallback 处理微信 OAuth 回调
// 微信授权后回调到后端接口，携带 code 和 state（state = CSRF token|前端页面路径）
// 处理完 code 后返回前端重定向 URL（使用 fragment 携带 token，避免泄露）
func (s *WXAuthService) OAuthCallback(code, state string) (string, error) {
	// 0. 解析并校验 CSRF token
	csrfToken, redirectURI, err := s.validateState(state)
	if err != nil {
		s.log.Error("state 校验失败", "error", err)
		return "", fmt.Errorf("授权参数校验失败：%w", err)
	}
	_ = csrfToken // 校验通过即视为有效，已删除缓存防止重放

	// 1. 使用 code 换取 openid，完成登录
	wxUser, accessToken, refreshToken, _, accessExpire, err := s.LoginByCode(code)
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
	redirectTarget := fmt.Sprintf(
		"%s/#%s#token=%s&refresh_token=%s&expire=%d",
		frontendURL, redirectURI, accessToken, refreshToken, accessExpire.Unix(),
	)
	s.log.Info("微信回调处理成功，重定向到前端", "openId", wxUser.OpenID, "target", redirectTarget)
	return redirectTarget, nil
}

// validateState 解析并校验 state 参数中的 CSRF token
// 格式：CSRF token|前端页面路径
// 校验成功后删除缓存，防止重放攻击
func (s *WXAuthService) validateState(state string) (string, string, error) {
	parts := strings.SplitN(state, "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("无效的 state 参数格式")
	}
	csrfToken, redirectURI := parts[0], parts[1]

	if csrfToken == "" || redirectURI == "" {
		return "", "", fmt.Errorf("state 参数不完整")
	}

	// 从 Redis 校验 CSRF token
	csrfKey := csrfCachePrefix + csrfToken
	value, err := s.cache.Get(csrfKey)
	if err != nil {
		s.log.Error("CSRF token 校验失败（可能已过期）", "error", err)
		return "", "", fmt.Errorf("授权已过期，请重新授权")
	}

	// 校验 redirectURI 是否与缓存一致（防止 state 被篡改）
	if value != redirectURI {
		s.log.Error("state 中的 redirectURI 与缓存不一致", "cached", value, "state", redirectURI)
		// 清理非法缓存
		_ = s.cache.Delete(csrfKey)
		return "", "", fmt.Errorf("授权参数校验失败")
	}

	// 删除缓存，防止重放攻击
	if err := s.cache.Delete(csrfKey); err != nil {
		s.log.Warn("CSRF token 缓存删除失败", "error", err)
	}

	return csrfToken, redirectURI, nil
}

// LoginByCode 通过微信授权码登录
func (s *WXAuthService) LoginByCode(code string) (*entity.WXUser, string, string, time.Time, time.Time, error) {
	// 1. 换取 openid
	tokenResult, err := s.getAccessToken(code)
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

	// 3. 生成会话 token 返回给客户端
	accessToken, refreshToken, accessExpire, refreshExpire, err := s.auth.GenerateToken(fmt.Sprintf("%d", wxUser.ID), "", "", "user", dictconst.UserTypeWechat, wxUser.OpenID)
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
		Scope:        "user",
		IsActive:     true,
		ExpiresAt:    accessExpire,
	}
	if err := s.wxAuthSessionRepo.Create(session); err != nil {
		s.log.Error("创建授权会话失败", "error", err)
	}

	s.log.Info("微信登录成功", "wxUserId", wxUser.ID, "openId", wxUser.OpenID)
	return wxUser, accessToken, refreshToken, accessExpire, refreshExpire, nil
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
	tokenResult, err := s.getAccessToken(code)
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

// getAccessToken 获取微信访问令牌（含重试）
func (s *WXAuthService) getAccessToken(code string) (*response.WechatAuthResult, error) {
	appIDConfig, err := s.baseConfigRepository.GetByKey("wechat_app_id")
	if err != nil {
		s.log.Error("获取微信 AppID 配置失败", "error", err)
		return nil, fmt.Errorf("获取微信配置失败：%w", err)
	}
	appID := appIDConfig.ConfigValue

	appSecretConfig, err := s.baseConfigRepository.GetByKey("wechat_app_secret")
	if err != nil {
		s.log.Error("获取微信 AppSecret 配置失败", "error", err)
		return nil, fmt.Errorf("获取微信配置失败：%w", err)
	}
	appSecret := appSecretConfig.ConfigValue

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		appID, appSecret, code,
	)

	// 重试 3 次（微信服务器偶发超时或错误，重试可自愈）
	var lastErr error
	for i := 0; i < 3; i++ {
		if i > 0 {
			// 指数退避：500ms -> 1s -> 2s
			backoff := time.Duration(500*(1<<uint(i-1))) * time.Millisecond
			time.Sleep(backoff)
			s.log.Warn("重试获取微信 access_token", "attempt", i+1)
		}

		var result response.WechatAuthResult
		resp, reqErr := http.Get(url)
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
func generateNonceStr() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "abcdefghijklmnopqrstuvwxyz0123456789"
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}