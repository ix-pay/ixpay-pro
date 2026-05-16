package task

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// TaskTemplate 任务模板
type TaskTemplate struct {
	helper *TaskHelper
	log    logger.Logger
}

// NewTaskTemplate 创建任务模板
func NewTaskTemplate(helper *TaskHelper, log logger.Logger) *TaskTemplate {
	return &TaskTemplate{
		helper: helper,
		log:    log,
	}
}

// HealthCheckTemplate 健康检查任务模板
// 定期调用指定 URL 检查服务健康状态
func (t *TaskTemplate) HealthCheckTemplate(
	taskID string,
	url string,
	intervalMinutes int,
) error {
	params := map[string]interface{}{
		"url":             url,
		"method":          "GET",
		"timeout":         5,
		"expected_status": 200,
	}

	cronExpr := fmt.Sprintf("0 */%d * * * *", intervalMinutes)
	return t.helper.CreateCronTask(taskID, TaskTypeHTTP, cronExpr, params,
		WithGroup("monitoring"),
		WithDescription(fmt.Sprintf("健康检查: %s", url)),
	)
}

// DataCleanupTemplate 数据清理任务模板
// 定期清理指定表中的过期数据
func (t *TaskTemplate) DataCleanupTemplate(
	taskID string,
	tableName string,
	daysToKeep int,
	cronExpr string,
) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE created_at < NOW() - INTERVAL '%d days'", tableName, daysToKeep)
	params := map[string]interface{}{
		"query":   query,
		"db_type": "postgres",
	}

	return t.helper.CreateCronTask(taskID, TaskTypeDatabase, cronExpr, params,
		WithGroup("maintenance"),
		WithDescription(fmt.Sprintf("清理 %s 表 %d 天前数据", tableName, daysToKeep)),
		WithTimeout(30*time.Minute),
	)
}

// CacheRefreshTemplate 缓存刷新任务模板
// 定期刷新指定的缓存键
func (t *TaskTemplate) CacheRefreshTemplate(
	taskID string,
	cacheKeys []string,
	intervalMinutes int,
) error {
	params := map[string]interface{}{
		"action":     "refresh",
		"cache_keys": cacheKeys,
		"ttl":        3600,
	}

	cronExpr := fmt.Sprintf("0 */%d * * * *", intervalMinutes)
	return t.helper.CreateCronTask(taskID, TaskTypeCache, cronExpr, params,
		WithGroup("cache"),
		WithDescription(fmt.Sprintf("刷新缓存: %v", cacheKeys)),
	)
}

// ReportGenerateTemplate 报表生成任务模板
// 定期生成指定类型的报表
func (t *TaskTemplate) ReportGenerateTemplate(
	taskID string,
	reportType string,
	cronExpr string,
) error {
	scriptPath := fmt.Sprintf("/scripts/generate_%s_report.sh", reportType)
	params := map[string]interface{}{
		"command":  "bash",
		"args":     []string{scriptPath},
		"timeout":  300,
		"work_dir": "/opt/app",
	}

	return t.helper.CreateCronTask(taskID, TaskTypeScript, cronExpr, params,
		WithGroup("report"),
		WithDescription(fmt.Sprintf("生成 %s 报表", reportType)),
		WithTimeout(10*time.Minute),
	)
}

// PaymentTimeoutTemplate 支付超时任务模板
// 添加支付超时延迟任务（注意：这是延迟任务，不是定时任务）
func (t *TaskTemplate) PaymentTimeoutTemplate(
	orderID string,
	timeoutMinutes int,
) error {
	return t.helper.AddPaymentTimeoutTask(orderID, time.Duration(timeoutMinutes)*time.Minute)
}

// EmailNotificationTemplate 邮件通知任务模板
// 定期执行邮件通知脚本
func (t *TaskTemplate) EmailNotificationTemplate(
	taskID string,
	cronExpr string,
) error {
	params := map[string]interface{}{
		"command":  "python3",
		"args":     []string{"/scripts/send_email_notification.py"},
		"timeout":  60,
		"work_dir": "/opt/app",
		"env": map[string]string{
			"SMTP_HOST": "smtp.example.com",
			"SMTP_PORT": "587",
		},
	}

	return t.helper.CreateCronTask(taskID, TaskTypeScript, cronExpr, params,
		WithGroup("notification"),
		WithDescription("邮件通知任务"),
		WithTimeout(2*time.Minute),
	)
}

