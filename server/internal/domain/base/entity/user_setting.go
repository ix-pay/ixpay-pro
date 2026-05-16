package entity

import "time"

// UserSetting 用户设置领域实体
// 存储用户的个性化系统设置
// 与用户实体是一对一关系
// 纯业务模型，无 GORM 标签
type UserSetting struct {
	ID               int64     // 用户设置 ID
	UserID           int64     // 用户 ID
	ThemeColor       string    // 主题颜色
	SidebarColor     string    // 侧边栏颜色
	NavbarColor      string    // 导航栏颜色
	FontSize         int       // 字体大小
	Language         string    // 语言
	AutoLogin        bool      // 自动登录
	RememberPassword bool      // 记住密码
	CreatedBy        int64     // 创建人 ID
	CreatedAt        time.Time // 创建时间
	UpdatedBy        int64     // 更新人 ID
	UpdatedAt        time.Time // 更新时间
}

// IsAutoLogin 检查是否启用自动登录
func (s *UserSetting) IsAutoLogin() bool {
	return s.AutoLogin
}

// IsRememberPassword 检查是否启用记住密码
func (s *UserSetting) IsRememberPassword() bool {
	return s.RememberPassword
}
