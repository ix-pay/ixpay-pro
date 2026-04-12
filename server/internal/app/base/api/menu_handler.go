package baseapi

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// MenuController 菜单控制器
//
//	@Summary		菜单相关API
//	@Description	提供菜单管理功能
//	@Tags			菜单管理
//	@Router			/api/admin/menu [get]
type MenuController struct {
	service *service.MenuService
	log     logger.Logger
}

// NewMenuController 创建菜单控制器
func NewMenuController(service *service.MenuService, log logger.Logger) *MenuController {
	return &MenuController{
		service: service,
		log:     log,
	}
}

// convertToMenuResponse 将 entity.Menu 转换为 response.MenuResponse
// 递归处理子菜单，确保所有 ID 字段都转为字符串
func convertToMenuResponse(menu *entity.Menu) *response.MenuResponse {
	if menu == nil {
		return nil
	}

	// 转换子菜单（递归）
	var children []response.MenuResponse
	if len(menu.Children) > 0 {
		children = make([]response.MenuResponse, len(menu.Children))
		for i, child := range menu.Children {
			childResp := convertToMenuResponse(child)
			if childResp != nil {
				children[i] = *childResp
			}
		}
	} else {
		children = []response.MenuResponse{}
	}

	return &response.MenuResponse{
		ID:           menu.ID,
		ParentID:     menu.ParentID,
		Path:         menu.Path,
		Name:         menu.Name,
		Component:    menu.Component,
		Title:        menu.Title,
		Icon:         menu.Icon,
		Hidden:       menu.Hidden,
		Sort:         menu.Sort,
		Status:       menu.Status,
		IsExt:        menu.IsExt,
		Redirect:     menu.Redirect,
		Permission:   menu.Permission,
		KeepAlive:    menu.KeepAlive,
		DefaultMenu:  menu.DefaultMenu,
		Breadcrumb:   menu.Breadcrumb,
		ActiveMenu:   menu.ActiveMenu,
		Affix:        menu.Affix,
		Type:         int(menu.Type),
		FrameLoading: menu.FrameLoading,
		Meta: &response.MenuMetaResp{
			Title:        menu.Meta.Title,
			Icon:         menu.Meta.Icon,
			KeepAlive:    menu.Meta.KeepAlive,
			DefaultMenu:  menu.Meta.DefaultMenu,
			Breadcrumb:   menu.Meta.Breadcrumb,
			ActiveMenu:   menu.Meta.ActiveMenu,
			Affix:        menu.Meta.Affix,
			FrameLoading: menu.Meta.FrameLoading,
		},
		Children: children,
	}
}

// GetMenuList 获取菜单列表
//
//	@Summary		获取菜单列表
//	@Description	获取系统菜单列表
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success	200	{object}	baseRes.Response{data=[]entity.Menu,msg=string}	"菜单列表"
//	@Failure		401	{object}	map[string]string								"未授权"
//	@Failure		500	{object}	map[string]string								"服务器内部错误"
//	@Router			/api/admin/menu [get]
func (c *MenuController) GetMenuList(ctx *gin.Context) {
	// 从上下文中获取用户ID
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 从上下文中获取用户角色
	role, roleExists := ctx.Get("role")
	if !roleExists {
		c.log.Error("获取用户角色失败")
		baseRes.FailWithMessage("获取用户角色失败", ctx)
		return
	}

	menus, err := c.service.GetUserMenus(role.(string))
	if err != nil {
		c.log.Error("获取菜单列表失败", "error", err)
		baseRes.FailWithMessage("获取菜单列表失败", ctx)
		return
	}

	baseRes.OkWithDetailed(menus, "获取菜单列表成功", ctx)
}

