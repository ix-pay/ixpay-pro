// Package wxpayv3 微信支付 APIv3 协议层（直连/服务商模式）
// 严格使用微信支付官方 WXPayUtility 工具类（Go 版）完成签名、验签、解密、请求头构建，
// 与 Gpay 项目 wechatpay 协议层的实现方式保持一致。
package wxpayv3

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	wxpay_utility "github.com/ix-pay/ixpay-pro/internal/infrastructure/payment/wxpayutility"
)

const (
	// Host 微信支付 APIv3 主域名
	Host = "https://api.mch.weixin.qq.com"

	// HeaderSerial 敏感字段加密时的序列号请求头
	HeaderSerial = "Wechatpay-Serial"
)

// ClientConfig V3 客户端配置
type ClientConfig struct {
	MchID       string          // 签名主体商户号（直连=商户号，服务商=sp_mchid）
	SerialNo    string          // 商户 API 证书序列号
	APIv3Key    string          // APIv3 密钥（回调解密）
	PrivateKey  *rsa.PrivateKey // 商户私钥（请求签名/调起支付签名/敏感字段加密）
	PublicKeyID string          // 微信支付公钥 ID（应答/回调验签用，可空：跳过应答验签）
	PublicKey   *rsa.PublicKey  // 微信支付公钥
	// 服务商模式（可选）
	IsPartner bool   // 是否服务商模式
	SpMchID   string // 服务商商户号
	SubMchID  string // 子商户号
	SubAppID  string // 子商户 AppID（可选）
}

// Client 微信支付 APIv3 客户端（直连/服务商共用）
type Client struct {
	cfg        ClientConfig
	httpClient *http.Client
}

// NewClient 创建 V3 客户端
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// IsPartner 是否服务商模式
func (c *Client) IsPartner() bool {
	return c.cfg.IsPartner
}

// GetPublicKey 获取微信支付公钥（医保敏感字段/异常退款卡号姓名加密用）
func (c *Client) GetPublicKey() *rsa.PublicKey {
	return c.cfg.PublicKey
}

// GetPublicKeyID 获取微信支付公钥 ID
func (c *Client) GetPublicKeyID() string {
	return c.cfg.PublicKeyID
}

// APIError 微信支付API错误（应答非2XX，与官方 ApiException 同构）
type APIError struct {
	StatusCode   int
	ErrorCode    string
	ErrorMessage string
	Body         string
}

func (e *APIError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("微信支付API错误[%d %s]: %s", e.StatusCode, e.ErrorCode, e.ErrorMessage)
	}
	return fmt.Sprintf("微信支付API错误[%d]: %s", e.StatusCode, e.Body)
}

// doRequest 发送请求并验证应答签名（urlPath为相对主域名的路径）
// extraHeaders: 附加请求头（如医保下单/异常退款的 Wechatpay-Serial）
func (c *Client) doRequest(method, urlPath string, body, result any, extraHeaders map[string]string) error {
	respHeader, respBody, err := c.doHTTP(method, Host+urlPath, body, extraHeaders)
	if err != nil {
		return err
	}

	// 验证应答签名（已配置微信支付公钥时，委托官方工具类）
	if c.cfg.PublicKey != nil {
		if err := wxpay_utility.ValidateResponse(c.cfg.PublicKeyID, c.cfg.PublicKey, &respHeader, respBody); err != nil {
			return fmt.Errorf("应答验签失败: %w", err)
		}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("解析应答失败: %w", err)
		}
	}
	return nil
}

