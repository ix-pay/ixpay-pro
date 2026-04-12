package response

// TaskResponse 任务响应 DTO
type TaskResponse struct {
	TaskID      string `json:"task_id"`
	Type        string `json:"type"`
	Expression  string `json:"expression"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	LastRunAt   string `json:"last_run_at,omitempty"`
	NextRunAt   string `json:"next_run_at,omitempty"`
	RetryCount  int    `json:"retry_count"`
	MaxRetries  int    `json:"max_retries"`
}

// TaskListResponse 任务列表响应 DTO
type TaskListResponse struct {
	List []TaskResponse `json:"list"`
}

// TaskExecutionLogResponse 任务执行日志响应 DTO
type TaskExecutionLogResponse struct {
	ID          string `json:"id,string"`
	TaskID      string `json:"taskId,string"`
	TaskName    string `json:"taskName"`
	Group       string `json:"group"`
	ExecuteAt   string `json:"executeAt"`
	Duration    int64  `json:"duration"`
	Result      string `json:"result"`
	ErrorInfo   string `json:"errorInfo"`
	RetryCount  int    `json:"retryCount"`
	CronExpr    string `json:"cronExpr"`
	TriggerType string `json:"triggerType"`
	OperatorID  string `json:"operatorId,string"`
}

// TaskExecutionLogsResponse 任务执行日志列表响应 DTO
type TaskExecutionLogsResponse struct {
	Logs  []TaskExecutionLogResponse `json:"logs"`
	Total int64                      `json:"total"`
}

// TaskStatisticsResponse 任务统计响应 DTO
type TaskStatisticsResponse struct {
	TaskID        string  `json:"taskId,string"`
	TaskName      string  `json:"taskName"`
	Group         string  `json:"group"`
	TotalExecutes int64   `json:"totalExecutes"`
	SuccessCount  int64   `json:"successCount"`
	FailedCount   int64   `json:"failedCount"`
	SuccessRate   float64 `json:"successRate"`
	AvgDuration   float64 `json:"avgDuration"`
	LastExecuteAt string  `json:"lastExecuteAt"`
	NextExecuteAt string  `json:"nextExecuteAt"`
}

// TaskStatisticsListResponse 任务统计列表响应 DTO
type TaskStatisticsListResponse struct {
	List []TaskStatisticsResponse `json:"list"`
}
