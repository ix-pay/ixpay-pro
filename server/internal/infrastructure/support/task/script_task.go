package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// ScriptTask 脚本任务
type ScriptTask struct {
	TaskID  string            `json:"-"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	WorkDir string            `json:"work_dir"`
	Env     map[string]string `json:"env"`
	Timeout int               `json:"timeout"`
}

// CreateScriptTask 创建脚本任务
func CreateScriptTask(taskID string, params json.RawMessage) (Task, error) {
	var scriptTask ScriptTask
	if err := json.Unmarshal(params, &scriptTask); err != nil {
		return nil, fmt.Errorf("解析脚本任务参数失败: %w", err)
	}

	scriptTask.TaskID = taskID
	if scriptTask.Timeout == 0 {
		scriptTask.Timeout = 60
	}
	if scriptTask.Command == "" {
		return nil, fmt.Errorf("命令不能为空")
	}

	return &scriptTask, nil
}

// Run 执行脚本
func (t *ScriptTask) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(t.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.Command, t.Args...)
	cmd.Dir = t.WorkDir

	env := cmd.Environ()
	for key, value := range t.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("脚本执行失败: %w, 输出: %s", err, string(output))
	}

	return nil
}

// GetName 获取任务名称
func (t *ScriptTask) GetName() string {
	return t.TaskID
}
