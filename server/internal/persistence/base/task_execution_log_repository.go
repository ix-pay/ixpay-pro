package persistence

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// taskExecutionLogModel 任务执行日志数据库模型
type taskExecutionLogModel struct {
	database.SnowflakeBaseModel
	TaskID      int64  `gorm:"index;not null"`
	TaskName    string `gorm:"size:100;not null"`
	Group       string `gorm:"size:50;index"`
	ExecuteAt   string `gorm:"size:50;index"`
	Duration    int64  `gorm:"default:0"`
	Result      string `gorm:"size:20;index"`
	ErrorInfo   string `gorm:"type:text"`
	RetryCount  int    `gorm:"default:0"`
	CronExpr    string `gorm:"size:100"`
	TriggerType string `gorm:"size:20"`
	OperatorID  int64  `gorm:"index"`
}

// TableName 指定表名
func (taskExecutionLogModel) TableName() string {
	return "base_task_execution_logs"
}

// toDomain 将数据库模型转换为领域实体
func (m *taskExecutionLogModel) toDomain() *entity.TaskExecutionLog {
	if m == nil {
		return nil
	}
	return &entity.TaskExecutionLog{
		ID:          m.ID,
		TaskID:      m.TaskID,
		TaskName:    m.TaskName,
		Group:       m.Group,
		ExecuteAt:   m.ExecuteAt,
		Duration:    m.Duration,
		Result:      m.Result,
		ErrorInfo:   m.ErrorInfo,
		RetryCount:  m.RetryCount,
		CronExpr:    m.CronExpr,
		TriggerType: m.TriggerType,
		OperatorID:  m.OperatorID,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   m.UpdatedBy,
		UpdatedAt:   m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainTaskExecutionLog(log *entity.TaskExecutionLog) (*taskExecutionLogModel, error) {
	return &taskExecutionLogModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        log.ID,
			CreatedBy: log.CreatedBy,
			UpdatedBy: log.UpdatedBy,
		},
		TaskID:      log.TaskID,
		TaskName:    log.TaskName,
		Group:       log.Group,
		ExecuteAt:   log.ExecuteAt,
		Duration:    log.Duration,
		Result:      log.Result,
		ErrorInfo:   log.ErrorInfo,
		RetryCount:  log.RetryCount,
		CronExpr:    log.CronExpr,
		TriggerType: log.TriggerType,
		OperatorID:  log.OperatorID,
	}, nil
}

// taskExecutionLogRepository Repository 实现
type taskExecutionLogRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.TaskExecutionLogRepository = (*taskExecutionLogRepository)(nil)

// NewTaskExecutionLogRepository 创建任务执行日志仓库实现
func NewTaskExecutionLogRepository(db *database.PostgresDB) repo.TaskExecutionLogRepository {
	return &taskExecutionLogRepository{db: db}
}

