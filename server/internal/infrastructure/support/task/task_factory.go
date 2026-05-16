package task

import (
	"encoding/json"
	"fmt"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeHTTP             TaskType = "http"
	TaskTypeDatabase         TaskType = "database"
	TaskTypeCache            TaskType = "cache"
	TaskTypeScript           TaskType = "script"
	TaskTypeStreamMaintenance TaskType = "stream_maintenance"
)

// TaskCreator 任务创建函数
type TaskCreator func(taskID string, params json.RawMessage) (Task, error)

// TaskFactory 任务工厂
type TaskFactory struct {
	creators map[TaskType]TaskCreator
}

// NewTaskFactory 创建任务工厂
func NewTaskFactory() *TaskFactory {
	factory := &TaskFactory{
		creators: make(map[TaskType]TaskCreator),
	}

	// 注册内置任务类型
	factory.Register(TaskTypeHTTP, CreateHTTPTask)
	factory.Register(TaskTypeDatabase, CreateDatabaseTask)
	factory.Register(TaskTypeCache, CreateCacheTask)
	factory.Register(TaskTypeScript, CreateScriptTask)
	factory.Register(TaskTypeStreamMaintenance, CreateStreamMaintenanceTask)

	return factory
}

// Register 注册任务类型
func (f *TaskFactory) Register(taskType TaskType, creator TaskCreator) {
	f.creators[taskType] = creator
}

// CreateTask 创建任务实例
func (f *TaskFactory) CreateTask(taskID string, taskType TaskType, params json.RawMessage) (Task, error) {
	creator, exists := f.creators[taskType]
	if !exists {
		return nil, fmt.Errorf("不支持的任务类型: %s", taskType)
	}

	return creator(taskID, params)
}