// AddMenu 创建菜单
func (c *MenuController) AddMenu(ctx *gin.Context) {
	var req request.AddMenuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 直接使用 string 类型的 ParentID
	// 将字符串类型的 Status 转换为 int
	status, err := strconv.Atoi(req.Status)
	if err != nil {
		c.log.Error("Status 格式错误", "error", err, "status", req.Status)
		baseRes.FailWithMessage("Status 格式错误", ctx)
		return
	}

	// 获取当前登录用户 ID 作为创建者
	createdBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 构建菜单实体
	menu := entity.Menu{
		ParentID:     req.ParentID,
		Title:        req.Title,
		Name:         req.Name,
		Path:         req.Path,
		Component:    req.Component,
		Icon:         req.Icon,
		Sort:         req.Sort,
		Status:       status,
		Type:         entity.MenuType(req.Type),
		Hidden:       req.Hidden,
		IsExt:        req.IsExt,
		Redirect:     req.Redirect,
		Permission:   req.Permission,
		KeepAlive:    req.KeepAlive,
		DefaultMenu:  req.DefaultMenu,
		Breadcrumb:   req.Breadcrumb,
		ActiveMenu:   req.ActiveMenu,
		Affix:        req.Affix,
		FrameLoading: req.FrameLoading,
	}

	// 调用服务层创建菜单
	if err := c.service.CreateMenu(&menu, createdBy.(string)); err != nil {
		c.log.Error("创建菜单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 获取完整的菜单信息用于响应
	menus, _, err := c.service.GetMenuList(1, 100, map[string]interface{}{"id": menu.ID})
	if err == nil && len(menus) > 0 {
		// 转换为响应结构（包含字符串格式的 ID）
		menuResp := convertToMenuResponse(menus[0])
		baseRes.OkWithDetailed(menuResp, "创建菜单成功", ctx)
		return
	}

	// 如果获取完整信息失败，至少返回创建的基本信息（转换为响应结构）
	menuResp := &response.MenuResponse{
		ID:       menu.ID,
		ParentID: menu.ParentID,
		Title:    menu.Title,
		Name:     menu.Name,
		Path:     menu.Path,
	}
	baseRes.OkWithDetailed(menuResp, "创建菜单成功", ctx)
}

// UpdateMenu 更新菜单
func (c *MenuController) UpdateMenu(ctx *gin.Context) {
	var req request.UpdateMenuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("请求参数错误", "error", err.Error(), "raw_req", ctx.Request.Body)
		baseRes.FailWithMessage("请求参数错误："+err.Error(), ctx)
		return
	}

	// 将字符串类型的 ID 转换为 int64
	// 直接使用 string 类型的 ID 和 ParentID
	// 将字符串类型的 Status 转换为 int
	status, err := strconv.Atoi(req.Status)
	if err != nil {
		c.log.Error("Status 格式错误", "error", err, "status", req.Status)
		baseRes.FailWithMessage("Status 格式错误", ctx)
		return
	}

	// 获取当前登录用户 ID 作为修改者
	updatedBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 构建菜单实体
	menu := entity.Menu{
		ParentID:     req.ParentID,
		Title:        req.Title,
		Name:         req.Name,
		Path:         req.Path,
		Component:    req.Component,
		Icon:         req.Icon,
		Sort:         req.Sort,
		Status:       status,
		Type:         entity.MenuType(req.Type),
		Hidden:       req.Hidden,
		IsExt:        req.IsExt,
		Redirect:     req.Redirect,
		Permission:   req.Permission,
		KeepAlive:    req.KeepAlive,
		DefaultMenu:  req.DefaultMenu,
		Breadcrumb:   req.Breadcrumb,
		ActiveMenu:   req.ActiveMenu,
		Affix:        req.Affix,
		FrameLoading: req.FrameLoading,
	}
	menu.ID = req.ID

	// 调用服务层更新菜单（包含 API 关联）
	if err := c.service.UpdateMenuWithAPIs(&menu, req.ApiIds, updatedBy.(string)); err != nil {
		c.log.Error("更新菜单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 获取完整的菜单信息用于响应
	menus, _, err := c.service.GetMenuList(1, 100, map[string]interface{}{"id": menu.ID})
	if err == nil && len(menus) > 0 {
		// 转换为响应结构（包含字符串格式的 ID）
		menuResp := convertToMenuResponse(menus[0])
		baseRes.OkWithDetailed(menuResp, "更新菜单成功", ctx)
		return
	}

	// 如果获取完整信息失败，至少返回更新的基本信息（转换为响应结构）
	menuResp := &response.MenuResponse{
		ID:       menu.ID,
		ParentID: menu.ParentID,
		Title:    menu.Title,
		Name:     menu.Name,
		Path:     menu.Path,
	}
	baseRes.OkWithDetailed(menuResp, "更新菜单成功", ctx)
}

// DeleteMenu 删除菜单
//
//	@Summary		删除菜单
//	@Description	删除菜单（管理员权限）
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int								true	"菜单ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/menu/:id [delete]
func (c *MenuController) DeleteMenu(ctx *gin.Context) {
	// 解析菜单ID
	menuID := ctx.Param("id")
	if menuID == "" {
		c.log.Error("菜单 ID 不能为空")
		baseRes.FailWithMessage("菜单 ID 不能为空", ctx)
		return
	}

	// 调用服务层删除菜单
	if err := c.service.DeleteMenu(menuID); err != nil {
		c.log.Error("删除菜单失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("删除菜单成功", ctx)
}

// GetMenuPage 获取菜单分页列表
//
//	@Summary		获取菜单分页列表
//	@Description	获取菜单分页列表（支持过滤）
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int																		true	"页码"
//	@Param			page_size	query		int																		true	"每页数量"
//	@Param			title		query		string																	false	"菜单标题"
//	@Param			status		query		int																		false	"状态"
//	@Success		200			{object}	baseRes.Response{data=baseRes.PageResult{list=[]response.MenuResponse},msg=string}	"菜单分页列表"
//	@Failure		400			{object}	map[string]string														"请求参数错误"
//	@Failure		401			{object}	map[string]string														"未授权"
//	@Failure		500			{object}	map[string]string														"服务器内部错误"
//	@Router			/api/admin/menu/page [get]
func (c *MenuController) GetMenuPage(ctx *gin.Context) {
	var req request.GetMenuPageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 构建筛选条件
	filters := make(map[string]interface{})
	if req.Title != "" {
		filters["title"] = req.Title
	}
	// 只有当明确传递了 status 参数时才添加过滤（通过检查原始请求）
	if ctx.Query("status") != "" {
		filters["status"] = req.Status
	}

	// 调用服务层获取分页菜单列表
	menus, total, err := c.service.GetMenuList(req.Page, req.PageSize, filters)
	if err != nil {
		c.log.Error("获取菜单分页列表失败", "error", err)
		baseRes.FailWithMessage("获取菜单分页列表失败", ctx)
		return
	}

	pageResult := baseRes.PageResult{
		List:     menus,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	baseRes.OkWithDetailed(pageResult, "获取菜单分页列表成功", ctx)
}

// GetMenuDeleteImpact 获取菜单删除影响评估
//
//	@Summary		获取菜单删除影响评估
//	@Description	获取删除菜单前的影响范围评估（管理员权限）
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int																				true	"菜单 ID"
//	@Success		200	{object}	baseRes.Response{data=entity.DeleteImpact,msg=string}	"影响评估结果"
//	@Failure		400	{object}	map[string]string																"请求参数错误"
//	@Failure		401	{object}	map[string]string																"未授权"
//	@Failure		404	{object}	map[string]string																"菜单不存在"
//	@Failure		500	{object}	map[string]string																"服务器内部错误"
//	@Router			/api/admin/menu/:id/delete-impact [get]
func (c *MenuController) GetMenuDeleteImpact(ctx *gin.Context) {
	// 获取菜单 ID
	menuID := ctx.Param("id")
	if menuID == "" {
		c.log.Error("菜单 ID 不能为空")
		baseRes.FailWithMessage("菜单 ID 不能为空", ctx)
		return
	}

	// 调用服务层计算删除影响
	impact, err := c.service.CalculateDeleteImpact(menuID)
	if err != nil {
		c.log.Error("计算删除影响失败", "error", err, "menuID", menuID)
		baseRes.FailWithMessage("计算删除影响失败", ctx)
		return
	}

	baseRes.OkWithDetailed(impact, "获取删除影响评估成功", ctx)
}

// GetMenuTree 获取菜单树结构
//
//	@Summary		获取菜单树结构
//	@Description	获取完整的菜单树结构（管理员权限）
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{data=[]response.MenuResponse,msg=string}						"菜单树结构"
//	@Failure		401	{object}	map[string]string																"未授权"
//	@Failure		500	{object}	map[string]string																"服务器内部错误"
//	@Router			/api/admin/menu/tree [get]
func (c *MenuController) GetMenuTree(ctx *gin.Context) {
	// 从上下文中获取用户 ID
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 调用服务层获取菜单树
	menus, err := c.service.GetAllMenuTree()
	if err != nil {
		c.log.Error("获取菜单树失败", "error", err)
		baseRes.FailWithMessage("获取菜单树失败", ctx)
		return
	}

	// 转换为响应结构（包含字符串格式的 ID）
	menuResponses := make([]response.MenuResponse, len(menus))
	for i, menu := range menus {
		resp := convertToMenuResponse(menu)
		if resp != nil {
			menuResponses[i] = *resp
		}
	}

	baseRes.OkWithDetailed(menuResponses, "获取菜单树成功", ctx)
}
