// request 包定义部门管理相关的请求模型
// 用于接收和验证 HTTP 请求参数
package request

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"` // 部门名称
	ParentID    string `json:"parentId"`                // 父部门 ID，默认为 0（顶级部门）
	LeaderID    string `json:"leaderId"`                // 部门负责人 ID
	Sort        int    `json:"sort"`                    // 排序，默认为 0
	Status      int    `json:"status"`                  // 状态：1-正常，0-禁用，默认为 1
	Description string `json:"description"`             // 部门描述
}

// UpdateDepartmentRequest 更新部门请求
type UpdateDepartmentRequest struct {
	ID          string `json:"id" binding:"required"`   // 部门 ID
	Name        string `json:"name" binding:"required"` // 部门名称
	ParentID    string `json:"parentId"`                // 父部门 ID
	LeaderID    string `json:"leaderId"`                // 部门负责人 ID
	Sort        int    `json:"sort"`                    // 排序
	Status      int    `json:"status"`                  // 状态：1-正常，0-禁用
	Description string `json:"description"`             // 部门描述
}

// UpdateDepartmentLeaderRequest 更新部门负责人请求
type UpdateDepartmentLeaderRequest struct {
	LeaderID string `json:"leaderId" binding:"required"` // 部门负责人 ID
}

// GetDepartmentListRequest 获取部门列表请求
type GetDepartmentListRequest struct {
	Page     int  `form:"page" binding:"required"`     // 页码
	PageSize int  `form:"pageSize" binding:"required"` // 每页数量
	ParentID int  `form:"parentId"`                    // 父部门 ID（可选筛选条件）
	Status   *int `form:"status"`                      // 状态（可选筛选条件）
}

// GetDepartmentByIDRequest 获取部门详情请求
type GetDepartmentByIDRequest struct {
	ID string `form:"id" binding:"required"` // 部门 ID
}

// DeleteDepartmentRequest 删除部门请求
type DeleteDepartmentRequest struct {
	ID string `json:"id" binding:"required"` // 部门 ID
}
