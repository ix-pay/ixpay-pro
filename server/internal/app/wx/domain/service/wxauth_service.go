package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/app/wx/domain/model/response"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/auth"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
)

// WXAuthService 实现微信认证服务接口
type WXAuthService struct {
	config            *config.Config
	log               logger.Logger
	jwtAuth           *auth.JWTAuth
	wxUserRepo        model.WXUserRepository
	wxAuthSessionRepo model.WXAuthSessionRepository
}

// NewWXAuthService 创建微信认证服务实例
func NewWXAuthService(config *config.Config, jwtAuth *auth.JWTAuth, log logger.Logger,
	wxUserRepo model.WXUserRepository, wxAuthSessionRepo model.WXAuthSessionRepository) model.WXAuthService {
	return &WXAuthService{
		config:            config,
		log:               log,
		jwtAuth:           jwtAuth,
		wxUserRepo:        wxUserRepo,
		wxAuthSessionRepo: wxAuthSessionRepo,
	}
}

// LoginByCode 通过微信授权码登录
func (s *WXAuthService) LoginByCode(code string) (*model.WXUser, string, string, error) {
	// 1. 换取openid和session_key
	tokenResult, err := s.getAccessToken(code)
	if err != nil {
		s.log.Error("Failed to get access token", "error", err)
		return nil, "", "", err
	}

	// 2. 查询或创建微信用户
	wxUser, err := s.wxUserRepo.GetByOpenID(tokenResult.OpenID)
	if err != nil {
		// 用户不存在，创建新用户
		wxUser = &model.WXUser{
			OpenID:    tokenResult.OpenID,
			UnionID:   tokenResult.UnionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.wxUserRepo.Create(wxUser); err != nil {
			s.log.Error("Failed to create wx user", "error", err)
			return nil, "", "", err
		}
	} else {
		// 更新用户信息
		wxUser.UpdatedAt = time.Now()
		if err := s.wxUserRepo.Update(wxUser); err != nil {
			s.log.Error("Failed to update wx user", "error", err)
		}
	}

	// 3. 生成会话token返回给客户端
	accessToken, refreshToken, err := s.jwtAuth.GenerateToken(wxUser.ID, "", "user", "wechat")
	if err != nil {
		s.log.Error("Failed to generate token", "error", err)
		return nil, "", "", err
	}

	// 4. 创建授权会话记录
	// 解析访问令牌过期时间
	accessTokenExpire, expireErr := time.ParseDuration(s.config.JWT.AccessTokenExpire)
	if expireErr != nil {
		s.log.Error("Failed to parse access token expire duration", "error", expireErr)
		accessTokenExpire = time.Hour * 24 // 默认24小时
	}
	session := &model.WXAuthSession{
		WXUserID:     wxUser.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenExpire.Seconds()),
		Scope:        "user",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(accessTokenExpire),
	}
	if err := s.wxAuthSessionRepo.Create(session); err != nil {
		s.log.Error("Failed to create auth session", "error", err)
	}

	s.log.Info("Wechat login successful", "wxUserId", wxUser.ID, "openId", wxUser.OpenID)
	return wxUser, accessToken, refreshToken, nil
}

// RefreshToken 刷新访问令牌
func (s *WXAuthService) RefreshToken(refreshToken string) (string, string, error) {
	// 使用JWTAuth提供的RefreshToken方法直接刷新令牌
	newAccessToken, newRefreshToken, err := s.jwtAuth.RefreshToken(refreshToken)
	if err != nil {
		s.log.Error("Failed to refresh token", "error", err)
		return "", "", err
	}

	// 解析新的访问令牌以获取用户ID
	claims, err := s.jwtAuth.ParseToken(newAccessToken)
	if err != nil {
		s.log.Error("Failed to parse new access token", "error", err)
		return "", "", err
	}

	// 获取用户ID
	wxUserID := uint(claims.UserID)

	// 更新会话信息
	session, err := s.wxAuthSessionRepo.GetActiveSessionByWXUserID(wxUserID)
	if err != nil {
		s.log.Error("Failed to get active session", "error", err)
		return newAccessToken, newRefreshToken, nil // 即使更新会话失败，也返回新令牌
	}

	// 更新会话
	session.AccessToken = newAccessToken
	session.RefreshToken = newRefreshToken
	session.UpdatedAt = time.Now()
	// 解析访问令牌过期时间
	accessTokenExpire, expireErr := time.ParseDuration(s.config.JWT.AccessTokenExpire)
	if expireErr != nil {
		s.log.Error("Failed to parse access token expire duration", "error", expireErr)
		accessTokenExpire = time.Hour * 24 // 默认24小时
	}
	session.ExpiresAt = time.Now().Add(accessTokenExpire)
	if err := s.wxAuthSessionRepo.Update(session); err != nil {
		s.log.Error("Failed to update session", "error", err)
	}

	s.log.Info("Token refreshed successfully", "wxUserId", wxUserID)
	return newAccessToken, newRefreshToken, nil
}

// GetUserInfo 获取微信用户信息
func (s *WXAuthService) GetUserInfo(accessToken string) (*model.WXUser, error) {
	// 解析访问令牌
	claims, err := s.jwtAuth.ParseToken(accessToken)
	if err != nil {
		s.log.Error("Failed to parse access token", "error", err)
		return nil, err
	}

	// 获取用户ID
	wxUserID := uint(claims.UserID)

	// 查询用户信息
	user, err := s.wxUserRepo.GetByID(wxUserID)
	if err != nil {
		s.log.Error("Failed to get user info", "error", err)
		return nil, err
	}

	// 检查会话是否有效
	session, err := s.wxAuthSessionRepo.GetActiveSessionByWXUserID(wxUserID)
	if err != nil || session.AccessToken != accessToken {
		s.log.Error("Invalid or expired session")
		return nil, fmt.Errorf("invalid or expired session")
	}

	s.log.Info("User info retrieved successfully", "wxUserId", wxUserID)
	return user, nil
}

