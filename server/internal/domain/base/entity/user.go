package entity

import "time"

// User 用户领域实体
// 表示系统中的用户基本信息
// 包含身份认证、个人资料、组织架构和权限相关字段
// 纯业务模型，无 GORM 标签
type User struct {
	ID                      int64       // 用户 ID
	Username                string      // 用户名，唯一标识
	PasswordHash            string      // 密码哈希值
	Nickname                string      // 用户昵称
	Email                   string      // 电子邮箱
	Phone                   string      // 手机号码
	Avatar                  string      // 用户头像 URL
	Status                  int         // 用户状态，1-正常，0-禁用
	Gender                  int         // 性别：0-未知，1-男，2-女
	Birthday                string      // 生日
	Address                 string      // 地址
	PositionID              int64       // 岗位 ID
	DepartmentID            int64       // 部门 ID
	EntryDate               string      // 入职日期
	LastLoginIP             string      // 最后登录 IP
	LastLoginTime           string      // 最后登录时间
	WechatOpenID            string      // 微信 OpenID，唯一
	CreatedBy               int64       // 创建人 ID
	CreatedAt               time.Time   // 创建时间
	UpdatedBy               int64       // 更新人 ID
	UpdatedAt               time.Time   // 更新时间
	RoleIds                 []int64     // 用户关联的角色 ID 列表
	Roles                   []*Role     // 角色列表
	Department              *Department // 所属部门
	Position                *Position   // 所属岗位
	SpecialPermissionIds    []int64     // 用户特殊权限 ID 列表
	SpecialBtnPermissionIds []int64     // 用户特殊按钮权限 ID 列表
}

// HasRole 检查用户是否拥有指定角色
func (u *User) HasRole(roleID int64) bool {
	for _, rid := range u.RoleIds {
		if rid == roleID {
			return true
		}
	}
	return false
}

// AddRole 为用户添加角色
func (u *User) AddRole(roleID int64) {
	if !u.HasRole(roleID) {
		u.RoleIds = append(u.RoleIds, roleID)
	}
}

// RemoveRole 移除用户角色
func (u *User) RemoveRole(roleID int64) {
	newRoles := make([]int64, 0, len(u.RoleIds))
	for _, rid := range u.RoleIds {
		if rid != roleID {
			newRoles = append(newRoles, rid)
		}
	}
	u.RoleIds = newRoles
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == 1
}

// IsAdmin 检查用户是否是管理员（通过角色判断）
func (u *User) IsAdmin(adminRoleID int64) bool {
	return u.HasRole(adminRoleID)
}
