package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// PermissionLogRepository 权限日志 Repository 接口
type PermissionLogRepository interface {
	// FindByID 根据 ID 查找权限日志
	FindByID(id int64) (*entity.PermissionLog, error)
	// FindByUserID 根据用户 ID 查找权限日志
	FindByUserID(userID int64, page, pageSize int) ([]*entity.PermissionLog, int64, error)
	// FindByRoleID 根据角色 ID 查找权限日志
	FindByRoleID(roleID int64, page, pageSize int) ([]*entity.PermissionLog, int64, error)
	// Create 创建权限日志
	Create(log *entity.PermissionLog) error
	// BatchDelete 批量删除权限日志
	BatchDelete(ids []int64) error
}
