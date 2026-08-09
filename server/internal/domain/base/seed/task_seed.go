package seed

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// TaskSeed 定时任务种子数据
type TaskSeed struct {
	taskRepo repo.TaskRepository
}

// NewTaskSeed 创建任务种子数据实例
func NewTaskSeed(taskRepo repo.TaskRepository) Seed {
	return &TaskSeed{
		taskRepo: taskRepo,
	}
}

// Version 返回种子数据版本
func (ts *TaskSeed) Version() string {
	return "v1.0.0"
}

// Name 返回种子数据名称
func (ts *TaskSeed) Name() string {
	return "task_seed"
}

// Order 返回初始化顺序（第十个执行，任务数据最后初始化）
func (ts *TaskSeed) Order() int {
	return 10
}

// Init 初始化任务种子数据
func (ts *TaskSeed) Init(db *database.PostgresDB, logger logger.Logger) error {
	logger.Info("开始初始化任务种子数据")

	tasks := []struct {
		taskID       string
		taskType     string
		scheduleType string
		expression   string
		description  string
		group        string
		params       string
	}{
		{
			taskID:       "http_health_check",
			taskType:     "http",
			scheduleType: "cron",
			expression:   "0 */5 * * * *",
			description:  "定时检查外部服务健康状态，每5分钟执行一次",
			group:        "monitoring",
			params:       `{"url": "http://localhost:8080/health", "method": "GET", "timeout": 5000}`,
		},
		{
			taskID:       "db_cleanup_expired_data",
			taskType:     "database",
			scheduleType: "cron",
			expression:   "0 0 2 * * *",
			description:  "定时清理过期数据，每天凌晨2点执行",
			group:        "maintenance",
			params:       `{"query": "DELETE FROM base_operation_logs WHERE created_at < NOW() - INTERVAL '30 days'", "db_type": "postgres"}`,
		},
		{
			taskID:       "cache_refresh_tokens",
			taskType:     "cache",
			scheduleType: "cron",
			expression:   "0 */30 * * * *",
			description:  "定时刷新缓存中的过期令牌，每30分钟执行一次",
			group:        "cache",
			params:       `{"action": "clear_expired", "prefix": "token:", "ttl": 1800}`,
		},
		{
			taskID:       "script_generate_daily_report",
			taskType:     "script",
			scheduleType: "cron",
			expression:   "0 0 8 * * *",
			description:  "定时生成日报表，每天早上8点执行",
			group:        "report",
			params:       `{"command": "/bin/bash", "args": ["-c", "echo generate_daily_report"], "work_dir": "/tmp"}`,
		},
		{
			taskID:       "stream_trim_delayed_tasks_ready",
			taskType:     "stream_maintenance",
			scheduleType: "cron",
			expression:   "0 0 */6 * * *",
			description:  "定时清理 delayed_tasks_ready Stream 已消费消息，每6小时执行一次",
			group:        "maintenance",
			params:       `{"stream_key": "stream:delayed_tasks_ready", "max_length": 10000, "trim_type": "approx", "timeout": 30}`,
		},
		{
			taskID:       "stream_trim_delayed_tasks_history",
			taskType:     "stream_maintenance",
			scheduleType: "cron",
			expression:   "0 0 3 * * *",
			description:  "定时清理 delayed_tasks_history Stream 历史消息，每天凌晨3点执行",
			group:        "maintenance",
			params:       `{"stream_key": "stream:delayed_tasks_history", "max_length": 50000, "trim_type": "approx", "timeout": 30}`,
		},
		{
			taskID:       "stream_trim_event_bus_buffer",
			taskType:     "stream_maintenance",
			scheduleType: "cron",
			expression:   "0 */30 * * * *",
			description:  "定时清理 event_bus Stream 已消费消息，每30分钟执行一次",
			group:        "maintenance",
			params:       `{"stream_key": "stream:event_bus", "max_length": 5000, "trim_type": "approx", "timeout": 30}`,
		},
	}

	for _, t := range tasks {
		existing, err := ts.taskRepo.GetByTaskID(t.taskID)
		if err == nil && existing != nil {
			logger.Info("任务已存在，跳过创建", "task_id", t.taskID)
			continue
		}

		task := &entity.Task{
			TaskID:      t.taskID,
			TaskType:    t.taskType,
			Type:        t.scheduleType,
			Expression:  t.expression,
			Description: t.description,
			Group:       t.group,
			Status:      1,
			Params:      t.params,
			RetryCount:  3,
			Concurrency: "allow",
			Timeout:     30,
		}

		if err := ts.taskRepo.Create(task); err != nil {
			logger.Error("创建任务失败", "task_id", t.taskID, "error", err)
			return err
		}
		logger.Info("创建任务成功", "task_id", t.taskID, "type", t.taskType)
	}

	logger.Info("任务种子数据初始化完成")
	return nil
}
