package repository

import (
	wxmodel "github.com/ix-pay/ixpay-pro/internal/app/wx/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/database"
)

// PaymentRepository 实现支付仓库接口
type PaymentRepository struct {
	db *database.PostgresDB
}

// NewPaymentRepository 创建支付仓库实例
func NewPaymentRepository(db *database.PostgresDB) wxmodel.PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}

// GetByID 根据ID获取支付记录
func (r *PaymentRepository) GetByID(id uint) (*wxmodel.Payment, error) {
	var payment wxmodel.Payment
	result := r.db.Preload("WechatPayInfo").First(&payment, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &payment, nil
}

// GetByOrderID 根据订单ID获取支付记录
func (r *PaymentRepository) GetByOrderID(orderID string) (*wxmodel.Payment, error) {
	var payment wxmodel.Payment
	result := r.db.Preload("WechatPayInfo").Where("order_id = ?", orderID).First(&payment)
	if result.Error != nil {
		return nil, result.Error
	}
	return &payment, nil
}

// GetByTransactionID 根据交易ID获取支付记录
func (r *PaymentRepository) GetByTransactionID(transactionID string) (*wxmodel.Payment, error) {
	var payment wxmodel.Payment
	result := r.db.Preload("WechatPayInfo").Where("transaction_id = ?", transactionID).First(&payment)
	if result.Error != nil {
		return nil, result.Error
	}
	return &payment, nil
}

// Create 创建支付记录
func (r *PaymentRepository) Create(payment *wxmodel.Payment) error {
	return r.db.Create(payment).Error
}

// Update 更新支付记录
func (r *PaymentRepository) Update(payment *wxmodel.Payment) error {
	return r.db.Save(payment).Error
}

// Delete 删除支付记录
func (r *PaymentRepository) Delete(id uint) error {
	// 先删除相关的微信支付信息
	if err := r.db.Where("payment_id = ?", id).Delete(&wxmodel.WechatPayInfo{}).Error; err != nil {
		return err
	}
	// 再删除支付记录
	return r.db.Delete(&wxmodel.Payment{}, id).Error
}

// ListByUser 获取用户的支付记录列表
func (r *PaymentRepository) ListByUser(userID uint, page, pageSize int) ([]*wxmodel.Payment, int64, error) {
	var payments []*wxmodel.Payment
	var total int64

	query := r.db.Model(&wxmodel.Payment{}).Where("user_id = ?", userID).Preload("WechatPayInfo")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// ListByStatus 获取指定状态的支付记录列表
func (r *PaymentRepository) ListByStatus(status wxmodel.PaymentStatus, page, pageSize int) ([]*wxmodel.Payment, int64, error) {
	var payments []*wxmodel.Payment
	var total int64

	query := r.db.Model(&wxmodel.Payment{}).Where("status = ?", status).Preload("WechatPayInfo")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}
