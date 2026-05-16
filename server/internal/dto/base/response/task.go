package response

// TaskResponse 任务响应 DTO
type TaskResponse struct {
	ID          int64  `json:"id,string"`
	TaskID      string `json:"taskId"`
	TaskType    string `json:"taskType"`
	Type        string `json:"type"`
	Expression  string `json:"expression"`
	Description string `json:"description"`
	Group       string `json:"group"`
	Status      string `json:"status"`
	StatusLabel string `json:"statusLabel"`
	CreatedAt   string `json:"createdAt"`
	LastRunAt   string `json:"lastRunAt,omitempty"`
	NextRunAt   string `json:"nextRunAt,omitempty"`
	RetryCount  int    `json:"retryCount"`
	MaxRetries  int    `json:"maxRetries"`
	Concurrency string `json:"concurrency"`
	Timeout     int    `json:"timeout"`
}

// TaskListResponse 任务列表响应 DTO
type TaskListResponse struct {
	List  []TaskResponse `json:"list"`
	Total int64          `json:"total"`
}

// TaskExecutionLogResponse 任务执行日志响应 DTO
type TaskExecutionLogResponse struct {
	ID          int64  `json:"id,string"`
	TaskID      int64  `json:"taskId,string"`
	TaskName    string `json:"taskName"`
	Group       string `json:"group"`
	ExecuteAt   string `json:"executeAt"`
	Duration    int64  `json:"duration"`
	Result      string `json:"result"`
	ErrorInfo   string `json:"errorInfo"`
	RetryCount  int    `json:"retryCount"`
	CronExpr    string `json:"cronExpr"`
	TriggerType string `json:"triggerType"`
	OperatorID  int64  `json:"operatorId,string"`
}

// TaskExecutionLogsResponse 任务执行日志列表响应 DTO
type TaskExecutionLogsResponse struct {
	Logs  []TaskExecutionLogResponse `json:"logs"`
	Total int64                      `json:"total"`
}

// TaskStatisticsResponse 任务统计响应 DTO
type TaskStatisticsResponse struct {
	TaskID        int64   `json:"taskId,string"`
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
