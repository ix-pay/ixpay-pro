package entity

import "time"

// Position 岗位领域实体
// 实现组织架构中的岗位管理功能
// 纯业务模型，无 GORM 标签
type Position struct {
	ID          int64     // 岗位 ID
	Name        string    // 岗位名称
	Sort        int       // 排序
	Status      int       // 状态：1-正常，0-禁用
	Description string    // 岗位描述
	CreatedBy   int64     // 创建人 ID
	CreatedAt   time.Time // 创建时间
	UpdatedBy   int64     // 更新人 ID
	UpdatedAt   time.Time // 更新时间
}

// IsActive 检查岗位是否激活
func (p *Position) IsActive() bool {
	return p.Status == 1
}
