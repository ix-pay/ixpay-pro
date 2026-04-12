package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// PositionResponse 岗位响应模型
// 所有 ID 字段使用 string 格式，避免前端精度丢失
type PositionResponse struct {
	ID          string `json:"id,string"`   // 岗位 ID
	Name        string `json:"name"`        // 岗位名称
	Code        string `json:"code"`        // 岗位编码
	Sort        int    `json:"sort"`        // 排序
	Status      int    `json:"status"`      // 状态：1-启用 0-禁用
	Description string `json:"description"` // 描述
	CreatedAt   string `json:"createdAt"`   // 创建时间
	UpdatedAt   string `json:"updatedAt"`   // 更新时间
}

// PositionListResponse 岗位列表响应模型
type PositionListResponse struct {
	baseRes.PageResult
	List []PositionResponse `json:"list"` // 岗位列表
}
