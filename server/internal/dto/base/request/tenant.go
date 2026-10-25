package request

// SaveTenantRequest 创建/更新租户请求（独立实例连接信息可选覆盖；dbPassword 明文输入服务端加密）
type SaveTenantRequest struct {
	ID          int64  `json:"id,string"`   // 更新时必传；创建时为空
	TenantCode  string `json:"tenantCode"`  // 租户编码（小写字母开头 3-31 位字母数字下划线）
	TenantName  string `json:"tenantName"`  // 租户名称
	DBHost      string `json:"dbHost"`      // 独立实例 Host（可选）
	DBPort      int    `json:"dbPort"`      // 独立实例端口（可选，0 = 沿用主实例）
	DBUser      string `json:"dbUser"`      // 独立实例用户（可选）
	DBPassword  string `json:"dbPassword"`  // 独立实例密码（明文输入；更新时空串 = 保留原值）
	Description string `json:"description"` // 描述
}

// CreateTenantRouteRequest 绑定应用路由请求
type CreateTenantRouteRequest struct {
	AppKey     string `json:"appKey" binding:"required"` // 前端实例标志
	ModuleCode string `json:"moduleCode"`                // 所属模块（空 = wx）
}

// CreateTenantGrantRequest 租户授权请求
type CreateTenantGrantRequest struct {
	UserID int64 `json:"userId,string" binding:"required"` // 管理平台用户 ID
}