// GetByID 根据 ID 查询任务执行日志
func (r *taskExecutionLogRepository) GetByID(id int64) (*entity.TaskExecutionLog, error) {
	var dbModel taskExecutionLogModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建任务执行日志
func (r *taskExecutionLogRepository) Create(log *entity.TaskExecutionLog) error {
	dbModel, err := fromDomainTaskExecutionLog(log)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	log.ID = dbModel.ID
	return nil
}

// Delete 删除任务执行日志
func (r *taskExecutionLogRepository) Delete(id int64) error {
	return r.db.Delete(&taskExecutionLogModel{}, id).Error
}

// List 分页查询任务执行日志列表
func (r *taskExecutionLogRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.TaskExecutionLog, int64, error) {
	var total int64
	var dbModels []taskExecutionLogModel

	query := r.db.Model(&taskExecutionLogModel{})

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("execute_at DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.TaskExecutionLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, total, nil
}

// GetByTaskID 根据任务 ID 查询执行日志
func (r *taskExecutionLogRepository) GetByTaskID(taskID int64, page, pageSize int) ([]*entity.TaskExecutionLog, int64, error) {
	var total int64
	var dbModels []taskExecutionLogModel

	query := r.db.Model(&taskExecutionLogModel{}).Where("task_id = ?", taskID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("execute_at DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.TaskExecutionLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, total, nil
}

// CountByTaskID 统计任务的执行次数
func (r *taskExecutionLogRepository) CountByTaskID(taskID int64) (int64, error) {
	var count int64
	result := r.db.Model(&taskExecutionLogModel{}).Where("task_id = ?", taskID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// CountByResult 统计指定结果的执行次数
func (r *taskExecutionLogRepository) CountByResult(result string) (int64, error) {
	var count int64
	resultErr := r.db.Model(&taskExecutionLogModel{}).Where("result = ?", result).Count(&count)
	if resultErr.Error != nil {
		return 0, resultErr.Error
	}

	return count, nil
}

// CountByTaskIDAndResult 统计指定任务和结果的执行次数
func (r *taskExecutionLogRepository) CountByTaskIDAndResult(taskID int64, result string) (int64, error) {
	var count int64
	resultErr := r.db.Model(&taskExecutionLogModel{}).Where("task_id = ? AND result = ?", taskID, result).Count(&count)
	if resultErr.Error != nil {
		return 0, resultErr.Error
	}

	return count, nil
}

// GetLatestByTaskID 获取任务最新的执行日志
func (r *taskExecutionLogRepository) GetLatestByTaskID(taskID int64, limit int) ([]*entity.TaskExecutionLog, error) {
	var dbModels []taskExecutionLogModel
	result := r.db.Where("task_id = ?", taskID).
		Order("execute_at DESC").
		Limit(limit).
		Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	logs := make([]*entity.TaskExecutionLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, nil
}

// ClearExpiredLogs 清理过期的日志
func (r *taskExecutionLogRepository) ClearExpiredLogs(beforeDate string) (int64, error) {
	result := r.db.Where("execute_at < ?", beforeDate).Delete(&taskExecutionLogModel{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// GetSuccessRateByTaskID 获取任务的成功率
func (r *taskExecutionLogRepository) GetSuccessRateByTaskID(taskID int64, days int) (float64, error) {
	var totalCount, successCount int64

	query := r.db.Model(&taskExecutionLogModel{}).Where("task_id = ?", taskID)

	// 根据 days 参数计算指定天数内的成功率
	if days > 0 {
		cutoffTime := time.Now().AddDate(0, 0, -days)
		query = query.Where("execute_at >= ?", cutoffTime)
	}

	// 获取总执行次数
	if err := query.Count(&totalCount).Error; err != nil {
		return 0, err
	}

	if totalCount == 0 {
		return 0, nil
	}

	// 获取成功次数
	if err := query.Where("result = ?", "SUCCESS").Count(&successCount).Error; err != nil {
		return 0, err
	}

	// 计算成功率
	successRate := float64(successCount) / float64(totalCount) * 100
	return successRate, nil
}

// GetGroupStatistics 获取任务分组统计
func (r *taskExecutionLogRepository) GetGroupStatistics() ([]*entity.TaskGroupStat, error) {
	type StatResult struct {
		Group         string
		TotalTasks    int64
		TotalExecutes int64
		SuccessCount  int64
		FailedCount   int64
	}

	var results []StatResult

	// 使用原生 SQL 进行分组统计
	sql := `
		SELECT 
			"group" as group,
			COUNT(DISTINCT task_id) as total_tasks,
			COUNT(*) as total_executes,
			SUM(CASE WHEN result = 'SUCCESS' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN result != 'SUCCESS' THEN 1 ELSE 0 END) as failed_count
		FROM base_task_execution_logs
		GROUP BY "group"
	`

	if err := r.db.Raw(sql).Scan(&results).Error; err != nil {
		return nil, err
	}

	// 转换为领域实体
	stats := make([]*entity.TaskGroupStat, len(results))
	for i, result := range results {
		stats[i] = &entity.TaskGroupStat{
			Group:         result.Group,
			TotalTasks:    result.TotalTasks,
			TotalExecutes: result.TotalExecutes,
			SuccessCount:  result.SuccessCount,
			FailedCount:   result.FailedCount,
			SuccessRate:   float64(result.SuccessCount) / float64(result.TotalExecutes) * 100,
		}
	}

	return stats, nil
}