// doHTTP 发送签名请求并返回应答头与原始报文（委托官方工具类构建 Authorization）
// 供账单下载使用：下载应答为账单数据本身，完整性由申请账单返回的 hash_value 校验
func (c *Client) doHTTP(method, fullURL string, body any, extraHeaders map[string]string) (http.Header, []byte, error) {
	u, err := url.Parse(fullURL)
	if err != nil {
		return nil, nil, fmt.Errorf("URL非法: %w", err)
	}
	if u.Scheme != "https" || (u.Host != "api.mch.weixin.qq.com" && u.Host != "api2.mch.weixin.qq.com") {
		return nil, nil, fmt.Errorf("仅允许访问微信支付域名，拒绝: %s", u.Host)
	}

	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
	}

	// 官方工具类：签名串 METHOD\nURL(含查询参数)\n时间戳\n随机串\n请求体\n
	auth, err := wxpay_utility.BuildAuthorization(c.cfg.MchID, c.cfg.SerialNo, c.cfg.PrivateKey, method, u.RequestURI(), bodyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("构造签名失败: %w", err)
	}

	req, err := http.NewRequest(method, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("请求微信支付失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("读取应答失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
		var errResp struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			apiErr.ErrorCode = errResp.Code
			apiErr.ErrorMessage = errResp.Message
		}
		return nil, nil, apiErr
	}
	return resp.Header, respBody, nil
}

// ---------- 请求/应答模型 ----------

// jsapiPrepayRequest JSAPI 下单请求
type jsapiPrepayRequest struct {
	AppID       string  `json:"appid"`
	MchID       string  `json:"mchid"`
	Description string  `json:"description"`
	OutTradeNo  string  `json:"out_trade_no"`
	NotifyURL   string  `json:"notify_url"`
	Amount      *amount `json:"amount,omitempty"`
	Payer       *payer  `json:"payer,omitempty"`
	TimeExpire  string  `json:"time_expire,omitempty"`
	Attach      string  `json:"attach,omitempty"`
}

// amount 订单金额
type amount struct {
	Total    int64  `json:"total"`
	Currency string `json:"currency"`
}

// payer 支付者
type payer struct {
	Openid string `json:"openid"`
}

// prepayResponse 下单应答
type prepayResponse struct {
	PrepayID string `json:"prepay_id"`
}

// QueryOrderResult 查单应答（与微信侧字段对齐）
type QueryOrderResult struct {
	AppID          string `json:"appid"`
	MchID          string `json:"mchid"`
	OutTradeNo     string `json:"out_trade_no"`
	TransactionID  string `json:"transaction_id"`
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	SuccessTime    string `json:"success_time,omitempty"`
	Amount         *struct {
		Total         int64  `json:"total"`
		PayerTotal    int64  `json:"payer_total,omitempty"`
		Currency      string `json:"currency"`
		PayerCurrency string `json:"payer_currency,omitempty"`
	} `json:"amount,omitempty"`
}

// JSAPIPayParams 前端调起支付参数（H5 公众号 JS-SDK 与小程序 wx.requestPayment 通用）
type JSAPIPayParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// ---------- 支付方法（直连/服务商自动切换，与 Gpay 一致） ----------

// JSAPIPrepay JSAPI 下单
// 直连: POST /v3/pay/transactions/jsapi
// 服务商: POST /v3/pay/partner/transactions/jsapi
func (c *Client) JSAPIPrepay(appID, openid, outTradeNo, description, notifyURL string, totalFee int64) (string, error) {
	if c.IsPartner() {
		req := &partnerPrepayRequest{
			SpAppID:     c.partnerSpAppID(),
			SpMchID:     c.cfg.SpMchID,
			SubMchID:    c.cfg.SubMchID,
			Description: description,
			OutTradeNo:  outTradeNo,
			NotifyURL:   notifyURL,
			Amount:      &amount{Total: totalFee, Currency: "CNY"},
			Payer:       &partnerPayer{},
		}
		// openid 主体：授权 appid=sub_appid 时用 sub_openid，否则 sp_openid
		if c.useSubAppID(appID) {
			req.SubAppID = c.cfg.SubAppID
			req.Payer = &partnerPayer{SubOpenid: openid}
		} else {
			req.Payer = &partnerPayer{SpOpenid: openid}
		}
		var resp prepayResponse
		if err := c.doRequest("POST", "/v3/pay/partner/transactions/jsapi", req, &resp, nil); err != nil {
			return "", err
		}
		if resp.PrepayID == "" {
			return "", fmt.Errorf("JSAPI下单失败: 应答缺少prepay_id")
		}
		return resp.PrepayID, nil
	}

	req := &jsapiPrepayRequest{
		AppID:       appID,
		MchID:       c.cfg.MchID,
		Description: description,
		OutTradeNo:  outTradeNo,
		NotifyURL:   notifyURL,
		Amount:      &amount{Total: totalFee, Currency: "CNY"},
		Payer:       &payer{Openid: openid},
	}
	var resp prepayResponse
	if err := c.doRequest("POST", "/v3/pay/transactions/jsapi", req, &resp, nil); err != nil {
		return "", err
	}
	if resp.PrepayID == "" {
		return "", fmt.Errorf("JSAPI下单失败: 应答缺少prepay_id")
	}
	return resp.PrepayID, nil
}

