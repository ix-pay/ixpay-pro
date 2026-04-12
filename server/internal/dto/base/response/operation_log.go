package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// OperationLogResponse 操作日志响应 DTO
type OperationLogResponse struct {
	ID            string `json:"id,string"`
	UserID        string `json:"userId,string"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	OperationType int    `json:"operationType"`
	Module        string `json:"module"`
	Description   string `json:"description"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	Params        string `json:"params"`
	ClientIP      string `json:"clientIp"`
	UserAgent     string `json:"userAgent"`
	StatusCode    int    `json:"statusCode"`
	Result        string `json:"result"`
	Duration      int64  `json:"duration"`
	ErrorMessage  string `json:"errorMessage"`
	IsSuccess     bool   `json:"isSuccess"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// OperationLogListResponse 操作日志列表响应 DTO
type OperationLogListResponse struct {
	baseRes.PageResult
	List []OperationLogResponse `json:"list"`
}

// OperationLogStatisticsResponse 操作日志统计响应 DTO
type OperationLogStatisticsResponse struct {
	TotalCount     int64 `json:"totalCount"`
	SuccessCount   int64 `json:"successCount"`
	FailedCount    int64 `json:"failedCount"`
	SuccessRate    int64 `json:"successRate"`
	AvgDuration    int64 `json:"avgDuration"`
	OperationTypes []struct {
		Type  int   `json:"type"`
		Count int64 `json:"count"`
	} `json:"operationTypes"`
}
