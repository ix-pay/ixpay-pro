package model

import (
	"time"
)

// User 用户实体
type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Username      string    `gorm:"size:50;unique;not null" json:"username"`
	PasswordHash  string    `gorm:"size:100;not null" json:"-"` // 不输出到JSON
	Nickname      string    `gorm:"size:50" json:"nickname"`
	Email         string    `gorm:"size:100;unique" json:"email"`
	Phone         string    `gorm:"size:20;unique" json:"phone"`
	Avatar        string    `gorm:"size:255" json:"avatar"`
	Role          string    `gorm:"size:20;default:'user'" json:"role"` // user, admin, etc.
	WechatOpenID  string    `gorm:"size:100;unique" json:"-"`           // 微信OpenID
	WechatUnionID string    `gorm:"size:100;unique" json:"-"`           // 微信UnionID
	Status        int       `gorm:"default:1" json:"status"`            // 1: active, 0: inactive
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserRepository 用户仓库接口
type UserRepository interface {
	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByPhone(phone string) (*User, error)
	GetByWechatOpenID(openID string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
	List(page, pageSize int, filters map[string]interface{}) ([]*User, int64, error)
}

// UserService 用户领域服务接口
type UserService interface {
	Register(username, password, email string) (*User, error)
	Login(username, password string) (*User, string, string, error) // 返回用户信息和令牌
	WechatLogin(code string) (*User, string, string, error)         // 微信登录
	GetUserInfo(userID uint) (*User, error)
	UpdateUserInfo(user *User) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
	RefreshToken(refreshToken string) (string, string, error) // 刷新令牌
}
