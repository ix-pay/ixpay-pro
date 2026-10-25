package request

// CreatePaymentRequest 创建支付（待支付记录）请求参数
// orderId 不传时由服务端自动生成；currency 默认 CNY；paymentMethod 默认 wechat
type CreatePaymentRequest struct {
	OrderID       string  `json:"orderId" binding:"omitempty"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"omitempty,oneof=CNY USD"`
	PaymentMethod string  `json:"paymentMethod" binding:"omitempty,oneof=wechat wechat_alipay alipay"`
	Description   string  `json:"description"`
}
