package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"gorm.io/gorm"
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
	// 扩展字段（普通/医保支付共用，对齐 Gpay model.Order）
	OrderType      string `gorm:"size:20;default:jsapi;index"`
	RefundedFee    int64  `gorm:"default:0"`
	Openid         string `gorm:"size:100;index"`
	CodeURL        string `gorm:"size:255"`
	TradeState     string `gorm:"size:30"`
	MixTradeNo     string `gorm:"size:100;index"`
	MedInfo        string `gorm:"type:text"` // 医保订单信息 JSON
	WxLastResponse string `gorm:"type:text"` // 微信侧原始回执
}

// TableName 指定表名（与 migrations 中的 wx_payments 表对齐）
func (paymentModel) TableName() string {
	return "wx_payments"
}

// toDomain 将数据库模型转换为领域实体
func (m *paymentModel) toDomain() *entity.Payment {
	if m == nil {
		return nil
	}
	payment := &entity.Payment{
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
		// 扩展字段
		OrderType:      m.OrderType,
		RefundedFee:    m.RefundedFee,
		Openid:         m.Openid,
		CodeURL:        m.CodeURL,
		TradeState:     m.TradeState,
		MixTradeNo:     m.MixTradeNo,
		WxLastResponse: m.WxLastResponse,
	}
	// 医保订单信息 JSON 反序列化
	if m.MedInfo != "" {
		med := &entity.MedOrderInfo{}
		if err := json.Unmarshal([]byte(m.MedInfo), med); err == nil {
			payment.MedInfo = med
		}
	}
	return payment
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainPayment(payment *entity.Payment) *paymentModel {
	model := &paymentModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        payment.ID,
			CreatedBy: payment.CreatedBy,
			UpdatedBy: payment.UpdatedBy,
		},
		OrderID:        payment.OrderID,
		UserID:         payment.UserID,
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		Method:         payment.Method,
		Status:         string(payment.Status),
		TransactionID:  payment.TransactionID,
		Description:    payment.Description,
		PaidAt:         payment.PaidAt,
		OrderType:      payment.OrderType,
		RefundedFee:    payment.RefundedFee,
		Openid:         payment.Openid,
		CodeURL:        payment.CodeURL,
		TradeState:     payment.TradeState,
		MixTradeNo:     payment.MixTradeNo,
		WxLastResponse: payment.WxLastResponse,
	}
	// 医保订单信息 JSON 序列化
	if payment.MedInfo != nil {
		if raw, err := json.Marshal(payment.MedInfo); err == nil {
			model.MedInfo = string(raw)
		}
	}
	return model
}

// paymentRepository Repository 实现（方法按 ctx 租户路由：业务数据归租户库）
type paymentRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.PaymentRepository = (*paymentRepository)(nil)

// g 按上下文租户路由 GORM 实例（无租户上下文/默认租户 → 主库）
func (r *paymentRepository) g(ctx context.Context) *gorm.DB {
	pdb := r.db
	if ctx != nil {
		if gdb, err := pdb.TenantGorm(ctx); err == nil && gdb != nil {
			return gdb
		}
	}
	return pdb.DB
}

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

// GetByMixTradeNo 根据医保混合订单号查询支付记录
func (r *paymentRepository) GetByMixTradeNo(mixTradeNo string) (*entity.Payment, error) {
	var dbModel paymentModel
	result := r.db.Where("mix_trade_no = ?", mixTradeNo).First(&dbModel)
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
	// Omit 审计列：更新业务字段时不覆盖 created_at/created_by 等创建审计信息
	return r.db.Omit("created_at", "created_by", "updated_by", "deleted_by", "deleted_at").
		Save(dbModel).Error
}

// Delete 删除支付记录
func (r *paymentRepository) Delete(id int64) error {
	return r.db.Delete(&paymentModel{}, id).Error
}

// List 分页查询支付记录列表（ctx 决定查询哪个租户库）
func (r *paymentRepository) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*entity.Payment, int, error) {
	var total64 int64
	var dbModels []paymentModel

	query := r.g(ctx).Model(&paymentModel{})

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

// GetReconSummary 对账汇总统计（管理平台对账模块，时间范围按 created_at；ctx 决定查询哪个租户库）
func (r *paymentRepository) GetReconSummary(ctx context.Context, start, end time.Time) (*repo.ReconSummary, error) {
	var out repo.ReconSummary
	err := r.g(ctx).Raw(`
		SELECT
			COUNT(*) AS total_count,
			COUNT(*) FILTER (WHERE status = 'success') AS success_count,
			COALESCE(SUM(amount) FILTER (WHERE status = 'success'), 0) AS success_amount,
			COUNT(*) FILTER (WHERE status = 'refunded') AS refund_count,
			COALESCE(SUM(refunded_fee) FILTER (WHERE status IN ('refunded', 'success')), 0) AS refund_amount,
			COUNT(*) FILTER (WHERE status = 'pending') AS pending_count
		FROM wx_payments
		WHERE created_at >= ? AND created_at <= ?
	`, start, end).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListByUser 根据用户 ID 分页查询支付记录
func (r *paymentRepository) ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]*entity.Payment, int, error) {
	return r.List(ctx, page, pageSize, map[string]interface{}{"user_id": userID})
}

// ListByStatus 根据状态分页查询支付记录
func (r *paymentRepository) ListByStatus(ctx context.Context, status entity.PaymentStatus, page, pageSize int) ([]*entity.Payment, int, error) {
	return r.List(ctx, page, pageSize, map[string]interface{}{"status": string(status)})
}
