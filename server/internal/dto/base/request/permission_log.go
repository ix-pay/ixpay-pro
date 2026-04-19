// request 包定义权限日志相关的请求模型
package request

// GetPermissionLogListRequest 获取权限日志列表请求
type GetPermissionLogListRequest struct {
	Page      int    `form:"page" binding:"required"`     // 页码
	PageSize  int    `form:"pageSize" binding:"required"` // 每页数量
	UserID    *int64 `form:"userId"`                      // 用户 ID（可选筛选）
	Username  string `form:"userName"`                    // 用户名（可选筛选）
	Operation string `form:"operation"`                   // 操作类型（可选筛选）
	Module    string `form:"module"`                      // 模块（可选筛选）
	StartDate string `form:"startDate"`                   // 开始日期（YYYY-MM-DD）
	EndDate   string `form:"endDate"`                     // 结束日期（YYYY-MM-DD）
}

// GetRolePermissionLogsRequest 获取角色权限日志请求
type GetRolePermissionLogsRequest struct {
	RoleID   int64 `uri:"roleId" binding:"required"`    // 角色 ID
	Page     int   `form:"page" binding:"required"`     // 页码
	PageSize int   `form:"pageSize" binding:"required"` // 每页数量
}
