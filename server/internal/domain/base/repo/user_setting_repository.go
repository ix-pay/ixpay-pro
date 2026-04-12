package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// UserSettingRepository 用户设置仓库接口
type UserSettingRepository interface {
	GetByUserID(userID string) (*entity.UserSetting, error)
	Create(setting *entity.UserSetting) error
	Update(setting *entity.UserSetting) error
	Delete(id string) error
}
