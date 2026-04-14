package request

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=30"`
	Email    string `json:"email" binding:"required,email"`
}

// UpdateUserRequest 更新用户信息请求参数
type UpdateUserRequest struct {
	ID       int64    `json:"id" binding:"required"`
	Nickname string   `json:"nickname" binding:"max=50"`
	Email    string   `json:"email" binding:"email"`
	Phone    string   `json:"phone" binding:"max=20"`
	Avatar   string   `json:"avatar" binding:"max=255"`
	Status   int      `json:"status" binding:"omitempty,min=0,max=1"` // 1: active, 0: inactive
	Roles    []string `json:"roles" binding:"omitempty"`              // 角色列表，支持多选
}

// ChangePasswordRequest 修改密码请求参数
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6,max=30"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=30"`
}

// ResetPasswordRequest 重置密码请求参数
type ResetPasswordRequest struct {
	UserID int64 `json:"userId" binding:"required"`
}

// AddUserRequest 增加用户请求参数
type AddUserRequest struct {
	Username     string   `json:"username" binding:"required,min=3,max=50"`
	Password     string   `json:"password" binding:"required,min=6,max=30"`
	Email        string   `json:"email" binding:"omitempty,email"`
	Nickname     string   `json:"nickname" binding:"max=50"`
	Phone        string   `json:"phone" binding:"max=20"`
	Avatar       string   `json:"avatar" binding:"max=255"`
	DepartmentID int64    `json:"departmentId" binding:"omitempty,min=1"`
	PositionID   int64    `json:"positionId" binding:"omitempty,min=1"`
	Status       int      `json:"status" binding:"omitempty,min=0,max=1"`
	Roles        []string `json:"roles" binding:"omitempty"`
}

// GetUserListRequest 获取用户列表请求参数
type GetUserListRequest struct {
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
	Username string `json:"username" form:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" form:"email" binding:"omitempty,email"`
	Role     string `json:"role" form:"role" binding:"omitempty,oneof=user admin"`
	Status   *int   `json:"status" form:"status" binding:"omitempty,min=0,max=1"`
}

// SwitchRoleRequest 切换角色请求参数
type SwitchRoleRequest struct {
	RoleID string `json:"roleId" binding:"required"`
}
