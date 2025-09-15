package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"
)

// WechatService 微信服务
type WechatService struct {
	config *config.Config
	log    logger.Logger
}

// WechatUserInfo 微信用户信息
type WechatUserInfo struct {
	OpenID    string `json:"openid"`
	UnionID   string `json:"unionid"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"headimgurl"`
	Sex       int    `json:"sex"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Language  string `json:"language"`
}

// WechatAuthResult 微信授权结果
type WechatAuthResult struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// NewWechatService 创建微信服务实例
func NewWechatService(config *config.Config, log logger.Logger) *WechatService {
	return &WechatService{
		config: config,
		log:    log,
	}
}

// GetUserInfoByCode 根据授权码获取用户信息
func (s *WechatService) GetUserInfoByCode(code string) (*WechatUserInfo, error) {
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

	// 设置UnionID
	userInfo.UnionID = tokenResult.UnionID

	return userInfo, nil
}

// getAccessToken 获取微信访问令牌
func (s *WechatService) getAccessToken(code string) (*WechatAuthResult, error) {
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
	var result WechatAuthResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查错误
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// getUserInfo 获取微信用户信息
func (s *WechatService) getUserInfo(accessToken, openID string) (*WechatUserInfo, error) {
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
	var result WechatUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GenerateWechatPayParams 生成微信支付参数
func (s *WechatService) GenerateWechatPayParams(orderID string, amount int64, description string) (map[string]interface{}, error) {
	// 这里应该实现微信支付参数的生成逻辑
	// 实际实现需要调用微信支付API，生成签名等

	// 简化返回模拟数据
	return map[string]interface{}{
		"appId":     s.config.Wechat.AppID,
		"timeStamp": fmt.Sprintf("%d", time.Now().Unix()),
		"nonceStr":  generateNonceStr(),
		"package":   "prepay_id=wx1234567890",
		"signType":  "MD5",
		"paySign":   "1234567890abcdef",
	}, nil
}

// HandleWechatPayNotify 处理微信支付通知
func (s *WechatService) HandleWechatPayNotify(notifyData []byte) (map[string]interface{}, error) {
	// 这里应该实现微信支付通知的处理逻辑
	// 实际实现需要验证签名，解析通知数据等

	// 简化返回成功响应
	return map[string]interface{}{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	}, nil
}

// generateNonceStr 生成随机字符串
func generateNonceStr() string {
	// 实现随机字符串生成逻辑
	// 简化实现，返回固定值
	return "abcdefghijklmnopqrstuvwxyz0123456789"
}
