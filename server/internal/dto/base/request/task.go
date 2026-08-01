package request

import "encoding/json"

// AddTaskRequest 添加任务请求参数
type AddTaskRequest struct {
	TaskID      string          `json:"taskId" binding:"required"`
	TaskType    string          `json:"taskType" binding:"required,oneof=http database cache script stream_maintenance"`
	Type        string          `json:"type" binding:"required,oneof=cron one_time"`
	Expression  string          `json:"expression" binding:"required"`
	Description string          `json:"description"`
	Group       string          `json:"group"`
	RetryCount  int             `json:"retryCount" binding:"min=0,max=10"`
	Params      json.RawMessage `json:"params" binding:"required"`
}

// SetTaskGroupRequest 设置任务分组请求参数
type SetTaskGroupRequest struct {
	Group string `json:"group" binding:"required"`
}

// GetTaskExecutionLogsRequest 获取任务执行日志请求参数
type GetTaskExecutionLogsRequest struct {
	TaskID   string `json:"task_id" form:"task_id"`
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
}

// SearchTaskExecutionLogsRequest 搜索任务执行日志请求参数（统一查询接口）
type SearchTaskExecutionLogsRequest struct {
	TaskID    string `json:"taskId" form:"taskId"`
	Result    string `json:"result" form:"result" binding:"omitempty,oneof=success failed"`
	StartDate string `json:"startDate" form:"startDate"`
	EndDate   string `json:"endDate" form:"endDate"`
	Page      int    `json:"page" form:"page" binding:"min=1"`
	PageSize  int    `json:"pageSize" form:"pageSize" binding:"min=1,max=100"`
}
