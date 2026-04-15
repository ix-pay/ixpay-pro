package request

// GetOperationLogListRequest 获取操作日志列表请求参数
type GetOperationLogListRequest struct {
	Page          int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize      int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
	StartTime     string `json:"startTime" form:"startTime" binding:"omitempty"`
	EndTime       string `json:"endTime" form:"endTime" binding:"omitempty"`
	Username      string `json:"userName" form:"userName" binding:"omitempty"`
	Module        string `json:"module" form:"module" binding:"omitempty"`
	OperationType int    `json:"operationType" form:"operationType" binding:"omitempty,min=0,max=9"`
	IsSuccess     *bool  `json:"isSuccess" form:"isSuccess" binding:"omitempty"`
}

// GetOperationLogByIDRequest 根据 ID 获取操作日志请求参数
type GetOperationLogByIDRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"`
}

// BatchDeleteOperationLogRequest 批量删除操作日志请求参数
type BatchDeleteOperationLogRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// GetOperationLogStatisticsRequest 获取操作日志统计请求参数
type GetOperationLogStatisticsRequest struct {
	StartTime string `json:"startTime" form:"startTime" binding:"omitempty"`
	EndTime   string `json:"endTime" form:"endTime" binding:"omitempty"`
}

// ClearOperationLogByTimeRangeRequest 清空操作日志请求参数
type ClearOperationLogByTimeRangeRequest struct {
	StartTime string `json:"startTime" binding:"required"`
	EndTime   string `json:"endTime" binding:"required"`
}
