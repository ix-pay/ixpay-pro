package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// wechatPayInfoModel 微信支付信息数据库模型
type wechatPayInfoModel struct {
	database.SnowflakeBaseModel
	PaymentID     int64  `gorm:"not null;uniqueIndex"`
	AppID         string `gorm:"size:100"`
	MCHID         string `gorm:"size:50"`
	NonceStr      string `gorm:"size:64"`
	PrepayID      string `gorm:"size:128"`
	CodeURL       string `gorm:"size:256"`
	Sign          string `gorm:"size:64"`
	Timestamp     string `gorm:"size:20"`
	Package       string `gorm:"size:256"`
	PaySign       string `gorm:"size:64"`
	ReturnCode    string `gorm:"size:20"`
	ReturnMsg     string `gorm:"size:256"`
	ResultCode    string `gorm:"size:20"`
	ErrCode       string `gorm:"size:64"`
	ErrCodeDes    string `gorm:"size:256"`
	TransactionID string `gorm:"size:100;index"`
	OpenID        string `gorm:"size:100"`
	BankType      string `gorm:"size:50"`
	TotalFee      int64  `gorm:"default:0"`
	CashFee       int64  `gorm:"default:0"`
	FeeType       string `gorm:"size:10"`
	NotifyData    string `gorm:"type:text"`
}

// TableName 指定表名
func (wechatPayInfoModel) TableName() string {
	return "wechat_pay_infos"
}

// toDomain 将数据库模型转换为领域实体
func (m *wechatPayInfoModel) toDomain() *entity.WechatPayInfo {
	if m == nil {
		return nil
	}
	return &entity.WechatPayInfo{
		ID:            m.ID,
		PaymentID:     m.PaymentID,
		AppID:         m.AppID,
		MCHID:         m.MCHID,
		NonceStr:      m.NonceStr,
		PrepayID:      m.PrepayID,
		CodeURL:       m.CodeURL,
		Sign:          m.Sign,
		Timestamp:     m.Timestamp,
		Package:       m.Package,
		PaySign:       m.PaySign,
		ReturnCode:    m.ReturnCode,
		ReturnMsg:     m.ReturnMsg,
		ResultCode:    m.ResultCode,
		ErrCode:       m.ErrCode,
		ErrCodeDes:    m.ErrCodeDes,
		TransactionID: m.TransactionID,
		OpenID:        m.OpenID,
		BankType:      m.BankType,
		TotalFee:      m.TotalFee,
		CashFee:       m.CashFee,
		FeeType:       m.FeeType,
		NotifyData:    m.NotifyData,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
		UpdatedBy:     m.UpdatedBy,
		UpdatedAt:     m.UpdatedAt,
	}
}

// fromDomainWechatPayInfo 将领域实体转换为数据库模型
func fromDomainWechatPayInfo(info *entity.WechatPayInfo) *wechatPayInfoModel {
	return &wechatPayInfoModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        info.ID,
			CreatedBy: info.CreatedBy,
			UpdatedBy: info.UpdatedBy,
		},
		PaymentID:     info.PaymentID,
		AppID:         info.AppID,
		MCHID:         info.MCHID,
		NonceStr:      info.NonceStr,
		PrepayID:      info.PrepayID,
		CodeURL:       info.CodeURL,
		Sign:          info.Sign,
		Timestamp:     info.Timestamp,
		Package:       info.Package,
		PaySign:       info.PaySign,
		ReturnCode:    info.ReturnCode,
		ReturnMsg:     info.ReturnMsg,
		ResultCode:    info.ResultCode,
		ErrCode:       info.ErrCode,
		ErrCodeDes:    info.ErrCodeDes,
		TransactionID: info.TransactionID,
		OpenID:        info.OpenID,
		BankType:      info.BankType,
		TotalFee:      info.TotalFee,
		CashFee:       info.CashFee,
		FeeType:       info.FeeType,
		NotifyData:    info.NotifyData,
	}
}

// wechatPayInfoRepository Repository 实现
type wechatPayInfoRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.WechatPayInfoRepository = (*wechatPayInfoRepository)(nil)

// NewWechatPayInfoRepository 创建微信支付信息仓库实现
func NewWechatPayInfoRepository(db *database.PostgresDB) repo.WechatPayInfoRepository {
	return &wechatPayInfoRepository{db: db}
}

// GetByID 根据 ID 查询
func (r *wechatPayInfoRepository) GetByID(id int64) (*entity.WechatPayInfo, error) {
	var dbModel wechatPayInfoModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

// GetByOrderID 根据订单 ID 查询（通过 payment 关联）
func (r *wechatPayInfoRepository) GetByOrderID(orderID string) (*entity.WechatPayInfo, error) {
	var dbModel wechatPayInfoModel
	result := r.db.Joins("JOIN payments ON payments.id = wechat_pay_infos.payment_id").
		Where("payments.order_id = ?", orderID).
		First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

// GetByTransactionID 根据交易 ID 查询
func (r *wechatPayInfoRepository) GetByTransactionID(transactionID string) (*entity.WechatPayInfo, error) {
	var dbModel wechatPayInfoModel
	result := r.db.Where("transaction_id = ?", transactionID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

// Create 创建微信支付信息
func (r *wechatPayInfoRepository) Create(info *entity.WechatPayInfo) error {
	dbModel := fromDomainWechatPayInfo(info)
	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}
	info.ID = dbModel.ID
	info.CreatedAt = dbModel.CreatedAt
	info.UpdatedAt = dbModel.UpdatedAt
	return nil
}

// Update 更新微信支付信息
func (r *wechatPayInfoRepository) Update(info *entity.WechatPayInfo) error {
	dbModel := fromDomainWechatPayInfo(info)
	return r.db.Save(dbModel).Error
}

// Delete 删除微信支付信息
func (r *wechatPayInfoRepository) Delete(id int64) error {
	return r.db.Delete(&wechatPayInfoModel{}, id).Error
}

// List 分页查询
func (r *wechatPayInfoRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.WechatPayInfo, int, error) {
	var total64 int64
	var dbModels []wechatPayInfoModel

	query := r.db.Model(&wechatPayInfoModel{})
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

	infos := make([]*entity.WechatPayInfo, len(dbModels))
	for i, model := range dbModels {
		infos[i] = model.toDomain()
	}
	return infos, total, nil
}

// ListByUser 根据用户 ID 分页查询
func (r *wechatPayInfoRepository) ListByUser(userID int64, page, pageSize int) ([]*entity.WechatPayInfo, int, error) {
	return r.List(page, pageSize, map[string]interface{}{"user_id": userID})
}

// ListByStatus 分页查询（按状态过滤）- 通过 payment 表关联
func (r *wechatPayInfoRepository) ListByStatus(page, pageSize int) ([]*entity.WechatPayInfo, int, error) {
	return r.List(page, pageSize, nil)
}