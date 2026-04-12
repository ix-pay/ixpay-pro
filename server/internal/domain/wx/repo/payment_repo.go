package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"

// PaymentRepository 支付仓库接口
// 定义支付数据访问的抽象接口
type PaymentRepository interface {
	GetByID(id string) (*entity.Payment, error)
	GetByOrderID(orderID string) (*entity.Payment, error)
	GetByTransactionID(transactionID string) (*entity.Payment, error)
	Create(payment *entity.Payment) error
	Update(payment *entity.Payment) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.Payment, int, error)
	ListByUser(userID string, page, pageSize int) ([]*entity.Payment, int, error)
	ListByStatus(status entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int, error)
}
