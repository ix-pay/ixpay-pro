package entity

import "time"

type Task struct {
	ID          int64      `json:"id" gorm:"column:id"`
	TaskID      string     `json:"taskId" gorm:"column:task_id"`
	TaskType    string     `json:"taskType" gorm:"column:task_type"`
	Type        string     `json:"type" gorm:"column:type"`
	Expression  string     `json:"expression" gorm:"column:expression"`
	Description string     `json:"description" gorm:"column:description"`
	Group       string     `json:"group" gorm:"column:group"`
	Status      int        `json:"status" gorm:"column:status"`
	Params      string     `json:"params" gorm:"column:params"`
	RetryCount  int        `json:"retryCount" gorm:"column:retry_count"`
	Concurrency string     `json:"concurrency" gorm:"column:concurrency"`
	Timeout     int        `json:"timeout" gorm:"column:timeout"`
	LastRunAt   *time.Time `json:"lastRunAt" gorm:"column:last_run_at"`
	NextRunAt   *time.Time `json:"nextRunAt" gorm:"column:next_run_at"`
	CreatedBy   int64      `json:"createdBy" gorm:"column:created_by"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedBy   int64      `json:"updatedBy" gorm:"column:updated_by"`
	UpdatedAt   time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

func (Task) TableName() string {
	return "base_tasks"
}
