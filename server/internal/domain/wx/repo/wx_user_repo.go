package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/wx/entity"

// WXUserRepository 微信用户仓库接口
// 提供对微信用户数据的访问方法
type WXUserRepository interface {
	GetByID(id int64) (*entity.WXUser, error)
	GetByOpenID(openID string) (*entity.WXUser, error)
	GetByUnionID(unionID string) (*entity.WXUser, error)
	GetByUserID(userID int64) (*entity.WXUser, error)
	Create(user *entity.WXUser) error
	Update(user *entity.WXUser) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.WXUser, int, error)
}
