package model

import (
	"time"
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"   // 待支付
	PaymentStatusSuccess   PaymentStatus = "success"   // 支付成功
	PaymentStatusFailed    PaymentStatus = "failed"    // 支付失败
	PaymentStatusRefunded  PaymentStatus = "refunded"  // 已退款
	PaymentStatusCancelled PaymentStatus = "cancelled" // 已取消
)

// Payment 支付实体
type Payment struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	OrderID       string         `gorm:"size:50;not null" json:"order_id"`
	UserID        uint           `gorm:"not null" json:"user_id"`
	Amount        int64          `gorm:"not null" json:"amount"` // 金额，单位：分
	Currency      string         `gorm:"size:10;default:'CNY'" json:"currency"`
	Method        string         `gorm:"size:20;not null" json:"method"` // wechat, alipay, etc.
	Status        PaymentStatus  `gorm:"size:20;default:'pending'" json:"status"`
	TransactionID string         `gorm:"size:100" json:"transaction_id"`
	Description   string         `gorm:"size:255" json:"description"`
	WechatPayInfo *WechatPayInfo `gorm:"foreignKey:PaymentID" json:"wechat_pay_info,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	PaidAt        *time.Time     `json:"paid_at,omitempty"`
}

// WechatPayInfo 微信支付信息
type WechatPayInfo struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PaymentID  uint      `gorm:"not null;uniqueIndex" json:"payment_id"`
	AppID      string    `gorm:"size:50" json:"app_id"`
	MCHID      string    `gorm:"size:50" json:"mch_id"`
	NonceStr   string    `gorm:"size:50" json:"nonce_str"`
	PrepayID   string    `gorm:"size:100" json:"prepay_id"`
	CodeURL    string    `gorm:"size:255" json:"code_url"`
	Sign       string    `gorm:"size:255" json:"sign"`
	Timestamp  string    `gorm:"size:20" json:"timestamp"`
	Package    string    `gorm:"size:50" json:"package"`
	PaySign    string    `gorm:"size:255" json:"pay_sign"`
	ReturnCode string    `gorm:"size:20" json:"return_code"`
	ReturnMsg  string    `gorm:"size:255" json:"return_msg"`
	ResultCode string    `gorm:"size:20" json:"result_code"`
	ErrCode    string    `gorm:"size:20" json:"err_code"`
	ErrCodeDes string    `gorm:"size:255" json:"err_code_des"`
	NotifyData string    `gorm:"type:text" json:"notify_data"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PaymentRepository 支付仓库接口
type PaymentRepository interface {
	GetByID(id uint) (*Payment, error)
	GetByOrderID(orderID string) (*Payment, error)
	GetByTransactionID(transactionID string) (*Payment, error)
	Create(payment *Payment) error
	Update(payment *Payment) error
	Delete(id uint) error
	ListByUser(userID uint, page, pageSize int) ([]*Payment, int64, error)
	ListByStatus(status PaymentStatus, page, pageSize int) ([]*Payment, int64, error)
}

// PaymentService 支付领域服务接口
type PaymentService interface {
	CreatePayment(userID uint, orderID string, amount int64, method string, description string) (*Payment, error)
	GetPayment(paymentID uint) (*Payment, error)
	GetPaymentByOrderID(orderID string) (*Payment, error)
	UpdatePaymentStatus(paymentID uint, status PaymentStatus) error
	CreateWechatPayment(userID uint, orderID string, amount int64, description string) (*Payment, error)
	HandleWechatPayNotify(notifyData []byte) (*Payment, error)
	CheckPaymentStatus(paymentID uint) (PaymentStatus, error)
	CancelPayment(paymentID uint) error
	RefundPayment(paymentID uint, refundAmount int64) error
}
