package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

type UserRelation string

const (
	DEPARTMENT UserRelation = "Department"
	POSITION   UserRelation = "Position"
	ROLES      UserRelation = "Roles"
)

// UserRepository 用户仓库接口
// 定义用户数据访问的抽象接口
type UserRepository interface {
	GetByID(id string, relations ...UserRelation) (*entity.User, error)
	GetByUsername(username string) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	GetByPhone(phone string) (*entity.User, error)
	GetByWechatOpenID(openID string) (*entity.User, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.User, int64, error)
	UpdateFields(id string, updates map[string]interface{}) error
	SetUserSpecialPermissions(userID string, apiIDs []string) error
	SetUserSpecialBtnPermissions(userID string, btnPermIDs []string) error
	GetUserSpecialPermissions(userID string) ([]*entity.API, error)
	GetUserSpecialBtnPermissions(userID string) ([]*entity.BtnPerm, error)
}
