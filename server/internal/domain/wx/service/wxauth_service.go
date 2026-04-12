package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	baseRepo "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	wxRepo "github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/dto/wx/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
)

// WXAuthService 实现微信认证服务接口
type WXAuthService struct {
	log                  logger.Logger
	auth                 *auth.JWTAuth
	wxUserRepo           wxRepo.WXUserRepository
	wxAuthSessionRepo    wxRepo.WXAuthSessionRepository
	baseConfigRepository baseRepo.ConfigRepository
}

// NewWXAuthService 创建微信认证服务实例
func NewWXAuthService(auth *auth.JWTAuth, log logger.Logger, wxUserRepo wxRepo.WXUserRepository, wxAuthSessionRepo wxRepo.WXAuthSessionRepository, baseConfigRepository baseRepo.ConfigRepository) *WXAuthService {
	return &WXAuthService{
		log:                  log,
		auth:                 auth,
		wxUserRepo:           wxUserRepo,
		wxAuthSessionRepo:    wxAuthSessionRepo,
		baseConfigRepository: baseConfigRepository,
	}
}

// LoginByCode 通过微信授权码登录
func (s *WXAuthService) LoginByCode(code string) (*entity.WXUser, string, string, time.Time, time.Time, error) {
	// 1. 换取 openid 和 session_key
	tokenResult, err := s.getAccessToken(code)
	if err != nil {
		s.log.Error("获取访问令牌失败", "error", err)
		return nil, "", "", time.Time{}, time.Time{}, err
	}

	// 2. 查询或创建微信用户
	wxUser, err := s.wxUserRepo.GetByOpenID(tokenResult.OpenID)
	if err != nil {
		// 用户不存在，创建新用户
		wxUser = &entity.WXUser{
			OpenID:  tokenResult.OpenID,
			UnionID: tokenResult.UnionID,
		}
		if err := s.wxUserRepo.Create(wxUser); err != nil {
			s.log.Error("创建微信用户失败", "error", err)
			return nil, "", "", time.Time{}, time.Time{}, err
		}
	} else {
		// 更新用户信息
		if err := s.wxUserRepo.Update(wxUser); err != nil {
			s.log.Error("更新微信用户失败", "error", err)
			return nil, "", "", time.Time{}, time.Time{}, err
		}
	}

	// 3. 生成会话 token 返回给客户端
	// 微信用户没有 nickname 字段，使用空字符串
	accessToken, refreshToken, accessExpire, refreshExpire, err := s.auth.GenerateToken(wxUser.ID, "", "", "user", "wechat")
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
	// 使用JWTAuth提供的RefreshToken方法直接刷新令牌
	newAccessToken, newRefreshToken, accessExpire, refreshExpire, err := s.auth.RefreshToken(refreshToken)
	if err != nil {
		s.log.Error("刷新令牌失败", "error", err)
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("刷新令牌失败: %w", err)
	}

	// 解析新的访问令牌以获取用户 ID
	claims, err := s.auth.ParseToken(newAccessToken)
	if err != nil {
		s.log.Error("解析新访问令牌失败", "error", err)
		return "", "", time.Time{}, time.Time{}, err
	}

	wxUserID := claims.UserID

	// 更新会话信息
	session, err := s.wxAuthSessionRepo.GetActiveSessionByWXUserID(wxUserID)
	if err != nil {
		s.log.Error("获取活跃会话失败", "error", err)
		return newAccessToken, newRefreshToken, accessExpire, refreshExpire, nil // 即使更新会话失败，也返回新令牌
	}

	// 更新会话
	session.AccessToken = newAccessToken
	session.RefreshToken = newRefreshToken
	session.ExpiresIn = int64(time.Until(accessExpire).Seconds())
	session.ExpiresAt = accessExpire
	if err := s.wxAuthSessionRepo.Update(session); err != nil {
		s.log.Error("更新会话失败", "error", err)
	}

	s.log.Info("令牌刷新成功", "wxUserId", wxUserID)
	return newAccessToken, newRefreshToken, accessExpire, refreshExpire, nil
}

