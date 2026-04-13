package persistence

import (
	"strconv"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// operationLogModel 操作日志数据库模型
type operationLogModel struct {
	database.SnowflakeBaseModel
	UserID        int64     `gorm:"index"`
	Username      string    `gorm:"size:50"`
	Nickname      string    `gorm:"size:50"`
	OperationType int       `gorm:"default:1"`
	Module        string    `gorm:"size:50"`
	Description   string    `gorm:"size:500"`
	Method        string    `gorm:"size:10"`
	Path          string    `gorm:"size:255"`
	Params        string    `gorm:"type:text"`
	ClientIP      string    `gorm:"size:50"`
	UserAgent     string    `gorm:"size:500"`
	StatusCode    int       `gorm:"default:200"`
	Result        string    `gorm:"type:text"`
	Duration      int64     `gorm:"default:0"`
	ErrorMessage  string    `gorm:"size:1000"`
	IsSuccess     bool      `gorm:"default:true"`
	ExecuteTime   time.Time `gorm:"index"`
}

// TableName 指定表名
func (operationLogModel) TableName() string {
	return "base_operation_logs"
}

// toDomain 将数据库模型转换为领域实体
func (m *operationLogModel) toDomain() *entity.OperationLog {
	if m == nil {
		return nil
	}
	return &entity.OperationLog{
		ID:            common.ToString(m.ID),
		UserID:        common.ToString(m.UserID),
		Username:      m.Username,
		Nickname:      m.Nickname,
		OperationType: entity.OperationType(m.OperationType),
		Module:        m.Module,
		Description:   m.Description,
		Method:        m.Method,
		Path:          m.Path,
		Params:        m.Params,
		ClientIP:      m.ClientIP,
		UserAgent:     m.UserAgent,
		StatusCode:    m.StatusCode,
		Result:        m.Result,
		Duration:      m.Duration,
		ErrorMessage:  m.ErrorMessage,
		IsSuccess:     m.IsSuccess,
		CreatedBy:     common.ToString(m.CreatedBy),
		CreatedAt:     m.CreatedAt,
		UpdatedBy:     common.ToString(m.UpdatedBy),
		UpdatedAt:     m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainOperationLog(log *entity.OperationLog) (*operationLogModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(log.ID, log.CreatedBy, log.UpdatedBy)

	return &operationLogModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		UserID:        common.TryParseInt64(log.UserID),
		Username:      log.Username,
		Nickname:      log.Nickname,
		OperationType: int(log.OperationType),
		Module:        log.Module,
		Description:   log.Description,
		Method:        log.Method,
		Path:          log.Path,
		Params:        log.Params,
		ClientIP:      log.ClientIP,
		UserAgent:     log.UserAgent,
		StatusCode:    log.StatusCode,
		Result:        log.Result,
		Duration:      log.Duration,
		ErrorMessage:  log.ErrorMessage,
		IsSuccess:     log.IsSuccess,
		ExecuteTime:   time.Now(),
	}, nil
}

// operationLogRepository Repository 实现
type operationLogRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.OperationLogRepository = (*operationLogRepository)(nil)

// NewOperationLogRepository 创建操作日志仓库实现
func NewOperationLogRepository(db *database.PostgresDB) repo.OperationLogRepository {
	return &operationLogRepository{db: db}
}

// Create 创建操作日志
func (r *operationLogRepository) Create(log *entity.OperationLog) error {
	dbModel, err := fromDomainOperationLog(log)
	if err != nil {
		return err
	}

	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}

	// 将生成的 ID 回写到领域实体
	log.ID = common.ToString(dbModel.ID)
	return nil
}

// BatchCreate 批量创建操作日志
func (r *operationLogRepository) BatchCreate(logs []*entity.OperationLog) error {
	if len(logs) == 0 {
		return nil
	}

	dbModels := make([]*operationLogModel, 0, len(logs))
	for _, log := range logs {
		dbModel, err := fromDomainOperationLog(log)
		if err != nil {
			return err
		}
		dbModels = append(dbModels, dbModel)
	}

	return r.db.CreateInBatches(dbModels, 100).Error
}

// GetByID 根据 ID 查询操作日志
func (r *operationLogRepository) GetByID(id string) (*entity.OperationLog, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var dbModel operationLogModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// List 分页查询操作日志列表
func (r *operationLogRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.OperationLog, int64, error) {
	var total int64
	var dbModels []operationLogModel

	query := r.db.Model(&operationLogModel{})

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("execute_time DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.OperationLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, total, nil
}

// Delete 删除操作日志
func (r *operationLogRepository) Delete(id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	return r.db.Delete(&operationLogModel{}, intID).Error
}

// DeleteByID 根据 ID 删除操作日志
func (r *operationLogRepository) DeleteByID(id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	return r.db.Delete(&operationLogModel{}, intID).Error
}

// BatchDelete 批量删除操作日志
func (r *operationLogRepository) BatchDelete(ids []string) error {
	intIDs, err := common.StringToInt64s(ids)
	if err != nil {
		return err
	}

	return r.db.Where("id IN ?", intIDs).Delete(&operationLogModel{}).Error
}

// DeleteByTimeRange 根据时间范围删除操作日志
func (r *operationLogRepository) DeleteByTimeRange(startTime, endTime time.Time) error {
	return r.db.Where("execute_time >= ? AND execute_time <= ?", startTime, endTime).
		Delete(&operationLogModel{}).Error
}

// CountByDate 统计指定日期的操作日志数
func (r *operationLogRepository) CountByDate(date time.Time) (int64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var count int64
	result := r.db.Model(&operationLogModel{}).
		Where("execute_time >= ? AND execute_time < ?", startOfDay, endOfDay).
		Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// CountByModule 统计指定模块的操作日志数
func (r *operationLogRepository) CountByModule(module string) (int64, error) {
	var count int64
	result := r.db.Model(&operationLogModel{}).Where("module = ?", module).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// CountByUser 统计用户的操作日志数
func (r *operationLogRepository) CountByUser(userID string) (int64, error) {
	intUserID, _ := strconv.ParseInt(userID, 10, 64)
	var count int64
	result := r.db.Model(&operationLogModel{}).Where("user_id = ?", intUserID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// CountByOperationType 统计指定操作类型的日志数
func (r *operationLogRepository) CountByOperationType(operationType entity.OperationType) (int64, error) {
	var count int64
	result := r.db.Model(&operationLogModel{}).Where("operation_type = ?", int(operationType)).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// CountByResult 统计指定结果的日志数
func (r *operationLogRepository) CountByResult(isSuccess bool) (int64, error) {
	var count int64
	result := r.db.Model(&operationLogModel{}).Where("is_success = ?", isSuccess).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}
