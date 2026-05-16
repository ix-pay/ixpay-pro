package task

import (
	"encoding/json"
	"fmt"
	"time"
)

// TaskBuilder 任务构建器
type TaskBuilder struct {
	taskID      string
	taskType    TaskType
	cronExpr    string
	group       string
	description string
	timeout     time.Duration
	concurrency ConcurrencyMode
	maxRetries  int
	params      map[string]interface{}
	err         error
}

// NewTaskBuilder 创建任务构建器
func NewTaskBuilder(taskID string) *TaskBuilder {
	return &TaskBuilder{
		taskID:      taskID,
		taskType:    TaskTypeHTTP,
		cronExpr:    "0 */5 * * * *",
		group:       "default",
		timeout:     5 * time.Minute,
		concurrency: ConcurrencySkip,
		maxRetries:  3,
		params:      make(map[string]interface{}),
	}
}

// Type 设置任务类型
func (b *TaskBuilder) Type(t TaskType) *TaskBuilder {
	b.taskType = t
	return b
}

// CronExpr 设置 Cron 表达式
func (b *TaskBuilder) CronExpr(expr string) *TaskBuilder {
	b.cronExpr = expr
	return b
}

// Group 设置任务分组
func (b *TaskBuilder) Group(group string) *TaskBuilder {
	b.group = group
	return b
}

// Description 设置任务描述
func (b *TaskBuilder) Description(desc string) *TaskBuilder {
	b.description = desc
	return b
}

// Timeout 设置任务超时时间
func (b *TaskBuilder) Timeout(timeout time.Duration) *TaskBuilder {
	b.timeout = timeout
	return b
}

// Concurrency 设置并发模式
func (b *TaskBuilder) Concurrency(mode ConcurrencyMode) *TaskBuilder {
	b.concurrency = mode
	return b
}

// MaxRetries 设置最大重试次数
func (b *TaskBuilder) MaxRetries(count int) *TaskBuilder {
	b.maxRetries = count
	return b
}

// Param 设置任务参数
func (b *TaskBuilder) Param(key string, value interface{}) *TaskBuilder {
	b.params[key] = value
	return b
}

// Params 批量设置任务参数
func (b *TaskBuilder) Params(params map[string]interface{}) *TaskBuilder {
	for k, v := range params {
		b.params[k] = v
	}
	return b
}

// HTTP 配置 HTTP 任务
func (b *TaskBuilder) HTTP() *HTTPTaskBuilder {
	b.taskType = TaskTypeHTTP
	return &HTTPTaskBuilder{parent: b}
}

// Database 配置数据库任务
func (b *TaskBuilder) Database() *DatabaseTaskBuilder {
	b.taskType = TaskTypeDatabase
	return &DatabaseTaskBuilder{parent: b}
}

// Cache 配置缓存任务
func (b *TaskBuilder) Cache() *CacheTaskBuilder {
	b.taskType = TaskTypeCache
	return &CacheTaskBuilder{parent: b}
}

// Script 配置脚本任务
func (b *TaskBuilder) Script() *ScriptTaskBuilder {
	b.taskType = TaskTypeScript
	return &ScriptTaskBuilder{parent: b}
}

// Build 构建任务
func (b *TaskBuilder) Build() (*ScheduledTask, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.taskID == "" {
		return nil, fmt.Errorf("任务 ID 不能为空")
	}

	if b.cronExpr == "" {
		return nil, fmt.Errorf("Cron 表达式不能为空")
	}

	params, err := json.Marshal(b.params)
	if err != nil {
		return nil, fmt.Errorf("序列化任务参数失败: %w", err)
	}

	taskInst, err := NewTaskFactory().CreateTask(b.taskID, b.taskType, params)
	if err != nil {
		return nil, fmt.Errorf("创建任务实例失败: %w", err)
	}

	return &ScheduledTask{
		Task:        taskInst,
		CronExpr:    b.cronExpr,
		Group:       b.group,
		Concurrency: b.concurrency,
		Timeout:     b.timeout,
	}, nil
}

// HTTPTaskBuilder HTTP 任务构建器
type HTTPTaskBuilder struct {
	parent *TaskBuilder
}

// URL 设置请求 URL
func (b *HTTPTaskBuilder) URL(url string) *HTTPTaskBuilder {
	b.parent.params["url"] = url
	return b
}

