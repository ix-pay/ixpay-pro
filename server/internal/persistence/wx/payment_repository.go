package persistence

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// paymentModel 支付数据库模型
type paymentModel struct {
	database.SnowflakeBaseModel
	OrderID       string              `gorm:"size:50;not null;index"`
	UserID        *int64              `gorm:"not null;default:0;index"`
	Amount        *int64              `gorm:"not null;default:0"`
	Currency      string              `gorm:"size:10;default:'CNY'"`
	Method        string              `gorm:"size:20;not null"`
	Status        string              `gorm:"size:20;default:'pending'"`
	TransactionID string              `gorm:"size:100;index"`
	Description   string              `gorm:"size:255"`
	WechatPayInfo *wechatPayInfoModel `gorm:"foreignKey:PaymentID"`
	PaidAt        *time.Time          `gorm:"index"`
}

// TableName 指定表名
func (paymentModel) TableName() string {
	return "wx_payments"
}

// toDomain 将数据库模型转换为领域实体
func (m *paymentModel) toDomain() *entity.Payment {
	if m == nil {
		return nil
	}
	var wechatPayInfo *entity.WechatPayInfo
	if m.WechatPayInfo != nil {
		wechatPayInfo = m.WechatPayInfo.toDomain()
	}
	return &entity.Payment{
		ID:            m.ID,
		OrderID:       m.OrderID,
		UserID:        *m.UserID,
		Amount:        *m.Amount,
		Currency:      m.Currency,
		Method:        m.Method,
		Status:        entity.PaymentStatus(m.Status),
		TransactionID: m.TransactionID,
		Description:   m.Description,
		WechatPayInfo: wechatPayInfo,
		PaidAt:        m.PaidAt,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
		UpdatedBy:     m.UpdatedBy,
		UpdatedAt:     m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainPayment(payment *entity.Payment) (*paymentModel, error) {
	var wechatPayInfo *wechatPayInfoModel
	if payment.WechatPayInfo != nil {
		wechatPayInfo = &wechatPayInfoModel{
			SnowflakeBaseModel: database.SnowflakeBaseModel{
				ID:        payment.WechatPayInfo.ID,
				CreatedBy: payment.WechatPayInfo.CreatedBy,
				UpdatedBy: payment.WechatPayInfo.UpdatedBy,
			},
			PaymentID:  common.Int64Ptr(payment.WechatPayInfo.PaymentID),
			AppID:      payment.WechatPayInfo.AppID,
			MCHID:      payment.WechatPayInfo.MCHID,
			NonceStr:   payment.WechatPayInfo.NonceStr,
			PrepayID:   payment.WechatPayInfo.PrepayID,
			CodeURL:    payment.WechatPayInfo.CodeURL,
			Sign:       payment.WechatPayInfo.Sign,
			Timestamp:  payment.WechatPayInfo.Timestamp,
			Package:    payment.WechatPayInfo.Package,
			PaySign:    payment.WechatPayInfo.PaySign,
			ReturnCode: payment.WechatPayInfo.ReturnCode,
			ReturnMsg:  payment.WechatPayInfo.ReturnMsg,
			ResultCode: payment.WechatPayInfo.ResultCode,
			ErrCode:    payment.WechatPayInfo.ErrCode,
			ErrCodeDes: payment.WechatPayInfo.ErrCodeDes,
			NotifyData: payment.WechatPayInfo.NotifyData,
		}
	}

	return &paymentModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        payment.ID,
			CreatedBy: 0,
			UpdatedBy: 0,
		},
		OrderID:       payment.OrderID,
		UserID:        common.Int64Ptr(payment.UserID),
		Amount:        common.Int64Ptr(payment.Amount),
		Currency:      payment.Currency,
		Method:        payment.Method,
		Status:        string(payment.Status),
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		WechatPayInfo: wechatPayInfo,
		PaidAt:        payment.PaidAt,
	}, nil
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
	dbModel, err := fromDomainPayment(payment)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	payment.ID = dbModel.ID
	return nil
}

// Update 更新支付记录
func (r *paymentRepository) Update(payment *entity.Payment) error {
	dbModel, err := fromDomainPayment(payment)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除支付记录
func (r *paymentRepository) Delete(id int64) error {
	return r.db.Delete(&paymentModel{}, id).Error
}

// ListByUser 根据用户 ID 查询支付列表
func (r *paymentRepository) ListByUser(userID int64, page, pageSize int) ([]*entity.Payment, int, error) {
	var total64 int64
	var dbModels []paymentModel

	query := r.db.Model(&paymentModel{}).Where("user_id = ?", userID)

	if err := query.Count(&total64).Error; err != nil {
		return nil, 0, err
	}
	total := int(total64)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	payments := make([]*entity.Payment, len(dbModels))
	for i, model := range dbModels {
		payments[i] = model.toDomain()
	}

	return payments, total, nil
}

// ListByStatus 根据状态查询支付列表
func (r *paymentRepository) ListByStatus(status entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int, error) {
	var total64 int64
	var dbModels []paymentModel

	query := r.db.Model(&paymentModel{}).Where("status = ?", string(status))

	if err := query.Count(&total64).Error; err != nil {
		return nil, 0, err
	}
	total := int(total64)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	payments := make([]*entity.Payment, len(dbModels))
	for i, model := range dbModels {
		payments[i] = model.toDomain()
	}

	return payments, total, nil
}

// List 查询支付列表（支持过滤）
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
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	payments := make([]*entity.Payment, len(dbModels))
	for i, model := range dbModels {
		payments[i] = model.toDomain()
	}

	return payments, total, nil
}
