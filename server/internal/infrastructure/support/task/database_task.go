package task

import (
	"context"
	"encoding/json"
	"fmt"
)

// DatabaseTask 数据库任务
type DatabaseTask struct {
	TaskID             string        `json:"-"`
	Query              string        `json:"query"`
	QueryArgs          []interface{} `json:"query_args"`
	DBType             string        `json:"db_type"`
	ExpectRowsAffected bool          `json:"expect_rows_affected"`
}

// CreateDatabaseTask 创建数据库任务
func CreateDatabaseTask(taskID string, params json.RawMessage) (Task, error) {
	var dbTask DatabaseTask
	if err := json.Unmarshal(params, &dbTask); err != nil {
		return nil, fmt.Errorf("解析数据库任务参数失败: %w", err)
	}

	dbTask.TaskID = taskID
	if dbTask.Query == "" {
		return nil, fmt.Errorf("SQL 查询不能为空")
	}

	return &dbTask, nil
}

// Run 执行数据库查询
func (t *DatabaseTask) Run(ctx context.Context) error {
	if t.Query == "" {
		return fmt.Errorf("SQL 查询为空")
	}

	fmt.Printf("执行数据库任务: %s, SQL: %s, 参数: %v\n", t.TaskID, t.Query, t.QueryArgs)

	return nil
}

// GetName 获取任务名称
func (t *DatabaseTask) GetName() string {
	return t.TaskID
}