// GetUserInfo 获取微信用户信息
// 根据会话 token 获取用户信息
func (s *WXAuthService) GetUserInfo(accessToken string) (*entity.WXUser, error) {
	// 解析访问令牌
	claims, err := s.auth.ParseToken(accessToken)
	if err != nil {
		s.log.Error("解析访问令牌失败", "error", err)
		return nil, err
	}

	wxUserID := claims.UserID

	// 查询用户信息
	user, err := s.wxUserRepo.GetByID(wxUserID)
	if err != nil {
		s.log.Error("获取用户信息失败", "error", err)
		return nil, err
	}

	// 检查会话是否有效
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
	// 获取访问令牌
	tokenResult, err := s.getAccessToken(code)
	if err != nil {
		s.log.Error("获取访问令牌失败", "error", err)
		return nil, err
	}

	// 获取用户信息
	userInfo, err := s.getUserInfo(tokenResult.AccessToken, tokenResult.OpenID)
	if err != nil {
		s.log.Error("获取用户信息失败", "error", err)
		return nil, err
	}

	s.log.Info("微信用户信息获取成功", "openId", tokenResult.OpenID)
	return userInfo, nil
}

// getAccessToken 获取微信访问令牌
func (s *WXAuthService) getAccessToken(code string) (*response.WechatAuthResult, error) {
	// 从配置读取服务获取微信配置
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

	// 构建请求 URL
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID,
		appSecret,
		code,
	)

	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result response.WechatAuthResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败：%w", err)
	}

	// 检查错误
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信错误：%d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// getUserInfo 获取微信用户详细信息
func (s *WXAuthService) getUserInfo(accessToken, openID string) (*response.WXUserInfoResponse, error) {
	// 构建请求URL
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		accessToken,
		openID,
	)

	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result response.WXUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败：%w", err)
	}

	return &result, nil
}

// BindWXUserToSystemUser 绑定微信用户到系统用户
func (s *WXAuthService) BindWXUserToSystemUser(openID string, userID string) error {
	// 查询微信用户
	wxUser, err := s.wxUserRepo.GetByOpenID(openID)
	if err != nil {
		s.log.Error("获取微信用户失败", "error", err)
		return fmt.Errorf("微信用户不存在：%w", err)
	}

	// 更新微信用户的系统用户 ID
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
	// 查询微信用户
	wxUser, err := s.wxUserRepo.GetByOpenID(openID)
	if err != nil {
		s.log.Error("获取微信用户失败", "error", err)
		return fmt.Errorf("微信用户不存在：%w", err)
	}

	// 解绑系统用户 ID（设置为空字符串）
	wxUser.UserID = ""
	if err := s.wxUserRepo.Update(wxUser); err != nil {
		s.log.Error("更新微信用户失败", "error", err)
		return fmt.Errorf("解绑用户失败：%w", err)
	}

	// 使所有会话失效
	if err := s.wxAuthSessionRepo.InvalidateAllSessionsByWXUserID(wxUser.ID); err != nil {
		s.log.Error("使会话失效失败", "error", err)
	}

	s.log.Info("微信用户解绑成功", "wxUserId", wxUser.ID)
	return nil
}

// UpdateWXUserInfo 更新微信用户信息
func (s *WXAuthService) UpdateWXUserInfo(user *entity.WXUser) error {
	// 验证用户存在
	existingUser, err := s.wxUserRepo.GetByID(user.ID)
	if err != nil {
		s.log.Error("用户不存在", "error", err)
		return fmt.Errorf("用户不存在：%w", err)
	}

	// 更新用户信息
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

	// 保存更新
	if err := s.wxUserRepo.Update(existingUser); err != nil {
		s.log.Error("更新用户信息失败", "error", err)
		return fmt.Errorf("更新用户信息失败：%w", err)
	}

	s.log.Info("微信用户信息更新成功", "wxUserId", user.ID)
	return nil
}

// Logout 用户登出
func (s *WXAuthService) Logout(accessToken string) error {
	// 解析访问令牌
	claims, err := s.auth.ParseToken(accessToken)
	if err != nil {
		s.log.Error("解析访问令牌失败", "error", err)
		return err
	}

	wxUserID := claims.UserID

	// 查询用户会话
	session, err := s.wxAuthSessionRepo.GetActiveSessionByWXUserID(wxUserID)
	if err != nil {
		s.log.Error("获取会话失败", "error", err)
		return err
	}

	// 使会话失效
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
		// 如果随机数生成失败，返回固定值
		return "abcdefghijklmnopqrstuvwxyz0123456789"
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}
