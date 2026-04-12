package request

// DictRequest 字典基础请求模型
type DictRequest struct {
	DictName    string `json:"dictName" binding:"required,max=100"` // 字典名称
	DictCode    string `json:"dictCode" binding:"required,max=50"`  // 字典编码
	Description string `json:"description" binding:"max=255"`       // 描述
	Status      int    `json:"status" binding:"oneof=0 1"`          // 状态：1-启用 0-禁用
}

// CreateDictRequest 创建字典请求模型
type CreateDictRequest struct {
	DictRequest
}

// UpdateDictRequest 更新字典请求模型
type UpdateDictRequest struct {
	ID string `json:"id" binding:"required,gt=0"` // 字典ID
	DictRequest
}

// DeleteDictRequest 删除字典请求模型
type DeleteDictRequest struct {
	ID string `json:"id" binding:"required,gt=0"` // 字典ID
}

// GetDictRequest 获取字典请求模型
type GetDictRequest struct {
	ID       string `form:"id" binding:"omitempty,gt=0"`          // 字典ID
	DictCode string `form:"dict_code" binding:"omitempty,max=50"` // 字典编码
}

// DictItemRequest 字典项基础请求模型
type DictItemRequest struct {
	DictID      string `json:"dictId" binding:"required,gt=0"`       // 字典 ID
	ItemKey     string `json:"itemKey" binding:"required,max=50"`    // 字典项键
	ItemValue   string `json:"itemValue" binding:"required,max=255"` // 字典项值
	Sort        int    `json:"sort" binding:"gte=0"`                 // 排序
	Description string `json:"description" binding:"max=255"`        // 描述
	Status      int    `json:"status" binding:"oneof=0 1"`           // 状态：1-启用 0-禁用
}

// CreateDictItemRequest 创建字典项请求模型
type CreateDictItemRequest struct {
	DictItemRequest
}

// UpdateDictItemRequest 更新字典项请求模型
type UpdateDictItemRequest struct {
	ID string `json:"id" binding:"required,gt=0"` // 字典项ID
	DictItemRequest
}

// DeleteDictItemRequest 删除字典项请求模型
type DeleteDictItemRequest struct {
	ID string `json:"id" binding:"required,gt=0"` // 字典项ID
}

// GetDictItemRequest 获取字典项请求模型
type GetDictItemRequest struct {
	ID     string `form:"id" binding:"omitempty,gt=0"`     // 字典项 ID
	DictID string `form:"dictId" binding:"omitempty,gt=0"` // 字典 ID
}
