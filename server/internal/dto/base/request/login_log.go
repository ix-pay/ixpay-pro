// request 包定义登录日志和在线用户管理相关的请求模型
// 用于接收和验证 HTTP 请求参数
package request

// GetLoginLogListRequest 获取登录日志列表请求
type GetLoginLogListRequest struct {
	Page      int    `form:"page" binding:"required"`     // 页码
	PageSize  int    `form:"pageSize" binding:"required"` // 每页数量
	UserID    *int64 `form:"userId"`                      // 用户 ID（可选筛选）
	Username  string `form:"username"`                    // 用户名（可选筛选）
	LoginIP   string `form:"loginIp"`                     // 登录 IP（可选筛选）
	Result    *int   `form:"result"`                      // 登录结果：0-失败，1-成功（可选筛选）
	StartDate string `form:"startDate"`                   // 开始日期（可选筛选）
	EndDate   string `form:"endDate"`                     // 结束日期（可选筛选）
}

// GetLoginStatisticsRequest 获取登录统计请求
type GetLoginStatisticsRequest struct {
	StartDate string `form:"startDate" binding:"required"` // 开始日期（YYYY-MM-DD）
	EndDate   string `form:"endDate" binding:"required"`   // 结束日期（YYYY-MM-DD）
}

// GetAbnormalLoginsRequest 获取异常登录请求
type GetAbnormalLoginsRequest struct {
	Page     int `form:"page" binding:"required"`     // 页码
	PageSize int `form:"pageSize" binding:"required"` // 每页数量
}

// GetOnlineUserListRequest 获取在线用户列表请求
type GetOnlineUserListRequest struct {
	Page     int `form:"page" binding:"required"`     // 页码（预留分页支持）
	PageSize int `form:"pageSize" binding:"required"` // 每页数量（预留分页支持）
}

// GetOnlineUserByIDRequest 获取在线用户详情请求
type GetOnlineUserByIDRequest struct {
	UserID int64 `uri:"userId" binding:"required"` // 用户 ID
}

// ForceOfflineRequest 强制下线请求
type ForceOfflineRequest struct {
	UserID int64  `json:"userId" binding:"required"` // 用户 ID
	Reason string `json:"reason"`                    // 下线原因
}

// BatchForceOfflineRequest 批量强制下线请求
type BatchForceOfflineRequest struct {
	UserIDs []int64 `json:"userIds" binding:"required"` // 用户 ID 列表
	Reason  string  `json:"reason"`                     // 下线原因
}

// RecordLoginRequest 记录登录日志请求（内部调用）
type RecordLoginRequest struct {
	UserID    int64  `json:"userId" binding:"required"`   // 用户 ID
	Username  string `json:"username" binding:"required"` // 用户名
	IP        string `json:"ip" binding:"required"`       // 登录 IP
	Place     string `json:"place"`                       // 登录地点
	Device    string `json:"device"`                      // 设备信息
	Browser   string `json:"browser"`                     // 浏览器信息
	OS        string `json:"os"`                          // 操作系统
	UserAgent string `json:"userAgent"`                   // 原始 User-Agent
	Success   bool   `json:"success"`                     // 是否成功
	ErrorMsg  string `json:"errorMsg"`                    // 错误信息
}
