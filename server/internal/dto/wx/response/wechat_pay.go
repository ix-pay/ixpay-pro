package response

// JSAPIParamsResponse JSAPI 支付参数响应
// 字段使用 camelCase 以匹配微信 JS-SDK 和 wxapp 前端期望的格式
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
	PrepayID    string              `json:"prepayId"`
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
	OAuthURL string `json:"oauthUrl"`
	State    string `json:"state,omitempty"`
}

// NativePrepayResponse Native 扫码下单响应
type NativePrepayResponse struct {
	CodeURL  string `json:"codeUrl"`  // 二维码链接(weixin://wxpay/bizpayurl?pr=xxx，有效期2小时)
	ExpireAt string `json:"expireAt"` // 订单有效期(time_expire)，前端倒计时用
}

// CombinePrepayResponse 合单支付下单响应
type CombinePrepayResponse struct {
	CombineOrderID string              `json:"combineOrderId"` // 合单商户订单号
	TotalAmount    float64             `json:"totalAmount"`    // 合单总金额(元)
	JSAPIParams    JSAPIParamsResponse `json:"jsapiParams"`    // 调起支付参数(与普通JSAPI一致)
}

// RefundResponse 退款单响应
type RefundResponse struct {
	OutRefundNo         string `json:"outRefundNo"`
	OutTradeNo          string `json:"outTradeNo"`
	RefundID            string `json:"refundId"`
	Reason              string `json:"reason,omitempty"`
	RefundFee           int64  `json:"refundFee"` // 分
	TotalFee            int64  `json:"totalFee"`  // 分
	Status              string `json:"status"`
	UserReceivedAccount string `json:"userReceivedAccount,omitempty"`
	SuccessTime         string `json:"successTime,omitempty"`
	CreatedAt           string `json:"createdAt,omitempty"`
}

// MedRefundResponse 医保退款单响应
type MedRefundResponse struct {
	HospOutRefundNo       string `json:"hospOutRefundNo"`
	OutTradeNo            string `json:"outTradeNo"`
	PartRefundType        string `json:"partRefundType,omitempty"`
	CashRefundFee         int64  `json:"cashRefundFee,omitempty"`
	InsRefundFee          int64  `json:"insRefundFee,omitempty"`
	MedRefundState        string `json:"medRefundState,omitempty"`
	CashRefundStatus      string `json:"cashRefundStatus,omitempty"`
	InsuranceRefundStatus string `json:"insuranceRefundStatus,omitempty"`
	Reason                string `json:"reason,omitempty"`
	CreatedAt             string `json:"createdAt,omitempty"`
}

// MedRefundQueryResponse 医保退款状态聚合响应
type MedRefundQueryResponse struct {
	HospOutRefundNo       string `json:"hospOutRefundNo"`
	HospOutTradeNo        string `json:"hospOutTradeNo,omitempty"`
	MedRefundID           string `json:"medRefundId,omitempty"`
	CashRefundID          string `json:"cashRefundId,omitempty"`
	InsuranceRefundID     string `json:"insuranceRefundId,omitempty"`
	CashRefundFee         int64  `json:"cashRefundFee,omitempty"`
	InsuranceRefundFee    int64  `json:"insuranceRefundFee,omitempty"`
	CashRefundStatus      string `json:"cashRefundStatus,omitempty"`
	InsuranceRefundStatus string `json:"insuranceRefundStatus,omitempty"`
	MedRefundState        string `json:"medRefundState,omitempty"`
	RefundEndTime         string `json:"refundEndTime,omitempty"`
	ResponseContent       string `json:"responseContent,omitempty"`
}

// MedAuthQueryResponse 医保授权查询响应
type MedAuthQueryResponse struct {
	UserName  string `json:"userName,omitempty"`
	CityID    string `json:"cityId,omitempty"`
	PayAuthNo string `json:"payAuthNo"`
	Longitude string `json:"longitude,omitempty"`
	Latitude  string `json:"latitude,omitempty"`
}

// MedInsPayParamsResponse 医保支付调起参数响应
// 注意：mixTradeNo/appid 为医保局调起协议要求的字段名（微信医保支付规范），豁免 camelCase 规则
type MedInsPayParamsResponse struct {
	MixTradeNo string `json:"mixTradeNo"`
	AppID      string `json:"appid"`
	TimeStamp  string `json:"timeStamp,omitempty"`
	NonceStr   string `json:"nonceStr,omitempty"`
	Package    string `json:"package,omitempty"`
	SignType   string `json:"signType,omitempty"`
	PaySign    string `json:"paySign,omitempty"`
}

// BillResponse 申请账单响应
type BillResponse struct {
	HashType    string `json:"hashType"`
	HashValue   string `json:"hashValue"`
	DownloadURL string `json:"downloadUrl"`
}
