package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

// permissionLogModel 权限日志数据库模型
type permissionLogModel struct {
	database.SnowflakeBaseModelWithoutDeleted
	UserID     int64  `gorm:"column:user_id"`
	Username   string `gorm:"column:user_name;size:64"`
	Operation  string `gorm:"column:operation;size:128"`
	Module     string `gorm:"column:module;size:64"`
	TargetType string `gorm:"column:target_type;size:64"`
	TargetID   int64  `gorm:"column:target_id"`
	OldValue   string `gorm:"column:old_value;type:text"`
	NewValue   string `gorm:"column:new_value;type:text"`
	IP         string `gorm:"column:ip;size:64"`
	UserAgent  string `gorm:"column:user_agent;size:512"`
}

// TableName 指定表名
func (permissionLogModel) TableName() string {
	return "sys_permission_logs"
}

// toDomain 将数据库模型转换为领域实体
func (m *permissionLogModel) toDomain() *entity.PermissionLog {
	if m == nil {
		return nil
	}
	return &entity.PermissionLog{
		SnowflakeBaseModelWithoutDeleted: database.SnowflakeBaseModelWithoutDeleted{
			ID:        m.ID,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
		UserID:     m.UserID,
		Username:   m.Username,
		Operation:  m.Operation,
		Module:     m.Module,
		TargetType: m.TargetType,
		TargetID:   m.TargetID,
		OldValue:   m.OldValue,
		NewValue:   m.NewValue,
		IP:         m.IP,
		UserAgent:  m.UserAgent,
	}
}

// fromDomainPermissionLog 将领域实体转换为数据库模型
func fromDomainPermissionLog(log *entity.PermissionLog) (*permissionLogModel, error) {
	return &permissionLogModel{
		SnowflakeBaseModelWithoutDeleted: database.SnowflakeBaseModelWithoutDeleted{
			ID:        log.ID,
			CreatedBy: log.CreatedBy,
			UpdatedBy: log.UpdatedBy,
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		UserID:     log.UserID,
		Username:   log.Username,
		Operation:  log.Operation,
		Module:     log.Module,
		TargetType: log.TargetType,
		TargetID:   log.TargetID,
		OldValue:   log.OldValue,
		NewValue:   log.NewValue,
		IP:         log.IP,
		UserAgent:  log.UserAgent,
	}, nil
}

// permissionLogRepository 权限日志 Repository 实现
type permissionLogRepository struct {
	db  *database.PostgresDB
	log logger.Logger
}

// NewPermissionLogRepository 创建权限日志 Repository 实例
func NewPermissionLogRepository(db *database.PostgresDB, log logger.Logger) repo.PermissionLogRepository {
	return &permissionLogRepository{
		db:  db,
		log: log,
	}
}

// FindByID 根据 ID 查找权限日志
func (r *permissionLogRepository) FindByID(id int64) (*entity.PermissionLog, error) {
	var model permissionLogModel
	result := r.db.Where("id = ?", id).First(&model)
	if result.Error != nil {
		return nil, result.Error
	}
	return model.toDomain(), nil
}

// FindByUserID 根据用户 ID 查找权限日志
func (r *permissionLogRepository) FindByUserID(userID int64, page, pageSize int) ([]*entity.PermissionLog, int64, error) {
	var models []permissionLogModel
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&permissionLogModel{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC, id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.PermissionLog, len(models))
	for i, model := range models {
		logs[i] = model.toDomain()
	}

	return logs, total, nil
}

// FindByRoleID 根据角色 ID 查找权限日志
func (r *permissionLogRepository) FindByRoleID(roleID int64, page, pageSize int) ([]*entity.PermissionLog, int64, error) {
	var models []permissionLogModel
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&permissionLogModel{}).Where("target_type = ? AND target_id = ?", "role", roleID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC, id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*entity.PermissionLog, len(models))
	for i, model := range models {
		logs[i] = model.toDomain()
	}

	return logs, total, nil
}

// Create 创建权限日志
func (r *permissionLogRepository) Create(log *entity.PermissionLog) error {
	dbModel, err := fromDomainPermissionLog(log)
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

// BatchDelete 批量删除权限日志
func (r *permissionLogRepository) BatchDelete(ids []int64) error {
	return r.db.Where("id IN ?", ids).Delete(&permissionLogModel{}).Error
}
