package baseapi

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// RoleController 角色控制器
type RoleController struct {
	roleService           *service.RoleService
	rolePermissionService *service.RolePermissionService
	apiService            *service.APIService
	log                   logger.Logger
}

// NewRoleController 创建角色控制器
func NewRoleController(roleService *service.RoleService, rolePermissionService *service.RolePermissionService, apiService *service.APIService, log logger.Logger) *RoleController {
	return &RoleController{
		roleService:           roleService,
		rolePermissionService: rolePermissionService,
		apiService:            apiService,
		log:                   log,
	}
}

// convertToRoleResponse 将 entity.Role 转换为 response.RoleResponse
func convertToRoleResponse(role *entity.Role) response.RoleResponse {
	return response.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Type:        role.Type,
		ParentId:    role.ParentID,
		Status:      role.Status,
		IsSystem:    role.IsSystem,
		Sort:        role.Sort,
		CreatedAt:   role.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   role.UpdatedAt.Format(time.RFC3339),
	}
}

// CreateRole 创建角色
//
//	@Summary		创建角色
//	@Description	创建新的角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.CreateRoleRequest	true	"角色信息"
//	@Success		200		{object}	baseRes.Response{data=response.RoleResponse}
//	@Router			/api/admin/role [post]
func (c *RoleController) CreateRole(ctx *gin.Context) {
	var req request.CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("创建角色参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	// 使用名称的小写加下划线作为默认编码
	code := strings.ToLower(strings.ReplaceAll(req.Name, " ", "_"))
	// 提供默认值：roleType=1（普通角色）, parentID="0"（顶级角色）, sort=0, isSystem=false
	role, err := c.roleService.CreateRole(req.Name, code, req.Description, "0", 1, userID.(string), req.Status, 0, false)
	if err != nil {
		c.log.Error("创建角色失败", "error", err.Error(), "userID", userID, "roleName", req.Name)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "创建角色失败", ctx)
		return
	}

	c.log.Info("创建角色成功", "userID", userID, "roleName", req.Name)
	baseRes.OkWithDetailed(convertToRoleResponse(role), "创建角色成功", ctx)
}

