package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// UserSimpleDTO 简单用户 DTO（列表使用）
// 所有 ID 字段使用 int64 格式返回，通过 json:",string" 标签自动序列化为字符串
type UserSimpleDTO struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// UserDetailDTO 详情用户 DTO（详情使用）
// 所有 ID 字段使用 int64 格式返回，通过 json:",string" 标签自动序列化为字符串
type UserDetailDTO struct {
	ID           int64     `json:"id,string"`
	Username     string    `json:"username"`
	Nickname     string    `json:"nickname"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Avatar       string    `json:"avatar"`
	Status       int       `json:"status"`
	DepartmentID int64     `json:"departmentId,string"`
	PositionID   int64     `json:"positionId,string"`
	Roles        []RoleDTO `json:"roles"`
	CreatedAt    string    `json:"createdAt"`
}

// UserSelectDTO 下拉选项 DTO（选择器使用）
// 所有 ID 字段使用 int64 格式返回，通过 json:",string" 标签自动序列化为字符串
type UserSelectDTO struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	PageResult baseRes.PageResult `json:"pageResult"`
	List       []UserSimpleDTO    `json:"list"`
}

// UserDetailResponse 用户详情响应
type UserDetailResponse struct {
	Data UserDetailDTO `json:"data"`
}

// UserSelectOptionsResponse 用户下拉选项响应
type UserSelectOptionsResponse struct {
	Data []UserSelectDTO `json:"data"`
}
