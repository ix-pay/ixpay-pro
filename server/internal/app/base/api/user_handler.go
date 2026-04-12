package baseapi

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// convertToUserResponse 将 entity.User 转换为 response.UserResponse
func convertToUserResponse(user *entity.User) response.UserResponse {
	return response.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Email:        user.Email,
		Phone:        user.Phone,
		Avatar:       user.Avatar,
		Status:       user.Status,
		DepartmentID: user.DepartmentID,
		PositionID:   user.PositionID,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
	}
}

// convertToUserSettingResponse 将 entity.UserSetting 转换为 response.UserSettingResponse
func convertToUserSettingResponse(setting *entity.UserSetting) response.UserSettingResponse {
	return response.UserSettingResponse{
		ID:               setting.ID,
		UserID:           setting.UserID,
		ThemeColor:       setting.ThemeColor,
		SidebarColor:     setting.SidebarColor,
		NavbarColor:      setting.NavbarColor,
		FontSize:         setting.FontSize,
		Language:         setting.Language,
		AutoLogin:        setting.AutoLogin,
		RememberPassword: setting.RememberPassword,
		CreatedAt:        setting.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        setting.UpdatedAt.Format(time.RFC3339),
	}
}

// UserController 用户控制器
// 处理用户相关的 HTTP 请求
// 提供用户注册、登录、信息查询等功能的 API 接口
// 字段:
//   - service: 用户服务，处理业务逻辑
//   - log: 日志记录器，记录操作日志
//     @Summary		用户相关 API
//     @Description	提供用户注册、登录、信息查询等功能
//     @Tags			用户管理
//     @Router			/api/admin//user [get]
type UserController struct {
	service *service.UserService // 用户服务接口
	log     logger.Logger        // 日志记录器
}

// NewUserController 创建用户控制器实例
// 参数:
// - service: 用户服务接口实现
// - log: 日志记录器
// 返回:
// - *UserController: 用户控制器实例
func NewUserController(service *service.UserService, log logger.Logger) *UserController {
	// 创建并返回用户控制器实例，注入依赖
	return &UserController{
		service: service,
		log:     log,
	}
}

// Register 用户注册
//
//	@Summary		用户注册
//	@Description	创建新用户账户
//	@Tags			基础服务
//	@Accept			json
//	@Produce		json
//	@Param			register	body		request.RegisterRequest							true	"注册请求参数"
//	@Success		201			{object}	baseRes.Response{data=entity.User,msg=string}	"注册成功"
//	@Failure		400			{object}	map[string]string								"请求参数错误"
//	@Router			/api/admin//auth/register [post]
func (c *UserController) Register(ctx *gin.Context) {
	var req request.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	user, err := c.service.Register(req.Username, req.Password, req.Email)
	if err != nil {
		c.log.Error("用户注册失败", "error", err)
		baseRes.FailWithMessage("用户注册失败", ctx)
		return
	}

	// 转换为 Response DTO
	userResponse := convertToUserResponse(user)
	baseRes.OkWithDetailed(userResponse, "注册成功", ctx)
}

