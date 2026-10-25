package response

// TenantResponse 租户响应（ID 雪花值经 ,string 标签序列化为字符串；不回显数据库密码密文）
type TenantResponse struct {
	ID          int64  `json:"id,string"`   // 租户 ID
	TenantCode  string `json:"tenantCode"`  // 租户编码
	TenantName  string `json:"tenantName"`  // 租户名称
	DBName      string `json:"dbName"`      // 业务库名（空 = 默认租户主库）
	DBHost      string `json:"dbHost"`      // 独立实例 Host
	DBPort      int    `json:"dbPort"`      // 独立实例端口
	DBUser      string `json:"dbUser"`      // 独立实例用户
	HasDBPass   bool   `json:"hasDbPass"`   // 是否已设置独立实例密码
	Status      string `json:"status"`      // init/active/disabled/archived/failed
	InitError   string `json:"initError"`   // 初始化失败原因
	Description string `json:"description"` // 描述
	CreatedAt   string `json:"createdAt"`   // 创建时间
	UpdatedAt   string `json:"updatedAt"`   // 更新时间
}

// GrantedTenantResponse 当前用户可见租户（租户切换器数据源）
type GrantedTenantResponse struct {
	ID         int64  `json:"id,string"`  // 租户 ID
	TenantCode string `json:"tenantCode"` // 租户编码
	TenantName string `json:"tenantName"` // 租户名称
	DBName     string `json:"dbName"`     // 业务库名
	Status     string `json:"status"`     // 状态
}

// TenantRouteResponse 应用路由响应（app_key → 租户）
type TenantRouteResponse struct {
	ID         int64  `json:"id,string"`       // 路由 ID
	AppKey     string `json:"appKey"`          // 前端实例标志
	TenantID   int64  `json:"tenantId,string"` // 所属租户 ID
	ModuleCode string `json:"moduleCode"`      // 所属模块
	CreatedAt  string `json:"createdAt"`       // 绑定时间
}

// TenantGrantResponse 租户授权用户响应
type TenantGrantResponse struct {
	ID        int64  `json:"id,string"`       // 授权 ID
	UserID    int64  `json:"userId,string"`   // 管理平台用户 ID
	TenantID  int64  `json:"tenantId,string"` // 租户 ID
	CreatedAt string `json:"createdAt"`       // 授权时间
}

// TenantOpLogResponse 租户运维审计日志响应
type TenantOpLogResponse struct {
	ID         int64  `json:"id,string"`         // 日志 ID
	TenantID   int64  `json:"tenantId,string"`   // 租户 ID
	Action     string `json:"action"`            // 动作
	Detail     string `json:"detail"`            // 详情
	OperatorID int64  `json:"operatorId,string"` // 操作人 ID
	CreatedAt  string `json:"createdAt"`         // 时间
}
