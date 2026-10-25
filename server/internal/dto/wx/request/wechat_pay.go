package request

// CreateUnifiedOrderRequest 微信统一下单请求参数
// 订单已存在（待支付）时金额/描述以已有记录为准，可不传；订单不存在时必传；
// mchId 不传时使用应用默认商户（多渠道），渠道由登录态 appKey 决定
type CreateUnifiedOrderRequest struct {
	OrderID     string   `json:"orderId" binding:"required"`
	Amount      *float64 `json:"amount" binding:"omitempty,gt=0"`
	Description string   `json:"description"`
	MchID       string   `json:"mchId"`
}

// NativePrepayRequest Native扫码下单请求参数（金额单位：元）
type NativePrepayRequest struct {
	OrderID     string   `json:"orderId" binding:"required"`
	Amount      *float64 `json:"amount" binding:"omitempty,gt=0"`
	Description string   `json:"description"`
	MchID       string   `json:"mchId"`
}

// NativeQRCodeRequest Native二维码请求参数
type NativeQRCodeRequest struct {
	OutTradeNo string `json:"outTradeNo" form:"outTradeNo" binding:"required"`
}

// RefundCreateRequest 退款申请请求参数（金额单位：分，对齐微信 V3）
type RefundCreateRequest struct {
	OutTradeNo string `json:"outTradeNo" binding:"required"`
	RefundFee  int64  `json:"refundFee" binding:"required,gt=0"`
	Reason     string `json:"reason"`
}

// RefundQueryRequest 查询退款请求参数
type RefundQueryRequest struct {
	OutRefundNo string `json:"outRefundNo" form:"outRefundNo" binding:"required"`
}

// RefundAbnormalRequest 发起异常退款请求参数（卡号/姓名明文入参，服务端加密后出网）
type RefundAbnormalRequest struct {
	OutRefundNo string `json:"outRefundNo" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=USER_BANK_CARD MERCHANT_BANK_CARD"`
	BankType    string `json:"bankType"`
	BankAccount string `json:"bankAccount"`
	RealName    string `json:"realName"`
}

// CombinePrepayRequest JSAPI 合单支付下单请求参数
// 官方文档: https://pay.weixin.qq.com/doc/v3/merchant/4012556926
// 将 2-10 笔待支付订单合并为一笔合单订单统一下单（仅普通商户/直连模式支持）
type CombinePrepayRequest struct {
	PaymentIDs []string `json:"paymentIds" binding:"required,min=2,max=10"`
	MchID      string   `json:"mchId"` // 不传时使用应用默认商户
}
