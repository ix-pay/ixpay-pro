package baseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// BtnPermController 按钮权限控制器
type BtnPermController struct {
	btnPermService    *service.BtnPermService
	permissionService *service.PermissionService
	roleService       *service.RoleService
	log               logger.Logger
}

// NewBtnPermController 创建按钮权限控制器
func NewBtnPermController(btnPermService *service.BtnPermService, permissionService *service.PermissionService, roleService *service.RoleService, log logger.Logger) *BtnPermController {
	return &BtnPermController{
		btnPermService:    btnPermService,
		permissionService: permissionService,
		roleService:       roleService,
		log:               log,
	}
}

// CreateBtnPerm 创建按钮权限
//
//	@Summary		创建按钮权限
//	@Description	创建新的按钮权限
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.CreateBtnPermRequest	true	"按钮权限信息"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/btn-perms [post]
func (c *BtnPermController) CreateBtnPerm(ctx *gin.Context) {
	var req request.CreateBtnPermRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("创建按钮权限参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	// 创建 BtnPerm 结构体实例
	btnPerm := &entity.BtnPerm{
		MenuID:      req.MenuID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}

	// 调用服务层创建按钮权限
	err := c.btnPermService.CreateBtnPerm(btnPerm, userID.(string))
	if err != nil {
		c.log.Error("创建按钮权限失败", "error", err.Error())
		baseRes.FailWithMessage("创建按钮权限失败", ctx)
		return
	}

	baseRes.OkWithMessage("创建成功", ctx)
}

// GetBtnPermByID 根据ID获取按钮权限
//
//	@Summary		根据ID获取按钮权限
//	@Description	根据ID获取按钮权限详情
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	query		uint	true	"按钮权限ID"
//	@Success		200	{object}	baseRes.Response{data=response.BtnPermResponse}
//	@Router			/api/admin/btn-perms/detail [get]
func (c *BtnPermController) GetBtnPermByID(ctx *gin.Context) {
	var req request.GetBtnPermByIDRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("获取按钮权限详情参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	btnPermDetail, err := c.btnPermService.GetBtnPermByID(req.ID)
	if err != nil {
		c.log.Error("获取按钮权限详情失败", "error", err.Error(), "btnPermID", req.ID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取按钮权限失败", ctx)
		return
	}

	baseRes.OkWithDetailed(btnPermDetail, "获取按钮权限成功", ctx)
}

// UpdateBtnPerm 更新按钮权限
//
//	@Summary		更新按钮权限
//	@Description	更新按钮权限信息
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.UpdateBtnPermRequest	true	"按钮权限信息"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/btn-perms [put]
func (c *BtnPermController) UpdateBtnPerm(ctx *gin.Context) {
	var req request.UpdateBtnPermRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("更新按钮权限参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	// 创建 BtnPerm 结构体实例
	btnPerm := &entity.BtnPerm{
		MenuID:      req.MenuID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}
	// 调用 UpdateBtnPerm 方法
	err := c.btnPermService.UpdateBtnPerm(btnPerm, userID.(string))
	if err != nil {
		c.log.Error("更新按钮权限失败", "error", err.Error(), "userID", userID, "btnPermID", req.ID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "更新按钮权限失败", ctx)
		return
	}

	c.log.Info("更新按钮权限成功", "userID", userID, "btnPermID", req.ID)
	baseRes.OkWithMessage("更新按钮权限成功", ctx)
}

// DeleteBtnPerm 删除按钮权限
//
//	@Summary		删除按钮权限
//	@Description	删除按钮权限
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.DeleteBtnPermRequest	true	"按钮权限ID"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/btn-perms [delete]
func (c *BtnPermController) DeleteBtnPerm(ctx *gin.Context) {
	var req request.DeleteBtnPermRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("删除按钮权限参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	err := c.btnPermService.DeleteBtnPerm(req.ID)
	if err != nil {
		c.log.Error("删除按钮权限失败", "error", err.Error(), "userID", userID, "btnPermID", req.ID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "删除按钮权限失败", ctx)
		return
	}

	c.log.Info("删除按钮权限成功", "userID", userID, "btnPermID", req.ID)
	baseRes.OkWithMessage("删除按钮权限成功", ctx)
}

// GetBtnPermList 获取按钮权限列表
//
//	@Summary		获取按钮权限列表
//	@Description	获取按钮权限列表
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		true	"页码"
//	@Param			pageSize	query		int		true	"每页数量"
//	@Param			menu_id		query		int64	false	"菜单 ID"
//	@Param			code		query		string	false	"按钮编码"
//	@Param			name		query		string	false	"按钮名称"
//	@Param			status		query		int		false	"状态(-1:全部,0:禁用,1:启用)"
//	@Success		200			{object}	baseRes.Response{data=baseRes.PageResult{list=[]entity.BtnPerm,total=int64,page=int,pageSize=int}}
//	@Router			/api/admin/btn-perms [get]
func (c *BtnPermController) GetBtnPermList(ctx *gin.Context) {
	var req request.GetBtnPermListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("获取按钮权限列表参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if req.MenuID != "" {
		filters["menu_id"] = req.MenuID
	}
	if req.Code != "" {
		filters["code"] = req.Code
	}
	if req.Name != "" {
		filters["name"] = req.Name
	}
	if req.Status != -1 {
		filters["status"] = req.Status
	}

	btnPermList, total, err := c.btnPermService.GetBtnPermList(req.Page, req.PageSize, filters)
	if err != nil {
		c.log.Error("获取按钮权限列表失败", "error", err.Error(), "page", req.Page, "pageSize", req.PageSize)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取按钮权限列表失败", ctx)
		return
	}

	c.log.Info("获取按钮权限列表成功", "page", req.Page, "pageSize", req.PageSize, "total", total)
	baseRes.OkWithDetailed(map[string]interface{}{
		"list":  btnPermList,
		"total": total,
		"page":  req.Page,
		"size":  req.PageSize,
	}, "获取按钮权限列表成功", ctx)
}

// AssignToBtnPerm 为按钮权限分配API路由
//
//	@Summary		为按钮权限分配API路由
//	@Description	为按钮权限分配API路由
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.AssignToBtnPermRequest	true	"按钮ID和API路由ID列表"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/btn-perms/assign-api-routes [post]
func (c *BtnPermController) AssignToBtnPerm(ctx *gin.Context) {
	var req request.AssignToBtnPermRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("分配API路由到按钮参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	// 遍历 API 路由 ID 列表，为每个 ID 调用 AssignToBtnPerm 方法
	for _, routeID := range req.IDs {
		err := c.btnPermService.AssignAPIToBtnPerm(req.BtnPermID, routeID)
		if err != nil {
			c.log.Error("分配API路由到按钮失败", "error", err.Error(), "userID", userID, "btnPermID", req.BtnPermID, "routeID", routeID)
			baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "分配API路由到按钮失败", ctx)
			return
		}
	}

	c.log.Info("分配API路由到按钮成功", "userID", userID, "btnPermID", req.BtnPermID, "routeCount", len(req.IDs))
	baseRes.OkWithMessage("分配API路由到按钮成功", ctx)
}

// RevokeFromBtnPerm 从按钮权限撤销API路由
//
//	@Summary		从按钮权限撤销API路由
//	@Description	从按钮权限撤销API路由
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.RevokeFromBtnPermRequest	true	"按钮ID和API路由ID"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/btn-perms/revoke-api-route [post]
func (c *BtnPermController) RevokeFromBtnPerm(ctx *gin.Context) {
	var req request.RevokeFromBtnPermRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("撤销API路由从按钮参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	err := c.btnPermService.RevokeAPIFromBtnPerm(req.BtnPermID, req.ID)
	if err != nil {
		c.log.Error("撤销API路由从按钮失败", "error", err.Error(), "userID", userID, "btnPermID", req.BtnPermID, "routeID", req.ID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "撤销API路由从按钮失败", ctx)
		return
	}

	c.log.Info("撤销API路由从按钮成功", "userID", userID, "btnPermID", req.BtnPermID, "routeID", req.ID)
	baseRes.OkWithMessage("撤销API路由从按钮成功", ctx)
}

// AssignBtnPermToRole 为角色分配按钮权限
//
//	@Summary		为角色分配按钮权限
//	@Description	为角色分配按钮权限
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.AssignBtnPermToRoleRequest	true	"角色ID和按钮权限ID列表"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/btn-perms/assign-to-role [post]
func (c *BtnPermController) AssignBtnPermToRole(ctx *gin.Context) {
	var req request.AssignBtnPermToRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("分配按钮权限到角色参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	// 使用角色服务批量分配按钮权限
	err := c.roleService.BatchAssignBtnPermsToRole(req.RoleID, req.BtnPermIDs)
	if err != nil {
		c.log.Error("分配按钮权限到角色失败", "error", err.Error(), "userID", userID, "roleID", req.RoleID)
		baseRes.FailWithMessage("分配按钮权限到角色失败", ctx)
		return
	}

	c.log.Info("分配按钮权限到角色成功", "userID", userID, "roleID", req.RoleID, "btnPermCount", len(req.BtnPermIDs))
	baseRes.OkWithMessage("分配按钮权限到角色成功", ctx)
}

// RevokeBtnPermFromRole 从角色撤销按钮权限
//
//	@Summary		从角色撤销按钮权限
//	@Description	从角色撤销按钮权限
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.RevokeBtnPermFromRoleRequest	true	"角色ID和按钮权限ID"
//	@Success		200		{object}	baseRes.Response
//	@Router			/api/admin/btn-perms/revoke-from-role [post]
func (c *BtnPermController) RevokeBtnPermFromRole(ctx *gin.Context) {
	var req request.RevokeBtnPermFromRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("撤销按钮权限从角色参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	userID, _ := ctx.Get("userID")
	// 使用角色服务撤销按钮权限
	err := c.roleService.RevokeBtnPermFromRole(req.RoleID, req.BtnPermID)
	if err != nil {
		c.log.Error("从角色撤销按钮权限失败", "error", err.Error(), "userID", userID, "roleID", req.RoleID, "btnPermID", req.BtnPermID)
		baseRes.FailWithMessage("从角色撤销按钮权限失败", ctx)
		return
	}

	c.log.Info("从角色撤销按钮权限成功", "userID", userID, "roleID", req.RoleID, "btnPermID", req.BtnPermID)
	baseRes.OkWithMessage("从角色撤销按钮权限成功", ctx)
}

// GetAPIRoutesByBtnPerm 获取按钮权限关联的API路由
//
//	@Summary		获取按钮权限关联的API路由
//	@Description	获取按钮权限关联的API路由
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			btnPermId	query		int64	true	"按钮权限 ID"
//	@Success		200			{object}	baseRes.Response{data=[]entity.API}
//	@Router			/api/admin/btn-perms/api-routes [get]
func (c *BtnPermController) GetAPIRoutesByBtnPerm(ctx *gin.Context) {
	var req struct {
		BtnPermID string `form:"btnPermId" binding:"required,gte=1"`
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("获取按钮关联API路由参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	apiRoutes, err := c.btnPermService.GetAPIsForBtnPerm(req.BtnPermID)
	if err != nil {
		c.log.Error("获取按钮关联API路由失败", "error", err.Error(), "btnPermID", req.BtnPermID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取按钮关联API路由失败", ctx)
		return
	}

	c.log.Info("获取按钮关联API路由成功", "btnPermID", req.BtnPermID, "routeCount", len(apiRoutes))
	baseRes.OkWithDetailed(apiRoutes, "获取按钮关联API路由成功", ctx)
}

// GetBtnPermsByAPIRoute 获取API路由关联的按钮权限
//
//	@Summary		获取API路由关联的按钮权限
//	@Description	获取API路由关联的按钮权限
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			routeId	query		int64	true	"API 路由 ID"
//	@Success		200		{object}	baseRes.Response{data=[]entity.BtnPerm}
//	@Router			/api/admin/btn-perms/for-route [get]
func (c *BtnPermController) GetBtnPermsByAPIRoute(ctx *gin.Context) {
	var req struct {
		RouteID string `form:"routeId" binding:"required,gte=1"`
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("获取路由关联按钮权限参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	btnPerms, err := c.btnPermService.GetBtnPermsForAPI(req.RouteID)
	if err != nil {
		c.log.Error("获取路由关联按钮权限失败", "error", err.Error(), "routeID", req.RouteID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取路由关联按钮权限失败", ctx)
		return
	}

	c.log.Info("获取路由关联按钮权限成功", "routeID", req.RouteID, "btnPermCount", len(btnPerms))
	baseRes.OkWithDetailed(btnPerms, "获取路由关联按钮权限成功", ctx)
}

// GetBtnPermsByRole 获取角色拥有的按钮权限
//
//	@Summary		获取角色拥有的按钮权限
//	@Description	获取角色拥有的按钮权限
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			roleId	query		int64	true	"角色 ID"
//	@Success		200		{object}	baseRes.Response{data=[]response.BtnPermForRole}
//	@Router			/api/admin/btn-perms/by-role [get]
func (c *BtnPermController) GetBtnPermsByRole(ctx *gin.Context) {
	var req struct {
		RoleID string `form:"roleId" binding:"required"`
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("获取角色关联按钮权限参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	// 通过 PermissionService 获取角色的按钮权限列表
	btnPerms, err := c.permissionService.GetBtnPermsByRole(req.RoleID)
	if err != nil {
		c.log.Error("获取角色关联按钮权限失败", "error", err.Error(), "roleID", req.RoleID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取角色关联按钮权限失败", ctx)
		return
	}

	// 转换为响应模型
	respList := make([]response.BtnPermForRole, 0, len(btnPerms))
	for _, btnPerm := range btnPerms {
		respList = append(respList, response.BtnPermForRole{
			ID:          btnPerm.ID,
			MenuID:      btnPerm.MenuID,
			Code:        btnPerm.Code,
			Name:        btnPerm.Name,
			Description: btnPerm.Description,
			Status:      btnPerm.Status,
			IsAssigned:  true,
		})
	}

	c.log.Info("获取角色关联按钮权限成功", "roleID", req.RoleID, "btnPermCount", len(btnPerms))
	baseRes.OkWithDetailed(respList, "获取角色关联按钮权限成功", ctx)
}

// GetBtnPermsByMenu 获取菜单下的按钮权限
//
//	@Summary		获取菜单下的按钮权限
//	@Description	获取菜单下的所有按钮权限
//	@Tags			按钮权限管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			menuId	query		int64	true	"菜单 ID"
//	@Success		200		{object}	baseRes.Response{data=[]entity.BtnPerm}
//	@Router			/api/admin/btn-perms/by-menu [get]
func (c *BtnPermController) GetBtnPermsByMenu(ctx *gin.Context) {
	var req struct {
		MenuID string `form:"menuId" binding:"required,gte=1"`
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("获取菜单关联按钮权限参数验证失败", "error", err.Error())
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "参数验证失败", ctx)
		return
	}

	btnPerms, err := c.btnPermService.GetBtnPermsByMenu(req.MenuID)
	if err != nil {
		c.log.Error("获取菜单关联按钮权限失败", "error", err.Error(), "menuID", req.MenuID)
		baseRes.FailWithDetailed(map[string]interface{}{"error": err.Error()}, "获取菜单关联按钮权限失败", ctx)
		return
	}

	c.log.Info("获取菜单关联按钮权限成功", "menuID", req.MenuID, "btnPermCount", len(btnPerms))
	baseRes.OkWithDetailed(btnPerms, "获取菜单关联按钮权限成功", ctx)
}
