package entity

import "time"

// UserSetting 用户设置领域实体
// 存储用户的个性化系统设置
// 与用户实体是一对一关系
type UserSetting struct {
	ID              int64     `json:"id"`              // 用户设置 ID
	UserID          int64     `json:"userId"`          // 用户 ID
	DarkMode        string    `json:"darkMode"`        // 主题模式：light-浅色, dark-深色, auto-自动
	PrimaryColor    string    `json:"primaryColor"`    // 主题颜色
	FontSize        int       `json:"fontSize"`        // 字体大小
	LayoutSideWidth int       `json:"layout_side_width"` // 侧边栏宽度
	ShowWatermark   bool      `json:"show_watermark"`  // 是否显示水印
	Language        string    `json:"language"`        // 语言
	MenuLayout      string    `json:"menuLayout"`      // 菜单布局：left-左侧, top-顶部
	UpdatedBy       int64     `json:"updatedBy"`       // 更新人 ID
	UpdatedAt       time.Time `json:"updatedAt"`       // 更新时间
}

// MenuLayoutOptions 返回菜单布局可选值
func (s *UserSetting) MenuLayoutOptions() []string {
	return []string{"left", "top"}
}
