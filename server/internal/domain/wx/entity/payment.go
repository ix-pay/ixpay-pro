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

// Payment 支付领域实体
// 表示支付订单的完整信息
// 纯业务模型，无 GORM 标签
type Payment struct {
	ID            string         // 支付 ID（string 类型，避免 JSON 精度丢失）
	OrderID       string         // 订单 ID
	UserID        string         // 用户 ID（string 类型，避免 JSON 精度丢失）
	Amount        int64          // 金额，单位：分
	Currency      string         // 货币类型，默认 CNY
	Method        string         // 支付方式：wechat, alipay 等
	Status        PaymentStatus  // 支付状态
	TransactionID string         // 交易 ID（第三方支付平台返回）
	Description   string         // 支付描述
	WechatPayInfo *WechatPayInfo // 微信支付详细信息
	PaidAt        *time.Time     // 支付成功时间
	CreatedBy     string         // 创建人 ID
	CreatedAt     time.Time      // 创建时间
	UpdatedBy     string         // 更新人 ID
	UpdatedAt     time.Time      // 更新时间
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
