package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

// PermissionGroupRepository 权限组仓库接口
type PermissionGroupRepository interface {
	GetByID(id string) (*entity.PermissionGroup, error)
	GetByName(name string) (*entity.PermissionGroup, error)
	Create(group *entity.PermissionGroup) error
	Update(group *entity.PermissionGroup) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]interface{}) ([]*entity.PermissionGroup, int64, error)
	GetAllGroups() ([]*entity.PermissionGroup, error)

	// 权限组关联 API 操作
	AddAPIToGroup(groupID, apiID string) error
	RemoveAPIFromGroup(groupID, apiID string) error
	GetAPIsByGroup(groupID string) ([]*entity.API, error)
	GetGroupsByAPI(apiID string) ([]*entity.PermissionGroup, error)

	// 权限组关联按钮权限操作
	AddBtnPermToGroup(groupID, btnPermID string) error
	RemoveBtnPermFromGroup(groupID, btnPermID string) error
	GetBtnPermsByGroup(groupID string) ([]*entity.BtnPerm, error)
	GetGroupsByBtnPerm(btnPermID string) ([]*entity.PermissionGroup, error)

	// 权限组关联角色操作
	GetRolesByGroup(groupID string) ([]*entity.Role, error)
	GetGroupsByRole(roleID string) ([]*entity.PermissionGroup, error)
}
