package task

import "context"

// Task 任务接口
type Task interface {
	Run(ctx context.Context) error
	GetName() string
}

// TaskWithGroup 带分组的任务接口
type TaskWithGroup interface {
	Task
	GetGroup() string
}
