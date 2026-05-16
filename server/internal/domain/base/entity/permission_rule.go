package entity

import "time"

// PermissionAttribute 权限属性定义
type PermissionAttribute struct {
	Key   string // 属性键
	Value string // 属性值
	Type  string // 属性类型：user, role, resource, environment
}

// PermissionRule 权限规则领域实体 - ABAC 核心
// 纯业务模型，无 GORM 标签
type PermissionRule struct {
	ID          int64                 // 规则 ID
	Name        string                // 规则名称
	Description string                // 规则描述
	Effect      string                // 效果：allow, deny
	APIPath     string                // API 路径
	Method      string                // HTTP 方法
	Conditions  string                // 条件表达式，JSON 格式
	Attributes  []PermissionAttribute // 属性列表
	Status      int                   // 状态：1-启用，0-禁用
	Sort        int                   // 排序
	IsSystem    bool                  // 是否系统规则
	RoleIds     []int64               // 关联角色 ID 列表
	Roles       []*Role               // 关联角色对象列表
	UserIds     []int64               // 关联用户 ID 列表
	Users       []*User               // 关联用户对象列表
	CreatedBy   int64                 // 创建人 ID
	CreatedAt   time.Time             // 创建时间
	UpdatedBy   int64                 // 更新人 ID
	UpdatedAt   time.Time             // 更新时间
}

// IsActive 检查规则是否启用
func (p *PermissionRule) IsActive() bool {
	return p.Status == 1
}

// IsAllow 检查规则是否允许
func (p *PermissionRule) IsAllow() bool {
	return p.Effect == "allow"
}

// IsDeny 检查规则是否拒绝
func (p *PermissionRule) IsDeny() bool {
	return p.Effect == "deny"
}

// IsSystemRule 检查是否是系统规则
func (p *PermissionRule) IsSystemRule() bool {
	return p.IsSystem
}

// HasRole 检查规则是否包含指定角色
func (p *PermissionRule) HasRole(roleID int64) bool {
	for _, rid := range p.RoleIds {
		if rid == roleID {
			return true
		}
	}
	return false
}

// HasUser 检查规则是否包含指定用户
func (p *PermissionRule) HasUser(userID int64) bool {
	for _, uid := range p.UserIds {
		if uid == userID {
			return true
		}
	}
	return false
}
