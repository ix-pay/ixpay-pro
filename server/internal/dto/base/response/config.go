package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// ConfigResponse 配置响应模型
// 所有 ID 字段使用 int64 格式，通过 json:",string" 标签自动序列化为字符串
type ConfigResponse struct {
	ID          int64  `json:"id,string"`   // 配置 ID
	ConfigKey   string `json:"configKey"`   // 配置键
	ConfigValue string `json:"configValue"` // 配置值
	ConfigType  string `json:"configType"`  // 配置类型
	Description string `json:"description"` // 描述
	Status      int    `json:"status"`      // 状态：1-启用 0-禁用
	CreatedAt   string `json:"createdAt"`   // 创建时间
	UpdatedAt   string `json:"updatedAt"`   // 更新时间
}

// ConfigListResponse 配置列表响应模型
type ConfigListResponse struct {
	baseRes.PageResult
	List []ConfigResponse `json:"list"` // 配置列表
}
