package service

import (
	"errors"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// TaskExecutionLogService 实现任务执行日志服务
type TaskExecutionLogService struct {
	repo repo.TaskExecutionLogRepository
	log  logger.Logger
}

// NewTaskExecutionLogService 创建任务执行日志服务实例
func NewTaskExecutionLogService(repo repo.TaskExecutionLogRepository, log logger.Logger) *TaskExecutionLogService {
	return &TaskExecutionLogService{
		repo: repo,
		log:  log,
	}
}

// RecordExecution 记录任务执行日志
func (s *TaskExecutionLogService) RecordExecution(
	taskID, taskName, group, cronExpr, triggerType string,
	duration int64, result string, errorInfo string,
	retryCount int, operatorID string,
) error {
	// 创建执行日志记录
	log := &entity.TaskExecutionLog{
		TaskID:      taskID,
		TaskName:    taskName,
		Group:       group,
		ExecuteAt:   time.Now().Format(time.RFC3339),
		Duration:    duration,
		Result:      result,
		ErrorInfo:   errorInfo,
		RetryCount:  retryCount,
		CronExpr:    cronExpr,
		TriggerType: triggerType,
		OperatorID:  operatorID,
	}

	// 保存到数据库
	if err := s.repo.Create(log); err != nil {
		s.log.Error("记录任务执行日志失败", "task_id", taskID, "error", err)
		return errors.New("记录任务执行日志失败")
	}

	s.log.Info("任务执行日志记录成功", "task_id", taskID, "result", result, "duration", duration)
	return nil
}

// GetExecutionHistory 查询任务执行历史
func (s *TaskExecutionLogService) GetExecutionHistory(taskID string, page, pageSize int) ([]*entity.TaskExecutionLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	logs, total, err := s.repo.GetByTaskID(taskID, page, pageSize)
	if err != nil {
		s.log.Error("查询任务执行历史失败", "task_id", taskID, "error", err)
		return nil, 0, errors.New("查询任务执行历史失败")
	}

	return logs, total, nil
}

// GetTaskStatistics 统计任务执行情况
func (s *TaskExecutionLogService) GetTaskStatistics(taskID string) (*entity.TaskStatistics, error) {
	// 获取最近执行记录
	latestLogs, err := s.repo.GetLatestByTaskID(taskID, 1)
	if err != nil {
		s.log.Error("获取任务最近执行记录失败", "task_id", taskID, "error", err)
		return nil, errors.New("获取任务统计失败")
	}

	// 获取总执行次数
	totalExecutes, err := s.repo.CountByTaskID(taskID)
	if err != nil {
		s.log.Error("统计任务执行次数失败", "task_id", taskID, "error", err)
		return nil, errors.New("获取任务统计失败")
	}

	// 获取成功次数
	successCount, err := s.repo.CountByResult("success")
	if err != nil {
		s.log.Error("统计任务成功次数失败", "task_id", taskID, "error", err)
		return nil, errors.New("获取任务统计失败")
	}

	// 计算成功率
	var successRate float64
	if totalExecutes > 0 {
		successRate = float64(successCount) / float64(totalExecutes) * 100
	}

	// 构建统计信息
	stats := &entity.TaskStatistics{
		TaskID:        taskID,
		TotalExecutes: totalExecutes,
		SuccessCount:  successCount,
		FailedCount:   totalExecutes - successCount,
		SuccessRate:   successRate,
	}

	// 填充最近执行信息
	if len(latestLogs) > 0 {
		latestLog := latestLogs[0]
		stats.TaskName = latestLog.TaskName
		stats.Group = latestLog.Group
		stats.LastExecuteAt = latestLog.ExecuteAt

		// 计算平均执行时长
		avgLogs, _ := s.repo.GetLatestByTaskID(taskID, 10)
		if len(avgLogs) > 0 {
			var totalDuration int64
			for _, log := range avgLogs {
				totalDuration += log.Duration
			}
			stats.AvgDuration = float64(totalDuration) / float64(len(avgLogs))
		}
	}

	return stats, nil
}

// GetAllTaskStatistics 获取所有任务统计
func (s *TaskExecutionLogService) GetAllTaskStatistics() ([]*entity.TaskStatistics, error) {
	// 获取所有日志记录
	logs, _, err := s.repo.List(1, 1000, map[string]interface{}{})
	if err != nil {
		s.log.Error("获取任务日志列表失败", "error", err)
		return nil, errors.New("获取所有任务统计失败")
	}

	// 按任务 ID 分组统计
	taskMap := make(map[string]*entity.TaskStatistics)
	for _, log := range logs {
		if _, exists := taskMap[log.TaskID]; !exists {
			taskMap[log.TaskID] = &entity.TaskStatistics{
				TaskID:   log.TaskID,
				TaskName: log.TaskName,
				Group:    log.Group,
			}
		}

		stats := taskMap[log.TaskID]
		stats.TotalExecutes++
		if log.Result == "success" {
			stats.SuccessCount++
		} else {
			stats.FailedCount++
		}

		// 更新最近执行时间
		if stats.LastExecuteAt == "" || log.ExecuteAt > stats.LastExecuteAt {
			stats.LastExecuteAt = log.ExecuteAt
		}
	}

	// 计算成功率和平均时长
	var result []*entity.TaskStatistics
	for _, stats := range taskMap {
		if stats.TotalExecutes > 0 {
			stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalExecutes) * 100
		}

		// 计算平均执行时长
		avgLogs, _ := s.repo.GetLatestByTaskID(stats.TaskID, 10)
		if len(avgLogs) > 0 {
			var totalDuration int64
			for _, log := range avgLogs {
				totalDuration += log.Duration
			}
			stats.AvgDuration = float64(totalDuration) / float64(len(avgLogs))
		}

		result = append(result, stats)
	}

	return result, nil
}

// ClearExpiredLogs 清理过期日志
func (s *TaskExecutionLogService) ClearExpiredLogs(days int) (int64, error) {
	if days <= 0 {
		days = 30 // 默认保留 30 天
	}

	// 计算过期日期
	beforeDate := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)

	// 清理过期日志
	deletedCount, err := s.repo.ClearExpiredLogs(beforeDate)
	if err != nil {
		s.log.Error("清理过期日志失败", "days", days, "error", err)
		return 0, errors.New("清理过期日志失败")
	}

	s.log.Info("清理过期日志成功", "days", days, "deleted_count", deletedCount)
	return deletedCount, nil
}

// GetGroupStatistics 获取任务分组统计
func (s *TaskExecutionLogService) GetGroupStatistics() ([]*entity.TaskGroupStat, error) {
	stats, err := s.repo.GetGroupStatistics()
	if err != nil {
		s.log.Error("获取任务分组统计失败", "error", err)
		return nil, errors.New("获取任务分组统计失败")
	}

	return stats, nil
}
