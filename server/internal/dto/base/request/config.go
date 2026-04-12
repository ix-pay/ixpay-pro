package request

// ConfigRequest 配置基础请求模型
type ConfigRequest struct {
	ConfigKey   string `json:"configKey" binding:"required,max=100"` // 配置键
	ConfigValue string `json:"configValue"`                          // 配置值
	ConfigType  string `json:"configType" binding:"max=20"`          // 配置类型
	Description string `json:"description" binding:"max=255"`        // 描述
	Status      int    `json:"status" binding:"oneof=0 1"`           // 状态：1-启用 0-禁用
}

// CreateConfigRequest 创建配置请求模型
type CreateConfigRequest struct {
	ConfigRequest
}

// UpdateConfigRequest 更新配置请求模型
type UpdateConfigRequest struct {
	ID string `json:"id" binding:"required,gt=0"` // 配置 ID
	ConfigRequest
}

// DeleteConfigRequest 删除配置请求模型
type DeleteConfigRequest struct {
	ID string `json:"id" binding:"required,gt=0"` // 配置 ID
}

// GetConfigRequest 获取配置请求模型
type GetConfigRequest struct {
	ID        string `form:"id" binding:"omitempty,gt=0"`           // 配置 ID
	ConfigKey string `form:"configKey" binding:"omitempty,max=100"` // 配置键
}
