package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/wx/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// wechatPayInfoModel 微信支付信息数据库模型
type wechatPayInfoModel struct {
	database.SnowflakeBaseModel
	PaymentID  int64  `gorm:"not null;uniqueIndex"`
	AppID      string `gorm:"size:50"`
	MCHID      string `gorm:"size:50"`
	NonceStr   string `gorm:"size:50"`
	PrepayID   string `gorm:"size:100"`
	CodeURL    string `gorm:"size:255"`
	Sign       string `gorm:"size:255"`
	Timestamp  string `gorm:"size:20"`
	Package    string `gorm:"size:50"`
	PaySign    string `gorm:"size:255"`
	ReturnCode string `gorm:"size:20"`
	ReturnMsg  string `gorm:"size:255"`
	ResultCode string `gorm:"size:20"`
	ErrCode    string `gorm:"size:20"`
	ErrCodeDes string `gorm:"size:255"`
	NotifyData string `gorm:"type:text"`
}

// TableName 指定表名
func (wechatPayInfoModel) TableName() string {
	return "wx_wechat_pay_infos"
}

// toDomain 将数据库模型转换为领域实体
func (m *wechatPayInfoModel) toDomain() *entity.WechatPayInfo {
	if m == nil {
		return nil
	}
	return &entity.WechatPayInfo{
		ID:         common.ToString(m.ID),
		PaymentID:  common.ToString(m.PaymentID),
		AppID:      m.AppID,
		MCHID:      m.MCHID,
		NonceStr:   m.NonceStr,
		PrepayID:   m.PrepayID,
		CodeURL:    m.CodeURL,
		Sign:       m.Sign,
		Timestamp:  m.Timestamp,
		Package:    m.Package,
		PaySign:    m.PaySign,
		ReturnCode: m.ReturnCode,
		ReturnMsg:  m.ReturnMsg,
		ResultCode: m.ResultCode,
		ErrCode:    m.ErrCode,
		ErrCodeDes: m.ErrCodeDes,
		NotifyData: m.NotifyData,
		CreatedBy:  common.ToString(m.CreatedBy),
		CreatedAt:  m.CreatedAt,
		UpdatedBy:  common.ToString(m.UpdatedBy),
		UpdatedAt:  m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainWechatPayInfo(info *entity.WechatPayInfo) (*wechatPayInfoModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(info.ID, info.CreatedBy, info.UpdatedBy)

	wechatPayInfoID := common.TryParseInt64(info.PaymentID)

	return &wechatPayInfoModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		PaymentID:  wechatPayInfoID,
		AppID:      info.AppID,
		MCHID:      info.MCHID,
		NonceStr:   info.NonceStr,
		PrepayID:   info.PrepayID,
		CodeURL:    info.CodeURL,
		Sign:       info.Sign,
		Timestamp:  info.Timestamp,
		Package:    info.Package,
		PaySign:    info.PaySign,
		ReturnCode: info.ReturnCode,
		ReturnMsg:  info.ReturnMsg,
		ResultCode: info.ResultCode,
		ErrCode:    info.ErrCode,
		ErrCodeDes: info.ErrCodeDes,
		NotifyData: info.NotifyData,
	}, nil
}

// wechatPayInfoRepository Repository 实现
type wechatPayInfoRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.WechatPayInfoRepository = (*wechatPayInfoRepository)(nil)

// NewWechatPayInfoRepository 创建支付仓库实现
func NewWechatPayInfoRepository(db *database.PostgresDB) repo.WechatPayInfoRepository {
	return &wechatPayInfoRepository{db: db}
}

// GetByID 根据 ID 查询支付记录
func (r *wechatPayInfoRepository) GetByID(id string) (*entity.WechatPayInfo, error) {
	idInt, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}
	var dbModel wechatPayInfoModel
	result := r.db.Where("id = ?", idInt).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByOrderID 根据订单 ID 查询支付记录
func (r *wechatPayInfoRepository) GetByOrderID(orderID string) (*entity.WechatPayInfo, error) {
	var dbModel wechatPayInfoModel
	result := r.db.Where("order_id = ?", orderID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetByTransactionID 根据交易 ID 查询支付记录
func (r *wechatPayInfoRepository) GetByTransactionID(transactionID string) (*entity.WechatPayInfo, error) {
	var dbModel wechatPayInfoModel
	result := r.db.Where("transaction_id = ?", transactionID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建支付记录
func (r *wechatPayInfoRepository) Create(WechatPayInfo *entity.WechatPayInfo) error {
	dbModel, err := fromDomainWechatPayInfo(WechatPayInfo)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	WechatPayInfo.ID = common.ToString(dbModel.ID)
	return nil
}

// Update 更新支付记录
func (r *wechatPayInfoRepository) Update(WechatPayInfo *entity.WechatPayInfo) error {
	dbModel, err := fromDomainWechatPayInfo(WechatPayInfo)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除支付记录
func (r *wechatPayInfoRepository) Delete(id string) error {
	idInt, err := common.ParseInt64(id)
	if err != nil {
		return err
	}
	return r.db.Delete(&wechatPayInfoModel{}, idInt).Error
}

// ListByUser 根据用户 ID 查询支付列表
func (r *wechatPayInfoRepository) ListByUser(userID string, page, pageSize int) ([]*entity.WechatPayInfo, int, error) {
	var total64 int64
	var dbModels []wechatPayInfoModel

	userIDint, err := common.ParseInt64(userID)
	if err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&wechatPayInfoModel{}).Where("user_id = ?", userIDint)

	if err := query.Count(&total64).Error; err != nil {
		return nil, 0, err
	}
	total := int(total64)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	wechatPayInfos := make([]*entity.WechatPayInfo, len(dbModels))
	for i, model := range dbModels {
		wechatPayInfos[i] = model.toDomain()
	}

	return wechatPayInfos, total, nil
}

// ListByStatus 根据状态查询支付列表
func (r *wechatPayInfoRepository) ListByStatus(page, pageSize int) ([]*entity.WechatPayInfo, int, error) {
	var total64 int64
	var dbModels []wechatPayInfoModel

	query := r.db.Model(&wechatPayInfoModel{})

	if err := query.Count(&total64).Error; err != nil {
		return nil, 0, err
	}
	total := int(total64)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	wechatPayInfos := make([]*entity.WechatPayInfo, len(dbModels))
	for i, model := range dbModels {
		wechatPayInfos[i] = model.toDomain()
	}

	return wechatPayInfos, total, nil
}

// List 查询支付列表（支持过滤）
func (r *wechatPayInfoRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.WechatPayInfo, int, error) {
	var total64 int64
	var dbModels []wechatPayInfoModel

	query := r.db.Model(&wechatPayInfoModel{})

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

	wechatPayInfos := make([]*entity.WechatPayInfo, len(dbModels))
	for i, model := range dbModels {
		wechatPayInfos[i] = model.toDomain()
	}

	return wechatPayInfos, total, nil
}
