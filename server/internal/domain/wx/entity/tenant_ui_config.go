package entity

import "time"

// TenantUIConfig 租户前端个性化配置（wxapp/miniapp 空壳按 app_config_id 加载渲染）
// 当前单库阶段存主库；多租户分库后随租户库迁移（表名模块无关）
type TenantUIConfig struct {
	ID          int64     // 配置 ID
	AppConfigID int64     // 关联 wx_app_configs.id（每个前端实例一份）
	UIJSON      string    // UI 配置 JSON（schema 见 specs/02，brand/home/payment/extras）
	Status      int       // 状态：1-启用，0-禁用
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}
