package baseapi

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/service"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/request"
	"github.com/ix-pay/ixpay-pro/internal/dto/base/response"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	"github.com/ix-pay/ixpay-pro/internal/utils/common/baseRes"
)

// PositionController 岗位控制器
// 处理岗位相关的 HTTP 请求
// 提供岗位增删改查等功能的 API 接口
// 字段:
//   - service: 岗位服务，处理业务逻辑
//   - log: 日志记录器，记录操作日志
//     @Summary		岗位管理 API
//     @Description	提供岗位创建、更新、删除、查询等功能
//     @Tags			岗位管理
//     @Router			/api/admin/position [get]
type PositionController struct {
	service *service.PositionService // 岗位服务
	log     logger.Logger            // 日志记录器
}

// NewPositionController 创建岗位控制器实例
// 参数:
// - service: 岗位服务接口实现
// - log: 日志记录器
// 返回:
// - *PositionController: 岗位控制器实例
func NewPositionController(service *service.PositionService, log logger.Logger) *PositionController {
	return &PositionController{
		service: service,
		log:     log,
	}
}

// convertToPositionResponse 将 entity.Position 转换为 response.PositionResponse
func convertToPositionResponse(position *entity.Position) response.PositionResponse {
	return response.PositionResponse{
		ID:          position.ID,
		Name:        position.Name,
		Code:        "", // Position 模型没有 Code 字段
		Sort:        position.Sort,
		Status:      position.Status,
		Description: position.Description,
		CreatedAt:   position.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   position.UpdatedAt.Format(time.RFC3339),
	}
}

