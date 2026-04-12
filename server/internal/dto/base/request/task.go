package request

// AddTaskRequest 添加任务请求参数
type AddTaskRequest struct {
	TaskID      string `json:"task_id" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=cron one_time"`
	Expression  string `json:"expression" binding:"required"`
	Description string `json:"description"`
	RetryCount  int    `json:"retry_count" binding:"min=0,max=10"`
}

// SetTaskGroupRequest 设置任务分组请求参数
type SetTaskGroupRequest struct {
	Group string `json:"group" binding:"required"`
}

// GetTaskExecutionLogsRequest 获取任务执行日志请求参数
type GetTaskExecutionLogsRequest struct {
	TaskID   string `json:"task_id" form:"task_id" binding:"required"`
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
}
