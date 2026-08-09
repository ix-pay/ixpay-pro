package entity

import "time"

// ConditionOperator 条件运算符
type ConditionOperator string

const (
	OperatorEQ    ConditionOperator = "EQ"    // 等于
	OperatorNEQ   ConditionOperator = "NEQ"   // 不等于
	OperatorIN    ConditionOperator = "IN"    // 包含于
	OperatorNotIN ConditionOperator = "NOT_IN" // 不包含于
	OperatorGT    ConditionOperator = "GT"    // 大于
	OperatorGTE   ConditionOperator = "GTE"   // 大于等于
	OperatorLT    ConditionOperator = "LT"    // 小于
	OperatorLTE   ConditionOperator = "LTE"   // 小于等于
	OperatorLIKE  ConditionOperator = "LIKE"  // 模糊匹配
)

// LogicOperator 逻辑运算符
type LogicOperator string

const (
	LogicAND LogicOperator = "AND" // 逻辑与
	LogicOR  LogicOperator = "OR"  // 逻辑或
	LogicNOT LogicOperator = "NOT" // 逻辑非
)

// PermissionAttribute 权限属性定义
type PermissionAttribute struct {
	Key      string            // 属性键
	Value    string            // 属性值
	Type     string            // 属性类型：user, role, resource, environment
	Operator ConditionOperator // 条件运算符（默认 EQ）
}

// ConditionNode 条件节点 - 支持嵌套的 AND/OR/NOT 逻辑
type ConditionNode struct {
	Operator   LogicOperator    `json:"operator,omitempty"`   // 逻辑运算符（AND/OR/NOT）
	Conditions []ConditionNode  `json:"conditions,omitempty"` // 子条件列表（嵌套条件）
	Attribute  *PermissionAttribute `json:"attribute,omitempty"` // 原子条件
	Not        bool             `json:"not,omitempty"`        // 是否取反
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
	Conditions  string                // 条件表达式，JSON 格式（支持新旧两种格式）
	Attributes  []PermissionAttribute // 属性列表（兼容旧格式）
	ConditionNode *ConditionNode      // 结构化条件节点（新格式）
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

// DataScope 数据权限范围
type DataScope int

const (
	DataScopeAll       DataScope = 1 // 全部数据
	DataScopeDept      DataScope = 2 // 本部门数据
	DataScopeDeptChild DataScope = 3 // 本部门及子部门数据
	DataScopeSelf      DataScope = 4 // 仅本人数据
	DataScopeCustom    DataScope = 5 // 自定义数据
)

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
