package entity

import "time"

// WXAuthSession 微信授权会话领域实体
// 用于保存微信登录授权的会话信息
// 纯业务模型，无 GORM 标签
type WXAuthSession struct {
	ID           int64     // 授权会话 ID
	WXUserID     int64     // 微信用户 ID（int64 类型）
	AccessToken  string    // 访问令牌
	RefreshToken string    // 刷新令牌
	ExpiresIn    int64     // 过期时间（秒）
	Scope        string    // 授权 scope
	IsActive     bool      // 是否激活
	ExpiresAt    time.Time // 过期时间
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

// IsExpired 检查会话是否已过期
func (s *WXAuthSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid 检查会话是否有效（未过期且激活）
func (s *WXAuthSession) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// MarkAsInactive 标记会话为失效
func (s *WXAuthSession) MarkAsInactive() {
	s.IsActive = false
}

// Refresh 刷新会话
func (s *WXAuthSession) Refresh(accessToken, refreshToken string, expiresIn int64) {
	s.AccessToken = accessToken
	s.RefreshToken = refreshToken
	s.ExpiresIn = expiresIn
	s.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
}
