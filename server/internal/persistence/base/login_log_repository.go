package persistence

import (
	"strconv"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// loginLogModel 登录日志数据库模型
type loginLogModel struct {
	database.SnowflakeBaseModel
	UserID     int64     `gorm:"index"`
	Username   string    `gorm:"size:50;not null"`
	LoginIP    string    `gorm:"size:50"`
	LoginTime  time.Time `gorm:"index"`
	LoginPlace string    `gorm:"size:100"`
	Device     string    `gorm:"size:50"`
	Browser    string    `gorm:"size:50"`
	OS         string    `gorm:"size:50"`
	Result     int       `gorm:"default:1"`
	ErrorMsg   string    `gorm:"size:500"`
	UserAgent  string    `gorm:"size:500"`
}

// TableName 指定表名
func (loginLogModel) TableName() string {
	return "base_login_logs"
}

// toDomain 将数据库模型转换为领域实体
func (m *loginLogModel) toDomain() *entity.LoginLog {
	if m == nil {
		return nil
	}
	return &entity.LoginLog{
		ID:         common.ToString(m.ID),
		UserID:     common.ToString(m.UserID),
		Username:   m.Username,
		LoginIP:    m.LoginIP,
		LoginTime:  m.LoginTime,
		LoginPlace: m.LoginPlace,
		Device:     m.Device,
		Browser:    m.Browser,
		OS:         m.OS,
		Result:     entity.LoginResult(m.Result),
		ErrorMsg:   m.ErrorMsg,
		UserAgent:  m.UserAgent,
		CreatedBy:  common.ToString(m.CreatedBy),
		CreatedAt:  m.CreatedAt,
		UpdatedBy:  common.ToString(m.UpdatedBy),
		UpdatedAt:  m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainLoginLog(log *entity.LoginLog) (*loginLogModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(log.ID, log.CreatedBy, log.UpdatedBy)

	return &loginLogModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		UserID:     common.TryParseInt64(log.UserID),
		Username:   log.Username,
		LoginIP:    log.LoginIP,
		LoginTime:  log.LoginTime,
		LoginPlace: log.LoginPlace,
		Device:     log.Device,
		Browser:    log.Browser,
		OS:         log.OS,
		Result:     int(log.Result),
		ErrorMsg:   log.ErrorMsg,
		UserAgent:  log.UserAgent,
	}, nil
}

// loginLogRepository Repository 实现
type loginLogRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.LoginLogRepository = (*loginLogRepository)(nil)

// NewLoginLogRepository 创建登录日志仓库实现
func NewLoginLogRepository(db *database.PostgresDB) repo.LoginLogRepository {
	return &loginLogRepository{db: db}
}

// Create 创建登录日志
func (r *loginLogRepository) Create(log *entity.LoginLog) error {
	dbModel, err := fromDomainLoginLog(log)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
}

// GetByID 根据 ID 查询登录日志
func (r *loginLogRepository) GetByID(id string) (*entity.LoginLog, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var dbModel loginLogModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// List 分页查询登录日志列表
func (r *loginLogRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.LoginLog, int64, error) {
	var total int64
	var dbModels []loginLogModel

	query := r.db.Model(&loginLogModel{})

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("login_time DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.LoginLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, total, nil
}

// GetByUserID 根据用户 ID 查询登录日志
func (r *loginLogRepository) GetByUserID(userID string, page, pageSize int) ([]*entity.LoginLog, int64, error) {
	intUserID, _ := strconv.ParseInt(userID, 10, 64)
	var total int64
	var dbModels []loginLogModel

	query := r.db.Model(&loginLogModel{}).Where("user_id = ?", intUserID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("login_time DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.LoginLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, total, nil
}

// GetByIP 根据 IP 查询登录日志
func (r *loginLogRepository) GetByIP(ip string, page, pageSize int) ([]*entity.LoginLog, int, error) {
	var total int64
	var dbModels []loginLogModel

	query := r.db.Model(&loginLogModel{}).Where("login_ip = ?", ip)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("login_time DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.LoginLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, int(total), nil
}

// GetFailedByIP 获取指定 IP 的失败登录记录
func (r *loginLogRepository) GetFailedByIP(ip string, hours int) ([]*entity.LoginLog, error) {
	var dbModels []loginLogModel
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	result := r.db.Where("login_ip = ? AND result = ? AND login_time >= ?", ip, 0, since).
		Order("login_time DESC").
		Find(&dbModels)
	if result.Error != nil {
		return nil, result.Error
	}

	logs := make([]*entity.LoginLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, nil
}

// GetByTimeRange 根据时间范围查询登录日志
func (r *loginLogRepository) GetByTimeRange(startTime, endTime time.Time, page, pageSize int) ([]*entity.LoginLog, int, error) {
	var total int64
	var dbModels []loginLogModel

	query := r.db.Model(&loginLogModel{}).
		Where("login_time >= ? AND login_time <= ?", startTime, endTime)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("login_time DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.LoginLog, len(dbModels))
	for i, model := range dbModels {
		logs[i] = model.toDomain()
	}

	return logs, int(total), nil
}

// CountByDate 统计指定日期的登录次数
func (r *loginLogRepository) CountByDate(date time.Time) (int64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var count int64
	result := r.db.Model(&loginLogModel{}).
		Where("login_time >= ? AND login_time < ?", startOfDay, endOfDay).
		Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// CountByUser 统计用户的登录次数
func (r *loginLogRepository) CountByUser(userID string) (int64, error) {
	intUserID, _ := strconv.ParseInt(userID, 10, 64)
	var count int64
	result := r.db.Model(&loginLogModel{}).Where("user_id = ?", intUserID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// CountByResult 统计指定结果的登录次数
func (r *loginLogRepository) CountByResult(result entity.LoginResult) (int64, error) {
	var count int64
	query := r.db.Model(&loginLogModel{}).Where("result = ?", int(result))
	resultErr := query.Count(&count)
	if resultErr.Error != nil {
		return 0, resultErr.Error
	}

	return count, nil
}

// GetStatistics 获取登录统计信息
func (r *loginLogRepository) GetStatistics(startTime, endTime time.Time) (*entity.LoginStatistics, error) {
	// TODO: 实现登录统计信息
	return &entity.LoginStatistics{}, nil
}