// GetUserInfo 获取用户信息
//
//	@Summary		获取用户信息
//	@Description	获取当前登录用户的详细信息
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{data=response.UserInfoResponse,msg=string}	"用户信息"
//	@Failure		401	{object}	map[string]string											"未授权"
//	@Failure		500	{object}	map[string]string											"服务器内部错误"
//	@Router			/api/admin//user/info [get]
func (c *UserController) GetUserInfo(ctx *gin.Context) {
	// 从上下文中获取用户 ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	user, err := c.service.GetUserInfo(userID.(string))
	if err != nil {
		c.log.Error("获取用户信息失败", "error", err)
		baseRes.FailWithMessage("获取用户信息失败", ctx)
		return
	}

	// 从 Redis 获取用户的当前角色 ID
	currentRoleIDStr, err := c.service.GetCurrentRoleID(userID.(string))
	if err != nil {
		c.log.Warn("获取当前角色 ID 失败", "error", err)
		// 不返回错误，继续执行
	}

	// 通过用户角色 ID 列表获取完整的角色信息
	roleInfos := make([]*response.RoleInfo, 0, len(user.RoleIds))
	for _, roleID := range user.RoleIds {
		role, err := c.service.GetRoleByID(roleID)
		if err != nil {
			c.log.Warn("获取角色信息失败", "roleID", roleID, "error", err)
			continue
		}
		roleInfo := &response.RoleInfo{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
			Type:        role.Type,
			ParentId:    role.ParentID,
			Status:      role.Status,
			IsSystem:    role.IsSystem,
			Sort:        role.Sort,
		}
		roleInfos = append(roleInfos, roleInfo)
	}

	// 确定当前角色
	var finalRoleID string
	var finalRoleName string
	if currentRoleIDStr != "" {
		// 检查当前角色 ID 是否在用户的角色列表中
		for _, roleInfo := range roleInfos {
			if roleInfo.ID == currentRoleIDStr {
				finalRoleID = roleInfo.ID
				finalRoleName = roleInfo.Name
				break
			}
		}
	}

	// 如果 Redis 中没有存储或角色不在列表中，使用第一个角色作为当前角色
	if finalRoleID == "" && len(roleInfos) > 0 {
		finalRoleID = roleInfos[0].ID
		finalRoleName = roleInfos[0].Name
	}

	// 构建响应数据
	response := response.UserInfoResponse{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Phone:         user.Phone,
		Avatar:        user.Avatar,
		Status:        user.Status,
		Roles:         roleInfos,
		CurrentRoleId: finalRoleID,
		Role:          finalRoleName,
		Authority: response.AuthorityInfo{
			DefaultRouter: "index",
		},
	}

	baseRes.OkWithDetailed(response, "获取用户信息成功", ctx)
}

