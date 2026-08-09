package persistence

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// paymentModel 支付数据库模型
type paymentModel struct {
	database.SnowflakeBaseModel
	OrderID       string     `gorm:"size:100;not null;uniqueIndex"`
	UserID        int64      `gorm:"not null;index"`
	Amount        int64      `gorm:"not null"`
	Currency      string     `gorm:"size:10;not null;default:CNY"`
	Method        string     `gorm:"size:50;not null"`
	Status        string     `gorm:"size:20;not null;index"`
	TransactionID string     `gorm:"size:100;index"`
	Description   string     `gorm:"size:500"`
	PaidAt        *time.Time `gorm:"index"`
}

// TableName 指定表名
func (paymentModel) TableName() string {
	return "payments"
}

// toDomain 将数据库模型转换为领域实体
func (m *paymentModel) toDomain() *entity.Payment {
	if m == nil {
		return nil
	}
	return &entity.Payment{
		ID:            m.ID,
		OrderID:       m.OrderID,
		UserID:        m.UserID,
		Amount:        m.Amount,
		Currency:      m.Currency,
		Method:        m.Method,
		Status:        entity.PaymentStatus(m.Status),
		TransactionID: m.TransactionID,
		Description:   m.Description,
		PaidAt:        m.PaidAt,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
		UpdatedBy:     m.UpdatedBy,
		UpdatedAt:     m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainPayment(payment *entity.Payment) *paymentModel {
	return &paymentModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        payment.ID,
			CreatedBy: payment.CreatedBy,
			UpdatedBy: payment.UpdatedBy,
		},
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Method:        payment.Method,
		Status:        string(payment.Status),
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		PaidAt:        payment.PaidAt,
	}
}

// paymentRepository Repository 实现
type paymentRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.PaymentRepository = (*paymentRepository)(nil)

// NewPaymentRepository 创建支付仓库实现
func NewPaymentRepository(db *database.PostgresDB) repo.PaymentRepository {
	return &paymentRepository{db: db}
}

// GetByID 根据 ID 查询支付记录
func (r *paymentRepository) GetByID(id int64) (*entity.Payment, error) {
	var dbModel paymentModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

// GetByOrderID 根据订单 ID 查询支付记录
func (r *paymentRepository) GetByOrderID(orderID string) (*entity.Payment, error) {
	var dbModel paymentModel
	result := r.db.Where("order_id = ?", orderID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

// GetByTransactionID 根据交易 ID 查询支付记录
func (r *paymentRepository) GetByTransactionID(transactionID string) (*entity.Payment, error) {
	var dbModel paymentModel
	result := r.db.Where("transaction_id = ?", transactionID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

// Create 创建支付记录
func (r *paymentRepository) Create(payment *entity.Payment) error {
	dbModel := fromDomainPayment(payment)
	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}
	payment.ID = dbModel.ID
	payment.CreatedAt = dbModel.CreatedAt
	payment.UpdatedAt = dbModel.UpdatedAt
	return nil
}

// Update 更新支付记录
func (r *paymentRepository) Update(payment *entity.Payment) error {
	dbModel := fromDomainPayment(payment)
	return r.db.Save(dbModel).Error
}

// Delete 删除支付记录
func (r *paymentRepository) Delete(id int64) error {
	return r.db.Delete(&paymentModel{}, id).Error
}

// List 分页查询支付记录列表
func (r *paymentRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Payment, int, error) {
	var total64 int64
	var dbModels []paymentModel

	query := r.db.Model(&paymentModel{})

	// 应用过滤器
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total64).Error; err != nil {
		return nil, 0, err
	}
	total := int(total64)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	payments := make([]*entity.Payment, len(dbModels))
	for i, model := range dbModels {
		payments[i] = model.toDomain()
	}

	return payments, total, nil
}

// ListByUser 根据用户 ID 分页查询支付记录
func (r *paymentRepository) ListByUser(userID int64, page, pageSize int) ([]*entity.Payment, int, error) {
	return r.List(page, pageSize, map[string]interface{}{"user_id": userID})
}

// ListByStatus 根据状态分页查询支付记录
func (r *paymentRepository) ListByStatus(status entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int, error) {
	return r.List(page, pageSize, map[string]interface{}{"status": string(status)})
}