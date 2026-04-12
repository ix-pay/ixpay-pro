package baseapi

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// APIController API 路由控制器
type APIController struct {
	apisService *service.APIService
	log         logger.Logger
}

// NewAPIController 创建 API 路由控制器实例
func NewAPIController(apisService *service.APIService, log logger.Logger) *APIController {
	return &APIController{
		apisService: apisService,
		log:         log,
	}
}

// 检查用户是否为管理员
func (c *APIController) checkAdminPermission(ctx *gin.Context) (string, bool) {
	role, exists := ctx.Get("role")
	if !exists || role != "admin" {
		c.log.Error("Permission denied", "user_role", role)
		baseRes.FailWithMessage("无权限访问", ctx)
		return "", false
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未找到用户信息")
		baseRes.FailWithMessage("未找到用户信息", ctx)
		return "", false
	}

	var userIDStr string
	switch v := userID.(type) {
	case string:
		userIDStr = v
	case int64:
		userIDStr = fmt.Sprintf("%d", v)
	case int:
		userIDStr = fmt.Sprintf("%d", v)
	default:
		c.log.Error("用户 ID 类型错误", "actual_type", fmt.Sprintf("%T", v))
		baseRes.FailWithMessage("用户 ID 类型错误", ctx)
		return "", false
	}

	return userIDStr, true
}