// NativePrepay Native 下单
// 直连: POST /v3/pay/transactions/native
// 服务商: POST /v3/pay/partner/transactions/native
// 返回二维码链接 code_url（有效期 2 小时）
func (c *Client) NativePrepay(appID, outTradeNo, description, notifyURL string, totalFee int64) (string, error) {
	if c.IsPartner() {
		req := &partnerPrepayRequest{
			SpAppID:     c.partnerSpAppID(),
			SpMchID:     c.cfg.SpMchID,
			SubMchID:    c.cfg.SubMchID,
			Description: description,
			OutTradeNo:  outTradeNo,
			NotifyURL:   notifyURL,
			Amount:      &amount{Total: totalFee, Currency: "CNY"},
		}
		var resp nativePrepayResponse
		if err := c.doRequest("POST", "/v3/pay/partner/transactions/native", req, &resp, nil); err != nil {
			return "", err
		}
		if resp.CodeURL == "" {
			return "", fmt.Errorf("Native下单失败: 应答缺少code_url")
		}
		return resp.CodeURL, nil
	}

	req := &jsapiPrepayRequest{
		AppID:       appID,
		MchID:       c.cfg.MchID,
		Description: description,
		OutTradeNo:  outTradeNo,
		NotifyURL:   notifyURL,
		Amount:      &amount{Total: totalFee, Currency: "CNY"},
	}
	var resp nativePrepayResponse
	if err := c.doRequest("POST", "/v3/pay/transactions/native", req, &resp, nil); err != nil {
		return "", err
	}
	if resp.CodeURL == "" {
		return "", fmt.Errorf("Native下单失败: 应答缺少code_url")
	}
	return resp.CodeURL, nil
}

