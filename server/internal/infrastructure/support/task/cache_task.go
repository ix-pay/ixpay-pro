package task

import (
	"context"
	"encoding/json"
	"fmt"
)

// CacheTask 缓存任务
type CacheTask struct {
	TaskID    string   `json:"-"`
	Action    string   `json:"action"`
	CacheKeys []string `json:"cache_keys"`
	Pattern   string   `json:"pattern"`
	TTL       int      `json:"ttl"`
}

// CreateCacheTask 创建缓存任务
func CreateCacheTask(taskID string, params json.RawMessage) (Task, error) {
	var cacheTask CacheTask
	if err := json.Unmarshal(params, &cacheTask); err != nil {
		return nil, fmt.Errorf("解析缓存任务参数失败: %w", err)
	}

	cacheTask.TaskID = taskID
	if cacheTask.Action == "" {
		cacheTask.Action = "refresh"
	}

	return &cacheTask, nil
}

// Run 执行缓存操作
func (t *CacheTask) Run(ctx context.Context) error {
	if t.Action == "" {
		return fmt.Errorf("缓存操作类型不能为空")
	}

	// 模拟缓存操作
	fmt.Printf("执行缓存任务: %s, 操作: %s, 键: %v, 模式: %s\n",
		t.TaskID, t.Action, t.CacheKeys, t.Pattern)

	return nil
}

// GetName 获取任务名称
func (t *CacheTask) GetName() string {
	return t.TaskID
}