// GetUserList 获取用户列表
//
//	@Summary		获取用户列表
//	@Description	获取用户列表（支持分页和筛选）
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int																		true	"页码"
//	@Param			page_size	query		int																		true	"每页数量"
//	@Param			username	query		string																	false	"用户名"
//	@Param			email		query		string																	false	"邮箱"
//	@Param			role		query		string																	false	"角色"
//	@Param			status		query		int																		false	"状态"
//	@Success		200			{object}	baseRes.Response{data=baseRes.PageResult{list=[]entity.User},msg=string}	"用户列表"
//	@Failure		401			{object}	map[string]string														"未授权"
//	@Failure		500			{object}	map[string]string														"服务器内部错误"
//	@Router			/api/admin//user [get]
func (c *UserController) GetUserList(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetUserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 构建筛选条件
	filters := make(map[string]interface{})
	if req.Username != "" {
		filters["username"] = req.Username
	}
	if req.Email != "" {
		filters["email"] = req.Email
	}
	if req.Role != "" {
		filters["role"] = req.Role
	}
	if req.Status != nil {
		filters["status"] = *req.Status
	}

	users, total, err := c.service.GetUserList(req.Page, req.PageSize, filters)
	if err != nil {
		c.log.Error("获取用户列表失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为 Response DTO 数组
	userResponses := make([]*response.UserResponse, len(users))
	for i, user := range users {
		userDTO := convertToUserResponse(user)
		userResponses[i] = &userDTO
	}

	pageResult := baseRes.PageResult{
		List:     userResponses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	baseRes.OkWithDetailed(pageResult, "获取用户列表成功", ctx)
}

// AddUser 增加用户
//
//	@Summary		增加用户
//	@Description	增加新用户（管理员权限）
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			add_user	body		request.AddUserRequest							true	"增加用户请求参数"
//	@Success		201			{object}	baseRes.Response{data=entity.User,msg=string}	"添加成功"
//	@Failure		400			{object}	map[string]string								"请求参数错误"
//	@Failure		401			{object}	map[string]string								"未授权"
//	@Failure		500			{object}	map[string]string								"服务器内部错误"
//	@Router			/api/admin//user [post]
func (c *UserController) AddUser(ctx *gin.Context) {
	var req request.AddUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 获取当前登录用户ID作为创建者
	createdBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 设置默认状态
	status := req.Status
	if status == 0 {
		status = 1 // 默认激活
	}

	user, err := c.service.AddUser(req.Username, req.Password, req.Email, req.Nickname, req.Phone, req.Avatar, req.DepartmentID, req.PositionID, createdBy.(string), status)
	if err != nil {
		c.log.Error("添加用户失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 分配角色
	if len(req.Roles) > 0 {
		if err := c.service.UpdateUserRoles(user.ID, req.Roles); err != nil {
			c.log.Error("分配用户角色失败", "error", err)
			baseRes.FailWithMessage("分配用户角色失败", ctx)
			return
		}
	}

	// 转换为 Response DTO
	userResponse := convertToUserResponse(user)
	baseRes.OkWithDetailed(userResponse, "添加用户成功", ctx)
}

// UpdateUserInfo 更新用户信息
//
//	@Summary		更新用户信息
//	@Description	更新用户的基本信息
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			update_user	body		request.UpdateUserRequest						true	"更新用户请求参数"
//	@Success		200			{object}	baseRes.Response{data=entity.User,msg=string}	"更新成功"
//	@Failure		400			{object}	map[string]string								"请求参数错误"
//	@Failure		401			{object}	map[string]string								"未授权"
//	@Failure		500			{object}	map[string]string								"服务器内部错误"
//	@Router			/api/admin//user/info [put]
func (c *UserController) UpdateUserInfo(ctx *gin.Context) {
	var req request.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 检查用户是否已登录
	updatedBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 获取要更新的用户信息（使用 string 类型的 ID）
	user, err := c.service.GetUserInfo(req.ID)
	if err != nil {
		c.log.Error("获取用户信息失败", "error", err)
		baseRes.FailWithMessage("获取用户信息失败", ctx)
		return
	}

	// 更新用户信息
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status >= 0 {
		user.Status = req.Status
	}

	// 调用服务层更新用户信息
	if err := c.service.UpdateUserInfo(user, updatedBy.(string)); err != nil {
		c.log.Error("更新用户信息失败", "error", err)
		baseRes.FailWithMessage("更新用户信息失败", ctx)
		return
	}

	// 更新用户角色
	if len(req.Roles) > 0 {
		if err := c.service.UpdateUserRoles(user.ID, req.Roles); err != nil {
			c.log.Error("更新用户角色失败", "error", err)
			baseRes.FailWithMessage("更新用户角色失败", ctx)
			return
		}
	}

	// 转换为 Response DTO
	userResponse := convertToUserResponse(user)
	baseRes.OkWithDetailed(userResponse, "更新用户信息成功", ctx)
}

// DeleteUser 删除用户
//
//	@Summary		删除用户
//	@Description	删除用户（管理员权限）
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int								true	"用户ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin//user/{id} [delete]
func (c *UserController) DeleteUser(ctx *gin.Context) {
	// 解析用户 ID（直接使用 string 类型）
	userID := ctx.Param("id")
	if userID == "" {
		c.log.Error("用户 ID 不能为空")
		baseRes.FailWithMessage("用户 ID 不能为空", ctx)
		return
	}

	if err := c.service.DeleteUser(userID); err != nil {
		c.log.Error("删除用户失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("删除用户成功", ctx)
}

// ChangePassword 修改密码
//
//	@Summary		修改密码
//	@Description	修改当前用户的密码
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			change_password	body		request.ChangePasswordRequest	true	"修改密码请求参数"
//	@Success		200				{object}	baseRes.Response{msg=string}	"修改成功"
//	@Failure		400				{object}	map[string]string				"请求参数错误"
//	@Failure		401				{object}	map[string]string				"未授权"
//	@Failure		500				{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin//user/password [put]
func (c *UserController) ChangePassword(ctx *gin.Context) {
	var req request.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	if err := c.service.ChangePassword(userID.(string), req.OldPassword, req.NewPassword); err != nil {
		c.log.Error("修改密码失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("修改密码成功", ctx)
}

// ResetPassword 重置密码
//
//	@Summary		重置密码
//	@Description	重置用户密码为默认密码123456（管理员权限）
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			reset_password	body		request.ResetPasswordRequest	true	"重置密码请求参数"
//	@Success		200				{object}	baseRes.Response{msg=string}	"重置成功"
//	@Failure		400				{object}	map[string]string				"请求参数错误"
//	@Failure		401				{object}	map[string]string				"未授权"
//	@Failure		500				{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin//user/reset-password [put]
func (c *UserController) ResetPassword(ctx *gin.Context) {
	var req request.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 获取当前登录用户ID作为修改者
	updatedBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 直接使用 string 类型的 UserID
	userID := req.UserID
	if userID == "" {
		c.log.Error("用户 ID 不能为空")
		baseRes.FailWithMessage("用户 ID 不能为空", ctx)
		return
	}

	if err := c.service.ResetPassword(userID, "123456", updatedBy.(string)); err != nil {
		c.log.Error("重置密码失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("重置密码成功", ctx)
}

// GetSelfSetting 获取用户设置
//
//	@Summary		获取用户设置
//	@Description	获取当前登录用户的系统设置
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{data=map[string]entity.UserSetting,msg=string}	"用户设置信息"
//	@Failure		401	{object}	map[string]string												"未授权"
//	@Failure		500	{object}	map[string]string												"服务器内部错误"
//	@Router			/api/admin//user/getSelfSetting [get]
func (c *UserController) GetSelfSetting(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	setting, err := c.service.GetSelfSetting(userID.(string))
	if err != nil {
		c.log.Error("获取用户设置失败", "error", err)
		baseRes.FailWithMessage("获取用户设置失败", ctx)
		return
	}

	// 转换为 Response DTO
	settingResponse := convertToUserSettingResponse(setting)

	// 返回符合前端期望的格式 { data: { settings: {...} } }
	baseRes.OkWithDetailed(map[string]response.UserSettingResponse{"settings": settingResponse}, "获取用户设置成功", ctx)
}

// SetSelfSetting 保存用户设置
//
//	@Summary		保存用户设置
//	@Description	保存当前登录用户的系统设置
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_setting	body		entity.UserSetting												true	"用户设置参数"
//	@Success		200				{object}	baseRes.Response{data=map[string]entity.UserSetting,msg=string}	"保存成功"
//	@Failure		400				{object}	map[string]string												"请求参数错误"
//	@Failure		401				{object}	map[string]string												"未授权"
//	@Failure		500				{object}	map[string]string												"服务器内部错误"
//	@Router			/api/admin//user/setSelfSetting [put]
func (c *UserController) SetSelfSetting(ctx *gin.Context) {
	var req entity.UserSetting
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	setting, err := c.service.SetSelfSetting(userID.(string), &req)
	if err != nil {
		c.log.Error("保存用户设置失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为 Response DTO
	settingResponse := convertToUserSettingResponse(setting)

	// 返回符合前端期望的格式 { data: { settings: {...} } }
	baseRes.OkWithDetailed(map[string]response.UserSettingResponse{"settings": settingResponse}, "保存用户设置成功", ctx)
}

// SwitchRole 切换用户角色
//
//	@Summary		切换用户角色
//	@Description	切换当前用户的活动角色（不修改用户角色关联关系）
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			switch_role	body		request.SwitchRoleRequest	true	"切换角色请求参数"
//	@Success		200			{object}	baseRes.Response{msg=string}	"切换成功"
//	@Failure		400			{object}	map[string]string				"请求参数错误"
//	@Failure		401			{object}	map[string]string				"未授权"
//	@Failure		500			{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin//user/switch-role [post]
func (c *UserController) SwitchRole(ctx *gin.Context) {
	var req request.SwitchRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 从上下文中获取用户 ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 调用服务层切换角色（使用 string 类型）
	if err := c.service.SwitchRole(userID.(string), req.RoleID); err != nil {
		c.log.Error("切换角色失败", "userID", userID, "roleID", req.RoleID, "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("切换角色成功", ctx)
}
