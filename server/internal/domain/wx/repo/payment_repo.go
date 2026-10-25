package repo

import (
	"context"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
)

// ReconSummary 对账汇总统计结果（管理平台对账模块）
type ReconSummary struct {
	TotalCount    int64 // 支付单总数
	SuccessCount  int64 // 支付成功笔数
	SuccessAmount int64 // 支付成功金额(分)
	RefundCount   int64 // 已退款笔数
	RefundAmount  int64 // 累计退款金额(分)
	PendingCount  int64 // 待支付笔数
}

// PaymentRepository 支付仓库接口
// 定义支付数据访问的抽象接口
type PaymentRepository interface {
	GetByID(id int64) (*entity.Payment, error)
	GetByOrderID(orderID string) (*entity.Payment, error)
	GetByTransactionID(transactionID string) (*entity.Payment, error)
	GetByMixTradeNo(mixTradeNo string) (*entity.Payment, error)
	Create(payment *entity.Payment) error
	Update(payment *entity.Payment) error
	Delete(id int64) error
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*entity.Payment, int, error)
	ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]*entity.Payment, int, error)
	ListByStatus(ctx context.Context, status entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int, error)
	GetReconSummary(ctx context.Context, start, end time.Time) (*ReconSummary, error)
}
