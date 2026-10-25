package entity

import "time"

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"   // 待支付
	PaymentStatusSuccess   PaymentStatus = "success"   // 支付成功
	PaymentStatusFailed    PaymentStatus = "failed"    // 支付失败
	PaymentStatusRefunded  PaymentStatus = "refunded"  // 已退款
	PaymentStatusCancelled PaymentStatus = "cancelled" // 已取消
)

// 订单类型（对齐 Gpay model.OrderType*）
const (
	PaymentOrderTypeJSAPI   = "jsapi"   // 普通JSAPI支付(微信内 H5/小程序)
	PaymentOrderTypeNative  = "native"  // Native扫码支付
	PaymentOrderTypeMedical = "medical" // 医保自费混合支付
	PaymentOrderTypeCombine = "combine" // JSAPI合单支付主单(2-10笔子单合并)
)

// Payment 支付领域实体
// 表示支付订单的完整信息（同时承载普通支付与医保支付订单，对齐 Gpay model.Order）
// 纯业务模型，无 GORM 标签
type Payment struct {
	ID            int64          // 支付 ID
	OrderID       string         // 订单 ID（商户订单号 out_trade_no）
	UserID        int64          // 用户 ID
	Amount        int64          // 金额，单位：分
	Currency      string         // 货币类型，默认 CNY
	Method        string         // 支付方式：wechat, alipay 等
	Status        PaymentStatus  // 支付状态
	TransactionID string         // 交易 ID（第三方支付平台返回）
	Description   string         // 支付描述
	WechatPayInfo *WechatPayInfo // 微信支付详细信息
	PaidAt        *time.Time     // 支付成功时间
	CreatedBy     int64          // 创建人 ID
	CreatedAt     time.Time      // 创建时间
	UpdatedBy     int64          // 更新人 ID
	UpdatedAt     time.Time      // 更新时间

	// 扩展字段（对齐 Gpay model.Order，普通/医保支付共用）
	OrderType      string        // 订单类型：jsapi | native | medical
	RefundedFee    int64         // 已累计退款成功金额(分)
	Openid         string        // 支付用户openid
	CodeURL        string        // Native二维码链接(有效期2小时)
	TradeState     string        // 微信侧交易状态 SUCCESS/NOTPAY/CLOSED/REFUND/PAYERROR/MIX_PAY_*
	MixTradeNo     string        // 医保混合订单号
	MedInfo        *MedOrderInfo // 医保订单信息(医疗订单必有)
	WxLastResponse string        // 微信侧原始回执(查单/回调原文，便于排查)
}

// IsPaid 检查支付是否已完成
func (p *Payment) IsPaid() bool {
	return p.Status == PaymentStatusSuccess
}

// IsPending 检查支付是否待处理
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsFailed 检查支付是否失败
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

// IsRefunded 检查支付是否已退款
func (p *Payment) IsRefunded() bool {
	return p.Status == PaymentStatusRefunded
}

// IsCancelled 检查支付是否已取消
func (p *Payment) IsCancelled() bool {
	return p.Status == PaymentStatusCancelled
}

// MarkAsPaid 标记支付为已支付
func (p *Payment) MarkAsPaid(transactionID string, paidAt time.Time) {
	p.Status = PaymentStatusSuccess
	p.TransactionID = transactionID
	p.PaidAt = &paidAt
}

// MarkAsFailed 标记支付为失败
func (p *Payment) MarkAsFailed() {
	p.Status = PaymentStatusFailed
}

// MarkAsRefunded 标记支付为已退款
func (p *Payment) MarkAsRefunded() {
	p.Status = PaymentStatusRefunded
}

// MarkAsCancelled 标记支付为已取消
func (p *Payment) MarkAsCancelled() {
	p.Status = PaymentStatusCancelled
}
