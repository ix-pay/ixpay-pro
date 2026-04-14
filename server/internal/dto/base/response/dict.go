package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// DictItemResponse 字典项响应模型
// 所有 ID 字段使用 int64 格式，通过 json:",string" 标签自动序列化为字符串
type DictItemResponse struct {
	ID          int64  `json:"id,string"`     // 字典项 ID
	DictID      int64  `json:"dictId,string"` // 字典 ID
	ItemKey     string `json:"itemKey"`       // 字典项键
	ItemValue   string `json:"itemValue"`     // 字典项值
	Sort        int    `json:"sort"`          // 排序
	Description string `json:"description"`   // 描述
	Status      int    `json:"status"`        // 状态：1-启用 0-禁用
	CreatedAt   string `json:"createdAt"`     // 创建时间
	UpdatedAt   string `json:"updatedAt"`     // 更新时间
}

// DictResponse 字典响应模型
type DictResponse struct {
	ID          int64              `json:"id,string"`   // 字典 ID
	DictName    string             `json:"dictName"`    // 字典名称
	DictCode    string             `json:"dictCode"`    // 字典编码
	Description string             `json:"description"` // 描述
	Status      int                `json:"status"`      // 状态：1-启用 0-禁用
	CreatedAt   string             `json:"createdAt"`   // 创建时间
	UpdatedAt   string             `json:"updatedAt"`   // 更新时间
	DictItems   []DictItemResponse `json:"dictItems"`   // 字典项列表
}

// DictListResponse 字典列表响应模型
type DictListResponse struct {
	baseRes.PageResult
	List []DictResponse `json:"list"` // 字典列表
}

// DictItemListResponse 字典项列表响应模型
type DictItemListResponse struct {
	baseRes.PageResult
	List []DictItemResponse `json:"list"` // 字典项列表
}