// Method 设置请求方法
func (b *HTTPTaskBuilder) Method(method string) *HTTPTaskBuilder {
	b.parent.params["method"] = method
	return b
}

// Header 设置请求头
func (b *HTTPTaskBuilder) Header(key, value string) *HTTPTaskBuilder {
	headers, ok := b.parent.params["headers"].(map[string]string)
	if !ok {
		headers = make(map[string]string)
		b.parent.params["headers"] = headers
	}
	headers[key] = value
	return b
}

// Body 设置请求体
func (b *HTTPTaskBuilder) Body(body string) *HTTPTaskBuilder {
	b.parent.params["body"] = body
	return b
}

// Timeout 设置请求超时时间（秒）
func (b *HTTPTaskBuilder) Timeout(seconds int) *HTTPTaskBuilder {
	b.parent.params["timeout"] = seconds
	return b
}

// ExpectedStatus 设置期望的状态码
func (b *HTTPTaskBuilder) ExpectedStatus(status int) *HTTPTaskBuilder {
	b.parent.params["expected_status"] = status
	return b
}

// Done 完成 HTTP 任务配置，返回父构建器
func (b *HTTPTaskBuilder) Done() *TaskBuilder {
	return b.parent
}

// DatabaseTaskBuilder 数据库任务构建器
type DatabaseTaskBuilder struct {
	parent *TaskBuilder
}

// Query 设置 SQL 查询
func (b *DatabaseTaskBuilder) Query(query string) *DatabaseTaskBuilder {
	b.parent.params["query"] = query
	return b
}

// QueryArgs 设置查询参数
func (b *DatabaseTaskBuilder) QueryArgs(args []interface{}) *DatabaseTaskBuilder {
	b.parent.params["query_args"] = args
	return b
}

// DBType 设置数据库类型
func (b *DatabaseTaskBuilder) DBType(dbType string) *DatabaseTaskBuilder {
	b.parent.params["db_type"] = dbType
	return b
}

// Done 完成数据库任务配置，返回父构建器
func (b *DatabaseTaskBuilder) Done() *TaskBuilder {
	return b.parent
}

// CacheTaskBuilder 缓存任务构建器
type CacheTaskBuilder struct {
	parent *TaskBuilder
}

// Action 设置缓存操作类型
func (b *CacheTaskBuilder) Action(action string) *CacheTaskBuilder {
	b.parent.params["action"] = action
	return b
}

// CacheKeys 设置缓存键
func (b *CacheTaskBuilder) CacheKeys(keys []string) *CacheTaskBuilder {
	b.parent.params["cache_keys"] = keys
	return b
}

// Pattern 设置缓存键模式
func (b *CacheTaskBuilder) Pattern(pattern string) *CacheTaskBuilder {
	b.parent.params["pattern"] = pattern
	return b
}

// TTL 设置缓存过期时间
func (b *CacheTaskBuilder) TTL(ttl int) *CacheTaskBuilder {
	b.parent.params["ttl"] = ttl
	return b
}

// Done 完成缓存任务配置，返回父构建器
func (b *CacheTaskBuilder) Done() *TaskBuilder {
	return b.parent
}

// ScriptTaskBuilder 脚本任务构建器
type ScriptTaskBuilder struct {
	parent *TaskBuilder
}

// Command 设置命令
func (b *ScriptTaskBuilder) Command(cmd string) *ScriptTaskBuilder {
	b.parent.params["command"] = cmd
	return b
}

// Args 设置命令参数
func (b *ScriptTaskBuilder) Args(args []string) *ScriptTaskBuilder {
	b.parent.params["args"] = args
	return b
}

// WorkDir 设置工作目录
func (b *ScriptTaskBuilder) WorkDir(dir string) *ScriptTaskBuilder {
	b.parent.params["work_dir"] = dir
	return b
}

// Env 设置环境变量
func (b *ScriptTaskBuilder) Env(key, value string) *ScriptTaskBuilder {
	env, ok := b.parent.params["env"].(map[string]string)
	if !ok {
		env = make(map[string]string)
		b.parent.params["env"] = env
	}
	env[key] = value
	return b
}

// Timeout 设置脚本超时时间（秒）
func (b *ScriptTaskBuilder) Timeout(seconds int) *ScriptTaskBuilder {
	b.parent.params["timeout"] = seconds
	return b
}

// Done 完成脚本任务配置，返回父构建器
func (b *ScriptTaskBuilder) Done() *TaskBuilder {
	return b.parent
}
