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

// DepartmentController 部门控制器
// 处理部门相关的 HTTP 请求
// 提供部门增删改查、树形结构等功能的 API 接口
// 字段:
//   - service: 部门服务，处理业务逻辑
//   - log: 日志记录器，记录操作日志
//     @Summary		部门管理 API
//     @Description	提供部门创建、更新、删除、查询等功能
//     @Tags			部门管理
//     @Router			/api/admin/dept [get]
type DepartmentController struct {
	service *service.DepartmentService // 部门服务接口
	log     logger.Logger              // 日志记录器
}

// NewDepartmentController 创建部门控制器实例
// 参数:
// - service: 部门服务接口实现
// - log: 日志记录器
// 返回:
// - *DepartmentController: 部门控制器实例
func NewDepartmentController(service *service.DepartmentService, log logger.Logger) *DepartmentController {
	return &DepartmentController{
		service: service,
		log:     log,
	}
}

// convertToDepartmentResponse 将 entity.Department 转换为 response.DepartmentResponse
func convertToDepartmentResponse(dept *entity.Department) response.DepartmentResponse {
	return response.DepartmentResponse{
		ID:        dept.ID,
		ParentID:  dept.ParentID,
		Name:      dept.Name,
		Sort:      dept.Sort,
		Status:    dept.Status,
		Leader:    "", // model 中没有直接的 Leader 字段，只有 LeaderID 和 Leader 关联
		Phone:     "", // model 中没有 Phone 字段
		Email:     "", // model 中没有 Email 字段
		CreatedAt: dept.CreatedAt.Format(time.RFC3339),
		UpdatedAt: dept.UpdatedAt.Format(time.RFC3339),
	}
}

