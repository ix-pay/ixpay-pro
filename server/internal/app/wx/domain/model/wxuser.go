package model

import (
	"time"
)

// WXUser 微信用户实体
// 用于保存微信登录用户的基本信息

type WXUser struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	OpenID        string     `gorm:"size:100;not null;uniqueIndex" json:"open_id"`
	UnionID       string     `gorm:"size:100;uniqueIndex" json:"union_id,omitempty"`
	Nickname      string     `gorm:"size:100" json:"nickname"`
	Avatar        string     `gorm:"size:255" json:"avatar"`
	Gender        int        `gorm:"default:0" json:"gender"` // 0:未知, 1:男, 2:女
	Country       string     `gorm:"size:50" json:"country"`
	Province      string     `gorm:"size:50" json:"province"`
	City          string     `gorm:"size:50" json:"city"`
	Language      string     `gorm:"size:20" json:"language"`
	Subscribe     bool       `gorm:"default:false" json:"subscribe"`
	SubscribeTime *time.Time `json:"subscribe_time,omitempty"`
	Remark        string     `gorm:"size:255" json:"remark"`
	GroupID       int64      `gorm:"default:0" json:"group_id"`
	UserID        uint       `json:"user_id"` // 关联系统用户ID
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// WXAuthSession 微信授权会话信息
// 用于保存微信登录授权的会话信息

type WXAuthSession struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	WXUserID     uint      `gorm:"not null;index" json:"wx_user_id"`
	AccessToken  string    `gorm:"size:255;not null" json:"access_token"`
	RefreshToken string    `gorm:"size:255;not null" json:"refresh_token"`
	ExpiresIn    int64     `gorm:"not null" json:"expires_in"`
	Scope        string    `gorm:"size:255" json:"scope"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// WXUserRepository 微信用户仓库接口
// 提供对微信用户数据的访问方法

type WXUserRepository interface {
	// GetByID 根据ID获取微信用户信息
	GetByID(id uint) (*WXUser, error)

	// GetByOpenID 根据OpenID获取微信用户信息
	GetByOpenID(openID string) (*WXUser, error)

	// GetByUnionID 根据UnionID获取微信用户信息
	GetByUnionID(unionID string) (*WXUser, error)

	// GetByUserID 根据系统用户ID获取微信用户信息
	GetByUserID(userID uint) (*WXUser, error)

	// Create 创建微信用户
	Create(user *WXUser) error

	// Update 更新微信用户信息
	Update(user *WXUser) error

	// Delete 删除微信用户
	Delete(id uint) error

	// List 获取微信用户列表
	List(page, pageSize int) ([]*WXUser, int64, error)
}

// WXAuthSessionRepository 微信授权会话仓库接口
// 提供对微信授权会话数据的访问方法

type WXAuthSessionRepository interface {
	// GetByID 根据ID获取微信授权会话
	GetByID(id uint) (*WXAuthSession, error)

	// GetActiveSessionByWXUserID 获取指定微信用户的有效会话
	GetActiveSessionByWXUserID(wxUserID uint) (*WXAuthSession, error)

	// Create 创建微信授权会话
	Create(session *WXAuthSession) error

	// Update 更新微信授权会话
	Update(session *WXAuthSession) error

	// InvalidateSession 使指定会话失效
	InvalidateSession(id uint) error

	// InvalidateAllSessionsByWXUserID 使指定微信用户的所有会话失效
	InvalidateAllSessionsByWXUserID(wxUserID uint) error
}

// WXAuthService 微信认证服务接口
// 提供微信登录认证相关的业务逻辑

type WXAuthService interface {
	// LoginByCode 通过微信授权码登录
	// 1. 换取openid和session_key
	// 2. 查询或创建微信用户
	// 3. 生成会话token返回给客户端
	LoginByCode(code string) (*WXUser, string, string, error)

	// RefreshToken 刷新访问令牌
	RefreshToken(refreshToken string) (string, string, error)

	// GetUserInfo 获取微信用户信息
	// 根据会话token获取用户信息
	GetUserInfo(accessToken string) (*WXUser, error)

	// BindWXUserToSystemUser 绑定微信用户到系统用户
	BindWXUserToSystemUser(openID string, userID uint) error

	// UnbindWXUserFromSystemUser 解绑微信用户与系统用户
	UnbindWXUserFromSystemUser(openID string) error

	// UpdateWXUserInfo 更新微信用户信息
	UpdateWXUserInfo(user *WXUser) error

	// GenerateWXAuthPayParams 生成微信支付参数
	GenerateWXAuthPayParams(orderID string, amount int64, description string) (map[string]interface{}, error)

	// HandleWXAuthPayNotify 处理微信支付通知
	HandleWXAuthPayNotify(notifyData []byte) (map[string]interface{}, error)

	// Logout 用户登出
	Logout(accessToken string) error
}
