package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// refundModel 退款单数据库模型（对齐 Gpay model.RefundRecord）
type refundModel struct {
	database.SnowflakeBaseModel
	OutRefundNo         string `gorm:"size:64;not null;uniqueIndex"`
	OutTradeNo          string `gorm:"size:100;not null;index"`
	RefundID            string `gorm:"size:64;index"`
	Reason              string `gorm:"size:255"`
	RefundFee           int64  `gorm:"not null"`
	TotalFee            int64  `gorm:"not null"`
	Status              string `gorm:"size:20;not null;index"`
	Applied             bool   `gorm:"default:false"`
	UserReceivedAccount string `gorm:"size:128"`
	SuccessTime         string `gorm:"size:32"`
	WxLastResponse      string `gorm:"type:text"`
}

// TableName 指定表名（与 migrations 中的 wx_refunds 表对齐）
func (refundModel) TableName() string {
	return "wx_refunds"
}

func (m *refundModel) toDomain() *entity.RefundRecord {
	if m == nil {
		return nil
	}
	return &entity.RefundRecord{
		ID:                  m.ID,
		OutRefundNo:         m.OutRefundNo,
		OutTradeNo:          m.OutTradeNo,
		RefundID:            m.RefundID,
		Reason:              m.Reason,
		RefundFee:           m.RefundFee,
		TotalFee:            m.TotalFee,
		Status:              m.Status,
		Applied:             m.Applied,
		UserReceivedAccount: m.UserReceivedAccount,
		SuccessTime:         m.SuccessTime,
		WxLastResponse:      m.WxLastResponse,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func fromDomainRefund(r *entity.RefundRecord) *refundModel {
	return &refundModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        r.ID,
			CreatedBy: 0,
			UpdatedBy: 0,
		},
		OutRefundNo:         r.OutRefundNo,
		OutTradeNo:          r.OutTradeNo,
		RefundID:            r.RefundID,
		Reason:              r.Reason,
		RefundFee:           r.RefundFee,
		TotalFee:            r.TotalFee,
		Status:              r.Status,
		Applied:             r.Applied,
		UserReceivedAccount: r.UserReceivedAccount,
		SuccessTime:         r.SuccessTime,
		WxLastResponse:      r.WxLastResponse,
	}
}

// medRefundModel 医保退款单数据库模型（对齐 Gpay model.MedRefundRecord）
type medRefundModel struct {
	database.SnowflakeBaseModel
	HospOutRefundNo       string `gorm:"size:64;not null;uniqueIndex"`
	OutTradeNo            string `gorm:"size:100;not null;index"`
	PartRefundType        string `gorm:"size:30"`
	CashRefundFee         int64  `gorm:"default:0"`
	InsRefundFee          int64  `gorm:"default:0"`
	RequestContent        string `gorm:"type:text"`
	Reason                string `gorm:"size:255"`
	MedRefundID           string `gorm:"size:64"`
	CashRefundID          string `gorm:"size:64;index"`
	InsuranceRefundID     string `gorm:"size:64"`
	MedRefundState        string `gorm:"size:20"`
	CashRefundStatus      string `gorm:"size:20"`
	InsuranceRefundStatus string `gorm:"size:20"`
	RefundEndTime         string `gorm:"size:32"`
	ResponseContent       string `gorm:"type:text"`
	CashApplied           bool   `gorm:"default:false"`
	InsApplied            bool   `gorm:"default:false"`
	LastResponse          string `gorm:"type:text"`
}

// TableName 指定表名（与 migrations 中的 wx_med_refunds 表对齐）
func (medRefundModel) TableName() string {
	return "wx_med_refunds"
}