// GetDepartmentList 获取部门列表
//
//	@Summary		获取部门列表
//	@Description	获取部门列表（支持分页和筛选）
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int																				true	"页码"
//	@Param			page_size	query		int																				true	"每页数量"
//	@Param			parent_id	query		int																				false	"父部门 ID"
//	@Param			status		query		int																				false	"状态 (0:禁用，1:启用)"
//	@Success		200			{object}	baseRes.Response{data=response.DepartmentListResponse,msg=string}	"部门列表"
//	@Failure		401			{object}	map[string]string																"未授权"
//	@Failure		500			{object}	map[string]string																"服务器内部错误"
//	@Router			/api/admin/dept [get]
func (c *DepartmentController) GetDepartmentList(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetDepartmentListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 构建筛选条件
	filters := make(map[string]interface{})
	if req.ParentID != 0 {
		filters["parent_id"] = req.ParentID
	}
	if req.Status != nil {
		filters["status"] = *req.Status
	}

	departments, total, err := c.service.GetDepartmentList(req.Page, req.PageSize, filters)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为响应结构
	responses := make([]*response.DepartmentResponse, 0, len(departments))
	for _, dept := range departments {
		resp := convertToDepartmentResponse(dept)
		responses = append(responses, &resp)
	}

	departmentListResponse := response.DepartmentListResponse{
		PageResult: baseRes.PageResult{
			List:     responses,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		List: responses,
	}

	baseRes.OkWithDetailed(departmentListResponse, "获取部门列表成功", ctx)
}

// GetDepartmentTree 获取部门树形结构
//
//	@Summary		获取部门树形结构
//	@Description	获取完整的部门树形结构数据
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{data=[]entity.Department,msg=string}	"部门树"
//	@Failure		401	{object}	map[string]string										"未授权"
//	@Failure		500	{object}	map[string]string										"服务器内部错误"
//	@Router			/api/admin/dept/tree [get]
func (c *DepartmentController) GetDepartmentTree(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	tree, err := c.service.GetDepartmentTree()
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(tree, "获取部门树成功", ctx)
}

// GetDepartmentByID 获取部门详情
//
//	@Summary		获取部门详情
//	@Description	根据 ID 获取部门详细信息
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	query		int													true	"部门 ID"
//	@Success		200	{object}	baseRes.Response{data=response.DepartmentResponse,msg=string}	"部门详情"
//	@Failure		400	{object}	map[string]string									"请求参数错误"
//	@Failure		401	{object}	map[string]string									"未授权"
//	@Failure		500	{object}	map[string]string									"服务器内部错误"
//	@Router			/api/admin/dept/:id [get]
func (c *DepartmentController) GetDepartmentByID(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetDepartmentByIDRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	department, err := c.service.GetDepartmentByID(req.ID)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(department, "获取部门详情成功", ctx)
}

// CreateDepartment 创建部门
//
//	@Summary		创建部门
//	@Description	创建新的部门
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.CreateDepartmentRequest						true	"部门信息"
//	@Success		200		{object}	baseRes.Response{data=response.DepartmentResponse,msg=string}	"创建成功"
//	@Failure		400		{object}	map[string]string									"请求参数错误"
//	@Failure		401		{object}	map[string]string									"未授权"
//	@Failure		500		{object}	map[string]string									"服务器内部错误"
//	@Router			/api/admin/dept [post]
func (c *DepartmentController) CreateDepartment(ctx *gin.Context) {
	var req request.CreateDepartmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 获取当前登录用户 ID 作为创建者
	createdBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 提供默认值：sort=0, status=1（启用）
	sort := req.Sort
	if sort == 0 {
		sort = 0
	}
	status := req.Status
	if status == 0 {
		status = 1 // 默认启用
	}

	department, err := c.service.CreateDepartment(
		req.Name,
		req.Description,
		req.ParentID,
		req.LeaderID,
		createdBy.(string),
		sort,
		status,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("创建部门成功", "id", department.ID, "name", req.Name)
	baseRes.OkWithDetailed(department, "创建部门成功", ctx)
}

// UpdateDepartment 更新部门
//
//	@Summary		更新部门
//	@Description	更新部门信息
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.UpdateDepartmentRequest						true	"部门信息"
//	@Success		200		{object}	baseRes.Response{data=response.DepartmentResponse,msg=string}	"更新成功"
//	@Failure		400		{object}	map[string]string									"请求参数错误"
//	@Failure		401		{object}	map[string]string									"未授权"
//	@Failure		500		{object}	map[string]string									"服务器内部错误"
//	@Router			/api/admin/dept [put]
func (c *DepartmentController) UpdateDepartment(ctx *gin.Context) {
	var req request.UpdateDepartmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 获取当前登录用户 ID 作为更新者
	updatedBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 直接使用 string 类型的 ID
	department, err := c.service.UpdateDepartment(
		req.ID,
		req.Name,
		req.Description,
		req.ParentID,
		req.LeaderID,
		updatedBy.(string),
		req.Sort,
		req.Status,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("更新部门成功", "id", req.ID, "name", req.Name)
	baseRes.OkWithDetailed(department, "更新部门成功", ctx)
}

// DeleteDepartment 删除部门
//
//	@Summary		删除部门
//	@Description	删除部门（管理员权限）
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"部门 ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/dept/:id [delete]
func (c *DepartmentController) DeleteDepartment(ctx *gin.Context) {
	// 直接使用 string 类型的部门 ID
	departmentID := ctx.Param("id")
	if departmentID == "" {
		c.log.Error("部门 ID 不能为空")
		baseRes.FailWithMessage("部门 ID 不能为空", ctx)
		return
	}

	if err := c.service.DeleteDepartment(departmentID); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("删除部门成功", ctx)
}

// UpdateDepartmentLeader 更新部门负责人
//
//	@Summary		更新部门负责人
//	@Description	更新部门负责人
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int										true	"部门 ID"
//	@Param			data	body		request.UpdateDepartmentLeaderRequest	true	"部门负责人信息"
//	@Success		200		{object}	baseRes.Response{msg=string}			"更新成功"
//	@Failure		400		{object}	map[string]string						"请求参数错误"
//	@Failure		401		{object}	map[string]string						"未授权"
//	@Failure		500		{object}	map[string]string						"服务器内部错误"
//	@Router			/api/admin/dept/:id/leader [put]
func (c *DepartmentController) UpdateDepartmentLeader(ctx *gin.Context) {
	// 直接使用 string 类型的部门 ID
	departmentID := ctx.Param("id")
	if departmentID == "" {
		c.log.Error("部门 ID 不能为空")
		baseRes.FailWithMessage("部门 ID 不能为空", ctx)
		return
	}

	var req request.UpdateDepartmentLeaderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 获取当前登录用户 ID 作为更新者
	updatedBy, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	if err := c.service.UpdateDepartmentLeader(departmentID, req.LeaderID, updatedBy.(string)); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("更新部门负责人成功", "id", departmentID, "leader_id", req.LeaderID)
	baseRes.OkWithMessage("更新部门负责人成功", ctx)
}
