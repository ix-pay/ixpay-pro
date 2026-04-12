// request 包定义岗位管理相关的请求模型
// 用于接收和验证 HTTP 请求参数
package request

// CreatePositionRequest 创建岗位请求
type CreatePositionRequest struct {
	Name        string `json:"name" binding:"required"` // 岗位名称
	Sort        int    `json:"sort"`                    // 排序，默认为 0
	Status      int    `json:"status"`                  // 状态：1-正常，0-禁用，默认为 1
	Description string `json:"description"`             // 岗位描述
}

// UpdatePositionRequest 更新岗位请求
type UpdatePositionRequest struct {
	ID          string `json:"id" binding:"required"`   // 岗位 ID
	Name        string `json:"name" binding:"required"` // 岗位名称
	Sort        int    `json:"sort"`                    // 排序
	Status      int    `json:"status"`                  // 状态：1-正常，0-禁用
	Description string `json:"description"`             // 岗位描述
}

// GetPositionListRequest 获取岗位列表请求
type GetPositionListRequest struct {
	Page     int  `form:"page" binding:"required"`     // 页码
	PageSize int  `form:"pageSize" binding:"required"` // 每页数量
	Status   *int `form:"status"`                      // 状态（可选筛选条件）
}

// GetPositionByIDRequest 获取岗位详情请求
type GetPositionByIDRequest struct {
	ID string `form:"id" binding:"required"` // 岗位 ID
}

// DeletePositionRequest 删除岗位请求
type DeletePositionRequest struct {
	ID string `json:"id" binding:"required"` // 岗位 ID
}
