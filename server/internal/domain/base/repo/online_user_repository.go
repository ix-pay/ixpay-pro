package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// OnlineUserRepository 在线用户仓库接口（基于 Redis）
type OnlineUserRepository interface {
	Add(user *entity.OnlineUser) error
	GetByUserID(userID string) (*entity.OnlineUser, error)
	GetBySessionID(sessionID string) (*entity.OnlineUser, error)
	UpdateActiveTime(userID string) error
	Remove(userID string) error
	RemoveBySessionID(sessionID string) error
	GetAll() ([]*entity.OnlineUser, error)
	GetCount() (int, error)
	Exists(userID string) (bool, error)
	ClearExpired() error
}
