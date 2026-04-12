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

// NoticeController 公告控制器
// 处理公告相关的 HTTP 请求
// 提供公告发布、编辑、删除、阅读状态跟踪等功能的 API 接口
// 字段:
//   - service: 公告服务，处理业务逻辑
//   - recordService: 阅读记录服务，处理阅读记录相关逻辑
//   - log: 日志记录器，记录操作日志
//     @Summary		公告管理 API
//     @Description	提供公告创建、更新、删除、查询、阅读标记等功能
//     @Tags			公告管理
//     @Router			/api/admin//notices [get]
type NoticeController struct {
	service       *service.NoticeService           // 公告服务接口
	recordService *service.NoticeReadRecordService // 阅读记录服务接口
	log           logger.Logger                    // 日志记录器
}

// NewNoticeController 创建公告控制器实例
// 参数:
// - service: 公告服务接口实现
// - recordService: 阅读记录服务接口实现
// - log: 日志记录器
// 返回:
// - *NoticeController: 公告控制器实例
func NewNoticeController(service *service.NoticeService, recordService *service.NoticeReadRecordService, log logger.Logger) *NoticeController {
	return &NoticeController{
		service:       service,
		recordService: recordService,
		log:           log,
	}
}

// GetNoticeList 获取公告列表
//
//	@Summary		获取公告列表
//	@Description	获取公告列表（支持分页和筛选）
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int																			true	"页码"
//	@Param			page_size	query		int																			true	"每页数量"
//	@Param			type		query		int																			false	"公告类型 (1:系统公告，2:活动公告，3:普通通知，4:紧急通知)"
//	@Param			status		query		int																			false	"公告状态 (0:草稿，1:已发布，2:已归档)"
//	@Param			is_top		query		bool																		false	"是否置顶"
//	@Success		200			{object}	baseRes.Response{data=response.NoticeListResponse,msg=string}	"公告列表"
//	@Failure		401			{object}	map[string]string															"未授权"
//	@Failure		500			{object}	map[string]string															"服务器内部错误"
//	@Router			/api/admin//notices [get]
func (c *NoticeController) GetNoticeList(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	var req request.GetNoticeListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.log.Error("请求参数错误", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 构建筛选条件
	filters := make(map[string]interface{})
	if req.Type != nil {
		filters["type"] = *req.Type
	}
	if req.Status != nil {
		filters["status"] = *req.Status
	}
	if req.IsTop != nil {
		filters["is_top"] = *req.IsTop
	}

	notices, total, err := c.service.GetNoticeList(req.Page, req.PageSize, filters)
	if err != nil {
		c.log.Error("获取公告列表失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为响应 DTO
	noticeResponses := make([]response.NoticeResponse, len(notices))
	for i, notice := range notices {
		noticeResponses[i] = response.NoticeResponse{
			ID:          notice.ID,
			Title:       notice.Title,
			Content:     notice.Content,
			Type:        int(notice.Type),
			Status:      int(notice.Status),
			PublisherID: notice.PublisherID,
			ViewCount:   notice.ViewCount,
			IsTop:       notice.IsTop,
			Sort:        notice.Sort,
			Description: notice.Description,
			CreatedAt:   notice.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   notice.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if notice.PublishTime != nil {
			noticeResponses[i].PublishTime = notice.PublishTime.Format("2006-01-02 15:04:05")
		}
	}

	noticeListResponse := response.NoticeListResponse{
		PageResult: baseRes.PageResult{
			List:     noticeResponses,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		List: noticeResponses,
	}

	baseRes.OkWithDetailed(noticeListResponse, "获取公告列表成功", ctx)
}

// GetNoticeByID 获取公告详情
//
//	@Summary		获取公告详情
//	@Description	根据 ID 获取公告详细信息
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string														true	"公告 ID"
//	@Success		200	{object}	baseRes.Response{data=response.NoticeResponse,msg=string}	"公告详情"
//	@Failure		400	{object}	map[string]string													"请求参数错误"
//	@Failure		401	{object}	map[string]string													"未授权"
//	@Failure		500	{object}	map[string]string													"服务器内部错误"
//	@Router			/api/admin//notices/:id [get]
func (c *NoticeController) GetNoticeByID(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 直接使用 string 类型的公告 ID
	noticeID := ctx.Param("id")
	if noticeID == "" {
		c.log.Error("公告 ID 不能为空")
		baseRes.FailWithMessage("公告 ID 不能为空", ctx)
		return
	}

	notice, err := c.service.GetNoticeByID(noticeID)
	if err != nil {
		c.log.Error("获取公告详情失败", "error", err, "id", noticeID)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为响应 DTO
	noticeResponse := response.NoticeResponse{
		ID:          notice.ID,
		Title:       notice.Title,
		Content:     notice.Content,
		Type:        int(notice.Type),
		Status:      int(notice.Status),
		PublisherID: notice.PublisherID,
		ViewCount:   notice.ViewCount,
		IsTop:       notice.IsTop,
		Sort:        notice.Sort,
		Description: notice.Description,
		CreatedAt:   notice.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   notice.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if notice.PublishTime != nil {
		noticeResponse.PublishTime = notice.PublishTime.Format("2006-01-02 15:04:05")
	}

	baseRes.OkWithDetailed(noticeResponse, "获取公告详情成功", ctx)
}

// CreateNotice 创建公告
//
//	@Summary		创建公告
//	@Description	创建新的公告（草稿状态）
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.CreateNoticeRequest								true	"公告信息"
//	@Success		200		{object}	baseRes.Response{data=response.NoticeResponse,msg=string}	"创建成功"
//	@Failure		400		{object}	map[string]string												"请求参数错误"
//	@Failure		401		{object}	map[string]string												"未授权"
//	@Failure		500		{object}	map[string]string												"服务器内部错误"
//	@Router			/api/admin//notices [post]
func (c *NoticeController) CreateNotice(ctx *gin.Context) {
	var req request.CreateNoticeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("创建公告参数验证失败", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 获取当前登录用户 ID 作为发布人
	publisherID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 提供默认值：sort=0（不置顶）
	sort := req.Sort

	notice, err := c.service.CreateNotice(
		req.Title,
		req.Content,
		req.Description,
		entity.NoticeType(req.Type),
		publisherID.(string),
		req.IsTop,
		sort,
	)
	if err != nil {
		c.log.Error("创建公告失败", "error", err, "title", req.Title)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为响应 DTO
	noticeResponse := response.NoticeResponse{
		ID:          notice.ID,
		Title:       notice.Title,
		Content:     notice.Content,
		Type:        int(notice.Type),
		Status:      int(notice.Status),
		PublisherID: notice.PublisherID,
		ViewCount:   notice.ViewCount,
		IsTop:       notice.IsTop,
		Sort:        notice.Sort,
		Description: notice.Description,
		CreatedAt:   notice.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   notice.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if notice.PublishTime != nil {
		noticeResponse.PublishTime = notice.PublishTime.Format("2006-01-02 15:04:05")
	}

	c.log.Info("创建公告成功", "id", notice.ID, "title", req.Title)
	baseRes.OkWithDetailed(noticeResponse, "创建公告成功", ctx)
}

// UpdateNotice 更新公告
//
//	@Summary		更新公告
//	@Description	更新公告信息（仅限草稿状态）
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			data	body		request.UpdateNoticeRequest												true	"公告信息"
//	@Success		200		{object}	baseRes.Response{data=response.NoticeResponse,msg=string}	"更新成功"
//	@Failure		400		{object}	map[string]string															"请求参数错误"
//	@Failure		401		{object}	map[string]string															"未授权"
//	@Failure		500		{object}	map[string]string															"服务器内部错误"
//	@Router			/api/admin//notices [put]
func (c *NoticeController) UpdateNotice(ctx *gin.Context) {
	var req request.UpdateNoticeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("更新公告参数验证失败", "error", err)
		baseRes.FailWithMessage("请求参数错误", ctx)
		return
	}

	// 获取当前登录用户 ID 作为更新者
	publisherID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	// 直接使用 string 类型的 ID
	notice, err := c.service.UpdateNotice(
		req.ID,
		req.Title,
		req.Content,
		req.Description,
		entity.NoticeType(req.Type),
		publisherID.(string),
		req.IsTop,
		req.Sort,
	)
	if err != nil {
		c.log.Error("更新公告失败", "error", err, "id", req.ID)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为响应 DTO
	noticeResponse := response.NoticeResponse{
		ID:          notice.ID,
		Title:       notice.Title,
		Content:     notice.Content,
		Type:        int(notice.Type),
		Status:      int(notice.Status),
		PublisherID: notice.PublisherID,
		ViewCount:   notice.ViewCount,
		IsTop:       notice.IsTop,
		Sort:        notice.Sort,
		Description: notice.Description,
		CreatedAt:   notice.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   notice.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if notice.PublishTime != nil {
		noticeResponse.PublishTime = notice.PublishTime.Format("2006-01-02 15:04:05")
	}

	c.log.Info("更新公告成功", "id", notice.ID, "title", req.Title)
	baseRes.OkWithDetailed(noticeResponse, "更新公告成功", ctx)
}

// DeleteNotice 删除公告
//
//	@Summary		删除公告
//	@Description	删除公告（管理员权限）
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"公告 ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"删除成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin//notices/{id} [delete]
func (c *NoticeController) DeleteNotice(ctx *gin.Context) {
	// 直接使用 string 类型的公告 ID
	noticeID := ctx.Param("id")
	if noticeID == "" {
		c.log.Error("公告 ID 不能为空")
		baseRes.FailWithMessage("公告 ID 不能为空", ctx)
		return
	}

	if err := c.service.DeleteNotice(noticeID); err != nil {
		c.log.Error("删除公告失败", "error", err, "id", noticeID)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithMessage("删除公告成功", ctx)
}

// PublishNotice 发布公告
//
//	@Summary		发布公告
//	@Description	将草稿状态的公告发布
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"公告 ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"发布成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin//notices/{id}/publish [post]
func (c *NoticeController) PublishNotice(ctx *gin.Context) {
	// 直接使用 string 类型的公告 ID
	noticeID := ctx.Param("id")
	if noticeID == "" {
		c.log.Error("公告 ID 不能为空")
		baseRes.FailWithMessage("公告 ID 不能为空", ctx)
		return
	}

	// 获取当前登录用户 ID 作为发布人
	publisherID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	if err := c.service.PublishNotice(noticeID, publisherID.(string)); err != nil {
		c.log.Error("发布公告失败", "error", err, "id", noticeID)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("发布公告成功", "id", noticeID, "publisher_id", publisherID)
	baseRes.OkWithMessage("发布公告成功", ctx)
}

// MarkAsRead 标记公告已读
//
//	@Summary		标记公告已读
//	@Description	标记公告为已读状态，并增加浏览次数
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"公告 ID"
//	@Success		200	{object}	baseRes.Response{msg=string}	"标记成功"
//	@Failure		400	{object}	map[string]string				"请求参数错误"
//	@Failure		401	{object}	map[string]string				"未授权"
//	@Failure		500	{object}	map[string]string				"服务器内部错误"
//	@Router			/api/admin//notices/:id/read [post]
func (c *NoticeController) MarkAsRead(ctx *gin.Context) {
	// 直接使用 string 类型的公告 ID
	noticeID := ctx.Param("id")
	if noticeID == "" {
		c.log.Error("公告 ID 不能为空")
		baseRes.FailWithMessage("公告 ID 不能为空", ctx)
		return
	}

	// 获取当前登录用户 ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	if err := c.service.MarkAsRead(noticeID, userID.(string)); err != nil {
		c.log.Error("标记公告已读失败", "error", err, "id", noticeID, "user_id", userID)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	c.log.Info("标记公告已读", "id", noticeID, "user_id", userID)
	baseRes.OkWithMessage("标记公告已读成功", ctx)
}

// GetStatistics 获取公告统计
//
//	@Summary		获取公告统计
//	@Description	获取公告的统计信息（总数、已发布、草稿、已归档）
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	baseRes.Response{data=response.NoticeStatisticsResponse,msg=string}	"统计信息"
//	@Failure		401	{object}	map[string]string																	"未授权"
//	@Failure		500	{object}	map[string]string																	"服务器内部错误"
//	@Router			/api/admin//notices/statistics [get]
func (c *NoticeController) GetStatistics(ctx *gin.Context) {
	// 检查用户是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	stats, err := c.service.GetStatistics()
	if err != nil {
		c.log.Error("获取公告统计失败", "error", err)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	// 转换为响应 DTO
	statsResponse := response.NoticeStatisticsResponse{
		TotalCount:     stats.TotalCount,
		PublishedCount: stats.PublishedCount,
		DraftCount:     stats.DraftCount,
		ArchivedCount:  stats.ArchivedCount,
	}

	baseRes.OkWithDetailed(statsResponse, "获取公告统计成功", ctx)
}

// CheckIsRead 检查用户是否已读公告
//
//	@Summary		检查用户是否已读公告
//	@Description	检查当前用户是否已阅读指定公告
//	@Tags			公告管理
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string									true	"公告 ID"
//	@Success		200	{object}	baseRes.Response{data=bool,msg=string}	"是否已读"
//	@Failure		400	{object}	map[string]string						"请求参数错误"
//	@Failure		401	{object}	map[string]string						"未授权"
//	@Failure		500	{object}	map[string]string						"服务器内部错误"
//	@Router			/api/admin//notices/:id/is-read [get]
func (c *NoticeController) CheckIsRead(ctx *gin.Context) {
	// 直接使用 string 类型的公告 ID
	noticeID := ctx.Param("id")
	if noticeID == "" {
		c.log.Error("公告 ID 不能为空")
		baseRes.FailWithMessage("公告 ID 不能为空", ctx)
		return
	}

	// 获取当前登录用户 ID
	userID, exists := ctx.Get("userID")
	if !exists {
		c.log.Error("未登录")
		baseRes.NoAuth("未登录", ctx)
		return
	}

	isRead, err := c.service.IsRead(noticeID, userID.(string))
	if err != nil {
		c.log.Error("检查阅读状态失败", "error", err, "id", noticeID, "user_id", userID)
		baseRes.FailWithMessage(err.Error(), ctx)
		return
	}

	baseRes.OkWithDetailed(isRead, "检查阅读状态成功", ctx)
}