// GetRoleByID 根据 ID 获取角色
//
//	@Summary		根据 ID 获取角色
//	@Description	根据 ID 获取角色详情
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	query		string	true	"角色 ID"
//	@Success	200	{object}	baseRes.Response{data=response.RoleDetailResponse}
//	@Router			/api/admin/role/detail [get]
func (c *RoleController) GetRoleByID(ctx *gin.Context) {
	var req request.GetRoleByIDRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("获取角色详情参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	// 直接使用 string 类型的 ID
	roleID := req.ID

	roleDetail, err := c.roleService.GetRoleByID(roleID)
	if err != nil {
		c.log.Error("获取角色详情失败", "error", err.Error(), "roleID", roleID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取角色失败", ctx)
		return
	}

	// 转换为响应结构
	roleResponse := convertToRoleResponse(roleDetail)
	roleDetailResponse := response.RoleDetailResponse{
		RoleResponse: roleResponse,
		Users:        []response.UserInfo{},
		Menus:        []response.MenuInfo{},
		Routes:       []response.RouteInfo{},
	}

	c.log.Info("获取角色详情成功", "roleID", roleID)
	baseRes.OkWithDetailed(roleDetailResponse, "获取角色成功", ctx)
}

// UpdateRole 更新角色
//
//	@Summary		更新角色
//	@Description	更新角色信息
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.UpdateRoleRequest	true	"角色信息"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/role [put]
func (c *RoleController) UpdateRole(ctx *gin.Context) {
	var req request.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("更新角色参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	// 使用名称的小写加下划线作为默认编码
	code := strings.ToLower(strings.ReplaceAll(req.Name, " ", "_"))
	// 更新角色时保持原有角色类型、父级、排序和系统标识不变
	err := c.roleService.UpdateRole(req.ID, req.Name, code, req.Description, "0", 1, userID.(string), req.Status, 0, false)
	if err != nil {
		c.log.Error("更新角色失败", "error", err.Error(), "userID", userID, "roleID", req.ID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "更新角色失败", ctx)
		return
	}

	c.log.Info("更新角色成功", "userID", userID, "roleID", req.ID)
	baseRes.OkWithMessage("更新角色成功", ctx)
}

// DeleteRole 删除角色
//
//	@Summary		删除角色
//	@Description	删除角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.DeleteRoleRequest	true	"角色ID"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/role [delete]
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	var req request.DeleteRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("删除角色参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	// 直接使用 string 类型的 ID
	roleID := req.ID
	err := c.roleService.DeleteRole(roleID)
	if err != nil {
		c.log.Error("删除角色失败", "error", err.Error(), "userID", userID, "roleID", req.ID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "删除角色失败", ctx)
		return
	}

	c.log.Info("删除角色成功", "userID", userID, "roleID", req.ID)
	baseRes.OkWithMessage("删除角色成功", ctx)
}

// GetRoleList 获取角色列表
//
//	@Summary		获取角色列表
//	@Description	获取角色列表
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		true	"页码"
//	@Param			pageSize	query		int		true	"每页数量"
//	@Param			name		query		string	false	"角色名称"
//	@Param			status		query		int		false	"状态(-1:全部,0:禁用,1:启用)"
//	@Success		200			{object}	baseRes.Response{data=response.RoleListResponse}
//	@Router			/api/admin/role [get]
func (c *RoleController) GetRoleList(ctx *gin.Context) {
	var req request.GetRoleListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("获取角色列表参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if req.Name != "" {
		filters["name"] = req.Name
	}
	if req.Status != nil && *req.Status != -1 {
		filters["status"] = *req.Status
	}

	roleList, total, err := c.roleService.GetRoleList(req.Page, req.PageSize, filters)
	if err != nil {
		c.log.Error("获取角色列表失败", "error", err.Error(), "page", req.Page, "pageSize", req.PageSize, "name", req.Name)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取角色列表失败", ctx)
		return
	}

	// 转换为响应结构
	responses := make([]response.RoleResponse, 0, len(roleList))
	for _, role := range roleList {
		responses = append(responses, convertToRoleResponse(role))
	}

	roleListResponse := response.RoleListResponse{
		PageResult: baseRes.PageResult{
			List:     responses,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		List: responses,
	}

	baseRes.OkWithDetailed(roleListResponse, "获取角色列表成功", ctx)
}

// AssignUserToRole 分配用户到角色
//
//	@Summary		分配用户到角色
//	@Description	分配用户到角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.AssignUserToRoleRequest	true	"角色ID和用户ID列表"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/role/assign-users [post]
func (c *RoleController) AssignUserToRole(ctx *gin.Context) {
	var req request.AssignUserToRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("分配用户到角色参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	// 遍历用户ID列表，逐个分配
	for _, userID := range req.UserIDs {
		err := c.roleService.AssignUserToRole(req.RoleID, userID)
		if err != nil {
			c.log.Error("分配用户到角色失败", "error", err.Error(), "roleID", req.RoleID, "userID", userID)
			baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "分配用户到角色失败", ctx)
			return
		}
	}

	c.log.Info("分配用户到角色成功", "roleID", req.RoleID, "userCount", len(req.UserIDs))
	baseRes.OkWithMessage("分配用户到角色成功", ctx)
}

// AssignMenuToRole 分配菜单到角色
//
//	@Summary		分配菜单到角色
//	@Description	分配菜单到角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.AssignMenuToRoleRequest	true	"角色ID和菜单ID列表"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/role/assign-menus [post]
func (c *RoleController) AssignMenuToRole(ctx *gin.Context) {
	var req request.AssignMenuToRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("分配菜单到角色参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	// 遍历菜单ID列表，逐个分配
	for _, menuID := range req.MenuIDs {
		err := c.roleService.AssignMenuToRole(req.RoleID, menuID)
		if err != nil {
			c.log.Error("分配菜单到角色失败", "error", err.Error(), "roleID", req.RoleID, "menuID", menuID)
			baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "分配菜单到角色失败", ctx)
			return
		}
	}

	c.log.Info("分配菜单到角色成功", "roleID", req.RoleID, "menuCount", len(req.MenuIDs))
	baseRes.OkWithMessage("分配菜单到角色成功", ctx)
}

// AssignToRole 分配API路由到角色
//
//	@Summary		分配API路由到角色
//	@Description	分配API路由到角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.AssignToRoleRequest	true	"角色ID和API路由ID列表"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/role/assign-api-routes [post]
func (c *RoleController) AssignAPIToRole(ctx *gin.Context) {
	var req request.AssignToRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("分配API路由到角色参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	// 使用批量分配方法，更高效
	err := c.roleService.BatchAssignAPIsToRole(req.RoleID, req.IDs)
	if err != nil {
		c.log.Error("分配API路由到角色失败", "error", err.Error(), "roleID", req.RoleID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "分配API路由到角色失败", ctx)
		return
	}

	c.log.Info("分配API路由到角色成功", "roleID", req.RoleID, "routeCount", len(req.IDs))
	baseRes.OkWithMessage("分配API路由到角色成功", ctx)
}

// GetAllRoles 获取所有角色
//
//	@Summary		获取所有角色
//	@Description	获取所有角色，用于前端选择角色选项
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{data=[]entity.Role}
//	@Router			/api/admin/role/all [get]
func (c *RoleController) GetAllRoles(ctx *gin.Context) {
	roles, err := c.roleService.GetAllRoles()
	if err != nil {
		c.log.Error("获取所有角色失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取所有角色失败", ctx)
		return
	}

	c.log.Info("获取所有角色成功", "count", len(roles))
	baseRes.OkWithDetailed(roles, "获取所有角色成功", ctx)
}

// SaveRolePermissions 保存角色权限
//
//	@Summary		保存角色权限
//	@Description	保存角色的菜单、按钮和 API 权限
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"角色 ID"
//	@Param			data	body		request.SaveRolePermissionsRequest	true	"权限信息"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/role/:id/permissions [post]
func (c *RoleController) SaveRolePermissions(ctx *gin.Context) {
	// 直接使用 string 类型的 ID
	roleID := ctx.Param("id")

	var req request.SaveRolePermissionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("保存角色权限参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	// 获取操作人 ID
	operatorID, _ := ctx.Get("userID")

	// 保存权限
	err := c.rolePermissionService.SaveRolePermissions(
		roleID,
		req.MenuIds,
		req.BtnPermIds,
		req.ApiRouteIds,
		operatorID.(string),
	)
	if err != nil {
		c.log.Error("保存角色权限失败", "error", err.Error(), "roleID", roleID)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("保存角色权限成功", "roleID", roleID, "menuCount", len(req.MenuIds), "btnPermCount", len(req.BtnPermIds), "apiCount", len(req.ApiRouteIds))
	baseRes.OkWithMessage("保存角色权限成功", ctx)
}

// GetRoleDetail 根据路径参数 ID 获取角色详情
//
//	@Summary		根据路径参数 ID 获取角色详情
//	@Description	根据路径参数 ID 获取角色详情，支持 RESTful 风格
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"角色 ID"
//	@Success	200	{object}	entity.Role
//	@Router			/api/admin/role/:id/detail [get]
func (c *RoleController) GetRoleDetail(ctx *gin.Context) {
	// 直接使用 string 类型的 ID
	roleID := ctx.Param("id")

	roleDetail, err := c.roleService.GetRoleByID(roleID)
	if err != nil {
		c.log.Error("获取角色详情失败", "error", err.Error(), "roleID", roleID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取角色失败", ctx)
		return
	}

	c.log.Info("获取角色详情成功", "roleID", roleID)
	baseRes.OkWithDetailed(roleDetail, "获取角色成功", ctx)
}

// GetAvailableAPIs 获取角色可用的 API 列表
//
//	@Summary		获取角色可用的 API 列表
//	@Description	获取尚未分配给该角色的 API 路由列表
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"角色 ID"
//	@Success		200	{object}	baseRes.Response{data=[]response.APIResponse}
//	@Router			/api/admin/role/:id/available-apis [get]
func (c *RoleController) GetAvailableAPIs(ctx *gin.Context) {
	// 直接使用 string 类型的 ID
	roleID := ctx.Param("id")

	// 获取该角色已分配的 API 列表
	assignedAPIs, err := c.roleService.GetAPIsForRole(roleID)
	if err != nil {
		c.log.Error("获取角色已分配 API 失败", "error", err.Error(), "roleID", roleID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取角色权限失败", ctx)
		return
	}

	// 创建已分配 API ID 的 map，用于快速查找
	assignedAPIMap := make(map[string]bool)
	for _, api := range assignedAPIs {
		assignedAPIMap[api.ID] = true
	}

	// 获取所有 API 列表
	allAPIs, err := c.apiService.GetRoutes()
	if err != nil {
		c.log.Error("获取所有 API 失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取 API 列表失败", ctx)
		return
	}

	// 过滤出未分配的 API（排除 auth_type = 0 的基础 API）
	availableAPIs := make([]*entity.API, 0)
	for _, api := range allAPIs {
		if !assignedAPIMap[api.ID] && api.AuthType != 0 {
			availableAPIs = append(availableAPIs, api)
		}
	}

	c.log.Info("获取角色可用 API 列表成功", "roleID", roleID, "totalCount", len(allAPIs), "availableCount", len(availableAPIs))
	baseRes.OkWithDetailed(availableAPIs, "获取可用 API 列表成功", ctx)
}
