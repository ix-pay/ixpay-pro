package service

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// OperationLogService 操作日志服务实现
type OperationLogService struct {
	repo repo.OperationLogRepository
	log  logger.Logger
}

// NewOperationLogService 创建操作日志服务实例
func NewOperationLogService(repo repo.OperationLogRepository, log logger.Logger) *OperationLogService {
	return &OperationLogService{
		repo: repo,
		log:  log,
	}
}

// CreateLog 创建操作日志
func (s *OperationLogService) CreateLog(log *entity.OperationLog) error {
	s.log.Info("创建操作日志", "module", log.Module, "username", log.Username)
	return s.repo.Create(log)
}

// BatchCreateLog 批量创建操作日志
func (s *OperationLogService) BatchCreateLog(logs []*entity.OperationLog) error {
	s.log.Info("批量创建操作日志", "count", len(logs))
	return s.repo.BatchCreate(logs)
}

// GetLogByID 根据 ID 获取操作日志
func (s *OperationLogService) GetLogByID(id int64) (*entity.OperationLog, error) {
	s.log.Info("根据 ID 获取操作日志", "id", id)
	return s.repo.GetByID(id)
}

// GetLogList 获取操作日志列表，支持分页和过滤
func (s *OperationLogService) GetLogList(page, pageSize int, filters map[string]interface{}) ([]*entity.OperationLog, int64, error) {
	s.log.Info("获取操作日志列表", "page", page, "pageSize", pageSize)
	list, total, err := s.repo.List(page, pageSize, filters)
	return list, total, err
}

// DeleteLogByID 根据 ID 删除操作日志
func (s *OperationLogService) DeleteLogByID(id int64) error {
	s.log.Info("根据 ID 删除操作日志", "id", id)
	return s.repo.Delete(id)
}

// BatchDeleteLog 批量删除操作日志
func (s *OperationLogService) BatchDeleteLog(ids []int64) error {
	s.log.Info("批量删除操作日志", "ids", ids)
	return s.repo.BatchDelete(ids)
}

// ClearLogByTimeRange 根据时间范围清空操作日志
func (s *OperationLogService) ClearLogByTimeRange(startTime, endTime time.Time) error {
	s.log.Info("根据时间范围清空操作日志", "startTime", startTime, "endTime", endTime)
	return s.repo.DeleteByTimeRange(startTime, endTime)
}

// GetLogStatistics 获取操作日志统计信息
func (s *OperationLogService) GetLogStatistics(startTime, endTime time.Time) (map[string]interface{}, error) {
	s.log.Info("获取操作日志统计信息", "startTime", startTime, "endTime", endTime)

	// 统计总操作数
	totalQuery := map[string]interface{}{
		"created_at between ? and ?": []interface{}{startTime, endTime},
	}
	_, total, err := s.repo.List(0, 0, totalQuery)
	if err != nil {
		return nil, err
	}

	// 统计成功操作数
	successQuery := map[string]interface{}{
		"created_at between ? and ?": []interface{}{startTime, endTime},
		"is_success = ?":             true,
	}
	_, successCount, err := s.repo.List(0, 0, successQuery)
	if err != nil {
		return nil, err
	}

	// 统计失败操作数
	failCount := total - successCount

	// 统计各操作类型数量
	typeStats := make(map[string]int64)
	operationTypes := []entity.OperationType{
		entity.OperationTypeCreate,
		entity.OperationTypeUpdate,
		entity.OperationTypeDelete,
		entity.OperationTypeQuery,
		entity.OperationTypeLogin,
		entity.OperationTypeLogout,
		entity.OperationTypeOther,
	}

	for _, opType := range operationTypes {
		opTypeQuery := map[string]interface{}{
			"created_at between ? and ?": []interface{}{startTime, endTime},
			"operation_type = ?":         opType,
		}
		_, count, err := s.repo.List(0, 0, opTypeQuery)
		if err != nil {
			return nil, err
		}
		opTypeName := s.getOperationTypeName(opType)
		typeStats[opTypeName] = int64(count)
	}

	// 统计最近7天的操作趋势
	trend := make(map[string]int64)
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		count, err := s.repo.CountByDate(date)
		if err != nil {
			return nil, err
		}
		trend[dateStr] = count
	}

	// 构建统计结果
	stats := map[string]interface{}{
		"total":     total,
		"success":   successCount,
		"fail":      failCount,
		"typeStats": typeStats,
		"trend":     trend,
		"startTime": startTime,
		"endTime":   endTime,
		"queryTime": time.Now(),
	}

	return stats, nil
}

// ExportLogs 导出操作日志
func (s *OperationLogService) ExportLogs(filters map[string]interface{}) ([]*entity.OperationLog, error) {
	s.log.Info("导出操作日志")
	// 导出所有符合条件的日志，不分页
	logs, _, err := s.repo.List(0, 0, filters)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// getOperationTypeName 获取操作类型名称
func (s *OperationLogService) getOperationTypeName(opType entity.OperationType) string {
	switch opType {
	case entity.OperationTypeCreate:
		return "创建"
	case entity.OperationTypeUpdate:
		return "更新"
	case entity.OperationTypeDelete:
		return "删除"
	case entity.OperationTypeQuery:
		return "查询"
	case entity.OperationTypeLogin:
		return "登录"
	case entity.OperationTypeLogout:
		return "登出"
	default:
		return "其他"
	}
}
