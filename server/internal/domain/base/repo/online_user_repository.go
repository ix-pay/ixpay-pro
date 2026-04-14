package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// OnlineUserRepository 在线用户仓库接口（基于 Redis）
type OnlineUserRepository interface {
	Add(user *entity.OnlineUser) error
	GetByUserID(userID int64) (*entity.OnlineUser, error)
	GetBySessionID(sessionID string) (*entity.OnlineUser, error)
	UpdateActiveTime(userID int64) error
	Remove(userID int64) error
	RemoveBySessionID(sessionID string) error
	GetAll() ([]*entity.OnlineUser, error)
	GetCount() (int, error)
	Exists(userID int64) (bool, error)
	ClearExpired() error
}
