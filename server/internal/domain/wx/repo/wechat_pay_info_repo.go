package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"

// WechatPayInfoRepository 支付仓库接口
// 定义支付数据访问的抽象接口
type WechatPayInfoRepository interface {
	GetByID(id string) (*entity.WechatPayInfo, error)
	GetByOrderID(orderID string) (*entity.WechatPayInfo, error)
	GetByTransactionID(transactionID string) (*entity.WechatPayInfo, error)
	Create(wechatPayInfo *entity.WechatPayInfo) error
	Update(wechatPayInfo *entity.WechatPayInfo) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.WechatPayInfo, int, error)
	ListByUser(userID string, page, pageSize int) ([]*entity.WechatPayInfo, int, error)
	ListByStatus(page, pageSize int) ([]*entity.WechatPayInfo, int, error)
}
