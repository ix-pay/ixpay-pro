package entity

import "time"

// Department 部门领域实体
// 实现组织架构中的部门管理功能
// 纯业务模型，无 GORM 标签
type Department struct {
	ID          string        // 部门 ID
	Name        string        // 部门名称
	ParentID    string        // 父部门 ID
	LeaderID    string        // 部门负责人 ID
	Sort        int           // 排序
	Status      int           // 状态：1-正常，0-禁用
	Description string        // 部门描述
	Children    []*Department // 子部门
	Parent      *Department   // 父部门（新增）
	Leader      *User         // 部门负责人（新增）
	CreatedBy   string        // 创建人 ID
	CreatedAt   time.Time     // 创建时间
	UpdatedBy   string        // 更新人 ID
	UpdatedAt   time.Time     // 更新时间
}

// IsActive 检查部门是否激活
func (d *Department) IsActive() bool {
	return d.Status == 1
}

// IsChildOf 检查当前部门是否是指定部门的子部门
func (d *Department) IsChildOf(parentID string) bool {
	return d.ParentID == parentID
}

// HasChildren 检查部门是否有子部门
func (d *Department) HasChildren() bool {
	return len(d.Children) > 0
}