// GetPositionList 获取岗位列表
//
//	@Summary		获取岗位列表
//	@Description	获取岗位列表（支持分页和筛选）
//	@Tags			岗位管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int																			true	"页码"
//	@Param			page_size	query		int																			true	"每页数量"
//	@Param			status		query		int																			false	"状态 (0:禁用，1:启用)"
//	@Success		200			{object}	baseRes.Response{data=response.PositionListResponse,msg=string}	"岗位列表"
//	@Failure		401			{object}	map[string]string															"未授权"
//	@Failure		500			{object}	map[string]string															"服务器内部错误"
//	@Router			/api/admin/position [get]
func (c *PositionController) GetPositionList(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetPositionListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 构建筛选条件
	filters := make(map[string]interface{})
	if req.Status != nil {
		filters["status"] = *req.Status
	}

	positions, total, err := c.service.GetPositionList(req.Page, req.PageSize, filters)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为响应结构
	responses := make([]response.PositionResponse, 0, len(positions))
	for _, position := range positions {
		responses = append(responses, convertToPositionResponse(position))
	}

	positionListResponse := response.PositionListResponse{
		PageResult: baseRes.PageResult{
			List:     responses,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		List: responses,
	}

	baseRes.OkWithDetailed(positionListResponse, "获取岗位列表成功", ctx)
}

// GetAllPositions 获取所有岗位
//
//	@Summary		获取所有岗位
//	@Description	获取完整的岗位列表数据
//	@Tags			岗位管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{data=[]entity.Position,msg=string}	"岗位列表"
//	@Failure		401	{object}	map[string]string									"未授权"
//	@Failure		500	{object}	map[string]string									"服务器内部错误"
//	@Router			/api/admin/position/all [get]
func (c *PositionController) GetAllPositions(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	positions, err := c.service.GetAllPositions()
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(positions, "获取所有岗位成功", ctx)
}

// GetPositionByID 获取岗位详情
//
//	@Summary		获取岗位详情
//	@Description	根据 ID 获取岗位详细信息
//	@Tags			岗位管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	query		int													true	"岗位 ID"
//	@Success		200	{object}	baseRes.Response{data=response.PositionResponse,msg=string}	"岗位详情"
//	@Failure		400	{object}	map[string]string									"请求参数错误"
//	@Failure		401	{object}	map[string]string									"未授权"
//	@Failure		500	{object}	map[string]string									"服务器内部错误"
//	@Router			/api/admin/position/:id [get]
func (c *PositionController) GetPositionByID(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetPositionByIDRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	position, err := c.service.GetPositionByID(req.ID)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(position, "获取岗位详情成功", ctx)
}

// CreatePosition 创建岗位
//
//	@Summary		创建岗位
//	@Description	创建新的岗位
//	@Tags			岗位管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.CreatePositionRequest						true	"岗位信息"
//	@Success		200		{object}	baseRes.Response{data=response.PositionResponse,msg=string}	"创建成功"
//	@Failure		400		{object}	map[string]string									"请求参数错误"
//	@Failure		401		{object}	map[string]string									"未授权"
//	@Failure		500		{object}	map[string]string									"服务器内部错误"
//	@Router			/api/admin/position [post]
func (c *PositionController) CreatePosition(ctx *gin.Context) {
	var req request.CreatePositionRequest
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

	// 将 createdBy 转换为 int64
	var createdByInt int64
	var err error
	switch v := createdBy.(type) {
	case string:
		createdByInt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.log.Error("用户 ID 格式错误", "error", err)
			baseRes.FailWithMessage("用户 ID 格式错误", ctx)
			return
		}
	case int64:
		createdByInt = v
	case int:
		createdByInt = int64(v)
	default:
		c.log.Error("用户 ID 类型错误", "actual_type", v)
		baseRes.FailWithMessage("用户 ID 类型错误", ctx)
		return
	}

	// 提供默认值：status=1（启用）
	status := req.Status
	if status == 0 {
		status = 1 // 默认启用
	}

	position, err := c.service.CreatePosition(
		req.Name,
		req.Description,
		createdByInt,
		req.Sort,
		status,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("创建岗位成功", "id", position.ID, "name", req.Name)
	baseRes.OkWithDetailed(position, "创建岗位成功", ctx)
}

// UpdatePosition 更新岗位
//
//	@Summary		更新岗位
//	@Description	更新岗位信息
//	@Tags			岗位管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.UpdatePositionRequest						true	"岗位信息"
//	@Success		200		{object}	baseRes.Response{data=response.PositionResponse,msg=string}	"更新成功"
//	@Failure		400		{object}	map[string]string									"请求参数错误"
//	@Failure		401		{object}	map[string]string									"未授权"
//	@Failure		500		{object}	map[string]string									"服务器内部错误"
//	@Router			/api/admin/position [put]
func (c *PositionController) UpdatePosition(ctx *gin.Context) {
	var req request.UpdatePositionRequest
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

	// 将 updatedBy 转换为 int64
	var updatedByInt int64
	var err error
	switch v := updatedBy.(type) {
	case string:
		updatedByInt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.log.Error("用户 ID 格式错误", "error", err)
			baseRes.FailWithMessage("用户 ID 格式错误", ctx)
			return
		}
	case int64:
		updatedByInt = v
	case int:
		updatedByInt = int64(v)
	default:
		c.log.Error("用户 ID 类型错误", "actual_type", v)
		baseRes.FailWithMessage("用户 ID 类型错误", ctx)
		return
	}

	// 调用服务层更新岗位
	position, err := c.service.UpdatePosition(
		req.ID,
		req.Name,
		req.Description,
		updatedByInt,
		req.Sort,
		req.Status,
	)
	if err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("更新岗位成功", "id", position.ID, "name", req.Name)
	baseRes.OkWithDetailed(position, "更新岗位成功", ctx)
}

// DeletePosition 删除岗位
//
//	@Summary		删除岗位
//	@Description	删除岗位（管理员权限）
//	@Tags			岗位管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"岗位 ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin/position/:id [delete]
func (c *PositionController) DeletePosition(ctx *gin.Context) {
	// 将字符串 ID 转换为 int64
	positionIDStr := ctx.Param("id")
	if positionIDStr == "" {
		c.log.Error("岗位 ID 不能为空")
		baseRes.FailWithMessage("岗位 ID 不能为空", ctx)
		return
	}

	positionID, err := strconv.ParseInt(positionIDStr, 10, 64)
	if err != nil {
		c.log.Error("无效的 ID 格式", "id", positionIDStr, "error", err)
		baseRes.FailWithMessage("无效的 ID 格式", ctx)
		return
	}

	if err := c.service.DeletePosition(positionID); err != nil {
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("删除岗位成功", ctx)
}
