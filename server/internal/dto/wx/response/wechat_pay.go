package response

// JSAPIParamsResponse JSAPI 支付参数响应
// 字段使用 camelCase 以匹配微信 JS-SDK 和 h5app 前端期望的格式
type JSAPIParamsResponse struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// UnifiedOrderResponse 统一下单响应
type UnifiedOrderResponse struct {
	PrepayID    string             `json:"prepayId"`
	JSAPIParams JSAPIParamsResponse `json:"jsapiParams"`
}

// WXJSAPIConfigResponse 微信 JS-SDK 配置响应（用于 wx.config()）
type WXJSAPIConfigResponse struct {
	AppID     string `json:"appId"`
	Timestamp string `json:"timestamp"`
	NonceStr  string `json:"nonceStr"`
	Signature string `json:"signature"`
}

// OAuthURLResponse OAuth 授权链接响应
type OAuthURLResponse struct {
	OAuthURL string `json:"oauth_url"`
	State    string `json:"state,omitempty"`
}