package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"

// RefundRepository 退款单仓库接口
type RefundRepository interface {
	GetByOutRefundNo(outRefundNo string) (*entity.RefundRecord, error)
	Create(record *entity.RefundRecord) error
	Update(record *entity.RefundRecord) error
	List(page, pageSize int) ([]*entity.RefundRecord, int, error)
	ListByOutTradeNo(outTradeNo string) ([]*entity.RefundRecord, error)
}

// MedRefundRepository 医保退款单仓库接口
type MedRefundRepository interface {
	GetByHospOutRefundNo(hospOutRefundNo string) (*entity.MedRefundRecord, error)
	Create(record *entity.MedRefundRecord) error
	Update(record *entity.MedRefundRecord) error
	List(page, pageSize int) ([]*entity.MedRefundRecord, int, error)
	LatestByOutTradeNo(outTradeNo string) (*entity.MedRefundRecord, error)
}
