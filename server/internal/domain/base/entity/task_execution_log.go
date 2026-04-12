package entity

import "time"

// TaskExecutionLog 任务执行日志领域实体
// 纯业务模型，无 GORM 标签
type TaskExecutionLog struct {
	ID          string    // 日志 ID
	TaskID      string    // 任务 ID
	TaskName    string    // 任务名称
	Group       string    // 任务分组
	ExecuteAt   string    // 执行时间
	Duration    int64     // 执行耗时（毫秒）
	Result      string    // 执行结果：success/failed
	ErrorInfo   string    // 错误信息
	RetryCount  int       // 重试次数
	CronExpr    string    // Cron 表达式
	TriggerType string    // 触发类型：cron/manual/retry
	OperatorID  string    // 操作人 ID（手动触发时）
	CreatedBy   string    // 创建人 ID
	CreatedAt   time.Time // 创建时间
	UpdatedBy   string    // 更新人 ID
	UpdatedAt   time.Time // 更新时间
}

// IsSuccess 检查任务执行是否成功
func (t *TaskExecutionLog) IsSuccess() bool {
	return t.Result == "success"
}

// IsFailed 检查任务执行是否失败
func (t *TaskExecutionLog) IsFailed() bool {
	return t.Result == "failed"
}

// HasError 检查是否有错误
func (t *TaskExecutionLog) HasError() bool {
	return t.ErrorInfo != ""
}

// IsRetry 检查是否是重试执行
func (t *TaskExecutionLog) IsRetry() bool {
	return t.RetryCount > 0
}

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	TaskID        string  // 任务 ID
	TaskName      string  // 任务名称
	Group         string  // 任务分组
	TotalExecutes int64   // 总执行次数
	SuccessCount  int64   // 成功次数
	FailedCount   int64   // 失败次数
	SuccessRate   float64 // 成功率
	AvgDuration   float64 // 平均耗时
	LastExecuteAt string  // 最后执行时间
	NextExecuteAt string  // 下次执行时间
}

// TaskGroupStat 任务分组统计
type TaskGroupStat struct {
	Group         string  // 任务分组
	TotalTasks    int64   // 总任务数
	TotalExecutes int64   // 总执行次数
	SuccessCount  int64   // 成功次数
	FailedCount   int64   // 失败次数
	SuccessRate   float64 // 成功率
}