// convertToAPIResponse 将 entity.API 转换为 response.APIResponse
func convertToAPIResponse(route *entity.API) *response.APIResponse {
	if route == nil {
		return nil
	}

	return &response.APIResponse{
		ID:           route.ID,
		Path:         route.Path,
		Method:       route.Method,
		Group:        route.Group,
		AuthRequired: route.AuthRequired,
		AuthType:     route.AuthType,
		Description:  route.Description,
		Status:       route.Status,
		RoleIds:      route.RoleIds,
		MenuIds:      route.MenuIds,
		BtnPermIds:   route.BtnPermIds,
		CreatedBy:    route.CreatedBy,
		CreatedAt:    route.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:    route.UpdatedBy,
		UpdatedAt:    route.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// GetAPIs 获取所有 API 路由列表（支持分页和搜索）
//
//	@Summary		获取所有 API 路由列表
//	@Description	获取系统中注册的所有 API 路由信息（管理员权限），支持分页和搜索
//	@Tags			API 路由管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int										false	"页码，默认为 1"
//	@Param			pageSize	query		int										false	"每页数量，默认为 100"
//	@Param			keyword		query		string									false	"搜索关键词（支持路径、描述、分组）"
//	@Param			group		query		string									false	"路由分组"
//	@Success		200			{object}	baseRes.Response{data=baseRes.PageResult}	"路由列表"
//	@Failure		401			{object}	baseRes.Response{msg=string}				"未授权"
//	@Failure		403			{object}	baseRes.Response{msg=string}				"无权限"
//	@Failure		500			{object}	baseRes.Response{msg=string}				"服务器内部错误"
//	@Router			/api/admin//apis [get]
func (c *APIController) GetAPIs(ctx *gin.Context) {
	_, ok := c.checkAdminPermission(ctx)
	if !ok {
		return
	}

	var req request.GetAPIPageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	filters := make(map[string]interface{})
	if req.Keyword != "" {
		filters["keyword"] = req.Keyword
	}
	if req.Group != "" {
		filters["group"] = req.Group
	}

	routes, total, err := c.apisService.GetAPIRouteList(req.Page, req.PageSize, filters)
	if err != nil {
		c.log.Error("Failed to get API routes", "error", err)
		baseRes.FailWithMessage("获取 API 路由列表失败", ctx)
		return
	}

	c.log.Info("Successfully retrieved API routes", "count", len(routes), "total", total)

	apiResponses := make([]response.APIResponse, 0, len(routes))
	for _, route := range routes {
		resp := convertToAPIResponse(route)
		if resp != nil {
			apiResponses = append(apiResponses, *resp)
		}
	}

	baseRes.OkWithDetailed(baseRes.PageResult{
		List:     apiResponses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", ctx)
}

// GetRouteByID 根据 ID 获取 API 路由
//
//	@Summary		根据 ID 获取 API 路由
//	@Description	根据 ID 获取指定的 API 路由信息（管理员权限）
//	@Tags			API 路由管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"路由 ID"
//	@Success		200	{object}	baseRes.Response{data=response.APIResponse}	"路由信息"
//	@Failure		401	{object}	baseRes.Response{msg=string}		"未授权"
//	@Failure		403	{object}	baseRes.Response{msg=string}		"无权限"
//	@Failure		404	{object}	baseRes.Response{msg=string}		"路由不存在"
//	@Failure		500	{object}	baseRes.Response{msg=string}		"服务器内部错误"
//	@Router			/api/admin//apis/{id} [get]
func (c *APIController) GetRouteByID(ctx *gin.Context) {
	_, ok := c.checkAdminPermission(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.log.Error("路由 ID 不能为空")
		baseRes.FailWithMessage("路由 ID 不能为空", ctx)
		return
	}

	route, err := c.apisService.GetRouteByID(id)
	if err != nil {
		c.log.Error("Failed to get route by ID", "error", err, "id", id)
		baseRes.FailWithMessage("获取路由信息失败", ctx)
		return
	}

	c.log.Info("Successfully retrieved route by ID", "id", id)
	routeResp := convertToAPIResponse(route)
	baseRes.OkWithDetailed(routeResp, "获取成功", ctx)
}

// CreateAPI 创建 API 路由
//
//	@Summary		创建 API 路由
//	@Description	创建新的 API 路由信息（管理员权限）
//	@Tags			API 路由管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.CreateAPIRequest		true	"API 路由信息"
//	@Success		200		{object}	baseRes.Response{data=response.APIResponse}	"创建成功的路由信息"
//	@Failure		401		{object}	baseRes.Response{msg=string}		"未授权"
//	@Failure		403		{object}	baseRes.Response{msg=string}		"无权限"
//	@Failure		400		{object}	baseRes.Response{msg=string}		"参数错误"
//	@Failure		500		{object}	baseRes.Response{msg=string}		"服务器内部错误"
//	@Router			/api/admin//apis [post]
func (c *APIController) CreateAPI(ctx *gin.Context) {
	operatorID, ok := c.checkAdminPermission(ctx)
	if !ok {
		return
	}

	var req request.CreateAPIRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("Invalid request parameters", "error", err)
		baseRes.FailWithMessage("参数验证失败", ctx)
		return
	}

	route := &entity.API{
		Path:         req.Path,
		Method:       req.Method,
		Group:        req.Group,
		AuthRequired: req.AuthRequired,
		AuthType:     req.AuthType,
		Description:  req.Description,
		Status:       req.Status,
		RoleIds:      req.RoleIds,
		MenuIds:      req.MenuIds,
		BtnPermIds:   req.BtnPermIds,
	}

	if err := c.apisService.CreateAPIRoute(route, operatorID); err != nil {
		c.log.Error("Failed to create API route", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("API route created successfully", "id", route.ID, "path", route.Path, "method", route.Method)
	routeResp := convertToAPIResponse(route)
	baseRes.OkWithDetailed(routeResp, "创建成功", ctx)
}

// UpdateAPI 更新 API 路由
//
//	@Summary		更新 API 路由
//	@Description	更新指定的 API 路由信息（管理员权限）
//	@Tags			API 路由管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"路由 ID"
//	@Param			data	body		request.UpdateAPIRequest		true	"更新的路由信息"
//	@Success		200		{object}	baseRes.Response{data=response.APIResponse}	"更新成功的路由信息"
//	@Failure		401		{object}	baseRes.Response{msg=string}		"未授权"
//	@Failure		403		{object}	baseRes.Response{msg=string}		"无权限"
//	@Failure		400		{object}	baseRes.Response{msg=string}		"参数错误"
//	@Failure		404		{object}	baseRes.Response{msg=string}		"路由不存在"
//	@Failure		500		{object}	baseRes.Response{msg=string}		"服务器内部错误"
//	@Router			/api/admin//apis/{id} [put]
func (c *APIController) UpdateAPI(ctx *gin.Context) {
	operatorID, ok := c.checkAdminPermission(ctx)
	if !ok {
		return
	}

	var req request.UpdateAPIRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("Invalid request parameters", "error", err)
		baseRes.FailWithMessage("参数验证失败", ctx)
		return
	}

	route := &entity.API{
		ID:           req.ID,
		Path:         req.Path,
		Method:       req.Method,
		Group:        req.Group,
		AuthRequired: req.AuthRequired,
		AuthType:     req.AuthType,
		Description:  req.Description,
		Status:       req.Status,
		RoleIds:      req.RoleIds,
		MenuIds:      req.MenuIds,
		BtnPermIds:   req.BtnPermIds,
	}

	if err := c.apisService.UpdateAPIRoute(route, operatorID); err != nil {
		c.log.Error("Failed to update API route", "error", err, "id", route.ID)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("API route updated successfully", "id", route.ID)
	routeResp := convertToAPIResponse(route)
	baseRes.OkWithDetailed(routeResp, "更新成功", ctx)
}

// DeleteAPI 删除 API 路由
//
//	@Summary		删除 API 路由
//	@Description	删除指定的 API 路由信息（管理员权限）
//	@Tags			API 路由管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"路由 ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		401	{object}	baseRes.Response{msg=string}	"未授权"
//	@Failure		403	{object}	baseRes.Response{msg=string}	"无权限"
//	@Failure		400	{object}	baseRes.Response{msg=string}	"参数错误"
//	@Failure		404	{object}	baseRes.Response{msg=string}	"路由不存在"
//	@Failure		500	{object}	baseRes.Response{msg=string}	"服务器内部错误"
//	@Router			/api/admin//apis/{id} [delete]
func (c *APIController) DeleteAPI(ctx *gin.Context) {
	_, ok := c.checkAdminPermission(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.log.Error("路由 ID 不能为空")
		baseRes.FailWithMessage("路由 ID 不能为空", ctx)
		return
	}

	if err := c.apisService.DeleteAPIRoute(id); err != nil {
		c.log.Error("Failed to delete API route", "error", err, "id", id)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("API route deleted successfully", "id", id)
	baseRes.OkWithMessage("删除成功", ctx)
}

// GetAPIList 获取 API 路由列表（分页）
//
//	@Summary		获取 API 路由列表
//	@Description	分页获取 API 路由列表，支持过滤（管理员权限）
//	@Tags			API 路由管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page			query		int														false	"页码，默认为 1"
//	@Param			pageSize		query		int														false	"每页数量，默认为 10"
//	@Param			group			query		string													false	"路由分组"
//	@Param			authRequired	query		bool													false	"是否需要认证"
//	@Success		200				{object}	baseRes.Response{data=baseRes.PageResult,msg=string}	"路由列表"
//	@Failure		401				{object}	baseRes.Response{msg=string}							"未授权"
//	@Failure		403				{object}	baseRes.Response{msg=string}							"无权限"
//	@Failure		500				{object}	baseRes.Response{msg=string}							"服务器内部错误"
//	@Router			/api/admin//apis/list [get]
func (c *APIController) GetAPIList(ctx *gin.Context) {
	_, ok := c.checkAdminPermission(ctx)
	if !ok {
		return
	}

	var req request.GetAPIPageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	filters := make(map[string]interface{})
	if req.Group != "" {
		filters["group"] = req.Group
	}
	if ctx.Query("authRequired") != "" {
		if ctx.Query("authRequired") == "true" {
			filters["auth_required"] = true
		} else if ctx.Query("authRequired") == "false" {
			filters["auth_required"] = false
		}
	}

	routes, total, err := c.apisService.GetAPIRouteList(req.Page, req.PageSize, filters)
	if err != nil {
		c.log.Error("Failed to get API route list", "error", err)
		baseRes.FailWithMessage("获取路由列表失败", ctx)
		return
	}

	c.log.Info("API route list retrieved successfully", "count", len(routes), "total", total)

	apiResponses := make([]response.APIResponse, 0, len(routes))
	for _, route := range routes {
		resp := convertToAPIResponse(route)
		if resp != nil {
			apiResponses = append(apiResponses, *resp)
		}
	}

	baseRes.OkWithDetailed(baseRes.PageResult{
		List:     apiResponses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", ctx)
}
