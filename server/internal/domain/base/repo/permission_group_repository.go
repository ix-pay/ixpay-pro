package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// PermissionGroupRepository 权限组仓库接口
type PermissionGroupRepository interface {
	GetByID(id int64) (*entity.PermissionGroup, error)
	GetByName(name string) (*entity.PermissionGroup, error)
	Create(group *entity.PermissionGroup) error
	Update(group *entity.PermissionGroup) error
	Delete(id int64) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.PermissionGroup, int64, error)
	GetAllGroups() ([]*entity.PermissionGroup, error)

	// 权限组关联 API 操作
	AddAPIToGroup(groupID, apiID int64) error
	RemoveAPIFromGroup(groupID, apiID int64) error
	GetAPIsByGroup(groupID int64) ([]*entity.API, error)
	GetGroupsByAPI(apiID int64) ([]*entity.PermissionGroup, error)

	// 权限组关联按钮权限操作
	AddBtnPermToGroup(groupID, btnPermID int64) error
	RemoveBtnPermFromGroup(groupID, btnPermID int64) error
	GetBtnPermsByGroup(groupID int64) ([]*entity.BtnPerm, error)
	GetGroupsByBtnPerm(btnPermID int64) ([]*entity.PermissionGroup, error)

	// 权限组关联角色操作
	GetRolesByGroup(groupID int64) ([]*entity.Role, error)
	GetGroupsByRole(roleID int64) ([]*entity.PermissionGroup, error)
}
