package entity

import "time"

// WXUser 微信用户领域实体
// 用于保存微信登录用户的基本信息
// 纯业务模型，无 GORM 标签
type WXUser struct {
	ID            string     // 微信用户 ID
	OpenID        string     // 微信 OpenID，唯一标识
	UnionID       string     // 微信 UnionID（同一开放平台下唯一）
	Nickname      string     // 用户昵称
	Avatar        string     // 用户头像 URL
	Gender        int        // 性别：0-未知，1-男，2-女
	Country       string     // 国家
	Province      string     // 省份
	City          string     // 城市
	Language      string     // 语言
	Subscribe     bool       // 是否关注公众号
	SubscribeTime *time.Time // 关注时间
	Remark        string     // 备注
	GroupID       string     // 分组 ID（string 类型，避免 JSON 精度丢失）
	UserID        string     // 关联系统用户 ID（string 类型，避免 JSON 精度丢失）
	CreatedAt     time.Time  // 创建时间
	UpdatedAt     time.Time  // 更新时间
}

// IsSubscribed 检查用户是否已关注公众号
func (w *WXUser) IsSubscribed() bool {
	return w.Subscribe
}

// HasUnionID 检查用户是否有 UnionID
func (w *WXUser) HasUnionID() bool {
	return w.UnionID != ""
}

// IsLinkedToSystemUser 检查是否已绑定系统用户
func (w *WXUser) IsLinkedToSystemUser() bool {
	return w.UserID != ""
}

// LinkToSystemUser 绑定到系统用户
func (w *WXUser) LinkToSystemUser(userID string) {
	w.UserID = userID
}

// UnlinkFromSystemUser 解绑系统用户
func (w *WXUser) UnlinkFromSystemUser() {
	w.UserID = ""
}

// UpdateProfile 更新用户资料
func (w *WXUser) UpdateProfile(nickname, avatar, country, province, city, language string, gender int) {
	w.Nickname = nickname
	w.Avatar = avatar
	w.Country = country
	w.Province = province
	w.City = city
	w.Language = language
	w.Gender = gender
}

// MarkAsSubscribed 标记为已关注
func (w *WXUser) MarkAsSubscribed(subscribeTime time.Time) {
	w.Subscribe = true
	w.SubscribeTime = &subscribeTime
}

// MarkAsUnsubscribed 标记为已取消关注
func (w *WXUser) MarkAsUnsubscribed() {
	w.Subscribe = false
	w.SubscribeTime = nil
}
