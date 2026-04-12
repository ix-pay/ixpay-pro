package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// DepartmentResponse 部门响应模型
// 所有 ID 字段使用 string 格式，避免前端精度丢失
type DepartmentResponse struct {
	ID        string `json:"id,string"`       // 部门 ID
	ParentID  string `json:"parentId,string"` // 父部门 ID
	Name      string `json:"name"`            // 部门名称
	Code      string `json:"code"`            // 部门编码
	Sort      int    `json:"sort"`            // 排序
	Status    int    `json:"status"`          // 状态：1-启用 0-禁用
	Leader    string `json:"leader"`          // 负责人
	Phone     string `json:"phone"`           // 联系电话
	Email     string `json:"email"`           // 邮箱
	CreatedAt string `json:"createdAt"`       // 创建时间
	UpdatedAt string `json:"updatedAt"`       // 更新时间
	// 可选字段：子部门列表
	Children []*DepartmentResponse `json:"children,omitempty"`
}

// DepartmentListResponse 部门列表响应模型
type DepartmentListResponse struct {
	baseRes.PageResult
	List []*DepartmentResponse `json:"list"` // 部门列表
}