// GetUserInfoByCode 根据授权码获取用户信息
func (s *WXAuthService) GetUserInfoByCode(code string) (*response.WXUserInfoResponse, error) {
	// 获取访问令牌
	tokenResult, err := s.getAccessToken(code)
	if err != nil {
		s.log.Error("Failed to get access token", "error", err)
		return nil, err
	}

	// 获取用户信息
	userInfo, err := s.getUserInfo(tokenResult.AccessToken, tokenResult.OpenID)
	if err != nil {
		s.log.Error("Failed to get user info", "error", err)
		return nil, err
	}

	s.log.Info("Wechat user info retrieved successfully", "openId", tokenResult.OpenID)
	return userInfo, nil
}

// getAccessToken 获取微信访问令牌
func (s *WXAuthService) getAccessToken(code string) (*response.WechatAuthResult, error) {
	// 构建请求URL
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.config.Wechat.AppID,
		s.config.Wechat.AppSecret,
		code,
	)

	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result response.WechatAuthResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查错误
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %d - %s", result.ErrCode, result.ErrMsg)
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
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result response.WXUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GenerateWXAuthPayParams 生成微信支付参数
func (s *WXAuthService) GenerateWXAuthPayParams(orderID string, amount int64, description string) (map[string]interface{}, error) {
	// 这里应该实现微信支付参数的生成逻辑
	// 实际实现需要调用微信支付API，生成签名等

	// 简化返回模拟数据
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := generateNonceStr()

	return map[string]interface{}{
		"appId":     s.config.Wechat.AppID,
		"timeStamp": timeStamp,
		"nonceStr":  nonceStr,
		"package":   "prepay_id=wx1234567890",
		"signType":  "MD5",
		"paySign":   "1234567890abcdef",
	}, nil
}

// HandleWXAuthPayNotify 处理微信支付通知
func (s *WXAuthService) HandleWXAuthPayNotify(notifyData []byte) (map[string]interface{}, error) {
	// 这里应该实现微信支付通知的处理逻辑
	// 实际实现需要验证签名，解析通知数据等

	// 简化返回成功响应
	return map[string]interface{}{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	}, nil
}

// BindWXUserToSystemUser 绑定微信用户到系统用户
func (s *WXAuthService) BindWXUserToSystemUser(openID string, userID uint) error {
	// 查询微信用户
	wxUser, err := s.wxUserRepo.GetByOpenID(openID)
	if err != nil {
		s.log.Error("Failed to get wx user", "error", err)
		return fmt.Errorf("wx user not found: %w", err)
	}

	// 更新微信用户的系统用户ID
	wxUser.UserID = userID
	wxUser.UpdatedAt = time.Now()
	if err := s.wxUserRepo.Update(wxUser); err != nil {
		s.log.Error("Failed to update wx user", "error", err)
		return fmt.Errorf("failed to bind user: %w", err)
	}

	s.log.Info("WX user bound to system user", "wxUserId", wxUser.ID, "systemUserId", userID)
	return nil
}

// UnbindWXUserFromSystemUser 解绑微信用户与系统用户
func (s *WXAuthService) UnbindWXUserFromSystemUser(openID string) error {
	// 查询微信用户
	wxUser, err := s.wxUserRepo.GetByOpenID(openID)
	if err != nil {
		s.log.Error("Failed to get wx user", "error", err)
		return fmt.Errorf("wx user not found: %w", err)
	}

	// 解绑系统用户ID
	wxUser.UserID = 0
	wxUser.UpdatedAt = time.Now()
	if err := s.wxUserRepo.Update(wxUser); err != nil {
		s.log.Error("Failed to update wx user", "error", err)
		return fmt.Errorf("failed to unbind user: %w", err)
	}

	// 使所有会话失效
	if err := s.wxAuthSessionRepo.InvalidateAllSessionsByWXUserID(wxUser.ID); err != nil {
		s.log.Error("Failed to invalidate sessions", "error", err)
	}

	s.log.Info("WX user unbound from system user", "wxUserId", wxUser.ID)
	return nil
}

// UpdateWXUserInfo 更新微信用户信息
func (s *WXAuthService) UpdateWXUserInfo(user *model.WXUser) error {
	// 验证用户存在
	existingUser, err := s.wxUserRepo.GetByID(user.ID)
	if err != nil {
		s.log.Error("User not found", "error", err)
		return fmt.Errorf("user not found: %w", err)
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
	existingUser.UpdatedAt = time.Now()

	// 保存更新
	if err := s.wxUserRepo.Update(existingUser); err != nil {
		s.log.Error("Failed to update user info", "error", err)
		return fmt.Errorf("failed to update user info: %w", err)
	}

	s.log.Info("WX user info updated", "wxUserId", user.ID)
	return nil
}

// Logout 用户登出
func (s *WXAuthService) Logout(accessToken string) error {
	// 解析访问令牌
	claims, err := s.jwtAuth.ParseToken(accessToken)
	if err != nil {
		s.log.Error("Failed to parse access token", "error", err)
		return err
	}

	// 获取用户ID
	wxUserID := uint(claims.UserID)

	// 查询用户会话
	session, err := s.wxAuthSessionRepo.GetActiveSessionByWXUserID(wxUserID)
	if err != nil {
		s.log.Error("Failed to get session", "error", err)
		return err
	}

	// 使会话失效
	if err := s.wxAuthSessionRepo.InvalidateSession(session.ID); err != nil {
		s.log.Error("Failed to invalidate session", "error", err)
		return err
	}

	s.log.Info("User logged out successfully", "wxUserId", wxUserID)
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
