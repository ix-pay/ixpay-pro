package entity

import "time"

// Dict 字典领域实体
// 纯业务模型，无 GORM 标签
type Dict struct {
	ID          int64      // 字典 ID
	DictName    string     // 字典名称
	DictCode    string     // 字典编码
	Description string     // 描述
	Status      int        // 状态：1-启用 0-禁用
	DictItems   []DictItem // 字典项列表（一对多关系）
	CreatedBy   int64      // 创建人 ID
	CreatedAt   time.Time  // 创建时间
	UpdatedBy   int64      // 更新人 ID
	UpdatedAt   time.Time  // 更新时间
}

// IsActive 检查字典是否启用
func (d *Dict) IsActive() bool {
	return d.Status == 1
}

// DictItem 字典项领域实体
// 纯业务模型，无 GORM 标签
type DictItem struct {
	ID          int64     // 字典项 ID
	DictID      int64     // 字典 ID
	ItemKey     string    // 字典项键
	ItemValue   string    // 字典项值
	Sort        int       // 排序
	Description string    // 描述
	Status      int       // 状态：1-启用 0-禁用
	CreatedBy   int64     // 创建人 ID
	CreatedAt   time.Time // 创建时间
	UpdatedBy   int64     // 更新人 ID
	UpdatedAt   time.Time // 更新时间
}

// IsActive 检查字典项是否启用
func (di *DictItem) IsActive() bool {
	return di.Status == 1
}
