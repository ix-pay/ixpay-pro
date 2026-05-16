package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPTask HTTP 请求任务
type HTTPTask struct {
	TaskID         string            `json:"-"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Body           string            `json:"body"`
	Timeout        int               `json:"timeout"`
	ExpectedStatus int               `json:"expected_status"`
}

// CreateHTTPTask 创建 HTTP 任务
func CreateHTTPTask(taskID string, params json.RawMessage) (Task, error) {
	var httpTask HTTPTask
	if err := json.Unmarshal(params, &httpTask); err != nil {
		return nil, fmt.Errorf("解析 HTTP 任务参数失败: %w", err)
	}

	httpTask.TaskID = taskID
	if httpTask.Method == "" {
		httpTask.Method = "GET"
	}
	if httpTask.Timeout == 0 {
		httpTask.Timeout = 30
	}
	if httpTask.ExpectedStatus == 0 {
		httpTask.ExpectedStatus = 200
	}

	return &httpTask, nil
}

// Run 执行 HTTP 请求
func (t *HTTPTask) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(t.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, t.Method, t.URL, strings.NewReader(t.Body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	for key, value := range t.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != t.ExpectedStatus {
		return fmt.Errorf("状态码不匹配: 期望 %d, 实际 %d, 响应: %s",
			t.ExpectedStatus, resp.StatusCode, string(body))
	}

	return nil
}

// GetName 获取任务名称
func (t *HTTPTask) GetName() string {
	return t.TaskID
}