// BackupDatabaseTemplate 数据库备份任务模板
// 定期执行数据库备份脚本
func (t *TaskTemplate) BackupDatabaseTemplate(
	taskID string,
	backupPath string,
	cronExpr string,
) error {
	params := map[string]interface{}{
		"command":  "bash",
		"args":     []string{"/scripts/backup_database.sh", backupPath},
		"timeout":  600,
		"work_dir": "/opt/app",
		"env": map[string]string{
			"BACKUP_PATH": backupPath,
			"DB_HOST":     "localhost",
			"DB_PORT":     "5432",
		},
	}

	return t.helper.CreateCronTask(taskID, TaskTypeScript, cronExpr, params,
		WithGroup("backup"),
		WithDescription(fmt.Sprintf("数据库备份: %s", backupPath)),
		WithTimeout(20*time.Minute),
	)
}

// SyncExternalDataTemplate 外部数据同步任务模板
// 定期从外部 API 同步数据
func (t *TaskTemplate) SyncExternalDataTemplate(
	taskID string,
	apiURL string,
	cronExpr string,
) error {
	params := map[string]interface{}{
		"url":    apiURL,
		"method": "GET",
		"headers": map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer ${API_TOKEN}",
		},
		"timeout":         30,
		"expected_status": 200,
	}

	return t.helper.CreateCronTask(taskID, TaskTypeHTTP, cronExpr, params,
		WithGroup("sync"),
		WithDescription(fmt.Sprintf("同步外部数据: %s", apiURL)),
		WithTimeout(5*time.Minute),
	)
}

// LogArchiveTemplate 日志归档任务模板
// 定期归档旧日志文件
func (t *TaskTemplate) LogArchiveTemplate(
	taskID string,
	cronExpr string,
) error {
	params := map[string]interface{}{
		"command":  "bash",
		"args":     []string{"/scripts/archive_logs.sh"},
		"timeout":  300,
		"work_dir": "/opt/app",
	}

	return t.helper.CreateCronTask(taskID, TaskTypeScript, cronExpr, params,
		WithGroup("maintenance"),
		WithDescription("日志归档任务"),
		WithTimeout(10*time.Minute),
	)
}

// MetricsCollectionTemplate 指标收集任务模板
// 定期收集系统指标
func (t *TaskTemplate) MetricsCollectionTemplate(
	taskID string,
	metricsURL string,
	intervalMinutes int,
) error {
	cronExpr := fmt.Sprintf("0 */%d * * * *", intervalMinutes)
	params := map[string]interface{}{
		"url":             metricsURL,
		"method":          "GET",
		"timeout":         10,
		"expected_status": 200,
	}

	return t.helper.CreateCronTask(taskID, TaskTypeHTTP, cronExpr, params,
		WithGroup("monitoring"),
		WithDescription(fmt.Sprintf("收集指标: %s", metricsURL)),
	)
}

// BatchProcessTemplate 批量处理任务模板
// 定期执行批量数据处理
func (t *TaskTemplate) BatchProcessTemplate(
	taskID string,
	query string,
	cronExpr string,
) error {
	params := map[string]interface{}{
		"query":   query,
		"db_type": "postgres",
	}

	return t.helper.CreateCronTask(taskID, TaskTypeDatabase, cronExpr, params,
		WithGroup("batch"),
		WithDescription("批量数据处理"),
		WithTimeout(30*time.Minute),
	)
}

// WebhookNotificationTemplate Webhook 通知任务模板
// 定期发送 Webhook 通知
func (t *TaskTemplate) WebhookNotificationTemplate(
	taskID string,
	webhookURL string,
	payload map[string]interface{},
	cronExpr string,
) error {
	payloadJSON, _ := json.Marshal(payload)
	params := map[string]interface{}{
		"url":    webhookURL,
		"method": "POST",
		"headers": map[string]string{
			"Content-Type": "application/json",
		},
		"body":            string(payloadJSON),
		"timeout":         10,
		"expected_status": 200,
	}

	return t.helper.CreateCronTask(taskID, TaskTypeHTTP, cronExpr, params,
		WithGroup("notification"),
		WithDescription(fmt.Sprintf("Webhook 通知: %s", webhookURL)),
	)
}

// StreamTrimTemplate Stream 裁剪任务模板
// 定时清理 Redis Stream 已消费消息，防止内存增长
func (t *TaskTemplate) StreamTrimTemplate(
	taskID string,
	streamKey string,
	maxLength int64,
	cronExpr string,
) error {
	params := map[string]interface{}{
		"stream_key":  streamKey,
		"max_length":  maxLength,
		"trim_type":   "approx",
		"timeout":     30,
	}

	return t.helper.CreateCronTask(taskID, TaskTypeStreamMaintenance, cronExpr, params,
		WithGroup("maintenance"),
		WithDescription(fmt.Sprintf("裁剪 Stream: %s, 最大长度: %d", streamKey, maxLength)),
	)
}