// QueryOrderByOutTradeNo 商户订单号查询订单（直连/服务商自动切换）
func (c *Client) QueryOrderByOutTradeNo(outTradeNo string) (*QueryOrderResult, error) {
	path := c.orderQueryPath(fmt.Sprintf("out-trade-no/%s", outTradeNo))
	var resp QueryOrderResult
	if err := c.doRequest("GET", path, nil, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryOrderByTransactionID 微信支付订单号查询订单
func (c *Client) QueryOrderByTransactionID(transactionID string) (*QueryOrderResult, error) {
	path := c.orderQueryPath(fmt.Sprintf("id/%s", transactionID))
	var resp QueryOrderResult
	if err := c.doRequest("GET", path, nil, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CloseOrder 关闭订单
func (c *Client) CloseOrder(outTradeNo string) error {
	if c.IsPartner() {
		path := fmt.Sprintf("/v3/pay/partner/transactions/out-trade-no/%s/close", outTradeNo)
		return c.doRequest("POST", path, &partnerCloseOrderRequest{SpMchID: c.cfg.SpMchID, SubMchID: c.cfg.SubMchID}, nil, nil)
	}
	path := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s/close", outTradeNo)
	return c.doRequest("POST", path, map[string]string{"mchid": c.cfg.MchID}, nil, nil)
}

// BuildJSAPIPayParams 构造前端调起支付参数并签名（小程序 wx.requestPayment 与公众号 H5 通用）
// 签名串: appId\n时间戳\n随机串\nprepay_id=xxx\n（官方工具类 SHA256withRSA 签名）
func (c *Client) BuildJSAPIPayParams(prepayID, appID string) (*JSAPIPayParams, error) {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr, err := wxpay_utility.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("生成随机串失败: %w", err)
	}
	packageStr := "prepay_id=" + prepayID

	message := appID + "\n" + timeStamp + "\n" + nonceStr + "\n" + packageStr + "\n"
	paySign, err := wxpay_utility.SignSHA256WithRSA(message, c.cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("调起支付参数签名失败: %w", err)
	}

	return &JSAPIPayParams{
		AppID:     appID,
		TimeStamp: timeStamp,
		NonceStr:  nonceStr,
		Package:   packageStr,
		SignType:  "RSA",
		PaySign:   paySign,
	}, nil
}

// ParseNotification 验签并解密回调通知（委托官方工具类），返回通知与解密后的业务 JSON
func (c *Client) ParseNotification(header http.Header, body []byte) (*wxpay_utility.Notification, []byte, error) {
	if c.cfg.PublicKey == nil {
		return nil, nil, fmt.Errorf("未配置微信支付公钥，无法验证回调签名")
	}
	n, err := wxpay_utility.ParseNotification(c.cfg.PublicKeyID, c.cfg.PublicKey, c.cfg.APIv3Key, &header, body)
	if err != nil {
		return nil, nil, fmt.Errorf("回调验签/解密失败: %w", err)
	}
	return n, []byte(n.Plaintext), nil
}

// ---------- 服务商模式辅助 ----------

// partnerPrepayRequest 服务商下单请求
type partnerPrepayRequest struct {
	SpAppID     string        `json:"sp_appid"`
	SpMchID     string        `json:"sp_mchid"`
	SubMchID    string        `json:"sub_mchid"`
	SubAppID    string        `json:"sub_appid,omitempty"`
	Description string        `json:"description"`
	OutTradeNo  string        `json:"out_trade_no"`
	NotifyURL   string        `json:"notify_url"`
	Amount      *amount       `json:"amount,omitempty"`
	Payer       *partnerPayer `json:"payer,omitempty"`
	TimeExpire  string        `json:"time_expire,omitempty"`
	Attach      string        `json:"attach,omitempty"`
}

// partnerPayer 服务商支付者
type partnerPayer struct {
	SpOpenid  string `json:"sp_openid,omitempty"`
	SubOpenid string `json:"sub_openid,omitempty"`
}

// partnerCloseOrderRequest 服务商关单请求
type partnerCloseOrderRequest struct {
	SpMchID  string `json:"sp_mchid"`
	SubMchID string `json:"sub_mchid"`
}

// nativePrepayResponse Native 下单应答
type nativePrepayResponse struct {
	CodeURL string `json:"code_url"`
}

// useSubAppID 服务商模式下是否以 sub_appid 为主体下单（授权/调起 appid 等于 sub_appid 时）
func (c *Client) useSubAppID(appID string) bool {
	return c.cfg.SubAppID != "" && appID == c.cfg.SubAppID
}

// partnerSpAppID 服务商模式下单的 sp_appid
func (c *Client) partnerSpAppID() string {
	if c.cfg.SubAppID != "" {
		return c.cfg.SubAppID
	}
	return ""
}

// orderQueryPath 查单路径（直连与服务商模式差异封装）
func (c *Client) orderQueryPath(idPath string) string {
	if c.IsPartner() {
		return fmt.Sprintf("/v3/pay/partner/transactions/%s?sp_mchid=%s&sub_mchid=%s", idPath, c.cfg.SpMchID, c.cfg.SubMchID)
	}
	return fmt.Sprintf("/v3/pay/transactions/%s?mchid=%s", idPath, c.cfg.MchID)
}
