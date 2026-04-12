package response

import "github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"

// UserSimpleDTO 简单用户 DTO（列表使用）
type UserSimpleDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// UserDetailDTO 详情用户 DTO（详情使用）
type UserDetailDTO struct {
	ID           string       `json:"id"`
	Username     string       `json:"username"`
	Nickname     string       `json:"nickname"`
	Email        string       `json:"email"`
	Phone        string       `json:"phone"`
	Avatar       string       `json:"avatar"`
	Status       int          `json:"status"`
	DepartmentID string       `json:"departmentId"`
	PositionID   string       `json:"positionId"`
	Roles        []RoleDTO    `json:"roles"`
	CreatedAt    string       `json:"createdAt"`
}

// UserSelectDTO 下拉选项 DTO（选择器使用）
type UserSelectDTO struct {
	ID       string `json:"id"`
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