func (m *medRefundModel) toDomain() *entity.MedRefundRecord {
	if m == nil {
		return nil
	}
	return &entity.MedRefundRecord{
		ID:                    m.ID,
		HospOutRefundNo:       m.HospOutRefundNo,
		OutTradeNo:            m.OutTradeNo,
		PartRefundType:        m.PartRefundType,
		CashRefundFee:         m.CashRefundFee,
		InsRefundFee:          m.InsRefundFee,
		RequestContent:        m.RequestContent,
		Reason:                m.Reason,
		MedRefundID:           m.MedRefundID,
		CashRefundID:          m.CashRefundID,
		InsuranceRefundID:     m.InsuranceRefundID,
		MedRefundState:        m.MedRefundState,
		CashRefundStatus:      m.CashRefundStatus,
		InsuranceRefundStatus: m.InsuranceRefundStatus,
		RefundEndTime:         m.RefundEndTime,
		ResponseContent:       m.ResponseContent,
		CashApplied:           m.CashApplied,
		InsApplied:            m.InsApplied,
		LastResponse:          m.LastResponse,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}

func fromDomainMedRefund(r *entity.MedRefundRecord) *medRefundModel {
	return &medRefundModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        r.ID,
			CreatedBy: 0,
			UpdatedBy: 0,
		},
		HospOutRefundNo:       r.HospOutRefundNo,
		OutTradeNo:            r.OutTradeNo,
		PartRefundType:        r.PartRefundType,
		CashRefundFee:         r.CashRefundFee,
		InsRefundFee:          r.InsRefundFee,
		RequestContent:        r.RequestContent,
		Reason:                r.Reason,
		MedRefundID:           r.MedRefundID,
		CashRefundID:          r.CashRefundID,
		InsuranceRefundID:     r.InsuranceRefundID,
		MedRefundState:        r.MedRefundState,
		CashRefundStatus:      r.CashRefundStatus,
		InsuranceRefundStatus: r.InsuranceRefundStatus,
		RefundEndTime:         r.RefundEndTime,
		ResponseContent:       r.ResponseContent,
		CashApplied:           r.CashApplied,
		InsApplied:            r.InsApplied,
		LastResponse:          r.LastResponse,
	}
}

// refundRepository 退款单仓储实现
type refundRepository struct {
	db *database.PostgresDB
}

var _ repo.RefundRepository = (*refundRepository)(nil)

// NewRefundRepository 创建退款单仓储实现
func NewRefundRepository(db *database.PostgresDB) repo.RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) GetByOutRefundNo(outRefundNo string) (*entity.RefundRecord, error) {
	var m refundModel
	if err := r.db.Where("out_refund_no = ?", outRefundNo).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *refundRepository) Create(record *entity.RefundRecord) error {
	m := fromDomainRefund(record)
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	record.ID = m.ID
	record.CreatedAt = m.CreatedAt
	record.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *refundRepository) Update(record *entity.RefundRecord) error {
	// Omit 审计列：更新业务字段时不覆盖 created_at/created_by 等创建审计信息
	return r.db.Omit("created_at", "created_by", "updated_by", "deleted_by", "deleted_at").
		Save(fromDomainRefund(record)).Error
}

func (r *refundRepository) List(page, pageSize int) ([]*entity.RefundRecord, int, error) {
	var total int64
	var models []refundModel
	query := r.db.Model(&refundModel{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	records := make([]*entity.RefundRecord, len(models))
	for i := range models {
		records[i] = models[i].toDomain()
	}
	return records, int(total), nil
}

func (r *refundRepository) ListByOutTradeNo(outTradeNo string) ([]*entity.RefundRecord, error) {
	var models []refundModel
	if err := r.db.Where("out_trade_no = ?", outTradeNo).Order("id DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	records := make([]*entity.RefundRecord, len(models))
	for i := range models {
		records[i] = models[i].toDomain()
	}
	return records, nil
}

// medRefundRepository 医保退款单仓储实现
type medRefundRepository struct {
	db *database.PostgresDB
}

var _ repo.MedRefundRepository = (*medRefundRepository)(nil)

// NewMedRefundRepository 创建医保退款单仓储实现
func NewMedRefundRepository(db *database.PostgresDB) repo.MedRefundRepository {
	return &medRefundRepository{db: db}
}

func (r *medRefundRepository) GetByHospOutRefundNo(hospOutRefundNo string) (*entity.MedRefundRecord, error) {
	var m medRefundModel
	if err := r.db.Where("hosp_out_refund_no = ?", hospOutRefundNo).First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *medRefundRepository) Create(record *entity.MedRefundRecord) error {
	m := fromDomainMedRefund(record)
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	record.ID = m.ID
	record.CreatedAt = m.CreatedAt
	record.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *medRefundRepository) Update(record *entity.MedRefundRecord) error {
	// Omit 审计列：更新业务字段时不覆盖 created_at/created_by 等创建审计信息
	return r.db.Omit("created_at", "created_by", "updated_by", "deleted_by", "deleted_at").
		Save(fromDomainMedRefund(record)).Error
}

func (r *medRefundRepository) List(page, pageSize int) ([]*entity.MedRefundRecord, int, error) {
	var total int64
	var models []medRefundModel
	query := r.db.Model(&medRefundModel{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	records := make([]*entity.MedRefundRecord, len(models))
	for i := range models {
		records[i] = models[i].toDomain()
	}
	return records, int(total), nil
}

func (r *medRefundRepository) LatestByOutTradeNo(outTradeNo string) (*entity.MedRefundRecord, error) {
	var m medRefundModel
	if err := r.db.Where("out_trade_no = ?", outTradeNo).Order("id DESC").First(&m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}
