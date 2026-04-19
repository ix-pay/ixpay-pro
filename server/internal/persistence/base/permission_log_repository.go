package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

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
	var log entity.PermissionLog
	result := r.db.Where("id = ?", id).First(&log)
	return &log, result.Error
}

// FindByUserID 根据用户 ID 查找权限日志
func (r *permissionLogRepository) FindByUserID(userID int64, page, pageSize int) ([]*entity.PermissionLog, int64, error) {
	var logs []*entity.PermissionLog
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&entity.PermissionLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// FindByRoleID 根据角色 ID 查找权限日志
func (r *permissionLogRepository) FindByRoleID(roleID int64, page, pageSize int) ([]*entity.PermissionLog, int64, error) {
	var logs []*entity.PermissionLog
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&entity.PermissionLog{}).Where("target_type = ? AND target_id = ?", "role", roleID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Create 创建权限日志
func (r *permissionLogRepository) Create(log *entity.PermissionLog) error {
	return r.db.Create(log).Error
}

// BatchDelete 批量删除权限日志
func (r *permissionLogRepository) BatchDelete(ids []int64) error {
	return r.db.Where("id IN ?", ids).Delete(&entity.PermissionLog{}).Error
}
