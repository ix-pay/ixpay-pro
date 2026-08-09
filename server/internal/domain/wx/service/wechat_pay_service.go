package service

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	baseRepo "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// 微信支付 API 地址
const wechatUnifiedOrderURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"

// WechatPayService 微信支付服务
// 处理与微信支付 API 的交互，包括统一下单、签名生成、通知验证等
// 优先从数据库（config_seed）读取微信配置，YAML 配置作为降级方案
type WechatPayService struct {
	cfg     config.WechatConfig    // YAML 配置（降级方案）
	cfgRepo baseRepo.ConfigRepository // 数据库配置仓库
	dbConfig *config.WechatConfig  // 数据库加载的配置（覆盖 YAML）
	log     logger.Logger
	http    *http.Client
	mu      sync.RWMutex
}

// NewWechatPayService 创建微信支付服务实例
func NewWechatPayService(cfg *config.Config, cfgRepo baseRepo.ConfigRepository, log logger.Logger) *WechatPayService {
	return &WechatPayService{
		cfg:     cfg.Wechat,
		cfgRepo: cfgRepo,
		log:     log,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// getConfig 返回当前有效的配置（数据库配置优先于 YAML 配置）
func (s *WechatPayService) getConfig() config.WechatConfig {
	s.mu.RLock()
	if s.dbConfig != nil {
		defer s.mu.RUnlock()
		return *s.dbConfig
	}
	s.mu.RUnlock()
	return s.cfg
}

// LoadConfigFromDB 从数据库加载微信配置，覆盖 YAML 配置
// 配置键在 config_seed.go 中定义：wechat_app_id, wechat_app_secret, wechat_mch_id, wechat_api_key, wechat_notify_url
func (s *WechatPayService) LoadConfigFromDB() error {
	// 从数据库批量读取微信配置
	keys := []string{"wechat_app_id", "wechat_app_secret", "wechat_mch_id", "wechat_api_key", "wechat_notify_url"}
	configMap := make(map[string]string, len(keys))

	for _, key := range keys {
		cfg, err := s.cfgRepo.GetByKey(key)
		if err != nil {
			s.log.Warn("从数据库读取微信配置失败", "key", key, "error", err)
			continue
		}
		configMap[key] = cfg.ConfigValue
	}

	// 如果有任何配置读取成功，则更新
	if len(configMap) == 0 {
		s.log.Warn("数据库微信配置为空，使用 YAML 配置降级")
		return fmt.Errorf("数据库微信配置为空")
	}

	// 构建配置：优先使用数据库值，否则使用 YAML 值
	yamlCfg := s.cfg
	dbCfg := &config.WechatConfig{
		AppID:            getConfigValue(configMap, "wechat_app_id", yamlCfg.AppID),
		AppSecret:        getConfigValue(configMap, "wechat_app_secret", yamlCfg.AppSecret),
		MCHID:            getConfigValue(configMap, "wechat_mch_id", yamlCfg.MCHID),
		APIKey:           getConfigValue(configMap, "wechat_api_key", yamlCfg.APIKey),
		NotifyURL:        getConfigValue(configMap, "wechat_notify_url", yamlCfg.NotifyURL),
		RedirectURL:      yamlCfg.RedirectURL,
		OAuthCallbackURL: yamlCfg.OAuthCallbackURL,
		FrontendURL:      yamlCfg.FrontendURL,
	}

	s.mu.Lock()
	s.dbConfig = dbCfg
	s.mu.Unlock()

	s.log.Info("从数据库加载微信配置成功")
	return nil
}

// getConfigValue 从 map 中取值，如果为空则使用默认值
func getConfigValue(m map[string]string, key, defaultVal string) string {
	if v, ok := m[key]; ok && v != "" && v != "your-wechat-app-id" && v != "your-wechat-app-secret" && v != "your-wechat-mch-id" && v != "your-wechat-api-key" && v != "http://your-server.com" {
		return v
	}
	return defaultVal
}

// ---------- XML 结构体定义 ----------

// unifiedOrderRequest 统一下单请求 XML 结构
type unifiedOrderRequest struct {
	XMLName        xml.Name `xml:"xml"`
	AppID          string   `xml:"appid"`
	MchID          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	Body           string   `xml:"body"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TotalFee       int64    `xml:"total_fee"`
	SpbillCreateIP string   `xml:"spbill_create_ip"`
	NotifyURL      string   `xml:"notify_url"`
	TradeType      string   `xml:"trade_type"`
	OpenID         string   `xml:"openid"`
}

// unifiedOrderResponse 统一下单响应 XML 结构
type unifiedOrderResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	AppID      string   `xml:"appid,omitempty"`
	MchID      string   `xml:"mch_id,omitempty"`
	NonceStr   string   `xml:"nonce_str,omitempty"`
	Sign       string   `xml:"sign,omitempty"`
	ResultCode string   `xml:"result_code,omitempty"`
	ErrCode    string   `xml:"err_code,omitempty"`
	ErrCodeDes string   `xml:"err_code_des,omitempty"`
	PrepayID   string   `xml:"prepay_id,omitempty"`
	TradeType  string   `xml:"trade_type,omitempty"`
	CodeURL    string   `xml:"code_url,omitempty"`
}

// WechatPayNotifyResult 微信支付通知结果
type WechatPayNotifyResult struct {
	XMLName       xml.Name `xml:"xml"`
	ReturnCode    string   `xml:"return_code"`
	ReturnMsg     string   `xml:"return_msg"`
	AppID         string   `xml:"appid"`
	MchID         string   `xml:"mch_id"`
	DeviceInfo    string   `xml:"device_info,omitempty"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	SignType      string   `xml:"sign_type,omitempty"`
	ResultCode    string   `xml:"result_code"`
	ErrCode       string   `xml:"err_code,omitempty"`
	ErrCodeDes    string   `xml:"err_code_des,omitempty"`
	OpenID        string   `xml:"openid"`
	IsSubscribe   string   `xml:"is_subscribe"`
	TradeType     string   `xml:"trade_type"`
	BankType      string   `xml:"bank_type"`
	TotalFee      int64    `xml:"total_fee"`
	SettlementTotalFee int64 `xml:"settlement_total_fee,omitempty"`
	FeeType       string   `xml:"fee_type,omitempty"`
	CashFee       int64    `xml:"cash_fee"`
	CashFeeType   string   `xml:"cash_fee_type,omitempty"`
	CouponFee     int64    `xml:"coupon_fee,omitempty"`
	CouponCount   int64    `xml:"coupon_count,omitempty"`
	TransactionID string   `xml:"transaction_id"`
	OutTradeNo    string   `xml:"out_trade_no"`
	Attach        string   `xml:"attach,omitempty"`
	TimeEnd       string   `xml:"time_end"`
}

// ---------- 公开方法 ----------

// JSAPIParams JSAPI 调起支付参数
type JSAPIParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// UnifiedOrder 调用微信统一下单 API，返回 prepay_id
func (s *WechatPayService) UnifiedOrder(openID, orderID, description, clientIP string, amount int64) (string, error) {
	nonceStr := generateNonceStr()

	cfg := s.getConfig()

	// 构建请求参数 map
	params := map[string]string{
		"appid":            cfg.AppID,
		"mch_id":           cfg.MCHID,
		"nonce_str":        nonceStr,
		"body":             description,
		"out_trade_no":     orderID,
		"total_fee":        fmt.Sprintf("%d", amount),
		"spbill_create_ip": clientIP,
		"notify_url":       cfg.NotifyURL,
		"trade_type":       "JSAPI",
		"openid":           openID,
	}

	// 生成签名
	sign := s.GenerateSign(params)

	// 构建 XML 请求体
	reqBody := &unifiedOrderRequest{
		AppID:          cfg.AppID,
		MchID:          cfg.MCHID,
		NonceStr:       nonceStr,
		Sign:           sign,
		Body:           description,
		OutTradeNo:     orderID,
		TotalFee:       amount,
		SpbillCreateIP: clientIP,
		NotifyURL:      cfg.NotifyURL,
		TradeType:      "JSAPI",
		OpenID:         openID,
	}

	xmlBytes, err := xml.Marshal(reqBody)
	if err != nil {
		s.log.Error("统一下单 XML 序列化失败", "error", err)
		return "", fmt.Errorf("统一下单 XML 序列化失败：%w", err)
	}

	// 发送请求
	resp, err := s.http.Post(wechatUnifiedOrderURL, "application/xml", bytes.NewReader(xmlBytes))
	if err != nil {
		s.log.Error("统一下单请求失败", "error", err)
		return "", fmt.Errorf("统一下单请求失败：%w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Error("统一下单读取响应失败", "error", err)
		return "", fmt.Errorf("统一下单读取响应失败：%w", err)
	}

	// 解析响应
	var orderResp unifiedOrderResponse
	if err := xml.Unmarshal(bodyBytes, &orderResp); err != nil {
		s.log.Error("统一下单 XML 解析失败", "error", err)
		return "", fmt.Errorf("统一下单 XML 解析失败：%w", err)
	}

	// 检查返回码
	if orderResp.ReturnCode != "SUCCESS" {
		s.log.Error("统一下单返回失败", "return_code", orderResp.ReturnCode, "return_msg", orderResp.ReturnMsg)
		return "", fmt.Errorf("统一下单失败：%s - %s", orderResp.ReturnCode, orderResp.ReturnMsg)
	}

	// 检查业务结果
	if orderResp.ResultCode != "SUCCESS" {
		s.log.Error("统一下单业务失败", "err_code", orderResp.ErrCode, "err_code_des", orderResp.ErrCodeDes)
		return "", fmt.Errorf("统一下单业务失败：%s - %s", orderResp.ErrCode, orderResp.ErrCodeDes)
	}

	// 验证签名
	respSignMap := map[string]string{
		"return_code": orderResp.ReturnCode,
		"appid":       orderResp.AppID,
		"mch_id":      orderResp.MchID,
		"nonce_str":   orderResp.NonceStr,
		"prepay_id":   orderResp.PrepayID,
		"result_code": orderResp.ResultCode,
		"trade_type":  orderResp.TradeType,
	}
	if orderResp.CodeURL != "" {
		respSignMap["code_url"] = orderResp.CodeURL
	}

	expectedSign := s.GenerateSign(respSignMap)
	if orderResp.Sign != "" && !strings.EqualFold(orderResp.Sign, expectedSign) {
		s.log.Error("统一下单响应签名验证失败", "expected", expectedSign, "actual", orderResp.Sign)
		return "", fmt.Errorf("统一下单响应签名验证失败")
	}

	s.log.Info("统一下单成功", "prepay_id", orderResp.PrepayID, "out_trade_no", orderID)
	return orderResp.PrepayID, nil
}

// GenerateJSAPIParams 根据 prepay_id 生成 JSAPI 调起支付参数
func (s *WechatPayService) GenerateJSAPIParams(prepayID string) *JSAPIParams {
	cfg := s.getConfig()
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := generateNonceStr()
	packageStr := "prepay_id=" + prepayID

	params := map[string]string{
		"appId":     cfg.AppID,
		"timeStamp": timeStamp,
		"nonceStr":  nonceStr,
		"package":   packageStr,
		"signType":  "MD5",
	}

	paySign := s.GenerateSign(params)

	return &JSAPIParams{
		AppID:     cfg.AppID,
		TimeStamp: timeStamp,
		NonceStr:  nonceStr,
		Package:   packageStr,
		SignType:  "MD5",
		PaySign:   paySign,
	}
}

// VerifyNotify 验证支付通知签名并解析通知数据
// 返回解析后的通知数据和验证结果
func (s *WechatPayService) VerifyNotify(body []byte) (*WechatPayNotifyResult, error) {
	var notify WechatPayNotifyResult
	if err := xml.Unmarshal(body, &notify); err != nil {
		s.log.Error("通知 XML 解析失败", "error", err)
		return nil, fmt.Errorf("通知 XML 解析失败：%w", err)
	}

	// 检查通信返回码
	if notify.ReturnCode != "SUCCESS" {
		s.log.Warn("通知通信失败", "return_code", notify.ReturnCode, "return_msg", notify.ReturnMsg)
		return &notify, nil
	}

	// 验证签名
	params := make(map[string]string)
	params["return_code"] = notify.ReturnCode
	params["appid"] = notify.AppID
	params["mch_id"] = notify.MchID
	params["nonce_str"] = notify.NonceStr
	params["result_code"] = notify.ResultCode
	params["openid"] = notify.OpenID
	params["is_subscribe"] = notify.IsSubscribe
	params["trade_type"] = notify.TradeType
	params["bank_type"] = notify.BankType
	params["total_fee"] = fmt.Sprintf("%d", notify.TotalFee)
	params["fee_type"] = notify.FeeType
	params["cash_fee"] = fmt.Sprintf("%d", notify.CashFee)
	params["transaction_id"] = notify.TransactionID
	params["out_trade_no"] = notify.OutTradeNo
	params["time_end"] = notify.TimeEnd

	if notify.SettlementTotalFee > 0 {
		params["settlement_total_fee"] = fmt.Sprintf("%d", notify.SettlementTotalFee)
	}
	if notify.CouponFee > 0 {
		params["coupon_fee"] = fmt.Sprintf("%d", notify.CouponFee)
	}
	if notify.CouponCount > 0 {
		params["coupon_count"] = fmt.Sprintf("%d", notify.CouponCount)
	}
	if notify.Attach != "" {
		params["attach"] = notify.Attach
	}

	expectedSign := s.GenerateSign(params)
	if !strings.EqualFold(notify.Sign, expectedSign) {
		s.log.Error("通知签名验证失败", "expected", expectedSign, "actual", notify.Sign)
		return nil, fmt.Errorf("通知签名验证失败")
	}

	s.log.Info("通知签名验证成功", "out_trade_no", notify.OutTradeNo)
	return &notify, nil
}

// GenerateSign 生成微信支付 MD5 签名
// 规则：将参数按 key 字典序排序，拼接为 key1=value1&key2=value2...&key=APIKey，然后 MD5 加密转大写
func (s *WechatPayService) GenerateSign(params map[string]string) string {
	// 提取所有非空 key 并排序
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接签名字符串
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
	}
	buf.WriteString("&key=")
	buf.WriteString(s.getConfig().APIKey)

	// MD5 加密
	hash := md5.Sum([]byte(buf.String()))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// GenerateNonceStr 生成随机字符串（32 位）
func (s *WechatPayService) GenerateNonceStr() string {
	return generateNonceStr()
}

// GetMCHID 获取商户号
func (s *WechatPayService) GetMCHID() string {
	return s.getConfig().MCHID
}

// GetAppID 获取微信公众号 AppID
func (s *WechatPayService) GetAppID() string {
	return s.getConfig().AppID
}

// BuildSuccessXMLResponse 构建成功 XML 响应
func BuildSuccessXMLResponse() []byte {
	return []byte("<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>")
}

// BuildFailXMLResponse 构建失败 XML 响应
func BuildFailXMLResponse(msg string) []byte {
	return []byte(fmt.Sprintf("<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[%s]]></return_msg></xml>", msg))
}